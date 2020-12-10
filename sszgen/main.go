package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

const bytesPerLengthOffset = 4

func main() {
	var source string
	var objsStr string
	var output string
	var include string

	flag.StringVar(&source, "path", "", "")
	flag.StringVar(&objsStr, "objs", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&include, "include", "", "")

	flag.Parse()

	targets := decodeList(objsStr)
	includeList := decodeList(include)

	if err := encode(source, targets, output, includeList); err != nil {
		fmt.Printf("[ERR]: %v", err)
	}
}

func decodeList(input string) []string {
	if input == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSpace(input), ",")
}

// The SSZ code generation works in three steps:
// 1. Parse the Go input with the go/parser library to generate an AST representation.
// 2. Convert the AST into an Internal Representation (IR) to describe the structs and fields
// using the Value object.
// 3. Use the IR to print the encoding functions

func encode(source string, targets []string, output string, includePaths []string) error {
	files, err := parseInput(source) // 1.
	if err != nil {
		return err
	}

	// parse all the include paths as well
	include := map[string]*ast.File{}
	for _, i := range includePaths {
		files, err := parseInput(i)
		if err != nil {
			return err
		}
		for k, v := range files {
			include[k] = v
		}
	}

	// read package
	var packName string
	for _, file := range files {
		packName = file.Name.Name
	}

	e := &env{
		include:  include,
		source:   source,
		files:    files,
		objs:     map[string]*Value{},
		packName: packName,
		targets:  targets,
	}

	if err := e.generateIR(); err != nil { // 2.
		return err
	}

	// 3.
	var out map[string]string
	if output == "" {
		out, err = e.generateEncodings()
	} else {
		// output to a specific path
		out, err = e.generateOutputEncodings(output)
	}
	if err != nil {
		panic(err)
	}
	if out == nil {
		// empty output
		panic("No files to generate")
	}

	for name, str := range out {
		output := []byte(str)

		output, err = format.Source(output)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(name, output, 0644); err != nil {
			return err
		}
	}
	return nil
}

func isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func parseInput(source string) (map[string]*ast.File, error) {
	files := map[string]*ast.File{}

	ok, err := isDir(source)
	if err != nil {
		return nil, err
	}
	if ok {
		// dir
		astFiles, err := parser.ParseDir(token.NewFileSet(), source, nil, parser.AllErrors)
		if err != nil {
			return nil, err
		}
		for _, v := range astFiles {
			if !strings.HasSuffix(v.Name, "_test") {
				files = v.Files
			}
		}
	} else {
		// single file
		astfile, err := parser.ParseFile(token.NewFileSet(), source, nil, parser.AllErrors)
		if err != nil {
			return nil, err
		}
		files[source] = astfile
	}
	return files, nil
}

// Value is a type that represents a Go field or struct and his
// correspondent SSZ type.
type Value struct {
	// name of the variable this value represents
	name string
	// name of the Go object this value represents
	obj string
	// n is the fixed size of the value
	n uint64
	// auxiliary int number
	s uint64
	// type of the value
	t Type
	// array of values for a container
	o []*Value
	// type of item for an array
	e *Value
	// auxiliary boolean
	c bool
	// another auxiliary int number
	m uint64
	// ref is the external reference if the struct is imported
	// from another package
	ref string
	// new determines if the value is a pointer
	noPtr bool
}

func (v *Value) isListElem() bool {
	return strings.HasSuffix(v.name, "]")
}

func (v *Value) objRef() string {
	// global reference of the object including the package if the reference
	// is from an external package
	if v.ref == "" {
		return v.obj
	}
	return v.ref + "." + v.obj
}

func (v *Value) copy() *Value {
	vv := new(Value)
	*vv = *v
	vv.o = make([]*Value, len(v.o))
	for indx := range v.o {
		vv.o[indx] = v.o[indx].copy()
	}
	if v.e != nil {
		vv.e = v.e.copy()
	}
	return vv
}

// Type is a SSZ type
type Type int

