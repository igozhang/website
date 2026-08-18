# Linux性能诊断

```
CentOS Linux release 7.6.1810 (Core)
```

```
1. 常见问题及原因
2. 资源情况及诊断
3. tips,网卡磁盘速率诊断
```

```
journalctl is used to print the log entries stored in the journal by systemd-journald.service(8) and systemd-journal-remote.service(8)
journalctl -p err..alert
journalctl -f
jurnalctl -b
journalctl -u nginx.service
journalctl --since "2023-03-03" --until "2023-03-03 03:00"

/var/log/messages  主日志
/var/log/dmesg  启动日志
```

```
概述及常见问题及原因：

找出CPU/MEM/DISK_IO资源占用情况及按连接数排序IP
ps aux --sort=-cpu | head -n 6
ps aux --sort=-%mem | head -n 6
perf top 可以查占用cpu时间最多的函数（热点函数）或指令
iotop  -n1 |head -n 8
netstat -nat | awk '{print $5}' | cut -d: -f1 | sort | uniq -c | sort -nr
找出CPU及MEM消耗前5线程
top -H -b -n 1 -p $pid |sed -n '7p' && top -H -b -n 1 -p $pid |sed -n '8,$p' |sort -k9nr |head -n 5
top -H -b -n 1 -p $pid |sed -n '7p' && top -H -b -n 1 -p $pid |sed -n '8,$p' |sort -k10nr |head -n 5

echo "5秒内高频采样10次计算平均值...";total=10;tmp=$(mktemp);for ((i=0;i<total;i++)); do ps -eo pid,%cpu,%mem,rss,cmd --no-headers >> $tmp; sleep 0.5; done;cpu_total=$(grep -c ^processor /proc/cpuinfo);echo -e "\n=== CPU 占用 Top3（平均值） ===";awk -v cpus="$cpu_total" -v t="$total" '{cpu[$1]+=$2; mem[$1]+=$3; rss[$1]+=$4; cmd[$1]=$5} END{for(p in cpu){pcpu=cpu[p]/t; cc=pcpu/cpus; printf "PID: %-6s CPU: %.2fC (%.1f%%)  MEM: %.1f%%  RSS: %.0fK  命令: %s\n", p, cc, pcpu, mem[p]/t, rss[p]/t, cmd[p]}}' $tmp | sort -k7 -nr | head -3;echo -e "\n=== 内存占用 Top3（平均值） ===";awk -v t="$total" '{cpu[$1]+=$2; mem[$1]+=$3; rss[$1]+=$4; cmd[$1]=$5} END{for(p in cpu){mrss=rss[p]/t; if(mrss>1024*1024){m=sprintf("%.2fG", mrss/1024/1024)}else if(mrss>1024){m=sprintf("%.2fM", mrss/1024)}else{m=sprintf("%.0fK", mrss)} printf "PID: %-6s MEM: %-8s (%.1f%%)  CPU: %.1f%%  命令: %s\n", p, m, mem[p]/t, cpu[p]/t, cmd[p]}}' $tmp | sort -k6 -nr | head -3;rm -f $tmp

1. CPU负载过高
问题表现：当CPU负载过高时，系统可能会变得缓慢、响应时间慢甚至无法响应。可以使用命令top或htop查看系统负载情况。
问题定位方法：可以使用命令pidstat查看进程的CPU使用情况，找出哪些进程使用了过多的CPU资源。可以使用命令ps或pstree查看进程之间的关系。
常用命令：
    top：实时显示进程的CPU、内存、I/O使用情况
    htop：类似top，但提供了更多的功能和可视化界面
    pidstat：显示进程的CPU、内存、I/O使用情况
    ps：显示进程信息
    pstree：以树形结构显示进程之间的关系

2. 磁盘I/O负载过高 iotop
问题表现：当磁盘I/O负载过高时，系统可能会变得缓慢、响应时间慢或出现磁盘访问超时等问题。可以使用命令iostat查看磁盘I/O使用情况。
问题定位方法：可以使用命令iotop查看哪些进程正在导致磁盘I/O负载过高。也可以使用命令lsof查看哪些进程正在访问磁盘文件。
常用命令：
    iostat：显示系统的磁盘I/O使用情况
    iotop：显示哪些进程导致磁盘I/O负载过高(iotop -oP)
    lsof：显示进程正在访问的文件
	
3. 内存使用过高
问题表现：当内存使用过高时，系统可能会开始使用交换空间，导致系统变得缓慢、响应时间慢。可以使用命令free查看系统内存使用情况。
问题定位方法：可以使用命令ps或top查看进程的内存使用情况，找出哪些进程使用了过多的内存。也可以使用命令vmstat查看系统的虚拟内存使用情况。
常用命令：
    free：显示系统的内存使用情况
    ps：显示进程信息
    top：实时显示进程的CPU、内存、I/O使用情况
    vmstat：显示系统的虚拟内存使用情况

4. 网络流量过高
问题表现：网络响应缓慢，网络传输速度慢。
发现问题的过程：使用iftop或者nethogs命令查看网络流量情况。
解决方法：优化网络配置，增加带宽或者关闭不必要的网络连接。 
  网络测试工具：测试工具可以用于测试网络连接的质量、带宽、丢包率、延迟等参数，常见的网络测试工具有ping、traceroute、mtr等。
  网络监控工具：监控工具用于实时监控网络流量、连接数量、连接速度等参数，也可以用于监控网络设备的状态，常见的网络监控工具有Nagios、Zabbix、Cacti等。
  网络分析工具：分析工具可以用于分析网络流量、协议、数据包等，帮助排查网络故障和安全问题，常见的网络分析工具有Wireshark、tcpdump等。
  网络诊断工具：诊断工具可以用于诊断网络故障，常见的网络诊断工具有nslookup、dig、netstat、telnet、curl、wget等。
  性能测试工具：性能测试工具可以用于测试Web应用的性能，包括负载测试、稳定性测试、压力测试等，常见的性能测试工具有JMeter、LoadRunner、WebLOAD等。

5. 错误日志过多
问题表现：系统稳定性下降，日志文件过大。
发现问题的过程：查看系统日志文件，查找错误日志。
解决方法：解决引起错误的问题，或者优化日志记录方式。
```

