# shell查字典

查字典：

1. 字典文件cidian.igo,其中第一列是工号，第二列是名字，以空格分隔
2. 输入一行以逗号为分隔符的工号，输出一行以逗号为分隔符的姓名

```
#!/bin/bash

# 检查是否提供了输入数据
if [ -z "$1" ]; then
    echo "请提供输入数据作为参数"
    exit 1
fi

# 定义字典文件路径
DICT_FILE="cidian.igo"

# 检查字典文件是否存在
if [ ! -f "$DICT_FILE" ]; then
    echo "字典文件 $DICT_FILE 不存在"
    exit 1
fi

# 处理输入数据：去掉空格，将中文逗号转换为英文逗号
INPUT=$(echo "$1" | tr -d ' ' | sed 's/，/,/g')

# 读取字典文件到关联数组中
declare -A DICT
while IFS=' ' read -r KEY VALUE; do
    DICT["$KEY"]="$VALUE"
done < "$DICT_FILE"

# 输出处理后的结果
IFS=',' read -ra ADDR <<< "$INPUT"
for i in "${ADDR[@]}"; do
    # 如果在字典中有对应的键，则替换为值，否则保持原样
    if [[ ${DICT[$i]+_} ]]; then
        echo -n "${DICT[$i]},"
    else
        echo -n "$i,"
    fi
done
echo # 打印换行符结束输出
```
