# Kubernetes 增删改查（K8s CRUD）

> Kubernetes 中所有的内容都抽象为资源，资源实例化（被调用、被执行了）之后，叫做对象。

---

## 环境配置

### 配置 kubectl 命令别名

```bash
echo "alias k='kubectl'" >> ~/.bashrc
echo "alias kd='kubectl describe'" >> ~/.bashrc
echo "alias kg='kubectl get'" >> ~/.bashrc
echo "alias ka='kubectl apply'" >> ~/.bashrc
echo "alias kdl='kubectl delete'" >> ~/.bashrc
source ~/.bashrc
```

### 创建普通用户并授权

```bash
useradd -m -s /bin/bash igo 2>/dev/null || true
echo 'igo:igo@123' | chpasswd
printf '%s\n' 'igo ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/igo
chmod 440 /etc/sudoers.d/igo
mkdir -p /home/igo/.kube
chown -R igo.igo /home/igo
cp /root/.kube/config /home/igo/.kube/
```

### 一条命令加载 kubectl 并创建 config 目录

```bash
# 版本号从 docker ps | grep rancher 获取
VER=v1.19.16 
VER=v1.25.16 && \
curl -fsSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/${VER}/bin/linux/amd64/kubectl" && \
chmod +x /usr/local/bin/kubectl && \
kubectl version --client 2>&1 | head -15 && \
mkdir -p /root/.kube

cp kubectlv1.25.16 /usr/local/bin/kubectl
chmod +x /usr/local/bin/kubectl && \
kubectl version --client 2>&1 | head -15
```

### RKE 集群环境

**集群管理员权限：**

```bash
cp /etc/kubernetes/ssl/kubecfg-kube-controller-manager.yaml ~/.kube/config
```

**节点权限：**

```bash
cp /etc/kubernetes/ssl/kubecfg-kube-node.yaml ~/.kube/config
```

### RKE2 集群环境

```bash
echo "export PATH=\$PATH:/var/lib/rancher/rke2/bin" >> ~/.bashrc
echo "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml" >> ~/.bashrc
source ~/.bashrc
```

---

## 查询

### 节点资源查询

**查看已分配的资源：**

```bash
kubectl describe node k8s-pre-34 | grep -A 15 "Allocated resources"
```

```bash
kubectl describe nodes | awk '/^Name:/{print "\n" $0} /^Allocated resources:/{p=1} p{print} /^Events:/{p=0}'
```

**查看节点 CPU / 内存：**

```bash
kubectl get nodes -o custom-columns=NAME:.metadata.name,CPU:.status.allocatable.cpu,MEMORY:.status.allocatable.memory | \
  awk 'NR==1; NR>1{printf "%s\t%s\t%.1fGi\n", $1, $2, $3/1024/1024}' | column -t
```

> 节点扩容后需要重启 kubelet 才会生效：
>
> ```bash
> ansible 10.80.238.34,10.80.238.36 -m shell -a "systemctl restart kubelet"
> ```

### Pod 相关查询

**列出 Pod 里的 containers：**

```bash
kubectl get pod jenkins-0 -n jenkins -o jsonpath="{range .spec.containers[*]}{.name}{'\t'}{.image}{'\n'}{end}"
```

**查看 Deployment 环境变量：**

```bash
kubectl get deploy mes-server-deployment -n mes-app -o jsonpath='{.spec.template.spec.containers[0].env}'
```

---

## 增加（创建）

### 创建资源

```bash
kubectl create -f kubernetes-demo.yaml               # 使用预定义文件创建资源
kubectl apply -f ./igo01.yaml -f ./igo02.yaml       # 使用多个文件
kubectl apply -f https://k8s.io/examples/application/mysql/mysql-pv.yaml
kubectl run nginx --image=nginx --replicas=3 --dry-run
```

### 交互式创建 MySQL 客户端 Pod

```bash
kubectl run -it --rm --image=mysql:5.6 --restart=Never mysql-client -- mysql -h mysql -ppassword
```

### 创建 CronJob

```bash
kubectl create cronjob hello \
  --image=busybox:1.28 \
  --schedule="*/1 * * * *" \
  -- echo "Hello World"
```

### 给 Pod 添加标签

```bash
kubectl label pods/pod-demo env=production
```

### YAML 方式创建多个 Pod

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  containers:
  - name: busybox
    image: busybox:1.28
    args:
    - sleep
    - "1000000"
---
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep-less
spec:
  containers:
  - name: busybox
    image: busybox:1.28
    args:
    - sleep
    - "1000"
EOF
```

### 设置环境变量

```bash
kubectl set env deploy/mes-server-deployment TZ=Europe/Budapest -n mes-app
kubectl set env deploy --all TZ=Europe/Budapest -n mes-app
```

---

## 修改

### 修改标签

```bash
kubectl label pods/pod-demo env=testing --overwrite
```

### 副本扩容 / 缩容

```bash
kubectl scale --replicas=3 rs/igo                      # 将 ig o副本集调整到 3 副本
kubectl scale --current-replicas=2 --replicas=3 deployment/mysql   # 如果当前是 2，调整到 3
```

### 端口转发

监听本地 5000 转发到 ig o-deployment 创建的 Pod 的 6000 端口：

```bash
kubectl port-forward deploy/igo-deployment 5000:6000
```

### Deployment 扩容（补充）

```bash
# 查看 Deployment 副本数
kubectl get deployment rancher -n cattle-system

# 示例：将 cattle-system 命名空间下的 rancher Deployment 副本数改为 3
kubectl scale deployment rancher -n cattle-system --replicas=3

# 编辑 Deployment
kubectl edit deployment rancher -n cattle-system
```

---

## 删除

### 删除资源

```bash
kubectl label pods igo-pod bar-                      # 删除 igo-pod 容器的 bar 标签
kubectl delete pod igo-pod --now                     # 立刻删除
kubectl delete pods,deployment,svc -l name=myLabel   # 按标签批量删除
kubectl delete -f ./igopod.json                     # 根据文件删除
```

### 删除 PVC / PV

```bash
kubectl delete pvc mysql-pv-claim
kubectl delete pv mysql-pv-volume
```

---

## 交互

### 进入 Pod 执行命令

```bash
kubectl exec --stdin --tty igo-pod -- /bin/sh
```

### 查看资源使用量（按 CPU 排序）

```bash
kubectl top pod POD_NAME --sort-by=cpu
```

### 文件拷贝

```bash
kubectl cp /tmp/foo_dir igo-pod:/tmp/bar_dir
```