```
资源分析诊断
1. 通览，快速判断是否存在资源瓶颈
先zabbix，grafana等工具抓取问题时间段的数据，必要时使用shell命令抓取补充，主要有：
1.1 cpu_util ;
结合sar看下27号cpu队列及使用情况
sar -q -f /var/log/sa/sa27
sar -u -f /var/log/sa/sa27
间隔1秒查3次
sar -q 1 3

1.2 mem_util ;
sar -r -f /var/log/sa/sa27

1.3 disk_io ;
关注disk queue size队列（8以下正常）和disk_util

先测磁盘读写性能，再对比监控数据；
dd if=/dev/zero of=testfile bs=1M count=1000 conv=fdatasync
dd if=testfile of=/dev/null bs=1M count=1000
iostat -x 1
iostat -x /dev/sda1 1
iostat -x -p PID 1
如果 rkB/s 和 wkB/s 的数值较高，且 await 和 svctm 的数值也较高，就可能存在磁盘 IO 瓶颈；
如果只是util比较高，表示磁盘正在高负载运行；

await：表示磁盘平均等待时间，单位为毫秒。
svctm：表示磁盘平均服务时间，单位为毫秒。
util：表示磁盘的利用率，即磁盘正在处理 I/O 请求的时间比例。

1.4 net
a. 先看error和丢包dropped
ifconfig eth0

b. 再看当前吞吐，结合比较网络理论吞吐；
当前吞吐：
$ sar -n DEV 1 10 |grep -i eth0
$ iftop -i eth0   可以查看吞吐和峰值peak
理论吞吐： ethtool bond0

ping -c 5 igozhang.cn 检测连通性并输出延迟；
traceroute igozhang.cn 数据包从本机到目标主机所经过的中间节点，并输出每个节点的延迟时间

2. 诊断，找出具体程序，进程号，关键函数并判断是否存在异常或者可优化空间；

2.1 找出内存cpu消耗高程序
ps -eo pid,ppid,cmd,%cpu --sort=-%cpu | head
ps -eo pid,ppid,cmd,%mem --sort=-%mem | head
top -H -p $pid  查看线程信息
ps -T -p $pid

2.2 找出网络链接最多程序
a 找出连接数最多的ip，top1_ip;
$ netstat -nat | awk '{print $5}' | cut -d: -f1 | sort | uniq -c | sort -nr

b 以top1_ip为目标抓取数据包,可以从数据中看到流量最大的应用程序；
$ tcpdump -i eth0 -A -s 1500 src host top1_ip;

c 找出进程ID
$ lsof -i :80 | grep LISTEN

2.3 strace "trace system calls and signals"
跟踪应用程序的系统调用，例如文件读写、进程创建等，以便分析应用程序的行为和性能瓶颈。strace 命令还可以输出调用堆栈和错误信息等
strace -p 22388 -ff -o s_redis.txt
分多个文件分类记录
strace -cp 22388  -o s_redis.txt
汇总进程信息并写入到文件

-r timestamp -T timespent -t wall clock time

系统调用：程序使用了哪些系统调用，是否有重复或无用的系统调用等情况
信号：程序接收了哪些信号，是否有未处理的信号等情况
错误信息：程序发生了哪些错误，错误的原因和位置等信息
时间信息：每个系统调用的执行时间，找到程序的性能瓶颈

2.4 perf "Performance analysis tools for Linux"
系统性能分析，包括 CPU 周期、缓存命中率、内存访问等。perf 命令还可以生成火焰图和统计图等

perf top -p 22388 -g -K 可以敲回车具体查看
(-g 程序调用栈； -K应用层函数)

2.4.2 抓取cpu事件并生成火焰图
perf record -F 99 -p 22388 -g -- sleep 10
每秒99次的频率对进程所有cpu事件采样10秒，并记录程序调用栈(-g),(-a --all-cpus)

# 下载工具
$ git clone --depth 1 https://github.com/brendangregg/FlameGraph.git
# 折叠调用栈
$ FlameGraph/stackcollapse-perf.pl perf.data > perf.folded
# 生成火焰图
$ FlameGraph/flamegraph.pl perf.folded >out.svg

2.5 valgrind 
可以进行内存泄漏和错误检测，以及性能分析。valgrind 命令可以检查应用程序的内存使用情况、线程安全性等，并生成报告和统计信息。

2.6 gdb
进行应用程序的调试和性能分析，包括断点调试、变量监视、内存访问等。gdb 命令还可以输出调用堆栈和代码执行路径等，方便分析性能瓶颈。
```

