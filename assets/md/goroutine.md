# go并发编程，协程

go version go1.17.7 Mac

```
% cat goroutine_test.go
package groutine_test

import (
	"fmt"
	"testing"
	"time"
)

func TestGroutine(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func(i int) {
			//time.Sleep(time.Second * 1)
			fmt.Println(i)
		}(i)
	}
	//time.Sleep(time.Millisecond * 50) 为了避免运行太快可以加sleep
}

结果：
igo@igoMacBook golang % go test goroutine_test.go
ok  	command-line-arguments	0.575s

IDE-JetBrains_GoLand结果：
/usr/local/go/bin/go tool test2json -t /private/var/folders/0p/379lx5z921lcvpmc7k59c6f40000gn/T/GoLand/___goroutine_test_go.test -test.v -test.paniconexit0 -test.run ^\QTestGroutine\E$
=== RUN   TestGroutine
0
6
2
7
5
8
3
9
1
--- PASS: TestGroutine (0.00s)
4
PASS
```
