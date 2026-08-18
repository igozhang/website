# dbclient_ins

```
mysql_client-install
env
CentOS Linux release 7.6.1810 (Core)
Docker version 1.13.1, build 7d71120/1.13.1
mysql  Ver 14.14 Distrib 5.7.41, for Linux (x86_64)

# docker run -itd --name mysql57-igo -p3306:3306 -e MYSQL_ROOT_PASSWORD=123456 mysql:5.7
# rpm -ivh https://repo.mysql.com//mysql57-community-release-el7-11.noarch.rpm
# yum search mysql-community
# rpm --import https://repo.mysql.com/RPM-GPG-KEY-mysql-2022
# yum -y install mysql-community-client
# mysql -h 127.0.0.1 -uroot -p123456
```

```
Oracle_client-install
https://blog.csdn.net/chengyuqiang/article/details/80406159

rpm -ivh oracle-instantclient11.2-basic-11.2.0.4.0-1.x86_64.rpm  
rpm -ivh oracle-instantclient11.2-devel-11.2.0.4.0-1.x86_64.rpm    
rpm -ivh oracle-instantclient11.2-sqlplus-11.2.0.4.0-1.x86_64.rpm

useradd oracle
passwd oracle
su - oracle

export  ORACLE_HOME=/usr/lib/oracle/12.2/client64
export  TNS_ADMIN=$ORACLE_HOME/network/admin
export  LD_LIBRARY_PATH=$ORACLE_HOME/lib 
export  PATH=$ORACLE_HOME/bin:$PATH
export  NLS_LANG=AMERICAN_AMERICA.ZHS16GBK

sqlplus DB_USER/DB_PASSWD@oracle_db_01.igozhang.cn:1521/service_name
```
