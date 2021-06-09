package https

import (
	"fmt"
	"net/http"

	"github.com/inlined/go-functions/support/emulator"
)

type Request = http.Request
type ResponseWriter = http.ResponseWriter

type Options struct {
	MinInstances      int
	MaxInstances      int
	AvailableMemoryMB int
}

type Function struct {
	Callback func(ResponseWriter, *Request)
	RunWith  Options
}

func (h Function) AddBackendDescription(symbolName string, b *emulator.Backend) {
	// Runtime isn't specified from within the API?
	b.CloudFunctions = append(b.CloudFunctions, emulator.FunctionSpec{
		ApiVersion:        emulator.GCFv1,
		Id:                symbolName,
		EntryPoint:        fmt.Sprintf("%s.%s", symbolName, "Callback"),
		MinInstances:      h.RunWith.MinInstances,
		MaxInstances:      h.RunWith.MaxInstances,
		AvailableMemoryMB: h.RunWith.AvailableMemoryMB,
	})
}
