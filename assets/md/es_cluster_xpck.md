# es_cluster_xpck安全认证

## ENV

CentOS Linux release 7.9.2009 (Core)  
openjdk version "11.0.20" 2023-07-18 LTS  
ELK 7.17.9 (amd64)

## ES操作

生成证书  
/usr/share/elasticsearch/bin/elasticsearch-certutil ca  
/usr/share/elasticsearch/bin/elasticsearch-certutil cert --ca elastic-stack-ca.p12  
迁移证书到etc  
mv /usr/share/elasticsearch/elastic-\* /etc/elasticsearch/  
同步到其他es节点  
scp /etc/elasticsearch/elastic-\* root@es-node02:/etc/elasticsearch/  
scp /etc/elasticsearch/elastic-\* root@es-node03:/etc/elasticsearch/

chown -R elasticsearch.elasticsearch /etc/elasticsearch/elastic-\*

tee >>/etc/elasticsearch/elasticsearch.yml<<EOF  
xpack.security.enabled: true  
xpack.security.transport.ssl.enabled: true  
xpack.security.transport.ssl.verification\_mode: certificate  
xpack.security.transport.ssl.keystore.path: /etc/elasticsearch/elastic-certificates.p12  
xpack.security.transport.ssl.truststore.path: /etc/elasticsearch/elastic-certificates.p12  
EOF

systemctl restart elasticsearch

/usr/share/elasticsearch/bin/elasticsearch-setup-passwords interactive  
ab1234  
Changed password for user [apm\_system]  
Changed password for user [kibana\_system]  
Changed password for user [kibana]  
Changed password for user [logstash\_system]  
Changed password for user [beats\_system]  
Changed password for user [remote\_monitoring\_user]  
Changed password for user [elastic]

## logstash配置

加用户和密码  
/etc/logstash/conf.d/pipe01.conf  
input {  
file {  
path => "/data/nginx\_acc.log"  
}  
}

output {  
elasticsearch {  
hosts => ["http://10.10.2.31:29200","http://10.10.2.32:29200","http://10.10.2.33:29200"]  
index => "nglog-%{+YYYY.MM.dd}"  
user => "elastic"  
password => "ab1234"  
}  
}

## kibana配置

/etc/kibana/kibana.yml

i18n.locale: "en"  
server.port: 5601  
server.host: "10.10.2.36"  
elasticsearch.hosts: "http://10.10.2.33:29200"  
elasticsearch.username: "kibana"  
elasticsearch.password: "ab1234"  
xpack.reporting.encryptionKey: "a\_random\_string"  
xpack.security.encryptionKey: "something\_at\_least\_32\_characters"

```
curator 配置
curator --config /opt/es_clean/curator.yml  /opt/es_clean/delete_indices.yml
curator 配置文件/opt/es_clean/curator.yml里头添加认证:
  http_auth: "elastic:ab1234"
```
