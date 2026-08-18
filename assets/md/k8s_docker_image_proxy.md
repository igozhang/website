# k8s_docker_image_proxy内网代理

```
解决github等各种被墙问题

方案1.
搭一个内网机场clash代理，监听比如7890端口

mkdir -p /etc/systemd/system/docker.service.d/

tee >/etc/systemd/system/docker.service.d/proxy.conf <<EOF
[Service]
Environment="HTTP_PROXY=http://172.17.34.10:7890/"
Environment="HTTPS_PROXY=http://172.17.34.10:7890/"
Environment="NO_PROXY=127.0.0.0/8,172.17.0.0/16"
EOF

systemctl daemon-reload
systemctl restart docker

done
```
