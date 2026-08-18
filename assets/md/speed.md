# 源，国内加速配置

```
yum
centos7
curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo
curl -o /etc/yum.repos.d/epel-7.repo https://mirrors.aliyun.com/repo/epel-7.repo

curl -o /etc/yum.repos.d/epel-7.repo http://mirrors.cloud.tencent.com/repo/epel-7.repo
wget -O /etc/yum.repos.d/CentOS-Base.repo http://mirrors.cloud.tencent.com/repo/centos7_base.repo
centos8
curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-8.repo
wget -O /etc/yum.repos.d/CentOS-Base.repo http://mirrors.cloud.tencent.com/repo/centos8_base.repo

apt
/etc/apt/sources.list
替换
http://archive.ubuntu.com/
为
mirrors.aliyun.com
Ubuntu 20.04.3
deb http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse
deb-src http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse
apt-get update  #更新列表
apt-get upgrade  #更新软件

pypi
mkdir -p ~/.pip
tee > ~/.pip/pip.conf <<EOF
[global]
index-url = https://mirrors.aliyun.com/pypi/simple/

[install]
trusted-host=mirrors.aliyun.com
EOF

docker
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
  "exec-opts": ["native.cgroupdriver=systemd"]
}
EOF
```

```
docker pull k8s.gcr.io/xxx:yyy
相当于docker pull gcr.io/google-containers/xxx:yyy
那么使用中科大镜像，应该是这样拉取：
docker pull gcr.mirrors.ustc.edu.cn/google-containers/xxx:yyy
以拉取k8s.gcr.io/addon-resizer:1.8.3为例，如下：
docker pull gcr.mirrors.ustc.edu.cn/google-containers/addon-resizer:1.8.3

docker pull quay.io/xxx/yyy:zzz
docker pull quay.mirrors.ustc.edu.cn/coreos/kube-state-metrics:v1.5.0
exp：
Pulling container image quay.io/ceph/ceph:v15  timeout
podman pull quay.mirrors.ustc.edu.cn/ceph/ceph:v15
podman tag  3437f7bed968 quay.io/ceph/ceph:v14
```

```
GOPROXY
aliyun
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/       // Windows  
export GOPROXY=https://mirrors.aliyun.com/goproxy/          // macOS 或 Linux
七牛云
go env -w GOPROXY=https://goproxy.cn      // Windows  
export GOPROXY=https://goproxy.cn         // macOS 或 Linux

Bash
export GOPROXY=https://proxy.golang.com.cn,direct
# 还可以设置不走 proxy 的私有仓库或组，多个用逗号相隔（可选）
export GOPRIVATE=git.mycompany.com,github.com/my/private

Powershell
$env:GOPROXY = "https://proxy.golang.com.cn,direct"
# 还可以设置不走 proxy 的私有仓库或组，多个用逗号相隔（可选）
$env:GOPRIVATE = "git.mycompany.com,github.com/my/private"
```
