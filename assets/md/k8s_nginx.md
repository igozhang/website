# 基于K8s实现nginx静态页域名访问

## 简介

本文档详细记录了如何在Docker Desktop的单节点Kubernetes集群上部署一个简单的Nginx静态页面，并通过域名访问的完整实现过程。

## 环境准备

### 前提条件

- 已安装Docker Desktop
- 已在Docker Desktop中启用Kubernetes功能
- 已配置好kubectl命令行工具

### 环境检查

1. 检查Docker状态

```
docker info
```

2. 检查Kubernetes集群状态

```
kubectl cluster-info
```

## 实现步骤

### 步骤1：创建静态HTML页面

创建一个简单的静态HTML页面，作为Nginx的默认首页。

```
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to Nginx on Kubernetes</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f0f0f0;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
        }
        .container {
            background-color: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            text-align: center;
        }
        h1 {
            color: #333;
        }
        p {
            color: #666;
        }
        .info {
            margin-top: 20px;
            padding: 10px;
            background-color: #f9f9f9;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Welcome to Nginx on Kubernetes!</h1>
        <p>This is a simple static page served by Nginx running on a single-node Kubernetes cluster.</p>
        <div class="info">
            <p><strong>Environment:</strong> Docker Desktop Kubernetes</p>
            <p><strong>Date:</strong> 2026-02-19</p>
        </div>
    </div>
</body>
</html>
```

将上述内容保存为 `index.html` 文件。

### 步骤2：编写Dockerfile构建Nginx镜像

创建一个Dockerfile，使用官方Nginx镜像作为基础，并将我们的静态页面复制到Nginx的默认静态文件目录。

```
FROM nginx:alpine

COPY index.html /usr/share/nginx/html/

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

将上述内容保存为 `Dockerfile` 文件。

### 步骤3：构建Docker镜像

使用Docker命令构建包含静态页面的Nginx镜像。

```
docker build -t nginx-static .
```

验证镜像是否构建成功：

```
docker images | grep nginx-static
```

### 步骤4：创建Kubernetes部署配置

创建一个Kubernetes Deployment配置文件，定义如何部署我们的Nginx应用。

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-static-deployment
  labels:
    app: nginx-static
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-static
  template:
    metadata:
      labels:
        app: nginx-static
    spec:
      containers:
      - name: nginx-static
        image: nginx-static:latest
        ports:
        - containerPort: 80
        imagePullPolicy: Never  # 使用本地构建的镜像，不尝试从远程拉取
```

将上述内容保存为 `nginx-deployment.yaml` 文件。

### 步骤5：创建Kubernetes服务配置

创建一个Kubernetes Service配置文件，使用NodePort类型来暴露服务。

```
apiVersion: v1
kind: Service
metadata:
  name: nginx-static-service
  labels:
    app: nginx-static
spec:
  selector:
    app: nginx-static
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
    name: http
  type: NodePort
```

将上述内容保存为 `nginx-service.yaml` 文件。

### 步骤6：应用Kubernetes配置并验证

1. 应用Deployment配置：

```
kubectl apply -f nginx-deployment.yaml
```

2. 应用Service配置：

```
kubectl apply -f nginx-service.yaml
```

3. 验证Deployment状态：

```
kubectl get deployments nginx-static-deployment
```

4. 验证Pod状态：

```
kubectl get pods -l app=nginx-static
```

5. 验证Service状态，获取NodePort端口：

```
kubectl get service nginx-static-service
```

### 步骤7：配置域名访问

1. 编辑本地hosts文件，添加域名映射：

```
sudo bash -c 'echo "127.0.0.1 nginx-static.local" >> /etc/hosts'
```

2. 验证服务是否可以通过IP和端口访问：

```
curl http://localhost:32513  # 注意：32513是示例端口，实际端口以kubectl get service命令的输出为准
```

3. 验证服务是否可以通过域名访问：

```
curl http://nginx-static.local:32513  # 注意：32513是示例端口，实际端口以kubectl get service命令的输出为准
```

## 总结

通过以上步骤，我们成功实现了：

1. 在Docker Desktop的单节点Kubernetes集群上部署了一个Nginx静态页面
2. 通过NodePort类型的Service暴露了服务
3. 配置了域名访问（通过修改hosts文件）

## 后续步骤

如果需要进一步优化，可以考虑：

1. 使用Ingress控制器替代NodePort，实现更灵活的路由和域名管理
2. 添加配置管理，实现静态页面内容的动态更新
3. 配置资源限制和健康检查，提高应用的稳定性
4. 实现CI/CD流程，自动化部署过程

## 故障排查

如果遇到问题，可以尝试以下方法：

1. 检查Pod状态和日志：

```
   kubectl describe pod <pod-name>
   kubectl logs <pod-name>
```

2. 检查Service状态：

```
   kubectl describe service nginx-static-service
```

3. 检查Docker镜像是否存在：

```
   docker images | grep nginx-static
```

4. 验证本地hosts文件配置：

```
   cat /etc/hosts | grep nginx-static
```
