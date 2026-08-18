# elasticsearch_troubleshooting,es排错

```
集群状态的含义
红色：至少一个主分片未分配成功；
黄色：至少一个副本分片未分配成功；
绿色：全部主&副本都分配成功。

集群
GET _cluster/health
查看具体问题索引
GET _cluster/health?level=indices
GET /_cat/indices?v&health=yellow
GET /_cat/indices?v&health=red
具体问题分片
GET _cluster/health?level=shards
原因
GET _cluster/allocation/explain
GET _cat/shards?h=index,shard,prirep,state,unassigned.reason

问题及原因
1. INDEX_CREATED
Unassigned as a result of an API creation of an index.
2. CLUSTER_RECOVERED
Unassigned as a result of a full cluster recovery.
3. INDEX_REOPENED
Unassigned as a result of opening a closed index.
4. DANGLING_INDEX_IMPORTED
Unassigned as a result of importing a dangling index.
5. NEW_INDEX_RESTORED
Unassigned as a result of restoring into a new index.
6. EXISTING_INDEX_RESTORED
Unassigned as a result of restoring into a closed index.
7. REPLICA_ADDED
Unassigned as a result of explicit addition of a replica.
8. ALLOCATION_FAILED
Unassigned as a result of a failed allocation of the shard.
9. NODE_LEFT
Unassigned as a result of the node hosting it leaving the cluster.
10. REROUTE_CANCELLED
Unassigned as a result of explicit cancel reroute command.
11. REINITIALIZED
When a shard moves from started back to initializing, for example, with shadow replicas.
12. REALLOCATED_REPLICA
A better replica location is identified and causes the existing replica allocation to be cancelled.
```

```
es6.2.2跨集群迁移缓慢问题排查
（最终发现问题原因是其中一个节点vpn不可用导致）
curl -XGET localhost:9200/_cat/shards?h=index,shard,prirep,state,unassigned.reason| grep UNASSIGNED
定位yellow原因
curl -XDELETE localhost:9200/index_name

curl -XGET localhost:9200/_cluster/allocation/explain?pretty

curl http://172.27.240.174:29200/_cat/shards?h=index,shard,prirep,state,unassigned.reason |grep -i unassigned >2.unassigned

https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-reroute.html
curl -XPOST http://172.27.240.174:29200/_cluster/reroute?retry_failed=true

curl -XPUT 'localhost:9200/_cluster/settings' -d'
{
"transient": {
"discovery.zen.minimum_master_nodes": 3
}
}'˚

curl -X PUT "172.22.240.222:29200/_cluster/settings?pretty" -H 'Content-Type: application/json' -d'
{
  "transient" : {
    "cluster.routing.allocation.exclude._ip" : "172.27.240.174"
  }
}'

curl -XPUT "localhost:9200/<INDEX_NAME>/_settings?pretty" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "index.unassigned.node_left.delayed_timeout": "5m"
  }
}'
```

```
现象
最近两天日志kibana查不到

分析
systemctl status logstash -l
Jun 17 16:24:05 hostname-igo-88 logstash[563508]: [2022-06-17T16:24:05,939][WARN ][logstash.outputs.elasticsearch][pipe5044][2886358e326bdd08752d5eb359b30b4eadce4c6d6c3644ac9ec6b26cc13ac614] Could not index event to Elasticsearch. {:status=>400, :action=>["index", {:_id=>nil, :_index=>"index_was_sysout-2022.06.17", :routing=>nil, :_type=>"_doc"}, #<LogStash::Event:0x1318fbc2>], :response=>{"index"=>{"_index"=>"in_asy_was_sysout-2022.06.17", "_type"=>"_doc", "_id"=>nil, "status"=>400, "error"=>{"type"=>"illegal_argument_exception", "reason"=>"Validation Failed: 1: this action would add [2] total shards, but this cluster currently has [4000]/[4000] maximum shards open;"}}}}

找到历史(待删除)索引
# curl  http://10.21.189.82:29200/_cat/indices?v |grep 2022.04 |awk '{print $3}' >202204.indices
删除历史索引
for ind in $(cat 202204.indices);do echo $ind; curl -XDELETE http://10.21.189.82:29200/$ind ;done
```

```
kibana查日志报错:

错误:
加载数据时出错
[status_exception] error while executing search
search_phase_execution_exception
Elasticsearch exception [type=search_phase_execution_exception, reason=all shards failed]

solution
curl -XPUT "http://10.21.81.36:29200/app_err_log-2022.11.22/_settings" -H "Content-Type: application/json" -d '{ "index" : { "max_result_window" : 1000000} }'

验证
curl -XGET "http://10.21.81.36:29200/app_err_log-2022.11.22/_settings"
```

```
error:
Aug 28 10:25:19 cnabdabvdc05-189-88 logstash[19262]: [2023-08-28T10:25:19,093][WARN ][logstash.outputs.elasticsearch][pipe5044][1550d3a30817a46916734b252ab3196851ee9423aa11eca36e1b674f9b527c78] Could not index event to Elasticsearch. {:status=>400, :action=>["index", {:_id=>nil, :_index=>"systemerr-ca_asy_was12-2023.08.28", :routing=>nil, :_type=>"_doc"}, #<LogStash::Event:0x648f4081>], :response=>{"index"=>{"_index"=>"systemerr-ca_asy_was12-2023.08.28", "_type"=>"_doc", "_id"=>nil, "status"=>400, "error"=>{"type"=>"illegal_argument_exception", "reason"=>"Validation Failed: 1: this action would add [2] total shards, but this cluster currently has [4000]/[4000] maximum shards open;"}}}}

solution:
curl -XPUT -H "Content-Type: application/json" -d '{"transient":{"cluster":{"max_shards_per_node":10000}}}' 'http://localhost:29200/_cluster/settings'
```

```
分片数满错误：
this action would add [2] total shards, but this cluster currently has [6000]/[6000] maximum shards open
查询当前分片数，默认每节点1000：
curl -X GET "localhost:9200/_cat/shards?v=true&pretty"
curl -X GET "localhost:9200/_cluster/stats?human&pretty"
更精准：
curl -X GET "localhost:9200/_cluster/settings?include_defaults=true&filter_path=defaults.cluster.max_shards_per_node"
kibana-dev_tools:
GET _cluster/settings?include_defaults=true&filter_path=defaults.cluster.max_shards_per_node

terminal设置:
curl -u elastic:ab1234 -X PUT "10.25.48.54:29200/_cluster/settings" -H 'Content-Type: application/json' -d '{"persistent": {"cluster.max_shards_per_node": "1500"}}'
kibana-dev_tools设置:
PUT _cluster/settings
{
    "persistent": {
        "cluster.max_shards_per_node": "1500"
    }
}
```
