package test

import (
	"fmt"
	"testing"
)

func TestMap1(t *testing.T) {
	var mp map[string]int = map[string]int{}
	mp["1"] = 1
	mp2 := make(map[string]int, 0)
	mp2["2"] = 2
	fmt.Println(mp2["abc"])
}
