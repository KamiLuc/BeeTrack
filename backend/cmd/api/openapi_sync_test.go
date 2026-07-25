package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// muxHandleRoute matches mux.Handle("METHOD /path", ...) and
// mux.HandleFunc("METHOD /path", ...) registrations in main.go.
var muxHandleRoute = regexp.MustCompile(`mux\.Handle(?:Func)?\("([A-Z]+) ([^"]+)"`)

// TestOpenAPISpecMatchesRegisteredRoutes guards against exactly the drift
// found while writing docs/openapi.yaml (INF-05-BE): every route actually
// registered in main.go must appear in the OpenAPI spec, and vice versa.
// Regenerate the spec with `python docs/generate_openapi.py` after editing
// docs/api.md or adding/removing a route here.
func TestOpenAPISpecMatchesRegisteredRoutes(t *testing.T) {
	mainSrc, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	registered := map[string]bool{}
	for _, m := range muxHandleRoute.FindAllStringSubmatch(string(mainSrc), -1) {
		registered[m[1]+" "+m[2]] = true
	}
	if len(registered) == 0 {
		t.Fatal("found zero registered routes in main.go — regex is broken")
	}

	specPath := filepath.Join("..", "..", "..", "docs", "openapi.yaml")
	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}

	var spec struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(specBytes, &spec); err != nil {
		t.Fatalf("parse %s: %v", specPath, err)
	}

	documented := map[string]bool{}
	for path, methods := range spec.Paths {
		for method := range methods {
			documented[strings.ToUpper(method)+" "+path] = true
		}
	}
	if len(documented) == 0 {
		t.Fatal("found zero documented paths in openapi.yaml — is it empty?")
	}

	var missingFromSpec, missingFromCode []string
	for route := range registered {
		if !documented[route] {
			missingFromSpec = append(missingFromSpec, route)
		}
	}
	for route := range documented {
		if !registered[route] {
			missingFromCode = append(missingFromCode, route)
		}
	}
	sort.Strings(missingFromSpec)
	sort.Strings(missingFromCode)

	if len(missingFromSpec) > 0 {
		t.Errorf("routes registered in main.go but missing from docs/openapi.yaml "+
			"(add them to docs/api.md and re-run `python docs/generate_openapi.py`):\n%s",
			strings.Join(missingFromSpec, "\n"))
	}
	if len(missingFromCode) > 0 {
		t.Errorf("routes documented in docs/openapi.yaml but not registered in main.go "+
			"(stale doc entry — remove it from docs/api.md and re-run `python docs/generate_openapi.py`):\n%s",
			strings.Join(missingFromCode, "\n"))
	}
}
