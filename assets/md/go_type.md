# golang_datatype

go 1.17

[normal\_int,str](#intstr)  
[array](#array)  
[slice](#slice)  
[map](#map)  
[string and string\_func](#string)

```
cat /igo/0209/intstr_test.go

package type_test

import "testing"

type MyInt int64

func TestImplicit(t *testing.T) {
	var a int32 = 1
	var b int64
	b = int64(a)
	var c MyInt
	c = MyInt(b)
	t.Log(a, b, c)
}

func TestPoint(t *testing.T) {
	a := 1
	aPtr := &a
        // & 取指符，获得变量a的指针
	//aPtr = aPtr + 1
	t.Log(a, aPtr)
	t.Logf("%T %T", a, aPtr)
}

func TestString(t *testing.T) {
	var s string
	t.Log("*" + s + "*") //初始化零值是“”
	t.Log(len(s))

}

******************运行结果分隔符**********************
=== RUN   TestPoint
    noramltype_test.go:9: 1 0xc00000a248
    noramltype_test.go:10: int *int
--- PASS: TestPoint (0.00s)
=== RUN   TestString
    noramltype_test.go:15: **
    noramltype_test.go:16: 0
--- PASS: TestString (0.00s)

******************运行结果分隔符**********************
```

```
cat /igo/0209/array_test.go

package array_test

import "testing"

func TestArrayInit(t *testing.T) {
	var arr [3]int
	arr1 := [4]int{1, 2, 3, 4}
	arr3 := [...]int{1, 3, 4, 5}
	arr1[1] = 5
	t.Log(arr[1], arr[2])
	t.Log(arr1, arr3)
}

func TestArrayTravel(t *testing.T) {
	arr3 := [...]int{1, 3, 4, 5}
	for i := 0; i < len(arr3); i++ {
		t.Log(arr3[i])
	}
	for _, e := range arr3 {
		t.Log(e)
	}
}

func TestArraySection(t *testing.T) {
	arr3 := [...]int{1, 2, 3, 4, 5}
	arr3_sec := arr3[:]
	t.Log(arr3_sec)
}
```

```
cat /igo/0209/slice_test.go
slice 相当于指针数组

package slice_test

import "testing"

func TestSliceInit(t *testing.T) {
	var s0 []int
	t.Log(len(s0), cap(s0))
	s0 = append(s0, 1)
	t.Log(len(s0), cap(s0))

	s1 := []int{1, 2, 3, 4}
	t.Log(len(s1), cap(s1))

	s2 := make([]int, 3, 5)
	t.Log(len(s2), cap(s2))
	t.Log(s2[0], s2[1], s2[2])
	s2 = append(s2, 1)
	t.Log(s2[0], s2[1], s2[2], s2[3])
	t.Log(len(s2), cap(s2))
}

func TestSliceGrowing(t *testing.T) {
	s := []int{}
	for i := 0; i < 10; i++ {
		s = append(s, i)
		t.Log(len(s), cap(s))
	}
}

func TestSliceShareMemory(t *testing.T) {
	year := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep",
		"Oct", "Nov", "Dec"}
	Q2 := year[3:6]
	t.Log(Q2, len(Q2), cap(Q2))
	summer := year[5:8]
	t.Log(summer, len(summer), cap(summer))
	summer[0] = "Unknow"
	t.Log(Q2)
	t.Log(year)
}

func TestSliceComparing(t *testing.T) {
	a := []int{1, 2, 3, 4}
	b := []int{1, 2, 3, 4}
	// if a == b { //切片只能和nil比较
	// 	t.Log("equal")
	// }
	t.Log(a, b)
}
```

```
go_slice 结果

GOROOT=/usr/local/go #gosetup
GOPATH=/Users/igo/go #gosetup
/usr/local/go/bin/go test -c -o /private/var/folders/0p/379lx5z921lcvpmc7k59c6f40000gn/T/GoLand/___slice_test_go.test go/test #gosetup
/usr/local/go/bin/go tool test2json -t /private/var/folders/0p/379lx5z921lcvpmc7k59c6f40000gn/T/GoLand/___slice_test_go.test -test.v -test.paniconexit0 -test.run ^\QTestSliceInit\E|\QTestSliceGrowing\E|\QTestSliceShareMemory\E|\QTestSliceComparing\E$
=== RUN   TestSliceInit
    slice_test.go:7: 0 0
    slice_test.go:9: 1 1
    slice_test.go:12: 4 4
    slice_test.go:15: 3 5
    slice_test.go:16: 0 0 0
    slice_test.go:18: 0 0 0 1
    slice_test.go:19: 4 5
--- PASS: TestSliceInit (0.00s)
=== RUN   TestSliceGrowing
    slice_test.go:26: 1 1
    slice_test.go:26: 2 2
    slice_test.go:26: 3 4
    slice_test.go:26: 4 4
    slice_test.go:26: 5 8
    slice_test.go:26: 6 8
    slice_test.go:26: 7 8
    slice_test.go:26: 8 8
    slice_test.go:26: 9 16
    slice_test.go:26: 10 16
--- PASS: TestSliceGrowing (0.00s)
=== RUN   TestSliceShareMemory
    slice_test.go:34: [Apr May Jun] 3 9
    slice_test.go:36: [Jun Jul Aug] 3 7
    slice_test.go:38: [Apr May Unknow]
    slice_test.go:39: [Jan Feb Mar Apr May Unknow Jul Aug Sep Oct Nov Dec]
--- PASS: TestSliceShareMemory (0.00s)
=== RUN   TestSliceComparing
    slice_test.go:48: [1 2 3 4] [1 2 3 4]
--- PASS: TestSliceComparing (0.00s)
PASS

Process finished with the exit code 0
```

```
cat /igo/0209/map_test.go

package my_map

import "testing"

func TestInitMap(t *testing.T) {
	m1 := map[int]int{1: 1, 2: 4, 3: 9}
	t.Log(m1[2])
	t.Logf("len m1=%d", len(m1))
	m2 := map[int]int{}
	m2[4] = 16
	t.Logf("len m2=%d", len(m2))
	m3 := make(map[int]int, 10)
	t.Logf("len m3=%d", len(m3))
}

func TestAccessNotExistingKey(t *testing.T) {
	m1 := map[int]int{}
	t.Log(m1[1])
	m1[2] = 0
	t.Log(m1[2])
	m1[3] = 0
	if v, ok := m1[3]; ok {
		t.Logf("Key 3's value is %d", v)
	} else {
		t.Log("key 3 is not existing.")
	}
}

func TestTravelMap(t *testing.T) {
	m1 := map[int]int{1: 1, 2: 4, 3: 9}
	for k, v := range m1 {
		t.Log(k, v)
	}
}

cat map_ext_test.go

package map_ext

import "testing"

func TestMapWithFunValue(t *testing.T) {
	m := map[int]func(op int) int{}
	m[1] = func(op int) int { return op }
	m[2] = func(op int) int { return op * op }
	m[3] = func(op int) int { return op * op * op }
	t.Log(m[1](2), m[2](2), m[3](2))
}

func TestMapForSet(t *testing.T) {
	mySet := map[int]bool{}
	mySet[1] = true
	n := 3
	if mySet[n] {
		t.Logf("%d is existing", n)
	} else {
		t.Logf("%d is not existing", n)
	}
	mySet[3] = true
	t.Log(len(mySet))
	delete(mySet, 1)
	n = 1
	if mySet[n] {
		t.Logf("%d is existing", n)
	} else {
		t.Logf("%d is not existing", n)
	}
}
```

```
cat string_test.go

package string_test

import (
	"testing"
)

func TestString(t *testing.T) {
	var s string
	t.Log(s) //初始化为默认零值“”
	s = "hello"
	t.Log(len(s))
	//s[1] = '3' //string是不可变的byte slice
	//s = "\xE4\xB8\xA5" //可以存储任何二进制数据
	s = "\xE4\xBA\xBB\xFF"
	t.Log(s)
	t.Log(len(s))
	s = "中"
	t.Log(len(s)) //是byte数

	c := []rune(s)
	t.Log(len(c))
	//	t.Log("rune size:", unsafe.Sizeof(c[0]))
	t.Logf("中 unicode %x", c[0])
	t.Logf("中 UTF8 %x", s)
}

func TestStringToRune(t *testing.T) {
	s := "中华人民共和国"
	for _, c := range s {
		t.Logf("%[1]c %[1]x", c)
	}
}

cat string_fun_test.go

package string_test

import (
	"strconv"
	"strings"
	"testing"
)

func TestStringFn(t *testing.T) {
	s := "A,B,C"
	parts := strings.Split(s, ",")
	for _, part := range parts {
		t.Log(part)
	}
	t.Log(strings.Join(parts, "-"))
}

func TestConv(t *testing.T) {
	s := strconv.Itoa(10)
	t.Log("str" + s)
	if i, err := strconv.Atoi("10"); err == nil {
		t.Log(10 + i)
	}
}
```
