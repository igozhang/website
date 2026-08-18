# NFS使用

```
env
centos76
Network File System
```

## 3 NFS共享的实现

### 3.1 需求

在虚拟机 server0 上配置NFS服务，完成以下任务：

1. 只读的方式共享目录 /public，只能被 example.com 域中的系统访问
2. 可读写共享目录/protected，能被 example.com 域中的系统访问

然后在虚拟机 desktop0 上访问NFS共享目录

1. 将 server0 的 /public 挂到本地 /mnt/nfsmount
2. 这些文件系统在系统启动时自动挂载

### 3.2 方法

对于普通NFS共享来说：

- 服务端需要运行系统服务 nfs-server.service
- 客户端不需要运行特定的系统服务

配置NFS共享目录的记录格式：

1. 文件夹绝对路径        客户地址1(ro或rw等控制参数) 客户地址2(ro或rw等控制参数) .. ..

### 3.3 步骤

实现此案例需要按照如下步骤进行。

步骤一：在server0上发布NFS共享目录

1）准备需要共享的文件夹

1. [root@server0 ~]# mkdir /public
2. [root@server0 ~]# mkdir /protected

2）建立NFS共享配置

1. [root@server0 ~]# vim /etc/exports
2. /public 172.25.0.0/24(ro)
3. /protected 172.25.0.0/24(rw)

3）启动系统服务nfs-server，并设置开机自启

1. [root@server0 ~]# systemctl restart nfs-server
2. [root@server0 ~]# systemctl enable nfs-server
3. ln -s '/usr/lib/systemd/system/nfs-server.service' '/etc/systemd/system/nfs.target.wants/nfs-server.service'

步骤二：在desktop0上挂载NFS共享目录/public

1）创建挂载点

1. [root@desktop0 ~]# mkdir /mnt/nfsmount

2）列出server0上提供的NFS共享资源

1. [root@desktop0 ~]# showmount -e server0.example.com
2. Export list for server0.example.com:
3. /protected 172.25.0.0/24
4. /public 172.25.0.0/24

3）配置开机挂载server0的NFS共享目录/public

1. [root@desktop0 ~]# vim /etc/fstab
2. .. ..
3. server0.example.com:/public /mnt/nfsmount nfs \_netdev 0 0

4）测试挂载配置

1. [root@desktop0 ~]# mount -a
2. [root@desktop0 ~]# df -hT /mnt/nfsmount/
3. Filesystem Type Size Used Avail Use% Mounted on
4. server0.example.com:/public nfs4 10G 3.2G 6.8G 32% /mnt/nfsmount