const (
	// TypeUint is a SSZ int type
	TypeUint Type = iota
	// TypeBool is a SSZ bool type
	TypeBool
	// TypeBytes is a SSZ fixed or dynamic bytes type
	TypeBytes
	// TypeBitVector is a SSZ bitvector
	TypeBitVector
	// TypeBitList is a SSZ bitlist
	TypeBitList
	// TypeVector is a SSZ vector
	TypeVector
	// TypeList is a SSZ list
	TypeList
	// TypeContainer is a SSZ container
	TypeContainer
	// TypeReference is a SSZ reference
	TypeReference
)

func (t Type) String() string {
	switch t {
	case TypeUint:
		return "uint"
	case TypeBool:
		return "bool"
	case TypeBytes:
		return "bytes"
	case TypeBitVector:
		return "bitvector"
	case TypeBitList:
		return "bitlist"
	case TypeVector:
		return "vector"
	case TypeList:
		return "list"
	case TypeContainer:
		return "container"
	case TypeReference:
		return "reference"
	default:
		panic("not found")
	}
}

type env struct {
	source string
	// map of the include path for cross package reference
	include map[string]*ast.File
	// map of files with their Go AST format
	files map[string]*ast.File
	// name of the package
	packName string
	// map of structs with their Go AST format
	raw map[string]*astStruct
	// map of structs with their IR format
	objs map[string]*Value
	// map of files with their structs in order
	order map[string][]string
	// target structures to encode
	targets []string
	// imports in all the parsed packages
	imports []*astImport
}

const encodingPrefix = "_encoding.go"

func (e *env) generateOutputEncodings(output string) (map[string]string, error) {
	out := map[string]string{}

	orders := []string{}
	for _, order := range e.order {
		orders = append(orders, order...)
	}

	res, ok, err := e.print(true, orders)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	out[output] = res
	return out, nil
}

func (e *env) generateEncodings() (map[string]string, error) {
	outs := map[string]string{}

	firstDone := true
	for name, order := range e.order {
		// remove .go prefix and replace if with our own
		ext := filepath.Ext(name)
		name = strings.TrimSuffix(name, ext)
		name += encodingPrefix

		vvv, ok, err := e.print(firstDone, order)
		if err != nil {
			return nil, err
		}
		if ok {
			firstDone = false
			outs[name] = vvv
		}
	}
	return outs, nil
}

func (e *env) print(first bool, order []string) (string, bool, error) {
	tmpl := `// Code generated by fastssz. DO NOT EDIT.
	package {{.package}}
	
	import (
		ssz "github.com/ferranbt/fastssz" {{ if .imports }}{{ range $value := .imports }}
			{{ $value }} {{ end }}
		{{ end }}
	)

	{{ range .objs }}
		{{ .Marshal }}
		{{ .Unmarshal }}
		{{ .Size }}
		{{ .HashTreeRoot }}
	{{ end }}
	`

	data := map[string]interface{}{
		"package": e.packName,
	}

	type Obj struct {
		Size, Marshal, Unmarshal, HashTreeRoot string
	}

	objs := []*Obj{}
	imports := []string{}

	// Print the objects in the order in which they appear on the file.
	for _, name := range order {
		obj, ok := e.objs[name]
		if !ok {
			continue
		}

		// detect the imports required to unmarshal this objects
		refs := detectImports(obj)
		imports = appendWithoutRepeated(imports, refs)

		if obj.isFixed() && isBasicType(obj) {
			// we have an alias of a basic type (uint, bool). These objects
			// will be encoded/decoded inside their parent container and do not
			// require the sszgen functions.
			continue
		}
		objs = append(objs, &Obj{
			HashTreeRoot: e.hashTreeRoot(name, obj),
			Marshal:      e.marshal(name, obj),
			Unmarshal:    e.unmarshal(name, obj),
			Size:         e.size(name, obj),
		})
	}
	if len(objs) == 0 {
		// No valid objects found for this file
		return "", false, nil
	}
	data["objs"] = objs

	// insert any required imports
	importsStr, err := e.buildImports(imports)
	if err != nil {
		return "", false, err
	}
	if len(importsStr) != 0 {
		data["imports"] = importsStr
	}

	return execTmpl(tmpl, data), true, nil
}

