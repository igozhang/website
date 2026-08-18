# redis配置文件

```
配置解读
https://redis.io/topics/cluster-tutorial
```

**redis主从**配置

```
v504
cat /opt/redis/etc/redis.conf
daemonize yes

#slaveof 172.25.240.201 26379
#masterauth password123

pidfile /opt/redis/var/redis-server.pid
port 26379
tcp-backlog 511
bind 0.0.0.0
timeout 0
tcp-keepalive 0
loglevel notice
logfile /opt/redis/var/redis-server.log
databases 16
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /opt/redis/data
slave-serve-stale-data yes
slave-read-only yes
repl-diskless-sync no
repl-diskless-sync-delay 5
repl-disable-tcp-nodelay no
slave-priority 100
requirepass viuRDS2793
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
aof-load-truncated yes
lua-time-limit 5000
slowlog-log-slower-than 10000
slowlog-max-len 128
latency-monitor-threshold 0
notify-keyspace-events ""
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-entries 512
list-max-ziplist-value 64
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000
activerehashing yes
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60
hz 10
aof-rewrite-incremental-fsync yes
```

redis集群**配置**

```
daemonize yes

cluster-enabled yes
cluster-config-file node-26379.conf
cluster-node-timeout 8000
masterauth password123

pidfile /opt/redis/var/redis-26379.pid
port 26379
tcp-backlog 511
bind 0.0.0.0
timeout 0
tcp-keepalive 0
loglevel notice
logfile /opt/redis/var/redis-26379.log
databases 16
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump-26379.rdb
dir /opt/redis/data
slave-serve-stale-data yes
slave-read-only yes
repl-diskless-sync no
repl-diskless-sync-delay 5
repl-disable-tcp-nodelay no
slave-priority 100
requirepass viuRDS2793
appendonly yes
appendfilename "appendonly-26379.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
aof-load-truncated yes
lua-time-limit 5000
slowlog-log-slower-than 10000
slowlog-max-len 128
latency-monitor-threshold 0
notify-keyspace-events ""
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-entries 512
list-max-ziplist-value 64
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000
activerehashing yes
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60
hz 10
aof-rewrite-incremental-fsync yes
```

```
tips
redis内存清理，redis内存淘汰LRU算法
https://redis.io/topics/lru-cache

好几种算法：主要有超时删除，随机删除，使用频率删除
volatile-lru　  　设定超时时间的数据中，删除最不常用使用的数据
allkeys-lru　　 　查询所有的key中最不常使用的数据进行删除，这是应用最广泛的策略。
volatile-random 　在已经设定了超时的数据中随机删除
allkeys-random　　查询所有的key，之后随机删除
volatile-ttl　　　查询全部设定超时时间的数据，之后排序，将马上要国企的数据进行删除
noeviction　　　　如果设置为该属性，则不会进行删除操作，在内存溢出时报错返回
volatile-lfu　　　从所有配置了超时时间的键中删除使用频率最少的键
allkeys-lfu　　　从所有键中删除使用频率最少的键
```
