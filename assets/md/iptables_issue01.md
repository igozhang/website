# iptables问题导致业务丢包

```
env
redhat6.5
ngx16
```

```
现象：
业务反馈丢包，访问超时

观察：
机器负载低，CPU、内存等其他指标正常;
查看日志，/var/log/messages里看到nf_conntrack: table full, dropping packet；

查看nf_conntrack_max，最大值是65536
# sysctl -a |grep conntrack_max
查看当前连接情况，已经超过最大值
# grep conn /proc/slabinfo

解决：
service iptables stop
chkconfig iptables off
关闭iptables之后，访问不再经过netfilter，就会不再触发conntrack满引起丢包

原理：
nf_conntrack 跟踪所有网络连接，记录存储在 1 个哈希表里，并会保留一段时间，根据协议和状态有所不同，直到超时都没有收发包就会清除记录。如果服务器比较繁忙，新连接进来的速度远高于释放的速度，把哈希表塞满了，新连接的数据包就会被丢掉。
机器启用iptables的时候，会调用nf_conntrack相关模块，如果没指定NOTRACK，所有进来的访问都会通过netfilter跟踪连接状态，当访问请求量大的时候，出问题的机器会话数超过nf_conntrack_max，所以丢包。
```
