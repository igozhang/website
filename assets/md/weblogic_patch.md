# weblogic打补丁

```
WLS 10.3.4之前的版本是没有PSU补丁的。在10.3.4到12.1.1之间的PSU补丁，都是累积补丁，10.3.4.0.5包含了10.3.4.0.4，在打10.3.4.0.5之前要卸载10.3.4.0.4，否则会有冲突。这个问题在WLS 12.1.2之后得以解决，在应用PSU的时候，会自动卸载原来的补丁，与数据库的补丁方式逐渐一致。如何下载补丁和解决升级中的问题，参考文档 1306505.1 / 1470197.1等。

附几种weblogic打补丁的方式示例

示例1：在12c中，与数据库的补丁方式类似

https://updates.oracle.com/Orion/Services/download?type=readme&aru=23842278
$ export ORACLE_HOME=/data/app/weblogic/Middleware
$ opatch lspatches
$ unzip -d PATCH_TOP p31960985_122140_Generic.zip
$ cd PATCH_TOP/31960985
$ opatch apply

$ opatch lspatches
31960985;WLS PATCH SET UPDATE 12.2.1.4.201001

OPatch succeeded.

回滚:
$ cd PATCH_TOP/31960985
$ opatch rollback -id 31960985

示例2：在10.3.4 版本中
cd /home/weblogic/Oracle/Middleware/utils/bsu/

./bsu.sh -install -patch_download_dir=/home/weblogic/media/p16088411_1035_Generic -patchlist=L5TD -prod_dir=/home/weblogic/Oracle/Middleware/wlserver_10.3/

 

示例3：在10 版本中

#vi bsu.sh

#!/bin/sh
JAVA_HOME="/usr/java5"
"$JAVA_HOME/bin/java" -jar -Xms1024m -Xmx2048m patch-client.jar $*
```

```
近期的版本归属：

Oracle WebLogic Server12c 指的是12.1.x 12.2.x
Oracle WebLogic Server11gR1 指的是10.3.1之后的版本
Oracle WebLogic Server10gR3 指的是10.3

官方名称归属：

Oracle Cloud Application Foundation 12c Releases
Oracle WebLogic Server 12.1.x
Oracle WebLogic Server 12.2.x

Oracle Cloud Application Foundation 11gR1 Releases
Oracle WebLogic Server 11g R1(10.3.1 and higher)*

Application Grid 10gR3 Releases
Oracle WebLogic Server 10.3 *
```

```
踩坑12.2.1.4.0
1: optach 报错:
The OPatch version is not applicable for current OUI version
原因,机器上同时安装了weblogic和oracle数据库客户端,设置了ORACLE_HOME为数据库的:
#export ORACLE_HOME=$ORACLE_BASE/product/client_64
改为weblogic的:(打完补丁记得回退)
export ORACLE_HOME=/data/app/weblogic/Middleware

2: Exception occured :	fuser could not be located
yum install -y psmisc

3. "The patch type is invalid in patch location".
zip包权限问题(之前用root解压的)
改权限,不能用root安装
```

```
备份bak
备份weblogic 备份oracle
备份配置文件及连接文件oraloc

升级opatch
java -jar  /data/igo/0826/WLS_SPB_12.2.1.4.250707/tools/opatch/generic/6880880/opatch_generic.jar  -silent oracle_home=/data/app/weblogic/Middleware

升级weblogic_patch
先进入补丁目录
cd /data/igo/0826/WLS_SPB_12.2.1.4.250707/binary_patches

如果报错 fuser could not be located
就安装 yum install -y  psmisc
/data/app/weblogic/Middleware/OPatch/opatch napply -oh /data/app/weblogic/Middleware -phBaseFile /data/igo/0826/WLS_SPB_12.2.1.4.250707/binary_patches/linux64_patchlist.txt

如果需要回退：
/data/app/weblogic/Middleware/OPatch/opatch nrollback -oh /data/app/weblogic/Middleware -idFile rollback_patchlist.txt
```
