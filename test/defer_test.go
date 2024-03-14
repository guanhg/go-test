package test

import (
	"fmt"
	"testing"
)

func cacl(idx string, x, y int) int {
	ret := x + y
	fmt.Println(idx, x, y, ret)
	return ret
}

func TestDefer(t *testing.T) {
	a := 2
	b := 4
	defer cacl("1", a, cacl("15", a, b))
	a = 1
	defer cacl("2", a, cacl("30", a, b))
}
