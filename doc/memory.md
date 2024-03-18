[toc]

# Memory
Go语言内置运行时（就是Runtime），抛弃了传统的内存分配方式，改为自主管理。这样可以自主地实现更好的内存使用模式，比如内存池、预分配等等。这样，不会每次内存分配都需要进行系统调用。

>   Golang运行时的内存分配算法主要源自 Google 为 C 语言开发的TCMalloc算法，全称Thread-Caching Malloc。核心思想就是把内存分为多级管理，从而降低锁的粒度。它将可用的堆内存采用二级分配的方式进行管理。
> 每个线程都会自行维护一个独立的内存池，进行内存分配时优先从该内存池中分配，当内存池不足时才会向全局内存池申请，以避免不同线程对全局内存池的频繁竞争

## 分配malloc
### 内存模型
<img src=https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2019/11/26/16ea7188a1df53e1~tplv-t2oaga2asx-jj-mark:3024:0:0:0:q75.png>

> Go在程序启动的时候，内存管理器会先向操作系统申请一块内存，切成小块后自己进行管理。

申请到的内存块被分配了三个区域，在X64上分别是512MB，16GB，512GB大小。
**注意：** 这时还只是一段虚拟的地址空间，并不会真正地分配内存

- arena区域
    <img src=https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2019/11/26/16ea71849399e784~tplv-t2oaga2asx-jj-mark:3024:0:0:0:q75.png>
`真正的堆区`，Go动态分配的内存都是在这个区域，它把内存分割成8KB大小的页，一些页组合起来称为mspan。

> **mspan**：由多个连续的页（page [大小：8KB]）组成的大块内存，面向内部管理
  **object**：将mspan按照特定大小切分成多个小块，每一个小块都可以存储对象，面向对象分配

```go
	type mspan struct {
		next           *mspan    	
		prev           *mspan    	
		startAddr      uintptr   	// 起始序号
		npages         uintptr   	// 管理的页数
		manualFreeList gclinkptr 	// 待分配的 object 链表
		nelems 		uintptr 		// 块个数，表示有多少个块可供分配
		allocCount     uint16		// 已分配块的个数
		...
	}
type heapArena struct {
   // arena 的位标记
   bitmap [heapArenaBitmapBytes]byte
	// arena 的所有mspan指针
   spans [pagesPerArena]*mspan
   ...
}
```
- bitmap区域
	标识arena区域哪些地址保存了对象，并且用4bit标志位表示对象是否包含指针、GC标记信息。

- spans区域
	存放mspan的指针变量，每个指针对应一页，所以spans区域的大小就是512GB/8KB*8B=512MB。
	除以8KB是计算arena区域的页数，而最后乘以8是计算每个页指针的大小。

### 内存管理
**基本策略**
> 1.每次从操作系统申请一大块内存，以减少系统调用。
 将申请的大块内存按照特定的大小预先的进行切分成小块，构成链表。
 2.为对象分配内存时，只需从大小合适的链表提取一个小块即可。
 3.回收对象内存时，将该小块内存重新归还到原链表，以便复用。
 4.如果闲置内存过多，则尝试归还部分内存给操作系统，降低整体开销。

**注意：**内存分配器只管理内存块，并不关心对象状态，而且不会主动回收，垃圾回收机制在完成清理操作后，触发内存分配器的回收操作

---
**管理策略**

<img src=https://img.draveness.me/2020-02-29-15829868066457-multi-level-cache.png>
> 真实的内存管理策略比基本策略要复杂，go使用`多级缓存`模式去管理内存的创建和申请；管理器包含`页堆(heap)、中心缓存(central)、线程缓存(cache)`几个组件

#### heap 
管理闲置的mspan，需要时向操作系统申请，Go程序持有的所有堆空间，使用一个mheap的全局对象_mheap来管理堆内存
> 1. 当mcentral没有空闲的mspan时，会向mheap申请。而mheap没有资源时，会向操作系统申请新内存。
> 2. mheap主要用于`大对象`的内存分配，以及管理未切割的mspan，用于给mcentral切割`小对象`。当申请一个大对象时，直接分配到堆上，不经过cache、central。

