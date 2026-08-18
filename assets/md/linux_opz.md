# linux_optimize

```
1. #增大文件描述符
ulimit -n 65536
echo -ne "
* soft nofile 65536
* hard nofile 65536
" >>/etc/security/limits.conf

2. #修改系统线程限制
echo -ne "
* soft nproc 2048
* hard nproc 4096
" >>/etc/security/limits.conf

3. #针对ngx的dmesg“nf_conntrack: table full, dropping packet”
服务器随机丢弃请求问题
vim /etc/sysctl.conf
net.nf_conntrack_max = 655360 （状态跟踪表的最大行数，16G的服务器）
net.netfilter.nf_conntrack_tcp_timeout_established = 1200（设置超时时间）

4. #交换分区优先级
/etc/sysctl.conf
内存在使用到100-10=90%的时候开始使用swap
vm.swappiness = 10

5. #关闭swap，对es，mysql有效
```
