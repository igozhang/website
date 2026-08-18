# java_ra(Report analysis)JAVA报告分析解读

```
1. jcmd
2. jstat
```

```
jcmd

exp1:
Threads class SMR info:  //(Scan Multiple Regions)
_java_thread_list=0x00007fef6000f310, length=208, elements={
0x00007ff080bf5800, 0x00007ff080bf7800, 0x00007ff080bfd000, 0x00007ff080bff800,
0x00007ff080c01800, 0x00007ff080c03800, 0x00007ff080c05800, 0x00007ff080c5a800,
  SMR（Safe Memory Reclamation）信息，用于跟踪 Java 虚拟机中的所有线程。_java_thread_list 变量存储了所有线程的信息。length 表示线程数量，elements 则是一个线程 ID 的列表；
  _java_thread_list=0x00007fef6000f310，表示 Java 虚拟机中的所有线程信息存储在这个地址中；
  length=208，表示 Java 虚拟机中有 208 个线程；
  elements={...}，表示线程列表的具体元素。

exp2:
"Service Thread" #5 daemon prio=9 os_prio=0 cpu=1242.27ms elapsed=322912.80s tid=0x00007ff080bff800 nid=0x6a6e6 runnable  [0x0000000000000000]
   java.lang.Thread.State: RUNNABLE
结果解读：
  线程名称是 "Service Thread"，表示这是 Java 应用程序中的一个服务线程。
  线程类型是 daemon，表示这是 Java 虚拟机中的一个守护线程。
  线程优先级为 9，表示这个线程的优先级比较高。
  线程操作系统优先级为 0，表示这个线程的操作系统优先级比较低。
  线程 CPU 时间为 1242.27 毫秒，表示这个线程已经占用 CPU 的时间。
  线程已经运行的时间为 322912.80 秒，表示这个线程已经运行了很长时间。
  线程 ID 为 0x00007ff080bff800，是这个线程的唯一标识符。
  线程在操作系统中的 ID 为 0x6a6e6。
  线程状态为 RUNNABLE，表示这个线程正在运行。
  堆栈信息为 [0x0000000000000000]，表示这个线程的堆栈信息为空。
  Java 线程状态为 java.lang.Thread.State: RUNNABLE，表示这个线程正在运行。

exp3:
"PeriodicMetricReader-1" #12 daemon prio=5 os_prio=0 cpu=16544.61ms elapsed=322912.35s tid=0x00007ff080f98000 nid=0x6a6ee waiting on condition  [0x00007ff04819d000]
   java.lang.Thread.State: TIMED_WAITING (parking)
        at jdk.internal.misc.Unsafe.park(java.base@11.0.16.1/Native Method)
        - parking to wait for  <0x0000000600002af0> (a java.util.concurrent.locks.AbstractQueuedSynchronizer$ConditionObject)
        at java.util.concurrent.locks.LockSupport.parkNanos(java.base@11.0.16.1/LockSupport.java:234)
        at java.util.concurrent.locks.AbstractQueuedSynchronizer$ConditionObject.awaitNanos(java.base@11.0.16.1/AbstractQueuedSynchronizer.java:2123)
        at java.util.concurrent.ScheduledThreadPoolExecutor$DelayedWorkQueue.take(java.base@11.0.16.1/ScheduledThreadPoolExecutor.java:1182)
        at java.util.concurrent.ScheduledThreadPoolExecutor$DelayedWorkQueue.take(java.base@11.0.16.1/ScheduledThreadPoolExecutor.java:899)
        at java.util.concurrent.ThreadPoolExecutor.getTask(java.base@11.0.16.1/ThreadPoolExecutor.java:1054)
        at java.util.concurrent.ThreadPoolExecutor.runWorker(java.base@11.0.16.1/Unknown Source)
        at java.util.concurrent.ThreadPoolExecutor$Worker.run(java.base@11.0.16.1/Unknown Source)
        at java.lang.Thread.run(java.base@11.0.16.1/Thread.java:829

解读如下：两部分1.线程信息; 2.堆栈信息
1.线程名称是 "PeriodicMetricReader-1"，表示这是 Java 应用程序中的一个周期性指标读取线程。
  线程类型是 daemon，表示这是 Java 虚拟机中的一个守护线程。
  线程优先级为 5，表示这个线程的优先级比较低。
  线程操作系统优先级为 0，表示这个线程的操作系统优先级比较低。
  线程 CPU 时间为 16544.61 毫秒，表示这个线程已经占用 CPU 的时间。
  线程已经运行的时间为 322912.35 秒，表示这个线程已经运行了很长时间。
  线程 ID 为 0x00007ff080f98000，是这个线程的唯一标识符。
  线程在操作系统中的 ID 为 0x6a6ee。
  线程状态为 waiting on condition，表示这个线程正在等待某个条件发生。
2.堆栈信息显示这个线程正在等待一个条件，具体等待的对象是 java.util.concurrent.locks.AbstractQueuedSynchronizer$ConditionObject，这个对象的地址是 <0x0000000600002af0>。在等待期间，这个线程使用了 parkNanos 方法进行等待。
  Java 线程状态为 java.lang.Thread.State: TIMED_WAITING (parking)，表示这个线程正在等待一个条件，等待的时间是有限的，也就是说，这个线程会在一定时间内自动唤醒。

jcmd Thread.print 输出解读：(openjdk 11.0.16.1 LTS)jstack和jcmd输出结果基本相同，不重复解读
重点关注 BLOCKED 或者 WAITING
.) 查看线程状态：每个线程的状态。如果有线程处于 BLOCKED 或者 WAITING 状态，那么就可能存在线程阻塞或者死锁的问题。如果有线程处于 RUNNABLE 状态，占用了大量 CPU 时间，那么就可能存在 CPU 过载的问题。

.) 查看堆栈信息：每个线程的堆栈信息。如果有线程的堆栈信息显示了一个死循环，那么就需要检查这个线程是否存在死循环的问题。如果有线程的堆栈信息显示了大量的 I/O 操作，那么就可能存在 I/O 阻塞的问题。

.) 查看内存使用：当前 Java 进程的内存使用情况。如果堆内存或者元数据区的使用率过高，就可能存在内存泄漏或者内存溢出的问题。

.) 查看 GC 情况：当前 Java 进程的 GC 情况。如果频繁触发 GC，或者 GC 时间过长，就可能存在内存泄漏或者内存溢出的问题。
死循环：
1.  线程状态一直处于 RUNNABLE 状态，而没有退出或者进入 BLOCKED、WAITING 或者 TIMED_WAITING 状态。
2.  线程的堆栈信息中某个方法一直在执行，并且该方法没有出口，或者循环条件一直为真。
3.  线程占用的 CPU 时间非常高，甚至达到了 100%。
死锁：
1.  多个线程相互等待：在 jstack 输出结果中，多个线程的状态都处于 BLOCKED 状态，且相互之间互相等待资源的释放。
2.  线程状态不可恢复：在 jstack 输出结果中，多个线程的状态都处于 BLOCKED 状态，且这种阻塞状态不会被解除，即线程状态不可恢复。
3.  线程占用的 CPU 时间非常高：在 jstack 输出结果中，死锁可能导致线程资源被占用，其他线程无法获得 CPU 时间，从而导致系统性能下降甚至崩溃。
GC频繁：
1. 大量线程状态为 WAITING：在 jstack 输出结果中，大量线程的状态为 WAITING 或者 TIMED_WAITING，等待 GC 完成。
2. 线程的堆栈信息中存在大量 GC 相关的方法调用：在 jstack 输出结果中，线程的堆栈信息中可能会出现大量 GC 相关的方法调用，例如 finalize 方法等。
3. 系统响应时间变慢：由于 GC 过于频繁，会导致系统响应时间变慢，甚至出现卡顿、崩溃等问题。
内存泄漏：
1. 内存使用率不断上升：在 jstack 输出结果中，可以通过查看内存使用情况来判断是否存在内存泄漏。如果堆内存或者元数据区的使用率不断上升，那么就可能存在内存泄漏的问题。
2. 线程堆栈信息中存在大量无用对象：在 jstack 输出结果中，线程堆栈信息中可能会出现大量无用的对象或者对象引用，这些对象无法被 GC 回收。
3. 系统响应时间变慢：由于内存泄漏导致内存资源被耗尽，会导致系统响应时间变慢，甚至出现卡顿、崩溃等问题。
```

```
jstat

exp1:
# jstat -gcutil $pid 2000 2
  S0     S1     E      O      M     CCS    YGC     YGCT    FGC    FGCT    CGC    CGCT     GCT   
  0.00 100.00   8.80  10.65  94.76  83.80     40    1.064     0    0.000    10    0.029    1.093
  0.00 100.00   8.88  10.65  94.76  83.80     40    1.064     0    0.000    10    0.029    1.093

S0	第一个 Survivor 区域已使用的百分比
S1	第二个 Survivor 区域已使用的百分比
E	Eden 区域已使用的百分比
O	老年代已使用的百分比
M	Metaspace(元空间)已使用的百分比
CCS	压缩 Class Space 已使用的百分比
YGC	Young GC 的次数
YGCT	Young GC 的总耗时
FGC	Full GC 的次数
FGCT	Full GC 的总耗时
CGC	Concurrent Mark Sweep (CMS) GC 的次数
CGCT	CMS GC 的总耗时
GCT	所有 GC 的总耗时
其中，S0、S1、E、O、M、CCS表示堆内存区域的使用情况，YGC、YGCT、FGC、FGCT、CGC、CGCT、GCT表示垃圾回收的次数和耗时
```
