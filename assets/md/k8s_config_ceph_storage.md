# k8s配置ceph存储

```
cephfs
步骤：
1.安装ceph-common
2.ceph创建mds
3.获取admin.keyring
3.k8s_node挂载目录

1.
wget -q -O- 'https://download.ceph.com/keys/release.asc' | sudo apt-key add -
echo 'deb http://download.ceph.com/debian-15.2.11/ bionic main' > /etc/apt/sources.list.d/ceph15.list
apt update
apt install ceph-common
也可以
cephadm install ceph-common

2.
ceph osd pool create cephfs_data 64 64
ceph osd pool create cephfs_metadata 64 64
ceph fs new cephfs cephfs_metadata cephfs_data
ceph fs ls
ceph orch apply mds cephfs --placement="3 node1 node2 node3"
ceph orch apply mds cephfs --placement="label:mds"

3.
ceph auth get client.admin
key = AQC3l7lhBJaUORAAO2kySb6atOlPMkO/SCfArw==

4.
mkdir /igocephfs

vim /etc/fstab
172.17.35.31:6789,172.17.35.32:6789,172.17.35.33:6789:/   /export    ceph    name=admin,secret=AQC3l7lhBJaUORAAO2kySb6atOlPMkO/SCfArw==,_netdev,noatime 0 2

mount -a
```

```
ceph_csi,ceph_osd

https://docs.ceph.com/en/latest/rbd/rbd-kubernetes/

ceph端

CREATE A POOL
ceph osd pool create kubernetes
rbd pool init kubernetes

CONFIGURE CEPH-CSI
SETUP CEPH CLIENT AUTHENTICATION
创建ceph账号，我直接用admin
ceph auth get-or-create client.kubernetes mon 'profile rbd' osd 'profile rbd pool=kubernetes' mgr 'profile rbd pool=kubernetes'
[client.kubernetes]
    key = AQD9o0Fd6hQRChAAt7fMaSZXduT3NWEqylNpmg==

ceph auth get client.admin

fsid
ceph mon dump
fsid b9127830-b0cc-4e34-aa47-9d1a2e9949a8

k8s端

GENERATE CEPH-CSI CONFIGMAP
mkdir -p /data/csi && cd /data/csi

cat <<EOF > csi-config-map.yaml
---
apiVersion: v1
kind: ConfigMap
data:
  config.json: |-
    [
      {
        "clusterID": "b9127830-b0cc-4e34-aa47-9d1a2e9949a8",
        "monitors": [
          "192.168.1.1:6789",
          "192.168.1.2:6789",
          "192.168.1.3:6789"
        ]
      }
    ]
metadata:
  name: ceph-csi-config
EOF
kubectl apply -f csi-config-map.yaml

Recent versions of ceph-csi also require an additional ConfigMap object to define Key Management Service (KMS) provider details.
cat <<EOF > csi-kms-config-map.yaml
---
apiVersion: v1
kind: ConfigMap
data:
  config.json: |-
    {}
metadata:
  name: ceph-csi-encryption-kms-config
EOF
kubectl apply -f csi-kms-config-map.yaml

Recent versions of ceph-csi also require yet another ConfigMap object to define Ceph configuration to add to ceph.conf file inside CSI containers:
cat <<EOF > ceph-config-map.yaml
---
apiVersion: v1
kind: ConfigMap
data:
  ceph.conf: |
    [global]
    auth_cluster_required = cephx
    auth_service_required = cephx
    auth_client_required = cephx
  # keyring is a required key and its value should be empty
  keyring: |
metadata:
  name: ceph-config
EOF
kubectl apply -f ceph-config-map.yaml

GENERATE CEPH-CSI CEPHX SECRET
cat <<EOF > csi-rbd-secret.yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: csi-rbd-secret
  namespace: default
stringData:
  userID: kubernetes
  userKey: AQD9o0Fd6hQRChAAt7fMaSZXduT3NWEqylNpmg==
EOF
kubectl apply -f csi-rbd-secret.yaml

CONFIGURE CEPH-CSI PLUGINS
Create the required ServiceAccount and RBAC ClusterRole/ClusterRoleBinding Kubernetes objects. These objects do not necessarily need to be customized for your Kubernetes environment and therefore can be used as-is from the ceph-csi deployment YAMLs:
$ kubectl apply -f https://raw.githubusercontent.com/ceph/ceph-csi/master/deploy/rbd/kubernetes/csi-provisioner-rbac.yaml
$ kubectl apply -f https://raw.githubusercontent.com/ceph/ceph-csi/master/deploy/rbd/kubernetes/csi-nodeplugin-rbac.yaml
Finally, create the ceph-csi provisioner and node plugins. With the possible exception of the ceph-csi container release version, these objects do not necessarily need to be customized for your Kubernetes environment and therefore can be used as-is from the ceph-csi deployment YAMLs:
$ wget https://raw.githubusercontent.com/ceph/ceph-csi/master/deploy/rbd/kubernetes/csi-rbdplugin-provisioner.yaml
$ kubectl apply -f csi-rbdplugin-provisioner.yaml
$ wget https://raw.githubusercontent.com/ceph/ceph-csi/master/deploy/rbd/kubernetes/csi-rbdplugin.yaml
$ kubectl apply -f csi-rbdplugin.yaml

USING CEPH BLOCK DEVICES
CREATE A STORAGECLASS
$ cat <<EOF > csi-rbd-sc.yaml
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
   name: csi-rbd-sc
provisioner: rbd.csi.ceph.com
parameters:
   clusterID: b9127830-b0cc-4e34-aa47-9d1a2e9949a8
   pool: kubernetes
   imageFeatures: layering
   csi.storage.k8s.io/provisioner-secret-name: csi-rbd-secret
   csi.storage.k8s.io/provisioner-secret-namespace: default
   csi.storage.k8s.io/controller-expand-secret-name: csi-rbd-secret
   csi.storage.k8s.io/controller-expand-secret-namespace: default
   csi.storage.k8s.io/node-stage-secret-name: csi-rbd-secret
   csi.storage.k8s.io/node-stage-secret-namespace: default
reclaimPolicy: Delete
allowVolumeExpansion: true
mountOptions:
   - discard
EOF
$ kubectl apply -f csi-rbd-sc.yaml

CREATE A PERSISTENTVOLUMECLAIM
$ cat <<EOF > raw-block-pvc.yaml
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: raw-block-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Block
  resources:
    requests:
      storage: 1Gi
  storageClassName: csi-rbd-sc
EOF
$ kubectl apply -f raw-block-pvc.yaml
```
