# 容器部署zabbix5

```
CentOS Linux release 7.6.1810 (Core)
Docker version 1.13.1, build 7d71120/1.13.1

https://www.zabbix.com/documentation/3.4/zh/manual/installation/containers
有方法，还有compose文件
```

```
docker pull mysql:5.7
docker pull zabbix/zabbix-java-gateway:5.0.26-centos
docker pull zabbix/zabbix-server-mysql:5.0.26-centos
docker pull zabbix/zabbix-web-nginx-mysql:5.0.26-centos
```

```
拉起mysql数据库;
# docker run --name mysql-server -t \
             -e MYSQL_DATABASE="zabbix" \
             -e MYSQL_USER="zabbix" \
             -e MYSQL_PASSWORD="igozhang.cn" \
             -e MYSQL_ROOT_PASSWORD="igozhang.cn" \
             -v /etc/localtime:/etc/localtime \
             -d mysql:5.7 \
             --character-set-server=utf8 \
             --collation-server=utf8_bin

拉起Zabbix Java gateway,用于jvm监控;
# docker run --name zabbix-java-gateway -t \
             -v /etc/localtime:/etc/localtime \
             -d zabbix/zabbix-java-gateway:5.0.26-centos

拉起zabbix—serv并使用link关联mysql—serv容器（不推荐，link有顺序依赖关系，必须先启动被link容器，才能启动link容器，否则报错）;
# docker run --name zabbix-server-mysql -t \
             -e DB_SERVER_HOST="mysql-server" \
             -e MYSQL_DATABASE="zabbix" \
             -e MYSQL_USER="zabbix" \
             -e MYSQL_PASSWORD="igozhang.cn" \
             -e MYSQL_ROOT_PASSWORD="igozhang.cn" \
             -e ZBX_JAVAGATEWAY="zabbix-java-gateway" \
             --link mysql-server:mysql \
             --link zabbix-java-gateway:zabbix-java-gateway \
             -p 10051:10051 \
             -v /etc/localtime:/etc/localtime \
             -d zabbix/zabbix-server-mysql:5.0.26-centos

拉起zbx_web,并关联zbx_serv和mysql;
# docker run --name zabbix-web-nginx-mysql -t \
             -e DB_SERVER_HOST="mysql-server" \
             -e MYSQL_DATABASE="zabbix" \
             -e MYSQL_USER="zabbix" \
             -e MYSQL_PASSWORD="igozhang.cn" \
             -e MYSQL_ROOT_PASSWORD="igozhang.cn" \
             --link mysql-server:mysql \
             --link zabbix-server-mysql:zabbix-server \
             -p 80:8080 \
             -v /etc/localtime:/etc/localtime \
             -d zabbix/zabbix-web-nginx-mysql:5.0.26-centos
```

default: admin/zabbix
