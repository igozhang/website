# redis4集群部署及管理

```
env
cent72
redis408
redis_installation
```

## 1：部署redis集群

### 1.1 概览

- 准备集群环境
- 安装redis并创建集群
- 查看集群信息

### 1.2 规划

搭建redis集群，拓扑规划如图-1所示：

图－1

IP，端口规划如表-1所示：

表-1

### 1.3 步骤

实现此案例需要按照如下步骤进行。

步骤一：准备集群环境

1）按照表-1配置主机名，ip地址，配置yum源（系统源）这里不再操作

2）把redis的软件包传到6台数据库服务器上面，安装redis服务器，六台服务器同样操作（以51为例）

1. [root@redisA ~]# yum -y install gcc gcc-c++ make
2. [root@redisA ~]# cd redis
3. redis/ redis-cluster/
4. [root@redisA ~]# cd redis/
5. [root@redisA redis]# ls
6. lnmp redis-4.0.8.tar.gz
7. [root@redisA redis]# tar -xf redis-4.0.8.tar.gz
8. [root@redisA redis]# cd redis-4.0.8/
9. [root@redisA redis-4.0.8]# make && make install
10. [root@redisA redis-4.0.8]# ./utils/install\_server.sh
11. Welcome to the redis service installer
12. This script will help you easily set up a running redis server
13. Please select the redis port for this instance: [6379]
14. Selecting default: 6379
15. Please select the redis config file name [/etc/redis/6379.conf]
16. Selected default - /etc/redis/6379.conf
17. Please select the redis log file name [/var/log/redis\_6379.log]
18. Selected default - /var/log/redis\_6379.log
19. Please select the data directory for this instance [/var/lib/redis/6379]
20. Selected default - /var/lib/redis/6379
21. Please select the redis executable path [/usr/local/bin/redis-server]
22. Selected config:
23. Port : 6379
24. Config file : /etc/redis/6379.conf
25. Log file : /var/log/redis\_6379.log
26. Data dir : /var/lib/redis/6379
27. Executable : /usr/local/bin/redis-server
28. Cli Executable : /usr/local/bin/redis-cli
29. Is this ok? Then press ENTER to go on or Ctrl-C to abort.
30. Copied /tmp/6379.conf => /etc/init.d/redis\_6379
31. Installing service...
32. Successfully added to chkconfig!
33. Successfully added to runlevels 345!
34. Starting Redis server...
35. Installation successful!            //安装成功
36. [root@redisA redis-4.0.8]# ss -antlp | grep 6379            //查看时有端口
37. LISTEN 0 128 127.0.0.1:6379 \*:\* users:(("redis-server",pid=10788,fd=6))

2）修改配置文件，6台redis服务器都要修改（以51为例子）

1. [root@redisA redis-4.0.8]# /etc/init.d/redis\_6379 stop
2. //停止已经开启的redis服务
3. Stopping ...
4. Waiting for Redis to shutdown ...
5. Redis stopped
6. [root@redisA redis-4.0.8]# vim /etc/redis/6379.conf
7. ...
8. bind 192.168.4.51        //修改ip
9. port 6351        //不允许相同，只指定物理接口的ip
10. daemonize yes         //以守护进程方式运行
11. pidfile /var/run/redis\_6351.pid
12. cluster-enabled yes     //是否启用集群，前提是以守护进程方式运行
13. cluster-config-file nodes-6351.conf
14. //存储集群信息的配置文件，自动生成，不允许相同
15. cluster-node-timeout 5000        //集群节点通信超时时间
16. ...
17. [root@redisA redis-4.0.8]# /etc/init.d/redis\_6379 start     //启动服务
18. Starting Redis server...
19. [root@redisA redis-4.0.8]# ss -antlp | grep 6351            //查看有端口
20. LISTEN 0 128 192.168.4.51:6351 \*:\* users:(("redis-server",pid=11092,fd=6))
21. LISTEN 0 128 192.168.4.51:16351 \*:\* users:(("redis-server",pid=11092,fd=8))        //16051：集群中的主机通信时用的端口
22. [root@redisA redis-4.0.8]# ps -C redis
23. PID TTY TIME CMD

注意：其他几台主机在修改时请注意ip，端口等的修改，不要和51主机的一样

3）关闭防火墙51-56主机（以51为例子）

1. [root@redisA redis-4.0.8]# getenforce
2. Permissive
3. [root@redisA redis-4.0.8]# systemctl disable firewalld
4. //关闭防火墙不自启

4）查看集群信息

