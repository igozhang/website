# mysql_slave主从配置

1. 案例1：MySQL一主一从
2. 案例2：配置主从从同步结构
3. 配置半同步复制模式

## 1 案例1：MySQL一主一从

### 1.1 问题

- 构建 主-->从 复制结构
- 其中主机192.168.4.10作为主库
- 主机192.168.4.20作为从库

### 1.2 方案

使用2台RHEL 7虚拟机，如图-1所示。其中192.168.4.10是MySQL主服务器，负责提供同步源；另一台192.168.4.20作为MySQL从服务器，通过调取主服务器上的binlog日志，在本地重做对应的库、表，实现与主服务器的AB复制（同步）。

图－1

提前为两台MySQL服务器安装好MySQL-server、MySQL-Client软件包，并为数据库用户root修改密码；Linux客户机上则只需安装MySQL-Client软件包即可。

### 1.3 步骤

实现此案例需要按照如下步骤进行。

步骤一：初始化现有库

为了在启用binlog日志及同步之前保持主、从库的一致性，建议进行初始化——备份主服务器上现有的库，然后导入到从服务器上。

当现有库、表都采用MyISAM引擎时，可执行离线备份、恢复，这样更有效率；否则，可通过mysqldump等工具来实现库的导出、导入。

1）备份MySQL Master（192.168.4.10）上现有的库

如果服务器已经启用binlog，建议对日志做一次重置，否则可忽略：

1. [root@dbsvr1 ~]# mysql -u root -p
2. Enter password:                                 //以数据库用户root登入
3. .. ..
4. mysql> RESET MASTER;                             //重置binlog日志
5. Query OK, 0 rows affected (0.06 sec)
6. mysql> quit                                     //退出mysql> 环境
7. Bye

以备份mysql库、sys库为例，导出操作如下：

1. [root@dbsvr1 ~]# mysqldump -u root -p –all-databases > /root/mytest.sql
2. Enter password:                                     //验证口令
3. [root@dbsvr1 ~]# ls -lh /root/mytest.sql             //确认备份结果
4. -rw-r--r--. 1 root root 777172 4月 23 12:21 /root/mytest.sql

2）在MySQL Slave（192.168.4.20）上导入备份的库

先清理目标库，避免导入时冲突。主要是采用InnoDB引擎的库，授权库mysql多采用MyISAM引擎，可不做清理。

1. [root@dbsvr2 ~]# mysql -u root -p
2. Enter password:                                 //以数据库用户root登入
3. .. ..
4. mysql> DROP DATABASE test;                         //删除test库
5. Query OK, 0 rows affected (0.03 sec)
6. mysql> quit                                     //退出mysql> 环境
7. Bye

使用scp工具下载备份文件：

1. [root@dbsvr2 ~]# scp /root/mytest.sql root@192.168.4.20:/
2. root@dbsvr1's password:                         //验证对方系统用户root的口令
3. mytest.sql 100% 759KB 759.0KB/s 00:00
4. [root@dbsvr2 ~]# ls -lh mytest.sql             //确认下载结果
5. -rw-r--r--. 1 root root 759K 4月 23 12:22 /mytest.sql

执行导入操作：

1. [root@dbsvr2 ~]# mysql -u root -p < /mytest.sql
2. Enter password:                                 //验证口令

导入成功后，可重新登入 mysql> 环境，确认清理的目标库已恢复：

1. mysql> show databases;
2. +--------------------+
3. | Database |
4. +--------------------+
5. | information\_schema |
6. | mysql |
7. | performance\_schema |
8. | sys |
9. +--------------------+
10. 4 rows in set (0.00 sec)

步骤二：配置MySQL Master（主服务器，192.168.4.10）

1）修改/etc/my.cnf配置，重新启动MySQL服务程序

指定服务器ID号、允许日志同步：

