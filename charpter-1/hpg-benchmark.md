---
title: benchmark 基准测试
seo_title: Go 语言高性能编程
date: 2020-11-17 01:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。详细介绍如何测试/评估 Go 代码的性能，内容包括使用 testing 库进行基准测试(benchmark)，性能分析(profiling) 编译优化(compiler optimisations)，内存管理(memory management)和垃圾回收(garbage collect)、pprof 等内容。同时也介绍了使用 Go 语言如何写出高性能的程序和应用，包括不限于 Go 语言标准库、第三方库的使用方式和最佳实践。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 性能分析
keywords:
- golang
- benchmark
- 性能分析
image: post/hpg-benchmark/benchmark.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: benchmark 基准测试
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-1/hpg-benchmark.md
---

![benchmark & profiling - high performance with go](hpg-benchmark/benchmark.jpg)

## 1 稳定的测试环境

当我们尝试去优化代码的性能时，首先得知道当前的性能怎么样。Go 语言标准库内置的 testing 测试框架提供了基准测试(benchmark)的能力，能让我们很容易地对某一段代码进行性能测试。

性能测试受环境的影响很大，为了保证测试的可重复性，在进行性能测试时，尽可能地保持测试环境的稳定。

- 机器处于闲置状态，测试时不要执行其他任务，也不要和其他人共享硬件资源。
- 机器是否关闭了节能模式，一般笔记本会默认打开这个模式，测试时关闭。
- 避免使用虚拟机和云主机进行测试，一般情况下，为了尽可能地提高资源的利用率，虚拟机和云主机 CPU 和内存一般会超分配，超分机器的性能表现会非常地不稳定。

> 超分配是针对硬件资源来说的，商业上对应的就是云主机的超卖。虚拟化技术带来的最大直接收益是服务器整合，通过 CPU、内存、存储、网络的超分配（Overcommitment）技术，最大化服务器的使用率。例如，虚拟化的技能之一就是随心所欲的操控 CPU，例如一台 32U(物理核心)的服务器可能会创建出 128 个 1U(虚拟核心)的虚拟机，当物理服务器资源闲置时，CPU 超分配一般不会对虚拟机上的业务产生明显影响，但如果大部分虚拟机都处于繁忙状态时，那么各个虚拟机为了获得物理服务器的资源就要相互竞争，相互等待。Linux 上专门有一个指标，Steal Time(st)，用来衡量被虚拟机监视器(Hypervisor)偷去给其它虚拟机使用的 CPU 时间所占的比例。


## 2 benchmark 的使用

### 2.1 一个简单的例子

Go 语言标准库内置了支持 benchmark 的 `testing` 库，接下来看一个简单的例子：

使用 `go mod init example` 初始化一个模块，新增 `fib.go` 文件，实现函数 `fib`，用于计算第 N 个菲波那切数。

```go
// fib.go
package main

func fib(n int) int {
	if n == 0 || n == 1 {
		return n
	}
	return fib(n-2) + fib(n-1)
}
```

接下来，我们在 `fib_test.go` 中实现一个 benchmark 用例：

```go
// fib_test.go
package main

import "testing"

func BenchmarkFib(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fib(30) // run fib(30) b.N times
	}
}
```

- benchmark 和普通的单元测试用例一样，都位于 `_test.go` 文件中。
- 函数名以 `Benchmark` 开头，参数是 `b *testing.B`。和普通的单元测试用例很像，单元测试函数名以 `Test` 开头，参数是 `t *testing.T`。

### 2.2 运行用例

`go test <module name>/<package name>` 用来运行某个 package 内的所有测试用例。

- 运行当前 package 内的用例：`go test example` 或 `go test .`
- 运行子 package 内的用例： `go test example/<package name>` 或 `go test ./<package name>`
- 如果想递归测试当前目录下的所有的 package：`go test ./...` 或 `go test example/...`。

`go test` 命令默认不运行 benchmark 用例的，如果我们想运行 benchmark 用例，则需要加上 `-bench` 参数。例如：

```bash
$ go test -bench .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-8               200           5865240 ns/op
PASS
ok      example 1.782s
```

`-bench` 参数支持传入一个正则表达式，匹配到的用例才会得到执行，例如，只运行以 `Fib` 结尾的 benchmark 用例：

```bash
$ go test -bench='Fib$' .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-8               202           5980669 ns/op
PASS
ok      example 1.813s
```

### 2.3 benchmark 是如何工作的

benchmark 用例的参数 `b *testing.B`，有个属性 `b.N` 表示这个用例需要运行的次数。`b.N` 对于每个用例都是不一样的。

