# ipxe_CtrlCV备查手册

```
ipxe的一点理解和实践
iPXE -- Open Source Boot Firmware -- http://ipxe.org
PXE  Preboot eXecution Environment
iPXE是一个基于GPL的开源引导固件，完整的PXE实现并且扩展增强了功能,比如支持http,iscsi_SAN,FC_SAN

https://ipxe.org
https://ipxe.org/appnote/proxydhcp
https://wiki.archlinux.org/title/dnsmasq#PXE_server
```

```
1.iPXE安装

1.1编译安装
https://ipxe.org/appnote/buildtargets

git clone git://git.ipxe.org/ipxe.git
make [platform]/[driver].[extension]
Exp：
make bin-x86_64-efi/ipxe.efi
make bin/undionly.kpxe
支持的平台有：
bin (alias for bin-i386-pcbios)
bin-i386-pcbios
bin-i386-efi
bin-i386-linux
bin-x86_64-efi
bin-x86_64-linux
bin-x86_64-pcbios
bin-arm32-efi
bin-arm64-efi

支持embed脚本
tee >embed.ipxe <<EOF
#!ipxe
dhcp
chain tftp://${next-server}/main.ipxe || shell
EOF
make bin/undionly.kpxe EMBED=embed.ipxe
cp bin/undionly.kpxe /opt/tftp

1.2烧录安装
https://ipxe.org/download
烧录U盘，CD，网卡ROM

a.U盘，CD
$ make bin/ipxe.iso
$ make bin/ipxe.usb
$ dd if=bin/ipxe.usb of=/dev/sdX

b.网卡ROM
1.找到网卡PCI vendor和DeviceID
# lspci -nn
00:00.0 Host bridge [0600]: Intel Corporation 440BX/ZX/DX - 82443BX/ZX/DX Host bridge [8086:7190] (rev 03)
00:01.0 PCI bridge [0604]: Intel Corporation 440BX/ZX/DX - 82443BX/ZX/DX AGP bridge [8086:7191] (rev 03)
...
00:0d.0 Ethernet controller [0200]: 3Com Corporation 3c900B-TPO Etherlink XL [Cyclone] [10b7:9004] (rev 04)
PCI vendor 10b7
DeviceID  9004

2.make
make bin/10b79004.rom

3.烧网卡
https://ipxe.org/howto/romburning/intel
找到网卡厂商烧录软件（烧录一般用来改MAC，升级驱动，刷功能等场景）
烧录进去，一般需要转格式，记得备份原固件文件，容易搞挂
复杂功能，坑多多，不同品牌大不相同，不赘述
```

```
2.iPXE配置
https://gist.github.com/rikka0w0/50895b82cbec8a3a1e8c7707479824c1
https://ipxe.org/cfg 配置解读
```

```
3.iPXE_Shell
Ctrl+B 进入，也可以写http,tftp脚本嵌入
https://ipxe.org/cmdline

ifstat 列出可用网卡
ifopen 打开网卡
ifclose关闭网卡
dhcp   发送dhcp请求
route  显示路由，一般用于dhcp后检查
chain  链式引导
sanboot iscsi:192.168.3.221:::1:iqn.2021-11.com.igo:id.2004
sanhook iscsi:192.168.3.221:::1:iqn.2021-11.com.igo:id.2004
sanhoot 
示例及说明：
iPXE> ifopen net0
iPXE> set net0/ip 192.168.3.121
iPXE> set net0/netmask 255.255.255.0
iPXE> set net0/gateway 192.168.3.2
iPXE> ifclose net1
iPXE> route
  net0: 192.168.3.121/255.255.255.0 gw 192.168.3.2
  net1: 192.168.3.153/255.255.255.0 gw 192.168.3.2 (inaccessible)
iPXE> chain http://boot.ipxe.org/demo/boot.php
tips
使用嵌入式脚本不会有Ctrl+B提示，可以手动加入
prompt --key 0x02 --timeout 2000 Press Ctrl-B for the iPXE command line... && shell ||
```

```
tips
dhcp-match=set:ipxe,175
if option 175 is set in the DHCP request. (175 returns text string 'Etherboot', after original project name)

不同机器不同引导
dnsmasq:
dhcp-boot=net:ipxe,http://igozhang.cn:8080/ipxe.php?mac=${mac}&uuid=${uuid}
或者
ipxe.boot
chain http://igozhang.cn:8080/ipxe/boot?mac=${net0/mac}&ip=${net0/ip}
```
