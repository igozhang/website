# ceph_crud常规操作

![](doc-pic/ceph_install_cephadm_1.png)

CRUD基本操作手册

```
env
CentOS Linux release 8.2.2004
ceph version 15.2.15 octopus (stable)

增
添加主机
ssh-copy-id -f -i /etc/ceph/ceph.pub root@igo-ceph-h2
ceph orch host add igo-ceph-h2
添加标签
ceph orch host label add igo-ceph-mon1 mon
ceph orch host label add igo-ceph-mon1 mds
ceph orch host label add igo-ceph-osd1 osd 

添加osd节点
ceph orch apply osd --all-available-devices
手动指定添加特定osd主机
ceph orch apply osd --all-available-devices --unmanaged=true
ceph orch daemon add osd node1:/dev/sdb

添加mon节点
ceph orch apply mon igo-ceph-mon1,igo-ceph-mon2,igo-ceph-mon3
ceph orch apply mon label:mon

添加mds节点
ceph osd pool create igo-cephfs-data
ceph osd pool create igo-cephfs-matadata
ceph fs new igo-cephfs igo-cephfs-matadata igo-cephfs-data
ceph orch apply mds igo-cephfs --placement=3
ceph orch apply mds igo-cephfs --placement=“3 igo-ceph-mon1 igo-ceph-mon2 igo-ceph-mon3”
ceph orch apply mds igo-cephfs --placement="label:mds"

添加rbd块设备
ceph osd pool create igo-rbd-pool-1
rbd  pool init igo-rbd-pool-1
rbd create --size 2G igo-rbd-pool-1/mysql

添加权限
ceph auth get-or-create client.igo mon 'profile rbd' osd 'profile rbd pool=runtime_pool,profile rbd pool=data_pool'

删
rbd rm igo-rbd-1/igo-image-2
ceph fs rm igo-cephfs --yes-i-really-mean-it
删pool跑路的正确姿势
ceph tell mon.* injectargs '--mon_allow_pool_delete=true'
ceph osd pool delete igo-rbd-pool-1 igo-rbd-pool-1 --yes-i-really-really-mean-it

# 区分集群网络和数据网络
ceph config set mon public_network 10.11.0.0/16
ceph config set mon cluster_network 192.168.10.0/24
# cephadm删除集群
cephadm rm-cluster --force --fsid a52f3112-c99a-11eb-8a13-525400c20cc3
# 删除节点服务
ceph orch daemon rm mgr.sz-rick-test-1.rsdyew
# 重新部署服务
ceph orch daemon redeploy mon.es-test-2 [images]

改
CIDR
ceph config set mon public_network 10.1.2.0/24

默认mon数量（5.default->3）
ceph orch apply mon 3
池副本数
ceph osd pool set igo-rbd-1 size 2
ceph osd pool set-quota igo-rbd-pool-1 max_bytes 2097152
块设备扩缩容
rbd resize igo-rbd-1/igo-image-1 --size 2G
rbd resize  igo-rbd-1/igo-image-1 --allow-shrink  --size 1G

打标签
ceph orch host label add igo-cacher-05 mon

查
集群状态
ceph -s
ceph health detail
ceph status
ceph -v

查日志
ceph log last cephadm
ceph -W cephadm

ceph orch device ls

查询fsid集群信息
ceph mon dump

osd块设备查询
ceph osd lspools
ceph osd pool get igo-rbd-1  size
ceph osd pool get-quota igo-rbd-1
rbd ls igo-rbd-pool-1
rbd info igo-rbd-pool-1/mysql

查看空间使用情况
rados df
ceph  df

组件运行状态
ceph orch ps
ceph orch ls
ceph orch ps --daemon-type mds

容器运行状态
cephadm ls

纳管节点
cehp orch host ls

ceph mds stat
ceph fs status
ceph fs dump

权限查询
ceph auth ls
ceph auth get client.admin
ceph auth get client.igo
```
