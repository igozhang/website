# centos7部署dns服务器bind

env  
cent79  
bind-9.11.4  
评价：相较于dnsmasq，bind配置太反人类，大型网络环境或者中心节点可以考虑配置，普通办公网或者个人代理用简洁明了的dnsmasq足矣

## 安装配置

yum -y install bind\*  
rpm -lq bind #查看安装文件

### 配置

vim /etc/named.conf  
listen-on port 53 { 192.168.0.33; };  
allow-query { 192.168.0.0/24; };  
默认开启上级转发  
recursion yes;  
改成no不然无法解析外网  
dnssec-enable no;  
dnssec-validation no;

### 添加zone,每个域名为1个zone

tee >>/etc/named.rfc1912.zones <<EOF  
zone "igozhang.com" IN {  
type master;  
file "igozhang.com.zone";  
allow-update { none; };  
};  
EOF

### 新建zone文件

```
tee >/var/named/igozhang.com.zone <<EOF
$TTL 1D
@       IN SOA  @ igozhang.com. (
                                        0       ; serial
                                        1D      ; refresh
                                        1H      ; retry
                                        1W      ; expire
                                        3H )    ; minimum
        NS      @
        A       192.168.0.33
        AAAA    ::1
        PTR     localhost.
@                       IN      NS              ns1.igozhang.com.
@                       IN      NS              ns2.igozhang.com.
ns1                     IN      A               192.168.0.33
ns2                     IN      A               192.168.0.33
EOF
```

# 开防火墙

firewall-cmd --zone=public --add-port=53/tcp --permanent  
firewall-cmd --zone=public --add-port=53/udp --permanent  
firewall-cmd --reload  
systemctl restart firewalld.service  
或

> iptables -I INPUT -p tcp --dport 53 -j ACCEPT #放行TCP协议的53端口  
> iptables -I INPUT -p udp --dport 53 -j ACCEPT #放行UDP协议的53端口  
> /sbin/service iptables save