```
网卡磁盘速率判断
磁盘disk：
额定速度：
1. 机械硬盘的读写速度一般在100MB/s至200MB/s之间，取决于转速、缓存大小和接口类型等因素。
2. 固态硬盘的读写速度要比机械硬盘快很多，一般在500MB/s至600MB/s之间，高端产品甚至可以达到3000MB/s以上的读写速度。SSD的读写速度取决于闪存芯片类型、控制器性能和接口类型等因素。

实际速度
dd if=/dev/zero of=igoTfile bs=1M count=1000
dd if=igoTfile of=/dev/null bs=1M count=1000

最近运行读写速率
iostat -dxk 11 1  \\每隔11s输出1次磁盘读写信息1次
-d”参数表示显示磁盘统计信息，“-x”参数表示显示详细信息，“-k”参数表示显示单位为KB/s
 dstat -cdD total,sda 1 1  \\每隔1s统计-cdD(cpu,disk,具体sda)1次

网络net：
额定速度：(ethtool eth0 有speed字段)
当前主流的1G、10G、100G服务器网卡的大概实际读写速度如下：
1G服务器网卡的理论最大传输速度为1Gbps，但实际上由于协议开销等因素，实际传输速度一般在800Mbps左右，即100MB/s左右。
10G服务器网卡的理论最大传输速度为10Gbps，在8Gbps至9Gbps之间，即1GB/s至1.125GB/s之间。
100G服务器网卡的理论最大传输速度为100Gbps，在80Gbps至90Gbps之间，即10GB/s至11.25GB/s之间

实际传输速度：
iperf -s
iperf -c serv_ip

最近传输速度：使用zabbix，prometheus等工具吧
```
