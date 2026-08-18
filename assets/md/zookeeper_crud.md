# zookeeper_CRUD

```
Zookeeper节点特性

1.同一级节点 key 名称是唯一的
2.创建节点时，必须要带上全路径
  create /runoob/child/ch01 0
3.session关闭，临时节点自动清除

https://zookeeper.apache.org
https://www.sparksys.top/archives/24

目前在用zookeeper-3.4.14
```

**Leader**领导者 ：Leader 节点负责Zookeeper集群内部投票的发起和决议（一次事务操作），更新系统的状态；同时它也能接收并且响应Client端发送的请求  
**Follower** 跟随者： Follower节点用于接收并且响应Client端的请求，如果是事务操作，会将请求转发给Leader节点，发起投票，参与集群的内部投票  
**Observer** 观察者：Observer节点功能和Follower相同，只是Observer 节点不参与投票过程，只会同步Leader节点的状态。

```
CRUD
cZxid:创建节点时的事务id
pZxid:子节点列表最后一次被修改的事务id
cversion:节点版本号
dataCersion:数据版本号
aclVerson:acl权限版本号

/opt/zookeeper/bin/zkServer.sh start
/opt/zookeeper/bin/zkServer.sh status
./zkCli.sh -timeout 5000  -server 127.0.0.1:2181

查看
ls /
stat /zookeeper
查看节点工作状态
get path：获取指定节点的内容
ls2 path：列出path节点的子节点及详细状态信息

创建
create [-s] [-e] path data acl
[zk: 127.0.0.1:2181(CONNECTED) 15] create -e /node_1/node_1_1 234
Created /node_1/node_1_1
创建了一个临时节点（-e），并且携带数据234
[zk: 127.0.0.1:2181(CONNECTED) 1] create -s /node_1/node_1_1 234
Created /node_1/node_1_10000000001[zk: 127.0.0.1:2181(CONNECTED) 1] create -s /node_1/node_1_1 234
Created /node_1/node_1_10000000001
使用-s参数，创建一个顺序节点

修改
set path data [version]
如果带version的话，只有正确的version才可以修改成功
即version号等于dataVersion
set /runoob 0 6 (6等于dataVersion才能修改成功)

set /node_1 998
dataVersion会随着修改自增

删除
delete path [version]
delete /node_1/node_1_10000000001
rmr /node_1
delete只能删除不包含子节点的节点，如果要删除的节点包含子节点，使用rmr命令
新版本可以deleteall 命令递归删除

quit
退出
```

```
四字命令,四字代码

stat 命令用于查看 zk 的状态信息
echo stat | nc 127.0.0.1 22181

ruok 命令用于查看当前 zkserver 是否启动，若返回 imok 表示正常
echo ruok | nc 127.0.0.1 22181

dump 命令用于列出未经处理的会话和临时节点
echo dump | nc 127.0.0.1 22181

conf 命令用于查看服务器配置
echo conf | nc 127.0.0.1 22181

cons 命令用于展示连接到服务器的客户端信息
echo cons | nc 127.0.0.1 22181

envi 命令用于查看环境变量
echo envi | nc 127.0.0.1 22181

yum install nc
apt install netcat

conf	3.3.0版本引入的。打印出服务相关配置的详细信息。
cons  3.3.0版本引入的。列出所有连接到这台服务器的客户端全部连接/会话详细信息。包括"接受/发送"的包数量、会话id、操作延迟、最后的操作执行等等信息。
crst	3.3.0版本引入的。重置所有连接的连接和会话统计信息。
dump	列出那些比较重要的会话和临时节点。这个命令只能在leader节点上有用。
envi	打印出服务环境的详细信息。
reqs	列出未经处理的请求
ruok	测试服务是否处于正确状态。如果确实如此，那么服务返回"imok"，否则不做任何相应。
stat	输出关于性能和连接的客户端的列表。
srst	重置服务器的统计。
srvr	3.3.0版本引入的。列出连接服务器的详细信息
wchs	3.3.0版本引入的。列出服务器watch的详细信息。
wchc	3.3.0版本引入的。通过session列出服务器watch的详细信息，它的输出是一个与watch相关的会话的列表。
wchp	3.3.0版本引入的。通过路径列出服务器watch的详细信息。它输出一个与session相关的路径。
mntr	3.4.0版本引入的。输出可用于检测集群健康状态的变量列表
```
