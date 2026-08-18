# login_check

```
用于检查 CentOS 7 系统中今日通过 SSH 成功登录的用户和来源 IP 地址的 Shell 脚本

筛选今日成功登录记录

从 /var/log/secure 日志中提取今天所有 Accepted（成功登录）的 SSH 记录。
准确解析用户和IP

使用正则表达式匹配日志中的 for 用户名 from IP 模式，避免因字段位置变化导致解析错误。
结构化输出信息
列出每次成功登录的时间、用户名和客户端IP
汇总唯一的访问IP地址
汇总唯一登录的用户名
显示失败登录的IP及尝试次数（带统计）
增强容错能力

对格式异常的日志行会提示“解析失败”，便于排查问题。
```

```
#!/bin/bash
# ssh-access-today.sh - 改进版

LOG_FILE="/var/log/secure"
TODAY=$(date +%b" "%d)  # 注意：日期格式如 "Aug 14"

echo "=== 今天SSH访问记录 ==="
echo "时间: $(date)"
echo "日志文件: $LOG_FILE"
echo "过滤日期: $TODAY"

# 提取成功登录的记录（更精确的模式匹配）
echo -e "\n--- 成功登录 ---"
grep "Accepted" "$LOG_FILE" | grep "$TODAY" | while read line; do
    # 使用正则提取时间、用户名、IP
    if [[ $line =~ ([A-Za-z]{3}[[:space:]]+[0-9]{1,2}[[:space:]]+[0-9:]{8})?.*Accepted.*for[[:space:]]+([^[:space:]]+).*from[[:space:]]+([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+) ]]; then
        timestamp="${BASH_REMATCH[1]:-未知时间}"
        user="${BASH_REMATCH[2]}"
        ip="${BASH_REMATCH[3]}"
        echo "$timestamp - 用户: $user, IP: $ip"
    else
        # 如果正则不匹配，打印原始行用于调试
        echo "解析失败: $line"
    fi
done

echo -e "\n--- 唯一访问IP ---"
grep "Accepted" "$LOG_FILE" | grep "$TODAY" | \
    grep -oE 'from [0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | \
    cut -d' ' -f2 | sort | uniq

echo -e "\n--- 唯一登录用户 ---"
grep "Accepted" "$LOG_FILE" | grep "$TODAY" | \
    grep -oE 'for [^[:space:]]+ from' | \
    cut -d' ' -f2 | sort | uniq

echo -e "\n--- 失败尝试 ---"
grep "Failed" "$LOG_FILE" | grep "$TODAY" | \
    grep -oE 'from [0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | \
    cut -d' ' -f2 | sort | uniq -c | sort -nr
```
