# k8s_nfs

手工搭建nfs服务器,作为k8s的默认sc使用  
ubuntu20.04.10  
k8s v1.24.6

# 挂盘

```
使用lvm方式挂载/data “推荐，有利于后期扩容”
lsblk
apt install -y lvm2
pvcreate /dev/vdb
vgcreate datavg /dev/vdb
lvcreate -l 100%FREE -n data datavg
mkfs.ext4 /dev/datavg/data
mkdir -p /data
mount /dev/datavg/data /data
echo "/dev/datavg/data /data ext4 defaults 0 0" >> /etc/fstab
df -h /data

扩容 （将vdc-2T扩容到/data）
pvcreate /dev/vdc
vgextend datavg /dev/vdc
lvextend -l +100%FREE /dev/datavg/data
resize2fs /dev/datavg/data
df -h /data

pvcreate /dev/vdd
vgextend datavg /dev/vdd
lvextend -l +100%FREE /dev/datavg/data
resize2fs /dev/datavg/data
df -h /data
```

给服务器整块4T的盘并挂到/data, 直接格式化ext4并挂载，不推荐（不利于后期扩容）

```
mkdir -p /data
检查是否被挂载
lsblk -f | grep vdb
mkfs.ext4 /dev/vdb
mount /dev/vdb /data
blkid /dev/vdb
echo 'UUID="080c6954-160e-499f-8a59-9419cab969b6" /data ext4 defaults,nofail 0 2' | tee -a /etc/fstab
mkdir -p /data/data_nfs
mkdir -p /data/igo
echo '4test' | tee -a /data/igo/test.txt
```

## 搭建nfs-kernel-server

```
apt update
apt install -y nfs-kernel-server
mkdir -p /data/data_nfs
chown nobody:nogroup /data/data_nfs
chmod 777 /data/data_nfs
echo '/data/data_nfs 10.201.81.0/24(rw,sync,no_subtree_check,no_root_squash)' | tee -a /etc/exports
下一行为放行所有，无限制，通常应该用上一行
echo '/data/data_nfs  *(rw,sync,no_subtree_check,no_root_squash)' | tee -a /etc/exports

生效配置
exportfs -ra
systemctl enable nfs-kernel-server
systemctl restart nfs-kernel-server

客户端挂载
mount 服务端IP:/data/data_nfs /本地挂载目录

ansible all -m apt -a "name=nfs-common state=present update_cache=yes" --become
ansible all -m shell -a "dpkg -l | grep -q nfs-common && echo '已安装' || echo '未安装'" --become

apt install -y nfs-common
showmount -e 10.201.81.60
mkdir -p /igo/mnt
mount -t nfs 10.201.81.60:/data/data_nfs /igo/mnt
mount |grep nfs
echo '51test' | tee -a /igo/mnt/51test.txt
cat /igo/mnt/51test.txt
umount /igo/mnt
```

## 给k8s挂载sc使用

```
sealos 方式：
sealos run registry.cn-shanghai.aliyuncs.com/labring/nfs-subdir-external-provisioner:v4.0.18 \
-e HELM_OPTS="--set nfs.server=10.80.238.60 --set nfs.path=/data/data_nfs --set storageClass.defaultClass=true --set storageClass.reclaimPolicy=Delete --set storageClass.archiveOnDelete=true"
 验证
kubectl get deploy nfs-subdir-external-provisioner -n nfs-provisioner -o jsonpath='{.spec.template.spec.containers[0].env}'
kg sc nfs-client -o yaml |grep -i archive
```

```
helm 方式：
1. 所有节点需要安装nfs-common

2.
安装default-SC
helm install nfs-client ./nfs-subdir-external-provisioner-4.0.18.tgz \
  --set nfs.server=10.201.81.60 \
  --set nfs.path=/data/data_nfs \
  --set storageClass.name=nfs-client \
  --set storageClass.defaultClass=true \
  --set storageClass.reclaimPolicy=Delete \
  --set storageClass.archiveOnDelete=true
卸载：
helm uninstall nfs-client -n default

验证：
kubectl get deploy nfs-client-nfs-subdir-external-provisioner -n default -o jsonpath='{.spec.template.spec.containers[0].env}'

根据需要可以再添加令一个SC-retain
# cat retainSC.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: nfs-client-retain
  # 注意：不要添加 storageclass.kubernetes.io/is-default-class: "true"，避免覆盖默认SC
  annotations:
    meta.helm.sh/release-name: nfs-client
    meta.helm.sh/release-namespace: default
  labels:
    app: nfs-subdir-external-provisioner
    app.kubernetes.io/managed-by: Helm
    chart: nfs-subdir-external-provisioner-4.0.18
    heritage: Helm
    release: nfs-client
provisioner: cluster.local/nfs-client-nfs-subdir-external-provisioner
parameters:
allowVolumeExpansion: true
reclaimPolicy: Retain  # 核心修改：回收策略改为 Retain
volumeBindingMode: Immediate
```
