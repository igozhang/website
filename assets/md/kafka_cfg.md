# kafka_config

```
kafka_2.11

https://kafka.apache.org/documentation/#configuration

# grep -vE '^$|^#' /opt/kafka/config/server.properties

broker.id=1
num.network.threads=3
num.io.threads=8
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
log.dirs=/data/kafka-logs
num.partitions=1
num.recovery.threads.per.data.dir=1
offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1
default.replication.factor=3
log.retention.hours=24
log.segment.bytes=1073741824
log.retention.check.interval.ms=300000
host.name=172.22.240.154
listeners=PLAINTEXT://172.22.240.154:29092
zookeeper.connect=172.22.240.154:22181,172.22.240.155:22181,172.22.240.156:22181
zookeeper.connection.timeout.ms=6000
group.initial.rebalance.delay.ms=0
```

```
配置解读
<kafka_dir>/config/server.properties

# Zookeeper设置。连接多个Zookeeper用逗号隔开。
zookeeper.connect=localhost:2181

# 连接Zookeeper超时时间
zookeeper.connection.timeout.ms=6000

# Kafka 服务器的唯一 ID，每个 Kafka 服务器的 ID 不能重复。
broker.id=0

日志
# 数据达到多少条就将消息刷到磁盘
log.flush.interval.messages=10000

# 多久将累积的消息刷到磁盘，任何一个达到指定值就触发写入
log.flush.interval.ms=1000

日志存储
# 日志存储目录
log.dirs=/tmp/kafka-logs

# 是否允许自动创建Topic，线上建议设置为false
auto.create.topics.enable=true

# 每个Topic默认的分区数
num.partitions=1

# 每个分区的默认副本数，不能大于集群的服务器个数
default.replication.factor=1

# 每个数据目录用来日志恢复的线程数目
num.recovery.threads.per.data.dir=1

# 日志文件保留时长。当超过这个值，最早的segment数据会被删除。
log.retention.hours=168

# 分区存储的最大数据大小。当超过这个值，最早的segment数据会被删除。
#log.retention.bytes=1073741824

# 每个segment文件的大小。当超过这个值，会创建新的segment文件。
log.segment.bytes=1073741824

# 日志保留策略的检查周期。用于定时检查是否应该删除特定文件。
log.retention.check.interval.ms=300000

网络
# Kafka服务器监听的端口
listeners=PLAINTEXT://:9092

# broker处理消息的最大线程数，一般情况下数量为cpu核数
num.network.threads=3

# broker处理磁盘IO的线程数，数值为cpu核数2倍
num.io.threads=8

# socket的发送缓冲区
socket.send.buffer.bytes=102400

# socket的发接收缓冲区
socket.receive.buffer.bytes=102400

# socket请求的最大数值，防止serverOOM
socket.request.max.bytes=104857600

副本
# 当follower落后leader指定数值的消息数，则会将其从ISR（In Sync Replica）列表中移除。默认值是4000。
replica.lag.max.messages=4000
# 该follower超过指定时间未从leader fetch数据，则从ISR中移除。默认值时10000。
replica.lag.time.max.ms=10000

组
# 消费组重新分配消费分区的延时。一般线上设置为3000ms是比较合适的。
# 这个参数的主要效果就是让coordinator推迟空消费组接收到成员加入请求后本应立即开启的rebalance。在实际使用时，假设你预估你的所有consumer组成员加入需要在10s内完成，那么你就可以设置该参数=10000。
group.initial.rebalance.delay.ms=0
```
