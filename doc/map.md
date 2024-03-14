# map
存储K-V数据形式的结构
## 声明
```go 
var mp map[string]int      // 第一种 nil map
mp := make(map[string]int) // 第二种 如果指定长度 mp := make(map[string]int, 3)
mp := map[string]int{}		// 第三种 等效第二种
```
1. 第一种声明没有初始化，是`nil map`，不能插入数据，因为`存储数据的指针为nil`
2. 第二种声明比较常用，因为已初始化空间长度，可以插入数据

实例1
```go
func TestMap1(t *testing.T) {
    var mp map[string]int		// 声明1，nil map
    mp["1"] = 1 // 报错
    mp2 := make(map[string]int)// 声明2，初始化
    mp2["2"] = 2
    fmt.Println(mp["abc"], mp2["abc"])
}
```
## 获取
两种获取方式
```
mp := make(map[string]int)
val := mp["abc"]				// 第一种
for k, v := range mp {...}		// 第二种
```
## 实现
数据结构
```go
type hmap struct {
   count     int 	// 数据长度
   flags     uint8
   B         uint8  // 2^B个bucket
   noverflow uint16 // 溢出buckets的数量
   hash0     uint32 // hash seed

   buckets    unsafe.Pointer // bucket数组，长度为2^B，如果B=0,则为nil
   oldbuckets unsafe.Pointer // 用于扩容
   nevacuate  uintptr      

   extra *mapextra // optional fields 溢出桶
}

// bucket 运行时重构结构
type bmap struct {
    topbits  [8]uint8		// hash(高位)值
    keys     [8]keytype		// 每个bucket存8个key
    values   [8]valuetype	// 每个bucket存8个value
    pad      uintptr 
    overflow uintptr		// 指向溢出桶
}
```
- hmap.extra溢出桶：
> 如果具有相同hash(高位)值的Key数量过多`即超过8个`，也会存放在溢出同种，溢出桶的内存空间紧跟在正常桶的后面
- hash(key)值
> hash值的低位是用来确定key在哪个桶，经过位运算`hash(key)&hmap.B`来确定
> hash值的高8位是用来查找桶中的key，topbits保存hash(key)的高8为的值，是为了快速查找

### 初始化
![image.png](https://p1-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/aab6faae364c4b309c5560bb44f734f4~tplv-k3u1fbpfcp-jj-mark:3024:0:0:0:q75.awebp#?w=1656&h=794&s=415982&e=png&b=fffefe)

- 初始化流程
> 1. new(hmap)，并确定一个hash0
> 2. 根据hint(桶多少)计算B的大小，例如hint=10则B=2
> 3. 申请正常桶和溢出桶的内存空间
>> **注意: 正常桶空间是连续的即数组；尾部紧跟溢出桶空间，也是数组，长度为 hint-2^B**

rumtime.makemap
```go
func makemap(t *maptype, hint int, h *hmap) *hmap {
	// 判断需要的内存大小是否溢出，如果溢出就不生成bucket
   mem, overflow := math.MulUintptr(uintptr(hint), t.bucket.size)
   if overflow || mem > maxAlloc {
      hint = 0
   }
   
   if h == nil {
      h = new(hmap)
   }
   h.hash0 = fastrand()  // hash种子

   // 根据hint找到合适B，例如hint=10,则B=2
   B := uint8(0)
   for overLoadFactor(hint, B) {
      B++
   }
   h.B = B

   // allocate initial hash table
   // if B == 0, the buckets field is allocated lazily later (in mapassign)
   // If hint is large zeroing this memory could take a while.
   if h.B != 0 {
      var nextOverflow *bmap
	  // 分配内存空间，正常桶空间是连续的即数组；紧跟溢出桶空间，也是数组，长度为 hint-2^B
      h.buckets, nextOverflow = makeBucketArray(t, h.B, nil)
      if nextOverflow != nil {
         h.extra = new(mapextra)
         h.extra.nextOverflow = nextOverflow
      }
   }

   return h
}
```
### 写入
写入流程
> 1. 计算hash(key)的值
> 2. hash值后B位确认写入的桶bucket，高8位确定为hashtop
> 3. 遍历已经确认的正常桶bucket的topbits数组以及该桶关联的溢出桶bucket.overflow的topbits数组。如果桶中没有相等的hashtop，找一个空位写入KV；如果有相同的hashtop，则比较key的值是否相等，如果相等则更新value，否则继续查找空位写入。

(源码查看runtime.mapassign，易懂)

### 读取
读取流程
> 1. 计算hash(key)的值
> 2. hash值后B位确认读取的桶bucket，高8位确定为hashtop 
> 3. 遍历确认的桶bucket的topbits数组以及该桶关联的溢出桶bucket.overflow的topbits数组。如果存在hashtop值，则比较key的值，如果相同就读取，否则继续往下遍历

(源码查看runtime.mapaccess1，易懂)

## 扩容
- 触发条件
>  1. 装载因子 count/2^B > 6.5
>  2. 溢出桶过多：a. B<=15，已使用的溢出桶个数>=`2^B`; b. B>15,已使用的溢出桶个数>=`2^15`

- 扩容流程
> 1. 计算新的B值，从oldoverflow上获取需要扩容的桶bucket
> 2. 获取新的内存空间
> 3. 变流bucket的KV，计算新的hash(key)，并迁移数据
> > **注意：为了防止耗时过长，每次扩容只会扩容一或两个桶，开始扩容时会把 oldoverflow=buckets，再次进入扩容工作只需要判断 oldoverflow!=nil**

## 总结
- 通过key的哈希值将key散落到不同的桶中，每个桶中有8个cell。哈希值的低B位决定桶序号，高8位标识同一个桶中的不同key。

- 当向桶中添加了很多key造成元素过多，或者溢出桶太多，就会触发扩容。扩容分为等量扩容和翻倍容量扩容。扩容后，原来一个 bucket中的key一分为二，会被重新分配到两个桶中。

- 查找、赋值、删除的一个很核心的内容是如何定位到key所在的位置，需要重点理解。一旦理解，关于map的源码就可以看懂了

- 扩容过程是渐进的，主要是防止一次扩容需要搬迁的key数量过多，引发性能问题。触发扩容的时机是增加了新元素，bucket搬迁的时机则发生在赋值、删除期间，每次最多搬迁2个bucket

- 遍历过程，for range每次循环都会通过hiter结构去获取cell，并且随机从一个桶的一个cell开始遍历，若遍历的桶处于翻倍扩容时，只会遍历翻倍后到本桶的cell

> [参考] (https://juejin.cn/post/7314509615159935027?searchId=20240312022725ACE5A854EA390D0331F2)