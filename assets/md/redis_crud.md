# redis_crud

```
# telnet localhost 6379
> auth igo_pwd
> info
> monitor
> quit

# redis-cli -a igo_pwd
> info
> monitor

20个Client(并发),执行10000个请求,然后强制退出仅输出query/sec值
# redis-benchmark -h 192.168.3.3 -p 7001 -n 10000 -c 20 -q

监控3000行,分析一下热KEY,Top_command,耗时等
redis-cli -p port MONITOR | head -n 1000 | ./redis-faina.py
```

```
v5.4.2.5

增：
localhost:6379> set key:1 value1new
OK
for i in {1..100}; do echo "set igo$i  $i" | redis-cli -c -p 7001; done
redis-cli -p 8001 --pipe </tmp/7001.aof //导入aof

# redis-cli -a igo_pwd
> set igokey value
> expire igokey 60  //1分钟过期

删：
localhost:6379> del key:1
(integer) 1
127.0.0.1:8001> flushall 删除当前节点所有数据
192.168.3.181:8002> flushdb 删除集群数据

改：（改键值和增加使用相同命令set）
localhost:6379> set key:1 newvalue1
OK
SLAVEOF host port
slaveof no one
slot迁移:
集群3个master_slot分别为7001:5461,7002:5461,7004:5462
目标,将所有slot集中到7001
redis-trib.rb reshard --from 19bc53185ae3e103cf4e9ce15772d88f4900256a --to f2ebf72a78f2637eead82c219f2fca737a36a371 --slots 5461 --yes 192.168.1.91:7001
redis-trib.rb reshard --from 99439ae72385633d3c9530a5e4f2afdca22a3e25 --to f2ebf72a78f2637eead82c219f2fca737a36a371 --slots 5462 --yes 192.168.1.91:7001
redis-trib.rb check 192.168.1.91:7001
目标,将7001所有slot重新分配到3个master
./redis-cli --cluster reshard 192.168.0.113:7001 --cluster-from 2fd4ec03e71d6f99718621303dcdf191189df70e --cluster-to ccf4f69e46aa6c3918fb0ebb53172d463eba65e0 --cluster-slots 5461 --cluster-yes
redis-cli --cluster reshard 192.168.0.113:7001 --cluster-from 2fd4ec03e71d6f99718621303dcdf191189df70e --cluster-to 1840c7329348d4f42dc89fa93526f91a4e172d33 --cluster-slots 5462 --cluster-yes

查：
keys * 查询所有的键；
dbsize 查询键总数；
info clients
client list
info stats

查key存活时间
OBJECT IDLETIME igokey
查key过期时间
ttl igokey
# redis-cli -c --cluster call 192.168.3.181:8001 keys \*  查集群所有keys
# redis-cli -c --cluster call 192.168.3.171:7001 dbsize
type key 如果键hello是字符串类型，则返回string；如果键不存在，则返回none
randomkey
集群
cluster nodes
cluster info

$ redis-cli -h localhost -p 6379 ping
127.0.0.1:6379> auth igopassword
127.0.0.1:26379> info replication
127.0.0.1:26379> cluster info
localhost:6379> get key:1
“newvalue1”

连接数：(默认10000，可以通过配置文件修改：maxclients 10000)
示例：
./redis-cli -h 127.0.0.1 -p 6379 -a password info Clients |grep connected_clients
如果大于5000，可能是redis处理速度出现问题，需求排查原因；

拒绝连接数：
示例：
./redis-cli -h 127.0.0.1 -p 6379 -a password info |grep rejected_connections
值理想状态是0。如果大于0,说明连接数超过了maxclients，需要排查原因；

缓存命中率：
./redis-cli -h 127.0.0.1 -p 6379 -a password info |grep keyspace
keyspace_hits:4215 命中数
keyspace_misses:2667 未命中次数
keyspace_hits/(keyspace_hits+keyspace_misses)缓存命中率公式，一般大于90%。

$ ./src/redis-cli -h 192.168.3.112 -p 7801
cluster info ：打印集群的信息
cluster nodes ：列出集群当前已知的所有节点（ node），以及这些节点的相关信息
cluster meet <ip> <port> ：将 ip 和 port 所指定的节点添加到集群当中，让它成为集群的一份子。
cluster forget <node_id> ：从集群中移除 node_id 指定的节点。
cluster replicate <node_id> ：将当前节点设置为 node_id 指定的节点的从节点。
cluster saveconfig ：将节点的配置文件保存到硬盘里面。
```

```
3, redis数据备份恢复
方法一：dump
redis-cli -h 127.0.0.1 -p 6379 -a pwd save
redis-cli -h 127.0.0.1 -p 6379 -a pwd bgsave
两种方法都可以生成dump文件，区别在于bgsave后台运行；
恢复rdb文件只需要将rdb文件恢复到redis安装目录然后启动redis就可以：
redis 127.0.0.1:6379> CONFIG GET dir
1) “dir”
2) “/usr/local/redis/bin”

方法二：aof
redis-cli -h 127.0.0.1 -p 6379 -a pwd bgrewriteaof
方法同dump恢复，注意住过配置文件同时配置了aof和dump，只会以aof方式恢复数据

方法三：主从复制 原理信息参考链接：https://www.cnblogs.com/kismetv/p/9236731.html
redis-server –port 6380 –slaveof 127.0.0.1 6379 开启异步数据复制
SLAVEOF NO ONE可以是当前数据库停止接收其他数据库的同步，转成主Redis数据库，程序连接地址都改为新的redis库IP地址和端口
127.0.0.1:6379> info Replication 同步是否完成可以查看
```

```
redis一个主节点挂了后,立马很快回进入dis-connect状态,大概一个cluster-node-timeout(默认15S)会重新选举新master,挂master对应的slave当选主节点;
```

```
redis禁用危险命令，以下追加到配置文件末尾就行
rename-command keys ""
rename-command flushdb ""
rename-command flushall ""
rename-command config ""
rename-command debug ""
rename-command eval ""
rename-command script ""
rename-command evalsha ""
rename-command shutdown ""
```

```
redis5.0.4建立集群

cluster=c7000
rm -rf /data/$cluster
mkdir -p /data/$cluster/{etc,logs}
for code in {7001..7006};do
mkdir -p /data/$cluster/$code
tee >/data/$cluster/etc/$code.cfg <<EOF
port $code
bind 0.0.0.0
pidfile /var/run/redis_$code.pid
cluster-config-file /data/$cluster/etc/$code.clustercfg
logfile "/data/$cluster/logs/$code.log"
dir /data/$cluster/$code
appendfilename $code.aof
appendonly no
dbfilename $code.rdb
cluster-enabled yes
daemonize yes
protected-mode no
masterauth passwd123
requirepass passwd123
EOF
done

for port in {7001..7006};do redis-server /data/$cluster/etc/$port.cfg;done

redis-cli --cluster create \
192.168.3.111:7001 \
192.168.3.111:7002 \
192.168.3.111:7003 \
192.168.3.111:7004 \
192.168.3.111:7005 \
192.168.3.111:7006 \
--cluster-replicas 1 \
-a passwd123
```
