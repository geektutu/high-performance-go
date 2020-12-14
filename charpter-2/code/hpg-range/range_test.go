package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestRange(t *testing.T) {
	// slice
	words := []string{"Go", "语言", "高性能", "编程"}
	for i, s := range words {
		// words = append(words, "test")
		fmt.Println(i, s)
	}
	for i := range words {
		fmt.Println(i, words[i])
	}
	// map
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
	// channel
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
}

func TestModifyForAndRange(t *testing.T) {
	persons := []struct{ no int }{{no: 1}, {no: 2}, {no: 3}}
	for _, s := range persons {
		s.no += 10
	}
	for i := 0; i < len(persons); i++ {
		persons[i].no += 100
	}
	fmt.Println(persons)
}

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
