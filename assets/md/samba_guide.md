# samba使用

```
env
centos7.6

Samba是在Linux和UNIX系统上实现SMB协议的一个免费软件,可是共享文件和打印机
SMB(Server Messages Block)
```

## 1 ：配置SMB文件夹共享

### 1.1 需求

在虚拟机 server0 上发布两个共享文件夹，具体要求如下：

1. 此服务器必须是 STAFF 工作组的一个成员
2. 发布目录 /common，共享名为 common
3. 发布目录 /devops，共享名为 devops
4. 这两个共享必须是可浏览的，只有 example.com 域内的客户端可以访问
5. 用户 harry 对共享 common 只读，密码是 migwhisk
6. 用户 kenji 对共享 devops 只读，密码是 atenorth
7. 用户 chihiro 对共享 devops 可读写，密码是atenorth

### 1.2 方案

Samba的用途：为多个客户机提供共享使用的文件夹。

Samba服务端：软件包samba、系统服务smb

Samba客户端：软件包samba-client和cifs-utils、客户端工具smbclient

传输协议及端口：TCP 139、TCP 445

Samba服务端配置文件：/etc/samba/smb.conf

Samba共享账号：存在独立的账号数据文件里，必须有同名系统账号（方便给权限）

Samba账号管理工具：

- pdbedit -a 用户名
- pdbedit -L [用户名]
- pdbedit -x 用户名

测试Samba共享资源：

- smbclient -L 服务器地址 【密码为空（直接回车）】
- smbclient -U 用户名 //服务器地址/共享名 【需要密码】

### 1.3 步骤

实现此案例需要按照如下步骤进行。

步骤一：在服务器server0发布Samba共享文件夹

1）安装软件包samba

1. [root@server0 ~]# yum -y install samba
2. .. ..

2）创建共享账号

添加共享账号harry，密码为migwhisk：

1. [root@server0 ~]# useradd harry
2. [root@server0 ~]# pdbedit -a harry                 //根据提示设好密码migwhisk
3. new password:
4. retype new password:

添加共享账号kenji，密码为atenorth：

1. [root@server0 ~]# useradd kenji
2. [root@server0 ~]# pdbedit -a kenji                 //根据提示设好密码atenorth
3. new password:
4. retype new password:

添加共享账号chihiro，密码为atenorth：

1. [root@server0 ~]# useradd chihiro
2. [root@server0 ~]# pdbedit -a chihiro             //根据提示设好密码atenorth
3. new password:
4. retype new password:

确认共享账号：

1. [root@server0 ~]# pdbedit -L
2. harry:1003:
3. chihiro:1005:
4. kenji:1004:

3）准备共享文件夹

1. [root@server0 ~]# mkdir /common
2. [root@server0 ~]# mkdir /devops
3. [root@server0 ~]# setfacl -m u:chihiro:rwx /devops         //配置写入权限

4）调整SELinux开关策略，允许发布可写的Samba共享资源

1. [root@server0 ~]# getsebool -a | grep ^samba\_exp         //默认配置
2. samba\_export\_all\_ro --> off
3. samba\_export\_all\_rw --> off
4. [root@server0 ~]# setsebool -P samba\_export\_all\_rw=on         //永久打开设置
5. [root@server0 ~]# getsebool -a | grep ^samba\_exp         //查看结果
6. samba\_export\_all\_ro --> off
7. samba\_export\_all\_rw --> on

5）配置共享目录

1. [root@server0 ~]# vim /etc/samba/smb.conf
2. [global]
3. workgroup = STAFF
4. .. ..
5. [common]
6. path = /common
7. hosts allow = 172.25.0.0/24
8. [devops]
9. path = /devops
10. hosts allow = 172.25.0.0/24
11. write list = chihiro

6）启动系统服务smb，并设置开机自启

1. [root@server0 ~]# systemctl restart smb
2. [root@server0 ~]# systemctl enable smb
3. ln -s '/usr/lib/systemd/system/smb.service' '/etc/systemd/system/multi-user.target.wants/smb.service'
4. [root@server0 ~]# netstat -antpu | grep smb
5. tcp 0 0 0.0.0.0:445 0.0.0.0:\* LISTEN 4709/smbd
6. tcp 0 0 0.0.0.0:139 0.0.0.0:\* LISTEN 4709/smbd

步骤二：在客户机desktop0测试Samba共享资源

