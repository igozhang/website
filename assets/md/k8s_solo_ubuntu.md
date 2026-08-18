# k8s_solo_ubuntu

在ubuntu20.04上使用kubeadm部署k8s1.29.10单节点  
使用本地目录'/data/storageclass'做默认sc  
避免使用最新版本，应该使用较新的稳定版本，所有k8s组件指定版本安装

## 环境准备

```
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

modprobe overlay
modprobe br_netfilter

cat > /etc/sysctl.d/k8s.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sysctl --system

ufw disable
```

## 安装

```
1. 安装containerd
apt update
apt install -y apt-transport-https ca-certificates curl gnupg lsb-release

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt update
apt install -y containerd.io=1.6.28-1
mkdir -p /etc/containerd
containerd config default | tee /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
systemctl restart containerd
systemctl enable containerd
systemctl status containerd --no-pager

2. 安装kubeadm 
mkdir -p /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /" | tee /etc/apt/sources.list.d/kubernetes.list
apt update

K8S_VERSION=1.29.10-1.1
apt install -y \
  kubelet=${K8S_VERSION} \
  kubeadm=${K8S_VERSION} \
  kubectl=${K8S_VERSION}

apt-mark hold kubelet kubeadm kubectl
systemctl enable --now kubelet
systemctl status kubelet --no-pager

3. 初始化k8s集群
kubeadm init \
  --kubernetes-version=v1.29.10 \
  --pod-network-cidr=192.168.0.0/16 \
  --apiserver-advertise-address=$(hostname -I | awk '{print $1}')

kubeadm join 10.80.238.88:6443 --token zno637.nnu9cl7g51654888 \
        --discovery-token-ca-cert-hash sha256:f44b44b03587715beeaa5fdfe15bd6c39002b6717803e7e161aa24039ce5b888

mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.27.3/manifests/calico.yaml
这个Calico插件有bird内存泄露问题，以后需要安装其他网络组件
 #先配置了CNI插件Calico集群才能ready

kubectl get nodes
kubectl version
 # 移除污点
kubectl taint nodes --all node-role.kubernetes.io/control-plane-
kubectl taint nodes --all node-role.kubernetes.io/master-
```

## 其他组件ingress，默认sc

```
kubectl create namespace ingress-nginx 2>/dev/null || true
kubectl create secret tls tls-default \
  --cert=/data/igozhang.cn/tls/igozhang.cn.pem \
  --key=/data/igozhang.cn/tls/igozhang.cn.key \
  -n ingress-nginx \
  --dry-run=client -o yaml | kubectl apply -f -

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx \
  --create-namespace \
  --version 4.11.5 \
  --set controller.kind=DaemonSet \
  --set controller.hostNetwork=true \
  --set controller.extraArgs.default-ssl-certificate=ingress-nginx/tls-default \
  --set controller.service.enabled=false \
  --set controller.ingressClassResource.default=true
```

## 使用本地目录做默认SC

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


## nerdctl,buildkit

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
 添加k8s拉取默认key：
kubectl create secret generic tcr-pull \
  --from-file=.dockerconfigjson=/root/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson \
  -n igo
kubectl patch serviceaccount default -n igo --type merge -p \
  '{"imagePullSecrets":[{"name":"tcr-pull"}]}'


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