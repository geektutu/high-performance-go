---
title: Go sync.Once
seo_title: Go 语言高性能编程
date: 2021-01-07 23:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。sync.Once 是 Golang package 中使方法只执行一次的对象实现，作用与 init 函数类似，但也有所不同。本文还解释了 sync.Once 源码中,done 为什么作为第一个字段。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 并发编程
keywords:
- golang
- sync.Once
image: post/hpg-mutex/concurrent.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: sync.Once 如何提升性能
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-3/hpg-sync-once.md
---

![high performance go - concurrent programming](hpg-mutex/concurrent.jpg)

## 1 sync.Once 的使用场景

`sync.Once` 是 Go 标准库提供的使函数只执行一次的实现，常应用于单例模式，例如初始化配置、保持数据库连接等。作用与 `init` 函数类似，但有区别。

- init 函数是当所在的 package 首次被加载时执行，若迟迟未被使用，则既浪费了内存，又延长了程序加载时间。
- sync.Once 可以在代码的任意位置初始化和调用，因此可以延迟到使用时再执行，并发场景下是线程安全的。

在多数情况下，`sync.Once` 被用于控制变量的初始化，这个变量的读写满足如下三个条件：

- 当且仅当第一次访问某个变量时，进行初始化（写）；
- 变量初始化过程中，所有读都被阻塞，直到初始化完成；
- 变量仅初始化一次，初始化完成后驻留在内存里。

`sync.Once` 仅提供了一个方法 `Do`，参数 f 是对象初始化函数。

```go
func (o *Once) Do(f func())
```

## 2 使用示例

### 2.1 一个简单的 Demo

考虑一个简单的场景，函数 ReadConfig 需要读取环境变量，并转换为对应的配置。环境变量在程序执行前已经确定，执行过程中不会发生改变。ReadConfig 可能会被多个协程并发调用，为了提升性能（减少执行时间和内存占用），使用 `sync.Once` 是一个比较好的方式。

```go
type Config struct {
	Server string
	Port   int64
}

var (
	once   sync.Once
	config *Config
)

func ReadConfig() *Config {
	once.Do(func() {
		var err error
		config = &Config{Server: os.Getenv("TT_SERVER_URL")}
		config.Port, err = strconv.ParseInt(os.Getenv("TT_PORT"), 10, 0)
		if err != nil {
			config.Port = 8080 // default port
        }
        log.Println("init config")
	})
	return config
}

func main() {
	for i := 0; i < 10; i++ {
		go func() {
			_ = ReadConfig()
		}()
	}
	time.Sleep(time.Second)
}
```

- 在这个例子中，声明了 2 个全局变量，once 和 config；
- config 是需要在 ReadConfig 函数中初始化的(将环境变量转换为 Config 结构体)，ReadConfig 可能会被并发调用。

如果 ReadConfig 每次都构造出一个新的 Config 结构体，既浪费内存，又浪费初始化时间。如果 ReadConfig 中不加锁，初始化全局变量 config 就可能出现并发冲突。这种情况下，使用 sync.Once 既能够保证全局变量初始化时是线程安全的，又能节省内存和初始化时间。

运行结果如下：

```bash
$ go run .
2021/01/07 23:51:49 init config
```

`init config` 仅打印了一次，即 sync.Once 中的初始化函数仅执行了一次。


### 2.2 标准库中 sync.Once 的使用

`sync.Once` 在 Go 语言标准库中被广泛使用，我们可以简单地搜索一下：

```bash
$ grep -nr "sync\.Once" "$(dirname $(which go))/../src"
/usr/local/go/bin/../src/cmd/go/internal/cache/default.go:25:   defaultOnce  sync.Once
/usr/local/go/bin/../src/cmd/go/internal/cache/default.go:63:   defaultDirOnce sync.Once
/usr/local/go/bin/../src/cmd/go/internal/auth/netrc.go:23:      netrcOnce sync.Once
/usr/local/go/bin/../src/cmd/go/internal/modfetch/sumdb.go:53:  dbOnce sync.Once
/usr/local/go/bin/../src/cmd/go/internal/modfetch/sumdb.go:117: once    sync.Once
/usr/local/go/bin/../src/cmd/go/internal/modfetch/codehost/git.go:126:  refsOnce sync.Once
/usr/local/go/bin/../src/cmd/go/internal/modfetch/codehost/git.go:132:  localTagsOnce sync.Once
...
$ grep -nr "sync\.Once" "$(dirname $(which go))/../src" | wc -l
111
```

