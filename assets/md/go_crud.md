# go_crud

```
go env -w GOPROXY=https://goproxy.cn,direct
会写到:
/root/.config/go/env:GOPROXY=https://goproxy.cn,direct

export GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=on

go mod init gitee.com/igozhang/helloworld
go mod init helloworld

go mod .
go mod tidy
模块间互相调用时,如果远端没有代码需要使用本地代码:
go mod edit -replace gitee.com/igozhang/greetings=../greetings
```
