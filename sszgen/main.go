package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ferranbt/fastssz/sszgen/generator"
	"github.com/ferranbt/fastssz/sszgen/version"
)

func main() {
	args := os.Args[1:]

	var cmd string
	if len(args) != 0 {
		cmd = args[0]
	}
	switch cmd {
	case "version":
		fmt.Println(version.Version)
	default:
		generate()
	}
}

func generate() {
	var source string
	var objsStr string
	var output string
	var include string
	var excludeObjs string
	var suffix string

	flag.StringVar(&source, "path", "", "")
	flag.StringVar(&objsStr, "objs", "", "")
	flag.StringVar(&excludeObjs, "exclude-objs", "", "Comma-separated list of types to exclude from output")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&include, "include", "", "")
	flag.StringVar(&suffix, "suffix", "encoding", "")

	flag.Parse()

	targets := decodeList(objsStr)
	includeList := decodeList(include)
	excludeTypeNames := make(map[string]bool)
	for _, name := range decodeList(excludeObjs) {
		excludeTypeNames[name] = true
	}

	if !strings.HasPrefix(suffix, "_") {
		suffix = fmt.Sprintf("_%s", suffix)
	}
	if !strings.HasSuffix(suffix, ".go") {
		suffix = fmt.Sprintf("%s.go", suffix)
	}

	out, err := generator.Encode(source, targets, output, includeList, excludeTypeNames, suffix)
	if err != nil {
		exit(err)
	}

	for name, str := range out {
		if err := ioutil.WriteFile(name, []byte(str), 0644); err != nil {
			exit(err)
		}
	}
}

func decodeList(input string) []string {
	if input == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSpace(input), ",")
}

func exit(err error) {
	fmt.Printf("[ERR]: %v\n", err)
	os.Exit(1)
}
