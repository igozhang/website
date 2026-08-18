# redis-shake容灾方案，redis5集群安装搭建部署

```
env
cent79
redis5.0.9
redis-shakev2.1.1
https://github.com/alibaba/RedisShake/wiki/tutorial-about-how-to-set-up
https://github.com/alibaba/RedisShake/releases

tag redis_ins

资源清单
安装redis509
拉起集群A
拉起集群B
redis-shake
灾难模拟01
灾难模拟02
评价
```

```
list
首先,我需要两套redis集群
igo_redisA_node_01	192.168.3.171	7001,7002
igo_redisA_node_02	192.168.3.172	7003,7004
igo_redisA_node_03	192.168.3.173	7005,7006

igo_redisB_node_01	192.168.3.181	8001,8002
igo_redisB_node_02	192.168.3.182	8003,8004
igo_redisB_node_03	192.168.3.183	8005,8006

igo_redis-shake_191     192.168.3.191
```

```
安装redis
mkdir -p /{data,opt}/redisA/{7001,7002,7003,7004,7005,7006} 
mkdir -p /{data,opt}/redisB/{8001,8002,8003,8004,8005,8006} 
cd /opt
wget http://download.redis.io/releases/redis-5.0.9.tar.gz
tar -xf redis-5.0.9.tar.gz
mv redis-5.0.9 redis
cd redis
make
make install
```

```
拉起集群A
cluster=c7000
rm -rf /{opt,data}/$cluster
mkdir -p /{opt,data}/$cluster
for code in {7001..7006};do
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
cluster-enabled yes
daemonize yes
protected-mode no
EOF
done

[root@igo_redisA_node_01 redis]# redis-server /opt/redisA/7001/redis7001.conf
[root@igo_redisA_node_01 redis]# redis-server /opt/redisA/7002/redis7002.conf
[root@igo_redisA_node_02 redis]# redis-server /opt/redisA/7003/redis7003.conf
[root@igo_redisA_node_02 redis]# redis-server /opt/redisA/7004/redis7004.conf
[root@igo_redisA_node_03 redis]# redis-server /opt/redisA/7005/redis7005.conf
[root@igo_redisA_node_03 redis]# redis-server /opt/redisA/7006/redis7006.conf

redis-cli --cluster create --cluster-replicas 1 \
192.168.3.171:7001 \
192.168.3.171:7002 \
192.168.3.172:7003 \
192.168.3.172:7004 \
192.168.3.173:7005 \
192.168.3.173:7006

127.0.0.1:7001> set igo helloworld
OK
127.0.0.1:7001> get igo
"helloworld"
```

```
拉起集群B
cluster=redisB
for code in {8001..8006};do
tee >/opt/$cluster/$code/redis$code.conf <<EOF
port $code
bind 0.0.0.0
pidfile /var/run/redis_$code.pid
cluster-config-file /opt/$cluster/$code/cluster$code.cfg
logfile "/data/$cluster/$code/$code.log"
dir /data/$cluster/$code
appendfilename $code.aof
appendonly no
dbfilename $code.rdb
cluster-enabled yes
daemonize yes
protected-mode no
EOF
done

[root@igo_redisB_node_01 redis]# redis-server /opt/redisB/8001/redis8001.conf
[root@igo_redisB_node_01 redis]# redis-server /opt/redisB/8002/redis8002.conf
[root@igo_redisB_node_02 redis]# redis-server /opt/redisB/8003/redis8003.conf
[root@igo_redisB_node_02 redis]# redis-server /opt/redisB/8004/redis8004.conf
[root@igo_redisB_node_03 redis]# redis-server /opt/redisB/8005/redis8005.conf
[root@igo_redisB_node_03 redis]# redis-server /opt/redisB/8006/redis8006.conf
redis-cli --cluster create --cluster-replicas 1 \
-a RedisPwdPwdPIgo \
192.168.3.181:8001 \
192.168.3.181:8002 \
192.168.3.182:8003 \
192.168.3.182:8004 \
192.168.3.183:8005 \
192.168.3.183:8006

127.0.0.1:8001> set hello Bworld
OK
127.0.0.1:8001> get hello
"Bworld"
127.0.0.1:8001>
```

