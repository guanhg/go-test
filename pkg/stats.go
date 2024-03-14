package pkg

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// 非常简单的用于统计函数运行指标：错误、是否成功、耗时

type StatsTag struct {
	name string
	err  error
	fail bool
	cost time.Duration
}

func NewStatsTag(name string, cost time.Duration, err error) StatsTag {
	return StatsTag{
		name: name,
		err:  err,
		fail: err != nil,
		cost: cost,
	}
}

type StatsTask struct {
	mux   sync.Mutex
	count atomic.Int32
	succ  []StatsTag
	fail  []StatsTag

	wg sync.WaitGroup
}

func NewStatsTask() *StatsTask {
	return &StatsTask{}
}

func (s *StatsTag) ShowStats() {
	fmt.Printf("Name: %s | Fail: %t | cost: %f s | err: %v \n", s.name, s.fail, s.cost.Seconds(), s.err)
}

func (st *StatsTask) AddTag(tag StatsTag) {
	st.mux.Lock()
	defer st.mux.Unlock()

	if tag.fail {
		st.fail = append(st.fail, tag)
	} else {
		st.succ = append(st.succ, tag)
	}
}

func (st *StatsTask) ShowStats() {
	fmt.Println("Success: ")
	for _, tag := range st.succ {
		fmt.Printf("    ")
		tag.ShowStats()
	}
	fmt.Println("Fail")
	for _, tag := range st.fail {
		fmt.Printf("    ")
		tag.ShowStats()
	}
}

// 直接在stats中执行job，并统计结果
// jc: job通道，在外部声明，主要要close
// gc: 协程数
func (st *StatsTask) StartWithJobChan(jc chan Job, gc int) {
	for i := 0; i < gc; i++ {
		go func() {
			for job := range jc {
				st.wg.Add(1)
				cost, err := TimeCost(job)

				tag := StatsTag{
					name: reflect.Func.String(),
					fail: err != nil,
					err:  err,
					cost: cost,
				}
				st.count.Add(1)
				st.mux.Lock()
				if err != nil {
					st.fail = append(st.fail, tag)
				} else {
					st.succ = append(st.succ, tag)
				}
				st.mux.Unlock()
				st.wg.Done()
			}
		}()
	}
}

func (st *StatsTask) Wait() {
	st.wg.Wait()
}
