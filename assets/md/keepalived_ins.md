# keepalived_ins

在6台ubuntu20.04机器上通过keepalive维护一个虚拟IP实现ingress服务高可用：  
IP: 10.201.81.54-59  
VIP: 10.201.81.6  
apt -y install keepalived

## 主节点 54 配置

```
vim /etc/keepalived/keepalived.conf

! 精简版 Keepalived 配置（仅维护 VIP）
global_defs {
    router_id INGRESS_MASTER  # 节点标识，自定义即可，不参与算法，可随意
}

vrrp_instance VI_1 {
    state MASTER               # 主节点标识
    interface ens18             # 替换为你的实际网卡名（如 ens33）
    virtual_router_id 6       # 所有节点必须一致（1-255）
    priority 100               # 主节点优先级最高
    advert_int 1               # 心跳间隔 1 秒
    nopreempt                  # 非抢占模式（避免主节点恢复后抢占 VIP）

    # 认证配置（所有节点必须一致）
    authentication {
        auth_type PASS
        auth_pass 1111         # 自定义认证密码，建议修改
    }

    # 虚拟 IP 配置（VIP）
    virtual_ipaddress {
        10.201.81.6/24 dev ens18  # VIP + 子网掩码 + 网卡
    }
}

systemctl enable keepalived
systemctl start keepalived
```

## 其他节点配置

```
global_defs {
    router_id INGRESS_BACKUP  # 备节点标识，自定义
}

vrrp_instance VI_1 {
    state BACKUP              # 备节点标识
    interface ens18            # 与主节点一致的网卡
    virtual_router_id 6      # 与主节点一致
    priority PRIORITY_PLACEHOLDER               # 备节点优先级（主节点100最高，其余节点随意降低）
    advert_int 1
    nopreempt

    authentication {
        auth_type PASS
        auth_pass 1111        # 与主节点一致
    }

    virtual_ipaddress {
        10.201.81.6/24 dev ens18  # 与主节点一致
    }
}
```
