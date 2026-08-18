# elasticsearch_ELK 空间清理

```
Curator and Index Lifecycle Management
https://www.elastic.co/guide/en/elasticsearch/client/curator/current/index.html

env
CentOS Linux release 7.6.1810 (Core)
elasticsearch-curator-5.8.4-1.x86_64

手动删除过期(上月)索引
curl  http://10.21.189.82:29200/_cat/indices?v |grep 2022.07 |awk '{print $3}' >202207.indices
for ind in $(cat 202207.indices);do echo $ind; curl -XDELETE http://192.168.3.222:29200/$ind ;done

Install

方法1:
rpm --import https://packages.elastic.co/GPG-KEY-elasticsearch
tee >/etc/yum.repos.d/curator.repo <<EOF
[curator-5]
name=CentOS/RHEL 7 repository for Elasticsearch Curator 5.x packages
baseurl=https://packages.elastic.co/curator/5/centos/7
gpgcheck=1
gpgkey=https://packages.elastic.co/GPG-KEY-elasticsearch
enabled=1
EOF
yum install elasticsearch-curator

方法2:不如方法1靠谱
pip3 install elasticsearch-curator==5.6.0

crontal -l
01 04  * * * /usr/bin/curator --config /opt/es_clean/curator.yml  /opt/es_clean/delete_indices.yml

mkdir -p /opt/es_clean/
tee >/opt/es_clean/curator.yml <<EOF
---
# Remember, leave a key empty if there is no value.  None will be a string,
# not a Python "NoneType"
client:
  hosts:
    - 172.22.240.222
    - 172.22.240.223
    - 172.22.240.224
  port: 29200
  url_prefix:
  use_ssl: False
  certificate:
  client_cert:
  client_key:
  ssl_no_validate: False
  http_auth:
  timeout: 30
  master_only: False
logging:
  loglevel: INFO
  logfile: /var/log/curator.log
  logformat: default
  blacklist: ['elasticsearch', 'urllib3']
EOF

tee>/opt/es_clean/delete_indices.yml <<EOF
---
# Remember, leave a key empty if there is no value.  None will be a string,
# not a Python "NoneType"
#
# Also remember that all examples have 'disable_action' set to True.  If you
# want to use this action as a template, be sure to set this to False after
# copying it.
actions:
  1:
    action: delete_indices
    description: >-
      Delete indices older than 45 days (based on index name), for logstash-
      prefixed indices. Ignore the error if the filter does not result in an
      actionable list of indices (ignore_empty_list) and exit cleanly.
    options:
      ignore_empty_list: True
      timeout_override:
      continue_if_exception: False
      #disable_action: True
    filters:
    # - filtertype: pattern
    #   kind: prefix
    #   value: logstash-
    #   exclude:
    - filtertype: age
      source: name
      direction: older
      timestring: '%Y.%m.%d'
      unit: days
      unit_count: 7
      exclude:
EOF
```
