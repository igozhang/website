# kafka_ins

## 概念

Broker Kafka节点  
Topic 主题，用来承载消息  
Partition 分区，用于主题分片存储  
Producer 生产者，向主题发布消息的应用  
Consumer 消费者，从主题订阅消息的应用  
Consumer Group 消费者组，由多个消费者组成

## 下载

http://kafka.apache.org/downloads

```
curl -OL https://archive.apache.org/dist/kafka/2.4.1/kafka_2.11-2.4.1.tgz
ln -s kafka_2.11-2.4.1 kafka211

env:
Centos7.9
Kafka2.11
```

## 启停单节点

1. 先安装java环境

```
yum search java-11-openjdk
yum -y install java-11-openjdk java-11-openjdk-devel
```

2. 再起zookeeper，Kafka 在设计上严重依赖 ZooKeeper 提供的分布式协调服务。ZooKeeper 负责存储和管理 Kafka 集群的元数据和配置信息。  
   2.1 集群管理：ZooKeeper 管理 Kafka Broker 的注册、上下线状态、心跳检测及故障转移。当 Broker 发生变化时，Zookeeper 参与 Broker 中 Leader 的选举过程。  
   2.2 Topic 和 Partition 管理：所有与 Topic 相关的元数据，如 Topic 创建、删除、分区数量调整，以及 Partition 的 Leader 选举和分配，都在 ZooKeeper 上进行。  
   2.3 Consumer 管理：Kafka 消费者（Consumer）也通过 ZooKeeper 注册消费组信息，包括消费者所消费的 Partition 列表，以及在消费组成员发生变化时触发的 rebalance 过程。

```
nohup bin/zookeeper-server-start.sh config/zookeeper.properties &
```

3. 启停kafka

```
启动
nohup bin/kafka-server-start.sh config/server.properties &
停止
bin/kafka-server-stop.sh config/server.properties
```

4. 创建主题并生产消费消息

```
使用名称igotest创建一个分区一个备份的topic:
bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic igotest

生产消息：
bin/kafka-console-producer.sh --broker-list localhost:9092 --topic igotest
>mes1
>mes2
>igomes3
>^C^C 无法优雅退出，需要两个ctrl+C

消费消息:
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic igotest --from-beginning
mes1
mes2
igomes3
^CProcessed a total of 6 messages 无法优雅退出，需要一个ctrl+C
```

## 集群部署

### 建目录

```
mkdir -p /data/{kafka,zookeeper}/{1,2,3}/{config,data,logs}
检查
tree -L 3 /data
```

### 部署zookeeper集群

配置zookeeper\_id

```
for i in {1,2,3};do echo "$i" > /data/zookeeper/$i/data/myid;done
检查
for i in {1,2,3};do cat /data/zookeeper/$i/data/myid;done
```

配置zookeeper

```
pp=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')
for i in {1,2,3};do
tee >/data/zookeeper/$i/config/zoo.cfg <<EOF
tickTime=2000
initLimit=10
syncLimit=50
dataDir=/data/zookeeper/$i/data
dataLogDir=/data/zookeeper/$i/logs
clientPort=218$i
server.1=$pp:1188:1288
server.2=$pp:2188:2288
server.3=$pp:3188:3288
EOF
done

###ZooKeeper使用的端口分别是2181,2888和3888。 2181端口：主要用于对客户端提供服务;2888集群内通讯,3888主要用于选举leader
检查
for i in {1,2,3};do cat /data/zookeeper/$i/config/zoo.cfg;done

启动
nohup /opt/kafka211/bin/zookeeper-server-start.sh /data/zookeeper/1/config/zoo.cfg &
nohup /opt/kafka211/bin/zookeeper-server-start.sh /data/zookeeper/2/config/zoo.cfg &
nohup /opt/kafka211/bin/zookeeper-server-start.sh /data/zookeeper/3/config/zoo.cfg &

验证
ss -tlnp |grep 218
ss -tlnp |grep 88
/opt/kafka211/bin/zookeeper-shell.sh 127.0.0.1:2181
/opt/kafka211/bin/zookeeper-shell.sh 127.0.0.1:2182
/opt/kafka211/bin/zookeeper-shell.sh 127.0.0.1:2183
zookeeper-shell.sh命令交互模式敲任意字母符号会提示你正确的用法，主要用来查看get，列表ls,删除rmr，stat等
```

### 部署kafka集群

配置kafka

```
pp=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')
for i in {1,2,3};do
tee >/data/kafka/$i/config/kafka.cfg <<EOF
broker.id=$i
listeners=PLAINTEXT://$pp:909$i
advertised.listeners=PLAINTEXT://$pp:909$i
log.dirs=/data/kafka/$i/logs
zookeeper.connect=$pp:2181,$pp:2182,$pp:2183
EOF
done

检查配置
for i in {1,2,3};do echo $i'##########kafka_cfg';cat /data/kafka/$i/config/kafka.cfg;done

启动
/opt/kafka211/bin/kafka-server-start.sh -daemon  /data/kafka/1/config/kafka.cfg
/opt/kafka211/bin/kafka-server-start.sh -daemon  /data/kafka/2/config/kafka.cfg
/opt/kafka211/bin/kafka-server-start.sh -daemon  /data/kafka/3/config/kafka.cfg

验证
ss -tlnp |grep 909
连接第一个zookeeper节点创建igotest_topic
# /opt/kafka211/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic igotest
连接第二个kafka节点生产消息
# /opt/kafka211/bin/kafka-console-producer.sh --broker-list 10.10.8.213:9092 --topic igotest
>igomessage1
>igomessage2
>^C^C
连接第三个kafka节点消费消息
# /opt/kafka211/bin/kafka-console-consumer.sh --bootstrap-server 10.10.8.213:9093 --topic igotest --from-beginning
igomessage1
igomessage2
^C
```
