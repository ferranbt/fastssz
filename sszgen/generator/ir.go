package generator

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
	Size  uint64
	IsDyn bool // this is a fixed byte array but that is represented as a vector
}

func (b *Bytes) isValue() {}

type DynamicBytes struct {
	MaxSize uint64
}

func (d *DynamicBytes) isValue() {}

type BitList struct {
	Size uint64
}

func (b *BitList) isValue() {}

type Vector struct {
	Elem  Value2
	Size  uint64
	IsDyn bool // this is a fixed byte array but that is represented as a vector
}

func (v *Vector) isValue() {}

type List struct {
	Elem    Value2
	MaxSize uint64
}

func (l *List) isValue() {}

type Container struct {
	Elems map[string]*Value
}

func (c *Container) isValue() {}

type Time struct {
}

func (t *Time) isValue() {}
