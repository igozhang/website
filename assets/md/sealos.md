# 一剑k8s_sealos

```
sealos init \
--master 172.17.35.10 \
--master 172.17.35.11 \
--master 172.17.35.12 \
--node   172.17.35.13 \
--node   172.17.35.14 \
--node   172.17.35.15 \
--podcidr 10.244.0.0/10 \
--user root \
--passwd 'igo@1234' \
--version v1.19.11 \
--pkg-url=kube1.19.11.tar.gz
```

## 集群安装

```
export PROXY_PREFIX=https://ghfast.top

获取版本列表
curl --silent "https://api.github.com/repos/labring/sealos/releases" | jq -r '.[].tag_name'

软件列表
https://github.com/labring-actions/cluster-image-docs/blob/main/docs/docker/rootfs.md
https://github.com/labring-actions/cluster-image-docs/blob/main/docs/docker/apps.md

二进制amd64手动下载：
export VERSION="v5.1.1"
export VERSION="v4.3.6"
wget ${PROXY_PREFIX}/https://github.com/labring/sealos/releases/download/${VERSION}/sealos_${VERSION#v}_linux_amd64.tar.gz \
&& tar zxvf sealos_${VERSION#v}_linux_amd64.tar.gz sealos && chmod +x sealos && mv sealos /usr/bin
sealos version

sealos run registry.cn-shanghai.aliyuncs.com/labring/kubernetes-docker:v1.24.6 \
  registry.cn-shanghai.aliyuncs.com/labring/helm:v3.19.2  \
  registry.cn-shanghai.aliyuncs.com/labring/calico:3.26.5 \
  --masters 10.201.81.51,10.201.81.52,10.201.81.53 \
  --nodes 10.201.81.54,10.201.81.55,10.201.81.56,10.201.81.57,10.201.81.58,10.201.81.59 \
  --port 5008 \
  -p igopassword

echo "alias k='kubectl'" >> ~/.bashrc
echo "alias kd='kubectl describe'" >> ~/.bashrc
echo "alias kg='kubectl get'" >> ~/.bashrc
echo "alias ka='kubectl apply'" >> ~/.bashrc
source ~/.bashrc

To start administering your cluster from this node, you need to run the following as a regular user:
        mkdir -p $HOME/.kube
        sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
        sudo chown $(id -u):$(id -g) $HOME/.kube/config

增减节点
$ sealos add --nodes 192.168.64.21,192.168.64.19 
$ sealos add --masters 192.168.64.21,192.168.64.19 
$ sealos delete --nodes 192.168.64.21,192.168.64.19 
$ sealos delete --masters 192.168.64.21,192.168.64.19  
清理 K8s 集群
$ sealos reset

集群离线安装
下载打包pull,save
sealos pull registry.cn-shanghai.aliyuncs.com/labring/kubernetes-docker:v1.24.6
sealos pull registry.cn-shanghai.aliyuncs.com/labring/helm:v3.3.4
sealos pull registry.cn-shanghai.aliyuncs.com/labring/helm:v3.19.2
sealos pull registry.cn-shanghai.aliyuncs.com/labring/calico:3.26.5
sealos pull registry.cn-shanghai.aliyuncs.com/labring/ingress-nginx:4.1.0
sealos pull registry.cn-shanghai.aliyuncs.com/labring/ingress-nginx:v1.8.1
sealos pull registry.cn-shanghai.aliyuncs.com/labring/cert-manager:v1.12.13
sealos pull registry.cn-shanghai.aliyuncs.com/labring/rancher:v2.6.9
sealos pull registry.cn-shanghai.aliyuncs.com/labring/nfs-subdir-external-provisioner:v4.0.18

sealos save -o kubernetesv1.19.16.tar registry.cn-shanghai.aliyuncs.com/labring/kubernetes:v1.19.16
sealos save -o helmv3.3.4.tar         registry.cn-shanghai.aliyuncs.com/labring/helm:v3.3.4
sealos save -o ciliumv1.13.4.tar      registry.cn-shanghai.aliyuncs.com/labring/cilium:v1.13.4
sealos save -o ingress-nginx4.1.0    registry.cn-shanghai.aliyuncs.com/labring/ingress-nginx:4.1.0

上传加仓库
$ sealos load -i kubernetesv1.19.16.tar 就可以正常安装
sealos load -i nginx-latest.tar
$ sealos images 

也可以不导入，直接tar包安装
$ sealos run kubernetesv1.19.16.tar # 单机安装，集群安装同理

ctr -n k8s.io image tag harbor-dianbai.sunwoda-evb.com/rancher/mirrored-library-nginx@sha256:bd0aa91fe6a182db22032463c17644cd2ff3bbe415e7b84964283bba687acaa6 nginx:latest

ctr -n k8s.io image export /igo/soft/nginx-latest.tar nginx:latest
ctr -n k8s.io image import nginx-latest.tar
crictl images
```

