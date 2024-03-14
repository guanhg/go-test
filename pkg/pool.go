package pkg

// 协程池

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// 协程池

type work struct {
	f Job
}

func (w *work) execute() error {
	return w.f()
}

type Pool struct {
	size int
	pc   chan work

	cache []work     // 作缓存，不阻塞Add操作
	mux   sync.Mutex // 添加cache元素时加锁

	closed atomic.Bool    // 原子取值
	wg     sync.WaitGroup // 用于等待所有任务完成

	Stats StatsTask // 用来统计task运行时间和错误信息
}

func NewPool(size int) *Pool {
	p := &Pool{
		size: size,
		pc:   make(chan work, size),
	}
	p.closed.Store(false)

	return p
}

func (p *Pool) Start() {
	go func() {
		for {
			if p.closed.Load() {
				break
			}
			for _, w := range p.cache {
				p.pc <- w
			}
			p.cache = p.cache[:0]
			time.Sleep(10 * time.Millisecond)
		}
	}()

	for i := 0; i < p.size; i++ {
		go func() {
			for w := range p.pc {
				cost, err := TimeCost(w.f)
				p.wg.Done()
				// 统计结果
				p.Stats.AddTag(NewStatsTag("", cost, err))
			}
		}()
	}
}

func (p *Pool) AddJob(f Job) error {
	if p.closed.Load() {
		return fmt.Errorf("Pool is closed")
	}
	p.wg.Add(1)

	p.mux.Lock()
	defer p.mux.Unlock()
	p.cache = append(p.cache, work{f: f})
	return nil
}

func (p *Pool) Wait() error {
	if p.closed.Load() {
		return fmt.Errorf("Pool is closed")
	}
	p.wg.Wait()
	return nil
}

func (p *Pool) Close() {
	p.closed.Store(true)
	close(p.pc)
}
