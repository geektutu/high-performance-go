package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func printLenCap(nums []int) {
	fmt.Printf("len: %d, cap: %d %v\n", len(nums), cap(nums), nums)
}
func TestSliceLenAndCap(t *testing.T) {
	nums := []int{1}
	printLenCap(nums) // len: 1, cap: 1 [1]
	nums = append(nums, 2)
	printLenCap(nums) // len: 2, cap: 2 [1 2]
	nums = append(nums, 3)
	printLenCap(nums) // len: 3, cap: 4 [1 2 3]
	nums = append(nums, 3)
	printLenCap(nums) // len: 4, cap: 4 [1 2 3 3]
}

func TestSlice(t *testing.T) {
	nums := make([]int, 0, 8)
	nums = append(nums, 1, 2, 3, 4, 5)
	nums2 := nums[2:4]
	printLenCap(nums)  // len: 5, cap: 8 [1 2 3 4 5]
	printLenCap(nums2) // len: 2, cap: 6 [3 4]

	nums2 = append(nums2, 50, 60)
	printLenCap(nums)  // len: 5, cap: 8 [1 2 3 4 50]
	printLenCap(nums2) // len: 4, cap: 6 [3 4 50 60]
}

func lastNumsBySlice(origin []int) []int {
	return origin[len(origin)-2:]
}

func lastNumsByCopy(origin []int) []int {
	result := make([]int, 2)
	copy(result, origin[len(origin)-2:])
	return result
}

func generateWithCap(n int) []int {
	rand.Seed(time.Now().UnixNano())
	nums := make([]int, 0, n)
	for i := 0; i < n; i++ {
		nums = append(nums, rand.Int())
	}
	return nums
}

func printMem(t *testing.T) {
	t.Helper()
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	t.Logf("%.2f MB", float64(rtm.Alloc)/1024./1024.)
}

func testLastChars(t *testing.T, f func([]int) []int) {
	t.Helper()
	ans := make([][]int, 0)
	for k := 0; k < 100; k++ {
		origin := generateWithCap(128 * 1024) // 1M
		ans = append(ans, f(origin))
		runtime.GC()
	}
	printMem(t)
	_ = ans
}

func TestLastCharsBySlice(t *testing.T) { testLastChars(t, lastNumsBySlice) }
func TestLastCharsByCopy(t *testing.T)  { testLastChars(t, lastNumsByCopy) }