func isBasicType(v *Value) bool {
	return v.t == TypeUint || v.t == TypeBool || v.t == TypeBytes
}

func (e *env) buildImports(imports []string) ([]string, error) {
	res := []string{}
	for _, i := range imports {
		imp := e.findImport(i)
		if imp != "" {
			res = append(res, imp)
		}
	}
	return res, nil
}

func (e *env) findImport(name string) string {
	for _, i := range e.imports {
		if i.match(name) {
			return i.getFullName()
		}
	}
	return ""
}

func appendWithoutRepeated(s []string, i []string) []string {
	for _, j := range i {
		if !contains(j, s) {
			s = append(s, j)
		}
	}
	return s
}

func detectImports(v *Value) []string {
	// for sure v is a container
	// check if any of the fields in the container has an import
	refs := []string{}
	for _, i := range v.o {
		var ref string
		switch i.t {
		case TypeReference:
			if !i.noPtr {
				// it is not a typed reference
				ref = i.ref
			}
		case TypeContainer:
			ref = i.ref
		case TypeList, TypeVector:
			ref = i.e.ref
		default:
			ref = i.ref
		}
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

// All the generated functions use the '::' string to represent the pointer receiver
// of the struct method (i.e 'm' in func(m *Method) XX()) for convenience.
// This function replaces the '::' string with a valid one that corresponds
// to the first letter of the method in lower case.
func appendObjSignature(str string, v *Value) string {
	sig := strings.ToLower(string(v.name[0]))
	return strings.Replace(str, "::", sig, -1)
}

type astStruct struct {
	name     string
	obj      *ast.StructType
	typ      ast.Expr
	implFunc bool
	isRef    bool
}

type astResult struct {
	objs  []*astStruct
	funcs []string
}

func decodeASTStruct(file *ast.File) *astResult {
	res := &astResult{
		objs:  []*astStruct{},
		funcs: []string{},
	}

	funcRefs := map[string]int{}
	for _, dec := range file.Decls {
		if genDecl, ok := dec.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					obj := &astStruct{
						name: typeSpec.Name.Name,
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if ok {
						obj.obj = structType
					} else {
						obj.typ = typeSpec.Type
					}
					res.objs = append(res.objs, obj)
				}
			}
		}
		if funcDecl, ok := dec.(*ast.FuncDecl); ok {
			if funcDecl.Recv == nil {
				continue
			}
			if expr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
				// only allow pointer functions
				if i, ok := expr.X.(*ast.Ident); ok {
					objName := i.Name
					if ok := isFuncDecl(funcDecl); ok {
						funcRefs[objName]++
					}
				}
			}
		}
	}
	for name, count := range funcRefs {
		if count == 4 {
			// it implements all the interface functions
			res.funcs = append(res.funcs, name)
		}
	}
	return res
}

func isSpecificFunc(funcDecl *ast.FuncDecl, in, out []string) bool {
	check := func(types *ast.FieldList, args []string) bool {
		list := types.List
		if len(list) != len(args) {
			return false
		}

		for i := 0; i < len(list); i++ {
			typ := list[i].Type
			arg := args[i]

			var buf bytes.Buffer
			fset := token.NewFileSet()
			if err := format.Node(&buf, fset, typ); err != nil {
				panic(err)
			}
			if string(buf.Bytes()) != arg {
				return false
			}
		}

		return true
	}
	if !check(funcDecl.Type.Params, in) {
		return false
	}
	if !check(funcDecl.Type.Results, out) {
		return false
	}
	return true
}

