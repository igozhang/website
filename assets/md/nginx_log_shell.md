# nginx_log shell日志生成

```
#!/bin/bash
mkdir -p /data/igo
while true; do
  # 获取当前时间
  timestamp=$(date +"%Y-%m-%d %H:%M:%S")

  # 生成随机的IP地址
  ip=$(shuf -i 1-255 -n 1).$(shuf -i 0-255 -n 1).$(shuf -i 0-255 -n 1).$(shuf -i 1-255 -n 1)

  # 生成随机的URL路径
  path="/$(shuf -i 1-100 -n 1)/$(shuf -i 1-1000 -n 1)/$(shuf -i 1-10000 -n 1)"

  # 生成随机的HTTP状态码
  status=$(shuf -i 200-300 -n 1)

  # 生成随机的User-Agent字符串
  user_agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36"

  # 构建日志行
  log_line="$ip - - [$timestamp] \"GET $path HTTP/1.1\" $status 0 \"-\" \"$user_agent\""

  # 写入日志文件
  echo $log_line >> /data/nginx_acc.log

  # 等待1秒钟
  sleep 1
done
```

```
bash /opt/nglog.sh &
```
