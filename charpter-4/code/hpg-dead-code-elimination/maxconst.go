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

const a, b = 10, 20

func main() {
	if max(a, b) == a {
		fmt.Println(a)
	}
}