1. [root@dbsvr1 mysql]# vim /etc/my.cnf
2. [mysqld]
3. log\_bin=dbsvr1-bin                     //启用binlog日志，并指定文件名前缀
4. server\_id = 10                         //指定服务器ID号
5. ......

重启mysql服务：

1. [root@dbsvr1 ~]# systemctl restart mysqld.service

2）新建一个备份用户，授予复制权限

需要的权限为REPLICATION SLAVE，允许其从Slave服务器访问：

1. mysql> GRANT REPLICATION SLAVE ON \*.\* TO 'replicater'@'192.168.4.%' IDENTIFIED BY 'pwd123';
2. Query OK, 0 rows affected, 1 warning (0.09 sec)

3）检查Master服务器的同步状态

在已经初始化现有库的情况下，查看MASTER状态，记录下当前的日志文件名、偏移的位置（下面SLAVE发起复制时需要用到）：

1. mysql> SHOW MASTER STATUS\G
2. \*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\* 1. row \*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*
3. File: dbsvr1-bin.000001 //记住当前的日志文件名
4. Position: 154 //记住当前的位置
5. Binlog\_Do\_DB:
6. Binlog\_Ignore\_DB:
7. Executed\_Gtid\_Set:
8. 1 row in set (0.00 sec)

步骤三：配置MySQL Slave（从服务器，192.168.4.20）

1）修改/etc/my.cnf配置，重新启动MySQL服务程序

指定服务器ID号、允许日志同步：

1. [root@dbsvr2 ~]# vim /etc/my.cnf
2. [mysqld]
3. log\_bin=dbsvr2-bin                     //启动SQL日志，并指定文件名前缀
4. server\_id = 20                         //指定服务器ID号，不要与Master的相同
5. .. ..

在生产环境中，还可以根据需要设置更详细的同步选项。比如，指定当主、从网络中断时的重试超时时间（slave-net-timeout=60 ）等，具体可参考MySQL手册。

配置完成后，重启mysql服务：

1. [root@dbsvr2 ~]# systemctl restart mysqld.service

通过CHANGE MASTER语句指定MASTER服务器的IP地址、同步用户名/密码、起始日志文件、偏移位置（参考MASTER上的状态输出）：

1. mysql> CHANGE MASTER TO MASTER\_HOST='192.168.4.10',
2. -> MASTER\_USER='replicater',
3. -> MASTER\_PASSWORD='pwd123',
4. -> MASTER\_LOG\_FILE='dbsvr1-bin.000002',     //对应Master的日志文件
5. -> MASTER\_LOG\_POS=334;                         //对应Master的日志偏移位置
6. Query OK, 0 rows affected, 2 warnings (0.12 sec)

然后执行START SLAVE（较早版本中为SLAVE START）启动复制：

1. mysql> START SLAVE;                             //启动复制
2. Query OK, 0 rows affected (0.00 sec)

注意：一旦启用SLAVE复制，当需要修改MASTER信息时，应先执行STOP SLAVE停止复制，然后重新修改、启动复制。

通过上述连接操作，MASTER服务器的设置信息自动存为master.info文件，以后每次MySQL服务程序时会自动调用并更新，无需重复设置。查看master.info文件的开头部分内容，可验证相关设置：

1. [root@dbsvr2 ~]# ls -lh /var/lib/mysql/master.info
2. -rw-r-----. 1 mysql mysql 132 4月 23 12:06 /var/lib/mysql/master.info
3. [root@dbsvr2 ~]# head /var/lib/mysql/master.info
4. 25
5. dbsvr1-bin.000001
6. 154
7. 192.168.4.10
8. replicater
9. pwd123
10. 3306
11. 60
12. 0

2）检查Slave服务器的同步状态

通过SHOW SLAVE STATUS语句可查看从服务器状态，确认其中的IO线程、SQL线程正常运行，才能成功同步：

