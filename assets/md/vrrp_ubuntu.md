# vrrp_ubuntu

6台Ubuntu 20.04 服务器 VRRP 多节点 VIP 高可用配置（纯内核版）  
6台机器完全可以通过 VRRP 实现 VIP 高可用！核心逻辑是：**6台机器加入同一个 VRRP 组，通过优先级选举唯一的 Master 持有 VIP，其余为 Backup；Master 故障时，优先级最高的 Backup 自动接管**。

## 环境规划（核心规则）

| 节点 | 物理IP | 网卡名（示例） | VRRP组ID | 优先级 | VIP | 角色说明 |
| --- | --- | --- | --- | --- | --- | --- |
| 节点1 | <10.201.81.54/24> | ens33 | 81 | 150 | <10.201.81.100/24> | 主Master（最高优先级） |
| 节点2 | <10.201.81.55/24> | ens33 | 81 | 140 | <10.201.81.100/24> | 备Master1 |
| 节点3 | <10.201.81.56/24> | ens33 | 81 | 130 | <10.201.81.100/24> | 备Master2 |
| 节点4 | <10.201.81.57/24> | ens33 | 81 | 120 | <10.201.81.100/24> | 备Master3 |
| 节点5 | <10.201.81.58/24> | ens33 | 81 | 110 | <10.201.81.100/24> | 备Master4 |
| 节点6 | <10.201.81.59/24> | ens33 | 81 | 100 | <10.201.81.100/24> | 最低优先级备份 |

### 关键规则

1. 所有节点的 `VRRP组ID（vrid）` 必须一致（示例用81，范围1-255）；
2. 优先级值越大，成为Master的优先级越高（建议间隔10，方便后续调整）；
3. 所有节点的VIP、网卡名、通告间隔必须一致；
4. 仅Master节点对外响应VIP的ARP请求，Backup仅监听不响应。

---

## 第一步：所有节点前置准备（统一执行）

### 1. 检查/加载VRRP内核模块

```
# 检查内核是否支持VRRP
modinfo vrrp

# 加载VRRP模块（临时）
sudo modprobe vrrp

# 设置开机自动加载
echo "vrrp" | sudo tee /etc/modules-load.d/vrrp.conf
```

### 2. 放行VRRP流量（防火墙配置）

```
# 临时关闭ufw（测试环境）
sudo ufw disable

# 永久放行VRRP协议（协议号112）和组播地址224.0.0.18
sudo iptables -A INPUT -p 112 -j ACCEPT
sudo iptables -A INPUT -d 224.0.0.18/32 -j ACCEPT

# 保存iptables规则（Ubuntu 20.04需先装iptables-persistent）
sudo apt update && sudo apt install iptables-persistent -y
sudo iptables-save | sudo tee /etc/iptables/rules.v4
```

### 3. 确认网卡名称（替换为实际名称）

```
ip addr  # 查看实际网卡名（如ens33、eth0等，记下来后续替换）
```

---

## 第二步：各节点VRRP实例配置（分节点执行）

### 核心规则

- 仅**优先级**不同，其余参数（vrid、VIP、通告间隔、网卡名）完全一致；
- 命令为临时生效，后续需配置开机自启。

### 1. 节点1（<10.201.81.54>，优先级150）

```
# 创建VRRP实例（vrrp81为实例名，vrid=81，优先级150）
sudo ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 150

# 设置VRRP通告间隔（1秒，所有节点保持一致）
sudo ip link set vrrp81 arp_interval 1000

# 绑定VIP到VRRP实例
sudo ip addr add 10.201.81.100/24 dev vrrp81

# 启用VRRP实例
sudo ip link set vrrp81 up

# 验证配置
ip -d link show vrrp81
ip addr show vrrp81
```

### 2. 节点2（<10.201.81.55>，优先级140）

```
sudo ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 140
sudo ip link set vrrp81 arp_interval 1000
sudo ip addr add 10.201.81.100/24 dev vrrp81
sudo ip link set vrrp81 up
```

### 3. 节点3（<10.201.81.56>，优先级130）

```
sudo ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 130
sudo ip link set vrrp81 arp_interval 1000
sudo ip addr add 10.201.81.100/24 dev vrrp81
sudo ip link set vrrp81 up
```

### 4. 节点4（<10.201.81.57>，优先级120）

```
sudo ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 120
sudo ip link set vrrp81 arp_interval 1000
sudo ip addr add 10.201.81.100/24 dev vrrp81
sudo ip link set vrrp81 up
```

### 5. 节点5（<10.201.81.58>，优先级110）

```
sudo ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 110
sudo ip link set vrrp81 arp_interval 1000
sudo ip addr add 10.201.81.100/24 dev vrrp81
sudo ip link set vrrp81 up
```

### 6. 节点6（<10.201.81.59>，优先级100）

