[toc]

# channel
channel是实现CSP(通信顺序进程)的一种通信单元, 用于协程间的同步/通信, 是一个双向通信的管道

<img src=https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2019/12/8/16ee5c45b1dcb0ea~tplv-t2oaga2asx-jj-mark:3024:0:0:0:q75.awebp heigh=300 width=400 />

## 发送send
示例1
```go
	func TestChannel1(t *testing.T) {
		ch := make(chan int)
		defer close(ch)
		ch <- 1
		
		var ch2 chan int  // nil
		ch2 <- 2 	// 阻塞
	}
```

- 对于已经<font color=yellow>关闭的通道</font>不能再发送消息, 否则panic报错
- 对于nil channel发送消息, 会被一直阻塞

## 接收recv
示例2
```go
	func TestChannel2(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 1
		// ① 输出 1
		if n, ok := <-ch; !ok {
			fmt.Println("not ok")
		} else {
			fmt.Println(n)
		}
	
		close(ch)
		// ② 输出 not ok
		if n, ok := <-ch; !ok {
			fmt.Println("not ok")
		} else {
			fmt.Println(n)
		}
	}
```
- 对于已经<font color=yellow>关闭的通道</font>, 接收时不会panic, 也不会被阻塞, 但接收标志为false
- 对于nil channel接收消息, 会被一直阻塞
- channel的接收有多种方式: 1. for-range; 2. select; 3. v, ok := <-ch; 4. v := <- ch
### for-range
示例3
```go
	func TestChannel3(t *testing.T) {
		ch := make(chan int)
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
```
- channel的for-range会阻塞当前协程，一直等待channel输出新的值, 适合异步任务
  应用场景: 1.设计简单的协程池, 2.统计运行结果 等
### select
示例4
```go
	func TestChannel4(t *testing.T) {
		ch := make(chan int, 1)
		ch2 := make(chan string, 1)
		defer close(ch)
		defer close(ch2)
		ch <- 1
		ch2 <- "abc"
		select {
		case v := <-ch:
			fmt.Println(v)
		case v2 := <-c2:
			fmt.Println(v2)
		default:
			fmt.Println("default")
		}
	}
```
- select有多个case同时都有准备好时, 会随机选择一个case执行; 如果所有case都没准备好时, 如果有default, 则进入default, 否则一直阻塞. 适用"等待信号"
  应用场景: 1.等待协程终止信号 2.等待时间信号 3.多路选择 等
## 缓冲和无缓冲channel
示例5
```go
	func TestChannel5(t *testing.T) {
		// 带缓冲
		chs := make(chan int, 3)
		defer close(chs)
		// 不阻塞协程
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
		fmt.Println(<-ch)
	}
```
- 无缓冲channel, <font color=red>数据未recv时</font>, 当前协程会被一值阻塞; 所以send和recv通常不在同一个协程中, 要么异步send, 要么异步recv;
  应用场景: 1.两个协程之间同步状态/数据 等
- 带缓冲channel, send数据时, 如果有缓存空间, 则不会阻塞当前协程; 否则会阻塞
  应用场景: 1.生产-消费模式场景; 2.限制协程数量 等
```go
	// 限制3个协程运行
	chs := make(chan struct{}, 3)
	for j := 0; j < 30; j++ {         
		chs <- struct{}{}         
		go func(i int) {             
			defer func() {                 
				<-chs             
			}()            
			fmt.Printf("[%d]Do Something...\n", i)   
		}
	}
```
## nil channel
```go
	var ch chan struct{}	// nil channel
	ch <- struct{}{}		// 阻塞
	<- ch					 // 阻塞
	close(ch)				 // panic
```
- 对于nil channel, 不管是send还是recv, 都会阻塞当前协程
- close时, 产生panic

