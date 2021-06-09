package pubsub

import (
	"fmt"
	"os"

	"github.com/inlined/go-functions/runwith"
	"github.com/inlined/go-functions/support/emulator"
)

type EventType string

var V1 = struct {
	Publish EventType
}{
	Publish: "google.cloud.pubsub.topic.v1.messagePublished",
}

type Function struct {
	EventType EventType
	Topic     string
	Region    string
	RunWith   runwith.Options
	Callback  interface{}
}

type Message struct {
	Data       interface{}       `json:"data"`
	Attributes map[string]string `json:"attributes"`
}

type Event struct {
	EventID string  `json:"eventId"`
	Data    Message `json:"data"`
}

func (p Function) AddBackendDescription(symbolName string, b *emulator.Backend) {
	// A builder pattern could ensure Topic is always present...
	if p.Topic == "" {
		panic(fmt.Sprintf("pubsub.Function %s is missing required parameteer Topic", symbolName))
	}
	if p.EventType == "" {
		p.EventType = V1.Publish
	}

	b.CloudFunctions = append(b.CloudFunctions, emulator.FunctionSpec{
		ApiVersion: emulator.GCFv2,
		EntryPoint: fmt.Sprintf("%s.%s", symbolName, "Callback"),
		Id:         symbolName,
		Region:     p.Region,
		Trigger: emulator.EventTrigger{
			EventType: string(p.EventType),
			EventFilters: []emulator.EventFilter{
				{
					Attribute: "resource",
					Value:     fmt.Sprintf("projects/%s/topics/%s", os.Getenv("GCLOUD_PROJECT"), p.Topic),
				},
			},
		},
		MinInstances:      p.RunWith.MinInstances,
		MaxInstances:      p.RunWith.MaxInstances,
		AvailableMemoryMB: p.RunWith.AvailableMemoryMB,
	})
}
