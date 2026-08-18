# rabbitmq_ins

## ENV

Centos7.9  
Docker version 1.13.1, build 7d71120/1.13.1  
默认端口及作用  
NameServer 端口是9876  
Broker 端口是10911（主节点）和10909（从节点）对Producer和Consumer提供服务的端口号，默认为10911  
Web控制台端口，默认是8080

```
tee <<EOF>/etc/docker/daemon.json
{
   "registry-mirrors": [
       "https://mirror.ccs.tencentyun.com"
  ]
}
EOF
```

## 安装部署rabbitmq

### 部署nameserver

```
docker pull apache/rocketmq:4.9.3
docker network create rocketmq
docker run -d --name rocketmq-nameserver --network rocketmq -p 9875:9876 -v /home/docker/mount/rocketmq/namesrv/logs:/root/logs -v /home/docker/mount/rocketmq/namesrv/store:/root/store apache/rocketmq:4.9.3 sh mqnamesrv

docker update --restart=always rocketmq-nameserver(docker_id)
```

### 部署broker\_master

```
复制nameserver中的broker配置
mkdir -p /home/docker/mount/rocketmq/broker/broker-a/
docker cp rocketmq-nameserver:/home/rocketmq/rocketmq-4.9.3/conf /home/docker/mount/rocketmq/broker/broker-a/

tee <<EOF>/home/docker/mount/rocketmq/broker/broker-a/conf/broker.conf
# 所属集群名称，如果节点较多可以配置多个
brokerClusterName = DefaultCluster
#broker名称，master和slave使用相同的名称，表明他们的主从关系
brokerName = broker-a
#0表示Master，大于0表示不同的slave
brokerId = 0
#表示几点做消息删除动作，默认是凌晨4点
deleteWhen = 04
#在磁盘上保留消息的时长，单位是小时
fileReservedTime = 48
#有三个值：SYNC_MASTER，ASYNC_MASTER，SLAVE；同步和异步表示Master和Slave之间同步数据的机制；
brokerRole = SYNC_MASTER
#刷盘策略，取值为：ASYNC_FLUSH，SYNC_FLUSH表示同步刷盘和异步刷盘；SYNC_FLUSH消息写入磁盘后才返回成功状态，ASYNC_FLUSH不需要；
flushDiskType = SYNC_FLUSH
# 设置broker节点所在服务器的ip地址（**这个非常重要,主从模式下，从节点会根据主节点的brokerIP2来同步数据，如果不配置，主从无法同步，brokerIP1设置为自己外网能访问的ip，服务器双网卡情况下必须配置，比如阿里云这种，主节点需要配置ip1和ip2，从节点只需要配置ip1即可）
brokerIP1 = 10.10.8.11
#nameServer地址，分号分割
namesrvAddr=10.10.8.11:9875
#Broker 对外服务的监听端口,
listenPort = 10911
#是否允许Broker自动创建Topic
autoCreateTopicEnable = true
#是否允许 Broker 自动创建订阅组
autoCreateSubscriptionGroup = true
#linux开启epoll
useEpollNativeSelector = true

#数据存放的根目录
#storePathRootDir = /root/store/path
#commit log保存目录
#storePathCommitLog = /root/store/path/commitlog
#消费队列存储路径存储路径
#storePathConsumerQueue = /root/store/path/consumequeue

slaveReadEnable = true
EOF

#启动容器
默认内存
docker run -d --name rocketmq-broker-a --network rocketmq -p 10909:10909 -p 10911:10911 -v /home/docker/mount/rocketmq/broker/broker-a/logs:/root/logs -v /home/docker/mount/rocketmq/broker/broker-a/store:/root/store -v /home/docker/mount/rocketmq/broker/broker-a/conf:/home/rocketmq/rocketmq-4.9.3/conf apache/rocketmq:4.9.3 sh mqbroker -c /home/rocketmq/rocketmq-4.9.3/conf/broker.conf
指定内存
docker run -d --name rocketmq-broker-a --network rocketmq -p 10909:10909 -p 10911:10911 -v /home/docker/mount/rocketmq/broker/broker-a/logs:/root/logs -v /home/docker/mount/rocketmq/broker/broker-a/store:/root/store -v /home/docker/mount/rocketmq/broker/broker-a/conf:/home/rocketmq/rocketmq-4.9.3/conf -e "JAVA_OPT_EXT=-server -Xms1g -Xmx1g -Xmn512m" apache/rocketmq:4.9.3 sh mqbroker -c /home/rocketmq/rocketmq-4.9.3/conf/broker.conf

docker update --restart=always rocketmq-broker-a (docker_id)
```

### 部署console

```
docker run -d --name rocketmq-console --network rocketmq -e "JAVA_OPTS=-Drocketmq.namesrv.addr=rocketmq-nameserver:9876 -Dcom.rocketmq.sendMessageWithVIPChannel=false" -p 8000:8080 apacherocketmq/rocketmq-dashboard:latest

docker update --restart=always rocketmq-console (docker_id)
```

### 访问验证

http://10.10.8.11:8000/
