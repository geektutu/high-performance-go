---
title: 如何退出协程 goroutine (超时场景)
seo_title: Go 语言高性能编程
date: 2020-12-12 20:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。本文介绍了 Go 语言中实现超时(timeout)退出时常见的错误，即子协程(goroutine)不能正常关闭，导致内存泄漏。常见的实现超时的方式有 time.After 和 context.WithTimeout。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 并发编程
keywords:
- golang
- time.After
- WithTimeout
image: post/hpg-mutex/concurrent.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: 如何退出协程(超时场景)
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-3/hpg-timeout-goroutine.md
---

![high performance go - concurrent programming](hpg-mutex/concurrent.jpg)

## 1 超时返回时的陷阱

超时控制在网络编程中是非常常见的，利用 `context.WithTimeout` 和 `time.After` 都能够很轻易地实现。

### 1.1 time.After 实现超时控制

```go
func doBadthing(done chan bool) {
	time.Sleep(time.Second)
	done <- true
}

func timeout(f func(chan bool)) error {
	done := make(chan bool)
	go f(done)
	select {
	case <-done:
		fmt.Println("done")
		return nil
	case <-time.After(time.Millisecond):
		return fmt.Errorf("timeout")
	}
}

// timeout(doBadthing)
```

上述代码是一个典型的实现超时的例子。

- 利用 `time.After` 启动了一个异步的定时器，返回一个 channel，当超过指定的时间后，该 channel 将会接受到信号。
- 启动了子协程执行函数 f，函数执行结束后，将向 channel `done` 发送结束信号。
- 使用 select 阻塞等待 `done` 或 `time.After` 的信息，若超时，则返回错误，若没有超时，则返回 nil。


如果每次调用，函数 f 都能够在超时前正常结束，那么启动的子协程(goroutine)能够正常退出。那如果是超时场景呢？子协程能够正常退出么？

### 1.2 测试协程是否退出

在这个例子中超时时间为 1 ms，而 `doBadthing` 需要 1s 才能结束运行。因此 `timeout(doBadthing)` 一定会触发超时。我们利用单元测试，来看一看超时场景下协程的情况。

```go
func test(t *testing.T, f func(chan bool)) {
	t.Helper()
	for i := 0; i < 1000; i++ {
		timeout(f)
	}
	time.Sleep(time.Second * 2)
	t.Log(runtime.NumGoroutine())
}

func TestBadTimeout(t *testing.T)  { test(t, doBadthing) }
```

- `timeout(doBadthing)` 调用了 1000 次，理论上会启动 1000 个子协程。
- 利用 `runtime.NumGoroutine()` 打印当前程序的协程个数。
- 因为 `doBadthing` 执行时间为 1s，因此打印协程个数前，等待 2s，确保函数执行完毕。

测试结果如下：

```bash
$ go test -run ^TestBadTimeout$ . -v
=== RUN   TestBadTimeout
--- PASS: TestBadTimeout (3.43s)
    timeout_test.go:49: 1002
```

最终程序中存在着 1002 个子协程，说明即使是函数执行完成，协程也没有正常退出。那如果在实际的业务中，我们使用了上述的代码，那越来越多的协程会残留在程序中，最终会导致内存耗尽（每个协程约占 2K 空间），程序崩溃。

我们仔细阅读这段代码，其实是非常容易发现问题所在的。`done` 是一个无缓冲区的 channel，如果没有超时，`doBadthing` 中会向 done 发送信号，`select` 中会接收 done 的信号，因此 `doBadthing` 能够正常退出，子协程也能够正常退出。

但是，当超时发生时，select 接收到 `time.After` 的超时信号就返回了，`done` 没有了接收方(receiver)，而 `doBadthing` 在执行 1s 后向 `done` 发送信号，由于没有接收者且无缓存区，发送者(sender)会一直阻塞，导致协程不能退出。

## 2 如何避免

### 2.1 创建有缓冲区的 channel

即创建channel `done` 时，缓冲区设置为 1，即使没有接收方，发送方也不会发生阻塞。

```go
func timeoutWithBuffer(f func(chan bool)) error {
	done := make(chan bool, 1)
	go f(done)
	select {
	case <-done:
		fmt.Println("done")
		return nil
	case <-time.After(time.Millisecond):
		return fmt.Errorf("timeout")
	}
}

func TestBufferTimeout(t *testing.T) {
	for i := 0; i < 1000; i++ {
		timeoutWithBuffer(doBadthing)
	}
	time.Sleep(time.Second * 2)
	t.Log(runtime.NumGoroutine())
}
```

测试结果如下：

