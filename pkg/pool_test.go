package pkg

import (
	"testing"
)

func TestPool(t *testing.T) {
	p := NewPool(5)
	defer p.Close()

	p.Start()

	for i := 0; i < 10; i++ {
		p.AddJob(TaskImpTest)
	}

	if err := p.Wait(); err == nil {
		p.Stats.ShowStats()
	}
}
