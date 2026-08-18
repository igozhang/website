# Go执行顺序及init函数

1. Go执行顺序sequence
2. init函数作用  
   2.1 重置包级变量值  
   2.2 实现对包级变量的复杂初始化  
   2.3 实现“注册模式”（各类驱动）

```
1. Go执行顺序

Go 按照“深度优先”以及“常量 -> 变量 -> init 函数”的顺序进行初始化

a, init在包内其他语法元素的后面执行；
b, 每个init函数在整个Go程序生命周期内仅会被执行一次;
c, init是顺序执行的，当一个执行完毕后，才会去执行下一个init函数。

代码：

package main
​
import (
    "fmt"
​
    _ "github.com/bigwhite/prog-init-order/pkg1"
    _ "github.com/bigwhite/prog-init-order/pkg2"
)
​
var (
    _  = constInitCheck()
    v1 = variableInit("v1")
    v2 = variableInit("v2")
)
​
const (
    c1 = "c1"
    c2 = "c2"
)
​
func constInitCheck() string {
    if c1 != "" {
        fmt.Println("main: const c1 has been initialized")
    }
    if c2 != "" {
        fmt.Println("main: const c2 has been initialized")
    }
    return ""
}
​
func variableInit(name string) string {
    fmt.Printf("main: var %s has been initialized\n", name)
    return name
}
​
func init() {
    fmt.Println("main: first init func invoked")
}
​
func init() {
    fmt.Println("main: second init func invoked")
}
​
func main() {
    // do nothing
}

运行结果：

$go run main.go
pkg3: const c has been initialized
pkg3: var v has been initialized
pkg3: init func invoked
pkg1: const c has been initialized
pkg1: var v has been initialized
pkg1: init func invoked
pkg2: const c has been initialized
pkg2: var v has been initialized
pkg2: init func invoked
main: const c1 has been initialized
main: const c2 has been initialized
main: var v1 has been initialized
main: var v2 has been initialized
main: first init func invoked
main: second init func invoked
```

```
2. init函数作用
2.1 重置包级变量值

flag包在init函数中重置CommandLine的Usage字段：

代码：

var CommandLine = NewFlagSet(os.Args[0], ExitOnError)

func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
    f := &FlagSet{
        name:          name,
        errorHandling: errorHandling,
    }
    f.Usage = f.defaultUsage
    return f
}

func (f *FlagSet) defaultUsage() {
    if f.name == "" {
        fmt.Fprintf(f.Output(), "Usage:\n")
    } else {
        fmt.Fprintf(f.Output(), "Usage of %s:\n", f.name)
    }
    f.PrintDefaults()
}

func init() {
    CommandLine.Usage = commandLineUsage // 重置CommandLine的Usage字段
}

func commandLineUsage() {
    Usage()
}

var Usage = func() {
    fmt.Fprintf(CommandLine.Output(), "Usage of %s:\n", os.Args[0])
    PrintDefaults()
}
```

```
2.2 实现对包级变量的复杂初始化

示例：
标准库http包定义了一系列布尔类型的特性开关变量，默认处于关闭状态（即值为 false），但可以通过GODEBUG环境变量的值，开启相关特性开关

var (
    http2VerboseLogs    bool // 初始化时默认值为false
    http2logFrameWrites bool // 初始化时默认值为false
    http2logFrameReads  bool // 初始化时默认值为false
    http2inTests        bool // 初始化时默认值为false
)

func init() {
    e := os.Getenv("GODEBUG")
    if strings.Contains(e, "http2debug=1") {
        http2VerboseLogs = true // 在init中对http2VerboseLogs的值进行重置
    }
    if strings.Contains(e, "http2debug=2") {
        http2VerboseLogs = true // 在init中对http2VerboseLogs的值进行重置
        http2logFrameWrites = true // 在init中对http2logFrameWrites的值进行重置
        http2logFrameReads = true // 在init中对http2logFrameReads的值进行重置
    }
}
```

```
2.3 实现“注册模式”（各类驱动）

示例2.3.1:
以空导入的方式导入 lib/pq 包实现访问 PostgreSQL 数据库的代码

import (
    "database/sql"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "user=pqgotest dbname=pqgotest sslmode=verify-full")
    if err != nil {
        log.Fatal(err)
    }
    
    age := 21
    rows, err := db.Query("SELECT name FROM users WHERE age = $1", age)
    ...
}

init代码：
func init() {
    sql.Register("postgres", &Driver{})
}

示例2.3.2:
支持 png、jpeg、gif 三种格式的图片

使用标准库 image 包获取各种格式图片的宽和高

package main

import (
    "fmt"
    "image"
    _ "image/gif" // 以空导入方式注入gif图片格式驱动
    _ "image/jpeg" // 以空导入方式注入jpeg图片格式驱动
    _ "image/png" // 以空导入方式注入png图片格式驱动
    "os"
)

func main() {
    // 支持png, jpeg, gif
    width, height, err := imageSize(os.Args[1]) // 获取传入的图片文件的宽与高
    if err != nil {
        fmt.Println("get image size error:", err)
        return
    }
    fmt.Printf("image size: [%d, %d]\n", width, height)
}

func imageSize(imageFile string) (int, int, error) {
    f, _ := os.Open(imageFile) // 打开图文文件
    defer f.Close()

    img, _, err := image.Decode(f) // 对文件进行解码，得到图片实例
    if err != nil {
        return 0, 0, err
    }

    b := img.Bounds() // 返回图片区域
    return b.Max.X, b.Max.Y, nil
}

init代码：

// $GOROOT/src/image/png/reader.go
func init() {
    image.RegisterFormat("png", pngHeader, Decode, DecodeConfig)
}

// $GOROOT/src/image/jpeg/reader.go
func init() {
    image.RegisterFormat("jpeg", "\xff\xd8", Decode, DecodeConfig)
}

// $GOROOT/src/image/gif/reader.go
func init() {
    image.RegisterFormat("gif", "GIF8?a", Decode, DecodeConfig)
}
```
