# kvm显卡直通

```
env
ubuntu18

NvidiaRTX显卡虚拟机直通指南
```

```
1、确认BIOS里开启Vt-d功能。
egrep '(vmx|svm)' /proc/cpuinfo

2、开启iommu。

sed -i 's/GRUB_CMDLINE_LINUX_DEFAULT=.*/GRUB_CMDLINE_LINUX_DEFAULT=\"modprobe.blacklist=nouveau net.ifnames=0 biosdevname=0 intel_iommu=on\ "/g' /etc/default/grub
sudo update-grub
 
echo "options kvm ignore_msrs=1" > /etc/modprobe.d/kvm.conf
3、加载内核模块

printf "vfio \nvfio_iommu_type1 \nvfio_pci \nvfio_virqfd \nvhost-net" >> /etc/initramfs-tools/modules
update-initramfs -u

4、查看显卡信息
lspci -nn | grep NVI
02:00.0 VGA compatible controller [0300]: NVIDIA Corporation GA106 [GeForce RTX 3060] [10de:2503] (rev a1)
02:00.1 Audio device [0403]: NVIDIA Corporation Device [10de:228e] (rev a1)
```

| **输出值** | **含义** | **详细解释** |
| --- | --- | --- |
| 02:00.002:00.1 | 以 ”bus:slot.func“ 格式来唯一标识一个 PCI 功能设备 | 唯一定位一个 PCI 设备的虚拟功能，可以是一个物理设备，也可以是一个多功能设备的功能设备，一个多功能设备可以最多有8个功能。总线号（bus）：  从系统中的256条总线中选择一条，0--255。设备号（slot）：  在一条给定的总线上选择32个设备中的一个。0--31。功能号（func）：  选择多功能设备中的某一个功能，有八种功能，0--7。 PCI规范规定，功能0是必须实现的。 |
| 0300 | PCI 设备类型 | 指 PCI 设备的类型，来自不同厂商的同一类设备的类型码可以是相同的。 |
| 10de | 供应商识别字段（Vendor ID） | 该字段用一标明设备的制造者。一个有效的供应商标识由 PCI SIG 来分配，以保证它的唯一性。Intel 的 ID 为 0x8086，Nvidia 的 ID 为 0x10de |
| 2503 | 设备识别字段（Device ID） | 用以标明特定的设备，具体代码由供应商来分配。本例中表示的是 GPU 图形卡的设备 ID。 |
| a1 | 版本识别字段（Revision ID） | 用来指定一个设备特有的版本识别代码，其值由供应商提供 |

#查看驱动详情  
lspci -vv -s 02:00.0 | grep driver

```
5、显卡添加到直通组
devids=""
devid_list=$(lspci -nn | grep NVIDIA | grep -v GF108 | grep -o \\[....:....\\])
for i in $devid_list
do
devids=$devids,$i
done
vfio_id=$(echo $devids| sed 's/^,//g'|sed 's:\[::g'|sed 's:\]::g')
echo options vfio-pci ids=$vfio_id >> /etc/modprobe.d/local.conf

#重启服务器或者手工卸载显卡设备
virsh nodedev-detach pci_0000_02_00_0
virsh nodedev-detach pci_0000_02_00_1
 
#验证
lspci -vv -s 02:00.0 | grep driver
lspci -vv -s 02:00.1 | grep driver

Kernel driver in use: vfio-pci 即为正常

6、添加显卡到虚拟机
方式一：
通过图形化virt-manager 添加PCI设备。

方式二：
通过xml文件配置

<hostdev mode='subsystem' type='pci' managed='yes'>
<driver name='vfio'/>
<source>
<address domain='0x0000' bus='0x02' slot='0x00' function='0x0'/>
</source>
<alias name='hostdev0'/>
<address type='pci' domain='0x0000' bus='0x00' slot='0x08' function='0x0' multifunction='on'/>
</hostdev>
<hostdev mode='subsystem' type='pci' managed='yes'>
<driver name='vfio'/>
<source>
<address domain='0x0000' bus='0x02' slot='0x00' function='0x1'/>
</source>
</hostdev>
```

```
tips

1、隐藏vm状态：否则显卡驱动安装过程中自检会报错ERROR CODE 43，驱动不允许安装在虚拟机上（windows必须设置）。

<features>
<kvm>
<hidden state='on'/>
</kvm>
</features>

2、编辑VM XML
将<domain type='kvm'>
修改为
<domain type='kvm' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>

3、注入信息
找到</device>节点，在device后添加内容
<qemu:commandline>
<qemu:arg value='-cpu'/>
<qemu:arg value='host,kvm=off,hv_vendor_id=1234567890ab,hv_vapic,hv_time,hv_relaxed,hv_spinlocks=0x1fff,-hypervisor'/>
</qemu:commandline>
```
