# interface
- 1. interface本质是声明了零或多个行为方法的 <font color=yellow>数据类型</font>，struct也是一种<font color=yellow>数据类型</font>, 不仅包含行为方法的实现，还可以包含<font color=yellow>数据属性</font>
- 2. 实现: <font color=red>由type关键字声明的类型(struct/go基本类型等)都是一种数据类型</font>, 如果这个类型实现了interface中声明的所有方法, 那么可以称这个type类型是interface的实现者. 对于零个方法的interface类型, 被称为empty interface, 任何type声明的类型都是实现者
- 3. 存储: interface变量存储实现者的值
- 4. 值类型: go的运行时会自动检测interface变量存储的值类型
	 获取empty interface的值时, 通常需要类型断言, 形式: value, ok := in.(int), in是interface变量
	 不确定empty interface变量的值类型时, 通常使用switch-type来判断, 形式: switch in.(type){...} 或 switch v := in.(type){...}
## 实现
interface声明一组方法, 有type关键字声明的类型需要对interface的所有方法进行实现

示例1
```go
	// 一组方法的声明
	type T interface {
		Set(int)
		Get() int
	}
	// type声明的Tx类型, 其属性值的内存是struct存储格式
	type Tx struct {
		A int
	}
	func (t *Tx) Set(a int) {
		t.A = a
	}
	func (t Tx) Get() int {
		return t.A
	}
	
	func TestInterface1(t *testing.T) {
		tx := Tx{A: 2}
		f := func(in T) { 
			 in.Set(123)
        	fmt.Println(in.Get())
    	}
		f(&tx)
	}
```

示例11
```go
	type In interface {
		Set(int)
		Get() int
	}
	// type声明的Int类型, 其属性值的内存是int存储格式
	type Int int
	func (i *Int) Set(a int) {
		*i = *(*Int)(unsafe.Pointer(&a))
	}
	func (i *Int) Get() int {
		return int(*i)
	}
	
	func TestInterface11(t *testing.T) {
		var i Int = 2
		f := func(in In) {
			// ① 无法更改属性值 in=2
			in.Set(123)
			fmt.Println(in.Get())
		}
		f(&i)
	}
```
type声明的基本类型实现In接口, 如果基本类型接口方法(如Set)会修改属性值时
- 如果是指针形式接收方法<font color=red>func (i *Int)Set(a int)</font>需要用到unsafe.Pointer强转; 
因为不像示例1中的Tx struct, 可以拥有Tx.A属性, 基本类型的属性是它本身, 如果是指针形式接收方法, 就相当于两个不同的指针类型转换p1 *int = p2 *Int, 这是不允许的. 
- 如果采用值形式接收方法, 则重新返回一个实例 <font color=red>func Set(int) In</font>
### 实现者的receiver
receiver是指实现者方法的绑定, 它的形式分为:
- 值绑定(Value receiver): <font color=yellow>func (t T)F(){}</font>, 绑定的变量形式是值形式<font color=yellow>t T</font>
- 指针绑定(Pointer receiver): <font color=yellow>func (t *T)F(){}</font>, 绑定的变量是指针形式<font color=yellow>t *T</font>

---
在示例1中, 使用了<font color=yellow>f(&tx)</font>, 如果使用<font color=yellow>f(tx)</font>形式则会报错
```
cannot use tx (variable of type Tx) as type T in argument to func(in T) {…}:
```
分析: Tx实现者的Set方法receiver是个Pointer, Get方法receiver是个Value, 在f(tx)传递值的拷贝时,不满足Set receiver必须是个Pointer. 而f(&tx)传递值的拷贝时, 实际上传递的是指针, 会指向值, <font color=red>**go 会把指针进行隐式转换得到value，但反过来则不行**</font>
如果Set方法改成如下形式, f(tx)和f(&tx)都不会报错
```go
	// tx是value形式, 赋值实际上不起作用, 但不会报错
	func (t Tx) Set(a int) {
		t.A = a 
	}
```
**Value形式的属性赋值, 实际上不起作用. Pointer形式的属性赋值会被改变**, 如示例1

示例2
```go
	// 两个实现者, 一个采用Pointer形式, 一个采用Value形式
	type M interface {
		Show() string
	}
	type Mx struct{}
	func (m *Mx) Show() string {
		return "Mx"
	}
	type My struct{}
	func (m My) Show() string {
		return "My"
	}
	
	func TestInterface2(t *testing.T) {
		f := func(in []M) {    
			 for _, m := range ary {
			 	fmt.Println(m.Show())
			 }
    	}
		// ①
		ary := []M{Mx{}, My{}}
		f(ary)
		// ②
		ary2 := []M{&Mx{}, &My{}}
		f(ary2)
	}
```
① 输出报错, 因为Mx的receiver是Pointer, 而在初始化ary时, Mx是Value; ② 输出正常, 因为ary2初始化Pointer, 而Pointer可以隐式转化为Value
## 存储
interface变量存储的是实现者的值, 而empty interface变量可以存储任何类型的值

示例3
```go
	var m M = Mx{}
	var i interface{} = 1
```

## 值类型
go的运行时会自动判断interface变量的值类型. 
- 对于empty interface的值, 使用类型断言形式

示例4
```go
	func TestInterface4(t *testing.T) {
		var in interface{} = 1
		// ①
		fmt.Println(int(in)+1) 
		
		// ②
		if v, ok := in.(int); !ok {
            fmt.Errorf("not ok")
        } else {
            fmt.Println(v + 2)
        }
	}
```
① int(in)并不能强制类型转换, 会报错. ② 类型断言可以成功转换类型

- 不确定empty interface的值类型, 使用switch-type

示例5
```go
	type G interface {
		String() string
	}
	type Gx struct{}
	func (g *Gx) String() string {
		return "Gx"
	}
	func Type(in interface{}) string {
		var typeStr string
		switch in.(type) {
		case int:
			typeStr = "int"
		case string:
			typeStr = "string"
		case float64:
			typeStr = "float64"
		case G:	
			typeStr = "G"
		case Gx:
			typeStr = "Gx"
		default:
			typeStr = "unknown"
		}
		return typeStr
	}
	
	func TestInterface5(t *testing.T) {
		var ary []interface{} = []interface{}{
			1,			// 输出 int
			"abc",		// 输出 string
			2.3,        // 输出 float64
			Gx{},		// 输出 Gx, 因为Value类型并不是G的实现者, 而是Gx变量
			&Gx{},		// 输出 G, 指针型Gx是G的实现者
		}
		f := func(ins []interface{}) {
			for idx, in := range ins {
				fmt.Println(idx, Type(in))
			}
		}
		f(ary)
	}
```
# 总结
1. interface实现者作实参时, 尽量传递指针形式
2. 对基本类型的interface实现者, 在需要修改属性值时, 返回新实例
3. 对empty interface, 使用类型断言获取值, 使用Switch-type获取类型