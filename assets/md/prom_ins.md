# Prometheus安装，Prometheus监控方案简介

env  
CentOS Linux release 7.9.2009 (Core)  
prometheus-2.33.4.linux-amd64  
node\_exporter-1.3.1.linux-amd64  
grafana-8.4.1-1.x86\_64  
alertmanager-0.23.0.linux-amd64

```
Prometheus是go编写的开源监控系统，使用时序数据库（TSDB）,不依赖任何三方软件包
2012年开始由前Google工程师在Soundcloud以开源软件的形式进行研发，并且于2015年早期对外发布早期版本
https://prometheus.io/download

Prometheus安装部署
node_exporter安装部署
Grafana安装部署
Alertmanager安装部署
```

**数据流**

服务架构

## **Prometheus**安装部署

```
下载安装：
mkdir -p /opt /data/prom
cd /opt
export VERSION=2.33.4
curl -OL https://github.com/prometheus/prometheus/releases/download/v$VERSION/prometheus-$VERSION.linux-amd64.tar.gz
tar -xf prometheus-$VERSION.linux-amd64.tar.gz
mv prometheus-$VERSION.linux-amd64 prom
cd prom
touch $VERSION.readme
./prometheus --storage.tsdb.path="/data/prom"

注册systemd服务：
useradd prom
echo 'igopwd' |passwd --stdin prom
chown -R prom.prom /data/prom /opt/prom

tee >/etc/systemd/system/prometheus.service <<EOF
[Unit]
Description=Prometheus Server
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target
[Service]
User=prom
Restart=on-failure
ExecStart=/opt/prom/prometheus \
--config.file=/opt/prom/prometheus.yml \
--storage.tsdb.path=/data/prom
[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl start prometheus
systemctl enable prometheus
```

http://192.168.3.191:9090/

## node\_exporter安装部署

```
mkdir /data /opt
cd /opt
wget https://github.com/prometheus/node_exporter/releases/download/v1.3.1/node_exporter-1.3.1.linux-amd64.tar.gz
touch node_exporter-1.3.1.linux-amd64/node_exporter-1.3.1.linux-amd64.readme
mv node_exporter-1.3.1.linux-amd64 node_exporter
cd node_exporter

配置systemd服务：
tee >/usr/lib/systemd/system/node_exporter.service <<EOF
[Unit]
Description=node_exporter service
[Service]
User=root
ExecStart=/opt/node_exporter/node_exporter
TimeoutStopSec=10
Restart=on-failure
RestartSec=5
[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl start node_exporter
systemctl status node_exporter
systemctl enable node_exporter

配置prometheus_srv端：
[root@igo-prometheus-srv opt]#
cp /opt/prom/prometheus.yml{,.bak}
vim 追加配置 /opt/prom/prometheus.yml

  - job_name: "igo-111_node"
    static_configs:
      - targets: ["192.168.3.111:9100"]
        labels:
          label: "igo-golang-111"
```

node\_exporter

node\_exporter\_metrics

## Grafana安装部署

```
tee >/etc/yum.repos.d/grafana.repo <<EOF
[grafana]
name=grafana
baseurl=https://mirrors.aliyun.com/grafana/yum/rpm
repo_gpgcheck=0
enabled=1
gpgcheck=0
EOF

yum makecache
yum list grafana --showduplicates
yum install -y grafana-8.4.1-1
systemctl start grafana-server
systemctl enable grafana-server
```

grafana

igo-node\_exporter111,id号8919

## Alertmanager安装部署

```
容器安装
$ docker run --name alertmanager -d -p 127.0.0.1:9093:9093 quay.io/prometheus/alertmanager
http://localhost:9093/

Go安装
$ GO15VENDOREXPERIMENT=1 go get github.com/prometheus/alertmanager/cmd/...
# cd $GOPATH/src/github.com/prometheus/alertmanager
$ alertmanager --config.file=<your_file>
http://localhost:9093/

包安装Precompiled binaries
wget https://github.com/prometheus/alertmanager/releases/download/v0.23.0/alertmanager-0.23.0.linux-amd64.tar.gz
touch alertmanager-0.23.0.linux-amd64/alertmanager-0.23.0.linux-amd64.readme
mv alertmanager-0.23.0.linux-amd64 alertmanager
mkdir -p /data/alertmanager

注册服务systemd
tee >/usr/lib/systemd/system/alertmanager.service <<EOF
	[Unit]
	Description=alertmanager
	Documentation=https://github.com/prometheus/alertmanager
	After=network.target
	
	[Service]
	Type=simple
	User=root
	ExecStart=/opt/alertmanager/alertmanager --config.file=/opt/alertmanager/alertmanager.yml --storage.path=/data/alertmanager
	Restart=on-failure
	
	[Install]
	WantedBy=multi-user.target
EOF

systemctl start alertmanager
systemctl enable alertmanager

修改prometheus字段
vim /opt/prom/prometheus.yml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - 192.168.3.191:9093
		  
systemctl restart prometheus
```

http://192.168.3.191:9093/

```
Alertmanager_cfg-example
[root@igo-prometheus-srv alertmanager]# cat /opt/alertmanager/alertmanager.yml
route:
  group_by: ['alertname']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 1h
  receiver: 'web.hook'
receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://127.0.0.1:5001/'
inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
```
