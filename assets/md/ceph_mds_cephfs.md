# ceph_mds文件系统

ceph-cephfs-入坑手册,mds,[Multiple File Systems](https://docs.ceph.com/en/pacific/cephfs/multifs/)

```
ceph-cephfs-使用

env
CentOS Linux release 8.2.2004 (Core)

https://docs.ceph.com/en/pacific/cephfs/createfs/
https://access.redhat.com/documentation/zh-cn/red_hat_ceph_storage/4/html/file_system_guide/mounting-the-ceph-file-system-as-a-kernel-client_fs

Client_config
1.
安装ceph-common
yum -y install ceph-common
2.
配置ceph-common
scp root@igo-ceph-mon1:/etc/ceph/ceph.client.admin.keyring /etc/ceph
scp root@igo-ceph-mon1:/etc/ceph/ceph.conf /etc/ceph/
ceph -s

3.
mkdir /mnt/igo-cephfs
mount -t ceph 192.168.3.201:6789,192.168.3.202:6789,192.168.3.203:6789:/ /mnt/igo-cephfs/ -o name=admin

4.
echo helloWorld >/mnt/igo-cephfs/igo-test.txt

5.
[root@igo-test-1 ceph]# df -Th
Filesystem                                                 Type      Size  Used Avail Use% Mounted on
devtmpfs                                                   devtmpfs  379M     0  379M   0% /dev
tmpfs                                                      tmpfs     396M     0  396M   0% /dev/shm
tmpfs                                                      tmpfs     396M  5.7M  391M   2% /run
tmpfs                                                      tmpfs     396M     0  396M   0% /sys/fs/cgroup
/dev/mapper/cl-root                                        xfs        17G  1.9G   16G  12% /
/dev/nvme0n1p1                                             ext4      976M  128M  782M  15% /boot
tmpfs                                                      tmpfs      80M     0   80M   0% /run/user/0
192.168.3.201:6789,192.168.3.202:6789,192.168.3.203:6789:/ ceph      8.5G     0  8.5G   0% /mnt/igo-cephfs

Server_config
1.
ceph osd pool create igo-cephfs-data
ceph osd pool create igo-cephfs-matadata
2.

***enable the file system using the fs new command:
ceph fs new igo-cephfs igo-cephfs-matadata igo-cephfs-data
# ceph fs ls
name: igo-cephfs, metadata pool: igo-cephfs-matadata, data pools: [igo-cephfs-data ]

3.
# ceph mds stat
igo-cephfs:0
启动fs后，会自动拉起mds单节点

ceph orch apply mds igo-cephfs --placement=3
# ceph mds stat
igo-cephfs:1 {0=igo-cephfs.igo-ceph-mon1.xentym=up:active} 2 up:standby

# ceph -s
  cluster:
    id:     413469b6-46fe-11ec-ad54-000c29c10c35
    health: HEALTH_OK
 
  services:
    mon: 3 daemons, quorum igo-ceph-mon1,igo-ceph-mon3,igo-ceph-mon2 (age 2h)
    mgr: igo-ceph-mon1.mkciav(active, since 45h), standbys: igo-ceph-osd1.gqvolj
    mds: igo-cephfs:1 {0=igo-cephfs.igo-ceph-mon1.xentym=up:active} 2 up:standby
    osd: 3 osds: 3 up (since 30h), 3 in (since 43h)
 
  data:
    pools:   4 pools, 97 pgs
    objects: 52 objects, 15 MiB
    usage:   3.0 GiB used, 27 GiB / 30 GiB avail
    pgs:     97 active+clean
```
