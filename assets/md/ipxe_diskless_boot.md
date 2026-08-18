# ipxe无盘服务器搭建

```
env
CentOS Linux release 7.9.2009
VMware® Workstation 16 Pro 16.2.0
scsi-target-utils-1.0.55-4.el7
ipxe 1.21.1+ g90625

https://ipxe.org

概要：
利用ipxe搭建一套无盘系统做内网引导服务器
```

```
1.
iscsi_server
yum -y install scsi-target-utils.x86_64

tee >>/etc/tgt/targets.conf <<EOF
iSNSServerIP 192.168.3.221
iSNSAccessControl No
EOF
systemctl start tgtd && systemctl enable tgtd

2.
ipxe_Server
yum -y install gcc make git wget vim mlocate tree htop
mkdir -p /opt/tftpboot/menu 
cd /opt/
git clone git://git.ipxe.org/ipxe.git
cd ipxe/src
make bin-x86_64-efi/ipxe.efi
make bin/undionly.kpxe
or
wget http://boot.ipxe.org/undionly.kpxe
cp /opt/ipxe/src/bin-x86_64-efi/ipxe.efi /opt/tftpboot
cp /opt/ipxe/src/bin/undionly.kpxe /opt/tftpboot

3.
dhcp_Server
yum -y install dnsmasq
tee >/etc/dnsmasq.d/igoserv.conf <<EOF
# 禁用dns,默认53端口
port=0
#server=223.5.5.5
#address=/git.igozhang.cn/xx.xx.xx.xx
listen-address=192.168.3.221
interface=ens33
# tftp配置
enable-tftp
tftp-root=/opt/tftpboot
# dhcp配置
dhcp-range=192.168.3.151,192.168.3.160,255.255.255.0,24h
dhcp-option=3,192.168.3.2 #网关
dhcp-match=set:ipxe,175
dhcp-vendorclass=BIOS,PXEClient:Arch:00000
dhcp-boot=tag:!ipxe,tag:BIOS,undionly.kpxe # for BIOS
dhcp-boot=tag:!ipxe,tag:!BIOS,ipxe.efi # for UEFI
dhcp-boot=tag:ipxe,menu/boot.ipxe
#dhcp-host=00:e0:4c:43:52:b9,10.86.39.105
EOF
systemctl start dnsmasq && systemctl enable dnsmasq

4.
创建测试镜像
yum -y install qemu-img

ID=2004
qemu-img convert -f vmdk -O raw /data/u20.04.vmdk /data/id-$ID.raw

losetup /dev/loop$ID /data/id-$ID.raw
tgtadm --lld iscsi --op new --mode target --tid $ID -T iqn.2021-11.com.igo:id.$ID
tgtadm --lld iscsi --op new --mode logicalunit --tid $ID --lun 1 -b /dev/loop$ID
tgtadm --lld iscsi --op bind --mode target --tid $ID -I ALL

[root@igo-ipxe_Server data]# tgt-admin -s  |grep Target
Target 2004: iqn.2021-11.com.igo:id.2004

5.
boot_menu
boot_menu
tee >/opt/tftpboot/menu/boot.ipxe <<EOF
#!ipxe
:start
menu igoZhang Boot Options
# 菜单的title
item --key d default (D)Default
item --key s shell (S)iPXE shell
item --key r reboot (R)reboot
# 菜单显示的条目
choose --default default --timeout 3000 option && goto \${option}
:default
echo Start Default.....
dhcp
echo dhcp-01 192.168.3.221
echo IP address: \${net0/ip}
echo MAC address: \${net0/mac}
set keep-san 1
sanboot --no-describe iscsi:192.168.3.221:::1:iqn.2021-11.com.igo:id.2004
boot
:shell
shell
:reboot
reboot
EOF
```

6.Done

ubuntu20.04

```
tips

(Cent8没有scsi-target-utils包)

#查看
tgtadm --lld iscsi --op show --mode target |grep Target
tgtadm -C 0 --op show --mode target 
tgt-admin -s  |grep Target
losetup -l

#删除
ID=100108
tgtadm   --op delete --mode target --tid=$ID
losetup -d /dev/loop$ID
rm -rf /data/id-$ID.raw

保存配置：
tgt-admin --dump > /etc/tgt/conf.d/ipxe.conf
# cat /etc/tgt/conf.d/ipxe.conf
default-driver iscsi #保存的配置文件里这行需删除掉，否则启动报错

Vmware_workstation启用uefi
找到虚拟vmx文件
第一行后添加
firmware="efi"
效果如下
.encoding = "GBK"
firmware="efi"
config.version = "8"
virtualHW.version = "19"
```
