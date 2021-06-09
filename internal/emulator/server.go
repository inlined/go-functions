package emulator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/go-yaml/yaml"
)

type functionData map[string]function

func (f functionData) DescribeBackend(w http.ResponseWriter, r *http.Request) {
	b := Backend{
		RequiredAPIs:   map[string]string{},
		CloudFunctions: make([]FunctionSpec, 0),
		Topics:         make([]PubSubSpec, 0),
		Schedules:      make([]ScheduleSpec, 0),
	}
	for symbol, function := range f {
		function.AddBackendDescription(symbol, &b)
	}
	yaml.NewEncoder(w).Encode(b)
}

func getHandler(f function) func(http.ResponseWriter, *http.Request) {
	// TODO: Should we handle a Callback field that isn't interface{}?
	callback := reflect.ValueOf(f).FieldByName("Callback")
	if callback.Type().Kind() == reflect.Interface {
		callback = callback.Elem()
	}
	if !callback.IsValid() {
		fmt.Printf("Could not find Callback field in %+v\n", f)
	} else {
		fmt.Println("Callback is type", callback.Type())
	}
	if callback.Kind() != reflect.Func {
		panic("CloudFunctions should have a Callback function")
	}
	if callback.Type().NumIn() != 2 {
		panic("CloudFunctions' Callback should take two parameters")
	}

	if httpHandler, ok := callback.Interface().(func(http.ResponseWriter, *http.Request)); ok {
		return httpHandler
	}

	if callback.Type().In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		panic("Event-handling CloudFunctions should take a first parameter of *context.Context")
	}

	argType := callback.Type().In(1)
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		argPtr := reflect.New(argType)
		argValue := argPtr.Elem()
		arg := argValue.Interface()
		json.NewDecoder(r.Body).Decode(arg)
		// Must hold this in an intermediary to force the coercion
		var ctx context.Context = r.Context()
		callback.Call([]reflect.Value{reflect.ValueOf(ctx), argValue})
		w.WriteHeader(200)
	}
}

// Consider allowing HTTP handler functions to be directly detectable
// and turned into an HttpFunction(-like thing) with default options.
// This would require reimplementing a shim type because we can't have
// a circular reference between the https package and this package.
func Serve(symbols map[string]interface{}) {
	d := functionData{}
	mux := http.NewServeMux()
	adminMux := http.NewServeMux()

	for symbol, value := range symbols {
		if asFunc, ok := value.(function); ok {
			d[symbol] = asFunc
		}
	}

	for symbol, function := range d {
		fmt.Println("About to add symbol", symbol)
		mux.HandleFunc(fmt.Sprintf("/%s", symbol), getHandler(function))
	}

	adminMux.HandleFunc("/__/backend.yaml", d.DescribeBackend)

	var port int64 = 8080
	var err error
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err = strconv.ParseInt(portStr, 10, 16); err != nil {
			panic("environment variable PORT must be an int")
		}
	}
	// THIS SHOULD BE 0 WHEN NOT USING GO RUN
	var adminPort int64 = 8081
	if portStr := os.Getenv("ADMIN_PORT"); portStr != "" {
		if adminPort, err = strconv.ParseInt(portStr, 10, 16); err != nil {
			panic("environment varialbe ADMIN_PORT must be an int")
		}
	}

	// TODO: graceful shutdown on SIGINT
	done := make(chan struct{}, 2)
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
		done <- struct{}{}
	}()
	go func() {
		if adminPort != 0 {
			http.ListenAndServe(fmt.Sprintf(":%d", adminPort), adminMux)
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}

type function interface {
	AddBackendDescription(symbolName string, b *Backend)
}
