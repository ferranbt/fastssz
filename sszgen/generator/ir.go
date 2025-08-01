package generator

import (
	"fmt"
	"reflect"
)

// vector -> fixed size
// list -> variable size

type Value2 interface {
	isValue()
}

type Uint struct {
	Size uint64
}

func (u *Uint) isValue() {}

type Int struct {
	Size uint64
}

func (i *Int) isValue() {}

type Bool struct {
}

func (b *Bool) isValue() {}

type Bytes struct {
	Size    uint64
	IsList  bool
	IsGoDyn bool // this is a fixed byte array but that is represented as a vector
}

func (b *Bytes) IsFixed() bool {
	return !b.IsList && !b.IsGoDyn
}

func (b *Bytes) isValue() {}

type BitList struct {
	Size uint64
}

func (b *BitList) isValue() {}

type Vector struct {
	Elem  *Value
	Size  uint64
	IsDyn bool // this is a fixed byte array but that is represented as a vector
}

func (v *Vector) isValue() {}

type List struct {
	Elem    *Value
	MaxSize uint64
}

func (l *List) isValue() {}

type Container struct {
	Elems []*Value
}

func (c *Container) isValue() {}

type Time struct {
}

func (t *Time) isValue() {}

type Reference struct {
	Size uint64
}

func (r *Reference) isValue() {}

func getElem(v Value2) *Value {
	switch obj := v.(type) {
	case *List:
		return obj.Elem
	case *Vector:
		return obj.Elem
	default:
		panic(fmt.Errorf("getElem called with non-list/vector type %s", reflect.TypeOf(v)))
	}
}

func (v *Value) getObjs() []*Value {
	if obj, ok := v.typ.(*Container); ok {
		return obj.Elems
	}
	return nil
}

func (v *Value) Type() string {
	switch v.typ.(type) {
	case *Bool:
		return "bool"
	case *Uint:
		return "uint"
	case *Int:
		return "int"
	case *Bytes:
		return "bytes"
	case *BitList:
		return "bitlist"
	case *Vector:
		return "vector"
	case *List:
		return "list"
	case *Container:
		return "container"
	case *Reference:
		return "reference"
	case *Time:
		return "time"
	default:
		panic(fmt.Errorf("unknown type %s", reflect.TypeOf(v.typ)))
	}
}
