# CentOS快速部署k8s

Env  
8.2.2004  
配置内核参数，将桥接的IPv4流量传递到iptables的链  
cat > /etc/sysctl.d/k8s.conf <<EOF  
net.bridge.bridge-nf-call-ip6tables = 1  
net.bridge.bridge-nf-call-iptables = 1  
EOF

sysctl -p

swapoff -a  
vim /etc/fstab

I**nstall\_docker 安装docker**

yum -y install yum-utils  
yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo  
yum -y install docker-ce-19.03.15-3.el8  
systemctl start docker && systemctl enable docker

**Install\_k8s 安装kubernetes/K8s**

cat <<EOF > /etc/yum.repos.d/kubernetes.repo  
[kubernetes]  
name=Kubernetes  
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86\_64/  
enabled=1  
gpgcheck=1  
repo\_gpgcheck=1  
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg  
EOF

setenforce 0

yum -y install kubeadm-1.19.16-0 kubelet-1.19.16-0 kubectl-1.19.16-0

初**始化k8s**

```
kubeadm init --image-repository='registry.cn-hangzhou.aliyuncs.com/google_containers'
Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 10.0.20.8:6443 --token btmzgv.4pj5tx16fpjwcc31 \
    --discovery-token-ca-cert-hash sha256:7e4b9a50c1333c643d0afdd933646c952e0317866f839ea2da26a2ab15b7a1a2

搭建单节点
kubectl taint nodes --all node-role.kubernetes.io/master-

初始化网络插件
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
or 
(如果用flannel，先wget，保持Network": "10.244.0.0/16和 kubeadm init 的--pod-network-cidr 一致)
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

# 添加 completion，最好放入 .bashrc 中
source <(kubectl completion bash)
source <(kubeadm completion bash)

Done！
kubectl create ns igo
kubectl create deployment nginx --image=nginx --replicas=1 -n igo
# kubectl get pod -A
NAMESPACE     NAME                                       READY   STATUS    RESTARTS   AGE
igo           nginx-6799fc88d8-xxnwl                     1/1     Running   0          6m9s
kube-system   calico-kube-controllers-659bd7879c-rnrqr   1/1     Running   0          8m4s
kube-system   calico-node-zx42j                          1/1     Running   0          8m4s
kube-system   coredns-6c76c8bb89-k7h26                   1/1     Running   0          10m
kube-system   coredns-6c76c8bb89-vkt7c                   1/1     Running   0          10m
kube-system   etcd-igo-cent82                            1/1     Running   0          10m
kube-system   kube-apiserver-igo-cent82                  1/1     Running   0          10m
kube-system   kube-controller-manager-igo-cent82         1/1     Running   0          10m
kube-system   kube-proxy-56hzd                           1/1     Running   0          10m
kube-system   kube-scheduler-igo-cent82                  1/1     Running   0          10m
```

工具和思路

```
yum list docker-ce --showduplicates

error:
[WARNING IsDockerSystemdCheck]: detected "cgroupfs" as the Docker cgroup driver. The recommended driver is "systemd". Please follow the guide at https://kubernetes.io/docs/setup/cri/

cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": ["https://59xo2v7a.mirror.aliyuncs.com"],
  "exec-opts": ["native.cgroupdriver=systemd"]
}
EOF

yum install iproute-tc
否则报警：[preflight] WARNING: tc not found in system path

kubeadm init 参数
–kubernetes-version: 用于指定k8s版本
–apiserver-advertise-address：用于指定kube-apiserver监听的ip地址,就是 master本机IP地址
–pod-network-cidr：用于指定Pod的网络范围
–service-cidr：用于指定SVC的网络范围
–image-repository: 指定阿里云镜像仓库地址
```
