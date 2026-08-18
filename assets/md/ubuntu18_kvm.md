# ubuntu18_kvm快速部署

env  
Ubuntu 18.04 LTS

```
配置网络
echo 'PermitRootLogin yes' >> /etc/ssh/sshd_config
cp /etc/netplan/01-netcfg.yaml{,.bak`date +%Y%m%d`}
tee >/etc/netplan/01-netcfg.yaml <<EOF
network:
  version: 2
  renderer: networkd
  ethernets:
    ens33:
      dhcp4: no

  bridges:
    br0:
	  interfaces: [ens33]
      dhcp4: no
      addresses: [192.168.3.239/24]
      gateway4: 192.168.3.2
      nameservers:
          addresses: [223.5.5.5,223.6.6.6]
EOF

netplan apply

配置aliyun_ubuntu18源
echo deb http://mirrors.aliyun.com/ubuntu/ bionic main restricted universe multiverse > /etc/apt/sources.list
echo deb-src http://mirrors.aliyun.com/ubuntu/ bionic main restricted universe multiverse >> /etc/apt/sources.list
echo deb http://mirrors.aliyun.com/ubuntu/ bionic-security main restricted universe multiverse >> /etc/apt/sources.list
echo deb-src http://mirrors.aliyun.com/ubuntu/ bionic-security main restricted universe multiverse >> /etc/apt/sources.list
echo deb http://mirrors.aliyun.com/ubuntu/ bionic-updates main restricted universe multiverse >> /etc/apt/sources.list
echo deb-src http://mirrors.aliyun.com/ubuntu/ bionic-updates main restricted universe multiverse >> /etc/apt/sources.list
echo deb http://mirrors.aliyun.com/ubuntu/ bionic-proposed main restricted universe multiverse >> /etc/apt/sources.list
echo deb-src http://mirrors.aliyun.com/ubuntu/ bionic-proposed main restricted universe multiverse >> /etc/apt/sources.list
echo deb http://mirrors.aliyun.com/ubuntu/ bionic-backports main restricted universe multiverse >> /etc/apt/sources.list
echo deb-src http://mirrors.aliyun.com/ubuntu/ bionic-backports main restricted universe multiverse >> /etc/apt/sources.list

apt update

安装qemu-kvm
apt-get install -y qemu-kvm libvirt-bin virtinst bridge-utils virt-manager cpu-checker  lrzsz ipmitool  iperf iftop iotop curl vlan jq  libguestfs-tools

创建kvm目录
mkdir -p /data/kvm && cd /data/kvm

两种方式安装kvm虚拟机
1.
virt-install \
--name=igo_cent7 \
--ram=1024 \
--cpu=host \
--vcpus=2 \
--os-type=Linux \
--os-variant=centos7.0 \
--disk /data/kvm/cent79.qcow2,size=10 \
--cdrom /data/iso/CentOS-7.9-x86_64-Minimal-2009.iso \
--network=bridge=br0,model=virtio \
--graphics vnc,port=3344,listen=0.0.0.0 \
--boot uefi  

或者
--graphics=none --console=pty,target_type=serial \
--extra-args="console=tty0 console=ttyS0"

virt-install \
--name=win10-3090 \
--ram=65536 \
--cpu=host \
--vcpus=24 \
--os-type=Windows \
--os-variant=win10 \
--disk /data/kvm/win10-3090.qcow2,size=100 \
--cdrom /home/iso/cn_windows_10_business_editions_version_20h2_updated_nov_2020_x64_dvd_87114b18.iso \
--network=bridge=br0,model=virtio \
--graphics vnc,port=33344,listen=0.0.0.0 \

2.
拷贝模板qcow2文件和kvm
导出kvm_xml
virsh dumpxml igotest01 > /etc/libvirt/qemu/igotest01.xml
```

```
tips

宿主机进入虚拟机console：
virsh console 5
或者
virt-viewer -c qemu:///system 5  图形化
virt-viewer  vm_name

打开console以便调试网络
systemctl start serial-getty@ttyS0
systemctl enable serial-getty@ttyS0

虚拟机添加vnc设备
kvm_vnc
xml<devices>下面添加
<graphics type='vnc' port='33344' autoport='no' listen='0.0.0.0'>
<listen type='address' address='0.0.0.0'/>
</graphics>

apt -y install libosinfo-bin
osinfo-query os
virt-customize -a rhel-server-7.6.qcow2 --root-password password:StrongRootPassword
```
