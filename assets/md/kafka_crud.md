# Kafka_CRUD

```
Kafka 是一个分布式流处理平台，支持将数据从生产者发送到消费者。指定 key 消费是指消费者仅仅消费具有特定 key 的消息。

在 Kafka 中，每条消息都有一个 key 和一个值，key 用于确定消息的分区，由生产者设置。而消费者可以通过读取具有特定 key 的消息，来实现指定 key 消费。
```

```
kafka_2.11
启停服务
bin/kafka-server-start.sh -daemon config/server.properties
bin/kafka-server-stop.sh

topic
查看当前集群所有topic
./bin/kafka-topics.sh --list --zookeeper 127.0.0.1:22181/
查看topic详细信息
./bin/kafka-topics.sh --list --zookeeper 127.0.0.1:22181/igo_test_topic
创建topic
./bin/kafka-topics.sh --create --zookeeper 127.0.0.1:22181 --replication-factor 1 --partitions 1 --topic igo_test_topic
删除
./bin/kafka-topics.sh --delete --zookeeper 127.0.0.1:22181/ --topic igo_test_topic

生产消息（Ctrl+c退出）
./bin/kafka-console-producer.sh --broker-list 172.22.50.155:29092 --topic igo_test_topic
>
>hello
>igo
>new
>kafka4
>world5

消费消息
# bin/kafka-console-consumer.sh --bootstrap-server 172.22.50.155:29092 --topic igo_test_topic --from-beginning

hello
igo
new
kafka4
world5
```

