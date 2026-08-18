# ansible使用

```
env
CentOS Linux release 8.2.2004 (Core)
ansible-2.9.27-1.el8.noarch

yum -y install ansible

tee >~/.ansible.cfg <<EOF
[defaults]
host_key_checking = False
interpreter_python = auto
inventory = /data/hosts
EOF

*********************************手动分割
cat /data/hosts

[hostAll:children]
mgr
osd

[mgr]
igo-ceph-mgr1 ansible_ssh_host=192.168.3.211 ansible_ssh_user=root ansible_ssh_pass=xxx
igo-ceph-mgr2 ansible_ssh_host=192.168.3.212 ansible_ssh_user=root ansible_ssh_pass=xxx
igo-ceph-mgr3 ansible_ssh_host=192.168.3.213 ansible_ssh_user=root ansible_ssh_pass=xxx

[osd]
igo-ceph-osd1 ansible_ssh_host=192.168.3.214 ansible_ssh_user=root ansible_ssh_pass=xxx
igo-ceph-osd2 ansible_ssh_host=192.168.3.215 ansible_ssh_user=root ansible_ssh_pass=xxx
igo-ceph-osd3 ansible_ssh_host=192.168.3.216 ansible_ssh_user=root ansible_ssh_pass=xxx

*******************************手动分割

[root@igo-ceph-mon1 ~]# ansible hostAll -m shell -a 'chronyc makestep' -b -f6
igo-ceph-mon2 | CHANGED | rc=0 >>
200 OK
igo-ceph-osd3 | CHANGED | rc=0 >>
200 OK
igo-ceph-osd1 | CHANGED | rc=0 >>
200 OK
igo-ceph-osd2 | CHANGED | rc=0 >>
200 OK
igo-ceph-mon3 | CHANGED | rc=0 >>
200 OK
igo-ceph-mon1 | CHANGED | rc=0 >>
200 OK
```

```
tips
如果centos7安装ansible
检查下/etc/yum.repos.d/是否有epel.repo，有就删掉

yum -y remove epel-release
rpm -Uvh http://mirrors.ustc.edu.cn/epel/epel-release-latest-7.noarch.rpm
yum -y install epel-release
yum clean all
yum makecache
yum -y install ansible
```
