# proxy

ENV:  Centos76

```
squid

安装 : yum -y install squid
配置: vim /etc/squid/squid.conf
（默认无需配置，监听所有主机访问3128端口）
启动：systemctl start squid
客户机验证：
# export http_proxy=10.8.8.128:3128
# curl http://www.baidu.com
OK了
```

```
clash

wget https://github.com/Dreamacro/clash/releases/download/v1.6.0/clash-linux-amd64-v1.6.0.gz
file clash-linux-amd64-v1.6.0
mv clash-linux-amd64-v1.6.0  /usr/local/bin/clash
chmod +x /usr/local/bin/clash
clash --help
env |grep -i clash

supervisor：
# cat /etc/supervisord.d/clash.ini
[program:clash]
command=clash -f /opt/clash/blinkload.yml
autostart=true
autorestart=true

docker内网代理
```

```
Bash
# http代理
export http_proxy="http://10.80.238.40:3128"
export https_proxy="http://10.80.238.40:3128"

export http_proxy="http://10.201.81.69:3128"
export https_proxy="http://10.201.81.69:3128"

# socks5代理
export http_proxy="socks5://127.0.0.1:7890"
export https_proxy="socks5://127.0.0.1:7890"
unset http_proxy
unset https_proxy

CMD
# http代理
set http_proxy=http://127.0.0.1:7890
set https_proxy=http://127.0.0.1:7890
# socks5代理
set http_proxy=socks5://127.0.0.1:7890
set https_proxy=socks5://127.0.0.1:7890

set http_proxy=
set https_proxy=

go
export GOPROXY=https://goproxy.cn

git
# http代理
git config --global http.proxy http://127.0.0.1:7890
git config --global https.proxy http://127.0.0.1:7890
git config --global http.proxy "socks5://127.0.0.1:7890"
git config --global https.proxy "socks5://127.0.0.1:7890"
git config --global --unset http.proxy
git config --global --unset https.proxy

git clone  https://github.com/protocolbuffers/protobuf
以下:
git clone  https://gitclone.com/github.com/protocolbuffers/protobuf
OR
git config --global url."https://gitclone.com/".insteadOf https://
git config --global url."https://gitclone.com/github.com".insteadOf https://github.com

npm_proxy
npm config set proxy http://127.0.0.1:7890
npm config set https-proxy http://127.0.0.1:7890
npm config delete proxy
npm config delete https-proxy
```
