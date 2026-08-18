# Doris 集群节点间 Ping 偶发延迟尖峰 — 问题报告

**状态：已修复**（2026-06-03，101~105 均已切换 bond 活动口至 ens2f0）

---

## ENV（环境信息）

| 项目 | 版本/配置 |
|------|-----------|
| **操作系统** | Rocky Linux 9.3 (Blue Onyx) |
| **内核** | 5.14.0-362.8.1.el9_3.x86_64 |
| **集群节点** | doris101 ~ doris105（10.10.243.101 ~ 10.10.243.105/24） |
| **网关** | 10.10.243.254 |
| **DNS** | 10.10.240.70 / 10.10.240.71 |
| **网络拓扑** | 同网段、同 TOR 交换机 |
| **Bond** | bond0，`active-backup`（主备），双 10G 物理口 |
| **物理网卡** | ens1f0 + ens2f0（Broadcom bnx2x，Chip Revision: **everest3**） |
| **网卡驱动** | bnx2x（随内核 5.14.0-362.8.1.el9_3） |
| **网卡固件** | mbi **7.15.42** / bc **7.15.23**（assert 时 dump 记录为 FW **7_13_21**） |
| **链路** | 10000 Mbps / Full Duplex / Flow control ON |
| **中断合并** | rx-usecs=24，tx-usecs=48 |
| **网络管理** | NetworkManager |
| **Bond 主配置（含 `[bond]` 段）** | `/etc/NetworkManager/system-connections/bond0.nmconnection` |
| **Slave 配置** | `bond-slave-ens1f0.nmconnection`、`bond-slave-ens2f0.nmconnection`（同目录） |

---

## 1. 问题概述

| 项目 | 内容 |
|------|------|
| **现象** | 同网段节点互 ping，RTT 多数 **0.1~0.3 ms**，偶发 **2~26 ms** 尖峰，**无丢包** |
| **修复前** | bond 活动口均为 **ens1f0**，`Primary Slave: None` |
| **修复后** | bond 活动口均为 **ens2f0**，`Primary Slave: ens2f0` |
| **影响** | ping 尖峰本身影响有限；网卡 assert 严重时可能影响 BE 间 RPC、副本同步 tail latency |

---

## 2. 问题发现与确认

### 2.1 初步现象

**doris101 ping doris103（修复前）：**

```text
多数包: 0.1~0.3 ms
偶发尖峰: 10.2 ms、10.9 ms、14.6 ms
丢包率: 0%，TTL: 64
```

**doris105 ping doris101（修复前，105 仍在 ens1f0）：**

```text
rtt min/avg/max/mdev = 2.799/17.697/26.504/10.592 ms
```

**判断：** 非持续拥塞或物理断链，更像 **单包在故障网口上被延迟**。

### 2.2 网络路径确认

- 流量经 **bond0** 发出，修复前活动 slave 均为 **ens1f0**
- bond 模式：`fault-tolerance (active-backup)`，未设 Primary
- 备口 **ens2f0** 链路 up，但未承担流量
- `rx_crc_errors: 0` → 基本排除光纤/模块物理误码

### 2.3 根因：Broadcom bnx2x 活动网口 ens1f0 异常

#### 内核日志证据

**doris103（2026-06-02 12:51:29，完整故障链）：**

```text
bnx2x: [bnx2x_mc_assert:749(ens1f0)]CSTORM_ASSERT_INDEX ...
bnx2x: [bnx2x_mc_assert:749(ens1f0)]USTORM_ASSERT_INDEX ...
bnx2x: [bnx2x_mc_assert:759(ens1f0)]Chip Revision: everest3, FW Version: 7_13_21
bnx2x: [bnx2x_panic_dump:1202(ens1f0)]end crash dump -----------------
bnx2x: [bnx2x_sp_rtnl_task:10281(ens1f0)]Indicating link is down due to Tx-timeout
```

| 日志关键字 | 含义 |
|-----------|------|
| `bnx2x_mc_assert` | 网卡固件 microcode 内部 assert |
| `bnx2x_panic_dump ... end crash dump` | 固件崩溃转储，驱动随后 reset 网卡 |
| `Tx-timeout` | 发送队列超时，链路被判定 down 后恢复 |

#### 网卡硬件错误计数（修复前，活动口 ens1f0）

| 节点 | rx_discards | unrecoverable_errors | 备口 ens2f0 |
|------|-------------|----------------------|-------------|
| doris101 | 134,235,720 | 2 | 0 / 0 |
| doris103 | 219,677,808 | 3 | 0 / 0 |
| doris105 | 410,864,782 | 1 | 0 / 0 |

