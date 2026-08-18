# nacos_k8s

使用helm3 安装nacos-server:v2.2.1集群 到k8s1.24集群 的完整步骤  
使用mysql：

```
root@k8s-master51:/igo/soft/mysql8.4.5# kg svc -n mysql
NAME                               TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
mysql-cluster-primary              NodePort    10.96.2.79    <none>        3306:30306/TCP   90m
mysql-cluster-primary-headless     ClusterIP   None          <none>        3306/TCP         90m
mysql-cluster-secondary            NodePort    10.96.1.149   <none>        3306:30307/TCP   90m
mysql-cluster-secondary-headless   ClusterIP   None          <none>        3306/TCP         90m
```

## env

k8s 1.24  
helm v3.19.2  
nacos-server:v2.2.1  
https://hub.docker.com/r/nacos/nacos-server/tags

## 数据库建表，初始化

1. 建库  
   kubectl exec -it -n mysql mysql-cluster-primary-0 -- bash  
   mysql -uroot -p

```
CREATE DATABASE nacos_config DEFAULT CHARACTER SET utf8mb4;
CREATE USER 'nacos'@'%' IDENTIFIED BY 'Nacos@2026';
GRANT ALL PRIVILEGES ON nacos_config.* TO 'nacos'@'%';
FLUSH PRIVILEGES;
```

下载建表语句：  
wget -O mysql-schema-2.2.1.sql https://raw.githubusercontent.com/alibaba/nacos/2.2.1/distribution/conf/mysql-schema.sql  
拷贝进容器  
kubectl -n mysql cp /igo/soft/nacos/mysql-schema-2.2.1.sql mysql-cluster-primary-0:/tmp/mysql-schema-2.2.1.sql

2. 导入表结构  
   mysql> use nacos\_config;  
   Database changed  
   mysql> source /tmp/mysql-schema-2.2.1.sql;  
   Query OK, 0 rows affected, 3 warnings (0.03 sec)

随时测一把连通性：

```
kubectl -n nacos run mysql-client --rm -it --restart=Never \
  --image=mysql:8.0 \
  --command -- bash

mysql -h mysql-cluster-primary.mysql.svc.cluster.local -P 3306 -u nacos -p
mysql> use nacos_config
mysql> show tables;
```

## 安装

```
git clone https://github.com/nacos-group/nacos-k8s.git
cd nacos-k8s/helm
helm show values . > values-default.yaml

！！！由于新版的chart和2.2.1app不匹配，探针路径问题导致POD永远无法ready，需要关闭探针：
sed -i.bak '/startupProbe:/,/path: \/v3\/console\/health\/liveness/d' templates/statefulset.yaml

JWT_TOKEN=$(openssl rand -base64 32 | tr -d '\n')
helm upgrade --install nacos . -n nacos \
  --set global.mode=cluster \
  --set nacos.replicaCount=3 \
  --set nacos.image.tag=v2.2.1 \
  --set service.type=NodePort \
  --set nacos.storage.type=mysql \
  --set nacos.storage.db.host=mysql-cluster-primary.mysql.svc.cluster.local \
  --set nacos.storage.db.port=3306 \
  --set nacos.storage.db.name=nacos_config \
  --set nacos.storage.db.username=nacos \
  --set nacos.storage.db.password='Nacos@2026' \
  --set-string nacos.authToken="$JWT_TOKEN" \
  --set-string nacos.identityKey='serverIdentity' \
  --set-string nacos.identityValue='security' \
  --set nacos.health.enabled=false

kubectl patch svc nacos-cs -n nacos --type='json' -p='[
  {"op":"replace","path":"/spec/ports/0/nodePort","value":30848},
  {"op":"replace","path":"/spec/ports/1/nodePort","value":31848}
]'

后续密钥可以找回：
kubectl -n nacos exec sts/nacos -- env | grep NACOS_AUTH

会出现 POD 无法ready，log没有报错的情况，原因：探针配错版本（3.x chart 的探针规则（/v3/... + 8080），而镜像是 2.2.1）
    Liveness:  http-get http://:8080/v3/console/health/liveness delay=10s timeout=10s period=5s #success=1 #failure=3
    Startup:   http-get http://:8080/v3/console/health/readiness delay=180s timeout=10s period=5s #success=1 #failure=3

正确探针：kubectl exec -it nacos-0 -n nacos -- curl -i http://127.0.0.1:8848/nacos/actuator/health

文件写死： templates/statefulset.yaml
解决：（关闭探针）
sed -i.bak '/startupProbe:/,/path: \/v3\/console\/health\/liveness/d' templates/statefulset.yaml

kubectl -n nacos logs nacos-0 --tail=50
kubectl -n nacos logs nacos-1 --tail=50
kubectl -n nacos logs nacos-2 --tail=50
```

