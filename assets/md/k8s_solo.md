# k8s_solo

在wp服务器上部署k8s,kubeadm单节点，跑容器，用来共享igozhang.cn域名并提供服务  
示例： igozhang.cn/trans  
在wp服务器上部署k8s，跑容器，用来共享igozhang.cn域名并提供服务  
示例： igozhang.cn/trans  
kubeadm部署单节点k8s,用来共享igoa域名  
OpenCloudOS 9 (腾讯云主机)  
（兼容 RHEL9/CentOS Stream 9），全程 root 权限执行

## 一、彻底清理干净（回退）

```
kubeadm reset -f
systemctl stop kubelet containerd
yum remove -y kubeadm kubelet kubectl containerd
rm -rf /etc/kubernetes /var/lib/etcd /var/lib/kubelet /etc/cni/net.d $HOME/.kube
iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X
```

## 二、基础环境配置

```
swapoff -a
sed -i '/swap/s/^/#/' /etc/fstab
systemctl stop firewalld && systemctl disable firewalld
setenforce 0
sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

cat <<EOF | tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF
modprobe overlay
modprobe br_netfilter

cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
net.ipv4.ip_forward=1
EOF
sysctl --system
```

## 三、安装 containerd **系统自带稳定版（1.6.x 固定）**

这一步**不会装最新版**，是 OS 官方源里的稳定版，最安全：

```
# 安装 OS 自带稳定版 containerd (1.6.x LTS)
yum install -y containerd

# 生成配置 + 修改 systemd cgroup
containerd config default > /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml

systemctl enable --now containerd
```

## 四、安装必备依赖

```
yum install -y conntrack-tools
修复BUG：containernetworking-plugins 把二进制装在 /usr/libexec/cni，而 kubelet 默认只在 /opt/cni/bin 里找 CNI 插件，路径不一致导致报找不到 loopback
mkdir -p /opt/cni/bin
ln -sf /usr/libexec/cni/* /opt/cni/bin/
ls -l /opt/cni/bin/loopback /opt/cni/bin/bridge

systemctl restart kubelet
```

## 五、配置 K8s 1.29 源

```
cat <<EOF | tee /etc/yum.repos.d/kubernetes.repo
```

[kubernetes]

name=Kubernetes baseurl=https://pkgs.k8s.io/core:/stable:/v1.29/rpm/ enabled=1 gpgcheck=1 gpgkey=https://pkgs.k8s.io/core:/stable:/v1.29/rpm/repodata/repom.xml.key exclude=kubelet kubeadm kubectl EOF

## 六、安装 **K8s 固定版本 1.29.10**

```
yum install -y kubelet-1.29.10 kubeadm-1.29.10 kubectl-1.29.10 --disableexcludes=kubernetes
systemctl enable kubelet
```

## 七、初始化集群（指定版本）

```
kubeadm init --kubernetes-version=v1.29.10 --pod-network-cidr=10.244.0.0/16
```

## 八、初始化成功后执行（必须）

```
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config

kubectl taint nodes --all node-role.kubernetes.io/control-plane-

kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```

## 九、验证,OUTPUT

```
kubectl get nodes

Alternatively, if you are the root user, you can run:

  export KUBECONFIG=/etc/kubernetes/admin.conf

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 10.8.8.8:6443 --token dbdmt7.3lh3nugd6uk9qigo \
    --discovery-token-ca-cert-hash sha256:db9eebcc94e96e38e66e5a941e35469f7cca1e7633fc05dd7f18c9bc5fcd3igo
```

## 十、 添加master和worker

先装好环境：containerd、kubeadm、kubelet（版本必须和主节点一致：1.29.10）  
添加worker:

```
kubeadm join 10.8.8.8:6443 --token dbdmt7.3lh3nugd6uk9qigo \
--discovery-token-ca-cert-hash sha256:db9eebcc94e96e38e66e5a941e35469f7cca1e7633fc05dd7f18c9bc5fcd3igo
```

添加master

```
kubeadm init phase upload-certs --upload-certs  #这条命令会输出一个certificate-key
kubeadm join 10.8.8.8:6443 --token dbdmt7.3lh3nugd6uk9qigo \
--discovery-token-ca-cert-hash sha256:db9eebcc94e96e38e66e5a941e35469f7cca1e7633fc05dd7f18c9bc5fcd3igo \
--control-plane \
--certificate-key 上面获取的certificate-key
```

---

## 版本（全部稳定、可追溯）

- containerd：**1.6.x（OpenCloudOS 9 官方稳定版）**
- kubeadm / kubelet / kubectl：**1.29.10**
- Kubernetes 集群：**v1.29.10**

## 后续helm，ingress-nginx,harbor等

1. helm固定版本（可改此变量重跑以升级/切换版本）

