# java_boot,java启动参数

```
xms,xmx

测试取值参考方法如下：by东方龙马_邓钊星
1、设定好压测场景；
2、预估一个JVM Heap值；
3、如果是因为GC导致的性能问题，对GC进行turning,达到一个平衡值；--->取这个值，包括Heap设置，以及GC策略

-Xms2048m -Xmx2048m
```

```
pemsize,MetaspaceSize
are used to set size for Permanent Generation;
The permanent space is where the classes, methods, internalized strings, and similar objects used by the VM are stored and never deallocated;

java6
-XX:PermSize=1024M -XX:MaxPermSize=4096M
java8-java11
-XX:MetaspaceSize=200m -XX:MaxMetaspaceSize=256m;
网友和经验:-XX:MetaspaceSize=256M \  //大部分项目256够用,减少GC次数
```

```
gc_log

java8
-Xloggc:/data/logs/igo_gc.log -XX:+PrintGCDetails -XX:+PrintGCDateStamps \
java11
-Xlog:gc*:file=/data/logs/igo_gc.log:time,uptime:filecount=10,filesize=10m \
```

```
OOM_log

-XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/data/dump/
```

```
rocketmq_log
-Drocketmq.client.logRoot=/data/logs/rocketmq/app_r.log -Drocketmq.client.logLevel=warn \
```

```
Spring Boot HikariCP / JDBC
-Dmanagement.endpoint.health.group.readiness.includ="readinessState,db" \
-Dmanagement.endpoint.health.group.liveness.include="livenessStateProbeIndicator,diskSpace,refreshScope" \
-Dmanagement.endpoints.web.exposure.include="health,prometheus,metrics,loggers" \
-Dmanagement.endpoints.metrics.enabled=true \
-Dmanagement.health.probes.enabled=true \
-Dmanagement.metrics.export.prometheus.enabled=true \
加上重启后就可以看到数据了,然后添加grafana对应模板展示
http://10.18.18.18:8080/igoproject/actuator/prometheus
```

```
-XX:G1HeapRegionSize=4m
实测
在Centos7.6  openjdk11中大幅减少了YoungGC频率，从48S一次到12min一次；
可以使用java -XX:+PrintFlagsFinal -XX:G1HeapRegionSize=2m --version | grep -i G1HeapRegionSize
确认初始值；
```