**结论：** 各节点 **ens1f0** 均有 assert 历史及上亿 rx_discards；**ens2f0 全部正常**。bond 未设 Primary，流量持续走故障口。

### 2.4 根因判断：ens1f0 是如何逐渐变坏的（五个阶段）

> 以下为基于日志、计数、bond 配置与修复结果的 **机制推测**。已证实事实与推断分层如下：

| 层级 | 内容 | 可信度 |
|------|------|--------|
| **已证实** | ens1f0 有 assert/Tx-timeout、上亿 rx_discards；ens2f0 长期为 0；切 ens2f0 后 ping 恢复 | 高 |
| **强推断** | 五台节点均 ens1f0 坏、ens2f0 好 → 与 bond 长期只走 ens1f0 高度吻合 | 中高 |
| **待验证** | 固件 bug 具体版本、PCIe 槽位差异、哪类流量触发 assert | 中 |

#### 2.4.1 为什么同型号网卡，偏偏网卡 1（ens1f0）坏？

两块卡 **芯片型号、驱动、固件版本相同**，但运行条件不同：

| 对比项 | ens1f0（网卡 1） | ens2f0（网卡 2） |
|--------|------------------|------------------|
| bond 角色 | **Active，自开机起承担 100% 业务流量** | **Backup，仅链路 up，几乎不传数据** |
| PCIe 位置 | 0000:**01**:00.0 | 0000:**22**:00.0（不同 Root Complex） |
| 中断/队列 | 多次 reload，fp 队列 7→15→29 变化 | 未经历同等冲击 |
| rx_discards | 1~4 亿 | 0（切换前） |
| unrecoverable_errors | 1~3 | 0 |
| assert 日志 | 有 | 无 |

**判断：** 不是“网卡 1 型号差”，而是 **网卡 1 从部署起持续扛全量流量，网卡 2 长期热备空闲**。备口“没坏”不等于“更耐造”，可能只是 **尚未被同等负载考验**。

#### 2.4.2 五个演化阶段（推测时间线）

```text
阶段 0：部署期
  bond active-backup，未设 primary → 默认 ens1f0 成为 Active
  ens2f0 仅 MII 监测链路，无数据面流量
        ↓
阶段 1：长期高负载（数周~数月）
  Doris BE 副本同步 / RPC / 导入 等持续产生网络 I/O
  网卡 RX/TX ring 持续运转，MSI-X 中断频繁
  表现：ping 多数仍正常（0.1~0.3 ms），偶发小尖峰
        ↓
阶段 2：RX ring 开始跟不上（discards 持续累积）
  突发流量时 RX ring / 内部 BRB buffer 满 → rx_discards 累加
  单包需等待 ring 空位或重传 → ping 偶发 2~10 ms
  （dmesg 可能尚无 assert，但计数已上亿，说明本阶段持续很久）
        ↓
阶段 3：固件内部状态异常（assert 触发点）
  高 PPS + 驱动/固件边界条件下，microcode 内部引擎异常
  → bnx2x_mc_assert（CSTORM/USTORM）
  → bnx2x_panic_dump
  → unrecoverable_errors +1
        ↓
阶段 4：发送路径卡死（Tx-timeout）
  固件 reset 前后，TX 队列无法及时完成
  → "Indicating link is down due to Tx-timeout"
  → 链路闪断 / 驱动 reload（dmesg 中 fp 队列数变化）
        ↓
阶段 5：亚健康态（问题暴露期）
  链路恢复，空闲时计数不再涨
  但 assert 口已不可靠 → ping 仍偶发 10~26 ms
  切换 ens2f0 后恢复
```

**要点：** 并非“某天突然坏”，而是 **阶段 2 可能持续很久**（rx_discards 上亿），阶段 3~4 为 **数次致命事件**；现场现象为 **阶段 2 与阶段 5 叠加**。

#### 2.4.3 各阶段与可观测证据的对应

| 阶段 | 机制 | 可观测现象 | 本集群证据 |
|------|------|------------|------------|
| 0 部署 | 默认 Active 走 ens1f0 | bond 无 primary | `Primary Slave: None` |
| 1 高负载 | ring/中断持续工作 | ping 基线正常 | 多数 RTT 0.1~0.3 ms |
| 2 ring 溢出 | buffer 满丢包 | rx_discards 增大 | ens1f0 累计 1~4 亿 |
| 3 固件 assert | microcode 崩溃 | mc_assert、panic_dump | 101/103/105 dmesg |
| 4 Tx-timeout | 发送队列卡死 | 链路 down、驱动 reload | 同上 + fp 队列数变化 |
| 5 亚健康 | 口不可靠但未持续恶化 | 空闲计数不涨、负载下尖峰 | 10s 采样不变，ping >10ms |

