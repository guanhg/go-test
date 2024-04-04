[toc]

# defer
defer用于资源的释放, 会在函数返回之前进行调用, 一般用来完成资源、锁、连接等释放工作，或者 `recover` 可能发生的`panic`

## 特性
1. defer函数是按照后进先出的顺序执行
```golang
	//  输出54321
	func main() {
		for i := 1; i <= 5; i++ {
			defer fmt.Print(i)
		}
	}
```
2. defer函数的传入参数在定义时就已经明确
```go
	// 输出 2 1
	func main() {
		i := 1
		defer fmt.Println(i) // 传入1
		i = 2
		defer fmt.Println(i) // 传入2
		return
	}
```
3. defer函数可以读取和修改函数的命名返回值
```go
	// 输出 101
	func main() {
		fmt.Println(test())
	}

	func test() (i int) {
		defer func() {
			i++
		}()
		return 100
	}
```
## 分析
由于**return xxx这一条语句并不是一条原子指令!** 
- 函数返回的过程是这样的：先给返回值赋值，然后调用defer表达式，最后才是返回到调用函数中。
  等效如下
```
返回值 = xxx
调用defer函数
空的return
```
  <font color=yellow>这种等效需要注意上面的第2个特性</font>
```go
	func cacl(idx string, x, y int) int {
		ret := x + y
		fmt.Println(idx, x, y, ret)
		return ret
	}

	func main() {
		a := 2
		b := 4
		defer cacl("1", a, cacl("15", a, b))
		a = 1
		defer cacl("2", a, cacl("30", a, b))
	}
```
等效
```
	tmp1 := cacl("15", 2, 4)
	tmp2 := cacl("30", 1, 4)
	cacl("2", 1, tmp2)
	cacl("1", 2, tmp1)
```
