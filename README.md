# Go 语言高性能编程

[![high performance with go](charpter-0/high-performance-go/high-performance-go.jpg)](https://geektutu.com/post/high-performance-go.html)

## 订阅

最新动态可以关注：知乎 [Go语言](https://www.zhihu.com/people/gzdaijie) 或微博 [极客兔兔](https://weibo.com/geektutu)

订阅方式：**watch** [geektutu/blog](https://github.com/geektutu/blog) ，每篇文章都能收到邮件通知，或通过 [RSS](https://geektutu.com/feed.xml) 订阅。

## 目录

- 序言
    - [关于本书](https://geektutu.com/post/high-performance-go.html)

- 第一章 性能分析
    - [benchmark 基准测试](https://geektutu.com/post/hpg-benchmark.html)
    - [pprof 性能分析](https://geektutu.com/post/hpg-pprof.html)

- 第二章 常用数据结构
    - [字符串拼接性能及原理](https://geektutu.com/post/hpg-string-concat.html)
    - [切片(slice)性能及陷阱](https://geektutu.com/post/hpg-slice.html)
    - [for 和 range 的性能比较](https://geektutu.com/post/hpg-range.html)
    - [反射(reflect)性能](https://geektutu.com/post/hpg-reflect.html)
    - [使用空结构体节省内存](https://geektutu.com/post/hpg-empty-struct.html)
    - [内存对齐对性能的影响](https://geektutu.com/post/hpg-struct-alignment.html)

- 第三章 并发编程
    - [读写锁和互斥锁的性能比较](https://geektutu.com/post/hpg-mutex.html)
    - [如何退出协程(超时场景)](https://geektutu.com/post/hpg-timeout-goroutine.html)
    - [如何退出协程(其他场景)](https://geektutu.com/post/hpg-exit-goroutine.html)
    - [控制协程的并发数量](https://geektutu.com/post/hpg-concurrency-control.html)
    - [sync.Pool 复用对象](https://geektutu.com/post/hpg-sync-pool.html)
    - [sync.Once 如何提升性能](https://geektutu.com/post/hpg-sync-once.html)
    - [sync.Cond 条件变量](https://geektutu.com/post/hpg-sync-cond.html)

- 第四章 编译优化
    - [减小编译体积](https://geektutu.com/post/hpg-reduce-size.html)
    - [逃逸分析对性能的影响](https://geektutu.com/post/hpg-escape-analysis.html)
    - [死码消除与调试模式](https://geektutu.com/post/hpg-dead-code-elimination.html)

- 附录 Go 语言陷阱
    - [数组和切片](https://geektutu.com/post/hpg-gotchas-array-slice.html)

## 基础入门

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go Test 单元测试简明教程](https://geektutu.com/post/quick-go-test.html)
- [Go Protobuf 简明教程](https://geektutu.com/post/quick-go-protobuf.html)
- [Go RPC & TLS 鉴权简明教程](https://geektutu.com/post/quick-go-rpc.html)
- [Go Mock (gomock)简明教程](https://geektutu.com/post/quick-gomock.html)
- [Go Mmap 文件内存映射简明教程](https://geektutu.com/post/quick-go-mmap.html)
- [Go Context 并发编程简明教程](https://geektutu.com/post/quick-go-context.html)
- [Go WebAssembly (Wasm) 简明教程](https://geektutu.com/post/quick-go-wasm.html)
- [Go Gin 简明教程](https://geektutu.com/post/quick-go-gin.html)

## 进阶系列

- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
    - [Web框架Gee](https://geektutu.com/post/gee.html)
    - [分布式缓存GeeCache](https://geektutu.com/post/geecache.html)
    - [ORM框架GeeORM](https://geektutu.com/post/geeorm.html)
    - [RPC框架GeeRPC](https://geektutu.com/post/geerpc.html)
    - [项目地址](https://github.com/geektutu/7days-golang)
- [Go 语言笔试面试题](https://geektutu.com/post/qa-golang.html)
    - [基础语法](https://geektutu.com/post/qa-golang-1.html)
    - [实现原理](https://geektutu.com/post/qa-golang-2.html)
    - [并发编程](https://geektutu.com/post/qa-golang-3.html)
    - [代码输出](https://geektutu.com/post/qa-golang-c1.html)
