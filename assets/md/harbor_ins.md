# harbor_ins

ubuntu 20.04.6  
k8s"v1.24.6"

自建harbor仓库

```
sealos方式： 要求sealos5.1.1 高版本
sealos pull registry.cn-shanghai.aliyuncs.com/labring/harbor:v2.8.2
sealos run registry.cn-shanghai.aliyuncs.com/labring/harbor:v2.8.2 -e \
sealos run registry.cn-shanghai.aliyuncs.com/labring/harbor:v2.8.2 -e \
HELM_OPTS="--set externalURL=https://harbor-hu.igozhang.cn --set expose.ingress.hosts.core=harbor-hu.igozhang.cn --set harborAdminPassword=Harbor@2026   --set persistence.persistentVolumeClaim.registry.size=2000Gi --set persistence.persistentVolumeClaim.trivy.size=10Gi --set persistence.persistentVolumeClaim.jobservice.jobLog.size=5Gi"

kubectl edit ingress harbor-ingress -n harbor
完了注释掉证书行“secretName: harbor-ingress”就会自动使用ingress-nginx默认证书

卸载重装
helm uninstall harbor -n harbor
kubectl delete pvc --all -n harbor
kubectl delete namespace harbor
kubectl get  all -n harbor
```

```
其他方式：
1. kubectl create namespace harbor || true
2. kubectl -n harbor create secret tls harbor-tls --cert=./tls.pem --key=./tls.key
```

配置并检查环境

## 安装

```
sealos安装会被忽略参数，upgrade再安装一次,文件版本可以helm list -n harbor 或者 ~/.cache/helm/ 查看
sealos run registry.cn-shanghai.aliyuncs.com/labring/harbor:v2.8.2 -e HELM_OPTS="--set externalURL=https://harbor-huizhou01.sunwoda-evb.com"
helm uninstall harbor -n harbor
helm install harbor harbor/harbor -n harbor \
  --version 1.12.2 \
  --set externalURL=https://harbor-huizhou01.igozhang.cn\
  --set expose.ingress.hosts.core=harbor-huizhou01.igozhang.cn\
  --set expose.ingress.ingressClassName=nginx \
  --set expose.ingress.annotations."kubernetes\.io/ingress\.class"=nginx \
  --set expose.tls.enabled=true \
  --set expose.tls.certSource=secret \
  --set expose.tls.secret.secretName=harbor-tls \
  --set notary.enabled=false \
  --set harborAdminPassword=Harbor@2026
完了还是用默认密码登录，set未生效
用户名：admin
密码：Harbor12345

次选：（版本不要太高，否则报错）
helm repo add harbor https://helm.goharbor.io
helm repo update
helm search repo harbor --versions
helm pull harbor/harbor --version 1.12.6
helm install harbor ./harbor-1.12.6.tgz \
  -n harbor \
  --set externalURL=https://harbor-huizhou01.igozhang.cn\
  --set expose.ingress.hosts.core=harbor-huizhou01.igozhang.cn\
  --set expose.tls.enabled=true \
  --set expose.tls.secretName=harbor-tls \
  --set harborAdminPassword=Harbor@2026
这样安装完后，镜像拉不到
helm uninstall harbor -n harbor
```

## 将本地images推到私有habor

```
原有镜像：
# docker images
REPOSITORY                                                                 TAG                               IMAGE ID       CREATED         SIZE
harbor-dianbai.sunwoda-evb.com/mom/6x575i8_prod                            r116                              b0574344e80b   6 days ago      1.31GB

打tag
# 方式1：用原镜像名打标签（推荐，更清晰）
docker tag harbor-dianbai.sunwoda-evb.com/mom/6x575i8_prod:r116 harbor-huizhou01.sunwoda-evb.com/pub/mom:r116
# 方式2：用镜像ID打标签（兜底，避免原镜像名复杂出错）
# docker tag b0574344e80b harbor-huizhou01.sunwoda-evb.com/pub/mom:r116

login才能推送
docker login harbor-huizhou01.sunwoda-evb.com
docker push harbor-huizhou01.sunwoda-evb.com/pub/mom:r116

可以在其他机器拉取验证：
docker pull harbor-huizhou01.sunwoda-evb.com/pub/mom:r116
```

```
配置免密拉取harbor镜像
1. 创建拉取密钥，前提先docker login才会生成json
kubectl create secret generic harbor-secret \
  --from-file=.dockerconfigjson=/root/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson
2. 全局免密（其余空间加一个 -n ns_name参数）
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "harbor-secret"}]}'

方式二：
kubectl create secret docker-registry harbor-hu-secret \
  --namespace=logcenter \
  --docker-server=harbor-hu.igozhang.cn \
  --docker-username=admin \
  --docker-password=Harbor@2026

kubectl patch serviceaccount default -n logcenter \
  -p '{"imagePullSecrets": [{"name": "harbor-hu-secret"}]}'

给所有ns 添加默认密钥
for ns in $(kubectl get ns -o jsonpath={.items[*].metadata.name}); do
  kubectl patch serviceaccount default -n $ns \
    -p '{"imagePullSecrets": [{"name": "harbor-hu-secret"}]}'
done
```
