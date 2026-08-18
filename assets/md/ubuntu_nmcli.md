# ubuntu20.04使用nmcli管理网络

## 安装nmcli

```
apt -y install network-manager
```

### 配置

```
1. vim /etc/NetworkManager/NetworkManager.conf
managed=true 将false改为true

2. vim /etc/netplan/00-installer-config.yaml
vim /etc/netplan/00-installer-config.yaml
  renderer: NetworkManager 添加一行
```

### 应用配置

```
1. netplan apply
2. systemctl restart NetworkManager
```

### 检查配置

```
1. nmcli -p device 默认是unmanaged，可用状态为connected
2. nmcli con
```

### 配置网络

```
nmcli con add con-name nmcon-ens33 ifname ens33 type ethernet ipv4.method manual ipv4.add 10.10.8.81/24 ipv4.gateway 10.10.8.2 ipv4.dns 223.5.5.5

nmcli con up nmcon-ens33
```