#### 2.4.4 为何五台节点都是 ens1f0 坏？

说明 **非单台硬件抽奖**，而是 **同构配置导致同构故障**：

1. 相同服务器型号 → ens1f0 均在 PCIe 01:00.0  
2. 相同 bond 配置 → 均未设 primary，均默认走 ens1f0  
3. 相同业务（Doris）→ 类似网络负载  
4. 相同固件/驱动 → 相同 bug 触发面  

**综合判断：** 直接原因是 ens1f0 固件 assert 与 Tx-timeout；深层原因是 **主备 bond 长期 100% 流量集中于 ens1f0**，在高 PPS/高吞吐下触发 bnx2x 固件缺陷，经阶段 2~5 逐步恶化。如何避免 ens2f0 重蹈覆辙，见 **附录 B**。

---

## 3. 证据采集命令

```bash
# assert / crash dump / Tx-timeout（推荐 journalctl，重启后仍可能保留）
sudo journalctl -k --no-pager | grep -E "bnx2x_mc_assert|bnx2x_panic_dump|Tx-timeout|Chip Revision" | tail -60

# 日志 + 硬件计数交叉验证
sudo bash -c 'echo "=== $(hostname) ens1f0 ==="; ethtool -S ens1f0 | awk "/^[[:space:]]+rx_discards:/ && !/\[/ {print}"; ethtool -S ens1f0 | grep unrecoverable_errors; journalctl -k --no-pager | grep -E "bnx2x_mc_assert|Tx-timeout" | tail -5'

# bond 状态
grep -E "Primary Slave|Currently Active Slave" /proc/net/bonding/bond0

# 错误计数 10s 采样（是否仍在恶化）
sudo bash -c 'echo T1; ethtool -S ens1f0 | awk "/^[[:space:]]+rx_discards:/ && !/\[/ {print}"; sleep 10; echo T2; ethtool -S ens1f0 | awk "/^[[:space:]]+rx_discards:/ && !/\[/ {print}"'
```

---

## 4. 修复方案与实施

### 4.1 修复思路

将 bond **Primary / Active Slave** 从故障口 **ens1f0** 切换至健康备口 **ens2f0**，并持久化配置。**切换不修复 ens1f0**，仅 bypass 故障口。

### 4.2 临时切换（立即生效）

NetworkManager 管理的 bond 推荐 **仅设置 primary**（设置 primary 后驱动会自动切换活动口）：

```bash
sudo sh -c 'echo ens2f0 > /sys/class/net/bond0/bonding/primary'
grep -E "Primary Slave|Currently Active Slave" /proc/net/bonding/bond0
```

**预期：**

```text
Primary Slave: ens2f0 (primary_reselect always)
Currently Active Slave: ens2f0
```

> 说明：若同时执行 `echo +ens2f0 > active_slave` 可能报 `No such device`（活动口已被 primary 自动切换），可忽略。

**备选（nmcli 一步完成）：**

```bash
sudo nmcli connection modify bond0 bond.options "mode=active-backup,miimon=100,primary=ens2f0,primary_reselect=2"
sudo nmcli connection up bond0
```

### 4.3 持久化（NetworkManager，必须执行）

#### 配置文件位置

本集群使用 **NetworkManager** 管理网络，bond 相关配置分布在以下文件（**每台节点路径相同**，IP 不同）：

| 文件 | 完整路径 | 作用 |
|------|----------|------|
| **bond 主配置** | `/etc/NetworkManager/system-connections/bond0.nmconnection` | bond 模式、primary、IP/网关/DNS |
| ens1f0 从属 | `/etc/NetworkManager/system-connections/bond-slave-ens1f0.nmconnection` | 将 ens1f0 加入 bond0 |
| ens2f0 从属 | `/etc/NetworkManager/system-connections/bond-slave-ens2f0.nmconnection` | 将 ens2f0 加入 bond0 |

**`[bond] mode=active-backup` 写在 bond 主配置文件中**，即：

```text
/etc/NetworkManager/system-connections/bond0.nmconnection
```

查看命令：

```bash
sudo cat /etc/NetworkManager/system-connections/bond0.nmconnection
# 或仅看 [bond] 段
sudo grep -A10 '^\[bond\]' /etc/NetworkManager/system-connections/bond0.nmconnection
```

> 说明：Rocky Linux 9 默认 **不使用** `/etc/sysconfig/network-scripts/ifcfg-bond0`；若该文件不存在，以 NetworkManager 路径为准。配置文件权限通常为 `600`，属主 `root`，需 `sudo` 查看/编辑。

