# k8s添加node节点

```
Env
CentOS Linux release 8.2.2004
```

```
step：
1. 配置ipv4内核参数
2. 关闭swap
3. 安装docker
4. 安装 kubeadm kubelet kubectl
5. 检查环境一致性
5.1 hosts 一致
cat /etc/hosts
192.168.33.201 igo-k8s-1
192.168.33.202 igo-k8s-2
192.168.33.203 igo-k8s-3

5.2 daemon.json 一致
cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": ["https://59xo2v7a.mirror.aliyuncs.com"]，
  "exec-opts": ["native.cgroupdriver=systemd"]
}
EOF

6. kubeadm join 192.168.33.201:6443 --token el3q08.3h4gils1hjunxjm0 \
    --discovery-token-ca-cert-hash sha256:25a71a49f50141b5dd9fca76021d68918fc5657219a01ff15a085402ec474def

7. Done！
[root@igo-k8s-1 ~]# kubectl get node
NAME        STATUS   ROLES    AGE     VERSION
igo-k8s-1   Ready    master   43m     v1.19.16
igo-k8s-2   Ready    <none>   23m     v1.19.16
igo-k8s-3   Ready    <none>   8m32s   v1.19.16
```
