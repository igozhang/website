# 问题集

```
优点 有相对完整技术栈，快速学习能力
缺点 编程这块可以加强补短板
对核心概念的深刻理解，技术不求精求快速学习快速响应快速解决问题，保持业务稳定；掌握工具通用性能力及差异特性；
```

```
1.GC,垃圾回收

2.Redis,分布式锁

3.redis_ACID

4.乐观锁，悲观锁

5.redis_slot迁移中收到请求处理

6.数据库锁表机制

7.kafka mq 选择考虑

8.nginx服务检测

9.zk选举机制

10.zk kafka redis es mysql 等组件数据迁移
```

```
1.GC_垃圾回收
比较普及的CMS 和 G1 收集器
CMS（Concurrent Mark Sweep）基于“标记—清除”算法，关注点在于减少“pause time”也即因垃圾回收导致的stop the world时间
初始标记：标记一下GC Roots能直接关联到的对象，会“Stop The World”
并发标记：GC Roots Tracing，可以和用户线程并发执行。
重新标记：标记期间产生的对象存活的再次判断，修正对这些对象的标记，执行时间相对并发标记短，会“Stop The World”。
并发清除：清除对象,可以和用户线程并发执行。

G1垃圾收集器
G1（Garbage-First）是在JDK 7u4版本之后发布的垃圾收集器，并在jdk9中成为默认垃圾收集器
G1也是利用多CPU来缩短stop the world时间，并且是高效的并发垃圾收集器

https://tech.meituan.com/2020/11/12/java-9-cms-gc.html
https://cloud.tencent.com/developer/article/1769458
```

```
2.redis分布式锁
锁的存在是为了解决分布式系统不同操作对同一个临界资源的竞争
锁的特性有
互斥：锁的基本特征，同一时刻，只能有一个进程(线程)持有锁
超时释放：超过一定时间需要释放锁，防止出现异常导致锁无法释放进而引起不可预知的其他问题
高性能与高可用：加锁和释放锁应该低开销，且尽可能防止锁意外失效

https://juejin.cn/post/6995850152690843655
https://zhuanlan.zhihu.com/p/60196818
```

```
3.redis_ACID
ACID 原子性、一致性、隔离性、持久性
原子性就是一组操作动作命令的独立性和整体性，不可分割
https://juejin.cn/post/6970531814653820942
MULTI 显式的开始一个事务
事务包含的具体操作命令
EXEC  开始执行事务
```

```
4.乐观锁，悲观锁
悲观锁一般由数据库实现，利用数据库的加锁机制
在对记录进行修改前，先尝试为该记录加上排他锁（exclusive locking）。
如果加锁失败，说明该记录正在被修改，那么当前查询可能要等待或者抛出异常。具体响应方式由开发者根据实际需要决定。
如果成功加锁，那么就可以对记录做修改，事务完成后就会解锁了。
其间如果有其他事务对该记录做加锁的操作，都要等待当前事务解锁或直接抛出异常。

乐观锁一般无需借助数据库的加锁机制
主要就是两个步骤：冲突检测和数据更新。其实现方式有一种比较典型的就是Compare and Swap(CAS)技术
CAS是项乐观锁技术，当多个线程尝试使用CAS同时更新同一个变量时，只有其中一个线程能更新变量的值，而其它线程都失败，失败的线程并不会被挂起，而是被告知这次竞争中失败，并可以再次尝试。
CAS的实现中，在表中增加一个version字段，操作前先查询version信息，在数据提交时检查version字段是否被修改，如果没有被修改则进行提交，否则认为是过期数据。
```

```
5.redis_slot迁移
redis-cluster在迁移一个slot的时候具体流程如下：
1.对目标节点发送 cluster setslot <slot> importing <sourceNodeId> 命令，让目标节点准备导入槽的数据。
2.对源节点发送  cluster setslot <slot> migrating <targetNodeId> 命令，让源节点准备迁出槽的数据。
3.源节点循环执行 cluster getkeysinslot <slot> <count> 命令，获取count个属于槽slot的键。
4.在源节点上执行 migrate <targetIp> <targetPort> "" 0 <timeout> keys <keys...> 命令，把获取的键通过流水（pipeline）机制批量迁移到目标节点。
5.重复执行步骤3和步骤4直到槽下所有的键值数据迁移到目标节点。
6.向集群内所有主节点发送cluster setslot <slot> node  <targetNodeId>命令，通知槽分配给目标节点。为了保证槽节点映射变更及时传播，需要遍历发送给所有主节点更新被迁移的槽指向新节点。
tips：如果中途想取消一个迁移，可以向节点发送 cluster setslot <slot> stable 取消对槽 slot 的导入（import）或者迁移（migrate）状态。

迁移中的读写：
migrate命令是同步阻塞的，因此不会存在一个key正在被迁移又同时被读写的情况，但由于一个slot下可能有部分key被迁移完成，部分key正在等待迁移的情况，因此如果读写的key所属的slot正在被迁移，redis-cluster做如下处理：
客户端根据本地slots缓存发送命令到源节点，如果存在键对象则直接指向并返回结果给客户端。
如果key对象不存在，但key所在的slot属于本节点，则可能存在于目标节点，这时源节点会回复ASK重定向异常。例如如：error）ASK :。
客户端从ASK重定向异常提取出目标节点的信息，发送asking命令到目标节点打开客户端连接标识，再执行key命令。如果存在则执行，不存在则返回不存在信息。
如果key所在的slot不属于本节点，则返回MOVE重定向。格式如下：（error）MOVED :
```

```
6.数据库锁表机制
mysql举例
数据库锁定机制简单的来说，就是数据库为了保证数据的一致性与完整性，而使各种共享资源在被并发访问时变得有序所设计的一种规则。 对于任何一种数据库来说都需要有相应的锁机制，所以MySQL也不能例外。 MySQL数据库根据锁锁定数据的颗粒度可分为表级锁、行级锁和页级锁
简单分类：
共享锁（读锁）：其他事务可以读，但不能写。
排他锁（写锁）：其他事务不能读取，也不能写。
也可以分为：
全局锁
对整个数据库实例加锁。当你需要让整个库处于只读状态的时候 
表锁
lock tables … read/write，可以用 unlock tables 主动释放锁
也可以在客户端断开的时候自动释放
行锁
MySQL 的行锁是在引擎层由各个引擎自己实现的。但并不是所有的引擎都支持行锁
```

```
7.mq,kafka对比
Kafka适合日志处理；
RocketMQ适合业务处理;

kafka单机性能更好
RocketMQ可靠性更好

只做参考，有待验证
https://cloud.tencent.com/developer/article/1706532
https://zhuanlan.zhihu.com/p/60196818
```

```
8.nginx服务检测

原生模块 ngx_http_proxy_module和ngx_http_upstream_module 支持tcp握手，支持超时
三方模块 nginx_upstream_check_module 可以发送tcp，还可以发送7层mysql,http,ssl包
```

```
9.zookeeper,zk选举机制
ZK集群的选举机制 zk会进行多轮的投票，直到某一个节点的票数大于或等于半数以上，在3个节点中，总共会进行2轮的投票： 
第一轮，每个节点启动时投票给自己，那这样zk1，zk2，zk3各有一票。 
第二轮，每个节点投票给大于自己myid，那这样zk2启动时又获得一票。加上自己给自己投的那一票。总共有2票。2票大于了当前节点总数的半数，所以投票终止。zk2当选leader。
```
