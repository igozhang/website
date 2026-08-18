# Cent7_boot目录修复

步骤1

误删启动目录（验证救援操作）

步骤2

挂载cent7光盘并BIOS光盘引导

步骤3

进入救援模式

步骤4

选Recue a centos system

步骤5

选1 Continue

步骤6

恢复boot命令

步骤7

成功

```
代码简述：
 1. 以光盘启动引导系统 （在进入虚拟机的瞬间按下F2）
 2. chroot /mnt/sysimage
 3. mkdir  /mnt/temp
 4. mount /dev/sr0 /mnt/temp
 5. rpm -ivh /mnt/temp/Packages/kernel-3.10.0-693.e17.x86_64.rpm.rpm --force
 6. ls /boot
 7. grub2-install /dev/sda
 8. grub2-mkconfig -o /boot/grub2/grub.cfg
 9. sync
 10. exit
 11.reboot

也可以从其他同版本机器拷贝一个boot目录恢复，在Centos7,8均测试通过
```
