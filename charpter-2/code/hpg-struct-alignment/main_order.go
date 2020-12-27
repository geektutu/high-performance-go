package main

import (
	"fmt"
	"unsafe"
)

type demo1 struct {
	a int8
	b int16
	c int32
}

type demo2 struct {
	a int8
	c int32
	b int16
}

func main() {
	fmt.Println(unsafe.Sizeof(demo1{}))
	fmt.Println(unsafe.Sizeof(demo2{}))
}
