# jenkins_k8s

k8s\_1.24

## 安装

```
sealos run registry.cn-shanghai.aliyuncs.com/labring/jenkins:v2.401.3 \
-e HELM_OPTS="--set controller.ingress.enabled=true --set controller.ingress.hostName=jenkins-hu.igozhang.cn --set controller.ingress.ingressClassName=nginx --set controller.ingress.path=/ --set controller.ingress.pathType=Prefix --set controller.installLatestPlugins=false --set controller.installLatestSpecifiedPlugins=false --set controller.overwritePlugins=true --set-string controller.installPlugins[0]=workflow-aggregator:596.v8c21c963d92d --set-string controller.installPlugins[1]=git:5.1.0"

--set controller.ingress.hostName=jenkins-hu.sunwoda-evb.com：指定对外访问 Jenkins 的域名。
--set controller.ingress.path=/：配置 Ingress 匹配根路径 /。
--set controller.ingress.pathType=Prefix：使用前缀匹配路径（/ 会匹配所有子路径）。
--set controller.installLatestPlugins=false：关闭“自动把依赖插件升级到最新版本”。

helm upgrade --install jenkins jenkins/jenkins \
  --namespace jenkins \
  --version 5.8.115 \
  --create-namespace \
  --set controller.ingress.enabled=true \
  --set controller.ingress.hostName=jenkins-hu.igozhang.cn \
  --set controller.ingress.path=/ \
  --set controller.ingress.ingressClassName=nginx \
  --set persistence.size=20Gi \
  --set-json 'controller.installPlugins=[]'

helm uninstall jenkins -n jenkins
```

## 建立一个发布流程

代码提交 → 拉取代码 → 编译 → 构建镜像 → 推仓库 → 更新 K8s 镜像 → 发布 → 验证  
先拿一个模块来做

## OUTPUT

```
NAME: jenkins
LAST DEPLOYED: Sun Apr 19 09:41:41 2026
NAMESPACE: jenkins
STATUS: deployed
REVISION: 1
NOTES:
1. Get your 'admin' user password by running:
  kubectl exec --namespace jenkins -it svc/jenkins -c jenkins -- /bin/cat /run/secrets/additional/chart-admin-password && echo
2. Visit http://jenkins-hu.igozhang.cn

3. Login with the password from step 1 and the username: admin
4. Configure security realm and authorization strategy
5. Use Jenkins Configuration as Code by specifying configScripts in your values.yaml file, see documentation: http://jenkins-hu.sunwoda-evb.com/configuration-as-code and examples: https://github.com/jenkinsci/configuration-as-code-plugin/tree/master/demos

For more information on running Jenkins on Kubernetes, visit:
https://cloud.google.com/solutions/jenkins-on-container-engine

For more information about Jenkins Configuration as Code, visit:
https://jenkins.io/projects/jcasc/

NOTE: Consider using a custom image with pre-installed plugins
```
