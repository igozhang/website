# docker

```
env
CentOS Linux release 7.6.1810 (Core)
Docker Client version 20.10.13, build a224086
Docker Engine 19.03.15
Docker Compose version v2.2.2
```

```
容器的两个关键概念
Namespace 和 Cgroups
Namespace 和 Cgroups 可以让程序在一个资源可控的独立（隔离）环境中运行
Namespace 是一种隔离机制,让这些容器之间不能访问彼此的资源,主要包括PID Namespace,Network Namespace,Mount Namespace;(Linux 中所有的 Namespace：cgroup/ipc/network/mount/pid/time/user/uts)
Cgroups （Control Groups）了。Cgroups 可以对指定的进程做各种计算机资源的限制，比如限制 CPU 的使用率，内存使用量，IO 设备的流量等等,避免一个进程把所有资源耗尽;
Cgroups 通过不同的子系统限制了不同的资源，每个子系统限制一种资源。每个子系统限制资源的方式都是类似的，就是把相关的一组进程分配到一个控制组里，然后通过树结构进行管理，每个控制组都设有自己的资源控制参数;
https://man7.org/linux/man-pages/man7/cgroups.7.html
几种比较常用的 Cgroups 子系统：CPU 子系统，用来限制一个控制组（一组进程，你可以理解为一个容器里所有的进程）可使用的最大 CPU。memory 子系统，用来限制一个控制组最大的内存使用量。pids 子系统，用来限制一个控制组里最多可以运行多少个进程。cpuset 子系统， 这个子系统来限制一个控制组里的进程可以在哪几个物理 CPU 上运行。
```

```
安装:
删除旧版本:
yum remove docker \
                  docker-client \
                  docker-client-latest \
                  docker-common \
                  docker-latest \
                  docker-latest-logrotate \
                  docker-logrotate \
                  docker-engine

安装
腾讯:
yum install -y yum-utils
yum-config-manager --add-repo \
https://download.docker.com/linux/centos/docker-ce.repo
yum list docker-ce --showduplicates | sort -r
yum -y install docker-ce-19.03.15-3.el7
systemctl start docker
docker run hello-world

或者
阿里:
yum -y install yum-utils
yum-config-manager –add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
yum -y install docker-ce-19.03.15-3.el7
```

```
Docker镜像源加速
腾讯:
vim /etc/docker/daemon.json
{
   "registry-mirrors": [
       "https://mirror.ccs.tencentyun.com"
  ]
}
阿里:
cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": ["https://59xo2v7a.mirror.aliyuncs.com"],
  "exec-opts": ["native.cgroupdriver=systemd"] ##这一行主要是解决K8S文件驱动不一致导致镜像无法启动问题，也可以通过 在/etc/systemd/system/kubelet.service.d/10-kubeadm.conf添加一行--cgroup-driver=systemd解决；
}
EOF
```

```
docker-compose

curl -L "https://github.com/docker/compose/releases/download/v2.2.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
docker-compose --version
```
