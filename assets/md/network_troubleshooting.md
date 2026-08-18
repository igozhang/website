# Network_TroubleShooting

```
1.
kvm虚拟机配置好网络后，外网无法访问

原因
docker默认网段和虚拟机网段冲突
解决：
cat <<EOF >> /etc/docker/daemon.json
{
    "registry-mirrors": ["https://docker.mirrors.ustc.edu.cn"],
        "bip":"192.168.0.2/24",
    "graph": "/data/docker"
}
EOF

systemctl restart docker
```
