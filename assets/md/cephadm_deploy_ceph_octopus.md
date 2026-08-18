# cephadm部署ceph,Octopus

Cephadm installs and manages a Ceph cluster using containers and systemd, with tight integration with the CLI and dashboard GUI.

- cephadm only supports Octopus and newer releases.
- cephadm is fully integrated with the new orchestration API and fully supports the new CLI and dashboard features to manage cluster deployment.
- cephadm requires container support (podman or docker) and Python 3.

https://ceph.io  
https://docs.ceph.com/

```
env
CentOS Linux release 8.2.2004 (Core)

REQUIREMENTS
a. python3
b. systemd
c. podman/docker二选一，我这里选的是podman
d. Time synchronization
e. lvm2

hosts配置
cat /etc/hosts
192.168.3.201 igo-ceph-mon1
192.168.3.202 igo-ceph-mon2
192.168.3.203 igo-ceph-mon3
192.168.3.206 igo-ceph-osd1
192.168.3.207 igo-ceph-osd2

配置podman国内源
mkdir -p ~/.config/containers/
tee >~/.config/containers/registries.conf <<EOF
unqualified-search-registries = ["docker.io"]
[[registry]]
prefix = "docker.io"
location = "59xo2v7a.mirror.aliyuncs.com"
EOF

1.
通过standalone脚本安装cephadm （也可以通过apt,dnf,zypper等工具安装）
curl --silent --remote-name --location https://github.com/ceph/ceph/raw/octopus/src/cephadm/cephadm
chmod +x cephadm
./cephadm add-repo --release octopus
./cephadm add-repo --version 15.2.11  #可以指定版本
./cephadm install
验证
[root@igo-ceph-mon1 ~]# which cephadm
/usr/sbin/cephadm

2.
部署集群
2.1
引导第一个mon节点
cephadm bootstrap --mon-ip 192.168.3.201
This command will:
Create a monitor and manager daemon for the new cluster on the local host.
Generate a new SSH key for the Ceph cluster and add it to the root user’s /root/.ssh/authorized_keys file.
Write a copy of the public key to /etc/ceph/ceph.pub.
Write a minimal configuration file to /etc/ceph/ceph.conf. This file is needed to communicate with the new cluster.
Write a copy of the client.admin administrative (privileged!) secret key to /etc/ceph/ceph.client.admin.keyring.
Add the _admin label to the bootstrap host. By default, any host with this label will (also) get a copy of /etc/ceph/ceph.conf and /etc/ceph/ceph.client.admin.keyring.

success_output
***************************************手动分割
Ceph Dashboard is now available at:

	     URL: https://igo-ceph-mon1:8443/
	    User: admin
	Password: b9hjymo328

You can access the Ceph CLI with:

	sudo /usr/sbin/cephadm shell --fsid 81f6add6-4534-11ec-99e1-000c298b72d6 -c /etc/ceph/ceph.conf -k /etc/ceph/ceph.client.admin.keyring

Please consider enabling telemetry to help improve Ceph:

	ceph telemetry on

For more information see:

	https://docs.ceph.com/docs/master/mgr/telemetry/

Bootstrap complete.
[root@igo-ceph-mon1 ~]#
*****************************************手动分割

已经可以看到章鱼哥了
![](doc-pic/ceph_install_cephadm_1.png)

可以使用cephadm shell 了
cephadm shell -- ceph -s
cephadm install ceph-common
ceph -s
ceph -v
ceph status

2.2
其他主机加入集群
先分发公钥
ceph cephadm get-pub-key > ~/ceph.pub
ssh-copy-id -f -i ~/ceph.pub root@igo-ceph-mon2
ssh-copy-id -f -i ~/ceph.pub root@igo-ceph-mon3
ssh-copy-id -f -i ~/ceph.pub root@igo-ceph-osd1
ssh-copy-id -f -i ~/ceph.pub root@igo-ceph-osd2
再添加节点
ceph orch host add igo-ceph-mon2
ceph orch host add igo-ceph-mon3
ceph orch host add igo-ceph-osd1
ceph orch host add igo-ceph-osd2

指定mon节点
mon节点依赖正确的子网配置会随着集群数量的变化自动伸缩，需要手动指定mon节点首先需要关闭自动mon_daemon部署
ceph orch apply mon --unmanaged
ceph orch apply mon  igo-ceph-mon1,igo-ceph-mon2,igo-ceph-mon3
ceph orch daemon add mon newhost1:10.1.2.123
ceph orch daemon add mon newhost2:10.1.2.0/24

添加osd
To add storage to the cluster, either tell Ceph to consume any available and unused device
#ceph orch apply osd --all-available-devices

[root@igo-ceph-mon1 ~]# ceph -s
  cluster:
    id:     413469b6-46fe-11ec-ad54-000c29c10c35
    health: HEALTH_OK
 
  services:
    mon: 3 daemons, quorum igo-ceph-mon1,igo-ceph-mon3,igo-ceph-mon2 (age 93m)
    mgr: igo-ceph-mon1.mkciav(active, since 15h), standbys: igo-ceph-osd1.gqvolj
    osd: 3 osds: 3 up (since 2m), 3 in (since 13h)
 
  data:
    pools:   1 pools, 1 pgs
    objects: 0 objects, 0 B
    usage:   3.0 GiB used, 27 GiB / 30 GiB avail
    pgs:     1 active+clean
```