```bash
$ go test -run ^TestBufferTimeout$ . -v
=== RUN   TestBufferTimeout
--- PASS: TestBufferTimeout (3.36s)
    timeout_test.go:65: 2
```

协程数量下降为 2，创建的 1000 个子协程成功退出。

### 2.2 使用 select 尝试发送

设置缓冲区是一种方式，还有另一种方式：

```go
func doGoodthing(done chan bool) {
	time.Sleep(time.Second)
	select {
	case done <- true:
	default:
		return
	}
}

func TestGoodTimeout(t *testing.T) { test(t, doGoodthing) }
```

测试结果如下：

```bash
$ go test -run ^TestGoodTimeout$ . -v
=== RUN   TestGoodTimeout
--- PASS: TestGoodTimeout (3.40s)
    timeout_test.go:58: 2
```

使用 select 尝试向信道 done 发送信号，如果发送失败，则说明缺少接收者(receiver)，即超时了，那么直接退出即可。

### 2.3 更复杂的场景

还有一些更复杂的场景，例如将任务拆分为多段，只检测第一段是否超时，若没有超时，后续任务继续执行，超时则终止。

```go
func do2phases(phase1, done chan bool) {
	time.Sleep(time.Second) // 第 1 段
	select {
	case phase1 <- true:
	default:
		return
	}
	time.Sleep(time.Second) // 第 2 段
	done <- true
}

func timeoutFirstPhase() error {
	phase1 := make(chan bool)
	done := make(chan bool)
	go do2phases(phase1, done)
	select {
	case <-phase1:
		<-done
		fmt.Println("done")
		return nil
	case <-time.After(time.Millisecond):
		return fmt.Errorf("timeout")
	}
}

func Test2phasesTimeout(t *testing.T) {
	for i := 0; i < 1000; i++ {
		timeoutFirstPhase()
	}
	time.Sleep(time.Second * 3)
	t.Log(runtime.NumGoroutine())
}
```

测试结果如下：

```bash
$ go test -run ^Test2phasesTimeout$ . -v
=== RUN   Test2phasesTimeout
--- PASS: Test2phasesTimeout (4.43s)
    timeout_test.go:98: 2
```

这种场景在实际的业务中更为常见，例如我们将服务端接收请求后的任务拆分为 2 段，一段是执行任务，一段是发送结果。那么就会有两种情况：

- 任务正常执行，向客户端返回执行结果。
- 任务超时执行，向客户端返回超时。

这种情况下，就只能够使用 select，而不能能够设置缓冲区的方式了。因为如果给信道 phase1 设置了缓冲区，`phase1 <- true` 总能执行成功，那么无论是否超时，都会执行到第二阶段，而没有即时返回，这是我们不愿意看到的。对应到上面的业务，就可能发生一种异常情况，向客户端发送了 2 次响应：

- 任务超时执行，向客户端返回超时，一段时间后，向客户端返回执行结果。

缓冲区不能够区分是否超时了，但是 select 可以（没有接收方，信道发送信号失败，则说明超时了）。

## 3 强制 kill goroutine 可能吗？

### 3.1 答案是不能

上面的例子，即时超时返回了，但是子协程仍在继续运行，直到自己退出。那么有可能在超时的时候，就强制关闭子协程吗？

答案是不能，goroutine 只能自己退出，而不能被其他 goroutine 强制关闭或杀死。

> goroutine 被设计为不可以从外部无条件地结束掉，只能通过 channel 来与它通信。也就是说，每一个 goroutine 都需要承担自己退出的责任。(A goroutine cannot be programmatically killed. It can only commit a cooperative suicide.)

关于这个问题，Github 上也有讨论：

> [question: is it possible to a goroutine immediately stop another goroutine?](https://github.com/golang/go/issues/32610)

摘抄其中几个比较有意思的观点如下：

- 杀死一个 goroutine 设计上会有很多挑战，当前所拥有的资源如何处理？堆栈如何处理？defer 语句需要执行么？
- 如果允许 defer 语句执行，那么 defer 语句可能阻塞 goroutine 退出，这种情况下怎么办呢？


### 3.2 一些建议

因为 goroutine 不能被强制 kill，在超时或其他类似的场景下，为了 goroutine 尽可能正常退出，建议如下：

- 尽量使用非阻塞 I/O（非阻塞 I/O 常用来实现高性能的网络库），阻塞 I/O 很可能导致 goroutine 在某个调用一直等待，而无法正确结束。
- 业务逻辑总是考虑退出机制，避免死循环。
- 任务分段执行，超时后即时退出，避免 goroutine 无用的执行过多，浪费资源。

## 附 推荐与参考

- [is it possible to a goroutine immediately stop another goroutine? - Github](https://github.com/golang/go/issues/32610)
- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