## 访问说明

客户端访问：  
nacos-cs（ClusterIP）：客户端统一接入点，负载均衡到 Nacos Pod，且提供控制台（/nacos）、配置管理、服务发现等 HTTP 能力  
nacos-hs（Headless）：集群内节点直连，Nacos 集群高可用的基础  
后台访问：http://<集群任意节点IP>:30466/nacos

```
root@k8s-pre-31:/igo/soft/nacos/nacos-k8s/helm# kg svc -n nacos
NAME       TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                                                                                     AGE
nacos-cs   NodePort    10.96.3.63   <none>        8848:30466/TCP,9848:30403/TCP,9849:31985/TCP,8080:30909/TCP,9080:32287/TCP,7848:30000/TCP   72s
nacos-hs   ClusterIP   None         <none>        8848/TCP,9848/TCP,9849/TCP,8080/TCP,9080/TCP,7848/TCP\
```

8848 主 HTTP 端口（控制台 / OpenAPI） 走nacos-cs负载均衡Nacos  
9848 客户端 gRPC 通信 客户端配置使用，不直接浏览器访问Nacos  
9849 服务间 gRPC 同步 仅集群内nacos-hs通信使用Nacos  
7848 Raft 共识通信 仅集群内nacos-hs通信使用Nacos

## output

NAME: nacos  
LAST DEPLOYED: Fri Apr 10 09:08:10 2026  
NAMESPACE: nacos  
STATUS: deployed  
REVISION: 1  
TEST SUITE: None  
NOTES:

1. Get the application URL by running these commands:  
   export NODE\_PORT=$(kubectl get svc nacos-cs --namespace nacos -o jsonpath="{.spec.ports[3].nodePort}" )  
   export NODE\_IP=$(kubectl get nodes --namespace nacos -o jsonpath="{.items[0].status.addresses[0].address}")  
   echo http://$NODE\_IP:$NODE\_PORT/index.html
2. MODE:  
   standalone: you need to modify replicaCount in the values.yaml, .Values.replicaCount=1  
   cluster: kubectl scale sts nacos-nacos --replicas=3

默认账密： nacos/nacos

## 回退重装命令

```
kubectl delete sts nacos -n nacos
helm uninstall nacos -n nacos
kubectl delete all -l app.kubernetes.io/instance=nacos -n nacos --force --grace-period=0
kubectl delete sts nacos -n nacos --force --grace-period=0
kubectl delete configmap nacos-cm -n nacos
kubectl delete svc -n nacos --all
kubectl get all -n nacos
```

## 离线安装

```
docker pull nacos/nacos-server:v2.2.1
docker pull nacos/nacos-peer-finder-plugin:1.1
 nacos-peer-finder用来在 Kubernetes 集群模式下自动发现同组 Nacos 节点
docker save -o nacos-server-v2.2.1.tar nacos/nacos-server:v2.2.1
docker save -o nacos-peer-finder-plugin-1.1.tar nacos/nacos-peer-finder-plugin:1.1

docker load -i nacos-server-v2.2.1.tar
docker load -i nacos-peer-finder-plugin-1.1.tar
docker images | grep -E 'nacos-server|peer-finder'
```
