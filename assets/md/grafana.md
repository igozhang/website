# grafana

```
CentOS Linux release 7.9.2009 (Core)
Docker version 1.13.1, build 7d71120/1.13.1
grafana/grafana:6.7.4
```

```
yum -y install docker
cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": ["https://59xo2v7a.mirror.aliyuncs.com"]
}
EOF
systemctl start docker

docker pull grafana/grafana:6.7.4

docker run --name=igo-grafana -d \
-p 3000:3000 \
grafana/grafana:6.7.4

http://192.168.3.222:3000/
初始默认密码
admin/admin
```