1. mysql> SHOW SLAVE STATUS\G
2. Slave\_IO\_State: Waiting for master to send event
3. Master\_Host: 192.168.4.1
4. Master\_User: replicater
5. Master\_Port: 3306
6. Connect\_Retry: 60
7. Master\_Log\_File: dbsvr1-bin.000001
8. Read\_Master\_Log\_Pos: 154
9. Relay\_Log\_File: db2-relay-bin.000003
10. Relay\_Log\_Pos: 321
11. Relay\_Master\_Log\_File: dbsvr1-bin.000001
12. Slave\_IO\_Running: Yes //IO线程应该已运行
13. Slave\_SQL\_Running: Yes //SQL线程应该已运行
14. Replicate\_Do\_DB:
15. Replicate\_Ignore\_DB:
16. Replicate\_Do\_Table:
17. Replicate\_Ignore\_Table:
18. Replicate\_Wild\_Do\_Table:
19. Replicate\_Wild\_Ignore\_Table:
20. Last\_Errno: 0
21. Last\_Error:
22. Skip\_Counter: 0
23. Exec\_Master\_Log\_Pos: 154
24. Relay\_Log\_Space: 2490
25. Until\_Condition: None
26. Until\_Log\_File:
27. Until\_Log\_Pos: 0
28. Master\_SSL\_Allowed: No
29. Master\_SSL\_CA\_File:
30. Master\_SSL\_CA\_Path:
31. Master\_SSL\_Cert:
32. Master\_SSL\_Cipher:
33. Master\_SSL\_Key:
34. Seconds\_Behind\_Master: 0
35. Master\_SSL\_Verify\_Server\_Cert: No
36. Last\_IO\_Errno: 0
37. Last\_IO\_Error:
38. Last\_SQL\_Errno: 0
39. Last\_SQL\_Error:
40. Replicate\_Ignore\_Server\_Ids:
41. Master\_Server\_Id: 10
42. Master\_UUID: 2d4d8a11-27b7-11e7-ae78-52540055c180
43. Master\_Info\_File: /var/lib/mysql/master.info
44. SQL\_Delay: 0
45. SQL\_Remaining\_Delay: NULL
46. Slave\_SQL\_Running\_State: Slave has read all relay log; waiting for more updates
47. Master\_Retry\_Count: 86400
48. Master\_Bind:
49. Last\_IO\_Error\_Timestamp:
50. Last\_SQL\_Error\_Timestamp:
51. Master\_SSL\_Crl:
52. Master\_SSL\_Crlpath:
53. Retrieved\_Gtid\_Set:
54. Executed\_Gtid\_Set:
55. Auto\_Position: 0
56. Replicate\_Rewrite\_DB:
57. Channel\_Name:
58. Master\_TLS\_Version:
59. 1 row in set (0.00 sec)

若START SLAVE直接报错失败，请检查CHANGE MASTER相关设置是否有误，纠正后再重试；若IO线程或SQL线程有一个为“No”，则应检查服务器的错误日志，分析并排除故障后重启主从复制。

步骤四：测试主从同步效果

1）在Master上操作数据库、表、表记录

新建newdb库、newtable表，随意插入几条表记录：

1. mysql> CREATE DATABASE newdb;                         //新建库newdb
2. Query OK, 1 row affected (0.17 sec)
3. mysql> USE newdb;                                     //切换到newdb库
4. Database changed
5. mysql> CREATE TABLE newtable(id int(4));             //新建newtable表
6. Query OK, 0 rows affected (0.46 sec)
7. mysql> INSERT INTO newtable VALUES(1234),(5678);     //插入2条表记录
8. Query OK, 2 rows affected (0.24 sec)
9. Records: 2 Duplicates: 0 Warnings: 0
10. mysql> SELECT \* FROM newtable;                         //确认表数据
11. +------+
12. | id |
13. +------+
14. | 1234 |
15. | 5678 |
16. +------+
17. 2 rows in set (0.00 sec)

2）在Slave上确认自动同步的结果

