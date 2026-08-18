# nginx_server_conf

```
404页面，conf示例
cat igogame.conf
upstream igogame {
   server 172.21.240.104:28126;
   server 172.21.240.105:28126;
   server 172.21.240.106:28126;
   keepalive 1024;
}

server {
    listen 18126 default;
    listen [::]:18126 default;
    server_name localhost
    keepalive_timeout 80;

    access_log    /data/logs/nginx/igogame/access.log vhostu;
    error_log     /data/logs/nginx/igogame/error.log;

    location / {
        proxy_pass http://igogame;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Connection "";
        proxy_http_version 1.1;
    }
error_page  404 403 500 502 503 504  /404.html;
location = /404.html {
# 放错误页面的目录路径。
root   /usr/share/nginx/html;
}
}
```

```
$ cat accss-https.conf
upstream access-proxy-https {
  server  192.168.240.104:28106;
  server  192.168.240.105:28106;
  server  192.168.240.106:28106;
  hash $arg_token consistent;
}
upstream access-service-https {
  server  192.168.240.104:28107;
  server  192.168.240.105:28107;
  server  192.168.240.106:28107;
}

map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

#log_format main '$remote_addr - $remote_user [$time_local] "$request" '
#                      '$status $body_bytes_sent "$http_referer" '
#                      '"$http_user_agent" "$http_x_forwarded_for"' '"$upstream_addr"' '"$upstream_response_time"';

server {
    listen 18111 ssl default;
    listen [::]:18111 ssl default;
    server_name access.igo.com;
    keepalive_timeout 80;
    ssl_certificate /etc/nginx/ssl/ssl.crt;
    ssl_certificate_key /etc/nginx/ssl/ssl.key;
    ssl_session_timeout 5m;
    ssl_protocols SSLv3 TLSv1 TLSv1.1 TLSv1.2;
    ssl_ciphers "HIGH:!aNULL:!MD5 or HIGH:!aNULL:!MD5:!3DES";
    ssl_prefer_server_ciphers on;

    access_log    /data/logs/nginx/access.igo.com/access-https.log log_json;
    error_log    /data/logs/nginx/access.igo.com/error-https.log;

    location /v1/endpoint {
        proxy_pass http://access-proxy-https;

        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        proxy_http_version 1.1;
        proxy_connect_timeout 4s;
        proxy_read_timeout 60s;
        proxy_send_timeout 12s;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;

        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods 'GET, PUT, POST, OPTIONS';
        add_header Access-Control-Allow-Headers 'DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization';
    }

    location /v1/cloudgame {
        proxy_pass http://access-service-https;

        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        proxy_http_version 1.1;

        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods 'GET, PUT, POST, OPTIONS';
        add_header Access-Control-Allow-Headers 'DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization';
    }

    location / {
        root /export/servers/nginx/html;
        index index.htm index.html;
        try_files $uri $uri/ =404;
    }
}
```
