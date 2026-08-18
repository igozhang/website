# kvm增量镜像

```
env
Ubuntu 20.04.1
virsh 6.0.0
qemu-img 4.2.1

增量镜像可以快速恢复以及占用较小空间
```

```
server_samba基于系统镜像创建增量镜像
qemu-img create -f qcow2 -b os.image vm-19.qcow2

client_kvm挂载使用增量镜像
   <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='/mnt/igo_vm_image/v2/vm-19.qcow2'/>
      <backingStore type='file'>
        <format type='qcow2'/>
        <source file='/mnt/igo_vm_image/v2/os.image'/>
        <backingStore/>
      </backingStore>
      <target dev='vda' bus='virtio'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x07' function='0x0'/>
    </disk>
```
