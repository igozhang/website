# 黄金一分钟

> 快速收集Linux系统当前状态信息并记录到output\_file

```
#!/bin/bash

# 输出到文件
output_file="igo.out$(date +%Y%m%d%H%M%S)"

# 清空输出文件
> "$output_file"

# 写入基本信息
echo "System Information Report" > "$output_file"
echo "-------------------------" >> "$output_file"
echo "Report Generated: $(date)" >> "$output_file"
echo "Hostname: $(hostname)" >> "$output_file"
echo "User: $(whoami)" >> "$output_file"
echo "" >> "$output_file"

# 写入系统详细信息
echo "Kernel Version: $(uname -r)" >> "$output_file"
echo "Date and Time: $(date)" >> "$output_file"
echo "" >> "$output_file"

echo "Disk Usage:" >> "$output_file"
df -h >> "$output_file"
echo "" >> "$output_file"

echo "Memory Usage:" >> "$output_file"
free -m >> "$output_file"
echo "" >> "$output_file"

echo "Top CPU Processes:" >> "$output_file"
ps -eo pid,ppid,user,%mem,%cpu,command --sort=-%cpu | head -n 6 >> "$output_file"
echo "Top Memory Processes:" >> "$output_file"
ps -eo pid,ppid,user,%mem,%cpu,command --sort=-%mem | head -n 6 >> "$output_file"
echo "" >> "$output_file"

echo "Disk Load (last 5 seconds):" >> "$output_file"
iostat -d -x 5 1 >> "$output_file"
echo "" >> "$output_file"

echo "Network Connections:" >> "$output_file"
ss -tuln >> "$output_file"
echo "" >> "$output_file"

# 写入错误信息
tail -n 100 /var/log/messages | grep -Ei 'error|fail|critical|warning' >> "$output_file"
echo "Recent Errors from journalctl:" >> "$output_file"
journalctl -b -u '*' --since "5 minutes ago" --no-pager --no-full --priority=3..7 >> "$output_file"
echo "" >> "$output_file"

# 显示输出文件的内容
cat "$output_file"
```

shell
