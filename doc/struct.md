[toc]

# struct
struct类似对象语言的一个类, 是属性和方法的集合; 与类不同的是, struct没有固有的反射metadata和初始化等, 但比类更简洁高效.
## 方法
- 指针方法对**属性值的改变**会反应在变量本身
- 非指针方法对**属性值的改变**不会反应在变量本身，需要返回一个新实例才能查看属性的变化

示例1
```go
	type Foo struct {
		a int
		b int
	}
	// 非指针赋值，需要返回新实例才能查看a的变化
	func (f Foo) SetA(a int) Foo {
		f.a = a
		return f
	}
	// 指针赋值，不需要返回一个新实例
	func (f *Foo) SetB(b int) {
		f.b = b
	}
	// 输出: {0 2} {1 3}
	func Test(t *testing.T) {
		f := Foo{}
		f2 := f.SetA(1) // 返回新实例
		f.SetB(2)
		f2.SetB(3)

		fmt.Println(f, f2)
	}
```
## 属性
### 大小写
用首字母的大小写来区分<font color=yellow>公有属性和私有属性</font>
- 公有属性: 可以被其他package直接引用
- 私有属性: 只能被所在package直接引用
```go
	type T struct{
		A int  // 公有
		b int  // 私有
	}
```
### 继承
在struct可以直接继承其他struct/interface类型
1. 继承Value形式的struct, 在方法体中修改属性值, 无法反应到实例中
2. 继承Pointer形式的struct, 在方法体中修改属性值, 可以反应到实例中
3. 继承interface只能通过方法调用去修改属性

示例2
```go
	type I interface {
		String() string
	}
	// I的实现者
	type Imp struct{}
	func (i Imp) String() string {
		return "Imp"
	}
	
	type A struct {
		Name string
		Age  int
		I	// 继承I
	}
	type AB struct {
		A	// 继承A
		Gender bool
	}
	type AC struct {
		*A	// 继承*A
		Age int32
	}

	func TestStruct2(t *testing.T) {
		f1 := func(b AB, c AC, age int) {
			b.A.Age = age
			c.A.Age = age
		}

		i := Imp{} // 实例化I
		b := AB{Gender: true, A: A{Name: "R1", Age: 1, I: i}}
		c := AC{Age: 187, A: &A{Name: "R2", Age: 2}}
		f1(b, c, 3)

		fmt.Println(b.A.Age, c.A.Age, b.A.I.String())
	}
```
### tag
属性的tag是对**struct的编/解码**时的规范
```go
	type Tag struct {
		Name string `json:"name,optional,omitempty,default=abc" xml:"name"`
	}
```
使用tag来指定encoding到json后, 各个field在json中的key名称/可选/可忽略/默认值等信息, 可使用reflect包来读取tag信息

### 空struct
struct{}直接定义一个空的结构体并没有意义，但在并发编程中，channel之间的通讯，可以使用一个struct{}作为信号量, <font color=yellow>空struct的实例不占内存</font>
```go
	ch := make(chan struct{}) 
	ch <- struct{}{}
	
	es := struct{}{}
	fmt.Println(unsafe.Sizeof(es)) // 输出 0
```