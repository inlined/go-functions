package functions

import (
	"context"
	"fmt"

	"github.com/inlined/go-functions/https"
	"github.com/inlined/go-functions/pubsub"
	"github.com/inlined/go-functions/runwith"
)

var Webhook = https.Function{
	RunWith: https.Options{
		AvailableMemoryMb: 256,
	},
	Callback: func(w https.ResponseWriter, r *https.Request) {
		fmt.Fprintf(w, "Hello, world!")
	},
}

var PubSubListener = pubsub.Function{
	RunWith: runwith.Options{
		MinInstances: 1,
	},
	Topic: "topic",
	Callback: func(ctx context.Context, event pubsub.Event) {
		fmt.Printf("Got event %+v", event)
	},
}

var NotAFunction = "Non-functions can be safely dumped to emulator.Serve to simplify code gen"

func NotACloudFunction(x int) {
	fmt.Println(x)
}
