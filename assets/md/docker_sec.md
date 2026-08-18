# 容器安全,docker_sec

```
对于任意一个进程，在做任意一个特权操作的时候，都需要对应的 capability,比如iptables命令，需要CAP_NET_ADMIN;
如果要 mount文件系统，那么对应的进程需要有 CAP_SYS_ADMIN 这个 capability;
另外，CAP_SYS_ADMIN 这个 capability 里允许了大量的特权操作，包括文件系统，交换空间，还有对各种设备的操作，以及系统调试相关的调用等等;
在普通 Linux 节点上，非 root 用户启动的进程缺省没有任何 Linux capabilities，而 root 用户启动的进程缺省包含了所有的 Linux capabilities
Linux capabilities
https://man7.org/linux/man-pages/man7/capabilities.7.html
内核文件
https://github.com/torvalds/linux/blob/v5.4/include/uapi/linux/capability.h#L113
Capabilities: Why They Exist and How They Work
https://blog.container-solutions.com/linux-capabilities-why-they-exist-and-how-they-work
Linux Capabilities in Practice
https://blog.container-solutions.com/linux-capabilities-in-practice
```

```
命令
摘除root用户的CAP_NET_ADMIN这个capability,root执行iptables就会报错(Permission denied (you must be root));
/usr/sbin/capsh --keep=1 --user=root --drop=cap_net_admin -- -c './iptables -L;sleep 100'
可以通过 /proc 文件系统找到对应进程的 status，这样就能确认进程中的 CAP_NET_ADMIN 是否已经被移除:

# ps -ef | grep sleep
root     22603 22275  0 19:44 pts/1    00:00:00 sudo /usr/sbin/capsh --keep=1 --user=root --drop=cap_net_admin -- -c ./iptables -L;sleep 100
root     22604 22603  0 19:44 pts/1    00:00:00 /bin/bash -c ./iptables -L;sleep 100
 
# cat /proc/22604/status | grep Cap
CapInh:            0000000000000000
CapPrm:          0000003fffffefff
CapEff:             0000003fffffefff
CapBnd:          0000003fffffefff
CapAmb:         0000000000000000

当容器权限不够时,不建议直接开privilege也就是把所有的 capabilities 都赋予给容器;
因为容器中的权限越高，对系统安全的威胁显然也是越大的。比如说，如果容器中的进程有了 CAP_SYS_ADMIN 的特权之后，那么这些进程就可以在容器里直接访问磁盘设备，直接可以读取或者修改宿主机上的所有文件;
docker run --name mysql-test --privileged=true -e MYSQL_ROOT_PASSWORD=123456 -p 9092:3306  -d daocloud.io/library/mysql:5.7.6

比如容器里需要使用 iptables,只需要设置 CAP_NET_ADMIN 这个 capability 就行;
那么我们只要在运行 Docker 的时候，给这个容器再多加一个 NET_ADMIN 参数就可;
# docker run --name iptables --cap-add NET_ADMIN -it registry/iptables:v1 bash
[root@cfedf124dcf1 /]# iptables -L
Chain INPUT (policy ACCEPT)
target     prot opt source               destination
 
Chain FORWARD (policy ACCEPT)
target     prot opt source               destination
 
Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
```