#### 修复前配置（示例：doris105）

文件：`/etc/NetworkManager/system-connections/bond0.nmconnection`

```ini
[connection]
id=bond0
type=bond
interface-name=bond0

[bond]
mode=active-backup

[ipv4]
address1=10.10.243.105/24,10.10.243.254
method=manual
```

**问题：** `[bond]` 段仅有 `mode=active-backup`，**无 `primary=ens2f0`**，重启后默认仍走 ens1f0。

#### 修复后目标配置

持久化成功后，`[bond]` 段应包含：

```ini
[bond]
mode=active-backup
miimon=100
primary=ens2f0
primary_reselect=2
```

#### 持久化命令（每台节点）

```bash
sudo nmcli connection modify bond0 bond.options "mode=active-backup,miimon=100,primary=ens2f0,primary_reselect=2"
sudo nmcli connection up bond0
```

#### 验证

```bash
sudo grep -A10 '^\[bond\]' /etc/NetworkManager/system-connections/bond0.nmconnection
grep -E "Primary Slave|Currently Active Slave" /proc/net/bonding/bond0
```

### 4.4 修复结果

**doris104 ping doris105（修复后）：**

```text
15 packets transmitted, 15 received, 0% packet loss
rtt min/avg/max/mdev = 0.069/0.155/0.657/0.152 ms
```

| 对比项 | 修复前 | 修复后 |
|--------|--------|--------|
| ping max | 10~26 ms | **< 1 ms** |
| ping avg | 1.7~17 ms | **~0.15 ms** |
| 活动口 | ens1f0 | **ens2f0** |
| ens2f0 unrecoverable_errors | — | **0** |

**实施范围：** doris101 ~ doris105 均已切换至 ens2f0。

### 4.5 注意事项

- **勿主动切回 ens1f0**：故障根因在该口，切回后尖峰可能复现
- **ens1f0 历史计数不会清零**，但无流量后不应再增长
- **重启前务必完成 nmcli 持久化**，否则默认回到 ens1f0
- **中长期**：升级 bnx2x 固件/驱动；若 ens2f0 也出 assert，按硬件故障处理

---

## 5. 修复后确认命令

```bash
# 1. bond 活动口
grep -E "Primary Slave|Currently Active Slave" /proc/net/bonding/bond0

# 2. 活动口错误计数（取合计行）
sudo ethtool -S ens2f0 | awk '/^[[:space:]]+rx_discards:/ && !/\[/ {print}'; sudo ethtool -S ens2f0 | grep unrecoverable_errors

# 3. ping 延迟
ping -c 50 -i 0.2 <对端IP> | tail -3

# 4. 24h 内无新 assert
sudo journalctl -k --since "24 hours ago" --no-pager | grep -E "bnx2x_mc_assert|bnx2x_panic_dump|Tx-timeout"
```

---

## 6. 总结

| 项 | 内容 |
|----|------|
| **问题** | 节点间 ping 偶发 2~26 ms 尖峰 |
| **直接原因** | bond 活动口 ens1f0 固件 assert + Tx-timeout，RX 大量丢包，单包 RTT 被拉高 |
| **深层原因（推测）** | 主备 bond 长期 100% 流量走 ens1f0，经五个阶段逐渐恶化（见 **§2.4**）；ens2f0 作备口几乎无数据面负载 |
| **关键证据** | journalctl 中 `bnx2x_mc_assert`、`bnx2x_panic_dump`、`Tx-timeout` + `unrecoverable_errors` |
| **修复** | 101~105 bond 切换至 ens2f0，nmcli 持久化 `primary=ens2f0` |
| **结果** | ping max 由 10~26 ms 降至 **< 1 ms**，问题已修复 |
| **后续** | 按附录 B 监控 ens2f0；升级固件/驱动；规划 ens1f0 更换 |

---

## 附录 A：根因机制补充说明

> 五个演化阶段的完整判断见正文 **§2.4**。本节仅作补充。

### A.1 与 §2.4 的关系

| 正文 §2.4 | 本附录 |
|-----------|--------|
| 五阶段时间线、证据对应、五台同构分析 | ens2f0 风险与预防（见 **附录 B**） |

### A.2 阶段 2 与阶段 3 的关键区别

| | 阶段 2（ring 溢出） | 阶段 3（固件 assert） |
|---|---------------------|------------------------|
| **层次** | 驱动/ring 资源不足 | 网卡 microcode 内部崩溃 |
| **日志** | 通常无 assert | `bnx2x_mc_assert`、`panic_dump` |
| **计数** | rx_discards 持续增大 | unrecoverable_errors +1 |
| **可否 bypass** | 调 ring/IRQ 可能缓解 | 必须换口或升级 FW |

