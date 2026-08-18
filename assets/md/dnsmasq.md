# dnsmasq服务

## env

centos7.9  
dnsmasq方便，简洁，明了，可以同时做dhcp服务器使用

## 虚机安装配置

yum -y install dnsmasq-2.76

```
tee >/etc/dnsmasq.d/igo.conf <<EOF
address=/igozhang.cn/192.168.0.33
address=/ns1.igozhang.cn/192.168.0.33
address=/ns2.igozhang.cn/192.168.0.33
EOF
```

默认开始上级代理

重启生效  
systemctl restart dnsmasq

## docker\_dnsmasq安装配置

容器方式需要关闭selinux  
setenfoce 0  
sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config

yum -y install docker-1.13.1  
docker pull jpillora/dnsmasq

mkdir -p /data/dns/

```
tee >/data/dns/dnsmasq.conf <<EOF
#dns解析日志
log-queries
#域名与IP映射
address=/igozhang.cn/192.168.0.33
address=/ns1.igozhang.cn/192.168.0.33
address=/ns2.igozhang.cn/192.168.0.33
EOF
```

启动  
docker run -d\  
--name dnsmasq \  
--restart always \  
-p 53:53/udp \  
-p 8080:8080 \  
-v /data/dns/dnsmasq.conf:/etc/dnsmasq.conf \  
--log-driver json-file \  
--log-opt max-size=15m \  
--log-opt max-file=5 \  
-e "HTTP\_USER=admin" \  
-e "HTTP\_PASS=123456" \  
jpillora/dnsmasq

docker restart dnsmasq  
docker ps  
管理 192.168.0.33:8080
