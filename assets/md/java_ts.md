# java_ts

```
# Java问题思路

## CPU高:
arthas dashboard + thread
1. 使用arthas dashboard命令查看CPU消耗高的线程,记录线程号TID
2. thread $tid 查看线程堆栈 (相当于jstack tid)
3. 如果是偶发CPU突增可以抓火焰图 igozhang.cn/java_flame
4. 找到业务代码堆栈后可以使用jad反编译到业务代码
   jad com.igozhang.scenario.Scenarioi$1 --linenumber true
也可以直接使用java-cpu-aly.sh $pid 找到CPU占用高的线程并打印线程堆栈;

## MEM高:
1. 找到高MEM消耗进程;
2. 打dump  igozhang.cn/java_dump

实时监控GC，判断是否频繁GC和FULLGC（间隔1S）
jstat -gc $pid 1000

查看当前 Java 进程的堆内存配置和使用概况
可以判断是否堆设置过小或过大，以及是否达到了上限
jmap -heap $pid

触发一次fullGC,找出大于1M的类
jmap -histo:live $pid | awk '$3 > 1048576' | head -n 5
printf %x $tid (10转16进制再jsatck文档中搜索)

## IO突增:
1. 打线程dump
jstack $pid >/tmp/igostackdump.txt
2. 找到IO消耗高的线程号tid
pidstat -p $pid -t -d
3. 转换成16进制tid
printf %x $tid
4. 使用16进制tid在igostackdump文件中匹配对应堆栈信息
vim igostackdump.txt

## 数据库异常
### 连接异常
类似报错:
Caused by: java.sql.SQLTransientConnectionException: master - Connection is not available, request timed out after 3004ms.
The last packet successfully received from the server was 64,367 milliseconds ago. The last packet sent successfully to the server was 3,00 milliseconds ago.
hikari连接池max-lifetime 过大,大于数据库lifetime,导致连接池断开了连接,数据库还没有释放问题,调小到30000可以解决问题;
SHOW VARIABLES LIKE '%TIMEOUT%';
https://blog.csdn.net/weixin_45654405/article/details/126426078

### 连接池满
Unable to acquire JDBC Connection
Connection is not available, request timed out after 30000ms.
1. 先考虑调大连接池连接数参数
Hikari：maximum-pool-size
Druid：maxActive
2. 再考虑调大从连接池获取连接的等待时间
Hikari：connectionTimeout
Druid：max-wait

### 连接池满(数据库端异常)
报错类似Too many connections
找DBA查
1. 数据库最大连接数
show variables like '%max_connections%';
2. 查数据库当前连接数
show full processlist;

## 常见线程
C1 ComplierThread和C2 ComplierThread线程
VM Thread
GCTaskThread

线程状态
RUNNABLE 就绪状态
BLOCKED  等锁从而进入就绪状态
WAITING  该线程需等待其他线程通知或中断
TIMED_WAITING 指定时间返回
```

```
CPU消耗前5JAVA线程
top -H -b -n 1 -p $pid |sed -n '7p' && top -H -b -n 1 -p $pid |sed -n '8,$p' |sort -k9nr |head -n 5
pidstat -t -p $pid |head -8
MEM消耗前5JAVA线程
top -H -b -n 1 -p $pid |sed -n '7p' && top -H -b -n 1 -p $pid |sed -n '8,$p' |sort -k10nr |head -n 5
```

```
JAVA接口慢问题定位

首先找开发要接口对应【类全路径】和【方法名】
然后使用arthas的trace命令: trace 类路径 方法名:
trace com.igo.dms.business.outside.controller.OutsideDefectController create 
再定位到耗时最长的【类全路径】和【方法名】,然后继续
一层一层抓下去直到找到最终耗时的地方;
一般抓到不是自己公司的代码(底层三方公司代码)为止,比如我公司代码都是com.igo.XXX;

exp1:
开发提供的类全路径和方法
com.igo.dms.business.outside.controller.OutsideDefectController create
com.igo.dms.business.outside.controller.OutsideDefectController 类全路径
create 方法名
trace com.igo.dms.business.outside.controller.OutsideDefectController create 

exp2:
com.igo.dms.business.alm.vo.CreateWorkItemReq:setChangedBy() #255 开发提供的类全路径和方法
com.igo.dms.business.alm.vo.CreateWorkItemReq  类全路径
setChangedBy 方法名
#255在相关代码文件中的位置(以行数为单位)
```