func isFuncDecl(funcDecl *ast.FuncDecl) bool {
	name := funcDecl.Name.Name
	if name == "SizeSSZ" {
		return isSpecificFunc(funcDecl, []string{}, []string{"int"})
	}
	if name == "MarshalSSZTo" {
		return isSpecificFunc(funcDecl, []string{"[]byte"}, []string{"[]byte", "error"})
	}
	if name == "UnmarshalSSZ" {
		return isSpecificFunc(funcDecl, []string{"[]byte"}, []string{"error"})
	}
	if name == "HashTreeRootWith" {
		return isSpecificFunc(funcDecl, []string{"*ssz.Hasher"}, []string{"error"})
	}
	return false
}

type astImport struct {
	alias string
	path  string
}

func (a *astImport) getFullName() string {
	if a.alias != "" {
		return fmt.Sprintf("%s \"%s\"", a.alias, a.path)
	}
	return fmt.Sprintf("\"%s\"", a.path)
}

func (a *astImport) match(name string) bool {
	if a.alias != "" {
		return a.alias == name
	}
	return filepath.Base(a.path) == name
}

func trimQuotes(a string) string {
	return strings.Trim(a, "\"")
}

func decodeASTImports(file *ast.File) []*astImport {
	imports := []*astImport{}
	for _, i := range file.Imports {
		var alias string
		if i.Name != nil {
			if i.Name.Name == "_" {
				continue
			}
			alias = i.Name.Name
		}
		path := trimQuotes(i.Path.Value)
		imports = append(imports, &astImport{
			alias: alias,
			path:  path,
		})
	}
	return imports
}

func (e *env) generateIR() error {
	e.raw = map[string]*astStruct{}
	e.order = map[string][]string{}
	e.imports = []*astImport{}

	// we want to make sure we only include one reference for each struct name
	// among the source and include paths.
	addStructs := func(res *astResult, isRef bool) error {
		for _, i := range res.objs {
			if _, ok := e.raw[i.name]; ok {
				return fmt.Errorf("two structs share the same name %s", i.name)
			}
			i.isRef = isRef
			e.raw[i.name] = i
		}
		// include all the functions that implement the interfaces
		for _, name := range res.funcs {
			v, ok := e.raw[name]
			if !ok {
				return fmt.Errorf("cannot find %s struct", name)
			}
			v.implFunc = true
		}
		return nil
	}

	// add the imports to the environment, we want to make sure that we always import
	// the package with the same name and alias which is easier to logic with.
	addImports := func(imports []*astImport) error {
		for _, i := range imports {
			// check if we already have this import before
			found := false
			for _, j := range e.imports {
				if j.path == i.path {
					found = true
					if i.alias != j.alias {
						return fmt.Errorf("the same package is imported twice by different files of path %s and %s with different aliases: %s and %s", j.path, i.path, j.alias, i .alias)
					}
				}
			}
			if !found {
				e.imports = append(e.imports, i)
			}
		}
		return nil
	}

	// decode all the imports from the input files
	for _, file := range e.files {
		if err := addImports(decodeASTImports(file)); err != nil {
			return err
		}
	}

	// decode the structs from the input path
	for name, file := range e.files {
		res := decodeASTStruct(file)
		if err := addStructs(res, false); err != nil {
			return err
		}

		// keep the ordering in which the structs appear so that we always generate them in
		// the same predictable order
		structOrdering := []string{}
		for _, i := range res.objs {
			structOrdering = append(structOrdering, i.name)
		}
		e.order[name] = structOrdering
	}

	// decode the structs from the include path but ONLY include them on 'raw' not in 'order'.
	// If the structs are in raw they can be used as a reference at compilation time and since they are
	// not in 'order' they cannot be used to marshal/unmarshal encodings
	for _, file := range e.include {
		res := decodeASTStruct(file)
		if err := addStructs(res, true); err != nil {
			return err
		}
	}

	for name, obj := range e.raw {
		var valid bool
		if e.targets == nil || len(e.targets) == 0 {
			valid = true
		} else {
			valid = contains(name, e.targets)
		}
		if valid {
			if obj.isRef {
				// do not process imported elements
				continue
			}
			if _, err := e.encodeItem(name, ""); err != nil {
				return err
			}
		}
	}
	return nil
}

func contains(i string, j []string) bool {
	for _, a := range j {
		if a == i {
			return true
		}
	}
	return false
}