### 部署rancher2.6.9

```
建证书
kubectl create namespace cattle-system
kubectl -n cattle-system create secret \
  tls tls-rancher-ingress \
  --cert=./tls.pem \
  --key=./tls.key

必须安装cert-manager才能装rancher
sealos run registry.cn-shanghai.aliyuncs.com/labring/cert-manager:v1.12.13
sealos run registry.cn-shanghai.aliyuncs.com/labring/rancher:v2.6.9
rancher:v2.6.9要求k8s 1.25以下
sealos run registry.cn-shanghai.aliyuncs.com/labring/rancher:v2.6.9 --env hostname=rancher.my.org --env ingressClassName=nginx
    admin/rancher
安装完后开始改配置，用自己的域名和证书：
kubectl edit ingress rancher -n cattle-system
删掉以下三行（取消cert-manager证书托管，删掉自由证书使用ingress默认证书）
cert-manager.io/issuer: rancher
cert-manager.io/issuer-kind: Issuer
secretName: tls-rancher-ingress

卸载重装：
helm uninstall rancher -n cattle-system
kubectl patch namespace cattle-system   -p '{"metadata":{"finalizers":[]}}'   --type=merge
kubectl get namespace cattle-system -o json   | jq 'del(.spec.finalizers)'   | kubectl replace --raw "/api/v1/namespaces/cattle-system/finalize" -f -
kubectl delete namespace cattle-system
kubectl get all -n cattle-system
rancher重装后一些还是原来地址的报错，修改url（rancher_UI->全局设置->搜索server-url）

1. 修改关联配置：
kubectl patch ingress rancher -n cattle-system --type='json' -p='[
  {"op": "replace", "path": "/spec/rules/0/host", "value": "rancher-huizhou01.igozhang.cn"},
  {"op": "replace", "path": "/spec/tls/0/hosts/0", "value": "rancher-huizhou01.igozhang.cn"},
  {"op": "replace", "path": "/spec/tls/0/secretName", "value": "tls-rancher-ingress"},
  {"op": "remove", "path": "/metadata/annotations/cert-manager.io~1issuer"},
  {"op": "remove", "path": "/metadata/annotations/cert-manager.io~1issuer-kind"}
]'

2. 取消cert-manager托管，否则一直会用自动生成的证书匹配覆盖你的证书
kubectl get certificate -n cattle-system
kubectl delete certificate tls-rancher-ingress -n cattle-system
kubectl delete secret tls-rancher-ingress -n cattle-system

3. 并获取当前的 settings 配置
kubectl get settings server-url -n cattle-system -o json | \
jq '.value = "https://rancher-huizhou01.igozhang.cn"' | \
kubectl apply -f -

4. 建证书

5. 重启
kubectl rollout restart deployment rancher -n cattle-system
kubectl rollout restart deployment ingress-nginx-controller -n ingress-nginx
```

## 其他应用

```
sealos run registry.cn-shanghai.aliyuncs.com/labring/openebs:v3.9.0
sealos run registry.cn-shanghai.aliyuncs.com/labring/minio-operator:v4.5.5
```
