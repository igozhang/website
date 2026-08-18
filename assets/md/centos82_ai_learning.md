# CentOS82配置深度学习环境

```
Env:
Cent82 Nvidia3090GTX 显卡
安装基础包
yum install -y yum-utils \
  device-mapper-persistent-data \
  lvm2
设置稳定仓库
yum-config-manager \
    --add-repo \
    https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo

安装Docker Engine - Community
yum -y install docker-ce docker-ce-cli containerd.io

systemctl start docker
systemctl enable docker

安装显卡驱动
yum -y install pkg-config gcc gcc-c++ kernel-devel kernel-headers dkms epel-release

禁用 nouveau
grub2-editenv - set "$(grub2-editenv - list | grep kernelopts) nouveau.modeset=0"
echo -e "blacklist nouveau\noptions nouveau modeset=0" > /etc/modprobe.d/blacklist.conf
验证： 重启后
lsmod |grep nouveau

yum -y update

# grep nvidia /etc/modprobe.d/* /lib/modprobe.d/*
注释掉这行才能安装
/lib/modprobe.d/dist-blacklist.conf:blacklist nvidiafb

带内核路径
./NVIDIA-Linux-x86_64-470.57.02.run --kernel-source-path=/usr/src/kernels/4.18.0-305.10.2.el8_4.x86_64 -k $(uname -r)

安装cuda
# ./cuda_11.4.1_470.57.02_linux.run
Accept ->install

export PATH=/usr/local/cuda-11.4/bin:$PATH
export LD_LIBRARY_PATH=/usr/local/cuda-11.4/lib64:$LD_LIBRARY_PATH
export CUDA_HOME=/usr/local/cuda

nvcc -V

安装cudnn
tar -xf cudnn-11.4-linux-x64-v8.2.2.26.tgz
cp cuda/include/cudnn*.h /usr/local/cuda-11.4/include
cp cuda/lib64/libcudnn* /usr/local/cuda-11.4/lib64/
chmod a+r /usr/local/cuda-11.4/include/cudnn*.h /usr/local/cuda-11.4/lib64/libcudnn*
验证：
cat /usr/local/cuda/include/cudnn_version.h | grep CUDNN_MAJOR -A 2

给用户添加docker组
useradd igo;
passwd igo
# cat /etc/group | grep docker
docker:x:985:
usermod -aG docker igo
usermod -aG wheel  igo 添加sudo组
newgrp docker
```

```
配置完可以 跑 google_bert 3dmark benchmark 测试性能
```
