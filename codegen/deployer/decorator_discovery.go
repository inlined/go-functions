package main

import (
	"errors"
	"fmt"
	"go/types"
	"log"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func main() {
	const pkg = "acme.com/example"
	triggers, err := extractTriggers(pkg)
	if err != nil {
		log.Fatal(err)
	}

	s, err := generateHeader(pkg, triggers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)

	s, err = generateRoutes(triggers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)

	s, err = generateMain(triggers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)
}

func extractTriggers(pkg string) ([]Trigger, error) {
	exports, err := loadExports(pkg)
	if err != nil {
		return nil, err
	}

	var triggers []Trigger
	for _, exp := range exports {
		typeName := exp.Type().String()
		var trigger Trigger
		if typeName == httpTriggerType {
			trigger = newHTTPTrigger(pkg, exp)
		} else if typeName == pubsubTriggerType {
			trigger = newPubSubTrigger(pkg, exp)
		} else {
			return nil, fmt.Errorf("Unsupported trigger type: %s", typeName)
		}
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

func loadExports(pkg string) ([]types.Object, error) {
	cfg := &packages.Config{Mode: packages.LoadTypes}
	pkgs, err := packages.Load(cfg, pkg)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, err
	}

	var results []types.Object
	scope := pkgs[0].Types.Scope()
	for _, symbol := range scope.Names() {
		typeInfo := scope.Lookup(symbol)
		if isTrigger(symbol, typeInfo) {
			results = append(results, typeInfo)
		}
	}

	if results == nil {
		return nil, errors.New("no exported member")
	}

	return results, nil
}

func isTrigger(name string, typeInfo types.Object) bool {
	if !typeInfo.Exported() {
		return false
	}

	typeName := typeInfo.Type().String()
	for _, tt := range triggerTypes {
		if tt == typeName {
			return true
		}
	}

	return false
}

const httpTriggerType = "func(net/http.ResponseWriter, *net/http.Request)"
const pubsubTriggerType = "func(ctx context.Context, data encoding/json.RawMessage)"

var triggerTypes = []string{
	httpTriggerType,
	pubsubTriggerType,
}

type Trigger interface {
	Name() string
	Discover() (string, error)
	Route() (string, error)
}

type HTTPTrigger struct {
	Pkg  string
	Func string
}

func newHTTPTrigger(pkg string, typeInfo types.Object) *HTTPTrigger {
	return &HTTPTrigger{
		Pkg:  pkg,
		Func: typeInfo.Name(),
	}
}

var httpDiscovery = `package main

import (
	"context"
	"net/http"

	alias "{{ .Pkg }}"
)

func main() {
	discover := alias.{{ .Func }}
	ctx := context.WithValue(context.Background(), "firebase_discovery", true)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example", nil)
	if err != nil {
		panic(err)
	}
	discover(req, nil)
}
`

var httpRoute = `
func {{ .Func }}(w http.ResponseWriter, r *http.Request) {
	alias.{{ .Func }}(w, r)
}
`

func (ht *HTTPTrigger) Name() string {
	return ht.Func
}

func (ht *HTTPTrigger) Discover() (string, error) {
	tmpl, err := template.New("http_discovery").Parse(httpDiscovery)
	if err != nil {
		return "", err
	}

	b := new(strings.Builder)
	if err := tmpl.Execute(b, ht); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (ht *HTTPTrigger) Route() (string, error) {
	tmpl, err := template.New("http_route").Parse(httpRoute)
	if err != nil {
		return "", err
	}

	b := new(strings.Builder)
	if err := tmpl.Execute(b, ht); err != nil {
		return "", err
	}

	return b.String(), nil
}

type PubSubTrigger struct {
	Pkg  string
	Func string
}

func newPubSubTrigger(pkg string, typeInfo types.Object) *PubSubTrigger {
	return &PubSubTrigger{
		Pkg:  pkg,
		Func: typeInfo.Name(),
	}
}

var pubsubDiscovery = `package main

import (
	"context"
	"fmt"

	alias "{{ .Pkg }}"
)

func main() {
	discover := alias.{{ .Func }}
	ctx := context.WithValue(context.Background(), "firebase_discovery", true)
	discover(ctx, nil)
}
`

var pubsubRoute = `
func {{ .Func }}(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var data json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		panic(err)
	}

	alias.{{ .Func }}(r.Context(), data)
	w.WriteHeader(http.StatusOK)
}
`

func (ht *PubSubTrigger) Name() string {
	return ht.Func
}

func (pt *PubSubTrigger) Discover() (string, error) {
	tmpl, err := template.New("pubsub_discovery").Parse(pubsubDiscovery)
	if err != nil {
		return "", err
	}

	b := new(strings.Builder)
	if err := tmpl.Execute(b, pt); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (ht *PubSubTrigger) Route() (string, error) {
	tmpl, err := template.New("pubsub_route").Parse(pubsubRoute)
	if err != nil {
		return "", err
	}

	b := new(strings.Builder)
	if err := tmpl.Execute(b, ht); err != nil {
		return "", err
	}

	return b.String(), nil
}

var header = `
package main

import (
{{range .Imports}}	"{{.}}"
{{ end }}
	alias "{{ .Pkg }}"
)
`

type HeaderInfo struct {
	Pkg     string
	Imports []string
}

func generateHeader(pkg string, triggers []Trigger) (string, error) {
	imports := []string{
		"log",
		"net/http",
	}

	pubsubAvailable := false
	for _, trig := range triggers {
		if _, ok := trig.(*PubSubTrigger); ok {
			pubsubAvailable = true
			break
		}
	}

	if pubsubAvailable {
		imports = append(imports, "io/ioutil")
		imports = append(imports, "encoding/json")
	}

	sort.Strings(imports)
	headerInfo := HeaderInfo{
		Pkg:     pkg,
		Imports: imports,
	}

	tmpl, err := template.New("header").Parse(header)
	if err != nil {
		return "", err
	}

	b := new(strings.Builder)
	if err := tmpl.Execute(b, headerInfo); err != nil {
		return "", err
	}

	return b.String(), nil
}

func generateRoutes(triggers []Trigger) (string, error) {
	var routes []string
	for _, trig := range triggers {
		r, err := trig.Route()
		if err != nil {
			return "", err
		}

		routes = append(routes, r)
	}

	return strings.Join(routes, "\n"), nil
}

var mainTemplate = `
func main() {
	mux := http.NewServeMux()
{{ range .Names }}	mux.HandleFunc("/{{.}}", {{.}})
{{ end }}
	log.Println("Listening on port: 8080")
{{ range .Names }}	log.Println("Function {{.}} exposed at: /{{.}}")
{{ end }}
	http.ListenAndServe(":8080", mux)
}
`

type MainInfo struct {
	Names []string
}

func generateMain(triggers []Trigger) (string, error) {
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return "", err
	}

	var names []string
	for _, trig := range triggers {
		names = append(names, trig.Name())
	}
	mainInfo := &MainInfo{
		Names: names,
	}

	b := new(strings.Builder)
	if err := tmpl.Execute(b, mainInfo); err != nil {
		return "", err
	}

	return b.String(), nil
}
