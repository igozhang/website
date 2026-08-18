# linux_ra,liunx常见报告解读

```
ps
free
top
fio
ifconfig
ethtool
```

```
ps

exp01:
# ps aux --sort=-%mem | head -n 5
USER        PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root      85022  141 28.6 13140232 9400684 ?    Ssl  09:00   0:09 /usr/share/filebeat/bin/filebeat --environment systemd -c /etc/filebeat/filebeat.yml --path.home /usr/share/filebeat --path.config /etc/filebeat --path.data /var/lib/filebeat --path.logs /var/log/filebeat
root      98316  0.0  0.8 2477288 271732 ?      Sl    2021 262:19 /opt/TEC/WebServer/DWSvc
root     291106  0.1  0.0  39124 14684 ?        Sl    2022 697:54 ./processbeat -c ../etc/processbeat.conf
root      81845  6.0  0.0 720948 12828 ?        Ssl  08:52   0:27 /usr/local/bin/node_exporter --collector.tcpstat

虚拟内存——虚拟内存是计算机系统内存管理的一种技术。它使得应用程序认为它拥有连续可用的内存（一个连续完整的地址空间），而实际上，它通常是被分隔成多个物理内存碎片，还有部分暂时存储在外部磁盘存储器上，在需要时进行数据交换。与没有使用虚拟内存技术的系统相比，使用这种技术的系统使得大型程序的编写变得更容易，对真正的物理内存（例如RAM）的使用也更有效率。
java_new分配内存只是操作系统内部做了一个标记，不会立刻实际分配;
1，机器总内存16GB，一次性申请new 20GB内存，会申请失败。
2，三个进程，分别new 10GB内存，没问题。
3，free命令的输出无明显变化。“虚拟内存”的占用，在free命令无法展示出来。

单位均为KB
VSZ——virtual memory size of the process in KiB (1024-byte units).
RSS——resident set size, the non-swapped physical memory that a task has used (in kiloBytes).
```

```
fio_ra

tee >igofio.ini <<EOF
[global]
ioengine=libaio
runtime=30
time_based
direct=1
buffered=0
group_reporting

[read-test]
rw=read
bs=4k
numjobs=1
size=1G
directory=/data

[write-test]
rw=write
bs=4k
numjobs=1
size=1G
directory=/data
EOF

# global部分定义了引擎，运行时间是否直接访问磁盘等；
# r/w部分定义了读写方式、块大小、并发数、测试文件大小和路径等；

报告：
Run status group 0 (all jobs):
   READ: bw=18.8MiB/s (19.7MB/s), 18.8MiB/s-18.8MiB/s (19.7MB/s-19.7MB/s), io=563MiB (590MB), run=30001-30001msec
  WRITE: bw=5890KiB/s (6031kB/s), 5890KiB/s-5890KiB/s (6031kB/s-6031kB/s), io=173MiB (181MB), run=30001-30001msec

Disk stats (read/write):
  vda: ios=143224/44028, merge=1/179, ticks=30378/28537, in_queue=58915, util=99.67%
解读：
测试期间30S
读4k读平均18.8MiB/s，总读io590M；
写平均5890K/s，总写io173M；

测试期间磁盘状态（vda）
I/O 操作(IOS) : 143,224个读操作和44,028个写操作。
合并: 1个读操作和179个写操作被合并。
计时: 读操作花费30,378毫秒，写操作花费28,537毫秒。
In _ queue: 在队列中等待的时间为58,915毫秒。
利用率(util) : 99.67%-这表明磁盘在测试期间的利用率为99.67%
```