```
常用工具 jcmd jstat jstack gc_log jmc
jps,jcmd 查看当前java进程
jcmd $pid GC.heap_dump /tmp/$pid-$(date +%H%M).hprof 打heap_dump
jcmd $pid GC.heap_info 查进程堆使用情况
jcmd $pid Thread.print $tid 查进程里某个线程的堆使用情况
jcmd $pid VM.flags 输出java虚拟机参数
其中UseConcMarkSweepGC是CMS垃圾回收器，UseG1GC是G1回收器
jcmd $pid VM.uptime 输出java运行时间，单位秒

curl -L https://arthas.aliyun.com/install.sh | sh
https://arthas.aliyun.com/doc/commands.html
command:
dashboard  //打开面板
thread 94  //看线程栈
thread -b  //定位block线程
jad com.igozhang.scenario.Scenario2$2  //反编译代码

jstat [options] <vmid> <interval> [<count>] //jvm性能诊断工具
jstat -gcutil $pid 2000 2  //统计gc信息2次，间隔2000ms (次数是自JVM启动以来的次数，时间是时间和，单位是S秒)
jstat -gccause $pid <interval> <count>  \\这条命令会输出每次GC的原因、时间、耗时、Heap Memory使用量等信息，从中可以找出频繁进行GC的内存区域及频繁回收的对象类型。如果某个对象被频繁回收，说明它经常被创建和销毁。

jstat -gc $pid gc统计
S0C：第一个幸存区的大小
S1C：第二个幸存区的大小
S0U：第一个幸存区的使用大小
S1U：第二个幸存区的使用大小
EC：伊甸园区的大小
EU：伊甸园区的使用大小
OC：老年代大小
OU：老年代使用大小
MC：方法区大小
MU：方法区使用大小
CCSC:压缩类空间大小
CCSU:压缩类空间使用大小
YGC：年轻代垃圾回收次数
YGCT：年轻代垃圾回收消耗时间
FGC：老年代垃圾回收次数
FGCT：老年代垃圾回收消耗时间
GCT：垃圾回收消耗总时间

jstat -gccapacity $pid  堆内存统计
NGCMN：新生代最小容量
NGCMX：新生代最大容量
NGC：当前新生代容量
S0C：第一个幸存区大小
S1C：第二个幸存区的大小
EC：伊甸园区的大小
OGCMN：老年代最小容量
OGCMX：老年代最大容量
OGC：当前老年代大小
OC:当前老年代大小
MCMN:最小元数据容量
MCMX：最大元数据容量
MC：当前元数据空间大小
CCSMN：最小压缩类空间大小
CCSMX：最大压缩类空间大小
CCSC：当前压缩类空间大小
YGC：年轻代gc次数
FGC：老年代GC次数

jstat -compiler $pid
Compiled：编译数量。
Failed：失败数量
Invalid：不可用数量
Time：时间
FailedType：失败类型
FailedMethod：失败的方法

jstat -class $pid
Loaded:加载class的数量
Bytes：所占用空间大小
Unloaded：未加载数量
Bytes:未加载占用空间
Time：时间

jstack [option] pid
Java自带命令，主要看线程的状态，包括线程堆栈信息，线程的运行状态以及被阻塞等信息；
jstack命令适用于以下场景：
1. 应用程序出现了线程锁死等问题，需要查看线程的状态、堆栈信息，以及线程之间的锁信息。
2. 分析应用程序的性能瓶颈，了解应用程序线程的运行状态和堆栈信息，帮助分析线程运行情况。
   注意：jstack命令对Java应用程序的诊断对应用程序性能有一定的影响，生产环境要谨慎。另外，jstack命令输出的信息较为详细，可以用jstackLogParser，jhat，Eclipse MAT，VisualVM等工具分析。
FastThread
https://fastthread.io/
VisualVM
https://visualvm.github.io/services/threaddump.html

jstack $pid > jstack.out
输出进程中某个线程的状态和堆栈信息：
jstack -l $pid | grep $tid16x
输出进程中线程名为thread-name的线程的状态信息和堆栈信息。
jstack -F -l $pid | grep -A1 "State"
输出进程中线程的状态信息统计结果，并统计每个状态的线程数量。
jstack -l $pid | grep -Ei "(deadlock)|(locked)"
输出进程中的死锁信息，以及当前处于锁状态的线程信息。

gc_log:
启动命令: -Xlog:gc*:file=/data/logs/igoapp_gc.log:time,uptime:filecount=10,filesize=10m \
GCViewer：是一个免费的GC日志分析工具，可以生成可视化的GC报告
GCEasy：是一个在线的GC日志分析工具，可以分析和优化GC日志

JMC(Java Mission Control)
```