直接切换到newdb库，并查询newtable表的记录，应该与Master上的一样，这才说明主从同步已经成功生效：

1. mysql> USE newdb;                                     //直接切换到newdb库
2. Reading table information for completion of table and column names
3. You can turn off this feature to get a quicker startup with -A
4. Database changed
5. mysql> SELECT \* FROM newtable;                     //输出表记录
6. +------+
7. | id |
8. +------+
9. | 1234 |
10. | 5678 |
11. +------+
12. 2 rows in set (0.02 sec)

3）在Master服务器上可查看Slave主机的信息

1. mysql> SHOW SLAVE HOSTS;
2. +-----------+------+------+-----------+--------------------------------------+
3. | Server\_id | Host | Port | Master\_id | Slave\_UUID |
4. +-----------+------+------+-----------+--------------------------------------+
5. | 2 | | 3306 | 10 | 512cf7c1-27c4-11e7-8f4b-5254007b030b |
6. +-----------+------+------+-----------+--------------------------------------+
7. 1 row in set (0.00 sec)

## 2 案例2：配置主从从同步结构

### 2.1 问题

- 具体要求如下：
- 配置主机192.168.4.51为主数据库服务器
- 配置主机192.168.4.52为51主机的从库服务器
- 配置主机192.168.4.53为52主机的从库服务器
- 客户端连接主数据库服务器51主机创建的数据，连接52和53主机时，也可以访问到库、表、记录。

### 2.2 方案

使用3台RHEL 7虚拟机，如图-2所示。其中192.168.4.51是MySQL主服务器，负责提供同步源；另一台192.168.4.52作为192.168.4.51从服务器，最后一台192.168.4.53作为192.168.4.52从服务器，通过调取主服务器上的binlog日志，客户端访问主库51 时 创建库表记录 在52 和53 数据库服务器都可以看到

图－2

### 2.3 步骤

实现此案例需要按照如下步骤进行。

步骤一：环境准备

为了在启用binlog日志及同步之前保持主、从库的一致性，主从同步未配置之前，要保证从库上要有主库上的数据，禁用selinux，关闭防火墙服务，保证物理连接正常

1. [root@db51 ~]# systemctl stop firewalld
2. [root@db51 ~]# setenforce 0
3. [root@db51 ~]# ping -c 2 192.168.4.51
4. PING 192.168.4.51 (192.168.4.51) 56(84) bytes of data.
5. 64 bytes from 192.168.4.51: icmp\_seq=1 ttl=64 time=0.059 ms
6. 64 bytes from 192.168.4.51: icmp\_seq=2 ttl=64 time=0.063 ms
7. --- 192.168.4.51 ping statistics ---
8. 2 packets transmitted, 2 received, 0% packet loss, time 999ms
9. rtt min/avg/max/mdev = 0.059/0.061/0.063/0.002 ms
10. [root@db51 ~]# ping -c 2 192.168.4.52
11. PING 192.168.4.52 (192.168.4.52) 56(84) bytes of data.
12. 64 bytes from 192.168.4.52: icmp\_seq=1 ttl=64 time=0.698 ms
13. 64 bytes from 192.168.4.52: icmp\_seq=2 ttl=64 time=0.365 ms
14. --- 192.168.4.52 ping statistics ---
15. 2 packets transmitted, 2 received, 0% packet loss, time 1000ms
16. rtt min/avg/max/mdev = 0.365/0.531/0.698/0.168 ms

步骤二：配置主服务器192.168.4.51

1）对yaya用户进行授权

1. [root@db51 ~]# mysql -uroot -p123456
2. mysql> grant replication slave on \*.\* to yaya@"%" identified by "123456";
3. Query OK, 0 rows affected, 1 warning (0.03 sec)

2）启用binlog日志，修改/etc/my.cnf配置，重新启动MySQL服务程序

指定服务器ID号、允许日志同步：