那这个值是如何决定的呢？`b.N` 从 1 开始，如果该用例能够在 1s 内完成，`b.N` 的值便会增加，再次执行。`b.N` 的值大概以 1, 2, 3, 5, 10, 20, 30, 50, 100 这样的序列递增，越到后面，增加得越快。我们仔细观察上述例子的输出：

```bash
BenchmarkFib-8               202           5980669 ns/op
```

BenchmarkFib-8 中的 `-8` 即 `GOMAXPROCS`，默认等于 CPU 核数。可以通过 `-cpu` 参数改变 `GOMAXPROCS`，`-cpu` 支持传入一个列表作为参数，例如：

```bash
$ go test -bench='Fib$' -cpu=2,4 .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-2               206           5774888 ns/op
BenchmarkFib-4               205           5799426 ns/op
PASS
ok      example 3.563s
```

在这个例子中，改变 CPU 的核数对结果几乎没有影响，因为这个 Fib 的调用是串行的。

`202` 和 `5980669 ns/op` 表示用例执行了 202 次，每次花费约 0.006s。总耗时比 1s 略多。

### 2.4 提升准确度

对于性能测试来说，提升测试准确度的一个重要手段就是增加测试的次数。我们可以使用 `-benchtime` 和 `-count` 两个参数达到这个目的。

benchmark 的默认时间是 1s，那么我们可以使用 `-benchtime` 指定为 5s。例如：

```bash
$ go test -bench='Fib$' -benchtime=5s .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-8              1033           5769818 ns/op
PASS
ok      example 6.554s
```

> 实际执行的时间是 6.5s，比 benchtime 的 5s 要长，测试用例编译、执行、销毁等是需要时间的。

将 `-benchtime` 设置为 5s，用例执行次数也变成了原来的 5倍，每次函数调用时间仍为 0.6s，几乎没有变化。

`-benchtime` 的值除了是时间外，还可以是具体的次数。例如，执行 30 次可以用 `-benchtime=30x`：

```bash
$ go test -bench='Fib$' -benchtime=50x .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-8                50           6121066 ns/op
PASS
ok      example 0.319s
```

调用 50 次 `fib(30)`，仅花费了 0.319s。

`-count` 参数可以用来设置 benchmark 的轮数。例如，进行 3 轮 benchmark。

```bash
$ go test -bench='Fib$' -benchtime=5s -count=3 .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-8               975           5946624 ns/op
BenchmarkFib-8              1023           5820582 ns/op
BenchmarkFib-8               961           6096816 ns/op
PASS
ok      example 19.463s
```

### 2.5 内存分配情况

`-benchmem` 参数可以度量内存分配的次数。内存分配次数也性能也是息息相关的，例如不合理的切片容量，将导致内存重新分配，带来不必要的开销。

在下面的例子中，`generateWithCap` 和 `generate` 的作用是一致的，生成一组长度为 n 的随机序列。唯一的不同在于，`generateWithCap` 创建切片时，将切片的容量(capacity)设置为 n，这样切片就会一次性申请 n 个整数所需的内存。

```go
// generate_test.go
package main

import (
	"math/rand"
	"testing"
	"time"
)

func generateWithCap(n int) []int {
	rand.Seed(time.Now().UnixNano())
	nums := make([]int, 0, n)
	for i := 0; i < n; i++ {
		nums = append(nums, rand.Int())
	}
	return nums
}

func generate(n int) []int {
	rand.Seed(time.Now().UnixNano())
	nums := make([]int, 0)
	for i := 0; i < n; i++ {
		nums = append(nums, rand.Int())
	}
	return nums
}

func BenchmarkGenerateWithCap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		generateWithCap(1000000)
	}
}

func BenchmarkGenerate(b *testing.B) {
	for n := 0; n < b.N; n++ {
		generate(1000000)
	}
}
```

运行该用例的结果是：

```bash
go test -bench='Generate' .
goos: darwin
goarch: amd64
pkg: example
BenchmarkGenerateWithCap-8            44          24294582 ns/op
BenchmarkGenerate-8                   34          30342763 ns/op
PASS
ok      example 2.171s
```

可以看到生成 100w 个数字的随机序列，`GenerateWithCap` 的耗时比 `Generate` 少 20%。

我们可以使用 `-benchmem` 参数看到内存分配的情况：

```bash
goos: darwin
goarch: amd64
pkg: example
BenchmarkGenerateWithCap-8  43  24335658 ns/op  8003641 B/op    1 allocs/op
BenchmarkGenerate-8         33  30403687 ns/op  45188395 B/op  40 allocs/op
PASS
ok      example 2.121s
```

`Generate` 分配的内存是 `GenerateWithCap` 的 6 倍，设置了切片容量，内存只分配一次，而不设置切片容量，内存分配了 40 次。