1. [root@redisA redis-4.0.8]# redis-cli -h 192.168.4.51 -p 6351
2. 192.168.4.51:6351> ping
3. PONG
4. 192.168.4.51:6351> cluster info
5. cluster\_state:fail
6. cluster\_slots\_assigned:0
7. cluster\_slots\_ok:0
8. cluster\_slots\_pfail:0
9. cluster\_slots\_fail:0
10. cluster\_known\_nodes:1
11. cluster\_size:0
12. ...
13. 192.168.4.51:6351> cluster nodes
14. f81f997d5ed988ec1587558e78d5f7dbc96abcbf :6351@16351 myself,master - 0 0 0 connected

步骤二：创建集群（在任意一台上执行创建集群的脚本都可以）这里在51上面执行

1）部署ruby脚本运行环境（在51上面执行）

1. [root@redisA redis-4.0.8]# cd /root/redis-cluster/
2. [root@redisA redis-cluster]# ls
3. redis-3.2.1.gem ruby-devel-2.0.0.648-30.el7.x86\_64.rpm
4. [root@redisA redis-cluster]# yum -y install ruby rubygems
5. [root@redisA redis-cluster]# rpm -ivh –nodeps \
6. ruby-devel-2.0.0.648-30.el7.x86\_64.rpm
7. warning: ruby-devel-2.0.0.648-30.el7.x86\_64.rpm: Header V3 RSA/SHA256 Signature, key ID f4a80eb5: NOKEY
8. Preparing... ################################# [100%]
9. Updating / installing...
10. 1:ruby-devel-2.0.0.648-30.el7 ################################# [100%]
11. [root@redisA redis-cluster]# which gem
12. /usr/bin/gem
13. [root@redisA redis-cluster]# gem install redis
14. Successfully installed redis-3.2.1
15. Parsing documentation for redis-3.2.1
16. Installing ri documentation for redis-3.2.1
17. 1 gem installed

2）生成创建集群的脚本

1. [root@redisA redis-cluster]# cd /root/redis/redis-4.0.8/src/
2. [root@redisA src]# cp redis-trib.rb /usr/local/bin/
3. [root@redisA src]# ll /usr/local/bin/redis-trib.rb
4. -rwxr-xr-x. 1 root root 65991 Sep 27 16:12 /usr/local/bin/redis-trib.rb

3）创建集群

1. [root@redisA src]# redis-trib.rb create --replicas 1 \
2. 192.168.4.51:6351 192.168.4.52:6352 \
3. 192.168.4.53:6353 192.168.4.54:6354 \
4. 192.168.4.55:6355 192.168.4.56:6356
5. //--replicas 1 给每一个主配置一个从库
6. [root@redisA log]# redis-trib.rb create --replicas 1 192.168.4.51:6351 192.168.4.52:6352 192.168.4.53:6353 192.168.4.54:6354 192.168.4.55:6355 192.168.4.56:6356
7. >>> Creating cluster
8. >>> Performing hash slots allocation on 6 nodes...
9. Using 3 masters:
10. 192.168.4.51:6351
11. 192.168.4.52:6352
12. 192.168.4.53:6353
13. Adding replica 192.168.4.55:6355 to 192.168.4.51:6351
14. Adding replica 192.168.4.56:6356 to 192.168.4.52:6352
15. Adding replica 192.168.4.54:6354 to 192.168.4.53:6353
16. ...
17. ...
18. ...
19. [OK] All nodes agree about slots configuration.
20. >>> Check for open slots...
21. >>> Check slots coverage...
22. [OK] All 16384 slots covered.

4）查看集群信息，任意一台主机访问本机的redis服务查看即可

cluster info 查看集群信息

cluster nodes 查看集群节点信息

