# ali_arthas工具

```
支持JAVA 6+版本
Arthas 是一款线上监控诊断产品，通过全局视角实时查看应用 load、内存、gc、线程的状态信息，并能在不修改应用代码的情况下，对业务问题进行诊断，包括查看方法调用的出入参、异常，监测方法执行耗时，类加载信息等，大大提升线上问题排查效率
```

```
https://github.com/alibaba/arthas/releases
https://arthas.aliyun.com/doc/arthas-tutorials?language=cn
https://arthas.aliyun.com/doc/install-detail.html
```

```
可以连外网环境可用
wget https://arthas.aliyun.com/arthas-boot.jar
java -jar arthas-boot.jar

不能连外网环境需要下载全量版本
https://arthas.aliyun.com/doc/install-detail.html
(搜索全量安装)
https://arthas.aliyun.com/download/latest_version?mirror=aliyun

或者rpm,deb包安装
https://github.com/alibaba/arthas/releases
```

```
CRUD:

安装:
unzip arthas-bin.zip

启动:
java -jar arthas-boot.jar
(会要求选择监控进程)

dashboard:
展示当前进程的信息;
thread 1会打印线程 ID 1 的栈，通常是 main 函数的线程;

jad反编译:
jad demo.MathGame (方法名)
```

```
thread:
thread -n 5  # 显示 CPU 占用最高的 5 个线程
thread <thread-id>  # 查看指定线程的堆栈信息
thread -b  # 查找并显示死锁的线程

jvm 命令可以查看 JVM 的详细信息，包括 JVM 版本、JVM 参数、类加载器信息等。

sc:
sc com.example.MyClass
sc（Search Class）命令用于搜索和查看 JVM 中加载的类信息。你可以通过类名、包名或部分匹配来查找类
按包名查找：
sc com.example.*
sc -d com.example.MyClass  # 查看类的详细信息，包括类加载器、类的字节码大小等

trace:
trace 命令用于追踪方法的调用链路，显示方法内部的每一层调用及其耗时。这对于分析方法的性能瓶颈非常有用
trace com.example.MyClass myMethod

watch:
watch com.example.MyClass myMethod '{params, returnObj}' -x 2
其中，{params, returnObj} 表示同时监控方法的参数和返回值，-x 2 表示只打印前两次调用的结果
watch com.example.MyClass myMethod 'throwExp' -x 2
只打印方法抛出的异常

jad 命令用于反编译 Java 类文件，帮助你查看类的源代码。这对于在没有源码的情况下调试和理解类的行为非常有用
反编译类：
jad com.example.MyClass
反编译类中的特定方法：
jad com.example.MyClass.myMethod

系统属性和环境变量：
sysprop  这将显示 JVM 的系统属性，如 java.version、user.home 等
sysenv  这将显示系统的环境变量。

生成一个堆转储文件：
heapdump /path/to/heapdump.hprof

性能问题：使用 trace 和 watch 命令追踪方法的调用链路和输入输出，找出性能瓶颈。
线程问题：使用 thread 命令查看线程状态，查找死锁、阻塞等问题。
内存泄漏：使用 heapdump 命令生成堆转储文件，分析内存泄漏。
Bug 修复：使用 redefine 命令动态修改类，修复线上 Bug。
日志不足：使用 watch 和 trace 命令补充日志，查看方法的调用情况。
```
