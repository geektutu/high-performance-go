package main

import (
	"fmt"
	"unsafe"
)

type demo3 struct {
	c int32
	a struct{}
}

type demo4 struct {
	a struct{}
	c int32
}

func main() {
	fmt.Println(unsafe.Sizeof(demo3{}))
	fmt.Println(unsafe.Sizeof(demo4{}))
}
