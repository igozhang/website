# redisinfo_sh

```
#!/bin/bash

# 定义输出文件路径
LOG_FILE="/tmp/redis$(date +%Y%m%d%H%M).log"

function collect_info () {
# 清空日志文件
> $LOG_FILE

# 写入基本信息到日志文件
echo "Redis Cluster Information - $(date)" >> $LOG_FILE

# 收集并写入 INFO 命令的数据
redis-cli -a $_pwd -p $_port info server >> $LOG_FILE
redis-cli -a $_pwd -p $_port info clients >> $LOG_FILE
redis-cli -a $_pwd -p $_port info memory >> $LOG_FILE
redis-cli -a $_pwd -p $_port info persistence >> $LOG_FILE
redis-cli -a $_pwd -p $_port info stats >> $LOG_FILE
redis-cli -a $_pwd -p $_port info replication >> $LOG_FILE
redis-cli -a $_pwd -p $_port info cpu >> $LOG_FILE
redis-cli -a $_pwd -p $_port info commandstats >> $LOG_FILE
redis-cli -a $_pwd -p $_port info cluster >> $LOG_FILE

# 收集并写入 CLIENT LIST 数据
redis-cli -a $_pwd -p $_port client list >> $LOG_FILE

# 收集并写入 CONFIG GET 数据
redis-cli -a $_pwd -p $_port config get maxmemory >> $LOG_FILE
redis-cli -a $_pwd -p $_port config get maxclients >> $LOG_FILE
redis-cli -a $_pwd -p $_port config get timeout >> $LOG_FILE

# 最终输出文件信息
echo "Log file saved to: $LOG_FILE"
}

main () {
  _pwd=$1
  _port=$2
  if [ $# -lt 1 ]; then
     echo -e  "\e[41m you must put 1 argument [PID] at least \e[0m"
     exit
  else
    collect_info $@
 fi
}

main $@

# exp: ./redisinfo.sh PasswordIgo 7001
```

```
输入redis集群密码，当前节点端口号，获取redis集群状态信息
```
