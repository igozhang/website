# redis6_ins

redis6集群安装部署  
ubuntu20.04  
redis-6.2.14  
https://download.redis.io/releases/redis-6.2.14.tar.gz

## 编译安装

```
apt install -y gcc make tcl

tar -zxvf redis-6.2.14.tar.gz
cd redis-6.2.14
make
make install
mkdir -p /etc/redis/cluster/7000
mkdir -p /var/lib/redis/cluster/7000
mkdir -p /data/redis/7000
chown -R userroot:userroot /etc/redis/cluster
chown -R userroot:userroot /var/lib/redis/cluster
chown -R userroot:userroot /data/redis/7000

防火墙
sudo ufw allow 7000/tcp
sudo ufw allow 17000/tcp
sudo ufw reload
```

## 配置

```
tee >/etc/redis/cluster/7000/redis.conf <<EOF
bind 0.0.0.0
port 7000
daemonize yes
pidfile /var/run/redis_7000.pid

# 日志文件
logfile /data/redis/7000/redis.log

# 数据文件目录
dir /data/redis/7000

# 关闭 RDB 快照（完全关闭 dump）
save ""
rdbcompression no
rdbchecksum no

# 开启 AOF 持久化
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite yes
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb

# 集群配置
cluster-enabled yes
cluster-config-file nodes_7000.conf
cluster-node-timeout 5000

# 密码（所有节点必须一样）
requirepass Redis@2026
masterauth Redis@2026

# 安全优化
protected-mode no
timeout 0
tcp-keepalive 300
maxmemory-policy volatile-lru
EOF
```

## 启动节点

```
redis-server /etc/redis/cluster/7000/redis.conf

做成守护服务
cat > /etc/systemd/system/redis.service <<EOF
[Unit]
Description=Redis In-Memory Data Store
After=network.target

[Service]
Type=forking #集群daemonize yes这里就要配成forking,否则主进程退出systemctl误认为挂了就导致无限重启
User=userroot
Group=userroot
ExecStart=/usr/local/bin/redis-server /etc/redis/cluster/7000/redis.conf
ExecStop=/usr/local/bin/redis-cli -a 'Redis@2026' -p 7000 shutdown
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable redis
systemctl start redis
systemctl status redis
```

## 建集群

```
redis-cli -a Redis@2026 --cluster create \
10.80.238.71:7000 \
10.80.238.72:7000 \
10.80.238.73:7000 \
10.80.238.74:7000 \
10.80.238.75:7000 \
10.80.238.76:7000 \
--cluster-replicas 1

redis-cli -a Redis@2026 -p 7000
redis-cli -c -h 10.80.238.71 -p 7000 -a Redis@2026
CLUSTER INFO
CLUSTER NODES
set foo bar
get foo
```
