// This is a sample of a minimum-complexity boilerplate code
// that must be generated in order to serve the emulator routes
// (GCF and Run) and production routes (Run only) along with
// the admin interface used for backend discovery.
package main

import (
	"github.com/inlined/go-functions/internal/emulator"
	alias "github.com/inlined/go-functions/sample/lib"
)

func main() {
	emulator.Serve(map[string]interface{}{
		"Webhook":           alias.Webhook,
		"PubSubListener":    alias.PubSubListener,
		"NotAFunction":      alias.NotAFunction,
		"NotACloudFunction": alias.NotAFunction,
	})
}
