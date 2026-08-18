# Calico BIRD 进程内存泄漏问题报告

## 1\. 问题现象

- **环境信息**：Kubernetes 集群，k8s\-master52 节点，16GB 物理内存，Calico Operator 部署（`calico\-system`命名空间），BIRD 1\.6\.8 版本，系统运行 46 天

- **核心问题**：服务器内存占用 100%，`free \-g`显示可用内存为 0

- **进程异常**：`bird`和`bird6`进程各占用 38\.1% 系统内存，合计约 12\.2GB

- **异常特征**：BIRD 进程路由表为空（0 条路由），但内存持续暴涨

- **风险等级**：高，随时可能触发系统 OOM 杀死关键进程

## 2\. 问题原因

- **根本原因**：Calico 集成的**BIRD 1\.6\.8 版本存在已知严重内存泄漏 bug**

- **触发条件**：进程长期运行（通常超过 30 天），即使无路由条目也会持续消耗内存

- **影响范围**：所有使用 Calico 3\.x 系列版本且未升级 BIRD 组件的 Kubernetes 集群

## 3\. 解决方案

### 紧急处理（按顺序执行）

1. **定位内存泄漏的 BIRD 进程**

```bash
# 查找占用内存最高的bird和bird6进程
ps aux --sort=-rss | grep -E 'bird|bird6' | head -n 2
```

2. **强制杀死内存泄漏进程**

```bash
# 替换为上一步查到的实际PID
kill -9 <bird_pid> <bird6_pid>
```

3. **验证进程自动恢复**

```bash
# 等待5秒后检查新启动的BIRD进程内存使用
sleep 5 && ps aux | grep -E 'bird|bird6'
```

- **处理效果**：Calico 会自动重新启动 BIRD 进程，内存立即恢复至正常水平（每个进程约 10\-20MB）

- **影响评估**：网络中断时间 \&lt; 5 秒，不影响正在运行的 Pod，不丢失任何网络配置

### 永久禁用 IPv6 BIRD 进程（推荐）

如果集群不使用 IPv6 网络，可彻底禁用 bird6 进程避免再次泄漏：

```bash
# 修改Calico Operator部署的配置（正确命名空间）
kubectl patch daemonset calico-node -n calico-system --type=json -p='[{"op": "add", "path": "/spec/template/spec/containers/0/env/-", "value": {"name": "CALICO_IPV6POOL_CIDR", "value": "none"}}]'

# 重启所有Calico节点使配置生效
kubectl rollout restart daemonset calico-node -n calico-system

# 手动杀死所有节点上已存在的bird6进程（配置只对新容器生效）
for node in $(kubectl get nodes -o name); do
  kubectl exec -n calico-system $(kubectl get pods -n calico-system -l k8s-app=calico-node -o name | grep ${node#node/}) -- killall bird6 2>/dev/null || true
done
```

### 验证配置生效

```bash
# 检查环境变量是否已正确设置
kubectl describe daemonset calico-node -n calico-system | grep -A 2 CALICO_IPV6POOL_CIDR

# 在所有节点验证bird6进程已停止
ps aux | grep -c bird6
```

**预期输出**：`0`

## 4\. 后续预防

1. **版本升级**：将 Calico 升级至 3\.25 \+ 版本，该版本已修复 BIRD 内存泄漏问题

2. **监控告警**：添加 BIRD 进程内存使用监控，设置单进程 \&gt; 500MB 告警阈值

3. **临时措施**：在升级完成前，设置每月一次的 BIRD 进程定时重启任务

4. **配置优化**：确认集群网络需求，禁用不需要的 IPv6 功能

## 代码
```shell
批量检查:
ansible k8s_pre -f 1 -m shell -a "echo -e '===== 节点：{{ inventory_hostname }} ====='; free -h; echo -e '\n--- BIRD 进程内存占用 ---'; ps aux --sort=-rss | grep -E 'bird|bird6' | head -n 3; echo -e '\n--- BIRD 路由表状态 ---'; birdc show route count 2>/dev/null || echo 'birdc未找到'; birdc6 show route count 2>/dev/null || echo 'birdc6未找到'; echo -e '\n--- bird6 进程状态 ---'; ps aux | grep -c bird6; echo ''"
批量杀死bird,bird6进程（进程会自己重启，无需担心）
ansible k8s_pre -f 1 -m shell -a "ps aux | grep -E 'bird|bird6' | grep -v grep | awk '{print \$2}' | xargs kill -9 2>/dev/null || true"
禁用ipv6_bird进程；
kubectl patch daemonset calico-node -n calico-system --type=json -p='[{"op": "add", "path": "/spec/template/spec/containers/0/env/-", "value": {"name": "CALICO_IPV6POOL_CIDR", "value": "none"}}]'
kubectl rollout restart daemonset calico-node -n calico-system


```