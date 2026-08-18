# ceph_rbd块设备

```
env
CentOS Linux release 8.2.2004 (Core)
ceph version 15.2.15

osd_pool_default_pg_num = 128
osd_pool_default_pgp_num = 128

创建块设备pool
ceph osd pool create igo-rbd-01
ceph osd pool ls
  建议用dashboard设置，可选项多而且直观
rbd pool init igo-rbd-01
ceph auth get-or-create client.igo osd 'allow rw  pool=igo-pool-1' -o igo.key

[root@igo-ceph-mon1 ~]# ceph auth get client.igo
exported keyring for client.igo
[client.igo]
	key = AQCS85Rh1EoGGBAAtthfxP7OucicmmyLsPv4mg==
	caps osd = "allow rw  pool=igo-pool-1"

创建块设备image
rbd create --size 1G igo-rbd-1/igo-image-2
# rbd ls igo-rbd-1
igo-image-1
igo-image-2

查看块设备大小
[root@igo-ceph-mon1 ~]# rbd du igo-rbd-1/igo-image-1
NAME         PROVISIONED  USED  
igo-image-1        1 GiB  28 MiB
块设备扩容缩容
rbd resize igo-rbd-1/igo-image-1 --size 2G
rbd resize  igo-rbd-1/igo-image-1 --allow-shrink  --size 1G

本地挂载块设备
rbd info igo-rbd-1/igo-image-1
将image映射为本地设备
rbd map igo-rbd-1/igo-image-1
[root@igo-ceph-mon1 ~]# rbd showmapped
id  pool       namespace  image        snap  device   
0   igo-rbd-1             igo-image-1  -     /dev/rbd0

mkfs.ext4 /dev/rbd0
mkdir -p /mnt/ceph-block-device
mount /dev/rbd0 /mnt/ceph-block-device
touch /mnt/ceph-block-device/igo.txt

删除块设备
rbd unmap igo-rbd-1/igo-image-2
rbd rm igo-rbd-1/igo-image-2

客户机（远端）挂载ceph块设备
1.
安装ceph-common
yum -y install ceph-common
配置ceph-common
scp root@igo-ceph-mon1:/etc/ceph/ceph.client.admin.keyring /etc/ceph
scp root@igo-ceph-mon1:/etc/ceph/ceph.conf /etc/ceph/
ceph -s

rbd ls igo-rbd-1
rbd map igo-rbd-1/igo-image-500
mkfs.ext4 /dev/rbd0
mount /dev/rbd0 /mnt/ceph500
vim /etc/fstab
/dev/rbd0     /mnt/ceph500 ext4 defaults 0 0
echo igo500> /mnt/ceph500/igo.txt

  实现开机自动挂载
1.
自动mapping
vim /etc/ceph/rbdmap
igo-rbd-1/igo-image-500 id=admin,keyring=/etc/ceph/ceph.client.admin.keyring
systemctl enable rbdmap.service
2.
开机自动mount
mount /dev/rbd0 /mnt/ceph500
```

```
补充：
tee >/etc/yum.repos.d/ceph.repo <<EOF
[Ceph]
name=Ceph packages for \$basearch
baseurl=https://mirrors.aliyun.com/ceph/rpm-octopus/el8/\$basearch
enabled=1
gpgcheck=0
type=rpm-md

[Ceph-noarch]
name=Ceph noarch packages
baseurl=https://mirrors.aliyun.com/ceph/rpm-octopus/el8/noarch
enabled=1
gpgcheck=0
type=rpm-md

[ceph-source]
name=Ceph source packages
baseurl=https://mirrors.aliyun.com/ceph/rpm-octopus/el8/SRPMS
enabled=1
gpgcheck=0
type=rpm-md
EOF
```