1. [root@redisA log]# redis-cli -h 192.168.4.52 -p 6352
2. 192.168.4.52:6352> CLUSTER INFO
3. cluster\_state:ok            //状态
4. cluster\_slots\_assigned:16384
5. cluster\_slots\_ok:16384
6. cluster\_slots\_pfail:0
7. cluster\_slots\_fail:0
8. cluster\_known\_nodes:6
9. cluster\_size:3
10. cluster\_current\_epoch:6
11. cluster\_my\_epoch:2
12. cluster\_stats\_messages\_ping\_sent:367
13. cluster\_stats\_messages\_pong\_sent:327
14. cluster\_stats\_messages\_meet\_sent:5
15. cluster\_stats\_messages\_sent:699
16. cluster\_stats\_messages\_ping\_received:327
17. cluster\_stats\_messages\_pong\_received:372
18. cluster\_stats\_messages\_received:699
19. 192.168.4.52:6352> CLUSTER NODES        //查看集群节点信息
20. 63afbb5e7d63b1f142358634578a3488e3c9e634 192.168.4.54:6354@16354 slave bc5c4e082a5a3391b634cf433a6486c867cfc44b 0 1538039278871 4 connected
21. bc5c4e082a5a3391b634cf433a6486c867cfc44b 192.168.4.53:6353@16353 master - 0 1538039278571 3 connected 10923-16383
22. 28e06c5f24a2b6c6412f81369e09bc9653cc51ff 192.168.4.56:6356@16356 slave 8568fbd73cb296cad6915d524e34761b2114af47 0 1538039278069 6 connected
23. 7e8d9121f44d8331ff55b45c218b87df9bda1b70 192.168.4.55:6355@16355 slave a3083123bc5c87a76aab2655171634d4ee84f418 0 1538039278000 5 connected
24. 8568fbd73cb296cad6915d524e34761b2114af47 192.168.4.52:6352@16352 myself,master - 0 1538039277000 2 connected 5461-10922
25. a3083123bc5c87a76aab2655171634d4ee84f418 192.168.4.51:6351@16351 master - 0 1538039277869 1 connected 0-5460
26. 192.168.4.52:6352>

5）测试集群

命令：

redis-cli -c -h ip地址 -p 端口

1. [root@redisA log]# redis-cli -c -h 192.168.4.51 -p 6351
2. 192.168.4.51:6351> set name jim
3. -> Redirected to slot [5798] located at 192.168.4.52:6352
4. OK
5. 192.168.4.52:6352> get name
6. "jim"
7. 192.168.4.52:6352> set class linux
8. OK
9. 192.168.4.52:6352> get class
10. "linux"
11. 192.168.4.52:6352> set pay 26800
12. -> Redirected to slot [4013] located at 192.168.4.51:6351
13. OK
14. 192.168.4.51:6351> get pay
15. "26800"

集群不能用的情况：

有半数或者半数以上的主库机器挂掉，集群就不能用了

把一个从库升级成主，没有从库，集群不能用（前提是：有半数或者半数以上的主库机器挂掉）

一个主库挂掉，它的从库自动顶替为主库，正常使用（前提是：有半数或者半数以上的主库机器能用），挂掉的主库修复好后，会成为从库，不会抢占为主

6）集群节点选举策略（三主，三从）

停止某个主库的redis服务，对应的从库会自动升级为主库

先查看节点信息的主从情况

1. [root@redisA log]# redis-cli -c -h 192.168.4.51 -p 6351
2. 192.168.4.51:6351> CLUSTER nodes
3. ...
4. 8568fbd73cb296cad6915d524e34761b2114af47 192.168.4.52:6352@16352 master - 0 1538040400840 2 connected 5461-10922
5. 28e06c5f24a2b6c6412f81369e09bc9653cc51ff 192.168.4.56:6356@16356 slave 8568fbd73cb296cad6915d524e34761b2114af47 0 1538040400000 6 connected
6. ...
7. 192.168.4.51:6351>

看谁是谁的从库，如：

看节点前后的编号id是否有相同的

如：8568fbd73cb296cad6915d524e34761b2114af47

发现52的从库为56

停止主库52

1. [root@redisA log]# redis-cli -h 192.168.4.52 -p 6352 shutdown
2. [root@redisA log]# redis-cli -c -h 192.168.4.51 -p 6351
3. 192.168.4.51:6351> CLUSTER NODES
4. ...
5. 8568fbd73cb296cad6915d524e34761b2114af47 192.168.4.52:6352@16352 master,fail - 1538041052349 1538041051000 2 disconnected        //52的主库坏掉
6. 28e06c5f24a2b6c6412f81369e09bc9653cc51ff 192.168.4.56:6356@16356 master - 0 1538041066000 7 connected 5461-10922        //56成为主库
7. ...

开启52，发现52成为从库

1. [root@redisB redis-4.0.8]# /etc/init.d/redis\_6352 start
2. Starting Redis server...
3. [root@redisA log]# redis-cli -c -h 192.168.4.51 -p 6351
4. 192.168.4.51:6351> CLUSTER NODES
5. 8568fbd73cb296cad6915d524e34761b2114af47 192.168.4.52:6352@16352 slave 28e06c5f24a2b6c6412f81369e09bc9653cc51ff 0 1538041254000 7 connected

## 2：管理redis集群

### 2.1 需求

- 如下：
- 添加主机
- 删除主机

### 2.2 步骤

实现此案例需要按照如下步骤进行。

步骤一：添加主机

1）部署一台新redis服务器，ip为192.168.4.58，装包，初始化，启用集群配置，重启服务（这里之前已经操作过，不会的可以参考案例1）

