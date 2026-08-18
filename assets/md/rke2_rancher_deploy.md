# rke2_rancher_deploy

Ubuntu部署rke2+rancher  
Ubuntu 20.04  
docker 不可以也无需预安装  
RKE2 RKE2 v1.28.12+rke2r1 Kubernetes 1.28 LTS 版本，RKE2 该版本适配国内环境，兼容性最佳  
Rancher v2.9.3 2.9 系列长期稳定版，修复了 v2.9.12 的小 bug，社区反馈更成熟  
https://github.com/rancher/rke2/releases/tag/v1.28.12%2Brke2r1  
rancher-huizhou01.igozhang.cn

## IP规划信息

规划网段  
10.10.81.51-53 master  
10.10.81.54-59 worker  
gateway  
10.10.81.254  
dns  
10.10.81.210  
hostname  
k8s-master51-53  
k8s-worker54-59

## 初始化命令

```
useradd -m -G sudo userroot && passwd userroot

hostnamectl set-hostname k8s-master51
hostnamectl set-hostname k8s-master52
依次类推直到 k8s-worker59

在所有节点的 /etc/hosts 追加以下内容
cat >> /etc/hosts <<EOF
10.10.81.51 k8s-master51
10.10.81.52 k8s-master52
10.10.81.53 k8s-master53
10.10.81.54 k8s-worker54
10.10.81.55 k8s-worker55
10.10.81.56 k8s-worker56
10.10.81.57 k8s-worker57
10.10.81.58 k8s-worker58
10.10.81.59 k8s-worker59
EOF

1. 打开密码访问：
passwd root
vim /etc/ssh/sshd_config
找到 PermitRootLogin 行，修改为 yes（默认是 prohibit-password 或 no）
PermitRootLogin yes
PasswordAuthentication yes
根据需要修改port
Port 22
systemctl restart sshd

2. 关闭自动更新
systemctl stop apt-daily.timer
systemctl disable apt-daily.timer
systemctl stop apt-daily-upgrade.timer
systemctl disable apt-daily-upgrade.timer
sed -i -e 's/^APT::Periodic::Update-Package-Lists.*/APT::Periodic::Update-Package-Lists "0";/' -e 's/^APT::Periodic::Unattended-Upgrade.*/APT::Periodic::Unattended-Upgrade "0";/' /etc/apt/apt.conf.d/20auto-upgrades

3. 扩容磁盘：默认只占用30G，需要把剩余的磁盘分配给/目录
systemctl stop unattended-upgrades
systemctl disable unattended-upgrades  # 永久禁用（谨慎！会影响系统安全更新）
apt update
apt install cloud-guest-utils
growpart /dev/vda 3
partprobe /dev/vda
pvresize /dev/vda3
df -Th
lsblk
lvextend -l +100%FREE /dev/mapper/ubuntu--vg-ubuntu--lv
resize2fs /dev/mapper/ubuntu--vg-ubuntu--lv
df -h /

4. swapoff并永久关闭 '注释掉 /etc/fstab 中的 swap' 
swapoff -a
sed -i '/swap/s/^/#/' /etc/fstab

5. 防火墙ufw
ufw disable
systemctl stop ufw
systemctl disable ufw

6. 时区
timedatectl set-timezone Asia/Shanghai

timedatectl set-timezone Europe/Budapest

7. 加载内核模块
cat >> /etc/modules-load.d/k8s.conf <<EOF
overlay
br_netfilter
EOF
modprobe overlay
modprobe br_netfilter

8. 优化 sysctl 参数
cat >> /etc/sysctl.d/k8s.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
vm.swappiness = 0
fs.inotify.max_user_watches=524288
fs.inotify.max_user_instances=512
EOF
sysctl --system

验证
sysctl -n net.bridge.bridge-nf-call-iptables
sysctl -n net.bridge.bridge-nf-call-ip6tables
sysctl -n net.ipv4.ip_forward
都应该 输出1

9. 配置阿里源
mv /etc/apt/sources.list /etc/apt/sources.list.bak
cat > /etc/apt/sources.list << EOF
deb http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse
EOF

apt update && apt upgrade -y
  基础工具
apt install -y curl wget vim mlocate htop lrzsz jq socat conntrack ipvsadm ipset nfs-common
```

### 用一台机器克隆其他虚机后需检查项

1. IP,hostname
2. Machine-ID (机器唯一标识)

```
# 清空 machine-id并重新生成
truncate -s 0 /etc/machine-id
rm /var/lib/dbus/machine-id
ln -s /etc/machine-id /var/lib/dbus/machine-id
systemd-machine-id-setup
```

3. SSH Host Keys (SSH 主机密钥) /etc/ssh/ssh\_host\_\*文件包含了服务器的身份密钥  
   删除并重新生成

