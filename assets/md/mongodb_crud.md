# mongodb_CRUD

```
v4.2.5

./mongo --host 172.20.240.93 --port 40003
rs.status()
rs.slaveOk()
rs.conf（） 查看副本配置文档
db.printSlaveReplicationInfo() 副本节点中查看同步状态信息
rs.printReplicationInfo()  查oplog的大小,操作日志记录的起始时间
db.getReplicationInfo()  查oplog的状态、大小、存储的时间范围

备份：
mongodump -h 172.20.240.84:40001 -d d_union_user_game_area_storage_info -o /data/db.dumpdir
恢复：
mongorestore -h <hostname><:port> -d dbname <path>

增删改查：
db.runoob.find().pretty()
show dbs
db.runoob.insert({"name":"igozhang"})
db.runoob.remove({"rs-shard-01-132-1" : "1"})
db.runoob.insert({"rs-shard-01-132-1":"1"})
db.runoob.update({"rs-shard-01-132-1":"1"},{$set{"rs-shard-01-133-1":"1"}})
```
