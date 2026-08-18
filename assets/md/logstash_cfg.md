# logstash_config

```
参考网站:
https://github.com/logstash-plugins/logstash-patterns-core/tree/main/patterns
http://grokconstructor.appspot.com/
https://www.elastic.co/guide/en/logstash/7.9/plugins-filters-grok.html
```

```
调试
logstash 7.9.3

/usr/share/logstash/bin/logstash -e 'input { stdin { } } output { stdout {} }'
/usr/share/logstash/bin/logstash --config.reload.automatic -f pipe01.conf 

/usr/share/logstash/bin/logstash --config.test_and_exit  -f /etc/logstash/conf.d/5044.conf
logstash -f /etc/logstash/conf.d/t.conf

# cat /etc/logstash/conf.d/t.conf
input { stdin {} }

filter {
    grok {
        match => ["message", "%{TIMESTAMP_ISO8601:logdate}"]
        }
    date {
        match => ["logdate", "yyyy-MM-dd HH:mm:ss.SSS","MMM dd yyyy HH:mm:ss","MMM  d yyyy HH:mm:ss", "ISO8601"]
        locale => "cn"
        timezone => "Asia/Shanghai"
        target => "@timestamp"
}
    mutate {
         remove_field => [ "logdate" ]
    }
}

output {
 stdout { codec => rubydebug }
}
```

```
v7.9.3
双输出且带判断

# cat /etc/logstash/conf.d/5044.conf
input {
  beats {
    port => 5044
  }
}

output {
  elasticsearch {
    hosts => ["http://10.10.10.34:29200","http://10.10.10.35:29200","http://10.10.10.36:29200"]
    index => "%{[fields][log_topics]}-%{+YYYY.MM.dd}"
}
  if "igo" in [fields][log_topics] or "zhang" in [fields][log_topics] {
    kafka {
      bootstrap_servers => "10.10.20.243:9095,10.10.20.242:9095,10.10.20.241:9095"
      topic_id => "G-IGO_ZHANG"
      client_id => "igozhang_log_kafka"
}
}
}
```

```
logstash 7.9.3
cat /etc/logstash/conf.d/igoelk.conf

input {
  beats {
    port => 5044
 
  }
 
}
 
filter{
    mutate{
        remove_field => ["host"]
        remove_field => ["agent"]
        remove_field => ["ecs"]
        remove_field => ["tags"]
        remove_field => ["fields"]
        remove_field => ["@version"]
        remove_field => ["@timestamp"]
        remove_field => ["input"]
        remove_field => ["log"]
    }
}
 
output {
  elasticsearch {
    hosts => ["192.168.22.68:9200"]
    index => "english"
  }
  stdout { codec => rubydebug }
}
```

```
logstash6.2.2

# cat /opt/logstash/pipeline/logstash.conf
input {
        kafka {
                bootstrap_servers => "172.22.240.171:29092,172.22.240.172:9092,172.22.240.173:9092"
                topics => ["resource_nginx","nginx_charge_manager","nginx_dal","nginx_imgmgr","nginx_vm_manager","nginx_vm_manager_v2","nginx_storage_index","nginx_storage_scheduler","nginx_charge_pc","nginx_charge_tv","nginx_charge_mobile","nginx-resource_type","nginx_access-service","nginx_archiver_manager","nginx_servo","nginx_tj-union-report","nginx_tj-union-access","nginx_tj-resource-limit","nginx_access","nginx_access_error","nginx_servo_error","nginx_live","nginx_live_error","img-res","resource_type"]
                codec => "json"
                consumer_threads => 2
                auto_offset_reset => "latest"
                enable_auto_commit => true
                auto_commit_interval_ms => "1000"
                group_id => nginx_logstash
        }
}

filter {

      if [fields][log_topics] =~ "nginx|resource"{
         grok {
          match => { "message" => '\[%{TIMESTAMP_ISO8601:date}\] (%{IPORHOST:remote_addr}|-) (%{QS:remote_user}|-) - (%{IPORHOST:server_name}|%{DATA:server_name}|-) (%{DATA:http_referer}|-) (%{DATA:http_user_agent}|-) (%{DATA:http_x_forwarded_for}|-) %{DATA:request} - %{BASE10NUM:status:int} (?:%{BASE10NUM:body_bytes_sent:int}|-) (%{BASE16FLOAT:request_time:float}|-) - (%{HOSTPORT:upstream_addr}|%{SPACE}) (%{BASE10NUM:upstream_status:int}|-) (?:%{BASE10NUM:upstream_response_time:float}|-)' }
         }
      }

}
output {
  elasticsearch {
    codec => rubydebug
    hosts => ["http://172.22.240.222:29200","http://172.22.240.223:29200","http://172.22.240.224:29200"]
    index => "cgbackend-%{[fields][log_topics]}-%{+YYYY.MM.dd}"
    #user => "elastic"
    #    #    #password => "changeme"
    }
    #            stdout { codec => rubydebug }
}
```

```
示例01:
55.3.244.1 GET /index.html 15824 0.043

Grok:
filter {
      grok {
        match => { "message" => "%{IP:client} %{WORD:method} %{URIPATHPARAM:request} %{NUMBER:bytes} %{NUMBER:duration}" }
      }
    }
	
结果:
After the grok filter, the event will have a few extra fields in it:
client: 55.3.244.1
method: GET
request: /index.html
bytes: 15824
duration: 0.043
```

```
日志示例:
nginx_access
10.98.91.152 - - [14/Mar/2022:18:41:23 +0800] "POST /stomp/296/shoguvrq/xhr?t=1647254458793 HTTP/1.1" 200 33 "https://igozhang.cn/basicFunction/locmanage/demandAndLoc" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36" "-"
```
