package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func plusConcat(n int, str string) string {
	s := ""
	for i := 0; i < n; i++ {
		s += str
	}
	return s
}

func sprintfConcat(n int, str string) string {
	s := ""
	for i := 0; i < n; i++ {
		s = fmt.Sprintf("%s%s", s, str)
	}
	return s
}

func builderConcat(n int, str string) string {
	var builder strings.Builder
	// builder.Grow(n * len(str))
	for i := 0; i < n; i++ {
		builder.WriteString(str)
	}
	return builder.String()
}

func bufferConcat(n int, s string) string {
	buf := new(bytes.Buffer)
	for i := 0; i < n; i++ {
		buf.WriteString(s)
	}
	return buf.String()
}

func byteConcat(n int, str string) string {
	buf := make([]byte, 0)
	for i := 0; i < n; i++ {
		buf = append(buf, str...)
	}
	return string(buf)
}

func preByteConcat(n int, str string) string {
	buf := make([]byte, 0, n*len(str))
	for i := 0; i < n; i++ {
		buf = append(buf, str...)
	}
	return string(buf)
}

func benchmark(b *testing.B, f func(int, string) string) {
	var str = randomString(10)
	for i := 0; i < b.N; i++ {
		f(10000, str)
	}
}

func BenchmarkPlusConcat(b *testing.B)    { benchmark(b, plusConcat) }
func BenchmarkSprintfConcat(b *testing.B) { benchmark(b, sprintfConcat) }
func BenchmarkBuilderConcat(b *testing.B) { benchmark(b, builderConcat) }
func BenchmarkBufferConcat(b *testing.B)  { benchmark(b, bufferConcat) }
func BenchmarkByteConcat(b *testing.B)    { benchmark(b, byteConcat) }
func BenchmarkPreByteConcat(b *testing.B) { benchmark(b, preByteConcat) }

// $ go test -bench="Concat$" -benchmem .
// BenchmarkPlusConcat-8                 19          56289919 ns/op        530998079 B/op     10026 allocs/op
// BenchmarkSprintfConcat-8              10         112229496 ns/op        833648087 B/op     37435 allocs/op
// BenchmarkBuilderConcat-8            8901            125147 ns/op          522227 B/op         23 allocs/op
// BenchmarkBufferConcat-8             8130            144149 ns/op          423539 B/op         13 allocs/op
// BenchmarkByteConcat-8               8984            124875 ns/op          628724 B/op         24 allocs/op
// BenchmarkPreByteConcat-8           17379             65931 ns/op          212993 B/op          2 allocs/op

func TestBuilderConcat(t *testing.T) {
	var str = randomString(10)
	var builder strings.Builder
	cap := 0
	for i := 0; i < 10000; i++ {
		if builder.Cap() != cap {
			fmt.Print(builder.Cap(), " ")
			cap = builder.Cap()
		}
		builder.WriteString(str)
	}
}

// go test -run="TestBuilderConcat" . -v
// 16 32 64 128 256 512 1024 2048 2688 3456 4864 6144 8192 10240 13568 18432 24576 32768 40960 57344 73728 98304 122880 --- PASS: TestBuilderConcat (0.00s)
