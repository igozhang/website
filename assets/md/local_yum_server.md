# 搭建内网Yum服务器

Env:  
CentOS7.9

```
本地镜像源：
CentOS-7-x86_64-DVD-1708.iso
mkdir -m 755 -p /mnt/cent7
mount /dev/cdrom  /mnt/cent7

写repo： vi /etc/yum.repos.d/t.repo
[Cent7]
name=Cent7
baseurl=file:///mnt/cent7/
enabled=1
gpgcheck=0
```

HTTP内网Yum服务器

```
1.挂载光盘
2.安装nginx yum install -y nginx
3.配置nginx：
cat /etc/nginx/conf.d/cent7.conf
# cat /etc/nginx/conf.d/cent7.conf
server
{
location /cent79 {
    alias   /mnt/cent7/;
    autoindex on;
    autoindex_exact_size off;
    autoindex_localtime on;
}
}

4.重启生效配置nginx
nginx -s reload
5.客户端写repo：
 cat /etc/yum.repos.d/t.repo
[cent7]
name=cent7
baseurl=http://10.86.20.47/cent7/
enabled=1
gpgcheck=0
```

FTP内网Yum服务器

```
yum install vsftpd
systemctl start vsftpd
浏览器 ftp://10.86.20.47验证，默认root目录是/var/ftp/
mkdir -p /var/ftp/cent7
cp -r /mnt/cent79/Packages/* /var/ftp/cent7/
createrepo /var/ftp/cent7/    ##会生成一个repodata目录
ls /var/ftp/cent7/repodata/   

修改vsftp默认目录，配置添加两行
tail -2 /etc/vsftpd/vsftpd.conf
anon_root=/home/ftp
anon_upload_enable=NO
```

```
tips

mkdir -p /var/ftp/cent7
createrepo /var/ftp/cent7
reposync -r base -p /var/ftp/cent7
```
