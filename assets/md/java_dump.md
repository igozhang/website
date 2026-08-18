# java_dump

```
heapdump

java11可以用root打dump，java8需要用运行用户打dump，否则报错：target process not responding or HotSpot VM not loaded
手动生成
1. jmap -dump:format=b,file=/tmp/$pid-$(date +%H%M).hprof $pid
1.1 jmap -dump:live,format=b,file=/tmp/$pid-$(date +%H%M).hprof $pid
2. jcmd $pid GC.heap_dump /tmp/$pid-$(date +%H%M).bin
3. web后台手动转储或者生成dump
jmap 主要用于离线分析，通过内存Dump保存堆转储、线程转储、线程栈、JIT 编译器的统计信息，更全面的分析内存占用，垃圾回收，线程栈，对象引用等问题;
jcmd 只会生成运行时信息和对象数量等简单信息;
jmap 生成的dump 可以用这些工具查看：
  ha457（IBM HeapAnalyzer）专注内存泄漏和性能瓶颈；
  Memory Analyzer Tool (MAT)；
  JProfiler；

jcmd 生成的dump 可以用这些工具查看：
  JVisualVM,jhat（早期版本自带，目前已弃用）；
  Eclipse Memory Analyzer (MAT)；
  YourKit Java Profiler 或 JProfiler 商业软件，要授权；
  VisualVM，可以当成JVisualVM增强版；

自动生成(参数)
-XX:+HeapDumpOnOutOfMemoryError
触发了OOME的时侯会自动在user.dir下生成java_xxxx.hprof
也可以指定目录:
-XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/data/heapdump
```

```
threaddump
A thread dump is a snapshot of the state of all the threads of a Java proces;
一般使用jstack打，也有一些图形工具后台可以直接生成比如：JMC（Java Mission Control），jvisualvm等
jstack [-F] [-l] [-m] <pid>
jstack 17264 > /tmp/threaddump.txt

打包成zip格式上传到https://gceasy.io/index.jsp可以生成图形化分析

环境没有装devel包(无jstack命令,只能用kill -3 打包)
```

```
javaCore
javacore(ThreadDump)，它包含的信息包括内存分配使用情况，JVM内部的对象锁，JVM线程的状态以及调用栈。
如果是非OOME的问题诊断，又或者是程序内部问题，例如程序内部有同步的代码（synchronized)造成的性能问题，那基本上用这个就足够了
kill -3 514353
当使用 kill -3 生成 dump 文件时，dump 文件会被输出到标准错误流。假如你的应用运行在 tomcat 上，dump 内容将被发送到<TOMCAT_HOME>/logs/catalina.out 文件里
# du -sh /opt/IBM/WebSphere/AppServer/profiles/AppSrv01/javacore.20220707.164135.514352.0001.txt
5.2M	/opt/IBM/WebSphere/AppServer/profiles/AppSrv01/javacore.20220707.164135.514352.0001.txt
```

```
systemdump
但是如果问题发生在外部调用，单纯从javaCore看不出是和谁在交互，或者一些更精确的运行时信息，例如Oracle连接的SID，Serial#，SQL语句，甚至底层的端口等等,systemdump就包含这些信息.
```

```
分析工具
1. Eclipse Memory Analyzer Tool (MAT)
Eclipse Memory Analyzer Tool(简称MAT)是一款功能强大的内存分析工具，可以采用jmap生成的hprof格式Dump文件进行分析，并提供详细的内存占用信息及内存泄漏分析报告等功能。

2. Java VisualVM
Java VisualVM是一款集垃圾回收器日志、内存、线程、CPU等多方面信息于一身的性能监控工具，可以采用jcmd生成的dump文件及命令对应用程序进行实时监控和分析。

3. JProfiler
JProfiler是一款优秀的Java性能分析工具，使用简单、功能强大。可以集成各种常见Java虚拟机(RMI、Servlet、JSP、EJB等)进行深入分析，通过内存泄漏检查器、调用树分析、方法性能分析等功能帮助寻找和解决应用程序的性能问题。

4. IBM HeapAnalyzer
IBM HeapAnalyzer是IBM官方的Java内存分析器，主要用于分析jmap生成的hprof格式Dump文件，提供丰富的内存分析功能、内存泄漏分析、趋势分析等功能。

5. YourKit Java Profiler
YourKit Java Profiler 是一款可以进行 CPU 和内存性能分析的 Java 性能调试工具。它支持使用jmap生成的hprof文件，除了内存泄漏检测外，还提供了性能调度、线程挂起、方法追踪等功能。

exp:
java -jar -Xms28g -Xmx28g C:\Users\Administrator\Desktop\ha457.jar

jhat (oracle_jdk有,openjdk没有)
jhat heapdump.bin -> http://localhost:7000
```

```
ha457  (IBM)
启动命令:(设定较大内存会打开比较快)
java -jar -Xms28g -Xmx28g C:\Users\Administrator\Desktop\ha457.jar

东方龙马邓钊星老师推荐MAT比hat好:(1,新; 2,可以显示GC ROOT,线程方法帧的局部变量,相对完整的引用路径)
mat(Memory Analyzer Tool)http://www.eclipse.org/mat/downloads.php
双击图表启动:(调整MemoryAnalyzer.ini)
1. 设置javaw程序正确位置;
2. 设置合适xmx大小;
-startup
plugins/org.eclipse.equinox.launcher_1.6.400.v20210924-0641.jar
--launcher.library
plugins/org.eclipse.equinox.launcher.win32.win32.x86_64_1.2.400.v20211117-0650
-vm
C:\Program Files\Java\jdk-11.0.16.1\bin\javaw.exe
-vmargs
-Xmx28g 
-XX:-UseGCOverheadLimit
```

```
was8553
was_console -> 故障诊断
java核心就是javacore
堆转储就是HeapDump
系统转储就是SystemDump
```
