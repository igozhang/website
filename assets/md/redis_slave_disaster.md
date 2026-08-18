# redis_slaveof_Cluster容灾方案

```
CentOS Linux release 7.8.2003
redis5.0.5
概述:
src集群A,dst集群B通过集群B上的三个同主机slave节点实现数据的快速恢复

aof方式
rdb方式
```

```
aof方式
集群A,master7001,7002,7003
6575a0715c0ab46db6a4093f79337e00c028e060 10.21.81.28:7003@17003 master - 0 1646703504000 3 connected 10923-16383
7cd7aaf9f08acad113b0dd77007e628a8c4778c8 10.21.81.28:7002@17002 master - 0 1646703504000 2 connected 5461-10922
4698b6d89c59343e3871320ba716836a79e80e70 10.21.81.28:7001@17001 myself,master - 0 1646703503000 1 connected 0-5460

集群B,master8001,8002,8003
6cc61ee16f7ffd747479fbd89aaa4426998ddcbd 10.21.81.29:8003@18003 master - 0 1646703615816 3 connected 10923-16383
8a7bae74075cd82d73689aa73b984ba9c90e9936 10.21.81.29:8001@18001 myself,master - 0 1646703612000 1 connected 0-5460
d2e752fc949cbbfc4788ebca812d90d66f861294 10.21.81.29:8002@18002 master - 0 1646703615000 2 connected 5461-10922

中转slave节点:
cluster=c9000
rm -rf /{opt,data}/$cluster
mkdir -p /{opt,data}/$cluster
for code in {9001..9003};do
mkdir -p /data/$cluster/$code
tee >/opt/$cluster/cfg.$code <<EOF
port $code
bind 0.0.0.0
pidfile /var/run/redis_$code.pid
cluster-config-file /opt/$cluster/cluster$code.cfg
logfile "/data/$cluster/$code.log"
dir /data/$cluster/$code
appendfilename $code.aof
appendonly no
dbfilename $code.rdb
cluster-enabled no
daemonize yes
protected-mode no
EOF
done

8007,8008,8009单节点
# redis-cli -p 8007 slaveof 10.21.81.28 7001
OK
# redis-cli -p 8008 slaveof 10.21.81.28 7002
OK
# redis-cli -p 8009 slaveof 10.21.81.28 7003
OK

集群A写100条数据:
>>> Calling dbsize
10.21.81.28:7001: 35
10.21.81.28:7003: 28
10.21.81.28:7002: 37
10.21.81.28:7005: 35
10.21.81.28:7006: 37
10.21.81.28:7004: 28

中转数据
查看:
[root@cnabdabvdc03-81-29 ~]# redis-cli -p 8007 dbsize
(integer) 35
[root@cnabdabvdc03-81-29 ~]# redis-cli -p 8008 dbsize
(integer) 37
[root@cnabdabvdc03-81-29 ~]# redis-cli -p 8009 dbsize
(integer) 28
导入aof:
redis-cli -p 8001 --pipe </data/redisB/8007/8007.aof
redis-cli -p 8002 --pipe </data/redisB/8008/8008.aof
redis-cli -p 8003 --pipe </data/redisB/8009/8009.aof

redisB]# cp 8007.aof /data/redisB/8001/8001.aof
redisB]# cp 8008.aof /data/redisB/8002/8002.aof
redisB]# cp 8009.aof /data/redisB/8003/8003.aof
redis-cli -p 8001 shutdown
redis-cli -p 8002 shutdown
redis-cli -p 8003 shutdown
redis-server /opt/redisB/8001.conf
redis-server /opt/redisB/8002.conf
redis-server /opt/redisB/8003.conf

验证:可读可写,通过!
igo-redis# redis-cli -p 8003 -c
127.0.0.1:8003> get igo99
-> Redirected to slot [5746] located at 10.21.81.29:8002
"99"
10.21.81.29:8002> set igo101 101
-> Redirected to slot [13756] located at 10.21.81.29:8003
OK
10.21.81.29:8003>
```

```
rdb方案:

资源清单同aof方案;
起slave:
# redis-cli -p 8007 slaveof 10.21.81.28 7001
# redis-cli -p 8008 slaveof 10.21.81.28 7002
# redis-cli -p 8009 slaveof 10.21.81.28 7003

查slave日志并看key数量确定同步完成
# redis-cli -p 8007 dbsize
(integer) 35
[root@cnabdabvdc03-81-29 ~]# redis-cli -p 8008 dbsize
(integer) 37
[root@cnabdabvdc03-81-29 ~]# redis-cli -p 8009 dbsize
(integer) 28

覆盖到容灾:
cp 8007/dump.rdb 8001/
cp 8008/dump.rdb 8002/
cp 8009/dump.rdb 8003/

redis-cli -p 8001 shutdown
redis-server /opt/redisB/cfg.8001
redis-cli -p 8002 shutdown
redis-server /opt/redisB/cfg.8002
redis-cli -p 8003 shutdown
redis-server /opt/redisB/cfg.8003

验证:同步完成,可写可读,通过:
>>> Calling dbsize
10.21.81.29:8001: 35
10.21.81.29:8003: 28
10.21.81.29:8006: 28
10.21.81.29:8004: 35
10.21.81.29:8002: 37
10.21.81.29:8005: 37

127.0.0.1:8001> get igo99
-> Redirected to slot [5746] located at 10.21.81.29:8002
"99"
10.21.81.29:8002> set igo101 101
-> Redirected to slot [13756] located at 10.21.81.29:8003
OK
10.21.81.29:8003>
```

```
优点:
方案简单清楚无三方依赖,实施难度低;
利用原生主从复制机制,高效保存数据;

同步速度快,秒级同步,166445(16w6k)数据5秒内同步完成;

缺点:
对集群有侵入;
切换过程需手动介入;
```