func (e *env) encodeItem(name, tags string) (*Value, error) {
	v, ok := e.objs[name]
	if !ok {
		var err error
		raw, ok := e.raw[name]
		if !ok {
			return nil, fmt.Errorf("could not find struct with name '%s'", name)
		}
		if raw.implFunc {
			size, _ := getTagsInt(tags, "ssz-size")
			v = &Value{t: TypeReference, s: size, n: size, noPtr: raw.obj == nil}
		} else if raw.obj != nil {
			v, err = e.parseASTStructType(name, raw.obj)
		} else {
			v, err = e.parseASTFieldType(name, tags, raw.typ)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to encode %s: %v", name, err)
		}
		v.name = name
		v.obj = name
		e.objs[name] = v
	}
	return v.copy(), nil
}

// parse the Go AST struct
func (e *env) parseASTStructType(name string, typ *ast.StructType) (*Value, error) {
	v := &Value{
		name: name,
		t:    TypeContainer,
		o:    []*Value{},
	}

	for _, f := range typ.Fields.List {
		if len(f.Names) != 1 {
			continue
		}
		name := f.Names[0].Name
		if !isExportedField(name) {
			continue
		}
		if strings.HasPrefix(name, "XXX_") {
			// skip protobuf methods
			continue
		}
		var tags string
		if f.Tag != nil {
			tags = f.Tag.Value
		}

		elem, err := e.parseASTFieldType(name, tags, f.Type)
		if err != nil {
			return nil, err
		}
		if elem == nil {
			continue
		}
		elem.name = name
		v.o = append(v.o, elem)
	}

	// get the total size of the container
	for _, f := range v.o {
		if f.isFixed() {
			v.n += f.n
		} else {
			v.n += bytesPerLengthOffset
			// container is dynamic
			v.c = true
		}
	}
	return v, nil
}

func getObjLen(obj *ast.ArrayType) uint64 {
	if obj.Len == nil {
		return 0
	}
	value := obj.Len.(*ast.BasicLit).Value
	num, err := strconv.ParseUint(value, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("BUG: Failed to convert to uint64 %s: %v", value, err))
	}
	return num
}

