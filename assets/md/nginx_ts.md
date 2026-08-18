# NGINX问题排查

```
nginx_ts
```

```
env： centos7.6  nginx1.16.1
err01: 部分用户访问500报错
2023/08/28 08:50:38 [alert] 488073#488073: *1591754 socket() failed (24: Too many open files) while connecting to upstream, client: 10.98.88.174, server: igozhang.cn, request: "GET /api/ui/lookup/value/ss", upstream: "http://10.10.10.10:8080/igozhang/api/ui/lookup/value/ss", host: "igozhang.cn", referrer: "https://igozhang.cn/beep/order/orderConfig/workCenter"
2023/08/28 08:50:41 [crit] 488070#488070: accept4() failed (24: Too many open files)
分析:
1. 相关时间段系统日志
journalctl --since "2023-08-28 08:49" --until "2023-08-28 08:52"
未发现异常报错,判断非系统文件描述符超限问题

2. 查看相关文件限制
ulimit -n
/proc/sys/fs/file-max
/proc/sys/fs/nr_open
均为20480000
再查/proc/$pid/limits,发现Max open files 为1024,判断瓶颈在这里

解决:
1. 热修改
centos6在线修改  echo -n "Max open files=65536:65536" > /proc/$pid/limits
centos7在线修改  prlimit --nofile=65536:65536 -p $pid
值不能大于65536,否则可能导致系统重启失败
2. 永久修改
如果是systemctl启动的服务:
/usr/lib/systemd/system/nginx.service的[Service]下加上LimitNOFILE=65536这一行;
systemctl dameon-reload然后systemctl restart nginx生效
本例中非使用systemctl启动nginx
修改nginx配置项/etc/nginx/nginx.conf加上:worker_rlimit_nofile 65535;
3. nginx的limit重启的时候未指定的话会默认取父进程1systemd的值,比如又是1024:
sed -i 's/^.*DefaultLimitNOFILE.*$/DefaultLimitNOFILE=65536/g' /etc/systemd/system.conf
systemctl daemon-reexec
或者
prlimit --nofile=65536:65536 -p 1

验证:
1. /proc/$pid/limits里Max open files 已经为65536
2. nginx某worker进程突破1024后err_log未复现too many 报错
```

```
env： centos7.6  nginx1.16.1
err02:
/var/log/nginx/error.log
2023/08/28 14:58:19 [alert] 189848#189848: 1024 worker_connections are not enough

solution:
/etc/nginx/nginx.conf  worker_connections  10240;
nginx -s reload
```

```
err03:
nginx -s stop
nginx: [error] invalid PID number "" in "/var/run/nginx.pid"

solution:
echo 189840 >/var/run/nginx.pid
```
