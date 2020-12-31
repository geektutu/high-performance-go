---
title: Go sync.Pool
seo_title: Go 语言高性能编程
date: 2020-12-31 23:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。Go 语言标准库中的 sync.Pool 可以建立对象池，复用已有对象，解决内存分配碎片化的问题，有效减轻垃圾回收的压力(Garbage Collection, GC)，在特定场景下可以有效地降低内存占用，提升性能。最后介绍了 sync.Pool 的作用和用法。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 并发编程
keywords:
- golang
- sync.Pool
- 内存碎片
image: post/hpg-mutex/concurrent.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: sync.Pool 复用对象
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-3/hpg-sync-pool.md
---

![high performance go - concurrent programming](hpg-mutex/concurrent.jpg)

## 1 sync.Pool 的使用场景

> 一句话总结：保存和复用临时对象，减少内存分配，降低 GC 压力。

举个简单的例子：

```go
type Student struct {
	Name   string
	Age    int32
	Remark [1024]byte
}

var buf, _ = json.Marshal(Student{Name: "Geektutu", Age: 25})

func unmarsh() {
	stu := &Student{}
	json.Unmarshal(buf, stu)
}
```

json 的反序列化在文本解析和网络通信过程中非常常见，当程序并发度非常高的情况下，短时间内需要创建大量的临时对象。而这些对象是都是分配在堆上的，会给 GC 造成很大压力，严重影响程序的性能。

> 参考：[垃圾回收(GC)的工作原理](https://geektutu.com/post/qa-golang-2.html#Q5-简述-Go-语言GC-垃圾回收-的工作原理)

Go 语言从 1.3 版本开始提供了对象重用的机制，即 sync.Pool。sync.Pool 是可伸缩的，同时也是并发安全的，其大小仅受限于内存的大小。sync.Pool 用于存储那些被分配了但是没有被使用，而未来可能会使用的值。这样就可以不用再次经过内存分配，可直接复用已有对象，减轻 GC 的压力，从而提升系统的性能。

sync.Pool 的大小是可伸缩的，高负载时会动态扩容，存放在池中的对象如果不活跃了会被自动清理。

## 2 如何使用

sync.Pool 的使用方式非常简单：

### 2.1 声明对象池

只需要实现 New 函数即可。对象池中没有对象时，将会调用 New 函数创建。

```go
var studentPool = sync.Pool{
    New: func() interface{} { 
        return new(Student) 
    },
}
```

### 2.2 Get & Put

```go
stu := studentPool.Get().(*Student)
json.Unmarshal(buf, stu)
studentPool.Put(stu)
```

- `Get()` 用于从对象池中获取对象，因为返回值是 `interface{}`，因此需要类型转换。
- `Put()` 则是在对象使用完毕后，返回对象池。

## 3 性能测试

### 3.1 struct 反序列化

```go
func BenchmarkUnmarshal(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stu := &Student{}
		json.Unmarshal(buf, stu)
	}
}

func BenchmarkUnmarshalWithPool(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stu := studentPool.Get().(*Student)
		json.Unmarshal(buf, stu)
		studentPool.Put(stu)
	}
}
```

测试结果如下：

```bash
$ go test -bench . -benchmem
goos: darwin
goarch: amd64
pkg: example/hpg-sync-pool
BenchmarkUnmarshal-8           1993   559768 ns/op   5096 B/op 7 allocs/op
BenchmarkUnmarshalWithPool-8   1976   550223 ns/op    234 B/op 6 allocs/op
PASS
ok      example/hpg-sync-pool   2.334s
```

在这个例子中，因为 Student 结构体内存占用较小，内存分配几乎不耗时间。而标准库 json 反序列化时利用了反射，效率是比较低的，占据了大部分时间，因此两种方式最终的执行时间几乎没什么变化。但是内存占用差了一个数量级，使用了 `sync.Pool` 后，内存占用仅为未使用的 234/5096 = 1/22，对 GC 的影响就很大了。

### 3.2 bytes.Buffer

```go
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

var data = make([]byte, 10000)

func BenchmarkBufferWithPool(b *testing.B) {
	for n := 0; n < b.N; n++ {
		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Write(data)
		buf.Reset()
		bufferPool.Put(buf)
	}
}

func BenchmarkBuffer(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var buf bytes.Buffer
		buf.Write(data)
	}
}
```

测试结果如下：

```bash
BenchmarkBufferWithPool-8    8778160    133 ns/op       0 B/op   0 allocs/op
BenchmarkBuffer-8             906572   1299 ns/op   10240 B/op   1 allocs/op
```

这个例子创建了一个 `bytes.Buffer` 对象池，而且每次只执行一个简单的 `Write` 操作，存粹的内存搬运工，耗时几乎可以忽略。而内存分配和回收的耗时占比较多，因此对程序整体的性能影响更大。

## 4 在标准库中的应用

### 4.1 fmt.Printf

Go 语言标准库也大量使用了 `sync.Pool`，例如 `fmt` 和 `encoding/json`。

以下是 `fmt.Printf` 的源代码(go/src/fmt/print.go)：

```go
// go 1.13.6

// pp is used to store a printer's state and is reused with sync.Pool to avoid allocations.
type pp struct {
    buf buffer
    ...
}

var ppFree = sync.Pool{
	New: func() interface{} { return new(pp) },
}

// newPrinter allocates a new pp struct or grabs a cached one.
func newPrinter() *pp {
	p := ppFree.Get().(*pp)
	p.panicking = false
	p.erroring = false
	p.wrapErrs = false
	p.fmt.init(&p.buf)
	return p
}

// free saves used pp structs in ppFree; avoids an allocation per invocation.
func (p *pp) free() {
	if cap(p.buf) > 64<<10 {
		return
	}

	p.buf = p.buf[:0]
	p.arg = nil
	p.value = reflect.Value{}
	p.wrappedErr = nil
	ppFree.Put(p)
}

func Fprintf(w io.Writer, format string, a ...interface{}) (n int, err error) {
	p := newPrinter()
	p.doPrintf(format, a)
	n, err = w.Write(p.buf)
	p.free()
	return
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...interface{}) (n int, err error) {
	return Fprintf(os.Stdout, format, a...)
}
```

`fmt.Printf` 的调用是非常频繁的，利用 `sync.Pool` 复用 pp 对象能够极大地提升性能，减少内存占用，同时降低 GC 压力。

> 这个例子来源于：[深度解密 Go 语言之 sync.Pool](https://www.cnblogs.com/qcrao-2018/p/12736031.html)

## 附 推荐与参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)



