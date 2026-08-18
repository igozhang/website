# jda_mod;JDA优化

```
 Blue Yonder Platform Version: 2019.1.2.0
```

```
jda_console(7001,7101,7201)
Administration>ProcessManager>NodePoolSummary:
PoolName:MP
NumberofNodesDefined:4/2(多个SREServer就填2否则选4)
NodeJavaOptions: -Xms30480m -Xmx30480m -XX:+HeapDumpOnOutOfMemoryError(xms,xmx一致)
Associated Processes:(Assigned Processes:只需保留5个比较占大内存的,其余的放basic就行,避免资源浪费)
Assigned Processes:
Generate Master Plan (MAP) (DEFAULT).
Generate Master Plan (LpOpt) (DEFAULT).
Level SKUs (DEFAULT).
Level SKUs with search (DEFAULT).
Transfer Calendar (DEFAULT).

主频高的CPU提高计算速度;
SSD硬盘提高落盘时间；
服务器日志清理脚本统一；
服务器日志清理数据库表；
大表做分区，分析目前引擎过程使用的大表，并进行分区改造测试；
Transfer Calendar 节点是否可以下线；
Level SKU这个SRE Process性能优化；
Purge Plan：提前删除结果数据 [代码调整]
Purge Plan：启用Parallelize Prev Plan Purge功能 [配置]
Purge Plan：提高COMMITFREQUENCY [配置]
Purge Plan：大表改造为分区表DROP分区
Import：过滤掉不需要的记录 
Import：SQL优化，比如加索引；
Plan：SRE服务器使用更高主频的CPU (可与APP分离)
Plan：优化ESP数据模型，比如给替换料增加一个dummy层  [代码调整]
Export：只输出需要的结果数据或文件；
Export：启用多线程导出 [配置]
Sqlldr：配置使用直接加载模式
Sqlldr：配置sqlldr参数，提高rows和readsize等，提高普通加载模式下的sqlldr性能 [代码调整]
Sqlldr：并行运行sqlldr同时加载多个文件的记录 [代码调整]
Sqlldr： 多个文件的记录再平衡 [代码调整]
如果没有启Batch Number功能，GMP的max node设为1, MP对应的Node数量可以设为2 [配置]
配置更多的物理内存，提高JVM的最小内存值(内存富裕的话可以等于max值)；确保应用服务器不使用Swap交换空间
减少日志输出，日志级别调整为ERROR级别
Node Pool定期重启策略制定;
```
