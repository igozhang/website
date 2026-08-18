# mac_tips

macbook 速查 小操作

## mac指定域名走专用dns

以 igozhang.com 为例

### 先检查dns服务器是否可用

```
nslookup www.qq.com 10.1.1.1
OR
dig @10.1.1.1 www.baidu.com

检查udp,tcp端口是否可用，默认udp
nc -vzuv 10.1.1.1 53 
nc -vzt 10.1.1.1 53
```

### 特定域名走指定dns

1. 创建 resolver 目录并添加一条规则（一般不存在）

```
sudo mkdir -p /etc/resolver
sudo tee /etc/resolver/igozhang.com <<EOF
nameserver 10.1.1.1
EOF

其中一个三级域名单独走公网dns
sudo tee /etc/resolver/ftp.igozhang.com <<EOF
nameserver 8.8.8.8
EOF
```

2. 刷新 DNS（必须做）

```
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

3. 验证

```
scutil --dns | grep -I3 igozhang.com
输出大概这样：
resolver #8
  domain   : igozhang.com
  nameserver[0] : 10.1.1.1 53
  flags    : Request A records, Request AAAA records
```
