# rabbitmq_crud

```
rabbitmq_server-3.8.6
CentOS Linux release 7.6.1810 (Core)
```

```
rabbitmq-server start
service rabbitmq-server start

rabbitmqctl status
rabbitmqctl stop
rabbitmq-plugins enable rabbitmq_management
rabbitmqctl list_queues
rabbitmqctl list_exchanges 
rabbitmqctl list_bindings

用户管理:
rabbitmqctl add_user igo igopwd
rabbitmqctl set_user_tags igo administrator
rabbitmqctl list_users

权限管理:
rabbitmqctl set_permissions [-p vhostpath] {user} {conf} {write} {read}
    conf:一个正则表达式match哪些配置资源能够被该用户访问。
    write:一个正则表达式match哪些配置资源能够被该用户读。
    read:一个正则表达式match哪些配置资源能够被该用户访问。
查看（指定vhost）所有用户的权限信息：
rabbitmqctl list_permissions [-p vhostPath]
查看指定用户的权限信息：
rabbitmqctl list_user_permissions {username}
清除用户的权限信息：
rabbitmqctl clear_permissions [-p vhostPath] {username}

插件管理:
rabbitmq-plugins [-n node] {command} [command options ...]
# ./sbin/rabbitmq-plugins enable rabbitmq_management  //启动
# ./sbin/rabbitmq-plugins list  //所有插件 -v 参数显示描述
# ./sbin/rabbitmq-plugins list -v management  //显示名称含有 "management" 的插件
# ./sbin/rabbitmq-plugins list -e rabbit -v  //所有显式或者隐式启动的插件
```
