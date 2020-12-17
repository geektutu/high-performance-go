---
title: for 和 range 的性能比较
seo_title: Go 语言高性能编程
date: 2020-12-01 23:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。本文比较了普通的 for 循环和 range 在不同场景下的性能，并解释了背后的原理：range 迭代时返回迭代值的拷贝，如果每个迭代值占用内存过大，性能将显著地低于 for，将元素类型改为指针，能够解决这一问题。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 常用数据结构
keywords:
- golang
- range
image: post/hpg-string-concat/data-structure.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: for 和 range 的性能比较
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-2/hpg-range.md
---

![high performance go - data structure](hpg-string-concat/data-structure.jpg)

## 1 range 的简单回顾

Go 语言中，range 可以用来很方便地遍历数组(array)、切片(slice)、字典(map)和信道(chan)

### 1.1 array/slice

```go
words := []string{"Go", "语言", "高性能", "编程"}
for i, s := range words {
    words = append(words, "test")
    fmt.Println(i, s)
}
```

输出结果如下：

```bash
0 Go
1 语言
2 高性能
3 编程
```

- 变量 words 在循环开始前，仅会计算一次，如果在循环中修改切片的长度不会改变本次循环的次数。
- 迭代过程中，每次迭代的下标和值被赋值给变量 i 和 s，第二个参数 s 是可选的。
- 针对 nil 切片，迭代次数为 0。

range 还有另一种只遍历下标的写法，这种写法与 for 几乎没什么差异了。

```go
for i := range words {
	fmt.Println(i, words[i])
}
```

输出也是一样的：

```bash
0 Go
1 语言
2 高性能
3 编程
```

### 1.2 map

```go
m := map[string]int{
    "one":   1,
    "two":   2,
    "three": 3,
}
for k, v := range m {
    delete(m, "two")
    m["four"] = 4
    fmt.Printf("%v: %v\n", k, v)
}
```

输出结果为：

```bash
one: 1
four: 4
three: 3
```

- 和切片不同的是，迭代过程中，删除还未迭代到的键值对，则该键值对不会被迭代。
- 在迭代过程中，如果创建新的键值对，那么新增键值对，可能被迭代，也可能不会被迭代。
- 针对 nil 字典，迭代次数为 0

### 1.3 channel

```go
ch := make(chan string)
go func() {
    ch <- "Go"
    ch <- "语言"
    ch <- "高性能"
    ch <- "编程"
    close(ch)
}()
for n := range ch {
    fmt.Println(n)
}
```

- 发送给信道(channel) 的值可以使用 for 循环迭代，直到信道被关闭。
- 如果是 nil 信道，循环将永远阻塞。

## 2 for 和 range 的性能比较

### 2.1 []int

```go
func generateWithCap(n int) []int {
	rand.Seed(time.Now().UnixNano())
	nums := make([]int, 0, n)
	for i := 0; i < n; i++ {
		nums = append(nums, rand.Int())
	}
	return nums
}

func BenchmarkForIntSlice(b *testing.B) {
	nums := generateWithCap(1024 * 1024)
	for i := 0; i < b.N; i++ {
		len := len(nums)
		var tmp int
		for k := 0; k < len; k++ {
			tmp = nums[k]
		}
		_ = tmp
	}
}

func BenchmarkRangeIntSlice(b *testing.B) {
	nums := generateWithCap(1024 * 1024)
	for i := 0; i < b.N; i++ {
		var tmp int
		for _, num := range nums {
			tmp = num
		}
		_ = tmp
	}
}
```

运行结果如下：

```bash
$ go test -bench=IntSlice$ .
goos: darwin
goarch: amd64
pkg: example/hpg-range
BenchmarkForIntSlice-8              3603            324512 ns/op
BenchmarkRangeIntSlice-8            3591            322744 ns/op
```

- `generateWithCap` 用于生成长度为 n 元素类型为 int 的切片。
- 从最终的结果可以看到，遍历 []int 类型的切片，for 与 range 性能几乎没有区别。

### 2.2 []struct

那如果是稍微复杂一点的 `[]struct` 类型呢？