## 源码解读
[channel源码文件](https://github.com/golang/go/blob/master/src/runtime/chan.go)
```go
	// channel的主要字段
	type hchan struct {
		...
		dataqsiz uint  				// 缓冲数组大小
		buf      unsafe.Pointer		// 缓冲数组指针
		closed   uint32				// 是否已关闭
		recvq    waitq  			// recv协程队列 waitq是sudog的链表
		sendq    waitq  			// send协程队列 waitq是sudog的链表
	}
	type sudog struct {
		g *g 	 // 关联的goroutine
		c *hchan // 关联的channel
		elem     unsafe.Pointer // 接收或发送的元素v v:=<-ch or ch<-v
		...
		next     *sudog  // 用来构建二叉树
		prev     *sudog
	}
```
---
初始化
```go
	func makechan(t *chantype, size int) *hchan {
		...
		// 1. 计算缓冲区需要的总大小（缓冲区大小*元素大小），并判断是否超出最大可分配范围
		mem, overflow := math.MulUintptr(elem.size, uintptr(size))
		if overflow || mem > maxAlloc-hchanSize || size < 0 {
			panic(plainError("makechan: size out of range"))
		}

		var c *hchan
		switch {
		case mem == 0:
			// 2. 缓冲区大小为0(无缓冲通道), 或者channel中元素大小为0（struct{}{}）时, 只需分配必需的空间即可
			c = (*hchan)(mallocgc(hchanSize, nil, true))
			c.buf = c.raceaddr()
		case elem.kind&kindNoPointers != 0:
			// 2. 元素类型不是指针，分配一片连续内存空间，所需空间等于 缓冲区数组空间 + hchan必要空间。
			c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
			c.buf = add(unsafe.Pointer(c), hchanSize)
		default:
			// 2. 元素中包含指针，为hchan和缓冲区分别分配空间
			c = new(hchan)
			c.buf = mallocgc(mem, elem, true)
		}
		...
		return c
	}
```
1. 判断缓冲数组内存大小
2. 根据元素和数组内存大小, 分配内存
- <font color=red>对于无缓冲channel来说, hchan.dataqsiz=0, 通信过程不需要用到缓冲数组</font>

---

发送 chansend
```go
	func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
		// channel为nil
		if c == nil {
			// 如果是非阻塞，直接返回发送不成功
			if !block {
				return false
			}
			// 否则，当前Goroutine阻塞挂起
			gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
			throw("unreachable")
		}
		...
		// 如果channel已关闭
		if c.closed != 0 {
			// 解锁，直接panic
			unlock(&c.lock)
			panic(plainError("send on closed channel"))
		}

		// 除了以上情况，当channel未关闭时，就有以下几种情况：

		// 1、当存在等待接收的Goroutine
		if sg := c.recvq.dequeue(); sg != nil {
			// 那么直接把正在发送的值发送给等待接收的Goroutine
			send(c, sg, ep, func() { unlock(&c.lock) }, 3)
			return true
		}

		// 2、当缓冲区未满时, 赋值元素到缓冲数组
		if c.qcount < c.dataqsiz {
			...
			// 将当前发送的值拷贝到缓冲区
			typedmemmove(c.elemtype, qp, ep)
			...
			return true
		}

		// 3、当既没有等待接收的Goroutine，缓冲区也没有剩余空间，如果是非阻塞的发送，那么直接解锁，返回发送失败
		if !block {
			unlock(&c.lock)
			return false
		}
		
		// 4、如果发送阻塞，那么就获取一个sudog结构体并绑定g和chan，然后加入channel的sendq队列
		gp := getg()
		mysg := acquireSudog()
		...
		mysg.elem = ep
		mysg.g = gp
		mysg.c = c
		gp.waiting = mysg
		c.sendq.enqueue(mysg)

		// 5. 调用goparkunlock挂起当前Goroutine，并进入休眠等待被唤醒
		goparkunlock(&c.lock, waitReasonChanSend, traceEvGoBlockSend, 3)
		KeepAlive(ep)

		// 6. 被唤醒之后执行清理工作并释放sudog结构体
		...
		releaseSudog(mysg)
		return true
	}
```
- 第1步, 直接从recvq中获取等待协程, 如果存在, 则通过send函数去唤醒协程
- 第2步, <font color=red>对于无缓冲channel, 因为datasiz=0而跳过, 经过第3~5步后会被直接挂起</font>
- 第4步, 把初始化的sudog结构赋给了协程. 等待协程唤醒后运行
- 第5步, 通过系统阻塞当前协程, 唤醒后才会往下运行

---

接收chanrecv
```go
	func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
		// channel为nil
		if c == nil {
			// 非阻塞直接返回（false, false）
			if !block {
				return
			}
			// 阻塞接收，则当前Goroutine阻塞挂起
			gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
			throw("unreachable")
		}

		...

		// 如果channel已关闭，并且缓冲区无元素
		if c.closed != 0 && c.qcount == 0 {
			...
			// 有等待接收的变量（即 v = <-ch中的v）
			if ep != nil {
				//根据channel元素的类型清理ep对应地址的内存，即ep接收了channel元素类型的零值
				typedmemclr(c.elemtype, ep)
			}
			// 返回（true, false），即接收到值，但不是从channel中接收的有效值
			return true, false
		}

		// 除了以上非常规情况，还有有以下几种常见情况：
		// 1. 等待发送的队列sendq里存在Goroutine，那么有两种情况：当前channel无缓冲区，或者当前channel已满
		if sg := c.sendq.dequeue(); sg != nil {
			// 如果无缓冲区，从sendq接收数据；否则，从buf队列的头部接收数据，并把sendq的数据加到buf队列的尾部
			recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
			return true, true
		}

		// 2. 缓冲区buf中有元素
		if c.qcount > 0 {
			qp := chanbuf(c, c.recvx)
			...
			// 获取buf数据, 并拷贝到接收元素
			if ep != nil {
				typedmemmove(c.elemtype, ep, qp)
			}
			// 清理空间
			typedmemclr(c.elemtype, qp)
			...
			return true, true
		}

		// 3. 非阻塞模式，且没有数据可以接受
		if !block {
			// 解锁，直接返回接收失败
			unlock(&c.lock)
			return false, false
		}

		// 4. 阻塞模式，获取当前Goroutine，打包一个sudog
		gp := getg()
		mysg := acquireSudog()
		...
		mysg.elem = ep
		mysg.g = gp
		mysg.c = c
		gp.waiting = mysg
		c.recvq.enqueue(mysg)
		// 5. 挂起当前Goroutine，设置为_Gwaiting状态并解锁，进入休眠等待被唤醒
		goparkunlock(&c.lock, waitReasonChanReceive, traceEvGoBlockRecv, 3)

		// 6. 被唤醒之后执行清理工作并释放sudog结构体
		...
		releaseSudog(mysg)
		return true, !closed
	}
```
- 第1步, 直接从sendq中获取等待协程, 通过recv函数去唤醒等待发送的协程
- 第2步, <font color=red>对于无缓冲channel, 因为qcount=0而跳过, 经过第3~5步后会被直接挂起</font>
- 第4步, 把初始化的sudog结构赋给了协程. 等待协程唤醒后运行
- 第5步, 通过系统阻塞当前协程, 唤醒后才会往下运行