```
HELM_VERSION=v3.15.4
ARCH=amd64   # 若是 ARM 云主机改为 arm64

cd /tmp
curl -fsSL -o "helm-${HELM_VERSION}-linux-${ARCH}.tar.gz" \
  "https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz"
tar -xzf "helm-${HELM_VERSION}-linux-${ARCH}.tar.gz"
install -m 0755 "linux-${ARCH}/helm" /usr/local/bin/helm
rm -rf "linux-${ARCH}" "helm-${HELM_VERSION}-linux-${ARCH}.tar.gz"

helm version
```

2. ingress-nginx

```
kubectl create namespace ingress-nginx 2>/dev/null || true

kubectl create secret tls tls-default \
  --cert=/www/server/panel/vhost/cert/wordpress.local/igozhang.cn.pem \
  --key=/www/server/panel/vhost/cert/wordpress.local/igozhang.cn.key \
  -n ingress-nginx \
  --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --version 4.11.5 \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.replicaCount=1 \
  --set controller.service.type=NodePort \
  --set controller.service.nodePorts.http=30080 \
  --set controller.service.nodePorts.https=30443 \
  --set controller.config.use-forwarded-headers=true \
  --set controller.config.compute-full-forwarded-for=true \
  --set controller.config.use-proxy-protocol=false \
  --set-string 'controller.extraArgs.default-ssl-certificate=ingress-nginx/tls-default' \
  --wait \
  --timeout 10m
```

3. 使用本地目录做默认SC

```
mkdir -p /data/storageclass
chmod 775 /data/storageclass

kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.30/deploy/local-path-storage.yaml

kubectl -n local-path-storage patch configmap local-path-config --type merge -p \
'{"data":{"config.json":"{\n  \"nodePathMap\":[\n    {\n      \"node\":\"DEFAULT_PATH_FOR_NON_LISTED_NODES\",\n      \"paths\":[\"/data/storageclass\"]\n    }\n  ]\n}"}}'

kubectl -n local-path-storage rollout restart deployment local-path-provisioner

kubectl patch storageclass "$(kubectl get sc -o jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}' 2>/dev/null)" \
  -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}' 2>/dev/null || true

kubectl patch storageclass local-path \
  -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

4. nerdctl

```
NERDCTL_VERSION=2.1.6
ARCH=amd64

cd /tmp
curl -fsSL -o "nerdctl-${NERDCTL_VERSION}-linux-${ARCH}.tar.gz" \
  "https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${ARCH}.tar.gz"

tar -xzf "nerdctl-${NERDCTL_VERSION}-linux-${ARCH}.tar.gz"
install -m 0755 nerdctl /usr/local/bin/nerdctl
rm -f nerdctl "nerdctl-${NERDCTL_VERSION}-linux-${ARCH}.tar.gz"
echo 'export CONTAINERD_ADDRESS=/run/containerd/containerd.sock' >> /root/.bashrc
nerdctl --version

nerdctl pull nginx:alpine
nerdctl login --username=100010341179 ccr.ccs.tencentyun.com    广州
nerdctl login --username=100010341179 krccr.ccs.tencentyun.com  首尔
nerdctl tag nginx:alpine ccr.ccs.tencentyun.com/igozhang/nginx:alpine
nerdctl push ccr.ccs.tencentyun.com/igozhang/nginx:alpine
然后就可以pull了
nerdctl pull ccr.ccs.tencentyun.com/igozhang/nginx:alpine
nerdctl images

kubeadm / 集群安装不包含 BuildKit；nerdctl build 必须单独装 buildctl + buildkitd

BK_VERSION=v0.13.2
ARCH=amd64
cd /tmp
curl -fsSL -o "buildkit-${BK_VERSION}.linux-${ARCH}.tar.gz" \
  "https://github.com/moby/buildkit/releases/download/${BK_VERSION}/buildkit-${BK_VERSION}.linux-${ARCH}.tar.gz"
sudo tar -xzf "buildkit-${BK_VERSION}.linux-${ARCH}.tar.gz" -C /usr/local
rm -f "buildkit-${BK_VERSION}.linux-${ARCH}.tar.gz"
ls -l /usr/local/bin/buildctl /usr/local/bin/buildkitd

systemd 常驻 buildkitd（与 nerdctl 默认 socket 一致）
sudo tee /etc/systemd/system/buildkit.service >/dev/null <<'EOF'
[Unit]
Description=BuildKit
Documentation=https://github.com/moby/buildkit
After=network-online.target local-fs.target containerd.service
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/buildkitd --addr unix:///run/buildkit/buildkitd.sock --oci-worker=true --containerd-worker=false
Restart=always
RestartSec=2
RuntimeDirectory=buildkit

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now buildkit
systemctl is-active buildkit && buildctl --version
```
