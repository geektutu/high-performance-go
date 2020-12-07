---
title: Go 语言陷阱 - 数组和切片
seo_title: Go 语言高性能编程
date: 2020-12-07 01:00:00
description: Go 语言/golang 高性能编程(high performance go)，Go 语言进阶教程，Go 语言陷阱(gotchas)。这篇文章介绍了 Go 语言中数组(Array) 和切片(Slice)的常见陷阱和规避方式。例如数组作为参数，修改参数，原数组不会发生改变。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 语言陷阱
keywords:
- golang
- array
- 值类型
- value type
image: post/hpg-gotchas-array-slice/gotchas.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: 数组和切片
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-a1/hpg-gotchas-array-slice.md
---

![golang gotchas](hpg-gotchas-array-slice/gotchas.jpg)

## 1 第一个陷阱

### 1.1 下面程序的输出是

```go
func foo(a [2]int) {
	a[0] = 200
}

func main() {
	a := [2]int{1, 2}
	foo(a)
	fmt.Println(a)
}
```

### 1.2 答案

正确的输出是 `[1 2]`，数组 `a` 没有发生改变。

- 在 Go 语言中，数组是一种值类型，而且不同长度的数组属于不同的类型。例如 `[2]int` 和 `[20]int` 属于不同的类型。
- 当值类型作为参数传递时，参数是该值的一个拷贝，因此更改拷贝的值并不会影响原值。

我们在 [切片(slice)性能及陷阱](https://geektutu.com/post/hpg-slice.html) 这篇文章中也提到了，为了避免数组的拷贝，提高性能，建议传递数组的指针作为参数，或者使用切片代替数组。


### 1.3 更多

如果将上述程序替换为：

```go
func foo(a *[2]int) {
	(*a)[0] = 200
}

func main() {
	a := [2]int{1, 2}
	foo(&a)
	fmt.Println(a)
}
```

或

```go
func foo(a []int) {
	a[0] = 200
}

func main() {
	a := []int{1, 2}
	foo(a)
	fmt.Println(a)
}
```

输出将会变成 `[200 2]`。

在 [切片(slice)性能及陷阱](https://geektutu.com/post/hpg-slice.html) 这篇文章中，我们也提到了切片由三个值构成：

- `*ptr` 指向底层数组的指针
- `len` 长度
- `cap` 容量

因此，将切片作为参数时，拷贝了一个新切片，即拷贝了构成切片的三个值，包括底层数组的指针。对切片中某个元素的修改，实际上是修改了底层数组中的值，因此原切片也发生了改变。


## 2 第二个陷阱

### 2.1 下面程序的输出是

```go
func foo(a []int) {
	a = append(a, 1, 2, 3, 4, 5, 6, 7, 8)
	a[0] = 200
}

func main() {
	a := []int{1, 2}
	foo(a)
	fmt.Println(a)
}
```

### 2.2 答案

输出仍是 `[1 2]`，切片 `a` 没有发生改变。

传参时拷贝了新的切片，因此当新切片的长度发生改变时，原切片并不会发生改变。而且在函数 `foo` 中，新切片 `a` 增加了 8 个元素，原切片对应的底层数组不够放置这 8 个元素，因此申请了新的空间来放置扩充后的底层数组。这个时候新切片和原切片指向的底层数组就不是同一个了。因此，对新切片第 0 个元素的修改，并不会影响原切片的第 0 个元素。

如果如果希望 `foo` 函数的操作能够影响原切片呢？

两种方式：

- 设置返回值，将新切片返回并赋值给 `main` 函数中的变量 `a`。
- 切片也使用指针方式传参。

```go
func foo(a []int) []int {
	a = append(a, 1, 2, 3, 4, 5, 6, 7, 8)
	a[0] = 200
	return a
}

func main() {
	a := []int{1, 2}
	a = foo(a)
	fmt.Println(a)
}
```

或

```go
func foo(a *[]int) {
	*a = append(*a, 1, 2, 3, 4, 5, 6, 7, 8)
	(*a)[0] = 200
}

func main() {
	a := []int{1, 2}
	foo(&a)
	fmt.Println(a)
}
```

上述两个程序的输出均为：

```bash
[200 2 1 2 3 4 5 6 7 8]
```

从可读性上来说，更推荐第一种方式。

## 附 推荐与参考

- [Array won’t change](https://yourbasic.org/golang/gotcha-function-doesnt-change-array/)
- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)