2） 添加集群4.58（添加master节点）

格式：redis-trib.rb 选项 参数

选项： add-nade 添加主机（不指定角色为主）

由于之前是在51上面创建ruby脚本，所以只有51上面有redis-trib.rb命令，在51上面执行

1. [root@redisA ~]# redis-trib.rb add-node 192.168.4.58:6358 192.168.4.51:6351
2. >>> Adding node 192.168.4.58:6358 to cluster 192.168.4.51:6351
3. >>> Performing Cluster Check (using node 192.168.4.51:6351)
4. S: a3083123bc5c87a76aab2655171634d4ee84f418 192.168.4.51:6351
5. slots: (0 slots) slave
6. replicates 7e8d9121f44d8331ff55b45c218b87df9bda1b70
7. M: 7e8d9121f44d8331ff55b45c218b87df9bda1b70 192.168.4.55:6355
8. slots:0-5460 (5461 slots) master
9. 1 additional replica(s)
10. S: 8568fbd73cb296cad6915d524e34761b2114af47 192.168.4.52:6352
11. slots: (0 slots) slave
12. replicates 28e06c5f24a2b6c6412f81369e09bc9653cc51ff
13. M: bc5c4e082a5a3391b634cf433a6486c867cfc44b 192.168.4.53:6353
14. slots:10923-16383 (5461 slots) master
15. 1 additional replica(s)
16. S: 63afbb5e7d63b1f142358634578a3488e3c9e634 192.168.4.54:6354
17. slots: (0 slots) slave
18. replicates bc5c4e082a5a3391b634cf433a6486c867cfc44b
19. M: 28e06c5f24a2b6c6412f81369e09bc9653cc51ff 192.168.4.56:6356
20. slots:5461-10922 (5462 slots) master
21. 1 additional replica(s)
22. [OK] All nodes agree about slots configuration.
23. >>> Check for open slots...
24. >>> Check slots coverage...
25. [OK] All 16384 slots covered.
26. >>> Send CLUSTER MEET to node 192.168.4.58:6358 to make it join the cluster.
27. [OK] New node added correctly.

3）检查集群主机的状态信息

选项：check 检查权限

1. [root@redisA ~]# redis-trib.rb check 192.168.4.58:6358        //查看状态
2. >>> Performing Cluster Check (using node 192.168.4.58:6358)
3. M: c5e0da48f335c46a2ec199faa99b830f537dd8a0 192.168.4.58:6358
4. slots: (0 slots) master        //发现没有hash槽
5. 0 additional replica(s)
6. M: 7e8d9121f44d8331ff55b45c218b87df9bda1b70 192.168.4.55:6355
7. slots:0-5460 (5461 slots) master
8. 1 additional replica(s)
9. ...
10. S: a3083123bc5c87a76aab2655171634d4ee84f418 192.168.4.51:6351
11. slots: (0 slots) slave
12. replicates 7e8d9121f44d8331ff55b45c218b87df9bda1b70
13. [OK] All nodes agree about slots configuration.
14. >>> Check for open slots...
15. >>> Check slots coverage...
16. [OK] All 16384 slots covered.

4）手动对集群进行分片迁移

选项：reshard 重新分配hash槽

1. [root@redisA ~]# redis-trib.rb reshard 192.168.4.58:6358
2. How many slots do you want to move (from 1 to 16384)?4096
3. //拿出多少个hash 槽给主机192.168.4.58
4. What is the receiving node ID? c5e0da48f335c46a2ec199faa99b830f537dd8a0
5. //主机192.168.4.58的id值
6. Source node #1:all //从当前所有的主里面获取hash槽
7. Do you want to proceed with the proposed reshard plan (yes/no)?yes
8. ...
9. Moving slot 12283 from 192.168.4.53:6353 to 192.168.4.58:6358:
10. Moving slot 12284 from 192.168.4.53:6353 to 192.168.4.58:6358:
11. Moving slot 12285 from 192.168.4.53:6353 to 192.168.4.58:6358:
12. Moving slot 12286 from 192.168.4.53:6353 to 192.168.4.58:6358:
13. Moving slot 12287 from 192.168.4.53:6353 to 192.168.4.58:6358:

再次查看发现4.58有4096个hash slot

1. [root@redisA ~]# redis-trib.rb check 192.168.4.58:6358
2. >>> Performing Cluster Check (using node 192.168.4.58:6358)
3. M: c5e0da48f335c46a2ec199faa99b830f537dd8a0 192.168.4.58:6358
4. slots:0-1364,5461-6826,10923-12287 (4096 slots) master
5. 0 additional replica(s)