```
rm /etc/ssh/ssh_host_*
ssh-keygen -A
systemctl restart sshd
sshd -t
```

4. Cloud-Init 状态  
   如果母机是通过 Cloud-Init (常见于云厂商或 OpenStack/深信服) 初始化的，它会生成一个标记文件 /var/lib/cloud/instances/…。克隆后，新机器会认为自己已经初始化过了，不再执行用户数据脚本（如设置密码、写入 SSH 键）  
   清除 cloud-init 状态，使其下次启动时重新运行

```
cloud-init clean --logs
```

6. 历史命令与日志(可选但推荐)  
   克隆过来的 .bash\_history 里有母机的操作记录，日志里也有母机的时间戳，容易造成运维混淆  
   清理旧日志及历史命令

```
journalctl --rotate
journalctl --vacuum-time=1s
cat /dev/null > ~/.bash_history
history -c
```

### ssh 互信

```
第一步： k8s_master51第一台机器上生成 SSH 密钥
生成密钥对，一路回车即可（不要设置密码短语，否则脚本无法自动登录）
ssh-keygen -t rsa -b 4096
第二部： 将公钥分发到其他 8 台机器
ssh-copy-id root@10.10.81.52
```

## 部署rke2k8s

```
INSTALL_RKE2_VERSION=v1.28.12+rke2r1 curl -sfL https://get.rke2.io | sudo sh -
mkdir -p mkdir -p /root/rke2-artifacts
cp rke2.linux-amd64.tar.gz rke2-images.linux-amd64.tar.zst sha256sum-amd64.txt /root/rke2-artifacts
INSTALL_RKE2_ARTIFACT_PATH=/root/rke2-artifacts/ sh install.sh

mkdir -p /etc/rancher/rke2/
cat > /etc/rancher/rke2/config.yaml << EOF
write-kubeconfig-mode: "0600"
node-ip: "10.10.81.51"
tls-san:
  - "10.10.81.51"
  - "k8s-master51"
cluster-cidr: "10.42.0.0/16"
service-cidr: "10.43.0.0/16"
system-default-registry: "registry.cn-hangzhou.aliyuncs.com"
EOF

最后
systemctl enable rke2-server
systemctl start rke2-server

获取token
cat /var/lib/rancher/rke2/server/node-token

# 指定kubectl,kubeconfig路径并生效
echo "export PATH=$PATH:/var/lib/rancher/rke2/bin" >> ~/.bashrc 
echo "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml" >> ~/.bashrc
source ~/.bashrc

TIPS 跟踪进度：
手动下载一个大包：
nohup wget --progress=bar:force "https://github.com/rancher/rke2/releases/download/v1.28.12%2Brke2r1/rke2-images.linux-amd64.tar.zst" -O rke2-images.linux-amd64.tar.zst > download-progress.log 2>&1 &
tail /root/igo/soft/download-progress.log
查看启动日志
journalctl -u rke2-server --no-pager | head -30

# 其他master52，53
将三个文件都放到images目录后使用install脚本安装：
mkdir -p mkdir -p /root/rke2-artifacts
cp rke2.linux-amd64.tar.gz rke2-images.linux-amd64.tar.zst sha256sum-amd64.txt /root/rke2-artifacts
INSTALL_RKE2_ARTIFACT_PATH=/root/rke2-artifacts/ sh install.sh

rke2 --version

mkdir -p /etc/rancher/rke2
cat > /etc/rancher/rke2/config.yaml << EOF
write-kubeconfig-mode: "0600"
tls-san:
  - "10.10.81.52"
  - "k8s-master52"
server: "https://10.10.81.51:9345"
token: "K10c4414c0f71c4f56cfb5456b4c506d89d675dbaf7e62297685e80e977b15af9b1::server:69fec9ff4edc66d08cce20260a6a9fa9"
system-default-registry: "registry.cn-hangzhou.aliyuncs.com"
EOF

mkdir -p /etc/rancher/rke2
cat > /etc/rancher/rke2/config.yaml << EOF
write-kubeconfig-mode: "0600"
tls-san:
  - "10.10.81.53"
  - "k8s-master53"
server: "https://10.10.81.51:9345"
token: "K10c4414c0f71c4f56cfb5456b4c506d89d675dbaf7e62297685e80e977b15af9b1::server:69fec9ff4edc66d08cce20260a6a9fa9"
system-default-registry: "registry.cn-hangzhou.aliyuncs.com"
EOF

systemctl enable rke2-server
systemctl start rke2-server

# worker节点
mkdir -p mkdir -p /root/rke2-artifacts
cp rke2.linux-amd64.tar.gz rke2-images.linux-amd64.tar.zst sha256sum-amd64.txt /root/rke2-artifacts
INSTALL_RKE2_TYPE=agent INSTALL_RKE2_ARTIFACT_PATH=/root/rke2-artifacts/ sh install.sh

mkdir -p /etc/rancher/rke2
cat > /etc/rancher/rke2/config.yaml << EOF
node-ip: "10.10.81.54"
server: "https://10.10.81.51:9345"
token: "K10198217f703319d08db248d83e53c6e2e2e618c8b6e32b160026f42a20b77dadc::server:825193642931036706fd0ae0c144aa32"
system-default-registry: "registry.cn-hangzhou.aliyuncs.com"
EOF

systemctl enable rke2-agent
systemctl start rke2-agent
```

