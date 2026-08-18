# k8s_wp

在wp服务器上部署k8s，跑容器，用来共享igozhang.cn域名并提供服务  
示例： igozhang.cn/trans  
核心步骤：

1. 跑一个trans服务，ingress提供trans.igozhang.cn入口
2. wp的nginx添加trans.conf配置文件提供反代到这个域名  
   坑： location ^~ /trans/ 不能 location /trans/，这样才可以吧css等样式文件带上

## 部署trans服务

部署过程不赘述，可以git我的gitee  
部署后参考：

```
[root@VM-4-11-opencloudos tmp]# kg pod -n igo
NAME                     READY   STATUS    RESTARTS   AGE
trans-85d8b878c9-8m9fr   1/1     Running   0          147m
[root@VM-4-11-opencloudos tmp]# kg svc -n igo
NAME    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
trans   ClusterIP   10.98.208.21   <none>        80/TCP    5h2m
[root@VM-4-11-opencloudos tmp]# kg ingress -n igo
NAME    CLASS   HOSTS               ADDRESS         PORTS   AGE
trans   nginx   trans.igozhang.cn   10.106.144.22   80      5h2m
```

## 添加WP-NGINX配置

```
1. 添加一条server级别的include，注意不是http级别的include；

vim /www/server/panel/vhost/nginx/wordpress.local.conf
添加一条：
include /www/server/panel/vhost/nginx/k8s-includes/*.conf;

2. 添加反代配置文件
mkdir -p /www/server/panel/vhost/nginx/k8s-includes
tee /www/server/panel/vhost/nginx/k8s-includes/trans.conf <<'EOF'
# trans：igozhang.cn/trans → Ingress(trans.igozhang.cn)
location = /trans {
    return 302 /trans/;
}
location ^~ /trans/ {
    proxy_pass http://127.0.0.1:30080/;
    proxy_set_header Host trans.igozhang.cn;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_connect_timeout 7200s;
    proxy_read_timeout 7200s;
    proxy_send_timeout 7200s;
    client_max_body_size 0;
}
EOF

nginx -t && nginx -s reload
```
