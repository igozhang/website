# elasticsearch数据迁移

```
1.reindex
POST _reindex
{
  "source": {
    "index": "my-index-000001"
  },
  "dest": {
    "index": "my-new-index-000001"
  }
}
```

```
2.集群数据备份恢复
PUT /_snapshot/my_backup/snapshot_hamlet_index?wait_for_completion=true
{
  "indices": "hamlet_*",
  "ignore_unavailable": true,
  "include_global_state": false,
  "metadata": {
    "taken_by": "mingyi",
    "taken_because": "backup before upgrading"
  }
}

POST /_snapshot/my_backup/snapshot_hamlet_index/_restore
```

```
3.集群node加入退出迁移
打通新旧集群网络（vpn）
更改配置打造大集群
待数据平衡后（green）逐台踢掉旧节点
```

```
4.data目录打包迁移
```

```
5.工具迁移
elasticdump
elasticsearch-migration
```
