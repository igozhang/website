# 网络配置命令

```
临时配置IP：重启network后失效
ifconfig eth1 10.0.0.1 netmask 255.255.255.0 up
ip addr add 10.0.0.1/24 dev eth1
ip addr del ip/netmask dev eth1
子接口
ip addr add 3.3.3.3/16 dev eth0 label eth0:1
ip addr del ip/netmask dev eth0 [label eth0:1] 
ip link set eth1 up 

查看网络配置
ip a
ip a s eth0

查看路由
netstat -rn
route -n
查看ipv6路由
route -6n 

路由
添加默认路由
route add default gw 10.2.111.111
ip route add 10.2.111.112 via 10.1.111.112
route add -host 目标主机 gw 网关 dev 接口
route add -host 10.1.111.111 gw 10.2.111.111 dev eth0
route add -net 目标网络 netmask 子网掩码 gw 网关 dev 接口
route add -net 10.1.0.0 netmask 255.255.0.0 gw 10.2.111.111 dev eth0
route del default
route del -net 192.168.2.0/24
ip route del default [via 网关]

ip -6 route add <ipv6network>/<prefixlength> via <ipv6address>
ip -6 route add default via 2001:0db8:0:f101::1
route -A inet6 add <ipv6network>/<prefixlength> gw
route -A inet6 add default gw 2001:0db8:0:f101::1

重启网络
service network restart
systemctl restart network
systemctl restart networking

nmcli
nmcli c add type ethernet con-name ens786f0 ifname ens786f0 ipv6.addresses fd00::7:200/112 ipv6.gateway fd00::7:1 ipv6.method manual
nmcli c add type ethernet con-name ens817f0 ifname ens817f0 ipv4.addr 10.4.0.10/24 ipv4.gateway 10.4.0.1 ipv4.method manual

lsmod | grep 8021q  ||  modprobe 8021q
ip link add link enp1s0 name ens8.18 type vlan id 18
ip addr add 172.17.232.184/24 dev ens8.18
ip route add default via 172.17.232.254
ip link delete eth0.8

使用命令配置网络都是临时的，配置永久ip需要编写对应配置文件，具体参考另一篇文章 《网络配置文件》
```
