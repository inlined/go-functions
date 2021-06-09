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

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// Support both a Foo and a Foo* by tracking the arg type and value type separately.
		// If argType is a Foo then valueType is a Foo. If argType is a *Foo then valueType
		// is a Foo.
		argType := callback.Type().In(1)
		valueType := argType
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
		}
		valuePtr := reflect.New(valueType)
		json.NewDecoder(r.Body).Decode(valuePtr.Interface())

		// Now we need to get an actual argument of type argType.
		// valuePtr is type *Foo, so arg will start as Foo and
		// then become *Foo again if argType is a *Foo
		arg := valuePtr.Elem()
		if arg.Type() != argType {
			arg = arg.Addr()
		}

		callback.Call([]reflect.Value{reflect.ValueOf(r.Context()), arg})
		w.WriteHeader(200)
	}
}

// Consider allowing HTTP handler functions to be directly detectable
// and turned into an HttpFunction(-like thing) with default options.
// This would require reimplementing a shim type because we can't have
// a circular reference between the https package and this package.
func Serve(symbols map[string]interface{}) {
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
	fmt.Printf("Serving emulator at http://localhost:%d\n", port)
	if adminPort != 0 {
		fmt.Printf("Serving emulator admin API at http://localhost:%d\n", adminPort)
	}

	d := functionData{}
	mux := http.NewServeMux()
	adminMux := http.NewServeMux()

	for symbol, value := range symbols {
		if asFunc, ok := value.(function); ok {
			d[symbol] = asFunc
		}
	}

	for symbol, function := range d {
		fmt.Printf("Serving function at http://localhost:%d/%s\n", port, symbol)
		mux.HandleFunc(fmt.Sprintf("/%s", symbol), getHandler(function))
	}

	adminMux.HandleFunc("/backend.yaml", d.DescribeBackend)

	// TODO: graceful shutdown on SIGINT
	done := make(chan struct{}, 2)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), mux)
		if err != nil {
			fmt.Println("Emulator exited with error", err)
		}
		done <- struct{}{}
	}()
	go func() {
		if adminPort != 0 {
			err := http.ListenAndServe(fmt.Sprintf("localhost:%d", adminPort), adminMux)
			if err != nil {
				fmt.Println("Emulator admin API exited with error", err)
			}
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}

type function interface {
	AddBackendDescription(symbolName string, b *Backend)
}
