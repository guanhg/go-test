package test

import (
	"fmt"
	"testing"
)

// case 2
func TestChannel2(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	if n, ok := <-ch; !ok {
		fmt.Println("not ok")
	} else {
		fmt.Println(n)
	}

	close(ch)
	if n, ok := <-ch; !ok {
		fmt.Println("not ok")
	} else {
		fmt.Println(n)
	}

}

// case 3
func TestChannel3(t *testing.T) {
	ch := make(chan int, 1)
	// 由于for-range会阻塞，所以需要用go关键字
	go func() {
		for v := range ch {
			fmt.Println(v)
		}
	}()
	for i := 0; i < 10; i++ {
		ch <- i
	}
}

// case 4
func TestChannel4(t *testing.T) {
	ch := make(chan int, 1)
	ch2 := make(chan string, 1)
	for i := 1; i < 10; i++ {
		ch <- 1
		ch2 <- "abc"

		select {
		case v2 := <-ch2:
			fmt.Println(v2, <-ch)
		case v := <-ch:
			fmt.Println(v, <-ch2)
		default:
			fmt.Println("default")
		}
	}
}

// case 5
func TestChannel5(t *testing.T) {
	// 带缓冲
	chs := make(chan int, 3)
	defer close(chs)
	chs <- 1
	fmt.Println(<-chs)
	// 无缓冲
	ch := make(chan struct{})
	defer close(ch)
	// 阻塞当前协程
	ch <- struct{}{}
	// 异步send
	go func() {
		ch <- struct{}{}
	}()
}

// case6
func TestChannel6(t *testing.T) {
	var ch chan struct{}
	ch <- struct{}{}
}
