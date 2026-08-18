# weblogic配置数据源

添加配置海量数据源(VastData)  
登录weblogic\_console  
http://$ip:7001/console  
(密码有时候在启动脚本里，或者history |grep weblogic)

## 配置数据源

1. 添加通用数据源  
   锁定和编辑(Lock&Edit)2 服务(Services)2 数据源(DataSources)2 New2GenericDataSource
2. 补充配置信息  
   Create a New JDBC Data Source

- Name: JDBC VastData Source-1
- Scope: Global(default)
- JNDI Name: (VastData\_jndi)
- Database Type: PostgreSQL

Next（全默认）

- DatabaseType: PostgreSQL
- DatabaseDriver: PostgreSQL'sDriver(T)ype4)Versions:Any  
  Next（Transaction Options 全默认）

Next（Connection Properties)

- Database Name:
- Host Name:
- Port:
- Database User Name:
- Password:
- Confirm Password:

Next(Test Database Connection)

- Test configuration(左上角按钮)  
  测试成功会出现Connection test succeeded.

Next(Select Targets)  
根据需求选择server  
Finish

## 激活更改

左上角原来锁定和编辑(Lock&Edit)位置  
激活更改（Activate Changes）
