# redis排错

```
redis_troubleshooting,redis_ts
```

```
redis505
问题:aof数据导入报错
[root@cnabdabvdc03-81-29 ~]# redis-cli -p 8001 --pipe </tmp/8007.aof
All data transferred. Waiting for the last reply...
ERR unknown command `REDIS0009򪟷ith args beginning with: `redis-ver5.0.5
Xshell
^C

说明:
AOF files, when containing the RDB preamble, can't be just piped through the cli to the server - you'll have to place the AOF in the server's directory and let it load it during startup.
aof导入的时候如果包含rdb序言,就不能使用pipe加载,需要直接覆盖

解决:拷贝解决
cp 8007.aof /data/redisB/8001/8001.aof
redis-cli -p 8001 shutdown
redis-server 8001.conf
```
