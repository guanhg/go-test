package test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAtomic(t *testing.T) {
	var vi atomic.Int32
	var vc int32

	for i := 0; i < 1000; i++ {
		go func() {
			vc++
			vi.Add(1)
			if vc == vi.Load() {
				return
			}
			fmt.Println(vc, vi.Load())
		}()
	}
}

func TestAtomic2(t *testing.T) {
	// cfg := &Config{}
	var a []int
	// 读
	go func() {
		var i int
		for ; ; i++ {
			a = []int{i, i + 1, i + 2, i + 3}
			a = append(a, []int{i + 4, i + 5}...)
		}
	}()

	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for n := 0; n < 1000; n++ {
				fmt.Println(a)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestAtomic3(t *testing.T) {
	var v atomic.Value
	// 读
	go func() {
		var i int
		for ; ; i++ {
			a := []int{i, i + 1, i + 2, i + 3}
			a = append(a, []int{i + 4, i + 5}...)
			v.Store(a)
		}
	}()

	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for n := 0; n < 100; n++ {
				fmt.Println(v.Load())
				time.Sleep(time.Millisecond)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