在 go1.13.6 版本的源码目录下，可以 grep 到 111 处使用。

比如 package `html` 中，对象 entity 只被初始化一次：

```go
var populateMapsOnce sync.Once
var entity           map[string]rune

func populateMaps() {
    entity = map[string]rune{
        "AElig;":                           '\U000000C6',
        "AMP;":                             '\U00000026',
        "Aacute;":                          '\U000000C1',
        "Abreve;":                          '\U00000102',
        "Acirc;":                           '\U000000C2',
        // 省略 2000 项
    }
}

func UnescapeString(s string) string {
    populateMapsOnce.Do(populateMaps)
    i := strings.IndexByte(s, '&')

    if i < 0 {
            return s
    }
    // 省略后续的实现
}
```

- 字典 `entity` 包含 2005 个键值对，若使用 init 在包加载时初始化，若不被使用，将会浪费大量内存。
- `html.UnescapeString(s)` 函数是线程安全的，可能会被用户程序在并发场景下调用，因此对 entity 的初始化需要加锁，使用 `sync.Once` 能保证这一点。

## 3 sync.Once 的原理

首先：保证变量仅被初始化一次，需要有个标志来判断变量是否已初始化过，若没有则需要初始化。

第二：线程安全，支持并发，无疑需要互斥锁来实现。

### 3.1 源码实现

以下是 `sync.Once` 的源码实现，代码位于 `$(dirname $(which go))/../src/sync/once.go`：

```go
package sync

import (
    "sync/atomic"
)

type Once struct {
    done uint32
    m    Mutex
}

func (o *Once) Do(f func()) {
    if atomic.LoadUint32(&o.done) == 0 {
        o.doSlow(f)
    }
}

func (o *Once) doSlow(f func()) {
    o.m.Lock()
    defer o.m.Unlock()
    if o.done == 0 {
        defer atomic.StoreUint32(&o.done, 1)
        f()
    }
}
```

`sync.Once` 的实现与一开始的猜测是一样的，使用 `done` 标记是否已经初始化，使用锁 `m Mutex` 实现线程安全。

### 3.2 done 为什么是第一个字段

字段 `done` 的注释也非常值得一看：

```go
type Once struct {
    // done indicates whether the action has been performed.
    // It is first in the struct because it is used in the hot path.
    // The hot path is inlined at every call site.
    // Placing done first allows more compact instructions on some architectures (amd64/x86),
    // and fewer instructions (to calculate offset) on other architectures.
    done uint32
    m    Mutex
}
```

其中解释了为什么将 done 置为 Once 的第一个字段：done 在热路径中，done 放在第一个字段，能够减少 CPU 指令，也就是说，这样做能够提升性能。

简单解释下这句话：

1. 热路径(hot path)是程序非常频繁执行的一系列指令，sync.Once 绝大部分场景都会访问 `o.done`，在热路径上是比较好理解的，如果 hot path 编译后的机器码指令更少，更直接，必然是能够提升性能的。

2. 为什么放在第一个字段就能够减少指令呢？因为结构体第一个字段的地址和结构体的指针是相同的，如果是第一个字段，直接对结构体的指针解引用即可。如果是其他的字段，除了结构体指针外，还需要计算与第一个值的偏移(calculate offset)。在机器码中，偏移量是随指令传递的附加值，CPU 需要做一次偏移值与指针的加法运算，才能获取要访问的值的地址。因为，访问第一个字段的机器代码更紧凑，速度更快。

> 参考 [What does “hot path” mean in the context of sync.Once? - StackOverflow](https://stackoverflow.com/questions/59174176/what-does-hot-path-mean-in-the-context-of-sync-once)


## 附 推荐与参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
