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
		AvailableMemoryMB: 256,
	},
	Callback: func(w https.ResponseWriter, r *https.Request) {
		fmt.Fprintf(w, "Hello, world!\n")
	},
}

var PubSubListener = pubsub.Function{
	RunWith: runwith.Options{
		MinInstances: 1,
	},
	EventType: pubsub.V1.Publish,
	Topic:     "topic",
	Callback: func(ctx context.Context, event pubsub.Event) {
		fmt.Printf("Got event %+v\n", event)
	},
}

var NotAFunction = "Non-functions can be safely dumped to emulator.Serve to simplify code gen"

// Is this better?
// var PubSubListner2 = pubsub.RunWith(runwith.Options{ MinInstances: 2 }).Topic("foo").OnPublish(func (ctx context.Context, event pubsub.Event) {
//
// })

func NotACloudFunction(x int) {
	fmt.Println(x)
}
