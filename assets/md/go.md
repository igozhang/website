# golang

```
Go 语言的创始人有三位：
分别是图灵奖获得者、C 语法联合发明人、Unix 之父肯·汤普森（Ken Thompson），Plan 9 操作系统领导者；
UTF-8 编码的最初设计者罗伯·派克（Rob Pike）；
Java 的 HotSpot 虚拟机和 Chrome 浏览器的 JavaScript V8 引擎的设计者之一罗伯特·格瑞史莫（Robert Griesemer）

Go实现包括：
Docker、Kubernetes、Prometheus、Ethereum（以太坊）、Istio、CockroachDB、InfluxDB、Terraform、Etcd、Consul

GoLand
Visual Studio Code（vs code）加vscode-go插件
```

Go笔记，go version go1.17.6 windows/amd64

```
蔡超老师，极客，Mobvista总裁兼架构师，亚马逊架构师
tonyBai 白明老师 https://github.com/bigwhite/gopherdaily  tonybai.com

与其语言差异
main 主函数不支持传参，不支持返回值
如需传参可以使用 os.Args
如需返回值可以使用 os.Exit

go 不支持隐式类型转换

package 管理定义模块，使用 import 关键字导入使用
如果go 自带的包，则会去安装目录 $GOROOT/src 按包路径加载，如 fmt 包
如果是 go get 安装或自定义的包，则会去 $GOPATH/src 下加载
```

```
go_ins
wget http://go.dev/dl/go1.17.8.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.8.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
~/.bashrc
go version
```

```
go_helloworld

规则1:
整个 Go 程序中仅允许存在一个名为 main 的包，Go 程序本质上就是一组包的集合
规则2:
函数声明参数用(),函数体需被{}包裹

package main  //定义main包，
import "fmt“  //import 导入标准库 fmt 目录下的包；
func main() {   
//定义main函数
    fmt.Println("hello, world")
//调用fmt包中的首字母大写的Println函数（首字母小写函数仅包内可用
//Go采用符合UTF-8标准编码的Unicode字符集来实现helloworld的所见即所得
）
}

go run helloworld.go
编译
go build hello.go
./hello
```
