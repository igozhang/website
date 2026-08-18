# centos_init,CENTOS系统初始化

> centos7系统初始化

1. 主机名

```
tee >/etc/hostname <<EOF
igo_222
EOF

hostnamectl set-hostname igo_222
```

2. 网络

```
cp /etc/sysconfig/network-scripts/ifcfg-ens32{,.$(date +%Y%m%d)}
tee >/etc/sysconfig/network-scripts/ifcfg-ens32 <<EOF
TYPE=Ethernet
PROXY_METHOD=none
BROWSER_ONLY=no
DEFROUTE=yes
IPV4_FAILURE_FATAL=no
NAME=ens32
DEVICE=ens32
ONBOOT=yes
BOOTPROTO=static
IPADDR=10.10.2.222
NETMASK=255.255.255.0
GATEWAY=10.10.2.2
DNS1=223.5.5.5
EOF
systemctl restart network
systemctl stop NetworkManager
systemctl disable NetworkManager
tips: 配置复杂网络VLAN，子接口之类可以考虑打开NetworkManager使用nmcli工具配置
```

3. yum源

```
centos7
cd /etc/yum.repos.d/
tar -cf repos.tar *.repo
rm -rf *.repo
curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo
curl -o /etc/yum.repos.d/epel-7.repo http://mirrors.cloud.tencent.com/repo/epel-7.repo
wget -O /etc/yum.repos.d/CentOS-Base.repo http://mirrors.cloud.tencent.com/repo/centos7_base.repo
centos8
curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-8.repo
wget -O /etc/yum.repos.d/CentOS-Base.repo http://mirrors.cloud.tencent.com/repo/centos8_base.repo
```

4. 初始软件包  
   yum -y install vim mlocate htop
5. 初始目录  
   mkdir -p /data
6. 防火墙及selinux

```
systemctl stop firewalld
systemctl disable firewalld

setenfoce 0
sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config 

改ssh_port
vi /etc/ssh/sshd_config
#Port 22
Port 5022
firewall-cmd --permanent --add-port=5022/tcp
firewall-cmd --permanent --remove-service=ssh

firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
```

7. chrony  
   vim /etc/chrony.conf  
   server ntp1.aliyun.com iburst  
   server ntp2.aliyun.com iburst  
   server ntp3.aliyun.com iburst  
   server ntp4.aliyun.com iburst
8. 添加swap  
   swap大小一般为物理内存一般，exp:4G

```
dd if=/dev/zero of=/swapfile bs=1024 count=524288
mkswap /swapfile
chown root:root /swapfile
chmod 0600 /swapfile
swapon /swapfile
swapon -s

vi /etc/fstab
/swapfile              swap   swap     defaults     0 0

free -m
```

9. 更新  
   yum -y update

```
10. 按需
关selinux
setenforce 0
sed -i 's/enforcing/disabled/' /etc/selinux/config
关防火墙
systemctl stop firewalld && systemctl disable firewalld
```
