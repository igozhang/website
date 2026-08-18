# go代码片段备查

```
01 传参,Case
02 exec.pipe
03 导入本地包
04 shell管道命令
```

```
01 传参,Case
package main

import (
    "os"
    "fmt"
    "os/exec"
)

func main() {

//01 打印多行字符串
menu :=`
A simple funclist for shell scripts:
1. ls
2. tree
`
fmt.Printf("%s", menu)

//02 带参数运行
//args := os.Args //获取输入的所有参数,类型是[]string
//argone := args[1] //获取输入的第一个参数
//fmt.Println("the args01 you input is:",argone)	

//03 说明switch的用法
//04 比较CombinedOutput和Output区别
//exit status 2  Output只打印错误码;
//ls: cannot access /noneDir: No such file or directory; CombinedOutput打印具体错误;

//交互传参
var  choice string
fmt.Println("Please Enter Your Choice: \n")
fmt.Scanln(&choice)

switch argone {
    case "1" :
        cmd := exec.Command("ls","/noneDir")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))

    case "2" :
	cmd01 := exec.Command("ls","/noneDir")
	out01, err01 := cmd01.Output()
	if err01 != nil {
		fmt.Println(err01)
	}
	fmt.Println(string(out01))

    default:
        fmt.Println("Wrong! Your Input Should Be 1-2 !")
}
}
```

```
02 exec.pipe
# cat pipe.go

package main
 
import (
  "fmt"
  "io/ioutil"
  "os/exec"
)
 
func main(){
  cmd := exec.Command("/bin/bash", "-c", "grep test")
 
  stdin, _ := cmd.StdinPipe()
  stdout, _ := cmd.StdoutPipe()
 
  if err := cmd.Start(); err != nil{
    fmt.Println("Execute failed when Start:" + err.Error())
    return
  }
 
  stdin.Write([]byte("igo text for grep\n"))
  stdin.Write([]byte("igo test text for grep\n"))
  stdin.Write([]byte("igo test text for grep 02\n"))
  stdin.Close()
 
  out_bytes, _ := ioutil.ReadAll(stdout)
  stdout.Close()
 
  if err := cmd.Wait(); err != nil {
    fmt.Println("Execute failed when Wait:" + err.Error())
    return
  }
 
  fmt.Println("Execute finished:")
  fmt.Println(string(out_bytes))
}
```

```
# go run pipe.go 
Execute finished:
igo test text for grep
igo test text for grep 02
```

```
导入本地包:
需求:导入本地cmd包,使main函数可以调用cmd包里的各种实现函数
funclist]# tree
.
├── cmd
│   ├── Cmd_ls.go
│   └── Cmd_tree.go
├── funclist
├── go.mod
├── main.go

1. # vim go.mod
module funclist

go 1.17

2. head cmd/Cmd_tree.go
package cmd

3. # vim main.go 
package main

import (
    "fmt"
    "os/exec"
    "funclist/cmd"
)
func main(){
cmd.Cmd_tree()
}
```

```
shell管道命令的实现:

funclist]# cat cmd/Cmd_tree.go 
package cmd

import (
"os"
"os/exec"
)

func Cmd_tree(){
        c1:= exec.Command("curl","-s", "<脚本URL>")
        cmd:= exec.Command("/bin/bash")
        cmd.Stdin, _ = c1.StdoutPipe()
        cmd.Stdout = os.Stdout
        _ = cmd.Start()
        _ = c1.Run()
        _ = cmd.Wait()

}
// curl -s <脚本URL> | bash
```
