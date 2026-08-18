# ELK_config,es,logstash,filebeat,kibana

```
elasticsearch6.2.2

# cat /opt/elasticsearch/config/elasticsearch.yml
cluster.name: igo-log-es
node.name: node4
node.attr.rack: log-route-lb1
path.data: /data/elasticsearch/data
path.logs: /data/elasticsearch/logs
network.host: 192.168.240.222
http.port: 29200
transport.tcp.port: 29300
discovery.zen.ping.unicast.hosts: ["192.168.240.222:29300", "192.168.240.223:29300", "192.168.240.224:29300"]
discovery.zen.minimum_master_nodes: 2

tips
推迟5分钟分片分配,支持运行时进行修改
PUT /_all/_settings 
{
  "settings": {
    "index.unassigned.node_left.delayed_timeout": "5m" 
  }
}
```

```
filebeat6.2

# cat /opt/filebeat/filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /data/volume/igozhang.com/game_flow_all.log
  fields:
    log_topics: area1_game_flow_all
  json.keys_under_root: true
  json.overwrite_keys: true
  json.message_key: message
- type: log
  enabled: true
  paths:
    - /data/volume/igozhang.com/game_flow_all.log
  fields:
    log_topics: area2_game_flow_all
  json.keys_under_root: true
  json.overwrite_keys: true
  json.message_key: message
- type: log
  enabled: true
  paths:
    - /data/volume/igozhang.com/game_flow_all.log
  fields:
    log_topics: area3_game_flow_all
  json.keys_under_root: true
  json.overwrite_keys: true
  json.message_key: message
- type: log
  enabled: true
  paths:
    - /data/volume/igozhang.com/game_flow_all.log
  fields:
    log_topics: area4_game_flow_all
  json.keys_under_root: true
  json.overwrite_keys: true
  json.message_key: message
processors:
#  - add_host_metadata: ~
#  - add_cloud_metadata: ~
  - drop_fields:
     fields: ["beat","input_type","source"]
output.kafka:
  enabled: true
  max_message_bytes: 100000000
  hosts: ["192.168.240.171:29092","192.168.240.172:29092","192.168.240.173:29092"]
  topic: '%{[fields][log_topics]}'
```

```
kibana6.2.2

# cat /opt/kibana/config/kibana.yml
i18n.defaultLocale: "en"
server.host: "192.168.240.224"
server.name: "log-kibana"
elasticsearch.url: "http://192.168.240.224:29200"
```
