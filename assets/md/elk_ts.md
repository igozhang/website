# ELK_troubleShooting

```
elk 7.9.3
```

```
es
issue01:
master_not_discovered_exception
curl "10.21.81.31:29200/_cluster/health"
{"error":{"root_cause":[{"type":"master_not_discovered_exception","reason":null}],"type":"master_not_discovered_exception","reason":null},"status":503}

solution
至少在集群一个节点中指定至少一个具备选举资格的节点用于集群初始化,可以用命令,也可以在集群配置中指定
$ bin/elasticsearch -Ecluster.initial_master_nodes=node01   #(命令引导集群)
cluster.initial_master_nodes: node01                       #(配置引导集群)
https://www.elastic.co/guide/en/elasticsearch/reference/7.9/modules-discovery-bootstrap-cluster.html
```

```
logstash
issue01:
logstash启动后没有监听端口
/data/logstash/logs/logstash-plain.log  ExceptionInInitializerError

solution
从日志及搜索判断可能和jdk有关,替换jdk后解决
原版本# java -version
openjdk version "11-ea" 2018-09-25
新版本1.8
yum -y install java-1.8.0-openjdk
```

```
filebeat
issue01:
Feb 10 17:27:17 igocent79 systemd[1]: Failed to start Filebeat sends log files to Logstash or directly to Elasticsearch

检查配置
/usr/share/filebeat/bin/filebeat  test config --path.config /etc/filebeat
Exiting: error unpacking config data: more than one namespace configured accessing 'output' (source:'/etc/filebeat/filebeat.yml')

solution:
修复配置,去掉部分output解决问题
```

```
问题：版本filebeat7.9.3，es7.17.2
修改filebeat配置，直接输出到es后，启动报错

解决：
配置索引名后，还需要配置setup三行才可以，否则启动filebeat失败
cat /etc/filebeat/filebeat.yml
filebeat.inputs:
- type: log
  paths:
    - /var/log/nginx/access.log
  fields:
    log_topics: ng_acc
    serv_ip: 192.168.3.222

output.elasticsearch:
  hosts: ["localhost:9200"]
  index: "nginx-acc%{+yyyy.MM.dd}"

setup.template.name: "default@template"
setup.template.pattern: "nginx-*"
setup.ilm.enabled: false
```

```
env:
centos7.6  filebeat7.9.3

问题：JAVA进程意外被杀,oom_kill
Jul 17 19:56:52 cndgdlbvdc08-127-113 kernel: Out of memory: Kill process 370563 (filebeat) score 494 or sacrifice child
Jul 17 19:56:52 cndgdlbvdc08-127-113 kernel: Killed process 370563 (filebeat) total-vm:23919652kB, anon-rss:20952540kB, file-rss:0kB, shmem-rss:0kB
...skipping...
Jul 17 20:32:42 cndgdlbvdc08-127-113 kernel: Out of memory: Kill process 78064 (java) score 380 or sacrifice child
Jul 17 20:32:42 cndgdlbvdc08-127-113 kernel: Killed process 78064 (java) total-vm:25210748kB, anon-rss:9778272kB, file-rss:0kB, shmem-rss:0kB

原因： Filebeat吃太多内存20952540kB（约20G）触发oom,然后JAVA进程也被OOM_kill
分析： 目前程序OUT（console）日志输出量巨大，到一天6G，其中部分为方便开发排错日志部分，暂未拆走；
解决： 
vim /etc/filebeat/filebeat.yml 追加再文件末尾，output.logstash后面
queue.mem:
  events: 1024 
Number of events the queue can store.The default value is 4096 events.

相关命令：
ps aux | sort -k4,4nr | head -n 10  ```查内存占用前十程序
journalctl -p err..emerg  --since "yesterday"
cat /proc/$pid/status

相关字段：
VmHWM:    表示进程所占用物理内存的峰值
VmRSS:    表示进程当前占用物理内存的大小(与procrank中的RSS)

OOM 杀进程算法：
The function select_bad_process() is responsible for choosing a process to kill. It decides by stepping through each running task and calculating how suitable it is for killing with the function badness(). The badness is calculated as follows, note that the square roots are integer approximations calculated with int_sqrt();

badness_for_task = total_vm_for_task / (sqrt(cpu_time_in_seconds) *
sqrt(sqrt(cpu_time_in_minutes)))
This has been chosen to select a process that is using a large amount of memory but is not that long lived. Processes which have been running a long time are unlikely to be the cause of memory shortage so this calculation is likely to select a process that uses a lot of memory but has not been running long. If the process is a root process or has CAP_SYS_ADMIN capabilities, the points are divided by four as it is assumed that root privilege processes are well behaved. Similarly, if it has CAP_SYS_RAWIO capabilities (access to raw devices) privileges, the points are further divided by 4 as it is undesirable to kill a process that has direct access to hardware.
```

```
filebeat启动失败
查看日志：
journalctl -u filebeat.service -f -n 50
或者：
tail -200 /var/log/filebeat/filebeat
tail -200 /var/log/filebeat/filebeat.log

Can only start an input when all related states are finished（无法启动输入，因为关联的日志文件状态未标记为完成），同时伴随 /home/deploy/g-iplan-sop.log 文件截断（truncated）和注册表停止的报错，这是 Filebeat 7.x 版本常见的注册表状态紊乱问题

删注册表
rm -rf /var/lib/filebeat/registry/*
# 解除systemd的启动频率限制
systemctl reset-failed filebeat
# 重启服务
systemctl start filebeat
```