// parse the Go AST field
func (e *env) parseASTFieldType(name, tags string, expr ast.Expr) (*Value, error) {
	if tag, ok := getTags(tags, "ssz"); ok && tag == "-" {
		// omit value
		return nil, nil
	}

	switch obj := expr.(type) {
	case *ast.StarExpr:
		// *Struct
		switch elem := obj.X.(type) {
		case *ast.Ident:
			// reference to a local package
			return e.encodeItem(elem.Name, tags)

		case *ast.SelectorExpr:
			// reference of the external package
			ref := elem.X.(*ast.Ident).Name
			// reference to a struct from another package
			v, err := e.encodeItem(elem.Sel.Name, tags)
			if err != nil {
				return nil, err
			}
			v.ref = ref
			return v, nil

		default:
			return nil, fmt.Errorf("cannot handle %s", elem)
		}

	case *ast.ArrayType:
		if isByte(obj.Elt) {
			if fixedlen := getObjLen(obj); fixedlen != 0 {
				// array of fixed size
				return &Value{t: TypeBytes, c: true, s: fixedlen, n: fixedlen}, nil
			}
			// []byte
			if tag, ok := getTags(tags, "ssz"); ok && tag == "bitlist" {
				// bitlist requires a ssz-max field
				max, ok := getTagsInt(tags, "ssz-max")
				if !ok {
					return nil, fmt.Errorf("bitfield requires a 'ssz-max' field")
				}
				return &Value{t: TypeBitList, m: max, s: max}, nil
			}
			size, ok := getTagsInt(tags, "ssz-size")
			if ok {
				// fixed bytes
				return &Value{t: TypeBytes, s: size, n: size}, nil
			}
			max, ok := getTagsInt(tags, "ssz-max")
			if !ok {
				return nil, fmt.Errorf("[]byte expects either ssz-max or ssz-size")
			}
			// dynamic bytes
			return &Value{t: TypeBytes, m: max}, nil
		}
		if isArray(obj.Elt) && isByte(obj.Elt.(*ast.ArrayType).Elt) {
			f, fCheck, s, sCheck, t, err := getRootSizes(obj, tags)
			if err != nil {
				return nil, err
			}
			if t == TypeVector {
				// vector
				return &Value{t: TypeVector, c: fCheck, n: f * s, s: f, e: &Value{t: TypeBytes, c: sCheck, n: s, s: s}}, nil
			}
			// list
			return &Value{t: TypeList, s: f, e: &Value{t: TypeBytes, c: sCheck, n: s, s: s}}, nil
		}

		// []*Struct
		elem, err := e.parseASTFieldType(name, tags, obj.Elt)
		if err != nil {
			return nil, err
		}
		if size, ok := getTagsInt(tags, "ssz-size"); ok {
			// fixed vector
			v := &Value{t: TypeVector, s: size, e: elem}
			if elem.isFixed() {
				// set the total size
				v.n = size * elem.n
			}
			return v, err
		}
		// list
		maxSize, ok := getTagsInt(tags, "ssz-max")
		if !ok {
			return nil, fmt.Errorf("slice '%s' expects either ssz-max or ssz-size", name)
		}
		v := &Value{t: TypeList, e: elem, s: maxSize, m: maxSize}
		return v, nil

	case *ast.Ident:
		// basic type
		var v *Value
		switch obj.Name {
		case "uint64":
			v = &Value{t: TypeUint, n: 8}
		case "uint32":
			v = &Value{t: TypeUint, n: 4}
		case "uint16":
			v = &Value{t: TypeUint, n: 2}
		case "uint8":
			v = &Value{t: TypeUint, n: 1}
		case "bool":
			v = &Value{t: TypeBool, n: 1}
		default:
			// try to resolve as an alias
			vv, err := e.encodeItem(obj.Name, tags)
			if err != nil {
				return nil, fmt.Errorf("type %s not found", obj.Name)
			}
			return vv, nil
		}
		return v, nil

	case *ast.SelectorExpr:
		name := obj.X.(*ast.Ident).Name
		sel := obj.Sel.Name

		if sel == "Bitlist" {
			// go-bitfield/Bitlist
			maxSize, ok := getTagsInt(tags, "ssz-max")
			if !ok {
				return nil, fmt.Errorf("bitlist %s does not have ssz-max tag", name)
			}
			return &Value{t: TypeBitList, m: maxSize, s: maxSize}, nil
		} else if strings.HasPrefix(sel, "Bitvector") {
			// go-bitfield/Bitvector, fixed bytes
			size, ok := getTagsInt(tags, "ssz-size")
			if !ok {
				return nil, fmt.Errorf("bitvector %s does not have ssz-size tag", name)
			}
			return &Value{t: TypeBytes, s: size, n: size}, nil
		}
		// external reference
		vv, err := e.encodeItem(sel, tags)
		if err != nil {
			return nil, err
		}
		vv.ref = name
		vv.noPtr = true
		return vv, nil

	default:
		panic(fmt.Errorf("ast type '%s' not expected", reflect.TypeOf(expr)))
	}
}

func getRootSizes(obj *ast.ArrayType, tags string) (f uint64, fCheck bool, s uint64, sCheck bool, t Type, err error) {

	// check if we are in an array and we get the sizes from there
	f = getObjLen(obj)
	s = getObjLen(obj.Elt.(*ast.ArrayType))
	t = TypeVector

	if f != 0 {
		fCheck = true
	}
	if s != 0 {
		sCheck = true
	}

	if f != 0 && s != 0 {
		// all the sizes are set as arrays
		return
	}

	if f != s {
		// one of the values was not set as an array
		// check 'ssz-size' for vector or 'ssz-max' for a list
		size, ok := getTagsInt(tags, "ssz-size")
		if !ok {
			t = TypeList
			size, ok = getTagsInt(tags, "ssz-max")
			if !ok {
				err = fmt.Errorf("bad")
				return
			}
		}

		// fill the missing size
		if f == 0 {
			f = size
		} else {
			s = size
		}
		return
	}

	// Neither of the values was set as an array, we need
	// to get both sizes with the go tags
	var ok bool
	f, s, ok = getTagsTuple(tags, "ssz-size")
	if !ok {
		err = fmt.Errorf("[][]byte expects a ssz-size tag")
		return
	}
	if f == 0 {
		t = TypeList
		f, ok = getTagsInt(tags, "ssz-max")
		if !ok {
			err = fmt.Errorf("ssz-max not set after '?' field on ssz-size")
			return
		}
	}
	return
}

