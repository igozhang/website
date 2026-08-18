# ubuntu20.04mac地址绑定

```
问题现象：
一批主机的桥接口eth0,eth1mac地址发生互换，导致无法正常通信，解决办法，系统绑定mac地址

cp /etc/default/grub{,.bak}
sed -i '/GRUB_CMDLINE_LINUX_DEFAULT/c GRUB_CMDLINE_LINUX_DEFAULT="modprobe.blacklist=nouveau intel_iommu=on"' /etc/default/grub
grep ^GRUB_CMDLINE_LINUX_DEFAULT /etc/default/grub
update-grub2

wget 172.27.240.176:33133/macbind.sh -O /tmp/macbind.sh
bash /tmp/macbind.sh
ip a |grep -A2 eth0
cat /etc/systemd/network/10-eth0.link

cat macbind.sh
#!/bin/bash
eth0_mac=$(cat /sys/class/net/eth0/address)
eth1_mac=$(cat /sys/class/net/eth1/address)
file1="/etc/systemd/network/10-eth0.link"
file2="/etc/systemd/network/10-eth1.link"

if [ ! -f "${file1}" ];then
cat > ${file1} << EOF
[Match]
MACAddress=${eth0_mac}

[Link]
Name=eth0
EOF
else
    echo "${file1} is already exists"
fi

if [ ! -f "${file2}" ];then
cat > ${file2} << EOF
[Match]
MACAddress=${eth1_mac}

[Link]
Name=eth1
EOF
else
    echo "${file2} is already exists"
fi
```
