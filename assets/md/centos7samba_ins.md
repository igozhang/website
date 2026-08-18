# centos7samba服务器部署

## ENV

cent79  
samba4.10.16  
samba的权限管控比较方便

## 安装配置samba

\*\* 注意samba配置不允许后面留空格否则服务起不来报错 set\_variable\_helper(yes # 是否可写入): value is not boolean! \*\*  
yum -y install samba-4.10.16  
追加共享配置到 /etc/samba/smb.conf  
tee >>/etc/samba/smb.conf <<EOF  
[igosmb]  
comment = hello smb  
path= /data/smb  
writable = yes  
public = no  
EOF

## 添加linux用户

useradd -d /data/smb -s /sbin/nologin igo

## 将用户添加到samba

可以用pdbedit或者smbpasswd 二选一，建议pdbedit  
\  
pdbedit -a -u igo

## 开防火墙

\*\* 只开445就可以，其他的酌情开放 \*\*  
firewall-cmd --zone=public --add-port=139/tcp --permanent  
firewall-cmd --zone=public --add-port=137/udp --permanent  
firewall-cmd --zone=public --add-port=138/udp --permanent  
firewall-cmd --zone=public --add-port=445/tcp --permanent  
firewall-cmd --reload  
systemctl restart firewalld.service

Port 137 (UDP) - NetBIOS 名字服务 ； nmbd  
Port 138 (UDP) - NetBIOS 数据报服务  
Port 139 (TCP) - 文件和打印共享 ； smbd （基于SMB(Server Message Block)协议，主要在局域网中使用，文件共享协议）  
Port 389 (TCP) - 用于 LDAP (Active Directory Mode)  
Port 445 (TCP) - NetBIOS服务在windos 2000及以后版本使用此端口, (Common Internet File System，CIFS，它是SMB协议扩展到Internet后，实现Internet文件共享)  
Port 901 (TCP) - 用于 SWAT，用于网页管理Samba

## 开selinux

getsebool -a |grep samba  
setsebool -P samba\_domain\_controller on  
setsebool -P samba\_enable\_home\_dirs on  
setsebool -P allow\_smbd\_anon\_write=1  
chcon -Rt samba\_share\_t /data/smb  
getsebool -a |grep samba

\*\* samba\_domain\_controller \*\*  
Turn the samba\_domain\_controller Boolean on to allow a Samba PDC to use the useradd and groupadd family of binaries.  
When enabled, this Boolean allows Samba to act as a domain controller, as well as giving it permission to execute related commands such as useradd, groupadd, and passwd.  
\*\* samba\_enable\_home\_dirs \*\*  
Enabling this Boolean allows Samba to share users' home directories.  
\*\* allow\_smbd\_anon\_write \*\*  
Having this Boolean enabled allows smbd to write to a public directory, such as an area reserved for common files that otherwise has no special access restrictions.

## 启动服务

systemctl start smb

## 验证

windows验证 \192.168.0.33\igosmb

## tips

pdbedit -L ：查看samba用户  
pdbedit -a -u user：添加samba用户  
pdbedit -r -u user：修改samba用户  
pdbedit -x -u user：删除samba用户

smbpasswd -a user：添加一个samba用户  
smbpasswd -d user：禁用一个samba用户  
smbpasswd -e user：恢复一个samba用户  
smbpasswd -x user：删除一个samba用户
