# filebeat解析oracle日志

```
env
cent79
filebeat7.17.1
logstash7.9.3
```

```
注意这里filebeat7.17.1可以支持xml,7.9.3不支持process_xml
另外multiline不能写在processor里面,filebeat对格式要求非常严格
# vim /etc/filebeat/filebeat.yml

filebeat.inputs:
- type: log
  paths:
    - /data/oracle/diag/rdbms/asysit/asysit/trace/alert_asysit.log
  fields:
    log_topics: asy_sit_ca_ora_log
    tag: ora_alert
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
    - /data/oracle/diag/tnslsnr/igohostname/listener/alert/log.xml
  fields:
    log_topics: asy_sit_ca_ora_xml
    tag: ora_xml
  multiline.type: pattern
  multiline.pattern: '^\<msg\ time\='
  multiline.negate: true
  multiline.match: after

output.logstash:
  hosts: ["192.168.81.37:5044"]
```

```
 cat /etc/logstash/conf.d/5044.conf
input {
  beats {
    port => 5044
  }
}
filter {
if [fields][tag] == "ng_acc"{
    grok {
        match => { "message" => "%{COMBINEDAPACHELOG}"}
    }
}
}

output {
  elasticsearch {
    hosts => ["http://192.168.81.34:29200","http://192.168.34.81.35:29200","http://192.168.81.36:29200"]
    index => "%{[fields][log_topics]}-%{+YYYY.MM.dd}"
}
}
```
