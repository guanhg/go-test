[toc]

# 锁
sync.Locker 定义了锁的接口
## 互斥锁
sync.Mutex是一种互斥锁实现

> 当互斥锁已经被其他协程占用

- 当前协程因为需要获取系统信号而被挂起
- 当等待时间超过1ms，互斥锁会进入饥饿模式，该模式下，新的协程被放入FIFO队列等待；在非饥饿模式下，互斥锁被随机抢占

## 自旋锁
`自旋锁并不会让当前协程被挂起`，状态是可运行的，需要手动触发协程调度来让其他协程释放锁
实现
```go
import "sync/atomic"

type SpinLocker uint32
func (sl *SpinLocker) Lock() {
    // 自旋尝试获取锁
    // 自旋 判断 锁的状态值 是否为0 ，为0的情况下置为1 这样协程就会获取到锁
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
        runtime.Gosched()
	}
}

func (sl *SpinLock) UnLock() {
  // 将 锁的状态值  由 1 置为 0
	atomic.CompareAndSwapUint32((*uint32)(sl), 1, 0)
}
```

## 乐观锁
go的乐观锁通过`atomic.CompareAndSwap`系列函数来实现

```go
var id int32
var old_id int32
var new_id int32
if atomic.CompareAndSwapInt32(&id, old_id, new_id) {  // 版本值不同
...
}
```

> 参考
- [golang 自旋锁](https://studygolang.com/articles/16480)

