# helm

```
helm repo add elastic https://helm.elastic.co
helm repo update

# 列出 elastic 仓库下所有可用的 Chart
helm search repo elastic

# 语法：helm search repo <仓库名>/<chart名> --versions
helm search repo elastic/elasticsearch --versions
helm search repo elastic/kibana --versions
helm search repo elastic/filebeat --versions

# 查看指定 Chart 的默认配置
helm show values elastic/elasticsearch
helm show values elastic/kibana

# 导出为本地 yaml 文件
helm show values elastic/elasticsearch > elasticsearch-values.yaml
helm show values elastic/kibana > kibana-values.yaml

# 安装 Elasticsearch（命名空间：elastic，发布名：es）
helm install es elastic/elasticsearch -n elastic --create-namespace
helm install es elastic/elasticsearch --version 8.15.0 -n elastic
helm install es elastic/elasticsearch -f elasticsearch-values.yaml -n elastic

# 查看所有命名空间的 Helm 应用
helm list -A

# 查看指定命名空间
helm list -n elastic

回滚
helm history logstash -n elk
helm rollback logstash <REVISION> -n elk

# 卸载发布（保留 PVC 存储）
helm uninstall es -n elastic

# 卸载并删除 PVC（彻底清空数据）
helm uninstall es -n elastic --keep-pvc=false
```

## 安装命令参考

```
helm upgrade --install elasticsearch elastic/elasticsearch \
  --version 7.17.3 \
  --namespace elk \
  --create-namespace \
  --set imageTag=7.17.3 \
  --set replicas=1 \
  --set minimumMasterNodes=1
```