1. [root@db51 ~]# vim /etc/my.cnf
2. [mysqld]
3. log\_bin=db51                     //启用binlog日志，并指定文件名前缀
4. server\_id=51                     //指定服务器ID号
5. binlog-format="mixed" // 指定binlog日志格式

重启mysql服务：

1. [root@db51 ~]# systemctl restart mysqld

确保/var/lib/mysql下面有两个文件：

1. [root@db51 ~]# ls /var/lib/mysql/db51.\*
2. /var/lib/mysql/db51.000001 /var/lib/mysql/db51.index

查看主服务正在使用的日志信息

查看主服务器状态，记录下当前的日志文件名、偏移的位置（下面SLAVE发起复制时需要用到）：

1. mysql> show master status;
2. +-------------+----------+--------------+------------------+-------------------+
3. | File | Position | Binlog\_Do\_DB | Binlog\_Ignore\_DB | Executed\_Gtid\_Set |
4. +-------------+----------+--------------+------------------+-------------------+
5. | db51.000002 | 437 | | | |
6. +-------------+----------+--------------+------------------+-------------------+
7. 1 row in set (0.00 sec)

步骤三：配置从服务器192.168.4.52

1）在服务器192.168.4.52上对user53用户进行授权

1. [root@db52 ~]# mysql -u root -p123456
2. mysql> grant replication slave on \*.\* to user53@"192.168.4.53" identified by "654321";
3. Query OK, 0 rows affected, 1 warning (0.00 sec)

2）修改/etc/my.cnf配置，启用binlog日志，指定server\_id 和 允许级联复制

1. [root@db52 ~]# vim /etc/my.cnf
2. [mysqld]
3. server\_id=52
4. log-bin=db52
5. binlog-format="mixed"
6. log\_slave\_updates //允许级联复制

3）配置完成后，重启mysql服务：

1. [root@db52 ~]# systemctl restart mysqld

4）确保/var/lib/mysql下面有两个文件：

1. [root@db52 ~]# ls /var/lib/mysql/db52.\*
2. /var/lib/mysql/db52.000001 /var/lib/mysql/db52.index

5）查看正在使用的日志信息

1. [root@db52 ~]# mysql -uroot -p123456
2. mysql> show master status;
3. +-------------+----------+--------------+------------------+-------------------+
4. | File | Position | Binlog\_Do\_DB | Binlog\_Ignore\_DB | Executed\_Gtid\_Set |
5. +-------------+----------+--------------+------------------+-------------------+
6. | db52.000001 | 154 | | | |
7. +-------------+----------+--------------+------------------+-------------------+
8. 1 row in set (0.00 sec) //查看日志文件名、偏移的位置

6）验证主库的授权用户

[root@db52 ~]# mysql -h192.168.4.51 -uyaya -p123456

mysql: [Warning] Using a password on the command line interface can be insecure.

Welcome to the MySQL monitor. Commands end with ; or \g.

Your MySQL connection id is 4

Server version: 5.7.17-log MySQL Community Server (GPL)

Copyright (c) 2000, 2016, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its

affiliates. Other names may be trademarks of their respective

owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

1. mysql> //验证成功

7）通过change master语句指定master服务器的IP地址、同步用户名/密码、起始日志文件、偏移位置（参考master上的状态输出）：

[root@db52 ~]# mysql -uroot -p123456

mysql> change master to

-> master\_host="192.168.4.51",

-> master\_user="yaya",

-> master\_password="123456",

-> master\_log\_file="db51.000002",

-> master\_log\_pos=437;

Query OK, 0 rows affected, 2 warnings (0.43 sec)

8）启动slave进程

mysql> start slave;

Query OK, 0 rows affected (0.03 sec)

9）查看进程状态信息，通过show slave status语句可查看从服务器状态，确认其中的IO线程、SQL线程正常运行，才能成功同步,IO线程和SQL线程必须是Yes

