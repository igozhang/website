# kvm_optimize

```
env
Ubuntu20.04
```

```
1.开启电源性能模式

1.1 内存组双通道
1.2 BIOS中开启电源性能模式（不是所有主板都有这个选项）
1.3 如果有多个CPU,开启NUMA4.确保服务器散热器,机箱盖,cpu硅脂等散热条件正常，温度过高可能导致CPU降频

#!/bin/bash
value=`cat /sys/devices/system/cpu/cpu1/cpufreq/scaling_governor`
if
[ X$value !=
"Xperformance"
]
   
then
         
for
file in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor            
do
echo
"performance"
> $file
done
fi
```

```
2. 安装新内核
apt install linux-image-5.8.0-50-lowlatency

3. 开启hugepages
apt install libhugetlbfs-bin/etc/sysctl.conf
增加以下内核参数vm.nr_hugepages = x 
（x为page数量，x*2M要大于分配给虚拟机的内存之和，如果设置过大会导致系统内存不足卡顿，过小会导致虚拟机无法启动，提示无法分配内存)执行hugeadm --explain命令，会得出kernel.shmmax = xxxvm.hugetlb_shm_group = xxx的建议值
追加到 /etc/sysctl.conf
然后执行
sysctl -pecho always > /sys/kernel/mm/transparent_hugepage/enabled

4. 关闭KSM
编辑 /etc/default/qemu-kvmKVM_HUGEPAGES=1KSM_ENABLED=0service libvirt-bin restartecho 0 >/sys/kernel/mm/ksm/run重启才会生效

5.关闭kvm mrsr
/etc/modprobe.d/kvm.confoptions kvm ignore_msrs=Yoptions kvm report_ignored_msrs=N
修改/etc/default/grubGRUB_CMDLINE_LINUX_DEFAULT 增加 
kvm.ignore_msrs=1GRUB_CMDLINE_LINUX_DEFAULT="modprobe.blacklist=nouveau
net.ifnames=0 biosdevname=0
intel_iommu=on
kvm.ignore_msrs=1
mitigations=off"
执行update-grub重启生效

6. 虚拟机CPU绑定优化
1. emulatorpin和iothread绑定12核CPU,且不绑定在虚拟机第一个核心上，某些游戏（原神）对单cpu核心占用较多

2. 分配给虚拟机的cpu在系统隔离
修改/etc/default/grub,追加
isolcpus=1,2,3,4,5,6,11,24,25,26,27,28,29,30...所有中断在第一个物理cpu的第一个核心，所以一般不隔离该核心，也不绑定到虚拟机；
GRUB_CMDLINE_LINUX_DEFAULT="modprobe.blacklist=nouveau
net.ifnames=0 biosdevname=0
intel_iommu=on
isolcpus=1,2,3,4,5,6,11,24,25,26,27,28,29,30
kvm.ignore_msrs=1 mitigations=off"
执行update-grub重启生效
```

```
windows系统优化
1.激活系系统，未激活状态下系统部分功能受限
2.禁用winsat任务计划
3.关闭系统自动更新
```