5）删除master角色的主机

先删除主机占用的hash槽

1. [root@redisA ~]# redis-trib.rb reshard 192.168.4.58:6358
2. How many slots do you want to move (from 1 to 16384)?4096
3. //移除hash 槽的个数
4. What is the receiving node ID? bc5c4e082a5a3391b634cf433a6486c867cfc44b
5. //要移动给谁的id即目标主机（这里可以随机写一个master的ID）
6. Source node #1: c5e0da48f335c46a2ec199faa99b830f537dd8a0
7. //从谁那移动即源主机（这里写4.58的ID）
8. Source node #2:done         //设置完毕
9. ...
10. Moving slot 12282 from c5e0da48f335c46a2ec199faa99b830f537dd8a0
11. Moving slot 12283 from c5e0da48f335c46a2ec199faa99b830f537dd8a0
12. Moving slot 12284 from c5e0da48f335c46a2ec199faa99b830f537dd8a0
13. Moving slot 12285 from c5e0da48f335c46a2ec199faa99b830f537dd8a0
14. Moving slot 12286 from c5e0da48f335c46a2ec199faa99b830f537dd8a0
15. Moving slot 12287 from c5e0da48f335c46a2ec199faa99b830f537dd8a0
16. Do you want to proceed with the proposed reshard plan (yes/no)?yes        //提交
17. ...
18. Moving slot 12282 from 192.168.4.58:6358 to 192.168.4.53:6353:
19. Moving slot 12283 from 192.168.4.58:6358 to 192.168.4.53:6353:
20. Moving slot 12284 from 192.168.4.58:6358 to 192.168.4.53:6353:
21. Moving slot 12285 from 192.168.4.58:6358 to 192.168.4.53:6353:
22. Moving slot 12286 from 192.168.4.58:6358 to 192.168.4.53:6353:
23. Moving slot 12287 from 192.168.4.58:6358 to 192.168.4.53:6353:

删除集群主机4.58(删除之后redis服务自动关闭)

1. [root@redisA ~]# redis-trib.rb del-node 192.168.4.58:6358 \
2. c5e0da48f335c46a2ec199faa99b830f537dd8a0    //删除谁+删除的id
3. >>> Removing node c5e0da48f335c46a2ec199faa99b830f537dd8a0 from cluster 192.168.4.58:6358
4. >>> Sending CLUSTER FORGET messages to the cluster...
5. >>> SHUTDOWN the node.

6）添加从节点主机，随机添加

1. [root@redisA ~]# redis-trib.rb add-node --slave \
2. 192.168.4.57:6357 192.168.4.51:6351
3. >>> Adding node 192.168.4.57:6357 to cluster 192.168.4.51:6351
4. >>> Performing Cluster Check (using node 192.168.4.51:6351)
5. ……
6. ……
7. [OK] All 16384 slots covered.
8. Automatically selected master 192.168.4.51:6351
9. >>> Send CLUSTER MEET to node 192.168.4.57:6357 to make it join the cluster.
10. Waiting for the cluster to join.
11. >>> Configure node as replica of 192.168.4.51:6351.
12. [OK] New node added correctly.

7）移除从节点，从节点主机没有槽位范围，直接移除即可

命令格式：

redis-trib.rb del-node 192.168.4.57:6357 主机id值

1. [root@redisA ~]# redis-trib.rb del-node 192.168.4.57:6357 \
2. f6649ea99b2f01faca26217691222c17a3854381
3. >>> Removing node f6649ea99b2f01faca26217691222c17a3854381
4. from cluster 192.168.4.57:6351
5. >>> Sending CLUSTER FORGET messages to the cluster...
6. >>> SHUTDOWN the node.

```
tips

redis-trib.rb是在redis3.x版本时所用的一种部署redis集群的工具，redis-cli是redis4.x及更高版本所支持创建集群的工具，在redis3.x版本时redis-cli只是一个客户端连接管理工具。
redis-cli比redis-trib.rb多了一个可以认证集群密码的功能，后者创建的集群不能对有密码的集群节点进行很好的管理，所以后来官方直接废弃了这个工具。
用redis-trib.rb创建集群之前需要配置ruby环境，新版本的redis-cli可以直接创建集群环境而不用配置ruby环境

exp
/opt/redis/bin/redis-cli --cluster create --cluster-replicas 1 172.28.0.224:6380 172.28.0.225:6380 172.31.0.226:6380 172.31.0.227:6380 172.31.0.224:6380 172.31.0.225:6380 172.28.0.226:6380 172.28.0.227:6380 -a '123456'
```
