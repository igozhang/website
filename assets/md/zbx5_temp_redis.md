# zabbix5添加redis模板

```
Env
Zabbix 5.0.2
CentOS Linux release 7.6.1810 (Core)
template_db_redis.xml 5.0.19

https://git.zabbix.com/projects/ZBX/repos/zabbix/browse/templates/db/redis/template_db_redis.xml?at=5.0.19
```

```
/etc/zabbix/zabbix_agent2.conf
//追加两行
pp=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')
auth='zwp1PSqh8KTHYIgo'
tee >>/etc/zabbix/zabbix_agent2.conf <<EOF
Plugins.Redis.Sessions.Redis1.Uri=tcp://$pp:6379
Plugins.Redis.Sessions.Redis1.Password=$auth
EOF

systemctl restart zabbix-agent2
systemctl status zabbix-agent2

zabbix_web-->配置-->模板-->Template_DB_Redis-->宏-->继承模板的宏-->{$REDIS.CONN.URI} 可用值Redis1
```
