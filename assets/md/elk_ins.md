# ELK7.9.3日志系统

elk7.9.3

```
elasticsearch,es

hosts
tee >>/etc/hosts<<EOF
10.10.2.31 es-node01
10.10.2.32 es-node02
10.10.2.33 es-node03
10.10.2.36 logstash-kibana
EOF

yum -y install java-11-openjdk java-11-openjdk-devel java-11-openjdk-headless lrzsz htop mlocate

mkdir -p /opt
mkdir -p /data/es/{data,logs}
yum -y install ./elasticsearch-7.9.3-x86_64.rpm
chown -R elasticsearch.elasticsearch /data/es

tee >>/etc/elasticsearch/jvm.options<<EOF
-Xms1g
-Xmx1g
EOF

hname=$(hostname)
pp=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')
tee >/etc/elasticsearch/elasticsearch.yml<<EOF
cluster.name: log-es
path.data: /data/es/data
path.logs: /data/es/logs
node.name: $hname
network.host:  $pp
http.port: 29200
transport.tcp.port: 29300
discovery.seed_hosts: ["10.10.2.31:29300","10.10.2.32:29300","10.10.2.33:29300"]
discovery.zen.minimum_master_nodes: 2
cluster.initial_master_nodes: igo_31 #这行只有第一次启动第一台机器指定，启动完成后删除
EOF

http://10.10.2.32:29200/

elasticsearch_config
7.9.3

31 ~]# cat /etc/elasticsearch/elasticsearch.yml
cluster.name: log-es01
path.data: /data/es/data
path.logs: /data/es/logs
node.master: true
node.data: false
node.name: node01
network.host:  192.168.81.31
http.port: 29200
transport.tcp.port: 29300
discovery.seed_hosts: ["192.168.81.31:29300","192.168.81.32:29300","192.168.81.33:29300","192.168.81.34:29300","192.168.81.35:29300","192.168.81.36:29300"]
discovery.zen.minimum_master_nodes: 2
cluster.initial_master_nodes: node01

32 data]# cat /etc/elasticsearch/elasticsearch.yml
cluster.name: log-es01
path.data: /data/es/data
path.logs: /data/es/logs
node.master: true
node.data: false
node.name: node02
network.host:  192.168.81.32
http.port: 29200
transport.tcp.port: 29300
discovery.seed_hosts: ["192.168.81.31:29300","192.168.81.32:29300","192.168.81.33:29300","192.168.81.34:29300","192.168.81.35:29300","192.168.81.36:29300"]
discovery.zen.minimum_master_nodes: 2

33 data]# cat /etc/elasticsearch/elasticsearch.yml
cluster.name: log-es01
path.data: /data/es/data
path.logs: /data/es/logs
node.master: true
node.data: true
node.voting_only: true
node.name: node03
network.host:  192.168.81.33
http.port: 29200
transport.tcp.port: 29300
discovery.seed_hosts: ["192.168.81.31:29300","192.168.81.32:29300","192.168.81.33:29300","192.168.81.34:29300","192.168.81.35:29300","192.168.81.36:29300"]
discovery.zen.minimum_master_nodes: 2

34 data]# cat /etc/elasticsearch/elasticsearch.yml
cluster.name: log-es01
path.data: /data/es/data
path.logs: /data/es/logs
node.master: false
node.data: true
node.name: node04
network.host:  192.168.81.34
http.port: 29200
transport.tcp.port: 29300
discovery.seed_hosts: ["192.168.81.31:29300","192.168.81.32:29300","192.168.81.33:29300","192.168.81.34:29300","192.168.81.35:29300","192.168.81.36:29300"]
discovery.zen.minimum_master_nodes: 2

35 data]# cat /etc/elasticsearch/elasticsearch.yml
cluster.name: log-es01
path.data: /data/es/data
path.logs: /data/es/logs
node.master: false
node.data: true
node.name: node05
network.host:  192.168.81.35
http.port: 29200
transport.tcp.port: 29300
discovery.seed_hosts: ["192.168.81.31:29300","192.168.81.32:29300","192.168.81.33:29300","192.168.81.34:29300","192.168.81.35:29300","192.168.81.36:29300"]
discovery.zen.minimum_master_nodes: 2

# cat /etc/elasticsearch/elasticsearch.yml
cluster.name: log-es01
path.data: /data/es/data
path.logs: /data/es/logs
node.master: false
node.data: true
node.name: node06
network.host:  192.168.81.36
http.port: 29200
transport.tcp.port: 29300
discovery.seed_hosts: ["192.168.81.31:29300","192.168.81.32:29300","192.168.81.33:29300","192.168.81.34:29300","192.168.81.35:29300","192.168.81.36:29300"]
discovery.zen.minimum_master_nodes: 2
```

```
logstash

mkdir -p /data/logstash/{data,logs}
chown -R logstash.logstash /data/logstash

yum -y install java-1.8.0-openjdk
yum -y install ./logstash-7.9.3.rpm

37 data]# grep -vE '^$|^#' /etc/logstash/logstash.yml
path.data: /data/logstash/data
pipeline.ordered: auto
path.logs: /data/logstash/logs

https://www.elastic.co/guide/en/logstash/7.9/plugins-outputs-elasticsearch.html#plugins-outputs-elasticsearch-healthcheck_path
It is important to exclude dedicated master nodes from the hosts list to prevent LS from sending bulk requests to the master nodes. So this parameter should only reference either data or client nodes in Elasticsearch.

37 ~]# cat /etc/logstash/conf.d/pipe01.conf
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
    hosts => ["http://10.21.81.34:29200","http://10.21.81.35:29200","http://10.21.81.36:29200"]
    index => "%{[fields][log_topics]}-%{+YYYY.MM.dd}"
}
}
```

```
kibana

yum -y install ./kibana-7.9.3-x86_64.rpm

# cat /etc/kibana/kibana.yml
i18n.locale: "en"
server.port: 5601
server.host: "10.21.81.37"
elasticsearch.hosts: ["http://10.21.81.31:29200"]
```

```
filebeat 7.9.3
echo "message" | ./filebeat -c ~/data/filebeat_test/filebeat.yml -e 2> /dev/null
153 ~]# cat /etc/filebeat/filebeat.yml
filebeat.inputs:
- type: log
  paths:
    - /home/ops/logs/access.log
  fields:
    log_topics: app_acc_uat_platform
    tag: app_acc
  multiline.pattern: '^\[[0-9]{4}-[0-9]{2}-[0-9]{2}'
  multiline.negate: true
  multiline.match: after
- type: log
  paths:
    - /home/ops/logs/error.log
  fields:
    log_topics: app_err_uat_platform
    tag: app_err
  multiline.pattern: '^\[[0-9]{4}-[0-9]{2}-[0-9]{2}'
  multiline.negate: true
  multiline.match: after

output.logstash:
  hosts: ["10.21.81.37:5044"]
```
