---
title: 控制协程(goroutine)的并发数量
seo_title: Go 语言高性能编程
date: 2020-12-21 01:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。本文介绍了 goroutine 协程并发控制，避免并发过高，大量消耗系统资源，导致程序崩溃或卡顿，影响性能。主要通过 2 种方式控制，一是使用 channel 的缓冲区，二是使用第三方协程池，例如 tunny 和 ants。同时介绍了使用 ulimit 和虚拟内存(virtual memory)提高资源上限的技巧。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 并发编程
keywords:
- golang
- channel
- 并发控制
image: post/hpg-mutex/concurrent.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: 控制协程的并发数量
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-3/hpg-concurrency-control.md
---

![high performance go - concurrent programming](hpg-mutex/concurrent.jpg)

## 1 并发过高导致程序崩溃

我们首先看一个非常简单的例子：

```go
func main() {
	var wg sync.WaitGroup
	for i := 0; i < math.MaxInt32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Println(i)
			time.Sleep(time.Second)
		}(i)
	}
	wg.Wait()
}
```

这个例子实现了 `math.MaxInt32` 个协程的并发，约 2^31 = 2 亿个，每个协程内部几乎没有做什么事情。正常的情况下呢，这个程序会乱序输出 `1 -> 2^31` 个数字。

那实际运行的结果是怎么样的呢？

```bash
$ go run main.go
...
150577
150578
panic: too many concurrent operations on a single file or socket (max 1048575)

goroutine 1199236 [running]:
internal/poll.(*fdMutex).rwlock(0xc0000620c0, 0x0, 0xc0000781b0)
        /usr/local/go/src/internal/poll/fd_mutex.go:147 +0x13f
internal/poll.(*FD).writeLock(...)
        /usr/local/go/src/internal/poll/fd_mutex.go:239
internal/poll.(*FD).Write(0xc0000620c0, 0xc125ccd6e0, 0x11, 0x20, 0x0, 0x0, 0x0)
        /usr/local/go/src/internal/poll/fd_unix.go:255 +0x5e
fmt.Fprintf(0x10ed3e0, 0xc00000e018, 0x10d3024, 0xc, 0xc0e69b87b0, 0x1, 0x1, 0x11, 0x0, 0x0)
        /usr/local/go/src/fmt/print.go:205 +0xa5
fmt.Printf(...)
        /usr/local/go/src/fmt/print.go:213
main.main.func1(0xc0000180b0, 0x124c31)
...
```

运行的结果是程序直接崩溃了，关键的报错信息是：

```bash
panic: too many concurrent operations on a single file or socket (max 1048575)
```

对单个 file/socket 的并发操作个数超过了系统上限，这个报错是 `fmt.Printf` 函数引起的，`fmt.Printf` 将格式化后的字符串打印到屏幕，即标准输出。在 linux 系统中，标准输出也可以视为文件，内核(kernel)利用文件描述符(file descriptor)来访问文件，标准输出的文件描述符为 1，错误输出文件描述符为 2，标准输入的文件描述符为 0。

简而言之，系统的资源被耗尽了。

那如果我们将 `fmt.Printf` 这行代码去掉呢？那程序很可能会因为内存不足而崩溃。这一点更好理解，每个协程至少需要消耗 2KB 的空间，那么假设计算机的内存是 2GB，那么至多允许 2GB/2KB = 1M 个协程同时存在。那如果协程中还存在着其他需要分配内存的操作，那么允许并发执行的协程将会数量级地减少。

## 2 如何解决

不同的应用程序，消耗的资源是不一样的。比较推荐的方式的是：应用程序来主动限制并发的协程数量。

### 2.1 利用 channel 的缓存区

可以利用信道 channel 的缓冲区大小来实现：

```go
// main_chan.go
func main() {
	var wg sync.WaitGroup
	ch := make(chan struct{}, 3)
	for i := 0; i < 10; i++ {
		ch <- struct{}{}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			log.Println(i)
			time.Sleep(time.Second)
			<-ch
		}(i)
	}
	wg.Wait()
}
```

- `make(chan struct{}, 3)` 创建缓冲区大小为 3 的 channel，在没有被接收的情况下，至多发送 3 个消息则被阻塞。
- 开启协程前，调用 `ch <- struct{}{}`，若缓存区满，则阻塞。
- 协程任务结束，调用 `<-ch` 释放缓冲区。
- `sync.WaitGroup` 并不是必须的，例如 http 服务，每个请求天然是并发的，此时使用 channel 控制并发处理的任务数量，就不需要 `sync.WaitGroup`。

运行结果如下：

