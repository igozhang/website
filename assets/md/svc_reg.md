# 服务注册,linux-service_registration

## systemctl

```
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

## supervisor

```
```shell
安装
apt-get install supervisor 
yum install -y supervisor 
pip2 install supervisor

pgrep supervisord

/etc/supervisor/supervisord.conf 配置文件

filebeat
tee >/etc/supervisord.d/filebeat.ini <<EOF
[program:filebeat]
command=/opt/filebeat/filebeat -e -c /opt/filebeat/filebeat.yml
autorestart=true
autostart=true
stderr_logfile=/var/log/supervisor/filebeat_err.log
stdout_logfile=/var/log/supervisor/filebeat_out.log
user=root
stopsignal=INT
startsecs=10
EOF

supervisorctl status
supervisorctl start filebeat
supervisorctl restart filebeat
supervisorctl stop all
supervisorctl reload
supervisorctl update

```
```
