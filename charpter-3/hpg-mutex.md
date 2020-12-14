---
title: 读写锁和互斥锁的性能比较
seo_title: Go 语言高性能编程
date: 2020-12-05 23:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。介绍了读写锁(sync.RWMutex)和互斥锁(sync.Mutex)在不同的读写比情况下的性能开销。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 并发编程
keywords:
- golang
- sync.Mutex
- sync.RWMutex
image: post/hpg-mutex/concurrent.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: 读写锁和互斥锁的性能比较
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-3/hpg-mutex.md
---

![high performance go - concurrent programming](hpg-mutex/concurrent.jpg)

## 1 读写锁和互斥锁的区别

Go 语言标准库 `sync` 提供了 2 种锁，互斥锁(sync.Mutex)和读写锁(sync.RWMutex)。那这两种锁的区别是是什么呢？

### 1.1 互斥锁(sync.Mutex)

互斥即不可同时运行。即使用了互斥锁的两个代码片段互相排斥，只有其中一个代码片段执行完成后，另一个才能执行。

Go 标准库中提供了 sync.Mutex 互斥锁类型及其两个方法：

- Lock 加锁
- Unlock 释放锁

我们可以通过在代码前调用 Lock 方法，在代码后调用 Unlock 方法来保证一段代码的互斥执行，也可以用 defer 语句来保证互斥锁一定会被解锁。在一个 Go 协程调用 Lock 方法获得锁后，其他请求锁的协程都会阻塞在 Lock 方法，直到锁被释放。

### 1.2 读写锁(sync.RWMutex)

想象一下这种场景，当你在银行存钱或取钱时，对账户余额的修改是需要加锁的，因为这个时候，可能有人汇款到你的账户，如果对金额的修改不加锁，很可能导致最后的金额发生错误。读取账户余额也需要等待修改操作结束，才能读取到正确的余额。大部分情况下，读取余额的操作会更频繁，如果能保证读取余额的操作能并发执行，程序效率会得到很大地提高。

保证读操作的安全，那只要保证并发读时没有写操作在进行就行。在这种场景下我们需要一种特殊类型的锁，其允许多个只读操作并行执行，但写操作会完全互斥。

这种锁称之为 `多读单写锁` (multiple readers, single writer lock)，简称读写锁，读写锁分为读锁和写锁，读锁是允许同时执行的，但写锁是互斥的。一般来说，有如下几种情况：

- 读锁之间不互斥，没有写锁的情况下，读锁是无阻塞的，多个协程可以同时获得读锁。
- 写锁之间是互斥的，存在写锁，其他写锁阻塞。
- 写锁与读锁是互斥的，如果存在读锁，写锁阻塞，如果存在写锁，读锁阻塞。

Go 标准库中提供了 sync.RWMutex 互斥锁类型及其四个方法：

- Lock 加写锁
- Unlock 释放写锁
- RLock 加读锁
- RUnlock 释放读锁

读写锁的存在是为了解决读多写少时的性能问题，读场景较多时，读写锁可有效地减少锁阻塞的时间。

## 2 读写锁和互斥锁性能比较

接下来，我们测试三种情景下，互斥锁和读写锁的性能差异。

- 读多写少(读占 90%)
- 读少写多(读占 10%)
- 读写一致(各占 50%)

### 2.1 测试用例

接下来我们实现 2 个结构体 `Lock` 和 `RWLock`，并且都继承 `RW` 接口。`RW` 接口中定义了 2 个操作，读(Read)和写(Write)，为了降低其他指令对测试的影响，假定每个读写操作耗时 1 微秒(百万分之一秒)。

- Lock

```go
type RW interface {
	Write()
	Read()
}

const cost = time.Microsecond

type Lock struct {
	count int
	mu    sync.Mutex
}

func (l *Lock) Write() {
	l.mu.Lock()
	l.count++
	time.Sleep(cost)
	l.mu.Unlock()
}

func (l *Lock) Read() {
	l.mu.Lock()
	time.Sleep(cost)
	_ = l.count
	l.mu.Unlock()
}
```

- RWLock

```go
type RWLock struct {
	count int
	mu    sync.RWMutex
}

func (l *RWLock) Write() {
	l.mu.Lock()
	l.count++
	time.Sleep(cost)
	l.mu.Unlock()
}

func (l *RWLock) Read() {
	l.mu.RLock()
	_ = l.count
	time.Sleep(cost)
	l.mu.RUnlock()
}
```

### 2.2 基准测试

```go
func benchmark(b *testing.B, rw RW, read, write int) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for k := 0; k < read*100; k++ {
			wg.Add(1)
			go func() {
				rw.Read()
				wg.Done()
			}()
		}
		for k := 0; k < write*100; k++ {
			wg.Add(1)
			go func() {
				rw.Write()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}


func BenchmarkReadMore(b *testing.B)    { benchmark(b, &Lock{}, 9, 1) }
func BenchmarkReadMoreRW(b *testing.B)  { benchmark(b, &RWLock{}, 9, 1) }
func BenchmarkWriteMore(b *testing.B)   { benchmark(b, &Lock{}, 1, 9) }
func BenchmarkWriteMoreRW(b *testing.B) { benchmark(b, &RWLock{}, 1, 9) }
func BenchmarkEqual(b *testing.B)       { benchmark(b, &Lock{}, 5, 5) }
func BenchmarkEqualRW(b *testing.B)     { benchmark(b, &RWLock{}, 5, 5) }
```