```
redis-shake
安装：
cd /opt
wget https://github.com/alibaba/RedisShake/releases/download/release-v2.1.1-20210903/release-v2.1.1-20210903.tar.gz
tar -xf release-v2.1.1-20210903
mv release-v2.1.1-20210903 redis-shake

配置：
获取主节点：redis-cli -p 8001 cluster nodes
redis-shake.conf配置：(出问题先看日志,大概率是目标集群不为空)
log.file = /data/logs/redis-shake/out.log
source.type = cluster
source.address = 192.168.3.171:7001;192.168.3.173:7005;192.168.3.172:7003
target.type = cluster
target.address = 192.168.3.183:8005;192.168.3.181:8001;192.168.3.182:8003
或者:
tee >redis-shake.conf<<EOF
log.file = /data/logs/redis-shake/out.log
source.type = cluster
source.address = master@10.21.81.28:7001
target.type = cluster
target.address = master@10.21.81.137:8001
EOF

验证：
./redis-shake.linux -conf=redis-shake.conf -type=sync
for i in {1..100}; do echo "set igo$i  $i" | redis-cli -c -p 7001; done
目标机器上查询是否同步：
redis-cli -c --cluster call 192.168.3.181:8001 keys '*'
curl 127.0.0.1:9320/metric

注册服务：
tee >/usr/lib/systemd/system/redis-shake.service <<EOF
[Unit]
Description= redis-shake
After=network.target

[Service]
Type=simple
ExecStart= /opt/redis-shake/redis-shake.linux -conf=redis-shake.conf -type=sync
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

systemctl daemonreload
systemctl start redis-shake
systemctl enable redis-shake
```

```
灾难模拟01
模拟source集群redisA，master7002意外宕机

宕机前：
> cluster nodes
98bf8e70113ccd15d7d74d7e932bfce933d5127d 192.168.3.171:7002@17002 master - 0 1646573888403 7 connected 10923-16383
9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c 192.168.3.171:7001@17001 master - 0 1646573889410 1 connected 0-5460
f81b3e1aa45de87c191f22e15841f5e80e82cdae 192.168.3.172:7004@17004 slave 9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c 0 1646573886389 4 connected
0bdaf47622d7180d2b7db5bf9d4aeb98fb33d761 192.168.3.173:7005@17005 slave 98bf8e70113ccd15d7d74d7e932bfce933d5127d 0 1646573887395 7 connected
619844ef57de910f084d7fcc430d5296231cd10f 192.168.3.172:7003@17003 master - 0 1646573887000 3 connected 5461-10922
b8dac906284ec20afcf337fd544e45ad8eb80ab4 192.168.3.173:7006@17006 myself,slave 619844ef57de910f084d7fcc430d5296231cd10f 0 16465738890006 connected

宕机：
127.0.0.1:7002> shutdown

宕机后：
> cluster nodes
98bf8e70113ccd15d7d74d7e932bfce933d5127d 192.168.3.171:7002@17002 master,fail - 1646574284309 1646574283402 7 disconnected
9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c 192.168.3.171:7001@17001 master - 0 1646574361087 1 connected 0-5460
f81b3e1aa45de87c191f22e15841f5e80e82cdae 192.168.3.172:7004@17004 slave 9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c 0 1646574360000 4 connected
0bdaf47622d7180d2b7db5bf9d4aeb98fb33d761 192.168.3.173:7005@17005 master - 0 1646574359071 8 connected 10923-16383
619844ef57de910f084d7fcc430d5296231cd10f 192.168.3.172:7003@17003 master - 0 1646574360080 3 connected 5461-10922
b8dac906284ec20afcf337fd544e45ad8eb80ab4 192.168.3.173:7006@17006 myself,slave 619844ef57de910f084d7fcc430d5296231cd10f 0 16465743590006 connected

redis-shake 报错退出：
2022/03/06 21:44:48 [WARN] DbSyncer[1] Event:GetFakeSlaveOffsetFail	Id:redis-shake	Warn:EOF
2022/03/06 21:44:48 [PANIC] cannot connect to '192.168.3.171:7002'
[error]: dial tcp 192.168.3.171:7002: connect: connection refused
[stack]:
    2   github.com/alibaba/RedisShake/redis-shake/common/utils.go:77
            github.com/alibaba/RedisShake/redis-shake/common.OpenNetConn
    1   github.com/alibaba/RedisShake/redis-shake/common/utils.go:61
            github.com/alibaba/RedisShake/redis-shake/common.OpenRedisConnWithTimeout
    0   github.com/alibaba/RedisShake/redis-shake/dbSync/syncIncrease.go:77
            github.com/alibaba/RedisShake/redis-shake/dbSync.(*DbSyncer).fetchOffset
        ... ...
[root@igo_redis-shake-191 redis-shake]#

修改redis-shake配置后可以正常连接
# ./redis-shake.linux -conf=redis-shake.conf -type=sync
```

