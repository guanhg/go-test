package test

import (
	"fmt"
	"testing"
)

// case 2
type I interface {
	String() string
}
type Imp struct{}

func (i Imp) String() string {
	return "Imp"
}

type A struct {
	Name string
	Age  int
	I
}
type AB struct {
	A
	Gender bool
}

type AC struct {
	*A
	Age int32
}

func TestStruct2(t *testing.T) {
	f1 := func(b AB, c AC, age int) {
		b.A.Age = age
		c.A.Age = age
	}

	i := Imp{}
	b := AB{Gender: true, A: A{Name: "R1", Age: 1, I: i}}
	c := AC{Age: 187, A: &A{Name: "R2", Age: 2}}
	f1(b, c, 3)

	fmt.Println(b.A.Age, c.A.Age, b.A.I.String())
}
