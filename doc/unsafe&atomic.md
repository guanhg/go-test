[toc]

# unsafe
### uintptr
是buildin的一个定义类型: type uintptr uintptr, 本质是<font color=red>一个整型而不是指针</font>, 所以可以对uintptr作加减形式的地址运算. 
结合unsafe.Pointer, unsafe.Sizeof, unsafe.Alignof获取后续地址, 比如基于struct中一个字段的地址偏移量获取下一个字段值

### unsafe.Pointer
对任何类型变量的指针类型进行强制转换, 可以结合uintptr进行<font color=red>地址运算</font>

- 使用原则
```
1. 任意类型的指针可以转换为一个 Pointer；
2. 一个 Pointer 也可以被转为任意类型的指针；
3. uintptr 可以被转换为 Pointer;
4. Pointer 也可以被转换为 uintptr
```
示例1
```go
	func TestUnsafe1(t *testing.T){
		x := []int{1, 2, 3, 4}
		e3 := *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&x[0])) + 2*unsafe.Sizeof(x[0])))
		fmt.Println(e3) //  输出 3
	}
```

- 使用限制
```
1. 转化的目标类型（如uint64) 的 size 一定不能比原类型 (如float64)还大（二者size都是8个字节）；
2. 前后两种类型有等价的 memory layout；
```
示例2
```go
	func Int64To8(f int64) int8 {
		return *(*int8)(unsafe.Pointer(&f))
	}

	func Int8To64(f int8) int64 {
		return *(*int64)(unsafe.Pointer(&f))
	}
	
	func TestUnsafe2(t *testing.T){
		fmt.Println("64->8", Int64To8(5)) // 输出 1079252997 ,一个不确定值
		fmt.Println("8->64", Int8To64(5)) // 输出 5
	}
```
- 使用场景
  1. 获取struct/slice等结构的偏移地址的值, 如示例2
  2. 强制地址类型转换
  示例3 整型地址 -> 浮点型地址
  ```go
  	func TestUnsafe3(t *testing.T){
		var i int = 2
		fp := (*float32)(unsafe.Pointer(&i))
		fmt.Println(fp, &i)	 // ①
		i = 4
    	fmt.Println(*fp, i)  // ②
		
		ip := (*int32)(unsafe.Pointer(&i))
		i = 6
		fmt.Println(*ip, i)	  // ③
		
		f := *(*int32)(&i)   // 语法错误, 不同地址类型是无法直接转化的
	}
  ```
  ① 输出 0xc00001c298 0xc00001c298, 输出相同的地址值, 但内存结构不一样
  ② 输出 6e-45 4, 因为float32和int的内存结构不一样,所以导致值不同; 
  ③ 输出 6 6 因为int和int32内存结构相似,所以值一样
  
  示例4 byte[] <-> string
  ```go
  	func BytesToString(b []byte) string {
		return *(*string)(unsafe.Pointer(&b))
	}

	func StringToBytes(s string) []byte {
		return *(*[]byte)(unsafe.Pointer(&s))
	}
  ```
## 总结
- `unsafe.Pointer` 可以让变量在不同的指针类型转换，也就是表示为任意可寻址的指针类型。
- `unsafe.Pointer` 可以直接操作源数据地址, 无需复制, 在无复制场景下更高效

# atomic
go提供的原子化操作包括几个功能, 这些操作都是原子化的
`数据的存储和加载`, `基本数据的加减`, `比较和交换`
```go
	var v atomic.Value
	v.Store(any)
    v.Load(any)
    v.Swap(new any)
    v.CompareAndSwap(old, new any)
```
## 应用
1. v++/v--
在该场景下, 本地并发运行多协程, 会产生资源冲突
示例3
```
	func TestAtomic(t *testing.T) {
		var v atomic.Int32 // 初始化0
		var vc int32

		for i := 0; i < 1000; i++ {
			go func() {
				vc++		// 会产生冲突
				vi.Add(1)	// 原子化
				if vc != vi.Load() {
					fmt.Println(vc, vi.Load())
				}
			}()
		}
	}
```
2. 读写
对于像slice和map这类复杂结构数据, 在不同协程中读写时, 可能会出现无法预料的情况
示例4
```go
	func TestAtomic2(t *testing.T) {
		var a []int
		// 写
		go func() {
			for ; ; i++ {
				a = []int{0, 1, 2, 3}
				a = append(a, []int{4,5}...)
			}
		}()
		// 读
		var wg sync.WaitGroup
		for n := 0; n < 4; n++ {
			wg.Add(1)
			go func() {
				for n := 0; n < 100; n++ {
					fmt.Println(a)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
```
输出会出现: [0, 1, 2, 3]以及[0, 1, 2, 3, 4, 5]两种情况, 在读取a时出现了并发冲突

示例5
```go
	func TestAtomic3(t *testing.T) {
		var v atomic.Value
		// 读
		go func() {
			var i int
			for ; ; i++ {
				a := []int{0, 1, 2, 3}
				a = append(a, []int{4,5}...)
				v.Store(a)
			}
		}()

		var wg sync.WaitGroup
		for n := 0; n < 4; n++ {
			wg.Add(1)
			go func() {
				for n := 0; n < 1000; n++ {
					fmt.Println(v.Load())
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
```
输出只有一种情况: [0, 1, 2, 3, 4, 5], 存储和读取都是原子化

对比 sync.Once的实现
```go
	type Once struct {
		done uint32 // 标志操作是否操作
		m    Mutex // 锁，用来第一操作时候，加锁处理
	}
	func (o *Once) Do(f func()) {
		if atomic.LoadUint32(&o.done) == 0 {// 原子性加载o.done
			o.doSlow(f)
		}
	}

	func (o *Once) doSlow(f func()) {
		o.m.Lock()
		defer o.m.Unlock()
		if o.done == 0 { // 再次进行o.done是否等于0判断，因为存在并发调用doSlow的情况
			defer atomic.StoreUint32(&o.done, 1) // 将o.done值设置为1，用来标志操作完成
			f() // 执行操作
		}
	}
```