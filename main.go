package main

import "fmt"

type I interface {
	String() string
}

type IS struct{}

func (i *IS) String() string {
	return "ISSS"
}

var _ I = (*IS)(nil)

func main() {
	var i = (*IS)(nil)
	var j = IS{}
	fmt.Println(i, j)
}
