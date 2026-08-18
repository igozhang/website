# k8s服务暴露

```
env
CentOS Linux release 7.9.2009
k8s v1.19.16  3+3
```

```
1.
nodeport方式
kubectl create deployment nginx-01 --image=nginx -n igo
kubectl expose deployment nginx-01 --port=80  --type=NodePort -n igo
[root@igo-k8s-1 ingress]# kg svc -n igo
NAME       TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
nginx-01   NodePort   10.103.39.89   <none>        80:30563/TCP   109m
访问
http://192.168.3.201:30563/
```

```
2.
ingress,contour方式
apply -f ingress.yaml -n igo
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: igo.nginx-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$1
spec:
  rules:
    - host: igo.nginx-ingress.web
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: nginx-01
                port:
                  number: 80
添加hosts解析
192.168.3.201 igo.nginx-ingress.web
查看服务端口
[root@igo-k8s-1 ingress]# kg -n heptio-contour svc
NAME      TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
contour   LoadBalancer   10.104.142.96   <pending>     80:30033/TCP,443:30433/TCP   3h24m
```

```
3.
nginx代理
nginx绑定宿主机网络，代理后台svc

nginx绑定宿主机网络
kubectl label nodes node3 role=nginx
# cat area-nginx.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: area-nginx
  name: area-nginx
  namespace: area01
spec:
  replicas: 1
  selector:
    matchLabels:
      app: area-nginx
  template:
    metadata:
      labels:
        app: area-nginx
    spec:
      nodeSelector:
        role: nginx
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
        - name: nginx
          image: nginx:latest
          imagePullPolicy: IfNotPresent
          volumeMounts:
          - name: conf
            mountPath: /etc/nginx/conf.d
          - name: nginx-conf
            mountPath: /etc/nginx/nginx.conf
          - name: logdir
            mountPath: /var/log/nginx
          - name: certs
            mountPath: /etc/nginx/ssl
          - mountPath: /etc/localtime
            name: timezone
            subPath: Shanghai
      volumes:
      - name: logdir
        hostPath:
          path: /data/area01/logs/nginx
          type: DirectoryOrCreate
      - name: conf
        hostPath:
          path: /export/area01/area-nginx/sites-enabled
          type: DirectoryOrCreate
      - name: certs
        hostPath:
          path: /export/area01/area-nginx/ssl
          type: DirectoryOrCreate
      - name: nginx-conf
        hostPath:
          path: /export/area01/area-nginx/nginx.conf
          type: FileOrCreate
      - name: timezone
        configMap:
          name: area-timezone

代理后台svc
# cat union-access-https.conf
upstream union-access {
   server union-access-svc:30108 fail_timeout=60s;
   keepalive 1024;
}

server {
    listen 18112 ssl;
    listen [::]:18112 ssl;
    server_name beta-02-union-access.yuntiancloud.com;
    keepalive_timeout 80;

    access_log    /var/log/nginx/union-access-https-access.log vhostu;
    error_log    /var/log/nginx/union-access-https-error.log;

    proxy_next_upstream error timeout http_500;
    proxy_next_upstream_tries 2;
    proxy_next_upstream_timeout 2800ms;
    proxy_connect_timeout 100ms;
    proxy_read_timeout 2500ms;
    proxy_send_timeout 200ms;

    ssl_certificate /etc/nginx/ssl/yuntiancloud.com.crt;
    ssl_certificate_key /etc/nginx/ssl/yuntiancloud.com.key;
    ssl_session_timeout 5m;
    ssl_protocols SSLv3 TLSv1 TLSv1.1 TLSv1.2;
    ssl_ciphers "HIGH:!aNULL:!MD5 or HIGH:!aNULL:!MD5:!3DES";
    ssl_prefer_server_ciphers on;

    if ($uri !~ ^/(v1|v2)) {
        return 403;
        break;
    }

    location / {
        proxy_pass http://union-access;

        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Connection "";
        proxy_http_version 1.1;

        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods 'GET, PUT, POST, OPTIONS';
        add_header Access-Control-Allow-Headers 'DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization';
    }
}

验证：
# telnet 172.17.35.13 18112
Trying 172.17.35.13...
Connected to 172.17.35.13.
Escape character is '^]'.
```

```
4.
服务直接绑定宿主机网络
同3，不赘述
# cat union-elasticsearch.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: union
  labels:
    product: k8s-elastic
  name: elastic-config
data:
  elasticsearch.yml: |
    cluster.name: "docker-cluster"
    network.host: "0.0.0.0"

---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  namespace: union
  name: union-elasticsearch
spec:
  selector:
    matchLabels:
      run: elasticsearch
  serviceName: union-elasticsearch-coordinating-only
  replicas: 1
  template:
    metadata:
      labels:
        name: elasticsearch
        run: elasticsearch
      name: union-elasticsearch
    spec:
#      nodeSelector:
#        role: node2
#      hostNetwork: true
#      dnsPolicy: ClusterFirstWithHostNet
```
