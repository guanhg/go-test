package pkg

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	maxRuntime = 4
)

type (
	Job func() error
)

func TaskImpTest() error {
	rn := float32(rand.Intn(maxRuntime)) / maxRuntime
	defer func() {
		fmt.Printf("[Done]Sleep %02fs\n", rn)
	}()
	time.Sleep(time.Duration(rn*1000) * time.Millisecond)
	if rn < 0.5 {
		return fmt.Errorf("err")
	}
	return nil
}

func TimeCost(fn Job) (time.Duration, error) {
	start := time.Now()
	err := fn()
	return time.Since(start), err
}
