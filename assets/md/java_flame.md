# JAVA火焰图 java_flame

## env

1. 工具Async-profiler  
   Async-profiler  
   https://github.com/jvm-profiling-tools/async-profiler

```
git clone https://github.com/jvm-profiling-tools/async-profiler
cd async-profiler
make
```

2. 工具arthas\_profiler  
   https://arthas.aliyun.com/doc/profiler.html

## 简要思路

1. 确定是CPU问题还是MEM问题
2. 生成CPU/MEM火焰图
3. 查看火焰图(横条越长表示CPU/MEM资源占用越多)

## 获取PID

```
jps
ps -eo pid,ppid,cmd,%cpu --sort=-%cpu | head -n 4
```

## 生成火焰图

1. 生成CPU火焰图  
   Async-profiler  
   pid=2449  
   ./profiler.sh -d 20 -f $pid.svg $pid  
   命令对$pid进行采样20S，生成CPU采样文件

arthas\_profiler  
java -jar /tmp/arthas/arthas-boo.jar  
(需要和进程同用户，需要选进程)  
profiler start -d 20 --file /tmp/igozhang-%t%p\_cpu.jfr  
采样20S，生成文件名带%t时间%p进程号

2. 生成MEM火焰图  
   Async-profiler  
   pid=2449  
   ./profiler.sh -d 20 -e alloc -f $pid-alloc.svg $pid  
   命令对$pid进行采样内存信息20S

arthas\_profiler  
java -jar /tmp/arthas/arthas-boo.jar  
(需要和进程同用户，需要选进程)  
profiler start -e alloc -d 20 --file /tmp/igozhang\_mem.jfr  
(可以直接打包成html格式，直接打开)

## 查看火焰图

火焰图里，横条越长，代表资源使用越多，从下到上是调用堆栈信息

1. CPU火焰图
2. MEM火焰图

SVG直接chrome打开  
jfr需要找工具  
JDK Mission Control ： https://github.com/openjdk/jmc  
JProfiler ： https://github.com/alibaba/arthas/issues/1416
