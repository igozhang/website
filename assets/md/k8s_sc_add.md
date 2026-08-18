# k8s基于现有SC添加新存储类

```
env
Ubuntu 18.04.5 LTS
k8s    v1.19.11
rook-ceph

基于现有storageClass_yaml文件添加新的sc
方法
导出现有sc.yaml,改名，删除不必要项，apply，kg sc验证
```

```
步骤
1.查看现有sc
需要一个新rbd_sc，名字叫csi-rbd-sc，拷贝rook-ceph-block的yaml，删改后apply
# kg sc
NAME              PROVISIONER                     RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
rook-ceph-block   rook-ceph.rbd.csi.ceph.com      Delete          Immediate           true                   23h
rook-cephfs       rook-ceph.cephfs.csi.ceph.com   Delete          Immediate           true                   23h

2.导出yaml
time,resource字段删掉，名字改掉就可以
# kg sc rook-ceph-block -o yaml >rook-ceph-block.yaml
    time: "2021-12-30T06:50:04Z"
  name: igo-rbd
  resourceVersion: "52083"

3.k apply -f igo-rbd.yaml
4.验证
# kg sc
NAME              PROVISIONER                     RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
igo-rbd           rook-ceph.rbd.csi.ceph.com      Delete          Immediate           true                   6m5s
rook-ceph-block   rook-ceph.rbd.csi.ceph.com      Delete          Immediate           true                   24h
rook-cephfs       rook-ceph.cephfs.csi.ceph.com   Delete          Immediate           true                   24h
```

```
tips

建议k9s edit 得到的yaml比较干净直观，保存修改一样可以使用

allowVolumeExpansion: true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  creationTimestamp: "2021-12-30T06:50:04Z"
  name: rook-ceph-block
  resourceVersion: "52083"
  selfLink: /apis/storage.k8s.io/v1/storageclasses/rook-ceph-block
  uid: dacdca9e-603b-472d-b509-4f39e8b574b5
parameters:
  clusterID: rook-ceph
  csi.storage.k8s.io/controller-expand-secret-name: rook-csi-rbd-provisioner
  csi.storage.k8s.io/controller-expand-secret-namespace: rook-ceph
  csi.storage.k8s.io/fstype: ext4
  csi.storage.k8s.io/node-stage-secret-name: rook-csi-rbd-node
  csi.storage.k8s.io/node-stage-secret-namespace: rook-ceph
  csi.storage.k8s.io/provisioner-secret-name: rook-csi-rbd-provisioner
  csi.storage.k8s.io/provisioner-secret-namespace: rook-ceph
  imageFeatures: layering
  imageFormat: "2"
  pool: replicapool
provisioner: rook-ceph.rbd.csi.ceph.com
reclaimPolicy: Delete
volumeBindingMode: Immediate
```
