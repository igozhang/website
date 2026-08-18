# prometheus_k8s

在一套在k8s 通过helm部署 Prometheus + granfa 实现对mysql以及rocketmq的监控及展示

1. 使用较新且稳定的版本为宜
2. 所有参数指定使用 helm --set 方式指定
3. 命令考虑install 和upgrade 两种情况兼容，且包含创建namespace
4. 使用nodeport模式以便后续运维
5. grafana使用域名"grafana-hu.igozhang.cn"

集群内组件接入的标准链路：

1. 在 k8s 部署 exporter，暴露 `/metrics`
2. 通过 Service 暴露 exporter
3. 通过 ServiceMonitor 声明抓取目标
4. Prometheus Operator 自动发现并抓取
5. Grafana 通过 Prometheus 数据源查询展示

env

```
NAME         STATUS   ROLES           AGE    VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
k8s-pre-31   Ready    control-plane   12d    v1.24.6   10.80.238.31   <none>        Ubuntu 20.04.6 LTS   5.4.0-216-generic   docker://25.0.5
helm v3.19.2
root@k8s-pre-31:~# kg svc -n mysql
NAME                               TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
mysql-cluster-primary              NodePort    10.96.2.135   <none>        3306:30306/TCP   12d
mysql-cluster-primary-headless     ClusterIP   None          <none>        3306/TCP         12d
mysql-cluster-secondary            NodePort    10.96.2.76    <none>        3306:30307/TCP   12d
mysql-cluster-secondary-headless   ClusterIP   None          <none>        3306/TCP         12d
root@k8s-pre-31:~# kg svc -n rocketmq
NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
rocketmq-dashboard             NodePort    10.96.2.147   <none>        8082:30888/TCP   11d
rocketmq-nameserver            NodePort    10.96.1.183   <none>        9876:30417/TCP   11d
rocketmq-nameserver-headless   ClusterIP   None          <none>        9876/TCP         11d
```

## 部署

```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add bjw-s https://bjw-s-labs.github.io/helm-charts
helm repo update
helm search repo prometheus-community/kube-prometheus-stack --versions
helm search repo prometheus-community/prometheus-mysql-exporter --versions
helm search repo bjw-s/app-template --versions

创建MYSQL监控账号：
kubectl -n mysql exec -i mysql-cluster-primary-0 -- mysql -uroot -p"${MYSQL_ROOT_PWD}" -e "
CREATE USER IF NOT EXISTS 'exporter'@'%' IDENTIFIED BY 'Exporter@123456';
GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO 'exporter'@'%';
FLUSH PRIVILEGES;
SHOW GRANTS FOR 'exporter'@'%';
"
验证：
kubectl -n mysql exec -it mysql-cluster-primary-0 -- mysql -uexporter -p'Exporter@123456' -e "SELECT VERSION();"
```

### prometheus

```
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --version 56.21.4 \
  --set grafana.enabled=true \
  --set grafana.adminPassword='Admin@123456' \
  --set grafana.ingress.enabled=true \
  --set grafana.ingress.ingressClassName='nginx' \
  --set grafana.ingress.hosts[0]='grafana-hu.igozhang.cn' \
  --set grafana.ingress.paths[0]='/' \
  --set grafana.ingress.pathType='Prefix' \
  --set grafana.service.type='ClusterIP' \
  --set grafana.persistence.enabled=true \
  --set grafana.persistence.storageClassName='nfs-client' \
  --set grafana.persistence.accessModes[0]='ReadWriteOnce' \
  --set grafana.persistence.size='10Gi' \
  --set prometheus.prometheusSpec.retention='15d' \
  --set prometheus.prometheusSpec.scrapeInterval='30s' \
  --set prometheus.prometheusSpec.scrapeTimeout='10s' \
  --set prometheus.prometheusSpec.evaluationInterval='30s' \
  --set prometheus.prometheusSpec.serviceMonitorSelector.matchLabels.release='monitoring' \
  --set prometheus.prometheusSpec.serviceMonitorNamespaceSelector.matchNames[0]='monitoring' \
  --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.storageClassName='nfs-client' \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.accessModes[0]='ReadWriteOnce' \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage='50Gi' \
  --set prometheus.service.type='NodePort' \
  --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.storageClassName='nfs-client' \
  --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.accessModes[0]='ReadWriteOnce' \
  --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.resources.requests.storage='10Gi'

NAME: monitoring
LAST DEPLOYED: Fri Apr 24 09:44:14 2026
NAMESPACE: monitoring
STATUS: deployed
REVISION: 1
NOTES:
kube-prometheus-stack has been installed. Check its status by running:
  kubectl --namespace monitoring get pods -l "release=monitoring"

Visit https://github.com/prometheus-operator/kube-prometheus for instructions on how to create & configure Alertmanager and Prometheus instances using the Operator.
```

