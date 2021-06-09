package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

type HTTPTrigger struct {
	MemoryMB int32
	Callback func(r *http.Request, w http.ResponseWriter)
}

type PubSubTrigger struct {
	Topic    string
	MemoryMB int32
	Callback interface{}
}

func OnRequest(callback func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx != nil && ctx.Value("firebase_discovery") == true {
			fmt.Println("Discovery phase")
			return
		}

		callback(w, r)
	}
}

func OnPubSubEvent(topic string, callback interface{}) func(ctx context.Context, data json.RawMessage) {
	callbackType := reflect.TypeOf(callback)
	argType := callbackType.In(1)
	if argType.Kind() == reflect.Ptr {
		argType = argType.Elem()
	}

	val := reflect.New(argType)
	p := val.Interface()
	return func(ctx context.Context, data json.RawMessage) {
		if ctx.Value("firebase_discovery") == true {
			fmt.Println("Discovery phase")
			fmt.Printf("topic: %s\n", topic)
			return
		}

		if err := json.Unmarshal(data, p); err != nil {
			panic(err)
		}

		reflect.ValueOf(callback).Call([]reflect.Value{
			reflect.ValueOf(ctx),
			val,
		})
	}
}
