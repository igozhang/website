# go-redis

```
env
CentOS Linux release 7.9.2009 (Core)
Docker version 1.13.1, build 7d71120/1.13.1
go version go1.17.8 linux/amd64

https://github.com/go-redis/redis
```

```
yum -y install docker-1.13.1
docker run --name redis507 -p 6379:6379 -d redis:5.0.7
docker run -it --network host --rm redis:5.0.7 redis-cli

# cat redis.go
package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

//Background返回一个非空的Context。 它永远不会被取消，没有值，也没有期限。
//它通常在main函数，初始化和测试时使用，并用作传入请求的顶级上下文。
var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.3.241:6379",
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("连接redis出错，错误信息：%v", err)
	}
        fmt.Println("成功连接redis")

//'*'获取所有key
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(keys)

//获取key类型
	vType, err := rdb.Type(ctx, "igo01").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(vType) //string

}
```

```
# go run redis.go
成功连接redis
[igo01]
string
```
