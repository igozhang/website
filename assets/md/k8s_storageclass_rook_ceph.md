# k8s_StorageClass-rook_ceph使用

```
env

CentOS Linux release 8.2.2004 (Core)
k8s v1.19.16
docker_server v19.03.15
rook v1.6.11
```

```
1.创建存储类
kubectl apply -f csi/rbd/storageclass.yaml
验证
ceph_dashboard可以看到replicapool
[root@igo-k8s-1 ceph]# kubectl get sc
NAME              PROVISIONER                  RECLAIMPOLICY
rook-ceph-block   rook-ceph.rbd.csi.ceph.com   Delete

2.搭建一套网站系统
kubectl apply -f kubernetes/mysql.yaml
kubectl apply -f kubernetes/wordpress.yaml
验证
kubectl get pvc
kubectl get pv
```
