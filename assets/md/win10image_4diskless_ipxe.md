# Win10无盘镜像制作

```
Win10_iPXE镜像制作

Boot to a virtual hard disk
https://docs.microsoft.com/en-us/windows-hardware/manufacture/desktop/boot-to-vhd--native-boot--add-a-virtual-hard-disk-to-the-boot-menu?view=windows-11

msinfo32  查bios是uefi还是legacy
winver    查win版本
msconfig  可以更改启动项页面选择时间
高级系统设置 取消勾选“显示时间设置为你需要等待的时间”

蓝屏代码
蓝屏代码 inaccessible boot device  
判断和网卡驱动强相关，找不到ipxe可用网卡（驱动）

蓝屏代码 driver irql not less or equal

虚拟机镜像制作
使用vmware虚拟机制作虚拟机镜像（成功）
win10_2004_uefi 虚拟机可用
1.启动盘预格式化两个分区
2.禁用虚拟内存
3.直接使用vmdk转qcow
4.删除sanboot的 --no-describe 参数

物理机镜像制作
禁用虚拟内存后直接ccboot打包可用（成功）
120802_2004_efi_120GAll_nomem_reltek_ccboot.VHDX
疑似和安装ccboot安装时候自动弹出的网卡驱动有关，默认弹出reltek就可以
inaccessible boot device
这个在物理机环境一般安装网卡驱动可以解决

处理问题的一些方法

去掉sanboot 的--no-describe 参数

禁用虚拟内存

注册表Ndi
https://social.technet.microsoft.com/Forums/en-US/4147d300-2de7-4fa3-ad6f-1f04b22b1af3/windows-iscsi-boot-issue-on-second-interface?forum=winserverfiles
2004_legacy_VM没有解决,待确认
1.Boot from the Windows installation media.
2.At the first screen select ‘Next’.
3.At the next screen select ‘Repair your computer’.
4.At the ‘System Recovery Options’ screen, ensure the correct operating system is highlighted and select ‘Next’.
5.At the ‘Choose a recovery tool’ screen, select ‘Command Prompt’.
6.At the command prompt, type ‘regedit’ and press enter.
7.From the registry editor, highlight HKEY_LOCAL_MACHINE, then from the ‘File’ menu select ‘Load Hive’.
8.From the ‘Load Hive’ dialog, browse to the following directory on the drive where Windows is installed: %windir%\system32\config
9.Select the file named SYSTEM and click ‘Open’.
10.Provide a key name of your choice (e.g. temp).
11.Navigate to the following registry location in the hive that was loaded in step 10.
HKEY_LOCAL_MACHINE\<name from step 10>\ControlSet001\Control\ Network\{4d36e974-e325-11ce-bfc1-08002be10318}\{B70D6460-3635-4D42-B866-B8AB1A24454C}\Ndi
12.Double click the FilterRunType value in the right hand pane, and change the value to 2.
13.Highlight the registry hive loaded in step 10, and then from the ‘File’ menu select ‘Unload Hive’.
14.Close registry editor and restart the machine.
Once Windows is booted, open the registry editor again and change the FilterRunType value, from step 11 above, back to a value of 1. Before rebooting, ensure that the NDIS LWF is unbound from all iSCSI booted adapters.

uefi分区
shift+f10
********************手动分割
diskpart
list disk
select disk 0
convert gpt
create partition efi size=100
format quick fs=fat32
create partition primary
exit
exit
```
