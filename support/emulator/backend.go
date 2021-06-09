package emulator

import (
	"os"
)

type EventFilters []EventFilter

func (f EventFilters) MarshalYAML() (interface{}, error) {
	m := make(map[string]interface{}, len(f))
	for _, filter := range f {
		m[filter.Attribute] = filter.Value
	}
	return m, nil
}

type EventFilter struct {
	Attribute string `yaml:"attribute"`
	Value     string `yaml:"value"`
}

type EventTrigger struct {
	EventType           string       `yaml:"eventType,omitempty"`
	EventFilters        EventFilters `yaml:"eventFilters,omitempty"`
	ServiceAccountEmail string       `yaml:"serviceAccountEmail,omitempty"`
}

type ApiVersion int

const GCFv2 ApiVersion = 2
const GCFv1 ApiVersion = 1

type FunctionSpec struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	EntryPoint string     `yaml:"entryPoint"`
	Id         string     `yaml:"id"`
	Region     string     `yaml:"region,omitempty"`
	Project    string     `yaml:"project,omitempty"`
	// NOTE: In the current schema this is a union between
	// an HTTP and an EventTrigger. Since HTTP triggers have
	// no options in GCFv2 we can just use an empty EventTrigger
	// for now.
	Trigger           EventTrigger `yaml:"trigger"`
	MinInstances      int          `yaml:"minInstances,omitempty"`
	MaxInstances      int          `yaml:"maxInstances,omitempty"`
	AvailableMemoryMB int          `yaml:"availableMemoryMb,omitempty"`
}

type TargetService struct {
	Id      string `yaml:"id"`
	Region  string `yaml:"region,omitempty"`
	Project string `yaml:"project,omitempty"`
}

type PubSubSpec struct {
	Id            string        `yaml:"id"`
	Project       string        `yaml:"project,omitempty"`
	TargetService TargetService `yaml:"targetService"`
}

type ScheduleRetryConfig struct {
	RetryCount int `yaml:"retryCount,omitempty"`
}

type Transport string

const PubSubTransport Transport = "pubsub"
const HttpsTransport Transport = "https"

type ScheduleSpec struct {
	Id            string              `yaml:"id"`
	Project       string              `yaml:"project"`
	Schedule      string              `yaml:"schedule"`
	TimeZone      string              `yaml:"timeZone,omitempty"`
	RetryConfig   ScheduleRetryConfig `yaml:"retryConfig"`
	Transport     Transport           `yaml:"transport"`
	TargetService TargetService       `yaml:"targetService"`
}

type Backend struct {
	RequiredAPIs   map[string]string `yaml:"requiredAPIs,omitempty"`
	CloudFunctions []FunctionSpec    `yaml:"cloudFunctions"`
	Topics         []PubSubSpec      `yaml:"topics,omitempty"`
	Schedules      []ScheduleSpec    `yaml:"schedules,omitempty"`
}

func ProjectOrDefault(project string) string {
	if project != "" {
		return project
	}
	return os.Getenv("GCLOUD_PROJECT")
}