mysql> show slave status \G;

\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\* 1. row \*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*

Slave\_IO\_State: Waiting for master to send event

Master\_Host: 192.168.4.51

Master\_User: yaya

Master\_Port: 3306

Connect\_Retry: 60

Master\_Log\_File: db51.000002

Read\_Master\_Log\_Pos: 437

Relay\_Log\_File: db52-relay-bin.000002

Relay\_Log\_Pos: 315

Relay\_Master\_Log\_File: db51.000002

Slave\_IO\_Running: Yes

Slave\_SQL\_Running: Yes

Replicate\_Do\_DB:

Replicate\_Ignore\_DB:

Replicate\_Do\_Table:

Replicate\_Ignore\_Table:

Replicate\_Wild\_Do\_Table:

Replicate\_Wild\_Ignore\_Table:

Last\_Errno: 0

Last\_Error:

Skip\_Counter: 0

Exec\_Master\_Log\_Pos: 437

Relay\_Log\_Space: 521

Until\_Condition: None

Until\_Log\_File:

Until\_Log\_Pos: 0

Master\_SSL\_Allowed: No

Master\_SSL\_CA\_File:

Master\_SSL\_CA\_Path:

Master\_SSL\_Cert:

Master\_SSL\_Cipher:

Master\_SSL\_Key:

Seconds\_Behind\_Master: 0

Master\_SSL\_Verify\_Server\_Cert: No

Last\_IO\_Errno: 0

Last\_IO\_Error:

Last\_SQL\_Errno: 0

Last\_SQL\_Error:

Replicate\_Ignore\_Server\_Ids:

Master\_Server\_Id: 51

Master\_UUID: 81a13101-aa66-11e8-ad11-525400019e62

Master\_Info\_File: /var/lib/mysql/master.info

SQL\_Delay: 0

SQL\_Remaining\_Delay: NULL

Slave\_SQL\_Running\_State: Slave has read all relay log; waiting for more updates

Master\_Retry\_Count: 86400

Master\_Bind:

Last\_IO\_Error\_Timestamp:

Last\_SQL\_Error\_Timestamp:

Master\_SSL\_Crl:

Master\_SSL\_Crlpath:

Retrieved\_Gtid\_Set:

Executed\_Gtid\_Set:

Auto\_Position: 0

Replicate\_Rewrite\_DB:

Channel\_Name:

Master\_TLS\_Version:

1 row in set (0.00 sec)

步骤四：配置从服务器192.168.4.53

1）验证主库的授权用户

1. [root@db53 ~]# mysql -h192.168.4.52 -uuser53 -p654321
2. mysql: [Warning] Using a password on the command line interface can be insecure.
3. Welcome to the MySQL monitor. Commands end with ; or \g.
4. Your MySQL connection id is 7
5. Server version: 5.7.17-log MySQL Community Server (GPL)
6. Copyright (c) 2000, 2016, Oracle and/or its affiliates. All rights reserved.
7. Oracle is a registered trademark of Oracle Corporation and/or its
8. affiliates. Other names may be trademarks of their respective
9. owners.
10. Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
11. mysql> //验证成功

2）指定server\_id

1. [mysqld]
2. validate\_password\_policy=0
3. validate\_password\_length=6
4. server\_id=53

3）重新启动服务

1. [root@db53 ~]# systemctl restart mysqld

4）管理员登录指定主库信息

1. [root@db53 ~]# mysql -uroot -p123456
2. mysql> change master to
3. -> master\_host="192.168.4.52",
4. -> master\_user="user53",
5. -> master\_password="654321",
6. -> master\_log\_file=" db52.000001",
7. -> master\_log\_pos=154;
8. Query OK, 0 rows affected, 2 warnings (0.37 sec)

5）启动slave进程

1. mysql> start slave;
2. Query OK, 0 rows affected (0.04 sec)

6）查看进程状态信息