```bash
$ go run main_chan.go
2020/12/21 00:48:28 2
2020/12/21 00:48:28 0
2020/12/21 00:48:28 1
2020/12/21 00:48:29 3
2020/12/21 00:48:29 4
2020/12/21 00:48:29 5
2020/12/21 00:48:30 6
2020/12/21 00:48:30 7
2020/12/21 00:48:30 8
2020/12/21 00:48:31 9
```

从日志中可以很容易看到，每秒钟只并发执行了 3 个任务，达到了协程并发控制的目的。

### 2.2 利用第三方库

目前有很多第三方库实现了协程池，可以很方便地用来控制协程的并发数量，比较受欢迎的有：

- [Jeffail/tunny](https://github.com/Jeffail/tunny)
- [panjf2000/ants](https://github.com/panjf2000/ants)

以 `tunny` 举例：

```go
package main

import (
	"log"
	"time"

	"github.com/Jeffail/tunny"
)

func main() {
	pool := tunny.NewFunc(3, func(i interface{}) interface{} {
		log.Println(i)
		time.Sleep(time.Second)
		return nil
	})
	defer pool.Close()

	for i := 0; i < 10; i++ {
		go pool.Process(i)
	}
	time.Sleep(time.Second * 4)
}
```

- `tunny.NewFunc(3, f)` 第一个参数是协程池的大小(poolSize)，第二个参数是协程运行的函数(worker)。
- `pool.Process(i)` 将参数 i 传递给协程池定义好的 worker 处理。
- `pool.Close()` 关闭协程池。

运行结果如下：

```bash
$ go run main_tunny.go
2020/12/21 01:00:21 6
2020/12/21 01:00:21 1
2020/12/21 01:00:21 3
2020/12/21 01:00:22 8
2020/12/21 01:00:22 4
2020/12/21 01:00:22 7
2020/12/21 01:00:23 5
2020/12/21 01:00:23 2
2020/12/21 01:00:23 0
2020/12/21 01:00:24 9
```

## 3 调整系统资源的上限

### 3.1 ulimit

有些场景下，即使我们有效地限制了协程的并发数量，但是仍旧出现了某一类资源不足的问题，例如：

- too many open files
- out of memory
- ...

例如分布式编译加速工具，需要解析 gcc 命令以及依赖的源文件和头文件，有些编译命令依赖的头文件可能有上百个，那这个时候即使我们将协程的并发数限制到 1000，也可能会超过进程运行时并发打开的文件句柄数量，但是分布式编译工具，仅将依赖的源文件和头文件分发到远端机器执行，并不会消耗本机的内存和 CPU 资源，因此 1000 个并发并不高，这种情况下，降低并发数会影响编译加速的效率，那能不能增加进程能同时打开的文件句柄数量呢？

操作系统通常会限制同时打开文件数量、栈空间大小等，`ulimit -a` 可以看到系统当前的设置：

```bash
$ ulimit -a
-t: cpu time (seconds)              unlimited
-f: file size (blocks)              unlimited
-d: data seg size (kbytes)          unlimited
-s: stack size (kbytes)             8192
-c: core file size (blocks)         0
-v: address space (kbytes)          unlimited
-l: locked-in-memory size (kbytes)  unlimited
-u: processes                       1418
-n: file descriptors                12800
```

我们可以使用 `ulimit -n 999999`，将同时打开的文件句柄数量调整为 999999 来解决这个问题，其他的参数也可以按需调整。

### 3.2 虚拟内存(virtual memory)

虚拟内存是一项非常常见的技术了，即在内存不足时，将磁盘映射为内存使用，比如 linux 下的交换分区(swap space)。

在 linux 上创建并使用交换分区是一件非常简单的事情：

```bash
sudo fallocate -l 20G /mnt/.swapfile # 创建 20G 空文件
sudo mkswap /mnt/.swapfile    # 转换为交换分区文件
sudo chmod 600 /mnt/.swapfile # 修改权限为 600
sudo swapon /mnt/.swapfile    # 激活交换分区
free -m # 查看当前内存使用情况(包括交换分区)
```

关闭交换分区也非常简单：

```bash
sudo swapoff /mnt/.swapfile
rm -rf /mnt/.swapfile
```

磁盘的 I/O 读写性能和内存条相差是非常大的，例如 DDR3 的内存条读写速率很容易达到 20GB/s，但是 SSD 固态硬盘的读写性能通常只能达到 0.5GB/s，相差 40倍之多。因此，使用虚拟内存技术将硬盘映射为内存使用，显然会对性能产生一定的影响。如果应用程序只是在较短的时间内需要较大的内存，那么虚拟内存能够有效避免 `out of memory` 的问题。如果应用程序长期高频度读写大量内存，那么虚拟内存对性能的影响就比较明显了。

## 附 推荐与参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
