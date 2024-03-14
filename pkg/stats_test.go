package pkg

import "testing"

func TestPressTest(t *testing.T) {
	jc := make(chan Job, 1)
	defer close(jc)

	stat := NewStatsTask()
	go stat.StartWithJobChan(jc, 3)

	for i := 0; i < 10; i++ {
		jc <- TaskImpTest
	}

	stat.Wait()

	stat.ShowStats()

}