1）安装软件包samba-client

1. [root@server0 ~]# yum -y install samba-client
2. .. ..

2）浏览目标主机提供了哪些共享资源

1. [root@desktop0 ~]# smbclient -L server0.example.com
2. Enter root's password:                         //此处无需输入密码，直接回车
3. Anonymous login successful
4. Domain=[MYGROUP] OS=[Unix] Server=[Samba 4.1.1]
5. Sharename Type Comment
6. --------- ---- -------
7. common Disk
8. devops Disk
9. IPC$ IPC IPC Service (Samba Server Version 4.1.1)
10. Anonymous login successful
11. Domain=[MYGROUP] OS=[Unix] Server=[Samba 4.1.1]
12. Server Comment
13. --------- -------
14. Workgroup Master
15. --------- -------

3）连接到目标主机的共享目录

1. [root@desktop0 ~]# smbclient -U harry //server0.example.com/common
2. Enter harry's password:                             //输入harry的密码
3. Domain=[STAFF] OS=[Unix] Server=[Samba 4.1.1]
4. smb: \> ls                                         //检查是否可列出目录内容
5. . D 0 Sun Nov 27 03:07:29 2016
6. .. D 0 Sun Nov 27 03:07:32 2016
7. 40913 blocks of size 262144. 27826 blocks available
8. smb: \> quit                                         //退出smb:\>交互环境
9. [root@desktop0 ~]#

## 2 ：多用户Samba挂载

### 2.1 需求

在虚拟机 desktop0 上访问 server0 提供的共享 devops，特性如下：

1. 将此共享永久挂载在 /mnt/dev 目录
2. 挂载时以用户 kenji 作为认证
3. 必要的时候，任何普通用户都可以通过用户 chihiro 来临时获取写的权限

### 2.2 方案

Samba客户端的multiuser挂载：支持切换访问Samba共享的用户身份，但不需要重新挂载共享资源。挂载参数需要添加“multiuser,sec=ntlmssp”，客户机上的普通用户可以通过cifscreds命令提交新的身份凭据。

在客户端挂载Samba共享目录，需要软件包cifs-utils的支持。

为访问网络资源配置开机挂载时，注意添加参数“\_netdev”，表示等客户机网络配置可用以后才挂载对应资源。

### 2.3 步骤

实现此案例需要按照如下步骤进行。

步骤一：挂载Samba共享目录

1）创建挂载点

1. [root@desktop0 ~]# mkdir /mnt/dev

2）安装cifs-utils软件包

1. [root@desktop0 ~]# yum -y install cifs-utils
2. .. ..

3）配置开机挂载

1. [root@desktop0 ~]# vim /etc/fstab
2. .. ..
3. //server0.example.com/devops /mnt/dev cifs username=kenji,password=atenorth,\_netdev     0     0

4）测试挂载配置

1. [root@desktop0 ~]# mount -a
2. [root@desktop0 ~]# df -hT /mnt/dev
3. Filesystem Type Size Used Avail Use% Mounted on
4. //server0.example.com/devops cifs 10G 3.2G 6.8G 32% /mnt/dev

步骤二：启用multiuser多用户支持

1）修改挂载配置，添加多用户支持

1. [root@desktop0 ~]# vim /etc/fstab
2. .. ..
3. //server0.example.com/devops    /mnt/dev    cifs    username=kenji,password=atenorth,multiuser,sec=ntlmssp,\_netdev 0 0
4. [root@desktop0 ~]# umount /mnt/dev                 //卸载此共享
5. [root@desktop0 ~]# mount /mnt/dev                 //重新挂载此共享

2）验证多用户访问

切换到普通用户student验证，无权访问挂载点/mnt/dev：

1. [root@desktop0 ~]# su - student
2. Last login: Sun Nov 27 03:51:32 CST 2016 on pts/0
3. [student@desktop0 ~]$ ls /mnt/dev
4. ls: cannot access /mnt/dev: Permission denied

以共享用户chihiro身份提交新的访问凭据，再次验证，对挂载点/mnt/dev可读写：

1. [student@desktop0 ~]$ cifscreds -u chihiro add server0.example.com
2. Password:                                         //输入共享账号chihiro的密码
3. [student@desktop0 ~]$ touch /mnt/dev/a.txt
4. [student@desktop0 ~]$ ls /mnt/dev/a.txt
5. /mnt/dev/a.txt