- 三种场景，分别使用 `Lock` 和 `RWLock` 测试，共 6 个用例。
- 每次测试读写操作合计 1000 次，例如读多写少场景，读 900 次，写 100 次。
- 使用 `sync.WaitGroup` 阻塞直到读写操作全部运行结束。

运行结果如下：

```bash
$ go test -bench .
goos: darwin
goarch: amd64
pkg: example/hpg-mutex
BenchmarkReadMore-8                   86          13202572 ns/op
BenchmarkReadMoreRW-8                661           1748724 ns/op
BenchmarkWriteMore-8                  87          13109525 ns/op
BenchmarkWriteMoreRW-8                94          12090900 ns/op
BenchmarkEqual-8                      85          13150321 ns/op
BenchmarkEqualRW-8                   176           6770092 ns/op
PASS
ok      example/hpg-mutex       7.816s
```

- 读写比为 9:1 时，读写锁的性能约为互斥锁的 8 倍
- 读写比为 1:9 时，读写锁性能相当
- 读写比为 5:5 时，读写锁的性能约为互斥锁的 2 倍

### 2.3 改变读写操作的时间

如果将单位读写操作的时间降为 0.1 微秒，结果如何呢？

```go
const cost = time.Nanosecond * 100
```

测试结果如下：

```go
$ go test -bench .
goos: darwin
goarch: amd64
pkg: example/hpg-mutex
BenchmarkReadMore-8                  715           1835021 ns/op
BenchmarkReadMoreRW-8               2198            462859 ns/op
BenchmarkWriteMore-8                 685           1831686 ns/op
BenchmarkWriteMoreRW-8               709           1679783 ns/op
BenchmarkEqual-8                     625           1844344 ns/op
BenchmarkEqualRW-8                  1057           1068423 ns/op
PASS
ok      example/hpg-mutex       7.957s
```

单位读写操作时间下降后，读写锁的性能优势下降到 3 倍，这也是可以理解的，因加锁而阻塞的时间占比减小，互斥锁带来的损耗自然就减小了。

将单位读写操作时间增加到 10 微秒的结果呢？

```go
const cost = time.Microsecond * 10
```

测试结果如下：

```bash
goos: darwin
goarch: amd64
pkg: example/hpg-mutex
BenchmarkReadMore-8                   49          24507629 ns/op
BenchmarkReadMoreRW-8                414           2873828 ns/op
BenchmarkWriteMore-8                  49          24452297 ns/op
BenchmarkWriteMoreRW-8                51          22208048 ns/op
BenchmarkEqual-8                      45          24486665 ns/op
BenchmarkEqualRW-8                    93          12414773 ns/op
PASS
ok      example/hpg-mutex       7.394s
```

单位时间增加后，读写锁和互斥锁的性能比与 1 微秒时基本一致。

## 附 互斥锁如何实现公平

如果多个 goroutine 都在请求同一个锁，sync.Mutex 是如何实现分配公平的呢？[sync.mutex 源代码分析](https://colobu.com/2018/12/18/dive-into-sync-mutex/) 这篇文章介绍了 sync.Mutex 的演进历史和当前的实现机制。重要的部分引用如下：

根据Mutex的注释，当前的 Mutex 有如下的性质。这些注释将极大的帮助我们理解Mutex的实现。

> 互斥锁有两种状态：正常状态和饥饿状态。
>
> 在正常状态下，所有等待锁的 goroutine 按照FIFO顺序等待。唤醒的 goroutine 不会直接拥有锁，而是会和新请求锁的 goroutine 竞争锁的拥有。新请求锁的 goroutine 具有优势：它正在 CPU 上执行，而且可能有好几个，所以刚刚唤醒的 goroutine 有很大可能在锁竞争中失败。在这种情况下，这个被唤醒的 goroutine 会加入到等待队列的前面。 如果一个等待的 goroutine 超过 1ms 没有获取锁，那么它将会把锁转变为饥饿模式。
> 
> 在饥饿模式下，锁的所有权将从 unlock 的 goroutine 直接交给交给等待队列中的第一个。新来的 goroutine 将不会尝试去获得锁，即使锁看起来是 unlock 状态, 也不会去尝试自旋操作，而是放在等待队列的尾部。
> 
> 如果一个等待的 goroutine 获取了锁，并且满足一以下其中的任何一个条件：(1)它是队列中的最后一个；(2)它等待的时候小于1ms。它会将锁的状态转换为正常状态。
> 
> 正常状态有很好的性能表现，饥饿模式也是非常重要的，因为它能阻止尾部延迟的现象。

## 附 推荐与参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