```
灾难模拟02
模拟target集群redisB，master节点8005意外宕机

宕机前：
> cluster nodes
bbd4bc6876e3886797748ceba8cb63eada0f0dfc 192.168.3.181:8002@18002 slave 475f4152864966178ba8a62fea140f079d30792c 0 1646574821000 8 connected
2fccf39c336c2291d5ed5d34b05b69553dfba316 192.168.3.183:8006@18006 myself,slave 3f131e350abfcf7275bd58d47835524180c00854 0 16465748200006 connected
3f131e350abfcf7275bd58d47835524180c00854 192.168.3.182:8003@18003 master - 0 1646574822000 3 connected 5461-10922
475f4152864966178ba8a62fea140f079d30792c 192.168.3.183:8005@18005 master - 0 1646574822657 8 connected 10923-16383
ecc12347afe9546b1c2b9fa41e55b3d8563e5f25 192.168.3.181:8001@18001 master - 0 1646574820000 1 connected 0-5460
84ba48afe20e0b3d4860eb6d85d48131dd4bc2cb 192.168.3.182:8004@18004 slave ecc12347afe9546b1c2b9fa41e55b3d8563e5f25 0 1646574819000 4 connected

宕机：
127.0.0.1:8005> shutdown

报错：
redis-shake
2022/03/06 21:54:29 [PANIC] DbSyncer[2] Event:FlushFail	Id:redis-shake	Error:dial tcp 192.168.3.183:8005: connect: connection refused
[stack]:
    1   github.com/alibaba/RedisShake/redis-shake/dbSync/syncIncrease.go:369
            github.com/alibaba/RedisShake/redis-shake/dbSync.(*DbSyncer).sendTargetCommand.func1
    0   github.com/alibaba/RedisShake/redis-shake/dbSync/syncIncrease.go:423
            github.com/alibaba/RedisShake/redis-shake/dbSync.(*DbSyncer).sendTargetCommand
        ... ...
[root@igo_redis-shake-191 redis-shake]#
```

