package test

import (
	"fmt"
	"testing"
	"unsafe"
)

// case 2
type M interface {
	Show() string
}
type Mx struct{}

func (m *Mx) Show() string {
	return "Mx"
}

type My struct{}

func (m My) Show() string {
	return "My"
}

func TestInterface2(t *testing.T) {
	f := func(in []M) {
		for _, m := range in {
			fmt.Println(m.Show())
		}
	}
	// ①
	// ary := []M{Mx{}, My{}}
	// f(ary)
	// ②
	ary2 := []M{&Mx{}, &My{}}
	f(ary2)
}

// case 5

type G interface {
	String() string
}
type Gx struct{}

func (g *Gx) String() string {
	return "Gx"
}

type Gy struct{}

func (g Gy) String() string {
	return "Gy"
}
func Type(in interface{}) string {
	var typeStr string
	switch in.(type) {
	case int:
		typeStr = "int"
	case string:
		typeStr = "string"
	case float64:
		typeStr = "float64"
	case G:
		typeStr = "G"
	case Gx:
		typeStr = "Gx"
	default:
		typeStr = "unknown"

	}

	return typeStr
}

func TestInterface5(t *testing.T) {
	var ary []interface{} = []interface{}{
		1,
		"abc",
		2.3,
		Gx{},
		&Gx{},
	}
	f := func(ins []interface{}) {
		for idx, in := range ins {
			fmt.Println(idx, Type(in))
		}
	}
	f(ary)
}

// case 11

type In interface {
	Set(int)
	Get() int
}

// type声明的Int类型, 其属性值的内存是int存储格式
type Int int

func (i *Int) Set(a int) {
	*i = *(*Int)(unsafe.Pointer(&a))
}
func (i *Int) Get() int {
	return int(*i)
}

func TestInterface11(t *testing.T) {
	var i Int = 2
	f := func(in In) {
		// ① 无法更改属性值 in=2
		in.Set(123)
		fmt.Println(in.Get())
	}
	f(&i)
}
