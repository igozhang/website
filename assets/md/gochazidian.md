# go查字典

```
CentOS Linux release 7.9.2009 (Core)
go version go1.20.12 linux/amd64
```

1. 字典文件cidian.igo,其中第一列是工号，第二列是名字，以空格分隔
2. 输入一行以逗号为分隔符的工号，输出一行以逗号为分隔符的姓名  
   示例数据：  
   351237 张三  
   351238 李四  
   351239 王五  
   351240 赵六  
   351241 刘七  
   351242 郑八  
   351243 钱九  
   351244 孙十  
   351245 周十一  
   351246 吴十二  
   351247 郑十三  
   351248 王十四  
   351249 赵十五  
   351250 陈十六  
   351251 陆十七  
   351252 江十八  
   351253 冯十九  
   351254 殷二十  
   351255 朱二十一  
   351256 许二十二

```
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// 读取字典文件并存储到map中
	filename := "cidian.igo"
	employeeMap, err := loadEmployeeMap(filename)
	if err != nil {
		fmt.Println("Error loading employee map:", err)
		return
	}

	// 从标准输入读取工号列表
	fmt.Print("请输入工号列表（以逗号分隔）: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input) // 移除字符串首尾的空白字符

	// 解析工号列表
	ids := strings.Split(input, ",")

	// 查找并输出对应的姓名列表
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if name, ok := employeeMap[id]; ok {
			names = append(names, name)
		}
	}

	// 输出结果
	fmt.Println(strings.Join(names, ","))
}

// loadEmployeeMap 加载员工信息到map中
func loadEmployeeMap(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	employeeMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			employeeMap[parts[0]] = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return employeeMap, nil
}
```

```
编译代码  go build -o igochazidian 1125.go
直接运行  go run 1125.go
```