### k8s-mysql

```
helm upgrade --install mysql-exporter prometheus-community/prometheus-mysql-exporter \
  --namespace monitoring \
  --create-namespace \
  --version 2.10.0 \
  --set mysql.user='exporter' \
  --set mysql.pass='Exporter@123456' \
  --set mysql.host='mysql-cluster-primary.mysql.svc.cluster.local' \
  --set mysql.port=3306 \
  --set serviceMonitor.enabled=true \
  --set serviceMonitor.interval='30s' \
  --set serviceMonitor.labels.release='monitoring'

NAMESPACE: monitoring
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace monitoring -l "app.kubernetes.io/name=prometheus-mysql-exporter,app.kubernetes.io/instance=mysql-exporter" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:9104 to use your application"
  kubectl --namespace monitoring port-forward $POD_NAME 9104
```

### k8s-rocketmq

```
helm upgrade --install rocketmq-exporter bjw-s/app-template \
  --namespace monitoring \
  --create-namespace \
  --version 3.4.0 \
  --set controllers.main.type='deployment' \
  --set controllers.main.replicas=1 \
  --set controllers.main.containers.main.image.repository='apache/rocketmq-exporter' \
  --set controllers.main.containers.main.image.tag='0.0.2' \
  --set controllers.main.containers.main.args[0]='--rocketmq.config.namesrvAddr=rocketmq-nameserver.rocketmq.svc.cluster.local:9876' \
  --set controllers.main.containers.main.args[1]='--server.port=5557' \
  --set controllers.main.containers.main.args[2]='--rocketmq.config.webTelemetryPath=/metrics' \
  --set service.main.controller='main' \
  --set service.main.ports.http.port=5557 \
  --set serviceMonitor.main.enabled=true \
  --set serviceMonitor.main.serviceName='main' \
  --set serviceMonitor.main.endpoints[0].port='http' \
  --set serviceMonitor.main.endpoints[0].path='/metrics' \
  --set serviceMonitor.main.endpoints[0].interval='30s' \
  --set serviceMonitor.main.labels.release='monitoring'

NAME: rocketmq-exporter
LAST DEPLOYED: Fri Apr 24 10:07:55 2026
NAMESPACE: monitoring
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

### 外部组件-redis

```
for ip in 71; do
  helm upgrade --install redis-exporter-${ip} bjw-s/app-template \
    --namespace monitoring \
    --create-namespace \
    --version 3.4.0 \
    --set controllers.main.type='deployment' \
    --set controllers.main.replicas=1 \
    --set controllers.main.containers.main.image.repository='oliver006/redis_exporter' \
    --set controllers.main.containers.main.image.tag='v1.66.0' \
    --set controllers.main.containers.main.args[0]='--web.listen-address=:9121' \
    --set controllers.main.containers.main.args[1]="--redis.addr=redis://10.80.238.${ip}:7000" \
    --set controllers.main.containers.main.args[2]='--redis.password=Redis@2026' \
    --set controllers.main.containers.main.args[3]='--is-cluster=true' \
    --set service.main.controller='main' \
    --set service.main.ports.metrics.port=9121 \
    --set serviceMonitor.main.enabled=true \
    --set serviceMonitor.main.serviceName='main' \
    --set serviceMonitor.main.endpoints[0].port='metrics' \
    --set serviceMonitor.main.endpoints[0].path='/metrics' \
    --set serviceMonitor.main.endpoints[0].interval='30s' \
    --set serviceMonitor.main.labels.release='monitoring'

  kubectl -n monitoring patch servicemonitor redis-exporter-${ip} --type='merge' -p "{
    \"spec\": {
      \"selector\": {
        \"matchLabels\": {
          \"app.kubernetes.io/instance\": \"redis-exporter-${ip}\",
          \"app.kubernetes.io/name\": \"redis-exporter-${ip}\",
          \"app.kubernetes.io/service\": \"redis-exporter-${ip}\"
        }
      }
    }
  }"
done
```

## 验证

```
kubectl -n monitoring get pods
kubectl -n monitoring get svc
kubectl -n monitoring get ingress
kubectl -n monitoring get servicemonitor
```
