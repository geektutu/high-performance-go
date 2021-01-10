---
title: Go 死码消除与调试(debug)模式
seo_title: Go 语言高性能编程
date: 2021-01-11 01:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。本文介绍了编译器在死码消除(Dead code elimination, DCE) 方面的优化，在实际编程中如何利用这一优化提高程序性能。并结合构建标记(build tags) 增加调试模式。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 编译优化
keywords:
- golang
- 死码消除
- 死代码
- Dead code elimination
image: post/hpg-reduce-size/compiler.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: 死码消除与调试模式
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-4/hpg-dead-code-elimination.md
---

![golang compiler optimization](hpg-reduce-size/compiler.jpg)

## 1 什么是死码消除

以下摘自内容 [Dead code elimination - wikipedia](https://en.wikipedia.org/wiki/Dead_code_elimination)

> In compiler theory, dead code elimination (also known as DCE, dead code removal, dead code stripping, or dead code strip) is a compiler optimization to remove code which does not affect the program results. 

死码消除(dead code elimination, DCE)是一种编译器优化技术，用处是在编译阶段去掉对程序运行结果没有任何影响的代码。

> Removing such code has several benefits: it shrinks program size, an important consideration in some contexts, and it allows the running program to avoid executing irrelevant operations, which reduces its running time.

死码消除有很多好处：减小程序体积，程序运行过程中避免执行无用的指令，缩短运行时间。

## 2 Go 语言中的应用

### 2.1 使用常量提升性能

在某些场景下，将变量替换为常量，性能会有很大的提升。

举一个简单的例子，以下是 `maxvar.go` 的代码：

```go
// maxvar.go
func max(num1, num2 int) int {
	if num1 > num2 {
		return num1
	}
	return num2
}

var a, b = 10, 20

func main() {
	if max(a, b) == a {
		fmt.Println(a)
	}
}
```

- max 是一个非常简单的函数，返回两个值中的较大值。
- a 和 b 是两个全局变量，赋值为 10 和 20。
- 如果 a 大于 b，那么将会调用 time.Sleep() 休眠 3 秒。

拷贝 `maxvar.go` 为 `maxconst.go`，并将 `var a, b` 修改为 `const a, b`。

```go
// maxconst.go
func max(num1, num2 int) int {
	if num1 > num2 {
		return num1
	}
	return num2
}

const a, b = 10, 20

func main() {
	if max(a, b) == a {
		fmt.Println(a)
	}
}
```

编译 `maxvar.go` 和 `maxconst.go`，并比较编译后的二进制大小：

```bash
go build -o maxvar maxvar.go
go build -o maxconst maxconst.go
ls -l maxvar maxconst
-rwxr-xr-x  1 x x 1895424 Jan 10 00:01 maxconst
-rwxr-xr-x  1 x x 2120368 Jan 10 00:01 maxvar
```

我们可以看到 `maxconst` 比 `maxvar` 体积小了约 10% = 0.22 MB。

为什么会出现 11% 的差异呢？

我们使用 `-gcflags=-m` 参数看一下编译器做了哪些优化：

```bash
go build -gcflags=-m  -o maxvar maxvar.go
# command-line-arguments
./maxconst.go:7:6: can inline max
./maxconst.go:17:8: inlining call to max
```

max 函数被内联了，即被展开了，手动展开后如下：

```go
func main() {
	var result int
	if a > b {
		result = a
	} else {
		result = b
    }
	if result == a {
		fmt.Println(a)
	}
}
```

那如果 a 和 b 均为常量（const）呢？那在编译阶段就可以直接进行计算：

```go
func main() {
	var result int
	if 10 > 20 {
		result = 10
	} else {
		result = 20
    }
	if result == 10 {
		fmt.Println(a)
	}
}
```

计算之后，`10 > 20` 永远为假，那么分支消除后：

```go
func main() {
	if 20 == 10 {
		fmt.Println(a)
	}
}
```

进一步，`20 == 10` 也永远为假，再次分支消除：

```go
func main() {}
```

但是如果全局变量 a、b 不为常量，即 `maxvar` 中声明的一样，编译器并不知道运行过程中 a、b 会不会发生改变，因此不能够进行死码消除，这部分代码被编译到最终的二进制程序中。因此 `maxvar` 比 `maxconst` 二进制体积大了约 10%。

如果在 if 语句中，调用了更多的库，死码消除之后，体积差距会更大。

因此，在声明全局变量时，如果能够确定为常量，尽量使用 const 而非 var，这样很多运算在编译器即可执行。死码消除后，既减小了二进制的体积，又可以提高运行时的效率，如果这部分代码是 `hot path`，那么对性能的提升会更加明显。

### 2.2 可推断的局部变量

考虑另一种情况，a、b 作为局部变量呢？

```go
// maxvarlocal
func main() {
	var a, b = 10, 20
	if max(a, b) == a {
		fmt.Println(a)
	}
}
```

编译结果如下，大小与 `varconst` 一致，即 a、b 作为局部变量时，编译器死码消除是生效的。

```bash
$ go build -o maxvarlocal maxvarlocal.go
$ ls -l maxvarlocal                      
-rwxr-xr-x  1 x x 1895424 Jan 10 00:05 maxvarlocal
```

那如果再修改一下，函数中增加修改 a、b 变量的并发操作。

```go
func main() {
	var a, b = 10, 20
	go func() {
		b, a = a, b
	}()
	if max(a, b) == a {
		fmt.Println(a)
	}
}
```

编译结果如下，大小增加了 10%，此时，a、b 的值不能有效推断，死码消除失效。

```bash
$ go build -o maxvarlocal maxvarlocal.go
$ ls -l maxvarlocal                      
-rwxr-xr-x  1 x x 2120352 Jan 10 00:05 maxvarlocal
```

其实这个结果很好理解，包(package)级别的变量和函数内部的局部变量的推断难度是不一样的。函数内部的局部变量的修改只会发生在该函数中。但是如果是包级别的变量，对该变量的修改可能出现在：

- 包初始化函数 init() 中，init() 函数可能有多个，且可能位于不同的 `.go` 源文件。
- 包内的其他函数。
- 如果是 public 变量（首字母大写），其他包引用时可修改。

推断 package 级别的变量是否被修改难度是非常大的，从上述的例子看，Go 编译器只对局部变量作了优化。

> 以上例子，基于 go1.13.6 darwin/amd64

### 2.3 调试(debug)模式

我们可以在源代码中，定义全局常量 debug，值设置为 `false`，在需要增加调试代码的地方，使用条件语句 `if debug` 包裹，例如下面的例子：

```go
const debug = false

func main() {
	if debug {
		log.Println("debug mode is enabled")
	}
}
```

如果是正常编译，常量 debug 始终等于 `false`，调试语句在编译过程中会被消除，不会影响最终的二进制大小，也不会对运行效率产生任何影响。

那如果我们想编译出 debug 版本的二进制呢？可以将 debug 修改为 true 之后编译。这对于开发者日常调试是非常有帮助的，日常开发过程中，在进行单元测试或者是简单的集成测试时，希望能够执行一些额外的操作，例如打印日志，或者是修改变量的值。提交代码时，再将 debug 修改为 false，开发过程中增加的额外的调试代码在编译时会被消除，不会对正式版本产生任何的影响。

Go 语言源代码中有很多这样的例子：

```bash
$ grep -nr "const debug = false" "$(dirname $(which go))/../src"
/usr/local/go/bin/../src/cmd/go/internal/modfile/read.go:606:   const debug = false
/usr/local/go/bin/../src/cmd/compile/internal/syntax/parser.go:14:const debug = false
#  ...
/usr/local/go/bin/../src/net/http/transport_test.go:2037:       const debug = false
/usr/local/go/bin/../src/net/http/transport_test.go:2095:       const debug = false
/usr/local/go/bin/../src/go/types/initorder.go:23:      const debug = false
/usr/local/go/bin/../src/go/internal/gcimporter/gcimporter.go:22:const debug = false
```

### 2.4 条件编译

有没有不修改源代码，也能编译出 debug 版本的方式呢？

答案是肯定的：有，可结合 build tags 来实现条件编译。

新建 `release.go` 和 `debug.go`：

- debug.go

```go
// +build debug

package main

const debug = true
```

- release.go

```go
// +build !debug

package main

const debug = false
```

在 `main.go` 中去掉常量 debug 的定义：

```go
package main

import "log"

func main() {
	if debug {
		log.Println("debug mode is enabled")
	}
}
```

- `// +build debug` 表示 build tags 中包含 debug 时，该源文件参与编译。
- `// +build !debug` 表示 build tags 中不包含 debug 时，该源文件参与编译。

一个源文件中可以有多个 build tags，同一行的空格隔开的 tag 之间是逻辑或的关系，不同行之间的 tag 是逻辑与的关系。例如下面的写法表示：此源文件只能在 linux/386 或者 darwin/386 平台下编译。

```go
// +build linux darwin
// +build 386
```

接下来，我们编译一个 debug 版本并运行：

```bash
$ go build -tags debug -o debug .  
$ ./debug 
2021/01/11 00:10:40 debug mode is enabled
```

编译 release 版本并运行：

```bash
$ go build -o release .
$ ./release
# no output
```

除了全局布尔值常量 `debug` 以外，`debug.go` 和 `release.go` 还可以根据需要添加其他代码。例如，相同的函数定义，debug 和 release 模式下有不同的函数实现。

## 附 推荐与参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)

