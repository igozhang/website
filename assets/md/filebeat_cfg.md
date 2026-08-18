# filebeat_cfg

```
Cent79
filebeat7.9.3

https://www.elastic.co/guide/en/beats/filebeat/current/multiline-examples.html
```

```
调试命令
echo "message" | ./filebeat -c ~/data/filebeat_test/filebeat.yml -e 2> /dev/null
```

```
filebeat.yml
filebeat.inputs:
- type: log
  paths:
    - /var/log/nginx/access.log
  fields:
    log_topics: ng_acc_prdrz_platform
    tag: ng_acc
- type: log
  paths:
    - /var/log/nginx/error.log
  fields:
    log_topics: ng_err_prdrz_platform
    tag: ng_err

output.logstash:
  hosts: ["192.168.189.88:5044"]
```

```
聚合multiline
日志样例01-java_jar:
2022-03-29 00:55:39.660  INFO [,,] 24842 --- [AppConfigRefresh-1] c.o.t.f.s.c.impl.FSAppConfigService      : refresh app config success! config cnt: 3

multiline.type: pattern
multiline.pattern: '^\[[0-9]{4}-[0-9]{2}-[0-9]{2}'
multiline.negate: true
multiline.match: after

日志样例02-oracle:
Tue Mar 22 11:05:49 2022
Thread 1 advanced to log sequence 1406 (LGWR switch)
  Current log# 8 seq# 1406 mem# 0: /data/oracle/oradata/asysit/onlinelog/redo8.log

  multiline.type: pattern
  multiline.pattern: '\w{3}\s\w{3}\s\d{2}\s\d{2}\:\d{2}\:\d{2}\s\d{4}'
  multiline.negate: true
  multiline.match: after

日志样例03-oracle:
<msg time='2022-03-22T15:09:10.882+08:00' org_id='oracle' comp_id='tnslsnr'
 type='UNKNOWN' level='16' host_id='igok8s_node_02'
 host_addr='10.21.120.236'>
 <txt>22-MAR-2022 15:09:10 * service_update * asysit * 0
 </txt>
</msg>

  multiline.type: pattern
  multiline.pattern: '^\<msg\ time\='
  multiline.negate: true
  multiline.match: after
  
日志样例04:
[beat-logstash-some-name-832-2015.11.28] IndexNotFoundException[no such index]
   at org.elasticsearch.cluster.metadata.IndexNameExpressionResolver$WildcardExpressionResolver.resolve(IndexNameExpressionResolver.java:566)
   at org.elasticsearch.cluster.metadata.IndexNameExpressionResolver.concreteIndices(IndexNameExpressionResolver.java:133)

multiline.pattern: '^\['
multiline.negate: true
multiline.match: after

也可以:
multiline.pattern: '^[[:space:]]'
multiline.negate: false
multiline.match: after

日志样例05:
      multiline.pattern: '^\[0-9]{8}'
      multiline.negate: true
      multiline.match: after
```

```
filebeat-modules
https://www.elastic.co/guide/en/beats/filebeat/current/filebeat-modules-overview.html

apache模块文件位置
modules.d/apache.yml

./filebeat modules list
./filebeat modules enable apache2 mysql
```

```
重置filebeat抓取记录，删除filebeat文件偏移量
Filebeat 使用一个注册表文件来记录已经读取的日志文件及其偏移量

/var/lib/filebeat/registry
rm -rf /var/lib/filebeat/registry/filebeat/
```