```
Java进程的Heap Memory快速增长通常是由以下原因引起的：

    1. 内存泄漏：内存泄漏是指程序在使用过程中，因为某些原因导致内存无法被垃圾回收器回收而一直占用着内存。如果程序中存在内存泄漏，Heap Memory就会快速增长，最终导致OutOfMemoryError。
### 看日志是否oom；
### 在jhat的网页http://localhost:7000/左侧的导航栏中选择“Top Objects”查看大对象排行；

    2. 频繁的对象创建和销毁：如果程序中频繁创建和销毁对象，就会导致Heap Memory的增长。
###在jhat的网页http://localhost:7000/上，选择“Object Histogram”选项卡，然后点击对象类别的“Surviving Instances”或“Total Size”列，排序并查看排名前10的。在这里，Surviving Instances指的是在GC之后幸存下来的对象，Total Size指的是对象的大小之和。如果某个对象的Surviving Instances很少而Total Size很大，就说明该对象经常被创建和销毁。
###jstat -gccause $pid <interval> <count>

    3. 频繁的Full GC：Full GC会导致应用程序的暂停，如果Full GC太频繁就会导致应用程序的性能问题，同时也会导致Heap Memory的增长。

为了找出Heap Memory快速增长的原因，可以考虑以下几个方面：

    分析Heap Memory的Dump File：使用jmap工具生成Heap Dump文件，并使用MAT (Eclipse Memory Analyzer Tool) 工具进行分析，找出Heap Memory中占用最多的对象与类型，并查看占用内存的线程信息。

    监测GC日志：使用JVM的GC日志可以观察应用程序的堆内存使用情况，包括GC的时机、频率、时间等信息。

    分析应用程序的代码：查看应用程序在运行期间是否有频繁的对象创建和销毁。

    监控应用程序：使用监控工具 (如JVisualVM、YourKit Java Profiler等) 对应用程序进行监控，查看是否有线程占用过多内存等问题
```

```
压测工具
JMeter
```

```
调整年轻代大小YG：
1. 查看当前JVM的GC参数和年轻代大小：
java -XX:+PrintCommandLineFlags -XX:+PrintGCDetails -version
该命令会输出当前JVM的GC参数和年轻代大小等信息。

2. 调整年轻代大小：
-XX:NewSize=<size>  # 设置年轻代的初始大小
-XX:MaxNewSize=<size>  # 设置年轻代的最大大小(k,m,g)
exp:
-XX:NewSize=512m -XX:MaxNewSize=1g

一般可以设置成1/4到1/3的整体heap_mem大小，比如：
java -Xms16g -Xmx16g -XX:NewSize=4g -XX:MaxNewSize=4g

3. 调整GC参数：
-XX:MaxTenuringThreshold=<n>  # 设置对象晋升老年代的年龄阈值
-XX:SurvivorRatio=<n>  # 设置年轻代中Eden区域和Survivor区域的比例
-XX:NewRatio=<n>  # 设置年轻代和老年代的比例
-XX:+UseParNewGC  # 使用并行化的年轻代GC
-XX:+UseConcMarkSweepGC  # 启用CMS垃圾回收
-XX:+UseG1GC  # 启用G1垃圾回收

其中，<n>表示一个整数。
exp:
将对象晋升老年代的年龄阈值设置为15，年轻代中Eden区域和Survivor区域的比例设置为8:1，年轻代和老年代的比例设置为1:2，使用并行化的年轻代GC和CMS垃圾回收，可以使用以下参数：
-XX:MaxTenuringThreshold=15 -XX:SurvivorRatio=8 -XX:NewRatio=2 -XX:+UseParNewGC -XX:+UseConcMarkSweepGC

4. 重启Java进程
```

```
阿里建议JVM参数：
-XX:+UseContainerSupport -XX:InitialRAMPercentage=70.0 -XX:MaxRAMPercentage=70.0 -XX:+PrintGCDetails -XX:+PrintGCDateStamps -Xloggc:/home/admin/nas/gc-${POD_IP}-$(date '+%s').log -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/home/admin/nas/dump-${POD_IP}-$(date '+%s').hprof

https://help.aliyun.com/zh/sae/use-cases/best-practices-for-jvm-heap-size-configuration?spm=5176.26934566.main.25.20de4474BoIkyJ
```
