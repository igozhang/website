# mysql5单机主从

```
env

cent79
mysql-5.7.43-linux-glibc2.12-x86_64
同主机搭建两个节点3306主，3307从

1. 3306主实例部署
2. 3307从实例部署
3. 建主从
4. 报错表
```

```
一： 3306主节点部署
加用户
groupadd mysql
useradd -r -g mysql -s /bin/false mysql

建目录
mkdir -p /data/3306/{base,data,run}
chown -R mysql.mysql /data/3306/{base,data,run}

传文件
sftp->put
解压缩
tar -vxf mysql-5.7.43-linux-glibc2.12-x86_64.tar
tar -xvf mysql-5.7.43-linux-glibc2.12-x86_64.tar.gz
建软链
ln -s mysql-5.7.43-linux-glibc2.12-x86_64 mysql
环境变量
export PATH=$PATH:/data/mysql/bin

写配置
tee> /data/3306/my.cnf <<EOF
[mysqld]

user = mysql
basedir = /data/mysql
datadir = /data/3306/data
socket = /data/3306/run/mysql.sock
port=3306
server-id = 3306
log-bin=mysql-bin
EOF

初始化
/data/mysql/bin/mysqld --defaults-file=/data/3306/my.cnf  --initialize --user=mysql --basedir=/data/mysql
/data/mysql/bin/mysql_ssl_rsa_setup  创建证书及密钥对文件,这一步可以跳过

启动：
/data/mysql/bin/mysqld_safe --defaults-file=/data/3306/my.cnf --user=mysql &
mysql -uroot -S/data/3306/run/mysql.sock -p  #密码是初始化init步骤会生成一个临时密码
set password = password('123456ab');         #设置密码
alter user 'root'@'localhost'password expire never; #永不过期
flush privileges;

登录，碰到密码问题参考排错部分,分两次启动，第一次skip配置密码并设置不过期，然后第二次就可以正式启动了；
第一次启动服务（skip网络用来设置密码）
/data/mysql/bin/mysqld_safe --defaults-file=/data/3306/my.cnf --user=mysql --skip-grant-tables --skip-networking &
mysql -uroot -S/data/3307/run/mysql.sock -p #无密码直接登录
use mysql;
update mysql.user set authentication_string = password("123456") where user="root";
flush privileges;
set password = password('123456ab');
alter user 'root'@'localhost'password expire never; #永不过期
flush privileges;
shutdown

第二次启动服务（正式启动）
/data/mysql/bin/mysqld_safe --defaults-file=/data/3306/my.cnf --user=mysql &
mysql -uroot -S/data/3306/run/mysql.sock -p
```

```
二： 3307备节点部署
同机器启动3307进程做slave节点：
mkdir -p /data/3307/{base,data,run}
chown -R mysql.mysql /data/3307/{base,data,run}
写配置
tee> /data/3307/my.cnf <<EOF
[mysqld]

user = mysql
basedir = /data/mysql
datadir = /data/3307/data
socket = /data/3307/run/mysql.sock
port=3307
server-id = 3307
log-bin=mysql-bin
EOF
初始化
/data/mysql/bin/mysqld --defaults-file=/data/3307/my.cnf  --initialize --user=mysql --basedir=/data/mysql

启动：
/data/mysql/bin/mysqld_safe --defaults-file=/data/3307/my.cnf --user=mysql &
mysql -uroot -S/data/3307/run/mysql.sock -p  #密码是初始化init步骤会生成一个临时密码
set password = password('123456ab');         #设置密码
alter user 'root'@'localhost'password expire never; #永不过期
flush privileges;

登录，碰到密码问题参考排错部分,分两次启动，第一次skip配置密码并设置不过期，然后第二次就可以正式启动了；
第一次启动服务（skip网络用来设置密码）
/data/mysql/bin/mysqld_safe --defaults-file=/data/3307/my.cnf --user=mysql --skip-grant-tables --skip-networking &
mysql -uroot -S/data/3307/run/mysql.sock -p #无密码直接登录
use mysql;
update mysql.user set authentication_string = password("123456") where user="root";
flush privileges;
set password = password('123456ab');
alter user 'root'@'localhost'password expire never; #永不过期
flush privileges;
shutdown

第二次启动服务（正式启动）
/data/mysql/bin/mysqld_safe --defaults-file=/data/3307/my.cnf --user=mysql &
mysql -uroot -S/data/3307/run/mysql.sock -p
```

```
三： 建主从
主3306
mysql -uroot -S/data/3306/run/mysql.sock -p
GRANT REPLICATION SLAVE ON *.* to 'root'@'%' identified by '123456';
show master status; (Empty set问题需要配置log-bin=mysql-bin 并重启mysql)
+------------------+----------+--------------+------------------+-------------------+
| File             | Position | Binlog_Do_DB | Binlog_Ignore_DB | Executed_Gtid_Set |
+------------------+----------+--------------+------------------+-------------------+
| mysql-bin.000001 |      154 |              |                  |                   |
+------------------+----------+--------------+------------------+-------------------+

从3307
mysql -uroot -S/data/3307/run/mysql.sock -p
change master to master_host='192.168.3.33',master_user='root',master_password='123456',master_log_file='mysql-bin.000001',master_log_pos=154;
start slave;
######################
        Relay_Master_Log_File: mysql-bin.000001
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes

主建个表，从查看验证下，over；
```

```
报错1.
[ERROR] --initialize specified but the data directory has files in it. Aborting
需要将/etc/my.cnf 里指定的datadir里的mysql文件mv掉，因为这个报错表示数据库已经初始化过了
mv /var/lib/mysql/ /var/lib/mysqlbak

报错2.初始无密码无法登录
a. kill进程；
b. skip登录
/data/mysql/bin/mysqld_safe --defaults-file=/data/3306/my.cnf --user=mysql --skip-grant-tables --skip-networking &
c. 添加密码
use mysql;
update mysql.user set authentication_string = password("123456") where user="root";
flush privileges;

报错3. 密码过期
# mysqladmin -uroot  -S/data/3306/run/mysql.sock -p 123456
Enter password: 
mysqladmin: connect to server at 'localhost' failed
error: 'Your password has expired. To log in you must change it using a client that supports expired passwords.'

mysql -uroot -S/data/3306/run/mysql.sock -p123456
set password = password('123456ab');
alter user 'root'@'localhost'password expire never;
flush privileges;
```
