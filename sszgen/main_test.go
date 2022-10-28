package main

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/ferranbt/fastssz/sszgen/generator"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

type config struct {
	Source           string
	Targets          []string
	Output           string
	IncludePaths     []string
	ExcludeTypeNames []string
}

func TestGolden(t *testing.T) {
	tests := []config{
		{Source: "exclude_obj.go", ExcludeTypeNames: []string{"Bytes"}},
		{Source: "exclude_objs.go", ExcludeTypeNames: []string{"Case5Bytes", "Case5Roots"}},
		{Source: "anonymous_field.go"},
		{Source: "resolve-packages/single.go"},
		{Source: "resolve-packages/multiple.go", IncludePaths: []string{"resolve-packages/other", "resolve-packages/other2"}},
		{Source: "alias_array_size.go"},
		{Source: "time.go"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.Source, func(t *testing.T) {
			t.Parallel()
			runGoldenTest(t, test)
		})
	}
}

func runGoldenTest(t *testing.T, test config) {
	excludeTypeNames := make(map[string]bool)
	for _, name := range test.ExcludeTypeNames {
		excludeTypeNames[name] = true
	}
	out, err := generator.Encode(prefixPath(test.Source), test.Targets, test.Output, prefixPaths(test.IncludePaths), excludeTypeNames, "_encoding.go")
	if err != nil {
		t.Fatal(err)
	}
	for f, got := range out {
		want, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatalf("unable to load expected result: %v", err)
		}
		if got != string(want) {
			edits := myers.ComputeEdits(span.URIFromPath("got"), string(want), got)
			t.Fatalf("\n%s",gotextdiff.ToUnified("got", "want", string(want), edits))
		}
	}
}

func prefixPaths(s []string) (out []string) {
	out = make([]string, len(s))
	for i, s := range s {
		out[i] = prefixPath(s)
	}
	return out
}

func prefixPath(s string) string {
	return path.Join("testcases", s)
}
