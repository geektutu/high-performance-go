package main

import (
	"fmt"
	"testing"
)

func TestRange(t *testing.T) {
	words := []string{"Go", "语言", "高性能", "编程"}
	for i, s := range words {
		words = append(words, "test")
		fmt.Println(i, s)
	}
	m := map[string]int{
		true: 1, false: 0
	}
	for k, v := range m {
		fmt.Printf("%v: %v\n", k, v)
	}
}
