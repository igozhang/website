# Centos_frr_HA

```
env
CentOS Linux release 7.6.1810
frr-7.4

集群中多个服务器通过OSPF协议向网络设备宣告相同的IP地址(VIP)，网络设备通过ECMP路由将流量分散到这些服务器中。当其中某些服务器异常宕机或者维护下线时，网络设备会通过OSPF或者BGP协议的路由收敛机制检测到服务器下线，迅速将流量切走, 实现服务高可用.

vtysh
# show running-config
Building configuration...

Current configuration:
!
frr version 7.4
frr defaults traditional
hostname igo-frr-01
no ip forwarding
no ipv6 forwarding
log syslog informational
!
ip route 0.0.0.0/0 192.168.249.254
!
interface eth0
 ip address 192.168.249.201/24
 ip ospf bfd
!
interface lo
 ip address 192.168.250.219/32
 ip address 192.168.250.232/32
!
router ospf
 ospf router-id 192.168.249.201
 log-adjacency-changes detail
 network 192.168.249.0/24 area 0.0.0.0
 network 192.168.250.0/24 area 0.0.0.0
!
line vty
!
bfd
 peer 192.168.249.201 interface eth0
  detect-multiplier 5
  transmit-interval 50
  receive-interval 50
 !
!
end
```
