package test

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

// case1
func TestUnsafe1(t *testing.T) {
	x := []int{1, 2, 3, 4}
	e3 := *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&x[0])) + 2*unsafe.Sizeof(x[0])))
	fmt.Println(e3) //  输出 3
}

// case3

func TestUnsafe3(t *testing.T) {
	// var f int16 = 1
	var i int = 2
	fp := (*float32)(unsafe.Pointer(&i))
	fmt.Println(fp, &i)
	i = 4
	fmt.Println(*fp, i)

	ip := (*int32)(unsafe.Pointer(&i))
	i = 2e4
	fmt.Println(*ip, i)

	// f := *(*float32)(&i)
}

// case4

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func TestUnsafe4(t *testing.T) {
	b := []byte{'a', 'b', 'c'}
	s := "def"
	s1 := BytesToString(b)
	b1 := StringToBytes(s)

	fmt.Println(&s1, &b1[0], &b[0], &s)
}

type MutexAlias struct {
	sync.Mutex
	A int
	B string
}

func TestUnsafe5(t *testing.T) {
	var loc sync.Mutex
	var ma MutexAlias

	f := func(lock unsafe.Pointer) {
		l := (*sync.Mutex)(lock)
		l.Unlock()
		fmt.Println("unlock", lock)
	}

	loc.Lock()
	ma.Lock()
	f(unsafe.Pointer(&loc))
	f(unsafe.Pointer(&ma))
}
