package main

import (
	"errors"
	"fmt"
	"go/types"
	"log"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) != 1 {
		fmt.Println("Usage: go run deploy_functions.go <FunctionPackage>")
		os.Exit(1)
	}

	pkg := argsWithoutProg[0]
	triggers, err := extractTriggers(pkg)
	if err != nil {
		log.Fatalln(err)
	}

	code, err := generateEntrypoint(pkg, triggers)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(code)
}

type Trigger struct {
	Name string
	Type string
}

const httpTriggerType = "github.com/inlined/go-functions/https.Function"
const pubsubTriggerType = "github.com/inlined/go-functions/pubsub.Function"

var triggerTypes = []string{
	httpTriggerType,
	pubsubTriggerType,
}

func extractTriggers(pkg string) ([]Trigger, error) {
	exports, err := loadExports(pkg)
	if err != nil {
		return nil, err
	}

	var triggers []Trigger
	for _, exp := range exports {
		trig := Trigger{
			Name: exp.Name(),
			Type: exp.Type().String(),
		}
		triggers = append(triggers, trig)
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
		return nil, fmt.Errorf("one or mor errors")
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

const mainTemplate = `
package main

import (
	alias "{{ .Pkg }}"
	"github.com/inlined/go-functions/support/emulator"
)

func main() {
	emulator.Serve(map[string]interface{}{
	{{- range .Triggers }}
		"{{ .Name }}": alias.{{ .Name }},
	{{- end }}
	})
}
`

func generateEntrypoint(pkg string, triggers []Trigger) (string, error) {
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return "", err
	}

	b := new(strings.Builder)
	info := struct {
		Pkg      string
		Triggers []Trigger
	}{
		Pkg:      pkg,
		Triggers: triggers,
	}
	if err := tmpl.Execute(b, info); err != nil {
		return "", err
	}

	return b.String(), nil
}
