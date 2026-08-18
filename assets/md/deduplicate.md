# deduplicate

去重shell脚本

```
#!/bin/bash

# 检查参数是否提供
if [ $# -ne 1 ]; then
    echo "用法: $0 <文件名>"
    exit 1
fi

input_file="$1"

# 判断文件是否存在
if [ ! -f "$input_file" ]; then
    echo "错误：文件 '$input_file' 不存在"
    exit 1
fi

# 使用 awk 去除重复行，保留第一次出现的行
awk '!seen[$0]++' "$input_file"
```
