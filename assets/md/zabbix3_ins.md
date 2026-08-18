# zabbix3安装部署

```
Env：
CentOS7.4，内核3.10
zabbix版本3.4.15
```

```
初始化：
关防火墙，关selinux；
sed -i "s/SELINUX=enforcing/SELINUX=disabled/g" /etc/selinux/config
sed -i 's/SELINUXTYPE=targeted/#&/' /etc/selinux/config
setenforce 0 # 可以设置配置文件永久关闭
systemctl stop iptables.service
systemctl stop firewalld.service
```

```
安装zabbix
安装mariadb
初始化mariadb
修改zabbix_server配置文件
修改PHP时区
部署zabbix_agent
```

```
安装zabbix

$ rpm -i http://repo.zabbix.com/zabbix/3.4/rhel/7/x86_64/zabbix-release-3.4-2.el7.noarch.rpm
$ yum install zabbix-server-mysql zabbix-web-mysql zabbix-agent
```

```
安装mariadb

$ yum -y install mariadb mariadb-devel mariadb-server # centos7下安装的是mariadb
$ systemctl enable mariadb
$ systemctl start mariadb
    设置初始密码
$ mysqladmin -u root -h localhost password 'Ab123456'
    创建数据库 zabbix 并授权
$ mysql -uroot -pAb123456
> create database zabbix character set utf8 collate utf8_bin;
> grant all privileges on zabbix.* to zabbix@192.168.153.152 identified by 'zabbix'; 客户机授权
> CREATE USER 'zabbix'@'localhost' IDENTIFIED BY 'zabbix';
> grant all privileges on zabbix.* to zabbix@localhost identified by 'zabbix'
> quit
```

```
初始化数据库

$ zcat /usr/share/doc/zabbix-server-mysql*/create.sql.gz | mysql -uzabbix -pzabbix zabbix_database_name
```

```
修改zabbix_server配置文件

$ vim /etc/zabbix/zabbix_server.conf
DBHost=192.168.153.153
DBName=zabbix
DBUser=zabbix
DBPassword=zabbix
```

```
修改PHP时区

$ vim /etc/httpd/conf.d/zabbix.conf
将下面这行修改为亚洲上海
php_value date.timezone Asia/Shanghai
```

```
配置本机配置zabbix-agent

$ vim /etc/zabbix/zabbix_agentd.conf # 修改agent配置文件
Server=192.168.2.205 # zabbix服务端内网IP
ServerActive=192.168.2.205 # zabbix服务端内网IP，Active表示agent主动推送
Hostname=Jump Server # zabbix客户端主机名称，需要和web端添加的名称一致
Include=/etc/zabbix/zabbix_agentd.d/*.conf
```

```
启动进程
$ systemctl restart zabbix-server zabbix-agent httpd
$ systemctl enable zabbix-server zabbix-agent httpd
```

注意：必须要选择监控模板之后ZBX方块才会显示绿色！！，否则一直显示灰色
