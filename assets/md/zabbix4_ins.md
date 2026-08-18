# zabbix4部署

```
env
cent72
zabbix4.1
tag: zabbix4_ins
老笔记整理，现在的解决方式一般helm或者docker，追加docker-compose

zabbbix4
mysql5.6
docker-compose
```

```
yum -y install httpd
yum -y install php
```

```
zabbbix4安装

rpm -ivh https://repo.zabbix.com/zabbix/4.2/rhel/7/x86_64/zabbix-release-4.2-1.el7.noarch.rpm
yum -y install yum-utils
yum-config-manager --enable rhel-7-server-optional-rpms
server：
yum install -y zabbix-server-mysql
yum install -y zabbix-web-mysql  

proxy:
yum install zabbix-proxy-mysql
```

```
mysql5.6安装导库

yum -y install *.rpm
改密
建库授权：
create database zabbix character set utf8 collate utf8_bin;
grant all privileges on zabbix.* to zabbix@'192.168.5.%' identified by 'zabbix';
初始化数据库，建表
zcat create.sql.gz | mysql -uzabbix -pzabbix zabbix；
修改zbx_serv 配置：
grep ^DB /etc/zabbix/zabbix_server.conf
```

```
serv_conf
[root@VM_0_13_centos]#  grep ^DB /etc/zabbix/zabbix_server.conf
DBHost=localhost #或目标ip
DBName=zabbix
DBUser=zabbix
DBPassword=zabbix

改时区：

# vim /etc/httpd/conf.d/zabbix.conf
 php_value date.timezone Asia/Shanghai
```

```
docker-compose

docker-compose
wget https://raw.githubusercontent.com/sunsharing-note/zabbix/master/zabbix_server.conf
mv zabbix_server.conf zabbix/zabbixconfig/
```

```
vim docker-compose.yml
version: '3'
services:
  zabbix-mysql:
    image: mariadb:latest
    ports:
        - "3309:3306"
    volumes:
        - ./zabbix/data/mysqllib:/var/lib/mysql
    environment:
        - MYSQL_USER=zabbix
        - MYSQL_DATABASE=zabbix
        - MYSQL_PASSWORD=zabbix
        - MYSQL_ROOT_PASSWORD=123321
        - character-set-server=utf8
        - collation-server=utf8_bin
    restart: always
  zabbix-server:
    image: zabbix/zabbix-server-mysql:centos-latest
    restart: always
    ports:
      - "10052:10051"
    environment:
      - DB_SERVER_HOST=zabbix-mysql
      - MYSQL_USER=zabbix
      - MYSQL_DATABASE=zabbix
      - MYSQL_PASSWORD=zabbix
      - DB_SERVER_ROOT_USER=root
      - DB_SERVER_ROOT_PASS=123321
      - TZ＝'Asia/Shanghai'
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
      - ./zabbix/zabbixconfig/:/etc/zabbix/
      - ./zabbix/alertscripts:/usr/lib/zabbix/alertscripts
      - ./zabbix/externalscripts:/usr/lib/zabbix/externalscripts
      - ./zabbix/modules:/var/lib/zabbix/modules
      - ./zabbix/enc:/var/lib/zabbix/enc
      - ./zabbix/ssh_keys:/var/lib/zabbix/ssh_keys
      - ./zabbix/ssl/certs:/var/lib/zabbix/ssl/certs
      - ./zabbix/ssl/keys:/var/lib/zabbix/ssl/keys
      - ./zabbix/ssl/ssl_ca:/var/lib/zabbix/ssl/ssl_ca
      - ./zabbix/snmptraps:/var/lib/zabbix/snmptraps
      - ./zabbix/mibs:/var/lib/zabbix/mibs
    depends_on:
      - zabbix-mysql
  zabbix-web-nginx:
    image: zabbix/zabbix-web-nginx-mysql:latest
    restart: always
    ports:
      - "8088:80"
    environment:
      - DB_SERVER_HOST=zabbix-mysql
      - MYSQL_DATABASE=zabbix
      - MYSQL_USER=zabbix
      - MYSQL_PASSWORD=zabbix
      - MYSQL_ROOT_PASSWORD=123321
      - ZBX_SERVER_HOST=zabbix-server
      - PHP_TZ="Asia/Shanghai"
      - TZ＝'Asia/Shanghai'
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    links:
      - zabbix-mysql
      - zabbix-server
   zabbix-agent:
    image: zabbix/zabbix-agent:latest
    environment:
      - ZBX_HOSTNAME=Zabbix server
      - ZBX_SERVER_HOST=zabbix-server
      - ZBX_SERVER_PORT=10051
      - TZ='Asia/Shanghai'
    ports:
      - "10053:10050"
    restart: always
    privileged: true
```

```
docker-compose -f docker-compose.yml up -d
192.168.3.168:8088
```
