package pkg

import (
	"fmt"
	"sync"
	"testing"
)

func TestLogCollect(t *testing.T) {
	lc := NewLogCollector(20)
	defer lc.Close()

	var wg sync.WaitGroup
	f := func(i int, wg *sync.WaitGroup) {
		defer wg.Done()
		lc.Collect(fmt.Sprintf("%d", i))
	}

	for i := 0; i <= 130; i++ {
		wg.Add(1)
		go f(i, &wg)
	}

	for i := 0; i < 100; i++ {
		lc.Collect(fmt.Sprintf("%d-", i))
	}
	wg.Wait()
}
