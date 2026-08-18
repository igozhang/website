# Linux高可用FRR方案

## 概述

Linux系统高可用方案  
OSPF+VRRP  
方案概述：  
在出口服务器，利用vrrp维护一个虚拟IP 3.3.3.3，并通过ospf宣告给上游路由器，实现业务服务的高可用

## 概念

1. OSPF  
   OSPF（Open Shortest Path First）是IETF（Internet Engineering Task Force，互联网工程任务组）组织开发的一个基于链路状态的内部网关协议。目前针对IPv4协议使用的是OSPF Version 2。
2. VRRP  
   Virtual Router Redundancy Protocol (VRRP)  
   虚拟路由冗余协议VRRP（Virtual Router Redundancy Protocol）是一种用于提高网络可靠性的容错协议。通过VRRP，可以在主机的下一跳设备出现故障时，及时将业务切换到备份设备，从而保障网络通信的连续性和可靠性。

## 配置

### H3C配置

1. 接口配置  
   Serv01 （其他设备参照serv01）

```
system-view
int ge0/0
ip add 10.1.1.1 24
int ge0/1
ip add 192.168.1.1 24
[H3C]interface loopback0
[H3C-LoopBack0]ip add
[H3C-LoopBack0]ip address 1.1.1.1 32
```

2. OSPF路由配置  
   Serv01 （其他设备参照serv01）

```
<H3C>system-view
[H3C-ospf-1]ospf 1
[H3C-ospf-1]area 0
[H3C-ospf-1-area-0.0.0.0]network  10.1.1.1 0.0.0.255
[H3C-ospf-1-area-0.0.0.0]network 192.168.1.11 0.0.0.255
[H3C-ospf-1-area-0.0.0.0]network 1.1.1.1 0.0.0.0
[H3C-ospf-1-area-0.0.0.0]quit
[H3C-ospf-1]save
```

3. vrrp配置  
   Serv01 （在Serv01的g0/1接口和Serv02的g0/1接口上跑vrrp）

```
<H3C>system-view
[H3C]interface g0/1
[H3C-GigabitEthernet0/1] vrrp vrid 1 virtual-ip 192.168.1.88
[H3C-GigabitEthernet0/1] vrrp vrid 1 priority 110
[H3C-GigabitEthernet0/1] vrrp vrid 1 preempt-mode
[H3C-GigabitEthernet0/1] quit

三点：
a. vip虚在同网段
b. 默认优先级pri是100
c. 开启抢占模式preempt-mode，宕机重启后会将现有vip抢回来
```

4. 通过ospf将vrrp的vip宣告出去  
   Serv01，Serv02 都需要宣告192.168.1.88路由以实现高可用  
   由于前面已经宣告192.168.1.0/24网段，故此步骤可省略  
   Serv01配置

```
<H3C>system-view
[H3C]ospf 1
[H3C-ospf-1]area 0
[H3C-ospf-1-area-0.0.0.0]network 192.168.1.88 0.0.0.255
[H3C-ospf-1-area-0.0.0.0]quit
```

5. 检查  
   show interface brief  
   show vrrp  
   在非VRRP机器上ping 192.168.1.88

思考题：  
在两台服务器均使用ospf通告了192.168.1.0/24网段时，上游路由器如何判断192.168.1.88具体访问哪台？

ARP

### Linux\_frr配置

1. 安装frr

```
FRRVER="frr-stable"
# add RPM repository on CentOS 7
curl -O https://rpm.frrouting.org/repo/$FRRVER-repo-1-0.el7.noarch.rpm
yum -y install ./$FRRVER*
# install FRR
yum -y install frr frr-pythontools

# 添加lo1接口
ip addr add 1.1.1.1/32 dev lo:1
# 启用内核转发
echo "net.ipv4.conf.all.forwarding=1">>/etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1">>/etc/sysctl.conf
sysctl -p
```

2. 启用ospfd,vrrpd

```
vim /etc/frr/daemons
ospfd=yes
vrrpd=yes

systemctl restart frr
systemctl enable frr
```

3. 配置FRR  
   Serv01

```
[root@igo_11 ~]# vtysh
igo_11# configure terminal 
igo_11(config)# router ospf
igo_11(config-router)# network 192.168.3.0/24 area 0
```

4. 配置VRRP  
   Serv01

```
配置vrrp需要先添加macvlan设备
ip link add vrrp4-2-1 link ens33 addrgenmode random type macvlan mode bridge
ip link set dev vrrp4-2-1 address 00:00:5e:00:01:05
ip addr add 10.0.2.16/24 dev vrrp4-2-1
ip link set dev vrrp4-2-1 up

[root@igo_11 ~]# vtysh
igo_11# configure terminal 
igo_11(config)# interface ens33 
igo_11(config-if)# vrrp 5 version 3
igo_11(config-if)# vrrp 5 priority 100
igo_11(config-if)# vrrp 5 advertisement-interval 1500
igo_11(config-if)# vrrp 5 ip 192.168.3.33
igo_11(config-if)# do write file

Serv02配置除了priory 不一致其他与Serv01相同
其中mac地址'00:00:5e:00:01:05'和vrid'5'由RFC文档指定，不可以改动
```

5. 检查配置  
   vtysh -c 'show running-config'  
   vtysh -c "show ip ospf interface"  
   vtysh -c "show ip ospf neighbor"  
   vtysh -c "show vrrp"

igo\_11# show vrrp

Virtual Router ID 5  
Protocol Version 3  
Autoconfigured No  
Shutdown No  
Interface ens33  
VRRP interface (v4) vrrp4-2-1  
VRRP interface (v6) None  
Primary IP (v4) 192.168.3.11  
Primary IP (v6) ::  
Virtual MAC (v4) 00:00:5e:00:01:05  
Virtual MAC (v6) 00:00:5e:00:02:05  
Status (v4) Master  
Status (v6) Initialize  
Priority 110  
Effective Priority (v4) 110  
Effective Priority (v6) 110  
Preempt Mode Yes  
Accept Mode Yes  
Checksum with IPv4 Pseudoheader Yes  
Advertisement Interval 1500 ms  
Master Advertisement Interval (v4) Rx 1500 ms (stale)  
Master Advertisement Interval (v6) Rx 0 ms (stale)  
Advertisements Tx (v4) 3  
Advertisements Tx (v6) 0  
Advertisements Rx (v4) 0  
Advertisements Rx (v6) 0  
Gratuitous ARP Tx (v4) 1  
Neigh. Adverts Tx (v6) 0  
State transitions (v4) 2  
State transitions (v6) 0  
Skew Time (v4) 850 ms  
Skew Time (v6) 0 ms  
Master Down Interval (v4) 5350 ms  
Master Down Interval (v6) 0 ms  
IPv4 Addresses 1  
……………………………. 192.168.3.33  
IPv6 Addresses
