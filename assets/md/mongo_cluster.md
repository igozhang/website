# mongo_cluster

mongodb集群搭建

env  
centos7.9  
mongo v3.4.10  
环境说明：  
192.168.3.10 mo10  
192.168.3.11 mo11  
192.168.3.12 mo12  
mongos3节点分别使用10,11,12机器的20000端口；  
configrepl3节点分别使用21000端口  
shard1 22001  
shard2 22002  
shard3 22003  
参考资料：https://bbs.huaweicloud.com/blogs/289539

## 配置confrepl集群

cat>/data/mongodb/mongoconf/conf/mongoconf.conf<<EOF  
dbpath=/data/mongodb/mongoconf/data  
logpath=/data/mongodb/mongoconf/log/mongoconf.log  
logappend=true  
port = 21000  
maxConns = 1000 #链接数  
journal = true #日志开启  
journalCommitInterval = 200  
fork = true #后台执行  
syncdelay = 60  
oplogSize = 1000  
configsvr = true #配置服务器  
replSet=replconf #config server配置集replconf  
EOF  
\*\* 启动配置集群 \*\*  
/opt/mongodb-linux-x86\_64-rhel62-3.4.10/bin/mongod -f /data/mongodb/mongoconf/conf/mongoconf.conf

mongo mo10:21000  
config = {\_id:"replconf",members:[  
{\_id:0,host:"192.168.3.10:21000"},  
{\_id:1,host:"192.168.3.11:21000"},  
{\_id:2,host:"192.168.3.12:21000"},]  
}  
rs.initiate(config); #初始化集群  
rs.status(); #查看状态

## 配置shard集群

### shard1

cat >/data/mongodb/shard1/conf/shard.conf <<EOF  
dbpath=/data/mongodb/shard1/data  
logpath = /data/mongodb/shard1/log/shard1.log  
port = 22001  
logappend = true  
nohttpinterface = true  
fork = true  
oplogSize = 4096  
journal = true  
smallfiles=true  
shardsvr=true #shard服务器  
replSet=shard1 #副本集名称shard1  
EOF  
\*\* 启动 \*\*  
mongod -f /data/mongodb/shard1/conf/shard.conf  
\*\* 在mo10初始化 \*\*  
use admin  
config = {\_id:"shard1",members:[  
{\_id:0,host:"192.168.3.10:22001"},  
{\_id:1,host:"192.168.3.11:22001",arbiterOnly:true},  
{\_id:2,host:"192.168.3.12:22001"},]  
}  
rs.initiate(config);  
rs.status();

### shard2

cat >/data/mongodb/shard2/conf/shard.conf <<EOF  
dbpath=/data/mongodb/shard2/data  
logpath = /data/mongodb/shard2/log/shard2.log  
port = 22002  
logappend = true  
nohttpinterface = true  
fork = true  
oplogSize = 4096  
journal = true  
smallfiles=true  
shardsvr=true  
replSet=shard2  
EOF  
\*\* 启动 \*\*  
mongod -f /data/mongodb/shard2/conf/shard.conf  
\*\* 在mo11初始化 \*\*  
mongo 192.168.3.11:22002  
use admin  
config = {\_id:"shard2",members:[  
{\_id:0,host:"192.168.3.10:22002"},  
{\_id:1,host:"192.168.3.11:22002"},  
{\_id:2,host:"192.168.3.12:22002",arbiterOnly:true},]  
}  
rs.initiate(config);  
rs.status();

### shard3

cat >/data/mongodb/shard3/conf/shard.conf <<EOF  
dbpath=/data/mongodb/shard3/data  
logpath = /data/mongodb/shard3/log/shard3.log  
port = 22003  
logappend = true  
nohttpinterface = true  
fork = true  
oplogSize = 4096  
journal = true  
smallfiles=true  
shardsvr=true  
replSet=shard3  
EOF  
\*\* 启动 \*\*  
mongod -f /data/mongodb/shard3/conf/shard.conf

初始化  
mongo 192.168.3.12:22003  
use admin  
config = {\_id:"shard3",members:[  
{\_id:0,host:"192.168.3.10:22003", arbiterOnly:true},  
{\_id:1,host:"192.168.3.11:22003"},  
{\_id:2,host:"192.168.3.12:22003"},]  
}  
rs.initiate(config);  
rs.status();

## 配置mongos路由集群

cat >/data/mongodb/mongos/conf/mongos.conf<<EOF  
logpath=/data/mongodb/mongos/log/mongos.log  
logappend=true  
port = 20000  
maxConns = 1000  
configdb=replconf/192.168.3.10:21000,192.168.3.11:21000,192.168.3.12:21000 #制定config server集群  
fork = true  
EOF  
\*\* 启动mongos \*\*  
mongos -f /data/mongodb/mongos/conf/mongos.conf  
\*\* 设置mongos添加分片 \*\*  
mongo 192.168.3.10:20000  
use admin  
db.runCommand({addshard:"shard1/192.168.3.10:22001,192.168.3.11:22001,192.168.3.12:22001"})  
db.runCommand({addshard:"shard2/192.168.3.10:22002,192.168.3.11:22002,192.168.3.12:22002"})  
db.runCommand({addshard:"shard3/192.168.3.10:22003,192.168.3.11:22003,192.168.3.12:22003"})

## 测试验证

use admin  
建库  
db.runCommand( { enablesharding :"igodb"});  
指定需要分片的表和ID  
db.runCommand( { shardcollection : "igodb.table1",key : {\_id:"hashed"} } )  
查看分片信息  
db.runCommand({listshards:1});  
写入测试数据  
use igodb;  
for (var i = 1; i <= 10000; i++) db.table1.save({\_id:i,"test1":"testval1"});  
查看分片情况  
db.table1.stats()

## tips

关闭重置mongo\_shard集群：  
使用localhost(127.0.0.1)登录  
mongo localhost:22002  
db.adminCommand({ "shutdown" : 1,force: true })

rm -rf /data/mongodb/shard2/data/\*

> /data/mongodb/shard2/log/shard2.log

export PATH=/usr/bin:/usr/sbin:/bin:/sbin:/usr/X11R6/bin  
export PATH=/opt/mongodb-linux-x86\_64-rhel62-3.4.10/bin:$PATH
