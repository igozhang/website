# rocketmq_k8s

在k8s集群上通过helm部署rocketmq:4.9.4

1. 3节点集群
2. 使用nodeport模式
3. 部署命令使用helm upgrade --install
4. namespace等其他参数都尽量在helm安装命令里指定
5. 命令可复用（兼容install和upgrade场景）

## env

rocketmq:4.9.4  
k8s 1.24.6  
dcoker 25.0.5  
helm 3.19.2

## 安装

```
helm repo add rocketmq-repo https://helm-charts.itboon.top/rocketmq
helm repo update
收集values有助于判断正确的参数设置：
helm show values rocketmq-repo/rocketmq 

helm upgrade --install rocketmq \
  --namespace rocketmq \
  --create-namespace \
  --set nameserver.replicaCount=2 \
  --set nameserver.service.type=NodePort \
  --set nameserver.resources.limits.cpu=1 \
  --set nameserver.resources.limits.memory=2Gi \
  --set nameserver.resources.requests.cpu=500m \
  --set nameserver.resources.requests.memory=1Gi \
  --set nameserver.jvm.maxHeapSize=1G \
  \
  --set broker.size.master=3 \
  --set broker.size.replica=0 \
  --set broker.master.brokerRole=ASYNC_MASTER \
  --set broker.service.type=NodePort \
  --set broker.persistence.enabled=true \
  --set broker.persistence.size=50Gi \
  --set broker.persistence.storageClass="" \
  --set broker.master.resources.limits.cpu=4 \
  --set broker.master.resources.limits.memory=8Gi \
  --set broker.master.resources.requests.cpu=2 \
  --set broker.master.resources.requests.memory=4Gi \
  --set broker.master.jvm.maxHeapSize=4G \
  \
  --set image.repository=apache/rocketmq \
  --set image.tag=4.9.4 \
  --set proxy.enabled=false \
  \
  --set dashboard.enabled=true \
  --set dashboard.service.type=NodePort \
  --set dashboard.image.repository=apacherocketmq/rocketmq-dashboard \
  --set dashboard.image.tag=1.0.0 \
  --set dashboard.containerPort=8080 \
  rocketmq-repo/rocketmq

面对生产，强一致性场景，修改以下参数，达到2主2从，实时同步（数据强一致，牺牲性能）：
# 只改这3个参数，其他全部不变
--set broker.size.master=2 \
--set broker.size.replica=2 \
--set broker.master.brokerRole=SYNC_MASTER \
为什么要 SYNC_MASTER？
因为有从节点，要等同步完成才返回成功，否则主挂了消息会丢

RocketMQ 4.9.4 不支持 proxy 组件

卸载重装
helm uninstall rocketmq -n rocketmq
kubectl delete ns rocketmq

kubectl delete pvc -n rocketmq --all
```

## SVC信息

```
应用应该把两个nameserver的信息都填上，或者填svc地址（会自动负载均衡）
rocketmq-nameserver-0.rocketmq-nameserver-headless.rocketmq.svc:9876;rocketmq-nameserver-1.rocketmq-nameserver-headless.rocketmq.svc:9876
rocketmq-nameserver.rocketmq.svc:9876
运维用nodeport30888来看dashboard

root@k8s-master51:~# kg svc -n rocketmq
NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
rocketmq-dashboard             NodePort    10.96.0.43    <none>        8080:30327/TCP   8m10s
rocketmq-nameserver            NodePort    10.96.1.141   <none>        9876:32348/TCP   8m10s
rocketmq-nameserver-headless   ClusterIP   None          <none>        9876/TCP         8m10s

root@k8s-master51:~#
root@k8s-master51:~# kg pod -n rocketmq
NAME                                  READY   STATUS    RESTARTS   AGE
rocketmq-broker-master-0              1/1     Running   0          7m59s
rocketmq-broker-master-1              1/1     Running   0          7m19s
rocketmq-broker-master-2              1/1     Running   0          6m39s
rocketmq-dashboard-6594f45c9b-qqvsc   1/1     Running   0          2m53s
rocketmq-nameserver-0                 1/1     Running   0          7m59s
rocketmq-nameserver-1                 1/1     Running   0          7m59s
```

## OUTPUT

NAME: rocketmq  
LAST DEPLOYED: Thu Apr 16 04:26:35 2026  
NAMESPACE: rocketmq  
STATUS: deployed  
REVISION: 1  
TEST SUITE: None  
NOTES:

> > > Nameserver Address:  
> > > rocketmq-nameserver.rocketmq.svc:9876
> > >
> > > RocketMQ Dashboard Auth:  
> > > username: admin password: admin  
> > > username: user01 password: userPass  
> > > Modify "rocketmq-dashboard-cm" configmap to change the password

## 离线安装

导出chart和images等必要资源以便后续离线安装  
打包成tar，每个资源单独打包  
查看所有镜像：broker 和 nameserver公用一个镜像

```
kubectl get pods -n rocketmq -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].image}{"\n"}{end}'

docker pull apache/rocketmq:4.9.4
docker pull apacherocketmq/rocketmq-dashboard:2.1.0

docker save -o apache_rocketmq_4.9.4.tar apache/rocketmq:4.9.4
docker save -o apacherocketmq_rocketmq-dashboard_2.1.0.tar apacherocketmq/rocketmq-dashboard:2.1.0

docker load -i apache_rocketmq_4.9.4.tar
docker load -i apacherocketmq_rocketmq-dashboard_2.1.0.tar
```

## 验证 verify

```
# 进入 broker
kubectl exec -it rocketmq-broker-master-0 -n rocketmq -- sh

# 进入目录
cd /home/rocketmq/rocketmq-4.9.4/bin

# 1. 创建 topic
sh mqadmin updateTopic -n rocketmq-nameserver:9876 -t test_topic -c rocketmq-helm -r 4 -w 4

# 2. 查看 topic
sh mqadmin topicList -n rocketmq-nameserver:9876

# 3. 发送消息
sh mqadmin sendMessage -n rocketmq-nameserver:9876 -t test_topic -p "hello rocketmq"

# 4. 消费消息（正确命令！！）
sh mqadmin consumeMessage -n rocketmq-nameserver:9876 -t test_topic -g test_group
```
