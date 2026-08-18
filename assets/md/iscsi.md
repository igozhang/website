# iSCSI使用

```
env
Centos76

1：发布iSCSI网络磁盘
2：访问iSCSI网络磁盘
```

## 1：发布iSCSI网络磁盘

### 1.1 问题

server0 提供 iSCSI 服务，要求如下：

1. 磁盘名为iqn.2016-02.com.example:server0
2. 服务端口为 3260
3. 使用 iscsi\_store 作其后端卷，其大小为 3GiB
4. 此磁盘服务只能被 desktop0.example.com 访问

方案

首先利用parted或fdisk工具划分分区

安装targetcli服务端软件，执行targetcli命令进行配置

1. 创建后端存储
2. 创建iqn对象
3. 将iqn与后端存储绑定
4. 授权客户端的iqn
5. 指定监听地址及本机端口

### 1.2 步骤

实现此案例需要按照如下步骤进行。

步骤一：准备磁盘空间，划分分区

1）利用parted命令划分/dev/vdb

1. [root@server0 /]# parted /dev/vdb
2. (parted) mktable gpt                      //指定分区模式为gpt
3. (parted) mkpart primary ext4  0  3GiB
4. 警告: The resulting partition is not properly aligned for best performance.
5. 忽略/Ignore/放弃/Cancel? Ignore
6. (parted) print
7. ……
8. Number  Start  End    Size    File system  Name    标志
9. 1      17.4kB  3221MB  3221MB              primary
10. (parted) quit

2）命令行查看分区信息

1. [root@server0 /]# lsblk
2. NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
3. ……
4. └─vdb1        252:17  0    3G  0 part ln -s

3）查看分区设备文件

1. [root@server0 /]# ls /dev/vdb1
2. /dev/vdb1

步骤二：安装与配置iSCSI磁盘

1）安装软件包targetcli

1. [root@server0 ~]# yum  -y  install  targetcli

2）命令行运行targetcli配置iSCSI磁盘

1. [root@server0 /]# targetcli
2. /> backstores/block create iscsi\_store /dev/vdb1    //创建后端存储
3. /> iscsi/ create iqn.2016-02.com.example:server0    //创建iqn对象
4. /> iscsi/iqn.2016-02.com.example:server0/tpg1/luns create /backstores/block/iscsi\_store    //将iqn对象与后端存储绑定
5. /> iscsi/iqn.2016-02.com.example:server0/tpg1/acls create iqn.2016-02.com.example:desktop0  //设置客户端的iqn
6. /> iscsi/iqn.2016-02.com.example:server0/tpg1/portals/ create 0.0.0.0 3260  //开启本地监听的IP地址与端口
7. /> exit

3）启动target服务，并设置为开机自启

1. [root@server0 /]# systemctl restart target
2. [root@server0 /]# systemctl enable target
3. Created symlink from /etc/systemd/system/multi-user.target.wants/target.service to /usr/lib/systemd/system/target.service.

## 2：访问iSCSI网络磁盘

### 2.1

配置 desktop0 使用 iSCSI 磁盘，要求如下：

1. 目标磁盘组为 iqn.2016-02.com.example:server0
2. 此 iSCSI 设备在系统启动的期间自动加载
3. 此 iSCSI 块设备上包含一个大小为 2100MiB 的分区，并将其格式化为 ext4 文件系统
4. 此分区应该挂载到 /mnt/data 上，同时在系统启动的期间自动挂载

方案

首先安装客户端软件iscsi-initiator-utils

1. 设置客户端iqn
2. 重启iscsid服务刷新客户端iqn
3. 发现服务端iSCSI网络磁盘
4. 使用iSCSI共享磁盘空间

### 2.2 步骤

实现此案例需要按照如下步骤进行。

步骤一：安装客户端软件iscsi-initiator-utils

1）利用yum安装软件

1. [root@desktop0 /]# yum  repolist              //生成yum仓库信息缓存
2. [root@desktop0 /]# yum -y install iscsi-initiator-utils.i686  //在生成yum仓库信息缓存后，软件包名可以补全（tab）

步骤二：设置客户端iqn名称

1）修改配置文件

1. [root@desktop0 /]# vim /etc/iscsi/initiatorname.iscsi
2. InitiatorName=iqn.2016-02.com.example:desktop0

2）重启iscsid服务刷新iqn名称

1. [root@desktop0 /]# systemctl restart iscsid
2. Warning: iscsid.service changed on disk. Run 'systemctl daemon-reload' to reload units.
3. [root@desktop0 /]# systemctl daemon-reload
4. [root@desktop0 /]# systemctl restart iscsid
5. [root@desktop0 /]# systemctl enable iscsid
6. Created symlink from /etc/systemd/system/multi-user.target.wants/iscsid.service to /usr/lib/systemd/system/iscsid.service.

步骤三：发现服务端iSCSI网络磁盘

1）执行发现iSCSI网络磁盘命令

1. [root@desktop0 /]# iscsiadm --mode discoverydb --type sendtargets --portal 172.25.0.11 –discover    //此命令较长可以通过man  iscsiadm查看帮助信息
2. 172.25.0.11:3260,1 iqn.2016-02.com.example:server0

2）重启iscsi服务加载iSCSI网络磁盘

1. [root@desktop0 /]# systemctl restart iscsi
2. [root@desktop0 /]# lsblk
3. NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
4. sda            8:0    0    3G  0 disk
5. [root@desktop0 /]# systemctl enable iscsi

步骤四：客户端使用iSCSI网络磁盘

1）划分分区

1. [root@desktop0 /]# parted /dev/sda
2. (parted) mktable gpt
3. (parted) mkpart primary  ext4  0  2100MiB
4. 警告: The resulting partition is not properly aligned for best performance.
5. 忽略/Ignore/放弃/Cancel? Ignore
6. (parted) print
7. (parted) exit
8. [root@desktop0 /]# partprobe /dev/sda        //刷新分区表
9. [root@desktop0 /]# lsblk
10. NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
11. sda            8:0    0    3G  0 disk
12. └─sda1          8:1    0  2.1G  0 part

2）格式化使用分区，利用UUID实现开机自动挂载

1. [root@desktop0 /]# mkfs.ext4 /dev/sda1
2. [root@desktop0 /]# blkid /dev/sda1      #查看分区UUID值
3. /dev/sda1: UUID="402d1304-5781-4349-8c43-ba45eba8b427" TYPE="ext4" PARTLABEL="primary" PARTUUID="2e41f84d-5123-4511-aaca-7e779087579b"
4. [root@desktop0 /]# mkdir  /mnt/data
5. [root@desktop0 /]# vim /etc/fstab
6. UUID="402d1304-5781-4349-8c43-ba45eba8b427"  /mnt/data ext4 defaults,\_netdev 0 0
7. [root@desktop0 /]# mount -a
8. [root@desktop0 /]# df -h
9. 文件系统              容量  已用  可用 已用% 挂载点
10. ……
11. /dev/sda1              2.0G  6.2M  1.9G    1% /mnt/data
