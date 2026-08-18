# mysql添加zabbix-agent2监控

```
CentOS Linux release 7.6.1810 (Core)
Docker version 1.13.1, build 7d71120/1.13.1
zabbix-agent2-5.0.26-1.el7.x86_64
```

```
安装mysql
docker pull mysql:5.7.39
docker run \
--name mysql-srv \
-p 3306:3306 \
-e MYSQL_ROOT_PASSWORD='igozhang.cn' \
-e MYSQL_USER='igozhang' \
-e MYSQL_PASSWORD='igozhang.cn' \
-v /etc/localtime:/etc/localtime \
-d mysql:5.7.39 \
--character-set-server=utf8mb4 \
--collation-server=utf8mb4_unicode_ci
```

```
安装zabbix-agent2
rpm -Uvh https://mirrors.aliyun.com/zabbix/zabbix/5.0/rhel/7/x86_64/zabbix-release-5.0-1.el7.noarch.rpm
yum -y install zabbix-agent2

cp /etc/zabbix/zabbix_agent2.conf{,.$(date +%Y%m%d)}
tee >/etc/zabbix/zabbix_agent2.conf <<EOF
PidFile=/var/run/zabbix/zabbix_agent2.pid
LogFile=/var/log/zabbix/zabbix_agent2.log
LogFileSize=0
Server=192.168.3.222
ServerActive=192.168.3.222
Hostname=241mysql
Include=/etc/zabbix/zabbix_agent2.d/*.conf
ControlSocket=/tmp/agent.sock
EOF

systemctl start zabbix-agent2
```

```
配置zbx_srv
WEB端：
主机模板(Template DB MySQL by Zabbix agent 2)
宏：
{$MYSQL.DSN} tcp://192.168.3.241:3306
{$MYSQL.PASSWORD} igozhang.cn
{$MYSQL.USER} igozhang

重启zabbix-agent2就有数据了
systemctl restart zabbix-agent2
```
