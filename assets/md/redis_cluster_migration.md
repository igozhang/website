# redis集群间数据迁移

```
redis5.0.9
centos7.9
```

```
rdb迁移方案
aof迁移方案
```

```
rdb迁移方案
目标集群状态
[root@igo_redisB_node_01 redisB]# redis-cli -p 8001
127.0.0.1:8001> cluster nodes
475f4152864966178ba8a62fea140f079d30792c 192.168.3.183:8005@18005 master - 0 1646669738499 13 connected 10923-16383
84ba48afe20e0b3d4860eb6d85d48131dd4bc2cb 192.168.3.182:8004@18004 slave ecc12347afe9546b1c2b9fa41e55b3d8563e5f25 0 1646669736432 14 connected
bbd4bc6876e3886797748ceba8cb63eada0f0dfc 192.168.3.181:8002@18002 slave 475f4152864966178ba8a62fea140f079d30792c 0 1646669736000 13 connected
ecc12347afe9546b1c2b9fa41e55b3d8563e5f25 192.168.3.181:8001@18001 myself,master - 0 1646669738000 14 connected 0-5460
2fccf39c336c2291d5ed5d34b05b69553dfba316 192.168.3.183:8006@18006 master - 0 1646669736000 10 connected 5461-10922
3f131e350abfcf7275bd58d47835524180c00854 192.168.3.182:8003@18003 slave 2fccf39c336c2291d5ed5d34b05b69553dfba316 0 1646669737458 10 connected

步骤：
7001，7003，7005三个master节点上分别执行save命令生成rdb文件并拷贝到目标集群三个主节点8001，8005，8006，重启三个主节点加载
scp /data/redisA/7001/7001.rdb root@192.168.3.181:/data/redisB/8001/8001.rdb

源节点key数目
127.0.0.1:7001> dbsize
(integer) 66
拷贝完并重启后目标key数目
127.0.0.1:8001> dbsize
(integer) 66

127.0.0.1:7003> dbsize
(integer) 71
127.0.0.1:8005> dbsize
(integer) 71

第三个主节点数据迁移前，igo7记录到slot10750在8006节点，8006重启时间超过30秒被取代成slave节点，导致key不可用
127.0.0.1:8005> get igo7
-> Redirected to slot [10750] located at 192.168.3.183:8006
127.0.0.1:8005> get igo7
(error) CLUSTERDOWN Hash slot not served
```

```
aof方案测试

源集群
igo_redisA_node_01	192.168.3.171	7001,7002
igo_redisA_node_02	192.168.3.172	7003,7004
igo_redisA_node_03	192.168.3.173	7005,7006
主节点7001,7003,7005
[root@igo_redisA_node_01 ~]# redis-cli -p 7001 cluster nodes |grep master
eb7965bbdf695a1d1926c7e143e76097dc34fc59 192.168.3.171:7001@17001 myself,master - 0 1646699323000 1 connected 0-5460
81318fdfa0d72184be8f0e3ddf8c4cc1eb425417 192.168.3.173:7005@17005 master - 0 1646699326000 5 connected 10923-16383
b8a8e3163af66197abdc7f757627654e801c9c8c 192.168.3.172:7003@17003 master - 0 1646699323000 3 connected 5461-10922
[root@igo_redisA_node_01 ~]#

目标集群
igo_redisB_node_01	192.168.3.181	8001,8002
igo_redisB_node_02	192.168.3.182	8003,8004
igo_redisB_node_03	192.168.3.183	8005,8006
主节点8001,8003,8005
[root@igo_redisB_node_01 ~]# redis-cli -p 8001 cluster nodes |grep master
69a447ead34890cfbab4012ce9f245001ce17096 192.168.3.182:8003@18003 master - 0 1646698168665 3 connected 5461-10922
ea7c864898ed48cb034eb67f0e09f79bddb8bcac 192.168.3.181:8001@18001 myself,master - 0 1646698165000 1 connected 0-5460
425bd38a03300872956978843d3a4f04a33adac5 192.168.3.183:8005@18005 master - 0 1646698168000 5 connected 10923-16383
[root@igo_redisB_node_01 ~]#

迁移数据
迁移前数据量(1000)
>>> Calling dbsize
192.168.3.171:7001: 337
192.168.3.173:7005: 321
192.168.3.172:7004: 337
192.168.3.171:7002: 321
192.168.3.173:7006: 342
192.168.3.172:7003: 342
出
[root@igo_redisA_node_01 ~]# scp /data/redisA/7001/7001.aof root@192.168.3.181:/tmp/
[root@igo_redisA_node_02 ~]# scp /data/redisA/7003/7003.aof root@192.168.3.182:/tmp/
[root@igo_redisA_node_03 ~]# scp /data/redisA/7005/7005.aof root@192.168.3.183:/tmp/
入(秒级数据千条数据)
[root@igo_redisB_node_01 ~]# redis-cli -p 8001 --pipe </tmp/7001.aof
[root@igo_redisB_node_02 ~]# redis-cli -p 8003 --pipe </tmp/7003.aof
[root@igo_redisB_node_03 ~]# redis-cli -p 8005 --pipe </tmp/7005.aof
All data transferred. Waiting for the last reply...
Last reply received from server.
errors: 0, replies: 322

迁移后数据量
>>> Calling dbsize
192.168.3.181:8001: 337
192.168.3.182:8004: 337
192.168.3.181:8002: 321
192.168.3.183:8006: 342
192.168.3.182:8003: 342
192.168.3.183:8005: 321
```

```
结论：
rdb比较适用于单节点迁移，集群间迁移如果重启master时间超过30秒会倒是部分slot(key)不可用；
rbd记录的不只是key，还有key对应的slot关系；
```
