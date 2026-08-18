# GoModule_CRUD

1.[添加依赖](#1)  
2.[升级 / 降级依赖的版本](#2)  
3.[添加一个主版本号大于 1 的依赖](#3)  
4.[升级依赖版本到一个不兼容版本](#4)  
5.[移除一个依赖](#5)  
6.[使用vendor](#6)

```
1.添加依赖

代码：

package main

import (
  "github.com/google/uuid"
  "github.com/sirupsen/logrus"
)

func main() {
  logrus.Println("hello, go module mode")
  logrus.Println(uuid.NewString())
}

源码中，通过 import 语句导入了 github.com/google/uuid，并在 main 函数中调用了 uuid 包的函数 NewString。

添加依赖包：
$go get github.com/google/uuid
或者
$go mod tidy
```

```
2.升级 / 降级依赖的版本

假设go mod tidy 命令选择了 logrus 的最新发布版本 v1.8.1。如果需要将 logrus 版本降至某个之前发布的兼容版本，比如 v1.7.0

GoModule_降级：
$go get github.com/sirupsen/logrus@v1.7.0
go: downloading github.com/sirupsen/logrus v1.7.0
go get: downgraded github.com/sirupsen/logrus v1.8.1 => v1.7.0

也可以：
$go mod edit -require=github.com/sirupsen/logrus@v1.7.0
$go mod tidy       
go: downloading github.com/sirupsen/logrus v1.7.0

GoModule升级：

假设logrus v1.7.1修复了安全漏洞，我们需要升级

$go get github.com/sirupsen/logrus@v1.7.1
go: downloading github.com/sirupsen/logrus v1.7.1
go get: upgraded github.com/sirupsen/logrus v1.7.0 => v1.7.1
```

```
3.添加一个主版本号大于 1 的依赖

在 Go Module 构建模式下，当依赖的主版本号为 0 或 1 的时候，我们在 Go 源码中导入依赖包，不需要在包的导入路径上增加版本号：

import github.com/user/repo/v0 等价于 import github.com/user/repo
import github.com/user/repo/v1 等价于 import github.com/user/repo

如果我们要为项目引入主版本号大于 1 的依赖，比如 v2.0.0，那么由于这个版本与 v1、v0 开头的包版本都不兼容，我们在导入时，不能再使用 github.com/user/repo，而要使用像下面代码中那样不同的包导入路径：

import github.com/user/repo/v2/xxx

示例：
(以空导入的方式导入v7版本的github.com/go-redis/redis包)

package main

import (
  _ "github.com/go-redis/redis/v7" // “_”为空导入
  "github.com/google/uuid"
  "github.com/sirupsen/logrus"
)

func main() {
  logrus.Println("hello, go module mode")
  logrus.Println(uuid.NewString())
}

$go get github.com/go-redis/redis/v7
go: downloading github.com/go-redis/redis/v7 v7.4.1
go: downloading github.com/go-redis/redis v6.15.9+incompatible
go get: added github.com/go-redis/redis/v7 v7.4.1
```

```
4.升级依赖版本到一个不兼容版本(主版本号不一致)

示例：
(redis从V7升级到V8)

4.1 改代码里倒入路径的版本号

import (
  _ "github.com/go-redis/redis/v8"
  "github.com/google/uuid"
  "github.com/sirupsen/logrus"
)

4.2  go get

$go get github.com/go-redis/redis/v8
go: downloading github.com/go-redis/redis/v8 v8.11.1
go: downloading github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
go: downloading github.com/cespare/xxhash/v2 v2.1.1
go get: added github.com/go-redis/redis/v8 v8.11.1
```

```
5.移除一个依赖

示例：
(移除redis)

5.1
移除代码中的go-redis/redis，然后go build 项目，会发现
go.mod 的require字段以及
go list -m all 列出所有依赖
redis均继续存在

5.2
go mod tidy
将这个依赖项彻底从 Go Module 构建上下文中清除掉。go mod tidy 会自动分析源码依赖，而且将不再使用的依赖从 go.mod 和 go.sum 中移除。
```

```
6.使用vendor

vendor 机制虽然诞生于 GOPATH 构建模式主导的年代，但在 Go Module 构建模式下，它依旧被保留了下来，并且成为了 Go Module 构建机制的一个很好的补充。特别是在一些不方便访问外部网络，并且对 Go 应用构建性能敏感的环境，比如在一些内部的持续集成或持续交付环境 (CI/CD) 中，使用 vendor 机制可以实现与 Go Module 等价的构建。和 GOPATH 构建模式不同，Go Module 构建模式下，我们再也无需手动维护 vendor 目录下的依赖包了，Go 提供了可以快速建立和更新 vendor 的命令，我们还是以前面的 module-mode 项目为例，通过下面命令为该项目建立 vendor：

$go mod vendor
$tree -LF 2 vendor
vendor
├── github.com/
│   ├── google/
│   ├── magefile/
│   └── sirupsen/
├── golang.org/
│   └── x/
└── modules.txt

go mod vendor 命令在 vendor 目录下，创建了一份这个项目的依赖包的副本，并且通过 vendor/modules.txt 记录了 vendor 下的 module 以及版本.
```
