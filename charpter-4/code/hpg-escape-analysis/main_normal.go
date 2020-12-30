// main_normal.go
package main

import "fmt"

type Demo struct {
	name string
}

func createDemo(name string) *Demo {
	d := new(Demo) // 局部变量 d 逃逸到堆
	d.name = name
	return d
}

func test(demo *Demo) {
	fmt.Println(demo.name)
}

func main() {
	demo := createDemo("demo")
	test(demo)
}
