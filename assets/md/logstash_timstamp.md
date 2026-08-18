# logstash_timstamp,时间戳

```
env
es,logstash v7.9.3

使用日志时间戳替代kibana时间戳
```

```
# cat /etc/logstash/conf.d/5044.conf
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

if [fields][log_topics] == "app_acc_pet_env"{
    grok {
        match => ["message","%{TIMESTAMP_ISO8601:log.date}"]
    }
    date {
        match => ["log.date", "ISO8601"]
    }
}

}

output {
  elasticsearch {
    hosts => ["http://10.21.81.34:29200","http://10.21.81.35:29200","http://10.21.81.36:29200"]
    index => "%{[fields][log_topics]}-%{+YYYY.MM.dd}"
}
}
```

效果图

```
对所有日志生效:
$ cat /etc/logstash/conf.d/5044.conf
input {
  beats {
    port => 5044
  }
}
filter {

    grok {
        match => ["message","%{TIMESTAMP_ISO8601:log.date}"]
    }
    date {
        match => ["log.date", "ISO8601"]
    }
}

output {
  elasticsearch {
    hosts => ["http://10.21.81.34:29200","http://10.21.81.35:29200","http://10.21.81.36:29200"]
    index => "%{[fields][log_topics]}-%{+YYYY.MM.dd}"
}
}
```
