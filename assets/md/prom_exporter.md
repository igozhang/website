# Prometheus_exporter

```
env
CentOS Linux release 7.9.2009 (Core)
prometheus-2.33.4.linux-amd64
```

```
zookeeper_exporter
redis_exporter
rabbitmq_prometheus
nginx_prometheus
```

```
zookeeper_exporter v1.1.0
zookeeper-3.4.9 cent7.6

https://github.com/carlpett/zookeeper_exporter/releases
mkdir -p /data/prom/
wget http://github.com/carlpett/zookeeper_exporter/releases/download/v1.1.0/zookeeper_exporter -O /data/prom/zookeeper_exporter
chmod +x /data/prom/zookeeper_exporter
./zookeeper_exporter -h

注册服务:
tee >/usr/lib/systemd/system/zk-exporter.service <<EOF
[Unit]
Description= prometheus-zookeeper
After=network.target

[Service]
Type=simple
ExecStart= /data/prom/zookeeper_exporter -zookeeper 127.0.0.1:2181 -bind-addr :9143
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl start zk-exporter
systemctl enable zk-exporter

验证:
[root@igo-qq-h1 data]# curl localhost:9143/metrics
```

```
redis_exporter v1.35.1
Redis server v=5.0.5  cent7.6

Go编译安装:
https://github.com/oliver006/redis_exporter
git clone https://github.com/oliver006/redis_exporter.git
cd redis_exporter
go build .

二进制包安装:
mkdir -p /data/prom/
cd /data/prom/
wget https://github.com/oliver006/redis_exporter/releases/download/v1.35.1/redis_exporter-v1.35.1.linux-amd64.tar.gz
tar -xf redis_exporter-v1.35.1.linux-amd64.tar.gz
mv redis_exporter-v1.35.1.linux-amd64 redis_exporter
cd redis_exporter
./redis_exporter -h
./redis_exporter --version
./redis_exporter -redis.addr redis://localhost:6379
curl localhost:9121/metrics

服务注册
tee >/usr/lib/systemd/system/redis_exporter.service <<EOF
[Unit]
Description= redis_exporter
After=network.target

[Service]
Type=simple
ExecStart= /data/prom/redis_exporter/redis_exporter -redis.addr redis://localhost:6379 -web.listen-address :9121  -redis-only-metrics  -redis.password igopwd
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl start redis_exporter
systemctl enable redis_exporter
systemctl status redis_exporter
```

```
rabbitmq_prometheus 3.8.6

rabbitmq_server-3.8.6
CentOS Linux release 7.6.1810 (Core)

第一种方法:
3.8.0版本开始内置了rabbitmq_prometheus模块;
3.8.0之前版本需要单独安装prometheus_rabbitmq_exporter
https://github.com/deadtrickster/prometheus_rabbitmq_exporter

第二种选择:(单独部署exporter)
https://github.com/kbudde/rabbitmq_exporter
./rabbitmq_exporter -config-file config.example.json
http://127.0.0.1:15672

rabbitmq_prometheus 3.8.6

第一种方法的实现:

# ./sbin/rabbitmq-plugins enable rabbitmq_prometheus
# curl  "http://172.168.14.165:15692/metrics"
# curl -v -H "Accept:text/plain" "http://172.168.14.165:15692/metrics"
```
