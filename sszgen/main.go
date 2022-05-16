package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ferranbt/fastssz/sszgen/generator"
)

func main() {
	var source string
	var objsStr string
	var output string
	var include string
	var experimental bool
	var excludeObjs string

	flag.StringVar(&source, "path", "", "")
	flag.StringVar(&objsStr, "objs", "", "")
	flag.StringVar(&excludeObjs, "exclude-objs", "", "Comma-separated list of types to exclude from output")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&include, "include", "", "")
	flag.BoolVar(&experimental, "experimental", false, "")

	flag.Parse()

	targets := decodeList(objsStr)
	includeList := decodeList(include)
	excludeTypeNames := make(map[string]bool)
	for _, name := range decodeList(excludeObjs) {
		excludeTypeNames[name] = true
	}

	if err := generator.Encode(source, targets, output, includeList, excludeTypeNames, experimental); err != nil {
		fmt.Printf("[ERR]: %v\n", err)
		os.Exit(1)
	}
}

func decodeList(input string) []string {
	if input == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSpace(input), ",")
}
