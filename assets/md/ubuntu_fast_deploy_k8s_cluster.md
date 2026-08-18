# ubuntu快速部署k8s

```
有问题多看看官网  
https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/install-kubeadm/  
重要概念  
cgroups  
美团的技术团队有好的文章：  
https://tech.meituan.com/2015/03/31/cgroups.html
```

```
ubuntu部署k8s
 
Env:
Ubuntu20.04
swapoff -a
vim /etc/fstab

配置内核参数，将桥接的IPv4流量传递到iptables的链
cat > /etc/sysctl.d/k8s.conf <<EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system

docker_install
curl -fsSL http://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | sudo apt-key add -
add-apt-repository "deb [arch=amd64] http://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable"
apt -y install docker-ce=5:19.03.15~3-0~ubuntu-focal

cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": ["https://59xo2v7a.mirror.aliyuncs.com"]，
  "exec-opts": ["native.cgroupdriver=systemd"]
}
EOF

kubernetes_install
谷歌阿里二选一，国内建议阿里
信任证书和仓库，谷歌
sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list

信任证书和仓库，阿里
# 添加并信任APT证书
curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | sudo apt-key add -
# 添加源地址
add-apt-repository "deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main"
# 更新源并安装 kubenetes
apt updatel
apt -y install kubeadm=1.19.16-00 kubectl=1.19.16-00 kubelet=1.19.16-00
apt-mark hold kubelet kubeadm kubectl

启动主节点
kubeadm init --image-repository='registry.cn-hangzhou.aliyuncs.com/google_containers'

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 10.0.20.8:6443 --token y4lv3e.olrwcvbp5u8ie51d \
    --discovery-token-ca-cert-hash sha256:be21af30658baaa4bfa08ac6349ab572f5159a67d36aad35cb82f8773658598c

# 添加 completion，最好放入 .bashrc 中
source <(kubectl completion bash)
source <(kubeadm completion bash)

kubectl taint nodes --all node-role.kubernetes.io/master-
控制平面节点隔离：
默认情况下，出于安全原因，你的集群不会在控制平面节点上调度 Pod，运行上面命令相当于启动了一个单点k8s集群

root@igo-h1:~# kubectl get node
NAME     STATUS     ROLES    AGE   VERSION
igo-h1   NotReady   master   26m   v1.19.16

node节点为NotReady，因为corednspod没有启动，缺少网络组件
安装calico网络
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml

完，Done
root@igo-h1:~# kubectl create ns igo
namespace/igo created
root@igo-h1:~# kubectl create deployment nginx --image=nginx -n igo
deployment.apps/nginx created
root@igo-h1:~# kubectl get pod -A
NAMESPACE     NAME                                       READY   STATUS              RESTARTS   AGE
igo           nginx-6799fc88d8-s57zt                     0/1     ContainerCreating   0          14s
kube-system   calico-kube-controllers-659bd7879c-t6tc4   1/1     Running             0          3m44s
kube-system   calico-node-2g4tb                          1/1     Running             0          3m44s
kube-system   coredns-6c76c8bb89-2hwxt                   1/1     Running             0          6m28s
kube-system   coredns-6c76c8bb89-5lwhp                   1/1     Running             0          6m27s
kube-system   etcd-igo-h1                                1/1     Running             0          6m27s
kube-system   kube-apiserver-igo-h1                      1/1     Running             0          6m27s
kube-system   kube-controller-manager-igo-h1             1/1     Running             0          6m27s
kube-system   kube-proxy-dgcrw                           1/1     Running             0          6m28s
kube-system   kube-scheduler-igo-h1                      1/1     Running             0          6m27s
root@igo-h1:~#

工具和思路：

apt-cache madison vim

kubeadm join --token <token> <control-plane-host>:<control-plane-port> --discovery-token-ca-cert-hash sha256:<hash>
var1_token
kubeadm token list
var2_controlhost-controlhost
192.168.2.10:6443
var3_discovery-token-ca-cert-hash
openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null |    openssl dgst -sha256 -hex | sed 's/^.* //'

失败的话修复后需要重新reset：
kubeadm reset 再 kubeadm init
```
