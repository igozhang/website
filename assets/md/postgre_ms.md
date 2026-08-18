# postgre_ms

# postgre15.4单机主从搭建

env  
cent79  
postgresql-15.4.tar  
https://www.postgresql.org/ftp/source/v15.4/

基于主机的客户端认证配置文件  
http://www.postgres.cn/docs/13/auth-pg-hba-conf.html

adduser postgres  
cd /opt  
tar -xf postgresql-15.4.tar.gz  
ln -s postgresql-15.4 postgre  
yum -y install gcc readline-devel zlib  
cd /opt/postgre && ./configure --without-zlib  
make && make install  
export PATH=$PATH:/usr/local/pgsql/bin/

## 一，部署主节点8432，默认是5432端口

mkdir -p /opt/8432/{data,logs,archive}  
chown -R postgres.root /opt/8432  
su - postgres  
export PATH=$PATH:/usr/local/pgsql/bin/  
初始化数据库  
su - postgres  
/usr/local/pgsql/bin/initdb -D /opt/8432/data/

修改配置/opt/8432/data/postgresql.conf  
tee >>/opt/8432/data/postgresql.conf <<EOF  
listen\_addresses = '\*'  
port = 8432  
max\_connections = 100 #最大连接数，从库的 max\_connections 必须要大于主库的  
synchronous\_commit = on #开启同步复制  
max\_wal\_senders = 3 #同步最大的进程数量  
wal\_level = hot\_standby #启用热备模式  
wal\_sender\_timeout = 60s #流复制主机发送数据的超时时间  
archive\_mode = on  
archive\_command = 'cp %p /data/8432/archive/%f '  
EOF

修改配置/opt/8432/data/pg\_hba.conf  
tee >>/opt/8432/data/pg\_hba.conf <<EOF  
host all all 0.0.0.0/0 md5  
host replication repl 0.0.0.0/0 md5  
EOF

启动数据库  
/usr/local/pgsql/bin/pg\_ctl -D /opt/8432/data/ -l /opt/8432/logs/logfile start  
设置密码  
psql -U postgres -h 127.0.0.1 --port 8432 连接127.0.0.1不需要密码，已其他IP连接需要密码  
postgres=# ALTER USER postgres WITH PASSWORD 'ab1234';  
postgres=# create user repl REPLICATION LOGIN ENCRYPTED PASSWORD 'repl123'; #复制账号  
quit;

## 二，部署从节点8433

mkdir -p /opt/8433/{data,logs,archive}  
chown -R postgres.root /opt/8433  
su - postgres  
export PATH=$PATH:/usr/local/pgsql/bin/

如果已经初始化，需要删除数据目录，重新同步主库数据  
rm -rf /opt/8433/data/\*  
pg\_basebackup -h 192.168.3.33 -p 8432 -U repl -Fp -Xs -Pv -R -D /opt/8433/data  
-Fp: 表示以原始平面文件拷贝文件  
-Xs: 表示以流复制访问拷贝日志  
-Pv: 表示显示过程  
-R: 表示创建备库文件和恢复配置信息  
-c: 表示拷贝前做一次日志切换  
-D: 从节点的数据目录

修改配置文件vim /opt/8433/data/postgresql.conf  
将master同步过来的配置替换成  
listen\_addresses = '\*'  
port = 8433  
max\_connections = 1000 #最大连接数，从库的 max\_connections 必须要大于主库的  
hot\_standby = on # 开启热备  
max\_standby\_streaming\_delay = 30s # 数据流备份的最大延迟时间  
wal\_receiver\_status\_interval = 1s # 从节点向主节点报告自身状态的最长间隔时间  
hot\_standby\_feedback = on # 如果有错误的数据复制向主进行反馈  
synchronous\_commit = on #开启同步复制  
max\_wal\_senders = 3 #同步最大的进程数量  
wal\_level = hot\_standby #启用热备模式  
wal\_sender\_timeout = 60s #流复制主机发送数据的超时时间  
archive\_mode = on  
archive\_command = 'cp %p /data/8433/archive/%f'

primary\_conninfo = 'host=192.168.3.33 port=8432 user=repl password=repl123' #对应主库的连接信息  
recovery\_target\_timeline = 'latest' #流复制同步到最新的数据，原recovery.conf文件统一配置到postgresql.conf

chmod 0750 /opt/8433/data #权限要求，否则报错

启动数据库  
/usr/local/pgsql/bin/pg\_ctl -D /opt/8433/data/ -l /opt/8433/logs/logfile start

## 验证主从

主插条数据，从查数据，over

psql -U postgres -h 127.0.0.1 --port 8432  
创建数据库 create database igodb;  
查看数据库 \l  
进入数据库 \c igodb  
创建数据表 create table igotb(id int);  
查看表格信息 \d igotb  
表格插入数据 insert into igotb(id)values(18);

psql -U postgres -h 127.0.0.1 --port 8433  
进入数据库 \c igodb  
查数据 select \* from igotb;