func isArray(obj ast.Expr) bool {
	_, ok := obj.(*ast.ArrayType)
	return ok
}

func isByte(obj ast.Expr) bool {
	if ident, ok := obj.(*ast.Ident); ok {
		if ident.Name == "byte" {
			return true
		}
	}
	return false
}

func isExportedField(str string) bool {
	return str[0] <= 90
}

// getTagsTuple decodes tags of the format 'ssz-size:"33,32"'. If the
// first value is '?' it returns -1.
func getTagsTuple(str string, field string) (uint64, uint64, bool) {
	tupleStr, ok := getTags(str, field)
	if !ok {
		return 0, 0, false
	}

	spl := strings.Split(tupleStr, ",")
	if len(spl) != 2 {
		return 0, 0, false
	}

	// first can be either ? or a number
	var first uint64
	if spl[0] == "?" {
		first = 0
	} else {
		tmp, err := strconv.Atoi(spl[0])
		if err != nil {
			return 0, 0, false
		}
		first = uint64(tmp)
	}

	second, err := strconv.Atoi(spl[1])
	if err != nil {
		return 0, 0, false
	}
	return first, uint64(second), true
}

// getTagsInt returns tags of the format 'ssz-size:"32"'
func getTagsInt(str string, field string) (uint64, bool) {
	numStr, ok := getTags(str, field)
	if !ok {
		return 0, false
	}
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, false
	}
	return uint64(num), true
}

// getTags returns the tags from a given field
func getTags(str string, field string) (string, bool) {
	str = strings.Trim(str, "`")

	for _, tag := range strings.Split(str, " ") {
		if !strings.Contains(tag, ":") {
			return "", false
		}
		spl := strings.Split(tag, ":")
		if len(spl) != 2 {
			return "", false
		}

		tagName, vals := spl[0], spl[1]
		if !strings.HasPrefix(vals, "\"") || !strings.HasSuffix(vals, "\"") {
			return "", false
		}
		if tagName != field {
			continue
		}

		vals = strings.Trim(vals, "\"")
		return vals, true
	}
	return "", false
}

func (v *Value) isFixed() bool {
	switch v.t {
	case TypeVector:
		return v.e.isFixed()

	case TypeBytes:
		if v.s != 0 {
			// fixed bytes
			return true
		}
		// dynamic bytes
		return false

	case TypeContainer:
		return !v.c

	// Dynamic types
	case TypeBitList:
		fallthrough
	case TypeList:
		return false

	// Fixed types
	case TypeBitVector:
		fallthrough
	case TypeUint:
		fallthrough
	case TypeBool:
		return true

	case TypeReference:
		if v.s != 0 {
			return true
		}
		return false

	default:
		panic(fmt.Errorf("is fixed not implemented for type %s", v.t.String()))
	}
}

func execTmpl(tpl string, input interface{}) string {
	tmpl, err := template.New("tmpl").Parse(tpl)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, input); err != nil {
		panic(err)
	}
	return buf.String()
}

func uintVToName(v *Value) string {
	if v.t != TypeUint {
		panic("not expected")
	}
	switch v.n {
	case 8:
		return "Uint64"
	case 4:
		return "Uint32"
	case 2:
		return "Uint16"
	case 1:
		return "Uint8"
	default:
		panic("not found")
	}
}