```
集群A初始化输出：
[root@igo_redisA_node_01 opt]# redis-cli --cluster create --cluster-replicas 1 \
> 192.168.3.171:7001 \
> 192.168.3.171:7002 \
> 192.168.3.172:7003 \
> 192.168.3.172:7004 \
> 192.168.3.173:7005 \
> 192.168.3.173:7006
>>> Performing hash slots allocation on 6 nodes...
Master[0] -> Slots 0 - 5460
Master[1] -> Slots 5461 - 10922
Master[2] -> Slots 10923 - 16383
Adding replica 192.168.3.172:7004 to 192.168.3.171:7001
Adding replica 192.168.3.173:7006 to 192.168.3.172:7003
Adding replica 192.168.3.171:7002 to 192.168.3.173:7005
M: 9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c 192.168.3.171:7001
   slots:[0-5460] (5461 slots) master
S: 98bf8e70113ccd15d7d74d7e932bfce933d5127d 192.168.3.171:7002
   replicates 0bdaf47622d7180d2b7db5bf9d4aeb98fb33d761
M: 619844ef57de910f084d7fcc430d5296231cd10f 192.168.3.172:7003
   slots:[5461-10922] (5462 slots) master
S: f81b3e1aa45de87c191f22e15841f5e80e82cdae 192.168.3.172:7004
   replicates 9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c
M: 0bdaf47622d7180d2b7db5bf9d4aeb98fb33d761 192.168.3.173:7005
   slots:[10923-16383] (5461 slots) master
S: b8dac906284ec20afcf337fd544e45ad8eb80ab4 192.168.3.173:7006
   replicates 619844ef57de910f084d7fcc430d5296231cd10f
Can I set the above configuration? (type 'yes' to accept): yes
>>> Nodes configuration updated
>>> Assign a different config epoch to each node
>>> Sending CLUSTER MEET messages to join the cluster
Waiting for the cluster to join
...
>>> Performing Cluster Check (using node 192.168.3.171:7001)
M: 9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c 192.168.3.171:7001
   slots:[0-5460] (5461 slots) master
   1 additional replica(s)
S: b8dac906284ec20afcf337fd544e45ad8eb80ab4 192.168.3.173:7006
   slots: (0 slots) slave
   replicates 619844ef57de910f084d7fcc430d5296231cd10f
S: f81b3e1aa45de87c191f22e15841f5e80e82cdae 192.168.3.172:7004
   slots: (0 slots) slave
   replicates 9022d3f12139dcde83c1ac6a2bc21da6bf8ea03c
S: 98bf8e70113ccd15d7d74d7e932bfce933d5127d 192.168.3.171:7002
   slots: (0 slots) slave
   replicates 0bdaf47622d7180d2b7db5bf9d4aeb98fb33d761
M: 0bdaf47622d7180d2b7db5bf9d4aeb98fb33d761 192.168.3.173:7005
   slots:[10923-16383] (5461 slots) master
   1 additional replica(s)
M: 619844ef57de910f084d7fcc430d5296231cd10f 192.168.3.172:7003
   slots:[5461-10922] (5462 slots) master
   1 additional replica(s)
[OK] All nodes agree about slots configuration.
>>> Check for open slots...
>>> Check slots coverage...
[OK] All 16384 slots covered.
[root@igo_redisA_node_01 opt]#
```

```
评价：
优点：
兼容  2.8-5.0版本redis；
兼容  单点，哨兵，主备，集群
高效  能做到毫秒级支持，支持大key并发；
监控  支持自监控，curl 127.0.0.1:9320/metric；
同步速度快,秒级同步,166443(16w6k)数据5秒内同步完成;

缺点：
模拟一个从节点加入源redis集群，对集群有侵入性；
不够稳定，出现过自动退出，无报错故障；
dbshake断点续传不可靠,进程启动失败,键值冲突,需要手动清空目标redis,才能重新开始同步,有风险;
发生主备切换需要手动切换配置；
暂不支持断点续传，网络抖动及主备切换会造成全量重新同步；
不支持源端的slot_move操作

模式：
恢复restore：将RDB文件恢复到目的redis数据库。
备份dump：将源redis的全量数据通过RDB文件备份起来。
解析decode：对RDB文件进行读取，并以json格式解析存储。
同步sync：支持源redis和目的redis的数据同步，支持全量和增量数据的迁移，支持从云下到阿里云云上的同步，也支持云下到云下不同环境的同步，支持单节点、主从版、集群版之间的互相同步。需要注意的是，如果源端是集群版，可以启动一个RedisShake，从不同的db结点进行拉取，同时源端不能开启move slot功能；对于目的端，如果是集群版，写入可以是1个或者多个db结点。
同步rump：支持源redis和目的redis的数据同步，仅支持全量的迁移。采用scan和restore命令进行迁移，支持不同云厂商不同redis版本的迁移
```
