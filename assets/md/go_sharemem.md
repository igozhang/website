# go共享内存机制及协程锁

go version go1.17.7 Mac

```
package share_mem

import (
	"sync"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {

	counter := 0
	for i := 0; i < 5000; i++ {
		go func() {
			counter++
		}()
	}
	time.Sleep(1 * time.Second)
	t.Logf("counter = %d", counter)

}

func TestCounterThreadSafe(t *testing.T) {
	var mut sync.Mutex
	counter := 0
	for i := 0; i < 5000; i++ {
		go func() {
			defer func() {
				mut.Unlock()
			}()
			mut.Lock()
			counter++
		}()
	}
	time.Sleep(1 * time.Second)
	t.Logf("counter = %d", counter)

}

func TestCounterWaitGroup(t *testing.T) {
	var mut sync.Mutex
	var wg sync.WaitGroup
	counter := 0
	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				mut.Unlock()
			}()
			mut.Lock()
			counter++
			wg.Done()
		}()
	}
	wg.Wait()
	t.Logf("counter = %d", counter)

}
```

```
结果Mac：
igo@igoMacBook 0219 % go test share_mem_test.go
ok  	command-line-arguments	2.576s
```

```
结果goland：
/usr/local/go/bin/go tool test2json -t /private/var/folders/0p/379lx5z921lcvpmc7k59c6f40000gn/T/GoLand/___share_mem_test_go.test -test.v -test.paniconexit0 -test.run ^\QTestCounter\E|\QTestCounterThreadSafe\E|\QTestCounterWaitGroup\E$
=== RUN   TestCounter
    share_mem_test.go:18: counter = 4507
--- PASS: TestCounter (1.00s)
=== RUN   TestCounterThreadSafe
    share_mem_test.go:35: counter = 5000
--- PASS: TestCounterThreadSafe (1.00s)
=== RUN   TestCounterWaitGroup
    share_mem_test.go:55: counter = 5000
--- PASS: TestCounterWaitGroup (0.00s)
PASS

Process finished with the exit code 0
```
