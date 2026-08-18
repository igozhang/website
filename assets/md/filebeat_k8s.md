# filebeat_k8s

## 内部部署（K8S）

```
helm repo add elastic https://helm.elastic.co
helm repo update

helm upgrade --install filebeat elastic/filebeat \
  --version 7.17.3 \
  --namespace elk \
  --create-namespace \
  -f filebeat-values.yaml

验证：
kubectl exec -n elk elasticsearch-master-0 -- curl -s "http://localhost:9200/_cat/indices?v" | grep logstash

接新组件的时候：
1. 新组件一个 namespace（最常见）：编辑 filebeat-k8s-values.yaml 里 condition.or，增加一行
- equals: { kubernetes.namespace: "你的ns" }，然后执行：
helm upgrade filebeat elastic/filebeat --version 7.17.3 -n elk -f filebeat-k8s-values.yaml
2. 同一 namespace 里只采部分服务：启用上面 方式 B（hints） 注释块，并在对应 Deployment/StatefulSet 的 Pod 模板上加：
co.elastic.logs/enabled: "true"（只给需要进 ELK 的服务加）。
```

## 外部部署

```
wget https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-7.17.3-amd64.deb
dpkg -i filebeat-7.17.3-amd64.deb

cp /etc/filebeat/filebeat.yml /etc/filebeat/filebeat.yml.bak.$(date +%F-%H%M%S)

tee /etc/filebeat/filebeat.yml >/dev/null <<'EOF'
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /data/redis/7000/redis.log
    fields:
      index_prefix: hu-redis
      app: redis
      host_name: redis-xyl-71
    fields_under_root: true

output.logstash:
  hosts: ["10.80.238.54:30308","10.80.238.55:30308","10.80.238.56:30308"]

setup.template.enabled: false
setup.ilm.enabled: false

logging.level: info
EOF

filebeat test config -c /etc/filebeat/filebeat.yml
filebeat test output -c /etc/filebeat/filebeat.yml

systemctl enable filebeat
systemctl restart filebeat
systemctl status filebeat --no-pager -l
```

## 验证

```
查看最近5条日志
kubectl logs -n elk -l app=filebeat-filebeat --tail=5
# 1) 看今天是否有新索引
kubectl exec -n elk elasticsearch-master-0 -- \
  curl -s "http://localhost:9200/_cat/indices/logstash-*?v&s=index"

# 2) 看 rocketmq 最近5分钟写入量（按你ES时间）
kubectl exec -n elk elasticsearch-master-0 -- \
  curl -s -H 'Content-Type: application/json' \
  "http://localhost:9200/logstash-*/_count" \
  -d '{"query":{"bool":{"filter":[{"term":{"kubernetes.namespace":"rocketmq"}},{"range":{"@timestamp":{"gte":"now-5m"}}}]}}}'

# 3) 抽样看最新3条rocketmq日志
kubectl exec -n elk elasticsearch-master-0 -- \
  curl -s -H 'Content-Type: application/json' \
  "http://localhost:9200/logstash-*/_search?pretty" \
  -d '{"size":3,"sort":[{"@timestamp":"desc"}],"_source":["@timestamp","kubernetes.pod.name","message","log.file.path"],"query":{"term":{"kubernetes.namespace":"rocketmq"}}}'
```

## OUTPUT

```
NAME: filebeat
LAST DEPLOYED: Wed Apr 22 07:46:15 2026
NAMESPACE: elk
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
1. Watch all containers come up.
  $ kubectl get pods --namespace=elk -l app=filebeat-filebeat -w
```

## filebeat-values

```
cd /igo/soft/elk

tee filebeat-values.yaml >/dev/null <<'EOF'
daemonset:
  enabled: true

deployment:
  enabled: false

extraEnvs:
  - name: NODE_NAME
    valueFrom:
      fieldRef:
        fieldPath: spec.nodeName

filebeatConfig:
  filebeat.yml: |
    filebeat.autodiscover:
      providers:
        - type: kubernetes
          node: ${NODE_NAME}
          templates:
            - condition:
                and:
                  - equals:
                      kubernetes.namespace: rocketmq
                  - equals:
                      kubernetes.container.name: broker
              config:
                - type: container
                  paths:
                    - /var/log/containers/*-${data.kubernetes.container.id}.log
                  processors:
                    - add_fields:
                        target: ''
                        fields:
                          index_prefix: rocket-mq-broke
                          app: rocketmq-broker

            - condition:
                and:
                  - equals:
                      kubernetes.namespace: mysql
                  - equals:
                      kubernetes.container.name: mysql
              config:
                - type: container
                  paths:
                    - /var/log/containers/*-${data.kubernetes.container.id}.log
                  processors:
                    - add_fields:
                        target: ''
                        fields:
                          index_prefix: mysql
                          app: mysql

    output.logstash:
      hosts: ["logstash-logstash.elk.svc:5044"]

    setup.template.enabled: false
    setup.ilm.enabled: false
EOF

helm upgrade --install filebeat elastic/filebeat \
  --version 7.17.3 \
  --namespace elk \
  --create-namespace \
  -f filebeat-values.yaml
```
