# nginx_prometheus

```
nginx/1.10.3
CentOS Linux release 7.6.1810 (Core)
```

## 方案对比

```
方式一(官方):侵入较方式二高且配置复杂
Nginx metric library
依赖libnginx-mod-http-lua模块;
一段Lua脚本（prometheus.lua）;
https://github.com/knyar/nginx-lua-prometheus

方式二(官方):已停止维护
Nginx VTS exporter
https://github.com/hnlq715/nginx-vts-exporter
监控vts状态并通过http暴露给prometheus

方式三(非官方):
nginx_exporter
依赖ngx_http_stub_status_module 模块;
抓取统计页面 /nginx_status获取一些比较简单的指标;
https://github.com/nginxinc/nginx-prometheus-exporter

方式四(官方推荐)
nginx-module-vts 用于替换方式二
Nginx virtual host traffic status module
supports multiple vhost_traffic_status_display_format, like <json|html|jsonp|prometheus>
也叫nginx-vtx-module
https://github.com/vozlt/nginx-module-vts
```

## 方案四操作步骤

```
1. 下载对应版本nginx; https://nginx.org/
2. 重编译nginx;
3. 按需安装依赖;
4. 备份替换nginx;
5. 重启;

1.
wget https://nginx.org/download/nginx-1.10.3.tar.gz

2.
2.1 查看现有模块
sbin/nginx -V
nginx version: nginx/1.10.3
built by gcc 4.8.5 20150623 (Red Hat 4.8.5-28) (GCC) 
built with OpenSSL 1.0.2k-fips  26 Jan 2017
TLS SNI support enabled
configure arguments: --prefix=/opt/nginx --with-http_stub_status_module --with-http_ssl_module --with-http_stub_status_module --add-module=//upload/nginx_upstream_check_module-master/ --with-stream

2.3 下载vtx模块
cd 
shell> git clone git://github.com/vozlt/nginx-module-vts.git

2.3 添加vtx重编译,--add-module=/path/to/nginx-module-vts
./configure --prefix=/opt/nginx --with-http_stub_status_module --with-http_ssl_module --with-http_stub_status_module --add-module=/data/nginx_upstream_check_module/ --with-stream --add-module=/data/nginx-module-vts

make

3.
按需安装依赖
yum -y install zlib zlib-devel pcre pcre-devel openssl openssl-devel

4.
备份替换nginx
cp /data/nginx-1.10.3/objs/nginx /opt/nginx/nginx/sbin/nginx.new

5.
重启nginx
./nginx stop
./nginx.new -c /opt/nginx/conf/nginx.conf

验证:
http://172.168.14.166:8088/status/format/prometheus
支持目录:
/status/format/json
/status/format/html
/status/format/jsonp
/status/format/prometheus
/status/control
```

```
Synopsis:

http {
    vhost_traffic_status_zone;

    ...

    server {

        ...

        location /status {
            vhost_traffic_status_display;
            vhost_traffic_status_display_format html;
        }
    }
}
```
