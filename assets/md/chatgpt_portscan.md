# 端口探测 by Chatgpt

```
Port scanning
帮忙用shell写一个根据ip列表进行端口探测并返回成功或者失败的脚本
```

```
#!/bin/bash
IP列表

IP_LIST=("192.168.1.1" "192.168.1.2" "192.168.1.3")
端口列表

PORT_LIST=("80" "443" "22")

for ip in "${IP_LIST[@]}"; do
  for port in "${PORT_LIST[@]}"; do
    nc -z -w 2 "$ip" "$port"
    if [ "$?" -eq 0 ]; then
      echo "Port $port on $ip is open."
      break
    else
      echo "Port $port on $ip is closed."
    fi
  done
done
```
