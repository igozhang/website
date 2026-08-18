# mysql_k8s

k8s上安装mysql8.4.5主备模式

## env

k8s 1.24  
ubuntu20.04  
helm v3.19.2 要求大于3.8.8 否则拉取报oci错误

## 准备

```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm search repo bitnami/mysql --versions
"NAME            CHART VERSION   APP VERSION     DESCRIPTION"
"bitnami/mysql   9.14.4          8.0.35          MySQL is a fast, reliable, scalable, and easy t..."
"bitnami/mysql   12.3.5          8.4.5           MySQL is a fast, reliable, scalable, and easy t..."

先确认可以拉镜像：
docker pull docker.io/bitnamilegacy/mysql:8.4.5-debian-12-r0
```

## 部署

```
helm upgrade --install mysql-cluster bitnami/mysql \
  --version 12.3.5 \
  --namespace mysql \
  --create-namespace \
  --set architecture=replication \
  --set auth.rootPassword='Mysql@2026' \
  --set auth.replicationPassword='Mysql@2026' \
  --set secondary.replicaCount=1 \
  --set primary.persistence.size=200Gi \
  --set secondary.persistence.size=200Gi \
  --set image.registry=docker.io \
  --set image.repository=bitnamilegacy/mysql \
  --set global.security.allowInsecureImages=true \
  --set primary.service.type=NodePort \
  --set primary.service.nodePorts.mysql=30306 \
  --set secondary.service.type=NodePort \
  --set secondary.service.nodePorts.mysql=30307

卸载：helm uninstall mysql-cluster -n mysql
```

```
离线安装

所有节点：
docker load -i /igo/soft/mysql8.4.5/offline-images/docker.io_bitnamilegacy_mysql_8.4.5-debian-12-r0.tar
或者
ctr -n k8s.io images import /igo/soft/mysql8.4.5/offline-images/docker.io_bitnamilegacy_mysql_8.4.5-debian-12-r0.tar

helm upgrade --install mysql-cluster ./mysql-12.3.5.tgz \
  --namespace mysql \
  --create-namespace \
  --set architecture=replication \
  --set auth.rootPassword='Mysql@2026' \
  --set auth.replicationPassword='Mysql@2026' \
  --set secondary.replicaCount=1 \
  --set primary.persistence.size=200Gi \
  --set secondary.persistence.size=200Gi \
  --set image.registry=docker.io \
  --set image.repository=bitnamilegacy/mysql \
  --set image.tag=8.4.5-debian-12-r0 \
  --set image.pullPolicy=IfNotPresent \
  --set global.security.allowInsecureImages=true \
  --set primary.service.type=NodePort \
  --set primary.service.nodePorts.mysql=30306 \
  --set secondary.service.type=NodePort \
  --set secondary.service.nodePorts.mysql=30307
```

## 验证

```
kubectl get pods -n mysql
kubectl get svc -n mysql
kubectl exec -it -n mysql mysql-cluster-primary-0 -- bash
mysql -uroot -p
SHOW BINARY LOG STATUS;
8.4之前老命令：
SHOW MASTER STATUS;  # 记录File和Position值

# 切换到备节点Pod
kubectl exec -it -n mysql mysql-cluster-secondary-0 -- bash
mysql -uroot -p
# 查看备节点复制状态，确保Slave_IO_Running和Slave_SQL_Running均为Yes
SHOW REPLICA STATUS\G
8.4之前老命令：
SHOW SLAVE STATUS\G
```

## 输出

```
Tip:

  Watch the deployment status using the command: kubectl get pods -w --namespace mysql

Services:

  echo Primary: mysql-cluster-primary.mysql.svc.cluster.local:3306
  echo Secondary: mysql-cluster-secondary.mysql.svc.cluster.local:3306

Execute the following to get the administrator credentials:

  echo Username: root
  MYSQL_ROOT_PASSWORD=$(kubectl get secret --namespace mysql mysql-cluster -o jsonpath="{.data.mysql-root-password}" | base64 -d)

To connect to your database:

  1. Run a pod that you can use as a client:

      kubectl run mysql-cluster-client --rm --tty -i --restart='Never' --image  docker.io/bitnamilegacy/mysql:8.4.5-debian-12-r0 --namespace mysql --env MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD --command -- bash

  2. To connect to primary service (read/write):

      mysql -h mysql-cluster-primary.mysql.svc.cluster.local -uroot -p"$MYSQL_ROOT_PASSWORD"

  3. To connect to secondary service (read-only):

      mysql -h mysql-cluster-secondary.mysql.svc.cluster.local -uroot -p"$MYSQL_ROOT_PASSWORD"
```
