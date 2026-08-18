# go_troubleShooting,go排错

```
go 1.17
报错：
% go run gorouter.go
gorouter.go:8:2: no required module provides package github.com/julienschmidt/httprouter: go.mod file not found in current directory or any parent directory; see 'go help modules'

解决：
%  go mod init gorouter.go
go: creating new go.mod: module gorouter.go
```

```
go 1.17
Mac Monterey
报错：
装包超时
% go get -u github.com/julienschmidt/httprouter
//也可以使用 go mod tidy
go get: module github.com/julienschmidt/httprouter: Get "https://proxy.golang.org/github.com/julienschmidt/httprouter/@v/list": dial tcp 142.251.42.241:443: i/o timeout

解决：
% export GOPROXY=https://goproxy.cn
```

```
问题：
如何import自己本地尚未提交到github的moudle？

解答：
在Go 1.17版本及之前版本的解决方法是使用go mod的replace指示符(directive)。
假如你的module a要import的module b将发布到github.com/user/repo中，那么你可以手动在module的go.mod中的require块中手工加上一条：

require github.com/user/repo v1.0.0

注意v1.0.0这个版本号是一个临时的版本号；

然后在module a的go.mod中使用replace将上面对module b的require替换为本地的module b:

replace github.com/user/repo v1.0.0 => module b本地路径

这样go命令就会使用你本地正在开发、尚未提交github的module b了。
```

```
go1.17.7
报错：
# go build bookstore/cmd/bookstore
package bookstore/cmd/bookstore is not in GOROOT (/usr/local/go/src/bookstore/cmd/bookstore)

解决：
[root@igo-bookstore]# go env -w GOPATH=/golang
[root@igo-bookstore]# go build bookstore/cmd/bookstore
go: downloading github.com/gorilla/mux v1.8.0

$ go help gopath | grep -B 1 -A 4 "GOPATH and Modules"

GOPATH and Modules

When using modules, GOPATH is no longer used for resolving imports.
However, it is still used to store downloaded source code (in GOPATH/pkg/mod)
and compiled commands (in GOPATH/bin).
如上，开启模块后，GOPATH （默认是 ~/go）不再被用于解析包的导入， 也就是 go tool 不会从 GOPATH 中寻找应的包

所以更好的解决办法是
go env -w GO111MODULE=on
```

```
go1.17.7
报错:
# go run hidden.go
package command-line-arguments is not a main package

解决:
将package名字改成main解决
```

```
go1.17.7
报错:
# go run hidden.go
# command-line-arguments
./hidden.go:15:13: more than one character in rune literal

解决:
将字符串单引号改成双引号解决
fmt.Println("after foo",a)
```

```
go1.17.7
报错:
# go get -u https://github.com/jiankunking/zookeeper_exporter
go get: malformed module path "https:/github.com/jiankunking/zookeeper_exporter": invalid char ':'

解决:(路径不要带:)
# go get -u github.com/jiankunking/zookeeper_exporter
```

```
go1.17.7
报错:
./t.go:11:35: invalid character U+3000 in identifier

解决:
代码中有错误字符比如中文;
\n \t \r 等字符串删除，用代码编辑器的\t \n \r替换;
```

```
go1.17.7
报错:
./t.go:19:11: no new variables on left side of :=

原因:
This error means that you are trying to use Go's short form variable declaration ( a := "some-value" ) when the variable has already been declared;
变量重复声明错误;
```
