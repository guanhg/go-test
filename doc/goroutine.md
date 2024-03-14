[Go程序入口](https://github.com/golang/go/blob/master/src/runtime/asm_amd64.s#L170) | [G](https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L422)[M](https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L550)[P](https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L646)

# Goroutine
是一种`用户态`的轻量级的线程，由Go语言的运行时管理。它允许程序在不同的执行路径上并发地运行代码，而不需要显式地创建线程。通过关键字"go"来启动一个协程，这样的协程可以方便地进行并发编程，提高程序的性能和效率，

## 创建
在内部代码中, g结构体通过`getg()`获取. 而`getg()`的实现是汇编代码, 原理是从`当前线程的TLS(thread-local storage)`中获取正在运行的g结构体指针.

1. 程序启动时, 当前线程创建并获取一个g0协程, 为系统线程的主协程, g0用来做线程的线程执行其它普通协程时的启动和收尾工作, 普通协程的入口
   g0的汇编初始化
```asm
		MOVQ    $runtime·g0(SB), DI  //g0的地址放入DI寄存器
		LEAQ   (-64*1024+104)(SP), BX//BX = SP - 64*1024 + 104
		MOVQ   BX, g_stackguard0(DI)//g0.stackguard0 = SP - 64*1024 + 104
		MOVQ   BX, g_stackguard1(DI)//g0.stackguard1 = SP - 64*1024 + 104
		MOVQ   BX, (g_stack+stack_lo)(DI)//g0.stack.lo = SP - 64*1024 + 104
		MOVQ   SP, (g_stack+stack_hi)(DI)//g0.stack.hi = SP
```

2. m0的汇编初始化, 为整个程序的主线程，设置`m0.tls`, 并把g0绑定到m0的tls中，完成m0, g0的相互绑定
```asm
	LEAQ   runtime·m0+m_tls(SB), DI   //DI = &m0.tls，取m0的tls成员的地址到DI寄存器
	CALL runtime·settls(SB)  //调用settls设置线程本地存储，settls函数的参数在DI寄存器中

	// store through it, to make sure it works
	get_tls(BX) //获取fs段基地址并放入BX寄存器，其实就是m0.tls[1]的地址
	MOVQ   $0x123, g(BX) //把整型常量0x123拷贝到fs段基地址偏移-8的内存位置，也就是m0.tls[0] = 0x123
	MOVQ   runtime·m0+m_tls(SB), AX//AX = m0.tls[0]
	CMPQ   AX, $0x123//检查m0.tls[0]的值是否是通过线程本地存储存入的0x123来验证tls功能是否正常
	JEQ 2(PC)
	CALL runtime·abort(SB)

	ok:
   // set the per-goroutine and per-mach "registers"
   get_tls(BX)//获取fs段基址到BX寄存器
   LEAQ   runtime·g0(SB), CX //CX = g0的地址
   MOVQ   CX, g(BX)//把g0的地址保存在线程本地存储里面，也就是m0.tls[0]=&g0
   LEAQ   runtime·m0(SB), AX //AX = m0的地址

   // save m->g0 = g0
   MOVQ   CX, m_g0(AX)
   // save m0 to g0->m
   MOVQ   AX, g_m(CX)
   ...
   CALL runtime·osinit(SB) //执行的结果是全局变量 ncpu = CPU核数 
   CALL runtime·schedinit(SB)//调度系统初始化
```

3. 调度器初始化
schedinit初始化了最大线程数, m0的部分字段, allp全局变量等等, allp是全局processor的队列, 它的长度决定了多少个协程可以并行运行

4. 创建第一个普通协程, 通过`runtime.newproc->newproc1, 创建第一个普通协程. 初始化一个新的g结构体`, 并存储到m0.p中
```
	// create a new goroutine to start program
	MOVQ   $runtime·mainPC(SB), AX  // entry,mainPC是runtime.main
	//newproc的第二个参数入栈，也就是新的goroutine需要执行的函数
	PUSHQ  AX   //AX = &funcval{runtime·main},

	//newproc的第一个参数入栈，该参数表示runtime.main函数需要的参数大小，因为runtime.main没有参数，所以这里是0
	PUSHQ  $0   // arg size
	CALL runtime·newproc(SB)//创建main goroutine
	POPQ   AX
	POPQ   AX
```
```go
	func newproc(siz int32, fn *funcval) {
		//注意：argp指向fn函数的第一个参数，而不是newproc函数的参数
		//参数fn在栈上的地址+8的位置存放的是fn函数的第一个参数
		argp := add(unsafe.Pointer(&fn), sys.PtrSize)
		<span data-word-id="785" class="abbreviate-word">gp</span> := getg()//获取正在运行的g，初始化时是m0.g0
		//getcallerpc()返回一个地址，也就是调用newproc时由call指令压栈的函数返回地址，
		//对于我们现在这个场景来说，pc就是CALLruntime·newproc(SB)指令后面的POPQ AX这条指令的地址
		pc := getcallerpc()

		//systemstack的作用是切换到g0栈执行作为参数的函数
		//我们这个场景现在本身就在g0栈，因此什么也不做，直接调用作为参数的函数
		systemstack(func() {
			newg := newproc1(fn, argp, siz, gp, pc)

			_p_ := getg().m.p.ptr()
			runqput(_p_, newg, true)//把newg放入_p_的运行队列，初始化的时候一定是p的本地运行队列，其它时候可能因为本地队列满了而放入全局队列

			if mainStarted {//初始化时不会执行
				wakep()
			}
		})
	}
```

5. 程序进入调度循环, 调用链`mstart -> schedule -> execute, execute不会退出`, 从m.p或全局g队列中获取g结构, 执行过程中会把g结构放入到m.tls中, 方便通过`getg()从m.tls获取`
```asm
		// start this M 
		CALL runtime·mstart(SB) //主线程进入调度循环，运行刚刚创建的goroutine  
		CALL runtime·abort(SB)  // mstart should never return RET
```
```go
		func schedule() {
			_g_ := getg() //_g_ = g0
			...
			var gp *g//创建一个g指针
			...
			//接下来开始找到要调度执行的g
			if gp == nil {
				//为了保证调度的公平性，每进行61次调度就需要优先从全局运行队列中获取goroutine，
				//因为如果只调度本地队列中的g，那么全局运行队列中的goroutine将得不到运行
				if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
					lock(&sched.lock)
					gp = globrunqget(_g_.m.p.ptr(), 1)
					unlock(&sched.lock)
				}
			}

			if gp == nil {
				//从与m关联的p的本地运行队列中获取goroutine
				gp, inheritTime = runqget(_g_.m.p.ptr())
				// We can see gp != nil here even if the M is spinning,
				// if checkTimers added a local goroutine via goready.
			}

			if gp == nil {
				//如果从本地运行队列和全局运行队列都没有找到需要运行的goroutine，
				//则调用findrunnable函数从其它工作线程的运行队列中偷取，如果偷取不到，则当前工作线程进入睡眠，
				//直到获取到需要运行的goroutine之后findrunnable函数才会返回。
				gp, inheritTime = findrunnable() // blocks until work is available
			}
			...
			//当前运行的是runtime的代码，函数调用栈使用的是g0的栈空间
			//调用execte切换到gp的代码和栈空间去运行
			execute(gp, inheritTime)
		}
```
	
6. 使用第4步创建的第一个普通协程来执行runtime.main, 主要功能如下
	- 启动一个sysmon系统监控线程，该线程负责整个程序的gc、抢占调度以及netpoll等功能的监控，在后续分析抢占调度时我们再详细分析；
	- 执行runtime包的初始化；
	- 执行main包以及main包import的所有包的初始化；
	- 执行main.main函数；`即用户的main函数`
	- 从main.main函数返回后调用exit系统调用退出进程；

7. 在main.main函数即`用户main函数`中创建一个用户普通协程, 同第4步调用runtime.newproc函数, 保存g结构
```go
		func main(){
			go func test(){}()
			fmt.Println("Hello World")
		}
```
<img src=https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/3ecf20ac0f55448589215143b0b501d3~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp heigh=300 width=600>

关联图
<img src=https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/008b18cab409400b8f775f6ba93f1c7e~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp heigh=400 width=400>

## 运行

8. 在第7步中使用newproc创建新的g结构,并保存到m.p. 协程实际上是通过调用链`newproc -> wakep -> startm`来运行
```go
	func newproc(siz int32, fn *funcval) {
			...
			if mainStarted {//初始化时不会执行，执行runtime.main之后的所有调用都会执行这部分代码了
				wakep()
			}
	}
	// wakep 主要执行 startm
	func startm(_p_ *p, spinning bool) {
		mp := acquirem()
		lock(&sched.lock)
		if _p_ == nil {//没有指定p的话需要从p的空闲队列中获取一个p
			_p_ = pidleget() //从p的空闲队列中获取空闲p
			if _p_ == nil {
				unlock(&sched.lock)
				...
				releasem(mp)
				return
			}
		}

		nmp := mget() //从m空闲队列中获取正处于睡眠之中的工作线程，所有处于睡眠状态的m都在此队列中
		if nmp == nil { //没有处于睡眠状态的工作线程
			id := mReserveID()
			unlock(&sched.lock)
			...

			newm(fn, _p_, id) //创建新的工作线程
			// Ownership transfer of _p_ committed by start in newm.
			// Preemption is now safe.
			releasem(mp)
			return
		}
		unlock(&sched.lock)
		...

		//唤醒处于休眠状态的工作线程
		notewakeup(&nmp.park)
		// Ownership transfer of _p_ committed by wakeup. Preemption is now
		// safe.
		releasem(mp)
	}
```

## 退出
9. 当协程执行结束后，g中的sp寄存器的值就会弹到PC中来执行退出操作goexit 
- 调用链: `goexit -> goexit1 -> mcall -> goexit0 -> schedule`. 最后进入调度函数,  然后从m.p或全局g队列中获取一个g结构运行. 完成`步骤5~步骤9`的一次调度循环

## 调度
协程的调度包含方面几个目标: 调度模型GMP, 调度策略, 调度时机, 协程切换
### 调度模型GMP
<img src=https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/30e25af551cb4eb0b870ecf6d5e4f8f0~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp heigh=1000 width=400>
- G: 协程, 调度器包含一个`全局协程队列`
- M: 系统线程, 协程真正运行的单位, 切换由操作系统管理, 每个M关联一个P
- P: 协程队列, 每个P最多管理256个G, 有一个全局P数组变量

### 调度策略

#### G的创建存储与获取
- 创建
	调用runtime.newproc来创建一个普通协程时. 如果m.p本地队列没满, 则存入其中, 否则, `把m.p一半元素和newg放到全局队列sched.runq中`
- 获取
	调用runtime.schedule来判断把哪个g放到m上运行
	1. 调度器每切换61次, 则从全局队列sched.runq中获取, 保证全局队列中的g能够得到运行
	2. 否则, 从本地队列m.p中获取
	3. 否则, 再次检查全局队列sched.runq；否则，检查网络轮询netpoll结果；否则，从其他线程的队列窃取，即`伪随机从全局p队列allp中选取一个p，从p中获取g`
	
#### M的创建存储与获取
每个M都关联一个g0, 用于普通协程的传参和收尾
- 创建
	1. 调用链runtime.startm -> newm 来创建一个系统线程, 放到系统上去运行. 如果`m没有任务需要执行或m退出系统调用`时, 会把m结构放到调度器的m空闲列表中
- 获取
	1. 调用链runtime.startm -> acquirem 从调度器的m空闲队列中获取

#### P的创建与获取
- 创建
	1. 在调度器初始化时, 会创建一个全局队列allp, 是一个数组结构, 其长度默认为cpu的核数
- 获取
	1. 从全局队列allp中获取
## 调度时机
从调度机制上来讲，操作系统的调度可以分为协作式和抢占式。
- 协作式调度一般分为两种情形：
通过主动调用runtime.GoSched()函数来让出cpu资源
	1. 被动调度：goroutine执行某个操作因条件不满足需要等待而发生的调度（如等待锁、等待时间、等待IO资源就绪等）；
	2. 主动调度：goroutine主动调用Gosched()函数让出CPU而发生的调度；
	
协作式调度的调用链都会进入goschedImpl函数:`设置状态`, `解绑g和m`(把g放到全局队列allgs中), 进入新的调度等
```go
func goschedImpl(gp *g) {
	status := readgstatus(gp)
	if status&^ _Gscan != _Grunning {//当前的g为 _Grunning状态才能释放
		dumpgstatus(gp)
		throw("bad g status")
	}
	casgstatus(gp, _Grunning, _Grunnable)//修改g的运行状态
	dropg() //设置当前m.curg = nil, gp.m = nil
	lock(&sched.lock)
	globrunqput(gp)//把gp放入sched的全局运行队列runq
	unlock(&sched.lock)

	schedule()//进入新一轮调度
}
```

- 抢占式调度
抢占式分为`时间抢占`(本质上也是信号抢占)和`系统信号`抢占，通过`系统监控线程`来触发抢占行为，系统监控线程在runtime.main中被初始化
抢占调用链 runtime.sysmon -> retake
1. 遍历全局队列allp，如果p的运行时间超过10ms触发抢占
2. 遍历全局队列allp，如果p的状态是系统调用，寻找新的m来接管p，然后再通过运行时间触发抢占
```
func retake(now int64) uint32 {
    n := 0
    // Prevent allp slice changes. This lock will be completely
    // uncontended unless we're already stopping the world.
    lock(&allpLock)
    // We can't use a range loop over allp because we may
    // temporarily drop the allpLock. Hence, we need to re-fetch
    // allp each time around the loop.
    for i := 0; i < len(allp); i++ {//遍历所有p，然后根据p的状态进行抢占
        _p_ := allp[i]
        if _p_ == nil {
            // This can happen if procresize has grown
            // allp but not yet created new Ps.
            continue
        }
        pd := &_p_.sysmontick
        s := _p_.status
        sysretake := false
        if s == _Prunning || s == _Psyscall {
            t := int64(_p_.schedtick)
            if int64(pd.schedtick) != t {
                //pd.schedtick != t 说明(pd.schedwhen ～ now)这段时间发生过调度
                //重置跟sysmon相关的schedtick和schedwhen变量
                pd.schedtick = uint32(t)
                pd.schedwhen = now
            } else if pd.schedwhen+forcePreemptNS <= now {
                //连续运行超过10毫秒了，设置抢占请求
                preemptone(_p_)
                // In case of syscall, preemptone() doesn't
                // work, because there is no M wired to P.
                sysretake = true
            }
        }
        if s == _Psyscall {//系统调用抢占处理
            // Retake P from syscall if it's there for more than 1 sysmon tick (at least 20us).
            t := int64(_p_.syscalltick)
            if !sysretake && int64(pd.syscalltick) != t {//判断是否已发生过调度
				  pd.syscalltick = uint32(t)
				  pd.syscallwhen = now
                continue
            }
            // On the one hand we don't want to retake Ps if there is no other work to do,
            // but on the other hand we want to retake them eventually
            // because they can prevent the sysmon thread from deep sleep.
            // 1. 本地P中无等待运行的goroutine
            // 2. 有空闲的m、p
            // 3. 从上一次监控线程观察到 p 对应的 m 处于系统调用之中到现在小于 10 毫秒
            if runqempty(_p_) && atomic.Load(&sched.nmspinning)+atomic.Load(&sched.npidle) > 0 && pd.syscallwhen+10*1000*1000 > now {
                continue
            }
            // Drop allpLock so we can take sched.lock.
            unlock(&allpLock)
            // Need to decrement number of idle locked M's
            // (pretending that one more is running) before the CAS.
            // Otherwise the M from which we retake can exit the syscall,
            // increment nmidle and report deadlock.
            incidlelocked(-1)
            if atomic.Cas(&_p_.status, s, _Pidle) { //cas修改p的状态为 _Pidle
                if trace.enabled {
                    traceGoSysBlock(_p_)
                    traceProcStop(_p_)
                }
                n++
                _p_.syscalltick++
                handoffp(_p_) //寻找一个新的m出来接管P
            }
            incidlelocked(1)
            lock(&allpLock)
        }
    }
    unlock(&allpLock)
    return uint32(n)
}
```

## 总结
G状态切换图
<img src=https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/e6272ddbaeeb4a6ebb66b5502d4158c9~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp>

P状态切换图
<img src=https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/693ef311931b47a49b159c0cdecbaa3b~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp>

- `管道的阻塞`是一种协作式调度
	1. 调用gopark函数解绑g和m，进入新一轮进入调度
	2. 调用goready函数激活管道阻塞的协程，并把g放到m.p上，注意`goready是在系统栈上调用`
- `系统调用`抢占式调度
	系统调用会阻塞系统线程m，所以需要分离p和m，让p绑定其它m来运行其它协程。
	分为阻塞式系统调用和非阻塞式系统调用
	1. 非阻塞式调用runtime.entersyscall主要是`设置抢占信号和状态`，由监控线程安排调度，监控线程会分离m和p的绑定，为p重现找一个m去运行，如`读取文件`等
	2. 阻塞式调用runtime.entersyscallblock，分离p和m，然后由系统栈函数自动调用，如`系统休眠`等
	3. 退出调用，`设置抢占信号`触发新一轮调度
 - `网络轮询`抢占式调度
	网络轮询是由系统监控线程`定时读取已准备的协程`，然后设置协程状态
- `垃圾回收`抢占式调度
	1. 进入GC前，通过stopTheWorld把allp队列中的所有元素设置垃圾回收状态，然后触发抢占信号
	2. GC完成后，通过startTheWorld把allp队列中元素设置运行状态


## Ref
- [goroutine的创建及调度](https://juejin.cn/post/6986566386176753694)
- [深入runtime](https://zboya.github.io/post/go_scheduler/#pprocessor)