1. mysql> show slave status\G
2. \*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\* 1. row \*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*
3. Slave\_IO\_State: Waiting for master to send event
4. Master\_Host: 192.168.4.52
5. Master\_User: user53
6. Master\_Port: 3306
7. Connect\_Retry: 60
8. Master\_Log\_File: db52.000001
9. Read\_Master\_Log\_Pos: 154
10. Relay\_Log\_File: db53-relay-bin.000003
11. Relay\_Log\_Pos: 315
12. Relay\_Master\_Log\_File: db52.000001
13. Slave\_IO\_Running: Yes
14. Slave\_SQL\_Running: Yes

步骤五：客户端验证配置

1）在主服务器上在主库上授权访问gamedb库的用户

1. [root@db51 ~]# mysql -uroot -p123456
2. mysql> grant all on gamedb.\* to dada@"%" identified by "123456";
3. Query OK, 0 rows affected, 1 warning (0.03 sec)

2）客户端使用授权用户连接主库，建库、表、插入记录

1. [root@room9pc01 ~]# mysql -h192.168.4.51 -udada -p123456
2. Welcome to the MariaDB monitor. Commands end with ; or \g.
3. Your MySQL connection id is 7
4. Server version: 5.7.17-log MySQL Community Server (GPL)
5. Copyright (c) 2000, 2017, Oracle, MariaDB Corporation Ab and others.
6. Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
7. MySQL [(none)]> //验证成功
8. MySQL [(none)]> create database gamedb; //创建测试库
9. Query OK, 1 row affected (0.04 sec)
10. MySQL [(none)]> create table gamedb.t1(id int); //在gamedb下创建t1表
11. Query OK, 0 rows affected (0.17 sec)
12. MySQL [(none)]> insert into gamedb.t1 values(8888); //在t1表中插入数值
13. Query OK, 1 row affected (0.22 sec)

3）客户端使用授权用户连接2台从库时，也可以看到主库上新的库表记录

1. [root@room9pc01 ~]# mysql -h192.168.4.52 -udada -p123456 //验证52主机的状态
2. Welcome to the MariaDB monitor. Commands end with ; or \g.
3. Your MySQL connection id is 10
4. Server version: 5.7.17-log MySQL Community Server (GPL)
5. Copyright (c) 2000, 2017, Oracle, MariaDB Corporation Ab and others.
6. Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
7. MySQL [(none)]> select \* from gamedb.t1; //查询插入的表格
8. +------+
9. | id |
10. +------+
11. | 8888 |
12. +------+
13. 1 row in set (0.00 sec)
14. MySQL [(none)]> exit
15. [root@room9pc01 ~]# mysql -h192.168.4.53 -udada -p123456 //验证53主机的状态
16. Welcome to the MariaDB monitor. Commands end with ; or \g.
17. Your MySQL connection id is 6
18. Server version: 5.7.17 MySQL Community Server (GPL)
19. Copyright (c) 2000, 2017, Oracle, MariaDB Corporation Ab and others.
20. Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
21. MySQL [(none)]> select \* from gamedb.t1;
22. +------+
23. | id |
24. +------+
25. | 8888 |
26. +------+
27. 1 row in set (0.00 sec)

## 3 配置半同步复制模式

### 3.1 问题

- 开启案例1 主库192.168.4.51 半同步复制模式
- 开启案例1 从库192.168.4.52 半同步复制模式
- 开启案例1 从库192.168.4.53 半同步复制模式
- 查看半同步复制模式是否开启

### 3.2 步骤

实现此案例需要按照如下步骤进行。

步骤一：查看是否允许动态加载模块，

1）查看是否允许动态加载模块默认允许

1. mysql> show variables like 'have\_dynamic\_loading';
2. +----------------------+-------+
3. | Variable\_name | Value |
4. +----------------------+-------+
5. | have\_dynamic\_loading | YES |
6. +----------------------+-------+
7. 1 row in set (0.01 sec)

