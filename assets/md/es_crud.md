# elasticsearch_CRUD

```
测试版本 v6.2.2

常用
1  检查集群状态是否健康，
    绿色表示一切正常, 黄色表示所有的数据可用但是部分副本还没有分配,红色表示部分数据因为某些原因不可用.
curl 'localhost:9200/_cat/health?v'
curl -u elastic/igo1234 'localhost:9200/_cat/health?v'
2  获取集群的节点列表：
curl 'localhost:9200/_cat/nodes?v'
3  列出所有索引：
curl 'localhost:9200/_cat/indices?v'
4  创建并查看索引：
for i in {1..5}; do curl -s -X PUT "http://es_igo.local:19527/demo-index-$i" && echo " ✅ demo-index-$i created" || echo " ❌ demo-index-$i failed"; done

5. 查询集群配置信息：
localhost:9200/_cluster/settings?include_defaults=true

curl -XPUT 'localhost:9200/customer?pretty'
curl 'localhost:9200/_cat/indices?v'
http://172.29.240.174:29200/_nodes?pretty  节点详细信息
http://172.29.240.174:29200/_nodes/node-1?pretty
5  删除索引
curl -XDELETE 'http://192.168.121.128:9200/logstash-pipeline02-2022.02.11'
curl -XDELETE '10.1.1.54:29200/*2025.09*'

查记录
先查ID
# curl -XGET 'http://10.21.81.36:29200/test-test_log-2022.02/_search'
再查记录
GET index_name/_doc/id：获取指定索引中的指定文档
# curl -XGET 'http://10.21.81.36:29200/test-test_log-2022.02/_doc/wRjtBn8BQw8FL-Kto6S6'

健康查询：
GET _cluster/health
GET _cluster/health?level=indices
GET _cluster/health/my_index
GET _cluster/health?level=shards
GET _cluster/allocation/explain   返回第一个未分配 Shard 的原因
```

```
滚动重启
1.最好停止数据写入
2.停止分片分配
PUT /_cluster/settings
{
    "transient" : {
        "cluster.routing.allocation.enable" : "none"
    }
}

3.重启
4.开启分片分配
PUT /_cluster/settings
{
    "transient" : {
        "cluster.routing.allocation.enable" : "all"
    }
}
```

```
集群cluster
curl -X GET "localhost:9200/_cluster/health?pretty"
curl -X GET "http://172.22.240.224:29200/_cat/nodes?pretty&v"
curl -X GET "localhost:9200/_nodes/stats/http?pretty"
http是属性，另有indices, fs, http, jvm, os, process, thread_pool, discovery等，支持组合（如indices,fs,http）
分片
curl -X GET "localhost:9200/_cat/shards?v&pretty"

index
curl -X GET "localhost:9200/_cat/indices?v"
curl -X GET "localhost:9200/index_name/_stats?pretty"
查数据量
curl -X GET "localhost:9200/_cat/count/chat_index_alias?v&pretty"
PUT /megacorp/employee/1
GET /megacorp/employee/1   //指定
GET /megacorp/employee/_search //返回所有
GET /megacorp/employee/_search?q=last_name:Smith
q=Query （DSL）
DELETE /ecommerce
复杂查询
GET /megacorp/employee/_search
{
    "query" : {
        "bool": {
            "must": {
                "match" : {
                    "last_name" : "smith" 
                }
            },
            "filter": {
                "range" : {
                    "age" : { "gt" : 30 } 
                }
            }
        }
    }
}

cat 状态查询
可以后面加一个v，让输出内容表格显示表头; pretty则让输出缩进更规范
curl http://localhost:9200/_cat
=^.^=
/_cat/allocation
/_cat/shards
/_cat/shards/{index}
/_cat/master
/_cat/nodes
/_cat/tasks
/_cat/indices
/_cat/indices/{index}
/_cat/segments
/_cat/segments/{index}
/_cat/count
/_cat/count/{index}
/_cat/recovery
/_cat/recovery/{index}
/_cat/health
/_cat/pending_tasks
/_cat/aliases
/_cat/aliases/{alias}
/_cat/thread_pool
/_cat/thread_pool/{thread_pools}
/_cat/plugins
/_cat/fielddata
/_cat/fielddata/{fields}
/_cat/nodeattrs
/_cat/repositories
/_cat/snapshots/{repository}
/_cat/templates
```

```
REST,多行

1.节点间分片迁移
POST /_cluster/reroute
{
  "commands": [
    {
      "move": {
        "index": "indexname",
        "shard": 1,
        "from_node": "nodename",
        "to_node": "nodename"
      }
    }
  ]
}

2.节点优雅下线
PUT /_cluster/settings
{
  "transient": {
    "cluster.routing.allocation.exclude._ip": "122.5.3.55"
  }
}

3.强制刷新
确保当前仅存储在事务日志中的所有数据也永久存储在Lucene索引
POST /_flush

4.调整集群恢复速度
PUT /_cluster/settings
{
  "transient": {
    "indices.recovery.max_bytes_per_sec": "80mb"
  }
}

5.节点清理缓存
会降低性能，释放内存，拜托OOM
POST /_cache/clear

6.调整断路器，限制搜索内存，避免OOM
PUT /_cluster/settings
{
  "persistent": {
    "indices.breaker.total.limit": "40%"
  }
}
```
