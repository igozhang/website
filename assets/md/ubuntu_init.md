# ubuntu_init系统初始化

env  
Ubuntu 20.04.1 LTS

1. 主机名  
   tee >/etc/hostname <<EOF  
   igo\_223  
   EOF

hostnamectl set-hostname igo\_223

```
systemctl disable --now systemd-networkd-wait-online.service
systemctl stop systemd-networkd-wait-online.service
systemctl mask systemd-networkd-wait-online.service
```

2. 网络

```
u18及以前版本
cp /etc/network/interfaces{,.$(date +%Y%m%d)bak}
tee >/etc/network/interfaces <<EOF
auto ens32
iface ens32 inet static
  address 10.10.2.223
  netmask 255.255.255.0
  gateway 10.10.2.2
  dns-nameserver 223.5.5.5
EOF
netplan apply

u20及以后版本
u20 /etc/netplan/00-installer-config.yaml
cat > /etc/netplan/00-installer-config.yaml <<'EOF'
network:
  ethernets:
    ens18:
      addresses: [172.27.248.66/24]
      gateway4: 172.27.248.254
      nameservers:
        addresses: [8.8.8.8]
  version: 2
EOF
netplan apply
systemctl restart systemd-resolved

https://igozhang.cn/network_config

#启用bbr模块
echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf && \
echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf && \
sysctl -p && lsmod | grep bbr
```

3. apt\_src

```
cp /etc/apt/sources.list{,.$(date +%Y%m%d)bak}
替换archive.ubuntu.com为mirrors.aliyun.com
#添加阿里源
cat << EOF > /etc/apt/sources.list
deb http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse
EOF
apt-get update
```

4. 初始软件包  
   apt -y install vim mlocate htop curl git unzip wget supervisor net-tools
5. 初始目录  
   mkdir /data
6. 开启root登录

```
sed -i 's/^.PermitRootLogin .*/PermitRootLogin yes/g' /etc/ssh/sshd_config
sed -i 's/^.PasswordAuthentication .*/PasswordAuthentication yes/g' /etc/ssh/sshd_config
```

7. 添加swap

```
dd if=/dev/zero of=/var/swap bs=1024 count=2048000 && \
chmod 600 /var/swap && \
mkswap /var/swap && \
swapon /var/swap && \
echo "/var/swap swap swap defaults 0 0" >> /etc/fstab && \
free -m

如需删除
swapoff /var/swap && rm /var/swap
# /etc/fstab
```

8. 时区  
   timedatectl set-timezone Asia/Shanghai  
   timedatectl set-timezone Europe/Budapest
9. 更新  
   更新软件包列表  
   apt update  
   软件更新并自动删除不使用的软件包  
   apt upgrade -y && apt autoremove -y
10. 启用crontab 日志  
    vim /etc/rsyslog.d/50-default.conf  
    取消注释#cron.\* /var/log/cron.log，保存并退出  
    service rsyslog restart
11. 清空history  
    history -c && cat /dev/null > ~/.bash\_history