2）命令行加载插件，用户需有SUPER权限

主库上面操作：

1. mysql> INSTALL PLUGIN rpl\_semi\_sync\_master SONAME 'semisync\_master.so';

从库上面操作：

1. mysql> INSTALL PLUGIN rpl\_semi\_sync\_slave SONAME 'semisync\_slave.so';

查看系统库下的表，模块是否安装成功：

1. mysql> SELECT PLUGIN\_NAME, PLUGIN\_STATUS FROM INFORMATION\_SCHEMA.PLUGINS WHERE PLUGIN\_NAME LIKE '%semi%';
2. +----------------------+---------------+
3. | PLUGIN\_NAME | PLUGIN\_STATUS |
4. +----------------------+---------------+
5. | rpl\_semi\_sync\_master | ACTIVE |
6. | rpl\_semi\_sync\_slave | ACTIVE |
7. +----------------------+---------------+
8. 2 rows in set (0.00 sec)

3）启用半同步复制，在安装完插件后，半同步复制默认是关闭的

主库上面执行：

1. mysql> SET GLOBAL rpl\_semi\_sync\_master\_enabled = 1;
2. Query OK, 0 rows affected (0.00 sec)

从库上面执行：

1. mysql> SET GLOBAL rpl\_semi\_sync\_slave\_enabled = 1;
2. Query OK, 0 rows affected (0.00 sec)

查看半同步复制模式是否启用：

1. mysql> show variables like "rpl\_semi\_sync\_%\_enabled";
2. +------------------------------+-------+
3. | Variable\_name | Value |
4. +------------------------------+-------+
5. | rpl\_semi\_sync\_master\_enabled | ON |
6. | rpl\_semi\_sync\_slave\_enabled | ON |
7. +------------------------------+-------+
8. 2 rows in set (0.00 sec)

4）永久启用半同步复制

主库配置

1. [root@master51 ~]# vim /etc/my.cnf
2. [mysqld]
3. plugin-load=rpl\_semi\_sync\_master=semisync\_master.so
4. rpl\_semi\_sync\_master\_enabled=1

从库配置

1. [root@slave52 ~]# vim /etc/my.cnf
2. [mysqld]
3. plugin-load=rpl\_semi\_sync\_slave=semisync\_slave.so
4. rpl\_semi\_sync\_slave\_enabled=1

在高可用架构下，master和slave需同时启动，以便在切换后能继续使用半同步复制

1. [root@master51 ~]# vim /etc/my.cnf
2. [mysqld]
3. plugin-load \
4. ="rpl\_semi\_sync\_master=semisync\_master.so;rpl\_semi\_sync\_slave=semisync\_slave.so"
5. rpl-semi-sync-master-enabled = 1
6. rpl-semi-sync-slave-enabled = 1

```
tips

master:
GRANT SUPER,REPLICATION SLAVE,REPLICATION CLIENT ON *.* TO 'repl'@'%' IDENTIFIED BY 'replpwd';
show master status\G
             File: mysql-bin.000002
         Position: 154

slave:
/etc/my.cnf
添加一段
replicate-wild-ignore-table = mysql.%
replicate-wild-ignore-table = sys.%
replicate-wild-ignore-table = information_schema.%
replicate-wild-ignore-table = performance_schema.%
log_slave_updates=ON

CHANGE MASTER TO 
MASTER_HOST='10.31.240.40',MASTER_USER='repl',MASTER_PASSWORD='replpwd',MASTER_PORT=23306,MASTER_CONNECT_RETRY=5,MASTER_LOG_FILE='mysql-bin.000002',MASTER_LOG_POS=154;
START SLAVE;
show slave status\G
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes

如果开启了gtid，也可以用
CHANGE MASTER TO MASTER_HOST='10.31.240.40',MASTER_USER='repl',MASTER_PASSWORD='replpwd',MASTER_PORT=23306,master_auto_position=1;
```
