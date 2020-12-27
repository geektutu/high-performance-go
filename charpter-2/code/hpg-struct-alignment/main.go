package main

import (
	"fmt"
	"unsafe"
)

type Args struct {
	num1 int
	num2 int
}

type Flag struct {
	num1 int16
	num2 int32
}

func main() {
	fmt.Println(unsafe.Sizeof(Args{}))
	fmt.Println(unsafe.Sizeof(Flag{}))
	fmt.Println(unsafe.Alignof(Args{}))
	fmt.Println(unsafe.Alignof(Flag{}))
}