### 2.6 测试不同的输入

不同的函数复杂度不同，O(1)，O(n)，O(n^2) 等，利用 benchmark 验证复杂度一个简单的方式，是构造不同的输入。对刚才的 benchmark 稍作改造，便能够达到目的。

```go
// generate_test.go
package main

import (
	"math/rand"
	"testing"
	"time"
)

func generate(n int) []int {
	rand.Seed(time.Now().UnixNano())
	nums := make([]int, 0)
	for i := 0; i < n; i++ {
		nums = append(nums, rand.Int())
	}
	return nums
}
func benchmarkGenerate(i int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		generate(i)
	}
}

func BenchmarkGenerate1000(b *testing.B)    { benchmarkGenerate(1000, b) }
func BenchmarkGenerate10000(b *testing.B)   { benchmarkGenerate(10000, b) }
func BenchmarkGenerate100000(b *testing.B)  { benchmarkGenerate(100000, b) }
func BenchmarkGenerate1000000(b *testing.B) { benchmarkGenerate(1000000, b) }
```

这里，我们实现一个辅助函数 `benchmarkGenerate` 允许传入参数 i，并构造了 4 个不同输入的 benchmark 用例。运行结果如下：

```bash
$ go test -bench .                                                       
goos: darwin
goarch: amd64
pkg: example
BenchmarkGenerate1000-8            34048             34643 ns/op
BenchmarkGenerate10000-8            4070            295642 ns/op
BenchmarkGenerate100000-8            403           3230415 ns/op
BenchmarkGenerate1000000-8            39          32083701 ns/op
PASS
ok      example 6.597s
```

通过测试结果可以发现，输入变为原来的 10 倍，函数每次调用的时长也差不多是原来的 10 倍，这说明复杂度是线性的。

## 3 benchmark 注意事项

### 3.1 ResetTimer

如果在 benchmark 开始前，需要一些准备工作，如果准备工作比较耗时，则需要将这部分代码的耗时忽略掉。比如下面的例子：

```go
func BenchmarkFib(b *testing.B) {
	time.Sleep(time.Second * 3) // 模拟耗时准备任务
	for n := 0; n < b.N; n++ {
		fib(30) // run fib(30) b.N times
	}
}
```
运行结果是：

```bash
$ go test -bench='Fib$' -benchtime=50x .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-8                50          65912552 ns/op
PASS
ok      example 6.319s
```

50次调用，每次调用约 0.66s，是之前的 0.06s 的 11 倍。究其原因，受到了耗时准备任务的干扰。我们需要用 `ResetTimer` 屏蔽掉：

```go
func BenchmarkFib(b *testing.B) {
	time.Sleep(time.Second * 3) // 模拟耗时准备任务
	b.ResetTimer() // 重置定时器
	for n := 0; n < b.N; n++ {
		fib(30) // run fib(30) b.N times
	}
}
```

运行结果恢复正常，每次调用约 0.06s。

```bash
$ go test -bench='Fib$' -benchtime=50x .
goos: darwin
goarch: amd64
pkg: example
BenchmarkFib-8                50           6187485 ns/op
PASS
ok      example 6.330s
```

### 3.2 StopTimer & StartTimer

还有一种情况，每次函数调用前后需要一些准备工作和清理工作，我们可以使用 `StopTimer` 暂停计时以及使用 `StartTimer` 开始计时。

例如，如果测试一个冒泡函数的性能，每次调用冒泡函数前，需要随机生成一个数字序列，这是非常耗时的操作，这种场景下，就需要使用 `StopTimer` 和 `StartTimer` 避免将这部分时间计算在内。

例如：

```go
// sort_test.go
package main

import (
	"math/rand"
	"testing"
	"time"
)

func generateWithCap(n int) []int {
	rand.Seed(time.Now().UnixNano())
	nums := make([]int, 0, n)
	for i := 0; i < n; i++ {
		nums = append(nums, rand.Int())
	}
	return nums
}

func bubbleSort(nums []int) {
	for i := 0; i < len(nums); i++ {
		for j := 1; j < len(nums)-i; j++ {
			if nums[j] < nums[j-1] {
				nums[j], nums[j-1] = nums[j-1], nums[j]
			}
		}
	}
}

func BenchmarkBubbleSort(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		nums := generateWithCap(10000)
		b.StartTimer()
		bubbleSort(nums)
	}
}
```

执行该用例，每次排序耗时约 0.1s。

```bash
$ go test -bench='Sort$' .
goos: darwin
goarch: amd64
pkg: example
BenchmarkBubbleSort-8                  9         113280509 ns/op
PASS
ok      example 1.146s
```

## 附 推荐与参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
- [How to write benchmarks in Go](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
