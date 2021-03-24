---
title: Go sync.Cond
seo_title: Go 语言高性能编程
date: 2021-01-14 23:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。sync.Cond 是一个条件锁，也被称为条件变量，常用来一写多读(一个 goroutine 通知多个在等待的 goroutines)的场景。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 并发编程
keywords:
- golang
- sync.Cond
- 条件变量
image: post/hpg-mutex/concurrent.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: sync.Cond 条件变量
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-3/hpg-sync-cond.md
---

![high performance go - concurrent programming](hpg-mutex/concurrent.jpg)

## 1 sync.Cond 的使用场景

> 一句话总结：`sync.Cond` 条件变量用来协调想要访问共享资源的那些 goroutine，当共享资源的状态发生变化的时候，它可以用来通知被互斥锁阻塞的 goroutine。

`sync.Cond` 基于互斥锁/读写锁，它和互斥锁的区别是什么呢？

互斥锁 `sync.Mutex` 通常用来保护临界区和共享资源，条件变量 `sync.Cond` 用来协调想要访问共享资源的 goroutine。

`sync.Cond` 经常用在多个 goroutine 等待，一个 goroutine 通知（事件发生）的场景。如果是一个通知，一个等待，使用互斥锁或 channel 就能搞定了。

我们想象一个非常简单的场景：

有一个协程在异步地接收数据，剩下的多个协程必须等待这个协程接收完数据，才能读取到正确的数据。在这种情况下，如果单纯使用 chan 或互斥锁，那么只能有一个协程可以等待，并读取到数据，没办法通知其他的协程也读取数据。

这个时候，就需要有个全局的变量来标志第一个协程数据是否接受完毕，剩下的协程，反复检查该变量的值，直到满足要求。或者创建多个 channel，每个协程阻塞在一个 channel 上，由接收数据的协程在数据接收完毕后，逐个通知。总之，需要额外的复杂度来完成这件事。

Go 语言在标准库 sync 中内置一个 `sync.Cond` 用来解决这类问题。

## 2 sync.Cond 的四个方法

sync.Cond 的定义如下：

```go
// Each Cond has an associated Locker L (often a *Mutex or *RWMutex),
// which must be held when changing the condition and
// when calling the Wait method.
//
// A Cond must not be copied after first use.
type Cond struct {
        noCopy noCopy

        // L is held while observing or changing the condition
        L Locker

        notify  notifyList
        checker copyChecker
}
```

每个 Cond 实例都会关联一个锁 L（互斥锁 *Mutex，或读写锁 *RWMutex），当修改条件或者调用 Wait 方法时，必须加锁。

和 sync.Cond 相关的有如下几个方法：

### 2.1 NewCond 创建实例

```go
func NewCond(l Locker) *Cond
```

NewCond 创建 Cond 实例时，需要关联一个锁。

### 2.2 Broadcast 广播唤醒所有

```go
// Broadcast wakes all goroutines waiting on c.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
func (c *Cond) Broadcast()
```

Broadcast 唤醒所有等待条件变量 c 的 goroutine，无需锁保护。

### 2.3 Signal 唤醒一个协程

```go
// Signal wakes one goroutine waiting on c, if there is any.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
func (c *Cond) Signal()
```

Signal 只唤醒任意 1 个等待条件变量 c 的 goroutine，无需锁保护。


### 2.4 Wait 等待

```go
// Wait atomically unlocks c.L and suspends execution
// of the calling goroutine. After later resuming execution,
// Wait locks c.L before returning. Unlike in other systems,
// Wait cannot return unless awoken by Broadcast or Signal.
//
// Because c.L is not locked when Wait first resumes, the caller
// typically cannot assume that the condition is true when
// Wait returns. Instead, the caller should Wait in a loop:
//
//    c.L.Lock()
//    for !condition() {
//        c.Wait()
//    }
//    ... make use of condition ...
//    c.L.Unlock()
//
func (c *Cond) Wait()
```

调用 Wait 会自动释放锁 c.L，并挂起调用者所在的 goroutine，因此当前协程会阻塞在 Wait 方法调用的地方。如果其他协程调用了 Signal 或 Broadcast 唤醒了该协程，那么 Wait 方法在结束阻塞时，会重新给 c.L 加锁，并且继续执行 Wait 后面的代码。

对条件的检查，使用了 `for !condition()` 而非 `if`，是因为当前协程被唤醒时，条件不一定符合要求，需要再次 Wait 等待下次被唤醒。为了保险起见，使用 `for` 能够确保条件符合要求后，再执行后续的代码。

```go
   c.L.Lock()
   for !condition() {
       c.Wait()
   }
   ... make use of condition ...
   c.L.Unlock()
```

## 3 使用示例

接下来我们实现一个简单的例子，三个协程调用 `Wait()` 等待，另一个协程调用 `Broadcast()` 唤醒所有等待的协程。

```go
var done = false

func read(name string, c *sync.Cond) {
	c.L.Lock()
	for !done {
		c.Wait()
	}
	log.Println(name, "starts reading")
	c.L.Unlock()
}

func write(name string, c *sync.Cond) {
	log.Println(name, "starts writing")
	time.Sleep(time.Second)
	c.L.Lock()
	done = true
	c.L.Unlock()
	log.Println(name, "wakes all")
	c.Broadcast()
}

func main() {
	cond := sync.NewCond(&sync.Mutex{})

	go read("reader1", cond)
	go read("reader2", cond)
	go read("reader3", cond)
	write("writer", cond)

	time.Sleep(time.Second * 3)
}
```

- `done` 即互斥锁需要保护的条件变量。
- `read()` 调用 `Wait()` 等待通知，直到 done 为 true。
- `write()` 接收数据，接收完成后，将 done 置为 true，调用 `Broadcast()` 通知所有等待的协程。
- `write()` 中的暂停了 1s，一方面是模拟耗时，另一方面是确保前面的 3 个 read 协程都执行到 `Wait()`，处于等待状态。main 函数最后暂停了 3s，确保所有操作执行完毕。

运行结果如下：

```bash
$ go run main.go
2021/01/14 23:18:20 writer starts writing
2021/01/14 23:18:21 writer wakes all
2021/01/14 23:18:21 reader2 starts reading
2021/01/14 23:18:21 reader3 starts reading
2021/01/14 23:18:21 reader1 starts reading
```

writer 接收数据花费了 1s，同步通知所有等待的协程。

> 更多关于 sync.Cond 的讨论可参考 [How to correctly use sync.Cond? - StackOverflow](https://stackoverflow.com/questions/36857167/how-to-correctly-use-sync-cond)

## 附 推荐和参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
