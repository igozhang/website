# kvm_crud

# KVM运维

https://www.linux-kvm.org/page/Main\_Page  
https://www.redhat.com/zh/topics/virtualization/what-is-KVM  
https://www.qemu.org/  
基于内核的虚拟机（KVM）是一种内建于 Linux® 的开源虚拟化技术。具体而言，KVM 可帮助您将 Linux转变为虚拟机监控程序，使主机计算机能够运行多个隔离的虚拟环境，即虚拟客户机或虚拟机（VM）。  
KVM 是 Linux 的一部分。Linux 2.6.20 或更新版本包括 KVM。KVM 于 2006 年首次公布，并在一年后合并到主流 Linux 内核版本中。由于 KVM 属于现有的 Linux 代码，因此它能立即享受每一项新的 Linux 功能、修复和发展，无需进行额外工程。

[增](#add) [删](#del) [改](#mod) [查](#query)  
[命令整理](#cli)

## 增加

### 增加虚拟机

1. virt-install  
   virt-install 命令可以用来创建和安装虚拟机。以下是一个示例命令：

```
virt-install \
--name=igoVM \
--ram=2048 \
--vcpus=2 \
--disk path=/var/lib/libvirt/images/igoVM.qcow2,size=20 \
--os-type=linux \
--os-variant=rhel7 \
--network bridge=virbr0 \
--graphics none \
--console pty,target_type=serial \
--location /var/iso/CentOS-7-x86_64-Minimal-1611.iso \
--extra-args 'console=ttyS0,115200n8 serial'
```

这个命令会创建一个名为 igoVM 的虚拟机，分配 2048MB 内存和 2 个 vCPU，使用 qcow2 格式的 20GB 硬盘，安装 CentOS 7 操作系统，使用 virbr0 网络接口，关闭图形界面，使用串行控制台，并从 ISO 镜像文件进行安装。

2. virsh create  
   `virsh create igoVM.xml`

# 使用igoVM.xml创建虚拟机并立即开机

3. virsh define

```
virsh define igoVM.xml
#使用igoVM.xml创建虚拟机但不开机
virsh start igoVM
```

## 删除

### 删除虚拟机

```
virsh destroy igoVM
virsh undefine igoVM
rm /var/lib/libvirt/images/igoVM.qcow2
```

### 删除虚拟磁盘

```
virsh detach-disk <虚拟机名称> <磁盘设备名称> --persistent
rm -rf <虚拟磁盘文件>
```

其中，`virsh detach-disk` 命令用于从虚拟机上分离磁盘，`rm` 命令用于删除虚拟磁盘文件。

### 删除虚拟网络

```
virsh net-destroy <虚拟网络名称>
virsh net-undefine <虚拟网络名称>
其中，`virsh net-destroy` 命令用于关闭虚拟网络，`virsh net-undefine` 命令用于删除虚拟网络定义。
```

### 删除虚拟存储池

```
virsh pool-destroy <虚拟存储池名称>
virsh pool-undefine <虚拟存储池名称>
rm -rf <虚拟存储池目录>
其中，`virsh pool-destroy` 命令用于关闭虚拟存储池，`virsh pool-undefine` 命令用于删除虚拟存储池定义，`rm` 命令用于删除虚拟存储池目录及其下所有内容。
```

## 修改

### 修改虚拟机

```
virsh edit igoVM
virt-xml igoVM --edit --memory 4096 --vcpus 4
```

https://manpages.ubuntu.com/manpages/xenial/man1/virt-xml.1.html

### 虚拟机磁盘扩容

```
qemu-img info /var/lib/libvirt/images/igoVM.qcow2
这个命令会显示虚拟磁盘的大小、格式、实际大小等信息。
qemu-img resize /var/lib/libvirt/images/igoVM.qcow2 +30G
qemu-img check igoVM.qcow2 #检查磁盘信息
qemu-img convert igo-vmware.vmdk igo-kvm.qcow2

qemu-img create -f qcow2 igo.qcow2 4GB
or
qemu-img create -f qcow2 -o size=4GB igo.qcow2
virsh attach-disk igoVM igo.qcow2 vdb --persistent
```

## 查询

### 查询虚拟机

```
virsh list --all
virt-top --all --delay 5
#5s刷新一次虚拟机资源使用情况
virsh dominfo <虚拟机名称>
virsh vcpuinfo <虚拟机名称>
virsh domblklist <虚拟机名称>
virsh domiflist <虚拟机名称>
#磁盘
qemu-img info <虚拟磁盘文件>
#网络
virsh net-list --all
virsh net-info <虚拟网络名称>
#存储池
virsh pool-list --all
virsh pool-info <虚拟存储池名称>
#查虚拟机日志
virsh console igoVm
tail -f /var/log/libvirt/qemu/igoVm.log
```

## 命令整理

### qemu-img

qemu-img snapshot [-l | -a snapshot | -c snapshot | -d snapshot] filename

### virsh

https://www.linuxcool.com/virsh  
virsh dumpxml igoVm >/tmp/igoVm.xml

virsh list --all: 列出所有虚拟机的状态。  
virsh start igoVM: 启动名为 igoVM 的虚拟机。  
virsh shutdown igoVM: 关闭名为 igoVM 的虚拟机。  
virsh destroy igoVM: 强制关闭名为 igoVM 的虚拟机。  
virsh reboot igoVM: 重启名为 igoVM 的虚拟机。  
virsh suspend igoVM: 暂停名为 igoVM 的虚拟机。  
virsh resume igoVM: 恢复名为 igoVM 的虚拟机运行。  
virsh autostart igoVM: 设置名为 igoVM 的虚拟机在宿主机启动时自动启动。  
virsh undefine igoVM: 删除名为 igoVM 的虚拟机的定义。  
virsh edit igoVM: 编辑名为 igoVM 的虚拟机的 XML 配置文件。  
virsh dominfo igoVM: 显示名为 igoVM 的虚拟机的详细信息。  
virsh vcpuinfo igoVM: 显示名为 igoVM 的虚拟机的 vCPU 信息。  
virsh domblklist igoVM: 显示名为 igoVM 的虚拟机的磁盘信息。  
virsh domiflist igoVM: 显示名为 igoVM 的虚拟机的网络接口信息。  
virsh attach-disk igoVM /var/lib/libvirt/images/new\_disk.qcow2 vdb --persistent: 将名为 new\_disk.qcow2 的虚拟磁盘附加到名为 igoVM 的虚拟机上，并分配设备名称为 vdb。  
virsh detach-disk igoVM vdb --persistent: 从名为 igoVM 的虚拟机上分离设备名称为 vdb 的虚拟磁盘。  
virsh setvcpus igoVM 4 --maximum: 将名为 igoVM 的虚拟机的 CPU 核心数设置为 4，同时设置最大可用 CPU 核心数。  
virsh setmem igoVM 8192 --maximum: 将名为 igoVM 的虚拟机的内存大小设置为 8GB，同时设置最大可用内存大小。  
qemu-img create -f qcow2 /var/lib/libvirt/images/new\_disk.qcow2 50G: 创建一个大小为 50GB 的 qcow2 格式的虚拟磁盘，并保存在 /var/lib/libvirt/images 目录下。  
virt-install --name=igoVM --ram=2048 --vcpus=2 --disk path=/var/lib/libvirt/images/igoVM.qcow2,size=20 --os-type=linux --os-variant=rhel7 --network bridge=virbr0 --graphics none --console pty,target\_type=serial --location /var/iso/CentOS-7-x86\_64-Minimal-1611.iso --extra-args 'console=ttyS0,115200n8 serial': 使用 virt-install 命令创建一个名为 igoVM 的虚拟机，分配 2048MB 内存和 2 个 vCPU，使用 qcow2 格式的 20GB 硬盘，安装 CentOS 7 操作系统，使用 virbr0 网络接口，关闭图形界面，并将控制台输出重定向到串口。