```
ifconfig_ra

exp1
eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.0.20.8  netmask 255.255.252.0  broadcast 10.0.23.255
        inet6 fe80::5054:ff:fe7f:99e9  prefixlen 64  scopeid 0x20<link>
        ether 52:54:00:7f:99:e9  txqueuelen 1000  (Ethernet)
        RX packets 11765714  bytes 1970976644 (1.8 GiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 10941603  bytes 1698793476 (1.5 GiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

这是Linux系统中使用ifconfig命令查看eth0网络接口的详细信息。具体解释如下：

- flags：表示网络接口的标志，包括UP（接口已启用）、BROADCAST（接口支持广播）、RUNNING（接口正在运行）、MULTICAST（接口支持多播）等；
- mtu：表示网络接口的最大传输单元（Maximum Transmission Unit），即一次数据传输的最大长度；
- inet：表示网络接口的IPv4地址；
- netmask：表示网络接口的子网掩码；
- broadcast：表示网络接口的广播地址；
- inet6：表示网络接口的IPv6地址；
- prefixlen：表示网络接口的IPv6地址的前缀长度；
- scopeid：表示网络接口的作用域；
- ether：表示网络接口的MAC地址；
- txqueuelen：表示网络接口的传输队列长度；
- RX packets：表示接收到的数据包的数量；
- bytes：表示接收到的数据包的总字节数；
- RX errors：表示接收时发生错误的数据包的数量；
- dropped：表示接收时被丢弃的数据包的数量；
- overruns：表示接收缓冲区溢出的次数；
- frame：表示接收到帧错误的数据包的数量；
- TX packets：表示发送的数据包的数量；
- bytes：表示发送的数据包的总字节数；
- TX errors：表示发送时发生错误的数据包的数量；
- dropped：表示发送时被丢弃的数据包的数量；
- overruns：表示发送缓冲区溢出的次数；
- carrier：表示发送时发生载波错误的数据包的数量；
- collisions：表示发送时发生碰撞的数据包的数量。
```

```
ethtool_ra

exp1：
# ethtool eth0
Settings for eth0:
 Supported ports: [ TP ]
 Supported link modes:   1000baseT/Full 
                         10000baseT/Full 
 Supported pause frame use: No
 Supports auto-negotiation: No
 Advertised link modes:  Not reported
 Advertised pause frame use: No
 Advertised auto-negotiation: No
 Speed: 10000Mb/s
 Duplex: Full
 Port: Twisted Pair
 PHYAD: 0
 Transceiver: internal
 Auto-negotiation: off
 MDI-X: Unknown
 Supports Wake-on: uag
 Wake-on: d
 Link detected: yes
 
 解读：
  Supported link modes：表示网卡支持的连接模式，包括1000baseT/Full和10000baseT/Full，分别表示1Gbps和10Gbps的全双工模式；
  Supports auto-negotiation：表示网卡不支持自动协商；
  Advertised link modes：表示网卡未报告自己的连接模式；
  Speed：表示网卡的速率为10Gbps；
  Duplex：表示网卡的双工模式为全双工；
  Auto-negotiation：表示网卡的自动协商功能已关闭；
  Link detected：表示网卡检测到了连接。
```

```
free
exp01:
# free -m
              total        used        free      shared  buff/cache   available
Mem:           3694         281         167           0        3245        3121
Swap:             0           0           0

used—— Used memory (calculated as total - free - buffers - cache)（用户已使用的物理内存，可以看到其中排除了buff/cache）
free—— Unused memory (MemFree and SwapFree in /proc/meminfo)（整机未使用的物理内存）
buffers—— Memory used by kernel buffers (Buffers in /proc/meminfo)（内核用的）
cache—— Memory used by the page cache and slabs (Cached and Slab in /proc/meminfo)（内核用的）
buff/cache—— Sum of buffers and cache（可以理解为内核用来作缓存的一部分内存，可以随时释放出来给用户用）
available—— Estimation of how much memory is available for starting new applications, without swapping.（一个估算值，真正可供用户进程使用的内存值，比free那列更有意义）
总结：used + free + buff + cache = total
内核的策略是优先用free里面的，用到剩差不多100MB的时候，开始从缓存中释放内存给用户进程使用。
内存申请失败问题:内核Heuristic overcommit算法在每次申请时计算出来的一个值[5]，不考虑swap的话这个值大致就是available列的值;

Mem那行的公式是：total = used + free，这个used代表整机已使用的内存，基本等于新内核里used + buff + cache；free含义没变化。
buff/cache那行的公式也是：total = used + free，这个used代表：刨去buff/cache后用户使用的内存，即新内核里used的含义；free代表：Mem行的free + buff/cache，有点类似新内核的available（注意只是类似）。
现在你再看buff/cache那行开头的“-/+”符号，[7]分别代表Mem行的used减去、free加上（buff+cache）后的
```
