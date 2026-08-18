# kafka高可用失效问题排查

```
env
CentOS Linux release 7.6.1810
kafka_2.11-2.3.0
zookeeper-3.4.14
```

```
问题现象
3节点kafka挂掉一个节点之后部分topic无法消费，报错：
ClientResponse(receivedTimeMs=151589656205, disconnected=false, request=ClientRequest(expectResponse=true, callback=org.apache.kafka.clients.consumer.internals.ConsumerNetworkClient$RequestFutureCompletionHandler@488f3dd1, request=RequestSend(header={api_key=10,api_version=0,correlation_id=30281,client_id=consumer-1}, body={group_id=testGroup}), createdTimeMs=1515897558800, sendTimeMs=1515897561104), 
responseBody=**{error_code=15,coordinator={node_id=-1,host=,port=-1}})**

查看topic __consumer_offsets,确认所有Partition只有一份副本，对应leader主节点挂掉后，无法消费数据；
# /opt/kafka/bin/kafka-topics.sh --zookeeper localhost:22181 --describe --topic __consumer_offsets
Topic:__consumer_offsets	PartitionCount:50	ReplicationFactor:1	Configs:segment.bytes=104857600,cleanup.policy=compact,compression.type=producer
	Topic: __consumer_offsets	Partition: 0	Leader: 2	Replicas: 2	Isr: 2
	Topic: __consumer_offsets	Partition: 1	Leader: 3	Replicas: 3	Isr: 3
	Topic: __consumer_offsets	Partition: 2	Leader: 1	Replicas: 1	Isr: 1
```

```
现有配置
cat /opt/kafka/config/server.propertie

############################# Internal Topic Settings  #############################
# The replication factor for the group metadata internal topics "__consumer_offsets" and "__transaction_state"
# For anything other than development testing, a value greater than 1 is recommended for to ensure availability such as 3.
offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1
default.replication.factor=3
```

```
问题修复：
1.
将consumer_offsets副本数设置为3
offsets.topic.replication.factor=3
2.
关闭kafka集群，从zk删除kafka consumer_offsets topic （从kafka删除会导致集群彻底挂掉）
./zkCli.sh -server localhost:22181
rmr /brokers/topics/__consumer_offsets
3.
启动集群
```
