# 网络配置文件

## ubuntu22.04备份现有配置，设置bond并生效
```shell
sudo cp /etc/netplan/50-cloud-init.yaml /etc/netplan/50-cloud-init.yaml.bak && sudo tee /etc/netplan/50-cloud-init.yaml > /dev/null << 'EOF'
network:
  ethernets:
    ens9f1np1:
      dhcp4: no
    ens10f1np1:
      dhcp4: no
  bonds:
    bond0:
      interfaces: [ens9f1np1, ens10f1np1]
      dhcp4: no
      addresses: [10.145.241.152/24]
      routes:
        - to: default
          via: 10.145.241.254
      nameservers:
        addresses: [10.145.241.231]
      parameters:
        mode: active-backup
        mii-monitor-interval: 100
  version: 2
  renderer: networkd
EOF
sudo netplan apply
```
回退
```shell
cp /etc/netplan/50-cloud-init.yaml.bak /etc/netplan/50-cloud-init.yaml && sudo ip link del bond0 2>/dev/null && sudo netplan apply
```



ubuntu网络配置，network\_config

```
ubuntu18及以前用/etc/network/interfaces
ubuntu20及以后用/etc/netplan/00-installer-config.yaml

u18 /etc/network/interfaces
auto ens3
iface ens3 inet static
  address 172.21.240.105
  netmask 255.255.255.0
  gateway 172.21.240.254
  dns-nameserver 223.5.5.5
netplan apply

u20 /etc/netplan/00-installer-config.yaml
# This is the network config written by 'subiquity'
network:
  version: 2
  ethernets:
    ens32:
     dhcp4: false
     addresses: [10.10.2.223/24]
     gateway4: 10.10.2.2
     nameservers:
       addresses: [114.114.114.114,223.5.5.5]

u18子接口：
auto ens3:1
iface ens3:1 inet static
    address 172.25.240.212
    netmask 255.255.255.0
auto ens3:2
iface ens3:2 inet static
    address 172.25.240.212
    netmask 255.255.255.0

trunk，vlan：
auto enp4s0.240
iface enp4s0.240 inet manual
vlan-raw-device enp4s0
auto br0
iface br0 inet static
        address 172.21.240.102
        network 172.21.240.0
        netmask 255.255.255.0
        gateway 172.21.240.254
        dns-nameservers 114.114.114.114
        bridge_ports enp4s0.240
        bridge_stp off
        bridge_fd 0
        bridge_maxwait 0

bond：
auto bond0
  iface bond0 inet manual
  #bond-mode active-backup
  bond-mode 1
  bond-miimon 100
  bond-slaves enp4s0 enp130s0

Vlan配置：
interfaces:
cat /etc/network/interfaces
auto lo
iface lo inet loopback

auto eth0.101
iface eth0.101 inet manual
mtu 9000
vlan-raw-device eth0

auto eth0.201
iface eth0.201 inet manual
vlan-raw-device eth0

auto ibr0
iface ibr0 inet dhcp
bridge_ports eth0.101
bridge_stp off
bridge_fd 0
bridge_maxwait 0

auto obr0
iface obr0 inet dhcp
bridge_ports eth0.201
bridge_stp off
bridge_fd 0
bridge_maxwait 0

完整示例ubuntu18
# cat /etc/netplan/01-netcfg.yaml
network:
  version: 2
  renderer: networkd
  ethernets:
   enp132s0f0:
    dhcp4: no
    dhcp6: no
    optional: true
   eno1:
    dhcp4: no
    dhcp6: no
    optional: true
  vlans:
    vlan200:
     accept-ra: no
     id: 200
     link: bond0
    vlan201:
     accept-ra: no
     id: 201
     link: bond0
    vlan202:
     accept-ra: no
     id: 202
     link: bond0
    vlan203:
     accept-ra: no
     id: 203
     link: bond0
    vlan204:
     accept-ra: no
     id: 204
     link: bond0
  bridges:
   br0:
    interfaces: [bond0]
    dhcp4: no
    addresses: [10.33.240.211/24]
    gateway4: 10.33.240.254
    nameservers:
     addresses: [223.6.6.6,223.5.5.5]
    parameters:
     stp: false
     forward-delay: 0
   br200:
    interfaces: [vlan200]
    dhcp4: no
    parameters:
     stp: false
     forward-delay: 0
   br201:
    interfaces: [vlan201]
    dhcp4: no
    parameters:
     stp: false
     forward-delay: 0

  bonds:
   bond0:
    dhcp4: no
    dhcp6: no
    mtu: 9000
    interfaces:
    - eno1
    - enp132s0f0
    parameters:
       mode: active-backup
       mii-monitor-interval: 100

netplan apply
```

centos网络配置

```
vim /etc/sysconfig/network-scripts/ifcfg-eth0
TYPE=Ethernet
PROXY_METHOD=none
BROWSER_ONLY=no
DEFROUTE=yes
IPV4_FAILURE_FATAL=no
#IPV6INIT=yes
#IPV6_AUTOCONF=yes
#IPV6_DEFROUTE=yes
#IPV6_FAILURE_FATAL=no
#IPV6_ADDR_GEN_MODE=stable-privacy
NAME=eth0
#UUID=c180587d-598d-4299-97c3-149d08125496
DEVICE=eth0
ONBOOT=yes
BOOTPROTO=static
IPADDR=172.21.240.163
NETMASK=255.255.255.0
GATEWAY=172.21.240.254
DNS1=223.5.5.5
子接口:
vim /etc/sysconfig/network-scripts/ifcfg-eth0:1
TYPE=Ethernet
PROXY_METHOD=none
BROWSER_ONLY=no
BOOTPROTO=dhcp
DEFROUTE=yes
IPV4_FAILURE_FATAL=no
#IPV6INIT=yes
NAME=eth0:1
DEVICE=eth0:1
ONBOOT=yes
BOOTPROTO=static
IPADDR=172.21.240.124
NETMASK=255.255.255.0

Centos_vlan
vlan：
母卡：
DEVICE=ethX
TYPE=Ethernet
BOOTPROTO=none
ONBOOT=yes

vlan卡：
DEVICE=ethX.192
BOOTPROTO=none
ONBOOT=yes
IPADDR=192.168.1.1
PREFIX=24
NETWORK=192.168.1.0
VLAN=yes

systemctl restart network

cent_team

nmcli connection add type team con-name team1 ifname team1 config '{"runner":{"name":"activebackup"}}'
nmcli connection add type team-slave con-name team1-port1 ifname ens161 master team1
nmcli connection add type team-slave con-name team1-port2 ifname ens162 master team1
nmcli connection modify team1 ipv4.method manual ipv4.addresses 192.168.3.201/24 ipv4.gateway 192.168.3.1 ipv4.dns 223.5.5.5
关闭两slave卡的开机启动onboot=no,否则不会自动关联到team
nmcli connection modify ens161 autoconnect no
nmcli connection modify ens162 autoconnect no

systemctl restart network
teamdctl team1 state

cent_bond
nmcli connection add type bond ifname bond0 mode 1
nmcli connection add type bond-slave ifname eno1 master bond0
nmcli connection add type bond-slave ifname eno2 master bond0
systemctl restart network
```