本集群 ens1f0 **两个阶段证据均有**：discards 上亿（阶段 2）+ assert 日志（阶段 3~4）。

---

## 附录 B：如何避免 ens2f0 重蹈覆辙

ens2f0 切换后将 **首次承担与 ens1f0 相同的 100% 流量**，需主动预防。

### B.1 必做（高优先级）

| 措施 | 目的 | 操作 |
|------|------|------|
| **持久化 primary=ens2f0** | 防重启回到 ens1f0 | 已完成 nmcli 持久化则跳过 |
| **监控 ens2f0 计数** | 在 ring 溢出阶段提前发现 | 每周执行下方巡检命令 |
| **监控 dmesg assert** | 在固件崩溃前/后立刻告警 | 对 `bnx2x_mc_assert`、`Tx-timeout` 做日志告警 |
| **升级 bnx2x 固件/驱动** | 消除已知 microcode bug | 查 HPE/Broadcom everest3 更新，assert 时 FW 为 7_13_21 |

**巡检命令（每台节点，建议 cron/监控平台）：**

```bash
# 合计 rx_discards + unrecoverable_errors（活动口）
DEV=$(grep "Currently Active Slave" /proc/net/bonding/bond0 | awk '{print $4}')
echo "$(hostname) active=$DEV discards=$(ethtool -S $DEV | awk '/^[[:space:]]+rx_discards:/ && !/\[/ {print $2}') unrec=$(ethtool -S $DEV | grep unrecoverable_errors | awk '{print $2}')"
journalctl -k --since "24 hours ago" --no-pager | grep -cE "bnx2x_mc_assert|Tx-timeout" || true
```

**告警阈值建议：**

- `unrecoverable_errors` **任何增长** → 立即告警  
- 活动口 `rx_discards` **24h 增长 > 100万** → 告警（需按业务流量校准基线）  
- 24h 内出现 **assert / Tx-timeout** → 立即告警  

### B.2 建议做（降低触发固件 bug 的概率）

| 措施 | 原理 | 操作参考 |
|------|------|----------|
| **适当增大 RX ring** | 减少 burst 时 ring 溢出 → rx_discards | `ethtool -g ens2f0` 查看；若 max > current，可 `ethtool -G ens2f0 rx <值>`（需业务低峰测试） |
| **IRQ 绑核 / RPS 调优** | 减少 softirq 堆积导致处理延迟 | 查看 `cat /proc/interrupts \| grep ens2f0`；结合 `irqbalance` 或手动绑 NUMA 本地 CPU |
| **避免 ens1f0 再承担流量** | 故障口不应再被压测 | 保持 primary=ens2f0，勿手动切回 |
| **固件升级后压测** | 验证修复再决定是否复用 ens1f0 | 升级后对 ens1f0 做 isolated 压测（见 B.4） |

### B.3 可选做（架构层，中长期）

| 措施 | 说明 |
|------|------|
| **802.3ad LACP 双活** | 双口分担流量，避免单口 100% 负载（需交换机支持） |
| **业务网与容器网分离** | 节点存在 flannel/docker 网卡，增加 PPS；Doris 专用网卡尽量减少非业务流量 |
| **更换/报废 ens1f0** | 已有 unrecoverable_errors 的口不建议再作生产主口 |

### B.4 ens2f0 会不会也坏？如何判断

| 切换后表现 | 含义 | 下一步 |
|------------|------|--------|
| 3~7 天：discards 低且稳定，无 assert | **大概率是 ens1f0 个体/路径问题** | 继续监控，规划 ens1f0 更换 |
| 1~4 周：discards 缓慢持续上升，无 assert | **进入阶段 2**，ring/驱动参数需调优 | 增大 ring、查 IRQ、考虑 FW 升级 |
| 出现 assert / Tx-timeout | **固件级共性问题** | 立即 failover 预案；全员升级 FW/驱动；联系硬件厂商 |
| ens2f0 也坏但换槽/换卡后好 | **PCIe 槽位或批次问题** | 记录 01:00.0 vs 22:00.0 差异，统一硬件方案 |

### B.5 不建议的做法

- **清 zero 错误计数** → 不解决根因，掩盖阶段 2~3 征兆  
- **未升级固件即切回 ens1f0** → 高概率复现 assert  
- **无监控长期运行** → 可能在阶段 3 才被动发现  

---

*报告整理日期：2026-06-03*