```
sudo ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 100
sudo ip link set vrrp81 arp_interval 1000
sudo ip addr add 10.201.81.100/24 dev vrrp81
sudo ip link set vrrp81 up
```

---

## 第三步：配置开机自动生效（所有节点）

临时命令重启后失效，需写入 `/etc/rc.local` 实现开机自启。

### 1. 新建/编辑rc.local文件

```
sudo vim /etc/rc.local
```

### 2. 各节点rc.local内容（仅优先级不同）

#### 节点1（54）rc.local内容

```
#!/bin/bash
# 加载VRRP模块
modprobe vrrp
# 等待网卡初始化（避免开机时序问题）
sleep 8
# 配置VRRP实例
ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 150
ip link set vrrp81 arp_interval 1000
ip addr add 10.201.81.100/24 dev vrrp81
ip link set vrrp81 up
# 放行VRRP流量
iptables -A INPUT -p 112 -j ACCEPT
iptables -A INPUT -d 224.0.0.18/32 -j ACCEPT
exit 0
```

#### 节点2（55）rc.local内容（仅改优先级）

```
#!/bin/bash
modprobe vrrp
sleep 8
ip link add name vrrp81 link ens33 type vrrp vrid 81 priority 140
ip link set vrrp81 arp_interval 1000
ip addr add 10.201.81.100/24 dev vrrp81
ip link set vrrp81 up
iptables -A INPUT -p 112 -j ACCEPT
iptables -A INPUT -d 224.0.0.18/32 -j ACCEPT
exit 0
```

#### 节点3（56）：priority 130

#### 节点4（57）：priority 120

#### 节点5（58）：priority 110

#### 节点6（59）：priority 100

（其余内容和节点1完全一致，仅修改priority值）

### 3. 赋予执行权限并启用rc.local服务

```
# 添加执行权限
sudo chmod +x /etc/rc.local

# 启用rc-local服务（Ubuntu 20.04默认禁用）
sudo systemctl enable rc-local
sudo systemctl start rc-local

# 验证服务状态
sudo systemctl status rc-local  # 显示active(running)即正常
```

---

## 第四步：测试多节点高可用切换（核心验证）

### 1. 正常状态验证

```
# 在任意客户端ping VIP（验证连通性）
ping 10.201.81.100

# 在所有节点查看VRRP状态（仅节点1显示Master）
ip -d link show vrrp81
# 输出中节点1会显示"vrrp master"，其余节点显示"vrrp backup"
```

### 2. 模拟节点1（主Master）故障

```
# 在节点1执行（关闭VRRP实例）
sudo ip link set vrrp81 down

# 或直接关机：sudo poweroff
```

### 3. 验证切换效果

- 等待3秒（内核默认超时3个通告间隔），客户端ping VIP仅丢1-2个包后恢复；
- 在节点2执行 `ip -d link show vrrp81`，会显示"vrrp master"（节点2接管VIP）；
- 查看VIP响应：`arp -an | grep 10.201.81.100`，显示节点2的MAC地址。

### 4. 模拟级联故障（验证多备份切换）

```
# 关闭节点2的VRRP实例
sudo ip link set vrrp81 down
# 节点3会自动接管VIP（优先级130最高）
```

### 5. 恢复故障节点

```
# 重启节点1的VRRP实例
sudo ip link set vrrp81 up
# 节点1会因优先级最高（150），自动抢占回VIP（默认开启抢占模式）
```

---

## 关键注意事项（避坑指南）

1. **组播互通**：确保6台机器在同一二层网络，路由器/交换机未禁用组播（<224.0.0.18>）；
2. **优先级规划**：建议间隔10设置（如150/140/130），方便后续新增节点或调整优先级；
3. **抢占模式**：Linux内核VRRP默认开启抢占模式，高优先级节点恢复后会自动抢回VIP；
4. **切换延迟**：默认3秒（3\*1秒通告间隔），可通过调整`arp_interval`（如500ms）缩短，但不建议小于500ms；
5. **网卡名称**：务必替换为实际网卡名（如eth0、ens192等），否则配置失效。

---

### 总结

1. **核心逻辑**：6台节点加入同一VRRP组，通过优先级选举Master持有VIP，故障时按优先级顺位接管，实现多层备份的高可用；
2. **配置关键**：所有节点的VRRP组ID、VIP、通告间隔必须一致，仅优先级不同；
3. **稳定性**：纯内核层VRRP无用户态进程依赖，比keepalived更稳定，6节点场景下切换逻辑与2节点完全一致，仅多了优先级顺位。

如果需要调整切换延迟、关闭抢占模式（避免频繁切换），或添加服务健康检查（如检测Nginx/数据库状态后再切换），可以告诉我，我会补充对应配置。

> （注：文档部分内容可能由 AI 生成）