```go
type Item struct {
	id  int
	val [4096]byte
}

func BenchmarkForStruct(b *testing.B) {
	var items [1024]Item
	for i := 0; i < b.N; i++ {
		length := len(items)
		var tmp int
		for k := 0; k < length; k++ {
			tmp = items[k].id
		}
		_ = tmp
	}
}

func BenchmarkRangeIndexStruct(b *testing.B) {
	var items [1024]Item
	for i := 0; i < b.N; i++ {
		var tmp int
		for k := range items {
			tmp = items[k].id
		}
		_ = tmp
	}
}

func BenchmarkRangeStruct(b *testing.B) {
	var items [1024]Item
	for i := 0; i < b.N; i++ {
		var tmp int
		for _, item := range items {
			tmp = item.id
		}
		_ = tmp
	}
}
```

先看下 Benchmark 的结果：

```bash
$ go test -bench=Struct$ .
goos: darwin
goarch: amd64
pkg: example/hpg-range
BenchmarkForStruct-8             3769580               324 ns/op
BenchmarkRangeIndexStruct-8      3597555               330 ns/op
BenchmarkRangeStruct-8              2194            467411 ns/op
```

- 仅遍历下标的情况下，for 和 range 的性能几乎是一样的。
- `items` 的每一个元素的类型是一个结构体类型 `Item`，`Item` 由两个字段构成，一个类型是 int，一个是类型是 `[4096]byte`，也就是说每个 `Item` 实例需要申请约 4KB 的内存。
- 在这个例子中，for 的性能大约是 range (同时遍历下标和值) 的 2000 倍。

### 2.3 []int 和 []struct{} 的性能差异

与 for 不同的是，`range` 对每个迭代值都创建了一个拷贝。因此如果每次迭代的值内存占用很小的情况下，for 和 range 的性能几乎没有差异，但是如果每个迭代值内存占用很大，例如上面的例子中，每个结构体需要占据 4KB 的内存，这种情况下差距就非常明显了。

我们可以用一个非常简单的例子来证明 range 迭代时，返回的是拷贝。

```go
persons := []struct{ no int }{{no: 1}, {no: 2}, {no: 3}}
for _, s := range persons {
    s.no += 10
}
for i := 0; i < len(persons); i++ {
    persons[i].no += 100
}
fmt.Println(persons) // [{101} {102} {103}]
```
- `persons` 是一个长度为 3 的切片，每个元素是一个结构体。
- 使用 `range` 迭代时，试图将每个结构体的 no 字段增加 10，但修改无效，因为 range 返回的是拷贝。
- 使用 `for` 迭代时，将每个结构体的 no 字段增加 100，修改有效。

### 2.4 []*struct{}

那如果切片中是指针，而不是结构体呢？

```go
func generateItems(n int) []*Item {
	items := make([]*Item, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, &Item{id: i})
	}
	return items
}

func BenchmarkForPointer(b *testing.B) {
	items := generateItems(1024)
	for i := 0; i < b.N; i++ {
		length := len(items)
		var tmp int
		for k := 0; k < length; k++ {
			tmp = items[k].id
		}
		_ = tmp
	}
}

func BenchmarkRangePointer(b *testing.B) {
	items := generateItems(1024)
	for i := 0; i < b.N; i++ {
		var tmp int
		for _, item := range items {
			tmp = item.id
		}
		_ = tmp
	}
}
```

运行结果如下：

```bash
goos: darwin
goarch: amd64
pkg: example/hpg-range
BenchmarkForPointer-8             271279              4160 ns/op
BenchmarkRangePointer-8           264068              4194 ns/op
```

切片元素从结构体 `Item` 替换为指针 `*Item` 后，for 和 range 的性能几乎是一样的。而且使用指针还有另一个好处，可以直接修改指针对应的结构体的值。

## 3 总结

range 在迭代过程中返回的是迭代值的拷贝，如果每次迭代的元素的内存占用很低，那么 for 和 range 的性能几乎是一样，例如 `[]int`。但是如果迭代的元素内存占用较高，例如一个包含很多属性的 struct 结构体，那么 for 的性能将显著地高于 range，有时候甚至会有上千倍的性能差异。对于这种场景，建议使用 for，如果使用 range，建议只迭代下标，通过下标访问迭代值，这种使用方式和 for 就没有区别了。如果想使用 range 同时迭代下标和值，则需要将切片/数组的元素改为指针，才能不影响性能。

## 附 推荐与参考

- [4 basic range loop (for-each) patterns](https://yourbasic.org/golang/for-loop-range-array-slice-map-channel)
- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
