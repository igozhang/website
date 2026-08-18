# elk_k8s

env  
NAME STATUS ROLES AGE VERSION INTERNAL-IP EXTERNAL-IP OS-IMAGE KERNEL-VERSION CONTAINER-RUNTIME  
k8s-master51 Ready control-plane 13d v1.24.6 10.80.238.51 Ubuntu 20.04.6 LTS 5.4.0-216-generic docker://25.0.5  
Ubuntu 20.04.6 LTS  
helm v3.19.2

## 部署

```
添加仓库
helm repo add elastic https://helm.elastic.co
helm repo update

helm upgrade --install elasticsearch elastic/elasticsearch \
  --version 7.17.3 \
  --namespace elk \
  --create-namespace \
  --set imageTag=7.17.3 \
  --set replicas=3 \
  --set minimumMasterNodes=2 \
  --set antiAffinity=hard

Kibana 7（配置你的域名）
helm upgrade --install kibana elastic/kibana \
  --version 7.17.3 \
  --namespace elk \
  --create-namespace \
  --set imageTag=7.17.3 \
  --set elasticsearchHosts=http://elasticsearch-master:9200 \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set ingress.hosts[0].host=kibana-hu.sunwoda-evb.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix

Logstash
# 1.1 定义 pipeline（避免命令行引号转义问题）
read -r -d '' PIPELINE <<'EOF'
input { beats { port => 5044 } }

filter {
  ruby {
    code => "p = event.get('index_prefix'); event.set('target_index_prefix', p.to_s.empty? ? 'logstash' : p)"
  }
}

output {
  elasticsearch {
    hosts => ['http://elasticsearch-master:9200']
    index => "%{[target_index_prefix]}-%{+YYYY.MM.dd}"
  }
}
EOF
# 关键：转义逗号，避免 Helm --set 按逗号切分
PIPELINE_ESCAPED="${PIPELINE//,/\\,}"

helm upgrade --install logstash elastic/logstash \
  --version 7.17.3 \
  --namespace elk \
  --create-namespace \
  --set imageTag=7.17.3 \
  --set-json 'service.annotations={}' \
  --set service.type=NodePort \
  --set service.ports[0].name=beats \
  --set service.ports[0].port=5044 \
  --set service.ports[0].targetPort=5044 \
  --set service.ports[0].protocol=TCP \
  --set service.ports[1].name=http \
  --set service.ports[1].port=9600 \
  --set service.ports[1].targetPort=http \
  --set service.ports[1].protocol=TCP \
  --set extraPorts[0].name=beats \
  --set extraPorts[0].containerPort=5044 \
  --set-string 'logstashPipeline.logstash\.conf='"$PIPELINE_ESCAPED"
```

## OUTPUT

```
NAME: elasticsearch
LAST DEPLOYED: Wed Apr 22 02:19:04 2026
NAMESPACE: elk
STATUS: deployed
REVISION: 1
NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=elk -l app=elasticsearch-master -w2. Test cluster health using Helm test.
  $ helm --namespace=elk test elasticsearch

NAME: kibana
LAST DEPLOYED: Wed Apr 22 02:19:57 2026
NAMESPACE: elk
STATUS: deployed
REVISION: 1
TEST SUITE: None

NAME: logstash
LAST DEPLOYED: Wed Apr 22 02:44:17 2026
NAMESPACE: elk
STATUS: deployed
REVISION: 2
TEST SUITE: None
NOTES:
1. Watch all cluster members come up.
  $ kubectl get pods --namespace=elk -l app=logstash-logstash -w
```
