# nginx

```
编译动态模块
wget https://nginx.org/download/nginx-1.20.2.tar.gz
tar -xf nginx-1.20.2.tar.gz
cd nginx-1.20.2
./configure --help |more ##dynamic表示模块支持动态
  --with-http_perl_module            enable ngx_http_perl_module
  --with-http_perl_module=dynamic    enable dynamic ngx_http_perl_module
  --with-http_image_filter_module    enable ngx_http_image_filter_module
  --with-http_image_filter_module=dynamic

                                     enable dynamic ngx_http_image_filter_module

编译安装
./configure --prefix=/opt/nginx --with-http_image_filter_module=dynamic
make

验证
# ls /opt/nginx/modules
ngx_http_image_filter_module.so

vim /opt/nginx/conf/nginx.conf
load_module modules/ngx_http_image_filter_module.so; #首行
        location / {
            root   test;
            image_filter resize 150 150;
            index  index.html index.htm;
        }
nginx -s reload 
chrome_F12,network,disable cache禁用缓存再刷新
```