```go
type mheap struct {
	lock        mutex
	spans       []*mspan // spans: 指向mspans区域，用于映射mspan和page的关系
	bitmap      uintptr  // 指向bitmap首地址，bitmap是从高地址向低地址增长的
	arenas [1 << arenaL1Bits]*[1 << arenaL2Bits]*heapArena   // 所有arena指针
	arena_start uintptr  // 指示arena区首地址
	arena_used  uintptr  // 指示arena区已使用地址位置
	arena_end   uintptr  // 指示arena区末地址
	central [numSpanClasses]struct {
		mcentral mcentral
		pad      [sys.CacheLineSize-unsafe.Sizeof(mcentral{})%sys.CacheLineSize]byte
	}	//每个 central 对应一种 sizeclass，总共68种类型
}
```

#### central
为所有cache提供切割好的后备mspan资源，包括已分配出去的和未分配出去的。全局变量mheap_包含了一个长度为126的central数组，`数组中的每个central保存一种特定大小的全局mspan列表`，而每种mspan意味着object大小区间的不同。
```go
type mcentral struct {
	// span的类型，每种span用于保存特定大小区间的对象
   spanclass spanClass
	// partial和full：一个扫描使用中span，一个不扫描，这两个变量在一次gc后会互换
   partial [2]spanSet // list of spans with a free object
   full    [2]spanSet // list of spans with no free objects
}
```

#### cache
与工作线程绑定，缓存可用的mspan资源，用于object申请，因为同一个系统线程中不存在多个routine竞争，所以不需要加锁。
> mcache包含所有类型的mspan，每一种span意味着对象的大小在一个固定的区间内

```go
_NumSizeClasses = 68
numSpanClasses = _NumSizeClasses << 1 //136

type mcache struct {
	...
	alloc [numSpanClasses]*mspan//以numSpanClasses 为索引管理多个用于分配的 span
	...
}
```
mcache用Span Classes作为索引管理多个用于分配的mspan，它包含所有规格的mspan，它是 numClasses 的2倍，**这是为了加速之后内存回收的速度，数组里一半的mspan中分配的对象不包含指针，另一半则包含指针。对于无指针对象的mspan在进行垃圾回收的时候无需进一步扫描它是否引用了其他活跃的对象**。

### 总结
- new(object)调用runtime.mallocgc函数来分配对象

[分配流程](#)：

> 1.计算待分配对象的规格（size_class）
  2.从cache.alloc数组中找到规格相同的span
  3.从span.manualFreeList链表提取可用object
  4.如果span.manualFreeList为空，从central获取新的span
  5.如果central.nonempty为空，从mheap.free(freelarge) 获取，并切分成object链表
  6.如果heap没有大小合适的span，向操作系统申请新的内存

[释放流程](#)：
 
>1.将标记为可回收的object交还给所属的span.freelist
 2.该span被放回central，可以提供cache重新获取
 3.如果span以全部回收object，将其交还给heap，以便重新分切复用
 4.定期扫描heap里闲置的span，释放其占用的内存
  
注意：以上流程不包含大对象，它直接从heap分配和释放


Go语言的内存分配非常复杂，它的一个原则就是能复用的一定要复用。
- Go在程序启动时，会向操作系统申请一大块内存，之后自行管理。
- Go内存管理的基本单元是mspan，它由若干个页组成，每种mspan可以分配特定大小的object。
- mcache, mcentral, mheap是Go内存管理的三大组件，层层递进。mcache管理线程在本地缓存的mspan；mcentral管理全局的mspan供所有线程使用；mheap管理Go的所有动态分配内存。
- 一般小对象通过mspan分配内存；大对象则直接由mheap分配内存。

# 回收gc
垃圾回收一般分为标记阶段和清除阶段

## 三色标记
[三色标记法](https://community.apinto.com/d/34057-golang-gc)

## 垃圾清理
[gcStart源码解读](https://juejin.cn/post/7271620031510806586?searchId=2024031021071632C1320B60537253D9A1)