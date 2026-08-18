# k8s_zabbix6

k8s上部署完整 Zabbix6

1. 部署命令使用helm upgrade –install，兼容两种场景
2. namespace等其他参数都尽量在helm安装命令里指定，其中namespace在命令里创建
3. 使用6:latest 版本
4. 使用ingress模式
5. 使用域名"zabbix-hu.igozhang.cn"  
   env  
   helm v3.15.4  
   k8s v1.29.10

## 安装

```
export NAMESPACE='zabbix'
export RELEASE_NAME='zabbix'
export CHART_VERSION='4.4.1'
export INGRESS_CLASS='nginx'
export ZABBIX_DOMAIN='zabbix-hu.igozhang.cn'
export ZABBIX_SERVER_NODEPORT='31051'

helm repo add zabbix-community https://zabbix-community.github.io/helm-zabbix 2>/dev/null || true
helm repo update

helm upgrade --install "${RELEASE_NAME}" zabbix-community/zabbix \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --version "${CHART_VERSION}" \
  --dependency-update \
  --set zabbixImageTag=ubuntu-6.0-latest \
  --set postgresql.enabled=true \
  --set postgresql.persistence.enabled=true \
  --set postgresql.persistence.storageSize=200Gi \
  --set postgresAccess.user=zabbix \
  --set postgresAccess.password='Zabbix@2026' \
  --set postgresAccess.database=zabbix \
  --set zabbixServer.service.type=NodePort \
  --set-string zabbixServer.service.externalTrafficPolicy=Local \
  --set zabbixServer.service.nodePort="${ZABBIX_SERVER_NODEPORT}" \
  --set zabbixWeb.extraEnv[0].name=ZBX_SERVER_HOST \
  --set-string zabbixWeb.extraEnv[0].value="${RELEASE_NAME}-zabbix-server" \
  --set zabbixWeb.extraEnv[1].name=ZBX_SERVER_PORT \
  --set-string zabbixWeb.extraEnv[1].value=10051 \
  --set zabbixWeb.extraEnv[2].name=ZBX_SERVER_NAME \
  --set-string zabbixWeb.extraEnv[2].value='Zabbix' \
  --set ingress.enabled=true \
  --set-string ingress.ingressClassName="${INGRESS_CLASS}" \
  --set ingress.hosts[0].host="${ZABBIX_DOMAIN}" \
  --set ingress.hosts[0].paths[0].path='/' \
  --set ingress.hosts[0].paths[0].pathType=Prefix \
  --set ingress.pathType=Prefix

web 默认账密
Admin/zabbix

OUTPUT
NAME: zabbix
LAST DEPLOYED: Tue May 12 16:23:50 2026
NAMESPACE: zabbix
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace zabbix -l "app=zabbix-zabbix-web" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace zabbix $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:8888 to use your application"
  kubectl --namespace zabbix port-forward $POD_NAME 8888:$CONTAINER_PORT
```

## 接入

```
1. agent端：
cat <<'YAML' | ansible-playbook /dev/stdin -f 1
---
- hosts: "10.80.238.77"
  become: true
  gather_facts: false
  tasks:
    - ansible.builtin.shell: |
        set -euo pipefail
        Hostname=$(ip -4 addr show scope global | awk '/inet / {print $2}' | cut -d/ -f1 | head -n1)
        test -n "$Hostname"
        cp -a /usr/local/zabbix/conf/zabbix_agentd.conf "/usr/local/zabbix/conf/zabbix_agentd.conf.bak.$(date +%Y%m%d%H%M%S)"
        tee /usr/local/zabbix/conf/zabbix_agentd.conf >/dev/null <<EOF
        LogFile=/tmp/zabbix_agentd.log
        Server=10.80.238.38
        ServerActive=10.80.238.38:31051
        Hostname=${Hostname}
        Timeout=20
        ListenPort=10050
        EOF
        systemctl restart zabbix-agent
        systemctl --no-pager -l status zabbix-agent || true
      args:
        executable: /bin/bash
YAML

2. server端
登录 Zabbix Web → Configuration → Actions。
左上角 Event source 选 Autoregistration actions（自动注册）。
Create action：
Name：Auto register 10.80.238.0/24。
Conditions 增加一条：Host name | matches | 10.80.238.*
Operations 至少包含：
Add host
Add to host groups
Link to template：Template OS Linux by Zabbix agent
Save
```

postgresql 给 Zabbix 存配置、历史、趋势等数据的 PostgreSQL 数据库。  
zabbixServer 调度监控、处理 trap/主动 agent、写库、告警逻辑等，是核心引擎。  
zabbixWeb 浏览器里配置、看图、告警等；通过 DB 与 Server 协同。  
zabbixWebService 给前端提供部分 服务端能力（例如报表/导出等依赖的后端服务