## 部署rancher

rancher-huizhou01.igozhang.cn  
部署rancher  
Rancher v2.9.3  
需要安装：helm-v3.14.0，cert-manager-v1.12.0，和rancher-2.9.3

1. 安装helm

```
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
或者离线
tar -zxvf helm-v3.14.0-linux-amd64.tar.gz
mv linux-amd64/helm /usr/local/bin/
helm version
```

2. 安装rancher

```
kubectl create namespace cattle-system
kubectl -n cattle-system create secret \
  tls tls-rancher-ingress \
  --cert=./tls.pem \
  --key=./tls.key

helm install rancher ./rancher-2.9.3.tgz \
  --namespace cattle-system \
  --create-namespace \
  --set hostname=rancher-huizhou01.igozhang.cn \
  --set bootstrapPassword=admin \
  --set ingress.tls.source=secret \
  --set image.repository=registry.cn-hangzhou.aliyuncs.com/rancher/rancher \
  --set systemDefaultRegistry=registry.cn-hangzhou.aliyuncs.com

安装完了发现POD都起不来ImagePullBackOff状态需要离线下载镜像包并上传到POD节点，一般为所有WORKER节点
ctr没下载下来，装了个docker下：
docker pull registry.cn-hangzhou.aliyuncs.com/rancher/rancher:v2.9.3
docker images | grep "rancher/rancher"
docker save -o rancher-v2.9.3.tar rancher/rancher:v2.9.3
scp到所有worker节点，再导入，就可以拉起rancher了
ctr -n k8s.io images import rancher-v2.9.3.tar
ctr -n k8s.io images ls | grep rancher

公网环境使用  --set letsEncrypt.email=igozhang@example.com \配置来接受letsEncrypt的证书到期提醒

验证：
kubectl get pods -n cattle-system
kubectl get ingress -n cattle-system
kubectl rollout status deployment -n cattle-system rancher

设置一条hosts,然后就可以通过域名访问rancher了
10.10.81.55 rancher-huizhou01.igozhang.cn
```

## 补全环境变量

```
1. kubectl ctr crictl
echo "export PATH=$PATH:/var/lib/rancher/rke2/bin" >> ~/.bashrc 
echo "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml" >> ~/.bashrc
source ~/.bashrc

echo "export CONTAINERD_ADDRESS=/run/k3s/containerd/containerd.sock" >> ~/.bashrc
source ~/.bashrc
ctr -n k8s.io images ls

echo "export CRI_CONFIG_FILE=/var/lib/rancher/rke2/agent/etc/crictl.yaml" >> ~/.bashrc
source ~/.bashrc
crictl images
```

kubectl delete pod -n cattle-system -l app=rancher --force --grace-period=0

## 报错

```
Error: UPGRADE FAILED: failed to create resource: Internal error occurred: failed calling webhook "validate.nginx.ingress.kubernetes.io": failed to call webhook: Post "https://rke2-ingress-nginx-controller-admission.kube-system.svc:443/networking/v1/ingresses?timeout=10s": context deadline exceeded
root@k8s-master51:/igo/soft/rancher# kubectl edit validatingwebhookconfigurations rke2-ingress-nginx-admission

排查结论：**VXLAN 模式下，Worker 发出的 UDP 8472 包无法到达 Master**。
- Master 发出的 VXLAN 包可以到达 Worker
- Worker 发出的 VXLAN 包无法到达 Master（tcpdump 在 Master 上抓不到）
- Worker 本机 nc 测试发往 Master:8472 的包可以到达（说明链路本身可用）
- 推测为内核 VXLAN 封装或中间网络设备对部分 VXLAN 流量的处理异常（可能与 kernel 5.4 相关）

RKE2 v1.28.12这套版本有bug仅做安装参考
临时修复规避：

kubectl patch configmap rke2-canal-config -n kube-system --type merge -p '{"data":{"net-conf.json":"{\n  \"Network\": \"10.42.0.0/16\",\n  \"Backend\": {\n    \"Type\": \"host-gw\"\n  }\n}\n"}}'
kubectl rollout restart daemonset rke2-canal -n kube-system
```
