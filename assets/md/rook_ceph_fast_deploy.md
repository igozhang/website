# rook_ceph快速部署

rook架构

```
env
CentOS Linux release 8.2.2004 (Core)
k8s v1.19.16
docker_server v19.03.15

官网
https://rook.io/docs/rook/v1.6/ceph-quickstart.html

1.
先单手搭一套k8s集群，其中3个worker4,5,6节点各有一块未使用盘用作ceph_osd
# kubectl get node
NAME        STATUS   ROLES    AGE   VERSION
igo-k8s-1   Ready    master   57m   v1.19.16
igo-k8s-2   Ready    master   53m   v1.19.16
igo-k8s-3   Ready    master   52m   v1.19.16
igo-k8s-4   Ready    <none>   44m   v1.19.16
igo-k8s-5   Ready    <none>   43m   v1.19.16
igo-k8s-6   Ready    <none>   42m   v1.19.16

[root@igo-k8s-6 ~]# lsblk -f
NAME        FSTYPE      LABEL UUID                                   MOUNTPOINT
nvme0n1                                                              
├─nvme0n1p1 ext4              f69288fc-6ee5-4d2a-8aec-55c21ae85789   /boot
└─nvme0n1p2 LVM2_member       yoZnSA-oW1A-ARgy-YhHV-i0KE-TGCN-z1OODa 
  ├─cl-root xfs               b51ebb97-b006-4fa3-adca-4ea193714e79   /
  └─cl-swap swap              d9f94c3e-1853-459a-a57e-a7e312951470   
nvme0n2                                                              
[root@igo-k8s-6 ~]# 

2.快速部署rook
git clone --single-branch --branch v1.6.11 https://github.com/rook/rook.git
cd rook/cluster/examples/kubernetes/ceph
kubectl create -f crds.yaml -f common.yaml -f operator.yaml
kubectl create -f cluster.yaml

3.Done
等待pod_running大概用了27分钟
# kubectl get pod -n rook-ceph
NAME                                                  READY   STATUS      RESTARTS   AGE
csi-cephfsplugin-m8s9b                                3/3     Running     0          41m
csi-cephfsplugin-ncppf                                3/3     Running     0          41m
csi-cephfsplugin-provisioner-78d66674d8-5nn9p         6/6     Running     0          41m
csi-cephfsplugin-provisioner-78d66674d8-z9hvm         6/6     Running     0          41m
csi-cephfsplugin-zdgct                                3/3     Running     0          41m
csi-rbdplugin-2ncbh                                   3/3     Running     0          41m
csi-rbdplugin-5cvs9                                   3/3     Running     0          41m
csi-rbdplugin-bb6p9                                   3/3     Running     0          41m
csi-rbdplugin-provisioner-687cf777ff-4c2k9            6/6     Running     0          41m
csi-rbdplugin-provisioner-687cf777ff-xjpsn            6/6     Running     0          41m
rook-ceph-crashcollector-igo-k8s-4-558b944d7b-8xzgq   1/1     Running     0          39m
rook-ceph-crashcollector-igo-k8s-5-fd9ff5bc9-g2mb2    1/1     Running     0          39m
rook-ceph-crashcollector-igo-k8s-6-cb96564dd-qxh7s    1/1     Running     0          43m
rook-ceph-mds-myfs-a-74c968b8c8-9bcdz                 1/1     Running     0          39m
rook-ceph-mds-myfs-b-54dd4c76d4-6slqr                 1/1     Running     0          39m
rook-ceph-mgr-a-5c9fc7f9f5-jvcbw                      1/1     Running     0          44m
rook-ceph-mon-a-65bc6d6975-bnw5w                      1/1     Running     0          49m
rook-ceph-mon-b-547f9985f6-8bftl                      1/1     Running     0          49m
rook-ceph-mon-c-6767ff847b-gznlp                      1/1     Running     0          46m
rook-ceph-operator-6b78888745-6hscg                   1/1     Running     0          55m
rook-ceph-osd-0-59b5496599-7zrz6                      1/1     Running     0          43m
rook-ceph-osd-1-779f7f9cd8-gwxlx                      1/1     Running     0          43m
rook-ceph-osd-2-76dc87946f-7c9qw                      1/1     Running     0          39m
rook-ceph-osd-prepare-igo-k8s-4-sllrc                 0/1     Completed   0          39m
rook-ceph-osd-prepare-igo-k8s-5-47dz7                 0/1     Completed   0          39m
rook-ceph-osd-prepare-igo-k8s-6-hbslp                 0/1     Completed   0          39m

4.验证
kubectl apply -f toolbox.yaml
kubectl -n rook-ceph exec -it $(kubectl -n rook-ceph get pod -l "app=rook-ceph-tools" -o jsonpath='{.items[0].metadata.name}') bash
[root@rook-ceph-tools-656b876c47-cdrnc /]# ceph -s
  cluster:
    id:     f3d157d0-77d5-4648-96d0-6e7e56552767
    health: HEALTH_OK
 
  services:
    mon: 3 daemons, quorum a,b,c (age 7m)
    mgr: a(active, since 25m)
    osd: 3 osds: 3 up (since 21m), 3 in (since 4h)
 
  data:
    pools:   1 pools, 1 pgs
    objects: 0 objects, 0 B
    usage:   3.0 GiB used, 27 GiB / 30 GiB avail
    pgs:     1 active+cleanwr
	
5.dashboard
kubectl apply -f dashboard-external-http.yaml
查看密码
kubectl -n rook-ceph get secret rook-ceph-dashboard-password -o jsonpath="{['data']['password']}" | base64 --decode && echo
查看端口
[root@igo-k8s-1 ceph]# kubectl -n rook-ceph get svc | grep dashboard
rook-ceph-mgr-dashboard                 ClusterIP   10.105.90.58     <none>        7000/TCP            60m
rook-ceph-mgr-dashboard-external-http   NodePort    10.100.109.227   <none>        7000:31428/TCP      6m47s
可以开心的访问了
http://192.168.3.211:31428/
```

```
tips
失败需要重新部署时，在执行kubectl delete后需要删除所有node节点/var/lib/rook下的文件，并且DD/dev/vdb
rm -rf /var/lib/rook/*
dd if=/dev/zero  of=/dev/vdb bs=1M count=2048

dashboard添加用户
进入ceph-tools容器
echo igo@1234 >user.pass
ceph dashboard ac-user-create igo -i ./user.pass administrator
```
