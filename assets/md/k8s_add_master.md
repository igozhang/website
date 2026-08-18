# k8s添加master节点

```
Env
CentOS Linux release 7.9.2009 (Core)

step：
1. 配置ipv4内核参数
2. 关闭swap
3. 安装docker
4. 安装 kubeadm kubelet kubectl
5. 检查环境一致性
5.1 hosts 一致
cat /etc/hosts
192.168.3.211 igo-k8s-1 k8s-api.ilinux.io
192.168.3.212 igo-k8s-2
192.168.3.213 igo-k8s-3
192.168.3.214 igo-k8s-4
192.168.3.215 igo-k8s-5
192.168.3.216 igo-k8s-6

5.2 daemon.json 一致
cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": ["https://59xo2v7a.mirror.aliyuncs.com"],
  "exec-opts": ["native.cgroupdriver=systemd"]
}
EOF

6. kubeadm join k8s-api.ilinux.io:6443 --token d0u3px.ilkidxxcl7opgy92 --discovery-token-ca-cert-hash sha256:b83bd7b8c07398726707f7b5e30ab68ebde922eb0ab81a004b2d777707632ed4 --certificate-key 9493f7a35f570e55386b94473eb1ddf351fd34943bc3a8d457e4f5451c729a6b --control-plane
找不到密钥了在master1上运行
kubeadm init phase upload-certs --upload-certs
重新生成密钥并以secret方式重新上传证书到kube-system_namespace中

7. Done！
[root@igo-k8s-1 ~]# kubectl get node
NAME        STATUS   ROLES    AGE     VERSION
igo-k8s-1   Ready    master   34m     v1.19.16
igo-k8s-2   Ready    <none>   25m     v1.19.16
igo-k8s-3   Ready    master   9m48s   v1.19.16

tips:
igo-k8s-1第一个主节点初始化必须带
--control-plane-endpoint  （单master可以不加）
参数：
kubeadm init --image-repository='registry.cn-hangzhou.aliyuncs.com/google_containers' --control-plane-endpoint k8s-api.ilinux.io --v=6
同时hosts增加k8s-api.ilinux.io 解析
192.168.3.201 igo-k8s-1 k8s-api.ilinux.io
192.168.3.202 igo-k8s-2
192.168.3.203 igo-k8s-3
192.168.3.204 igo-k8s-4
```
