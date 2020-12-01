package main

import (
	"fmt"
	"testing"
)

func TestArrayAssign(t *testing.T) {
	a := [3]int{1, 2, 3}
	b := [4]int{2, 4, 5, 6}
	fmt.Println(a, b)
	// a = b // cannot use b (type [4]int) as type [3]int in assignment
}

func TestArrayCopy(t *testing.T) {
	a := [...]int{1, 2, 3}
	b := a
	a[0] = 100
	fmt.Println(a, b)
}

func square(arr *[3]int) {
	for i, num := range *arr {
		(*arr)[i] = num * num
	}
}

func TestArrayPointer(t *testing.T) {
	a := [...]int{1, 2, 3}
	square(&a)
	fmt.Println(a)
	if a[1] != 4 && a[2] != 9 {
		t.Fatal("failed")
	}
}
