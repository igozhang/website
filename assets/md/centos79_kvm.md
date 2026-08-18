# centos79_kvm,CENTOS部署KVM

## env

CentOS Linux release 7.9.2009 (Core)  
virsh 4.5.0  
virt-viewer-5.0-18  
virt-install-1.5.0-7.el7.noarch  
bridge-utils-1.5-9.el7.x86\_64  
libguestfs-tools-1.40.2-10.el7.noarch

## 安装配置

```
parted -s  /dev/sdb mklabel gpt mkpart primary xfs 0% 100%
mkfs.xfs /dev/sdb1
echo "/dev/sdb1  /vmdisk  xfs  defaults  0  0" >> /etc/fstab
mount -a

yum install -y qemu-kvm virt-install  virt-viewer  bridge-utils virt-manager cpu-checker  lrzsz ipmitool  iperf iftop iotop curl vlan jq  libguestfs-tools

虚拟化主机CPU需要打开（虚拟化IntelVT-x/EPT或AMD-V/RVI(V)）
nmcli con mod ens33 ipv4.method manual ipv4.addr 192.168.4.33/24 ipv4.gateway 192.168.4.2 ipv4.dns 192.168.4.2 autoconnect yes
systemctl restart network
hostnamectl set-hostname igo-kvm
systemctl stop firewalld
systemctl disable firewalld
setenforce 0
sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config 
timedatectl set-timezone Asia/Shanghai
rm -rf /etc/yum.repos.d/*.repo
curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo
curl -o /etc/yum.repos.d/epel-7.repo https://mirrors.aliyun.com/repo/epel-7.repo
mkdir -p /data/kvm

systemctl start libvirtd

网桥配置nmcli
nmcli con add type bridge ifname br0 con-name br0 stp no ipv4.dns "8.8.8.8 8.8.4.4"
nmcli con add type bridge-slave ifname ens33 master br0
nmcli con up br0
nmcli con mod eth0 ipv4.method manual ipv4.addr 192.168.3.222/24 ipv4.gateway 192.168.3.1 ipv4.dns 223.5.5.5 autoconnect yes
systemctl restart network

网桥配置iplink
ip link add name br0 type bridge
ip link set ens33 master br0
ip link set br0 up

ip link set ens33 nomaster 将ens33踢出网桥
ip link delete br0 删除网桥
ip link set br0 stp_state 0 关闭stp

网桥配置brctl包名为bridge-utils
brctl addbr br0
brctl addif br0 ens33
brctl show

brctl delif br0 ens33 将ens33踢出网桥
brctl delbr br0  删除网桥
brctl br0 off  关闭stp

virt-install \
--name=igo_cent7 \
--ram=1024 \
--cpu=host \
--vcpus=2 \
--os-type=Linux \
--os-variant=centos7.0 \
--disk /vmdisk/cent79.qcow2,size=10 \
--cdrom /data/iso/CentOS-7.9-x86_64-Minimal-2009.iso \
--network=bridge=virbr0,model=virtio \
--graphics vnc,port=13344,listen=0.0.0.0

或者直接使用默认nat网络192.168.122.0/24(CENTOS7目前无法使用)
--network network=default
```

## 调试

虚拟机里面打开console以便调试网络  
systemctl start serial-getty@ttyS0  
systemctl enable serial-getty@ttyS0

宿主机进入虚拟机console：  
virsh console 5  
或者  
virt-viewer -c qemu:///system 5 图形化  
virt-viewer vm\_name
