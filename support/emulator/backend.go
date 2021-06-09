package emulator

import (
	"os"
)

type EventFilter struct {
	Attribute string `yaml:"attribute"`
	Value     string `yaml:"value"`
}

type EventTrigger struct {
	EventType           string        `yaml:"eventType"`
	EventFilters        []EventFilter `yaml:"eventFilters"`
	ServiceAccountEmail string        `yaml:"serviceAccountEmail,omitempty"`
}

type ApiVersion int

const GCFv2 ApiVersion = 2

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
	RequiredAPIs   map[string]string `yaml:"requiredAPIs"`
	CloudFunctions []FunctionSpec    `yaml:"cloudFunctions"`
	Topics         []PubSubSpec      `yaml:"topics"`
	Schedules      []ScheduleSpec    `yaml:"schedules"`
}

func ProjectOrDefault(project string) string {
	if project != "" {
		return project
	}
	return os.Getenv("GCLOUD_PROJECT")
}
