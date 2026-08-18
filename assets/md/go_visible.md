# go显式

Go 设计哲学：简单、显式、组合、并发和面向工程

对比C的隐式转换说明

```
可编译的C代码：（隐式转换）

#include <stdio.h>

int main() {
    short int a = 5;

    int b = 8;
    long c = 0;
    
    c = a + b;
    printf("%ld\n", c);
}
```

```
对应的Go代码：（编译报错“invalid operation: a + b (mismatched types int16 and int)”）

package main

import "fmt"

func main() {
    var a int16 = 5
    var b int = 8
    var c int64

    c = a + b
    fmt.Printf("%d\n", c)
}
```

```
正确的Go代码：（显式转换）

c = int64(a) + int64(b)
fmt.Printf("%d\n", c)
```
