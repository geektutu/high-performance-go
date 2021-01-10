package main

import (
	"fmt"
)

func max(num1, num2 int) int {
	if num1 > num2 {
		return num1
	}
	return num2
}

func main() {
	var a, b = 10, 20
	// go func() {
	// 	b, a = a, b
	// }()
	if max(a, b) == a {
		fmt.Println(a)
	}
}
