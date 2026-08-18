# filebeat日志接入脚本

```
filebeat7.17
数据库_Oracle接入
Websphere接入

tips
topic索引名称不接受大写字母;
```

```
Oracle
redhat6.9

#!/bin/bash
#author:igo
#for oracle log collect

pp=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')
topic_xml=asy_bd_ora_xml
topic_alert=asy_bd_ora_alert

read -p  "input ora_alert_log: " ora_alert_log
read -p  "input ora_xml_log: " ora_xml_log

echo $pp
echo "ora_alert_log path is: "$ora_alert_log
echo "ora_xml_log path is: "$ora_xml_log

tee >/etc/filebeat/filebeat.yml<<EOF
filebeat.inputs:
- type: log
  paths:
    - $ora_alert_log
  fields:
    log_topics: $topic_alert
    tag: ora_alert
    serv_ip: $pp
  multiline.type: pattern
  multiline.pattern: '\w{3}\s\w{3}\s\d{2}\s\d{2}\:\d{2}\:\d{2}\s\d{4}'
  multiline.negate: true
  multiline.match: after
- type: log
  processors:
  - decode_xml:
      field: message
      target_field: ""
      overwrite_keys: true
      ignore_missing: true
      ignore_failure: true
  paths:
    - $ora_xml_log
  fields:
    log_topics: $topic_xml
    tag: ora_xml
    serv_ip: $pp
  multiline.type: pattern
  multiline.pattern: '^\<msg\ time\='
  multiline.negate: true
  multiline.match: after

output.logstash:
  hosts: ["10.21.189.88:5044"]
EOF

service filebeat restart
chkconfig filebeat on

service filebeat restart
chkconfig filebeat on
```

```
Websphere

cp /etc/filebeat/filebeat.yml{,.$(date +%Y%m%d)}
pp=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')
id=$(hostname |awk -F'-' '{print $NF}')
logid01=01
logid02=02
log_topics=cq_igo_was01
wasid=was03
tee >/etc/filebeat/filebeat.yml<<EOF
filebeat.inputs:
- type: log
  paths:
    - /opt/IBM/WebSphere/AppServer/profiles/AppSrv01/logs/member$id$logid01/SystemOut.log
    - /opt/IBM/WebSphere/AppServer/profiles/AppSrv01/logs/member$id$logid02/SystemOut.log
  fields:
    log_topics: systemout-$log_topics
    serv_ip: $pp
    wasid: $wasid
  multiline.type: pattern
  multiline.pattern: '^\['
  multiline.negate: true
  multiline.match: after

- type: log
  paths:
    - /opt/IBM/WebSphere/AppServer/profiles/AppSrv01/logs/member$id$logid01/SystemErr.log
    - /opt/IBM/WebSphere/AppServer/profiles/AppSrv01/logs/member$id$logid02/SystemErr.log
  fields:
    log_topics: systemerr-$log_topics
    serv_ip: $pp
    wasid: $wasid
  multiline.type: pattern
  multiline.pattern: '^\['
  multiline.negate: true
  multiline.match: after

output.logstash:
  hosts: ["192.168.89.88:5044"]
EOF

service filebeat start
chkconfig filebeat on
service filebeat start
```
