---
title: Go Reflect 提高反射性能
seo_title: Go 语言高性能编程
date: 2020-12-06 01:00:00
description: Go 语言/golang 高性能编程，Go 语言进阶教程，Go 语言高性能编程(high performance go)。本文介绍了反射的使用场景，并测试了反射的性能，以及某些场景下的替代方式。
tags:
- Go语言高性能编程
nav: 高性能编程
categories:
- 常用数据结构
keywords:
- golang
- reflect
image: post/hpg-string-concat/data-structure.jpg
github: https://github.com/geektutu/high-performance-go
book: Go 语言高性能编程
book_title: 反射(reflect)性能
github_page: https://github.com/geektutu/high-performance-go/blob/master/charpter-2/hpg-reflect.md
---

![high performance go - data structure](hpg-string-concat/data-structure.jpg)

## 1 反射的用途

标准库 [reflect](https://golang.org/pkg/reflect/) 为 Go 语言提供了运行时动态获取对象的类型和值以及动态创建对象的能力。反射可以帮助抽象和简化代码，提高开发效率。

Go 语言标准库以及很多开源软件中都使用了 Go 语言的反射能力，例如用于序列化和反序列化的 `json`、ORM 框架 `gorm/xorm` 等。

在 [7days-golang](https://github.com/geektutu/7days-golang) 这个项目中，也有好几处用到了反射。在 [七天用Go从零实现RPC框架](https://geektutu.com/post/geerpc.html) 中，我们使用反射在服务端，利用接收到的二进制报文动态创建对象，例如利用反射实现函数的动态调用。在 [7天用Go从零实现ORM框架GeeORM](https://geektutu.com/post/geeorm.html) 中，我们使用反射，实现了结构体(struct)类型和数据库表名的映射，结构体字段和数据库字段的映射。同样利用反射动态创建对象的能力，将数据库中查询到的记录转换为 Go 语言中的对象。

## 2 反射如何简化代码

接下来呢，我们利用反射实现一个简单的功能，来看看反射如何帮助我们简化代码的。

假设有一个配置类 Config，每个字段是一个配置项。为了简化实现，假设字段均为 string 类型：

```go
type Config struct {
	Name    string `json:"server-name"`
	IP      string `json:"server-ip"`
	URL     string `json:"server-url"`
	Timeout string `json:"timeout"`
}
```

配置默认从 `json` 文件中读取，如果环境变量中设置了某个配置项，则以环境变量中的配置为准。配置项和环境变量对应的规则非常简单：将 json 字段的字母转为大写，将 `-` 转为下划线，并添加 `CONFIG_` 前缀。

最终的对应结果如下：

```go
type Config struct {
	Name    string `json:"server-name"` // CONFIG_SERVER_NAME
	IP      string `json:"server-ip"`   // CONFIG_SERVER_IP
	URL     string `json:"server-url"`  // CONFIG_SERVER_URL
	Timeout string `json:"timeout"`     // CONFIG_TIMEOUT
}
```

实现这个功能非常简单，使用 `switch case` 或者 `if else` 硬编码很快就搞定了。但是，如果使用硬编码，`Config` 结构发生改变，例如修改 `json` 对应的字段，删除或新增了一个配置项，这块逻辑也需要发生改变。而更大的问题在于：容易出错，不好测试！！！

这个时候，就有了 reflect 的用武之地了。

```go
func readConfig() *Config {
	// read from xxx.json，省略
	config := Config{}
	typ := reflect.TypeOf(config)
	value := reflect.Indirect(reflect.ValueOf(&config))
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if v, ok := f.Tag.Lookup("json"); ok {
			key := fmt.Sprintf("CONFIG_%s", strings.ReplaceAll(strings.ToUpper(v), "-", "_"))
			if env, exist := os.LookupEnv(key); exist {
				value.FieldByName(f.Name).Set(reflect.ValueOf(env))
			}
		}
	}
	return &config
}

func main() {
	os.Setenv("CONFIG_SERVER_NAME", "global_server")
	os.Setenv("CONFIG_SERVER_IP", "10.0.0.1")
	os.Setenv("CONFIG_SERVER_URL", "geektutu.com")
	c := readConfig()
	fmt.Printf("%+v", c)
}
```

实现逻辑其实是非常简单的：

- 在运行时，利用反射获取到 `Config` 的每个字段的 `Tag` 属性，拼接出对应的环境变量的名称。
- 查看该环境变量是否存在，如果存在，则将环境变量的值赋值给该字段。

运行该程序，输出为：

```bash
&{Name:global_server IP:10.0.0.1 URL:geektutu.com Timeout:}
```

可以看到，环境变量中设置的三个配置项已经生效。之后无论结构体 `Config` 内部的字段发生任何改变，这部分代码无需任何修改即可完美的适配，出错概率也极大地降低。

## 3 反射的性能

毫无疑问的是，反射会增加额外的代码指令，对性能肯定会产生影响的。具体影响有多大，我们可以使用 Benchmark 来测试一番。

### 3.1 创建对象

```go
func BenchmarkNew(b *testing.B) {
	var config *Config
	for i := 0; i < b.N; i++ {
		config = new(Config)
	}
	_ = config
}

func BenchmarkReflectNew(b *testing.B) {
	var config *Config
	typ := reflect.TypeOf(Config{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config, _ = reflect.New(typ).Interface().(*Config)
	}
	_ = config
}
```

测试结果如下：

```bash
$ go test -bench .          
goos: darwin
goarch: amd64
pkg: example/hpg-reflect
BenchmarkNew-8                  26478909                40.9 ns/op
BenchmarkReflectNew-8           18983700                62.1 ns/op
PASS
ok      example/hpg-reflect     2.382s
```

通过反射创建对象的耗时约为 `new` 的 1.5 倍，相差不是特别大。

### 3.2 修改字段的值

通过反射获取结构体的字段有两种方式，一种是 `FieldByName`，另一种是 `Field`(按照下标)。前面的例子中，我们使用的是 `FieldByName`。

```go
func BenchmarkSet(b *testing.B) {
	config := new(Config)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Name = "name"
		config.IP = "ip"
		config.URL = "url"
		config.Timeout = "timeout"
	}
}

func BenchmarkReflect_FieldSet(b *testing.B) {
	typ := reflect.TypeOf(Config{})
	ins := reflect.New(typ).Elem()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ins.Field(0).SetString("name")
		ins.Field(1).SetString("ip")
		ins.Field(2).SetString("url")
		ins.Field(3).SetString("timeout")
	}
}

func BenchmarkReflect_FieldByNameSet(b *testing.B) {
	typ := reflect.TypeOf(Config{})
	ins := reflect.New(typ).Elem()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ins.FieldByName("Name").SetString("name")
		ins.FieldByName("IP").SetString("ip")
		ins.FieldByName("URL").SetString("url")
		ins.FieldByName("Timeout").SetString("timeout")
	}
}
```

测试结果如下：

```bash
$ go test -bench="Set$" .          
goos: darwin
goarch: amd64
pkg: example/hpg-reflect
BenchmarkSet-8                          1000000000               0.302 ns/op
BenchmarkReflect_FieldSet-8             33913672                34.5 ns/op
BenchmarkReflect_FieldByNameSet-8        3775234               316 ns/op
PASS
ok      example/hpg-reflect     3.066s
```

- 三种场景下，对象已经提前创建好，测试的均为给字段赋值所消耗的时间。
- 普通的赋值操作，每次耗时约为 0.3 ns，通过下标找到对应的字段再赋值，每次耗时约为 30 ns，通过名称找到对应字段再赋值，每次耗时约为 300 ns。

总结一下，对于一个普通的拥有 4 个字段的结构体 `Config` 来说，使用反射给每个字段赋值，相比直接赋值，性能劣化约 100 - 1000 倍。其中，`FieldByName` 的性能相比 `Field` 劣化 10 倍。

### 3.3 FieldByName 和 Field 性能差距

`FieldByName` 和 `Field` 十倍的性能差距让我对 `FieldByName` 的内部实现比较好奇，打开源代码一探究竟：

- reflect/value.go

```go
// FieldByName returns the struct field with the given name.
// It returns the zero Value if no field was found.
// It panics if v's Kind is not struct.
func (v Value) FieldByName(name string) Value {
	v.mustBe(Struct)
	if f, ok := v.typ.FieldByName(name); ok {
		return v.FieldByIndex(f.Index)
	}
	return Value{}
}
```

- reflect/type.go

```go
func (t *rtype) FieldByName(name string) (StructField, bool) {
	if t.Kind() != Struct {
		panic("reflect: FieldByName of non-struct type")
	}
	tt := (*structType)(unsafe.Pointer(t))
	return tt.FieldByName(name)
}

// FieldByName returns the struct field with the given name
// and a boolean to indicate if the field was found.
func (t *structType) FieldByName(name string) (f StructField, present bool) {
	// Quick check for top-level name, or struct without embedded fields.
	hasEmbeds := false
	if name != "" {
		for i := range t.fields {
			tf := &t.fields[i]
			if tf.name.name() == name {
				return t.Field(i), true
			}
			if tf.embedded() {
				hasEmbeds = true
			}
		}
	}
	if !hasEmbeds {
		return
	}
	return t.FieldByNameFunc(func(s string) bool { return s == name })
}
```

整个调用链条是比较简单的：

```bash
(v Value) FieldByName -> (t *rtype) FieldByName -> (t *structType) FieldByName
```

而 `(t *structType) FieldByName` 中使用 for 循环，逐个字段查找，字段名匹配时返回。也就是说，在反射的内部，字段是按顺序存储的，因此按照下标访问查询效率为 O(1)，而按照 `Name` 访问，则需要遍历所有字段，查询效率为 O(N)。结构体所包含的字段(包括方法)越多，那么两者之间的效率差距则越大。

## 4 如何提高性能

### 4.1 避免使用反射

使用反射赋值，效率非常低下，如果有替代方案，尽可能避免使用反射，特别是会被反复调用的热点代码。例如 RPC 协议中，需要对结构体进行序列化和反序列化，这个时候避免使用 Go 语言自带的 `json` 的 `Marshal` 和 `Unmarshal` 方法，因为标准库中的 json 序列化和反序列化是利用反射实现的。可选的替代方案有 [easyjson](https://github.com/mailru/easyjson)，在大部分场景下，相比标准库，有 5 倍左右的性能提升。

### 4.2 缓存

在上面的例子中可以看到，`FieldByName` 相比于 `Field` 有一个数量级的性能劣化。那在实际的应用中，就要避免直接调用 `FieldByName`。我们可以利用字典将 `Name` 和 `Index` 的映射缓存起来。避免每次反复查找，耗费大量的时间。

我们利用缓存，优化下刚才的测试用例：

```go
func BenchmarkReflect_FieldByNameCacheSet(b *testing.B) {
	typ := reflect.TypeOf(Config{})
	cache := make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		cache[typ.Field(i).Name] = i
	}
	ins := reflect.New(typ).Elem()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ins.Field(cache["Name"]).SetString("name")
		ins.Field(cache["IP"]).SetString("ip")
		ins.Field(cache["URL"]).SetString("url")
		ins.Field(cache["Timeout"]).SetString("timeout")
	}
}
```

测试结果如下：

```bash
$ go test -bench="Set$" . -v
goos: darwin
goarch: amd64
pkg: example/hpg-reflect
BenchmarkSet-8                                  1000000000               0.303 ns/op
BenchmarkReflect_FieldSet-8                     33429990                34.1 ns/op
BenchmarkReflect_FieldByNameSet-8                3612130               331 ns/op
BenchmarkReflect_FieldByNameCacheSet-8          14575906                78.2 ns/op
PASS
ok      example/hpg-reflect     4.280s
```

消耗时间从原来的 10 倍，缩小到了 2 倍。

## 附 推荐与参考

- [Go 语言笔试面试题汇总](https://geektutu.com/post/qa-golang.html)
- [七天用Go从零实现系列](https://geektutu.com/post/gee.html)
