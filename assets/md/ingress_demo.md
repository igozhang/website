# ingress_demo

## 需求说明

设计一个精简实验验证ingress可用性，所有配置写在一个yaml中  
指定POD 运行在node：k8s-worker55  
镜像使用： REPOSITORY TAG IMAGE ID CREATED SIZE  
registry.cn-shanghai.aliyuncs.com/labring/nginx 1.22.0 d62eacde2872 3 years ago 39.2 MB  
域名： igo.sunwoda-evb.com  
namespace： igo  
密钥 ：tls-igo-ingress

## 证书secret

```
kubectl create namespace igo
kubectl -n igo create secret \
  tls tls-igo-ingress \
  --cert=./tls.pem \
  --key=./tls.key
```

## 资源yaml

```
# cat ingress-test.yaml

# Nginx Deployment (指定运行在 k8s-worker55 节点)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-test
  namespace: igo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-test
  template:
    metadata:
      labels:
        app: nginx-test
    spec:
      # 强制调度到指定节点
      nodeSelector:
        kubernetes.io/hostname: k8s-worker55
      containers:
      - name: nginx
        image: docker.io/library/nginx:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 80
        # 健康检查确保容器可用
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 10
---
# 服务暴露 Pod
apiVersion: v1
kind: Service
metadata:
  name: nginx-test-svc
  namespace: igo
spec:
  selector:
    app: nginx-test
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
---
# Ingress 配置 (域名 + TLS)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-test-ingress
  namespace: igo
  # 若使用 nginx-ingress controller，需添加注解（根据实际 controller 调整）
  annotations:
    kubernetes.io/ingress.class: "nginx"  # 适配老版本 ingress-controller
    nginx.ingress.kubernetes.io/ssl-redirect: "true"  # 强制 HTTPS
spec:
  # TLS 配置（使用指定密钥）
  tls:
  - hosts:
    - igo.sunwoda-evb.com
    secretName: tls-igo-ingress
  rules:
  - host: igo.sunwoda-evb.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx-test-svc
            port:
              number: 80
```

## 验证

```
# 方式1：在集群内节点测试 (替换 INGRESS_CONTROLLER_IP 为实际 ingress-controller 地址)
curl -v https://igo.sunwoda-evb.com --resolve igo.sunwoda-evb.com:443:INGRESS_CONTROLLER_IP --insecure

# 方式2：配置本地 hosts (添加 igo.sunwoda-evb.com -> INGRESS_CONTROLLER_IP) 后，浏览器访问 https://igo.sunwoda-evb.com
```
