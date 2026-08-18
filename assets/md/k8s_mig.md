# k8s_mig

将A基地集群k8s业务复制到B基地集群  
总体目标：需要将A基地k8s集群 namespace mes-app2023的所有资源 复制到 B基地k8s集群 namespace mes-app  
总体思路：导出必要资源，清洗后导入目标集群

1. 先定位namespace mes-app2023里面所有deployment涉及的image，并导入目标harbor

```
kubectl get deploy -n mes-app2023 -o custom-columns=NAME:.metadata.name,IMAGE:.spec.template.spec.containers[*].image
```

1.1 将涉及的所有镜像PULL并PUSH到目标仓库  
src="harbor-a-mom.igozhang.cn/standard-auto-a-mon/$img"  
dst="harbor-b-mom.igozhang.cn/humon/$img"

2. 导出yaml  
   在老集群按照deployment区分导出文件（按deployment区分，同deployment的资源放在同一个文件夹）
3. 清洗yaml
4. 导入yaml

| 类别 | 规则 |
| --- | --- |
| 命名空间 | mes-app（Pod 模板 不再 写 namespace） |
| 镜像 | harbor-a-mon…/standard-auto-a-mon/ → harbor-hu.igozhang.cn/humon/ |
| Nacos | nacos-headless.nsfly-nacos:8848 → nacos-hs.nacos.svc.cluster.local:8848 |
| PostgreSQL | 10.129.241.31:2345 → 10.80.237.10:2345 |
| Redis | standalone → cluster：spring.redis.cluster.nodes + 6 节点 :7000；分布式锁改为 cluster + cluster.nodes |
| 域名/网关 | gateway-a-mon / app-a-mon / idm-a-mon / paas-a-mon → 对应 \*-hu / rancher-hu |
| Ingress 主机 | a-mon.igozhang.cn → mes-hu.igozhang.cn；app-a-mon… → app-hu…；a-mon-pre… → mes-hu-pre… |
| 资源 | CPU/内存、JAVA\_OPTS 中 -Xms/-Xmx 均为 原值 1/4 向上取整 |
| 清理 | 去掉 Rancher/Cattle/objectset 等注解与标签；Service 去掉 clusterIP |

### 拉镜像

1. 先定位namespace mes-app2023里面所有deployment涉及的image

```
kubectl get deploy -n mes-app2023 -o custom-columns=NAME:.metadata.name,IMAGE:.spec.template.spec.containers[*].image
```

xxx-mes-file-center harbor-a-mom.igozhang.cn/standard-auto-a-mon/xxx-mes-file-center:20240115\_162156-9cbe206b  
xxx-mes-auth-center harbor-a-mom.igozhang.cn/standard-auto-a-mon/xxx-mes-auth-center:20240109\_132438-1a9f9c9f  
xxx-mes-base-server harbor-a-mom.igozhang.cn/standard-auto-a-mon/xxx-mes-base-server:20240408\_093110-19336648

推拉镜像脚本：

```
# 先登录
docker login harbor-a-mom.igozhang.cn
docker login harbor-b-mom.igozhang.cn

# 批量同步脚本开始
images=(
"xxx-mes-file-center:20240115_162156-9cbe206b"
"xxx-mes-auth-center:20240109_132438-1a9f9c9f"
"xxx-mes-base-server:20240408_093110-19336648"
)

for img in "${images[@]}"; do
  src="harbor-a-mom.igozhang.cn/standard-auto-a-mon/$img"
  dst="harbor-b-mom.igozhang.cn/humon/$img"
  echo "=> $src → $dst"

  docker pull "$src"
  docker tag "$src" "$dst"
  docker push "$dst"
done
```

### 导出YAML

需要 jq,jy

```
#!/bin/bash
set -e

# 基础配置
NS="mes-app2023"
EXPORT_ROOT="k8s-export-${NS}"
mkdir -p "${EXPORT_ROOT}"

# 获取所有 Deployment 名称
DEPLOYS=$(kubectl get deploy -n "${NS}" -o jsonpath='{.items[*].metadata.name}')

for deploy in ${DEPLOYS}; do
  echo "=================================================="
  echo "导出: ${deploy}"

  DIR="${EXPORT_ROOT}/${deploy}"
  mkdir -p "${DIR}"

  # ------------------------------
  # 1. Deployment
  # ------------------------------
  kubectl get deploy "${deploy}" -n "${NS}" -o yaml |
    yq 'del(
      .metadata.uid,
      .metadata.resourceVersion,
      .metadata.creationTimestamp,
      .metadata.generation,
      .metadata.managedFields,
      .metadata.annotations."deployment.kubernetes.io/revision",
      .metadata.annotations."kubectl.kubernetes.io/last-applied-configuration",
      .metadata.annotations."field.cattle.io/*",
      .metadata.annotations."objectset.rio.cattle.io/*",
      .metadata.namespace,
      .status
    )' \
    > "${DIR}/deployment.yaml"

  # ------------------------------
  # 2. Service
  # ------------------------------
  if kubectl get svc "${deploy}" -n "${NS}" &>/dev/null; then
    kubectl get svc "${deploy}" -n "${NS}" -o yaml |
      yq 'del(
        .metadata.uid,
        .metadata.resourceVersion,
        .metadata.creationTimestamp,
        .metadata.managedFields,
        .metadata.namespace,
        .status
      )' \
      > "${DIR}/service.yaml"
  fi

  # ------------------------------
  # 3. Ingress（非常重要）
  # ------------------------------
  if kubectl get ingress -n "${NS}" -o name | grep -qwF -e "${deploy}"; then
    kubectl get ingress "${deploy}" -n "${NS}" -o yaml |
      yq 'del(
        .metadata.uid,
        .metadata.resourceVersion,
        .metadata.creationTimestamp,
        .metadata.managedFields,
        .metadata.namespace,
        .status
      )' \
      > "${DIR}/ingress.yaml"
  fi

  # ------------------------------
  # 4. ConfigMap
  # ------------------------------
  if kubectl get cm "${deploy}" -n "${NS}" &>/dev/null; then
    kubectl get cm "${deploy}" -n "${NS}" -o yaml |
      yq 'del(
        .metadata.uid,
        .metadata.resourceVersion,
        .metadata.creationTimestamp,
        .metadata.managedFields,
        .metadata.namespace
      )' \
      > "${DIR}/configmap.yaml"
  fi

  # ------------------------------
  # 5. Secret
  # ------------------------------
  if kubectl get secret "${deploy}" -n "${NS}" &>/dev/null; then
    kubectl get secret "${deploy}" -n "${NS}" -o yaml |
      yq 'del(
        .metadata.uid,
        .metadata.resourceVersion,
        .metadata.creationTimestamp,
        .metadata.managedFields,
        .metadata.namespace
      )' \
      > "${DIR}/secret.yaml"
  fi

  # ------------------------------
  # 6. ServiceAccount（如果有）
  # ------------------------------
  if kubectl get sa "${deploy}" -n "${NS}" &>/dev/null; then
    kubectl get sa "${deploy}" -n "${NS}" -o yaml |
      yq 'del(
        .metadata.uid,
        .metadata.resourceVersion,
        .metadata.creationTimestamp,
        .metadata.managedFields,
        .metadata.namespace,
        .secrets,
        .imagePullSecrets
      )' \
      > "${DIR}/serviceaccount.yaml"
  fi

  echo "完成: ${DIR}"
done

echo
echo "✅ 导出完成，目录结构："
tree "${EXPORT_ROOT}"
```
