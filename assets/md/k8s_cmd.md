# k8s_crud常用命令

```
http://docs.kubernetes.org.cn/683.html

source <(kubectl completion bash)
source <(kubeadm completion bash)
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

alias k='kubectl'
alias ka='kubectl apply'
alias kc='kubectl create'
alias kd='kubectl describe'
alias kg='kubectl get'

kubectl cluster-info
kubectl api-versions

kubectl label nodes k8s-node-02 role=igo-k8s-node2

生成模板文件
kubectl create deployment nginx --image=nginx -o yaml --dry-run > deployment.yaml
kubectl expose deployment nginx --port=80 --type=NodePort -o yaml --dry-run > svc.yaml

kubectl create ns igo
kubectl apply -f pod1.yaml
kubectl create deployment nginx --image=nginx -n igo
kubectl expose deployment nginx --port=80 --type=NodePort
kubectl create deployment nginx-01 -n igo --image=nginx:latest --dry-run=client -o yaml
kubectl create service nodeport nginx-01  --tcp=80:80 --node-port=30080 -n igo --dry-run=client -o yaml
kubectl create deployment nginx --image=nginx

kubectl exec -it nginx-6799fc88d8-s57zt -n igo -- /bin/bash

kubectl get pod -A
kubectl get pod -n igo
kubectl get pod -n igo  -o wide
kubectl get pod nginx-6799fc88d8-s57zt -n igo -o yaml
kubectl describe pod nginx-6799fc88d8-s57zt -n igo
kubectl logs nginx-6799fc88d8-s57zt -n igo

强制删除pod
kubectl delete pod wordpress-mysql-6965fc8cc8-tjhmp --grace-period=0 --force

kubectl scale deployment -n igo nginx --replicas=3

节点重启
kubectl get pod -n igo -o wide
kubectl drain igo-k8s-4 --delete-local-data --ignore-daemonsets
systemctl uncordon igo-k8s-4

修改默认端口范围
/etc/kubernetes/manifests/kube-apiserver.yaml
    - --service-node-port-range=10000-39000
不需要重启，kubernets会自动重启kubelet生效

添加时区cm
kubectl -n igo create cm igo-timezone --from-file=/usr/share/zoneinfo/Asia/Shanghai

回滚实验
kubectl rollout history deployment deploy-myapp -n igo
kubectl rollout undo deployment myapp-deploy   -n igo

灰度发布
kubectl -n igo patch deployment myapp-deploy -p '{"spec": {"strategy":{"rollingUpdate": {"maxSurge": 1, "maxUnavailable": 0}}}}'
kubectl  patch deployment myapp -p '{"spec": {"strategy":{"rollingUpdate": {"maxSurge": 1, "maxUnavailable": 0}}}}'
kubectl  set image deployment myapp myapp=ikubernets/myapp:v3 && kubectl rollout pause deployment myapp

添加 helm repo
helm repo add bitnami https://charts.bitnami.com/bitnami
helm search repo bitnami
helm install my-release bitnami/<chart>

https://github.com/bitnami/charts
helm repo --help
helm fetch bitnami/zookeeper
vim  mysql/values.yaml
helm install u-zook bitnami/zookeeper -f values.yaml -n igo
helm upgrade -n igo mysql bitnami/mysql  -f values.yaml
kubectl -n igo exec -it mysql-0 -- /bin/bash

跑个临时mysql客户端
kubectl run mysql-client --rm --tty -i --restart='Never' --image  docker.io/bitnami/mysql:5.7.34-debian-10-r3 --namespace igo --command -- bash
```
