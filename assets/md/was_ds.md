# Websphere创建数据源连接海量VastData

登录管理后台console  
https://$ip/ibm/console

## 先装驱动(JDBC提供程序)

1. 新建JDBC提供程序  
   资源2 JDBC2 JDBC提供程序2 新建
2. 创建新的JDBC提供程序  
   填充项：

- 作用域（默认）
- 数据库类型（用户定义的）
- 实现类名（org.postgresql.ds.PGConnectionPoolDataSource）
- 名称 （vastdata\_JDBC\_Provider）  
  下一步：  
  JDBC驱动程序类文件完整路径  
  exp: D:\VastbaseG100\_jdbc\_2.4p\_2022090917.jar  
  下一步2完成；

## 配置数据源

1. 新建数据源  
   资源2 JDBC2 数据源2 新建
2. 基本数据源信息

- 作用域（默认）
- 数据源名（vastdata）
- JNDI名称 （vastdata\_jndi）  
  下一步

1. 选择JDBC提供程序

- vastdata\_JDBC\_Provider  
  下一步

1. 输入数据源特定于数据库的属性  
   默认2下一步
2. 设置安全性别名  
   默认（3无）  
   下一步  
   完成

## 配置DB认证

1. 选择刚创建的数据源  
   资源2 JDBC2 数据源2 数据源名（vastdata)
2. JAAS － J2C 认证数据  
   右侧其他属性

- 新建  
  填充项问DBA要  
  别名，用户标识，密码
- 确定
- 保存

1. 安全性设置  
   选择刚创建的数据源  
   资源2 JDBC2 数据源2 数据源名（vastdata)  
   左侧常规属性下拉2 安全性设置

- 组件管理的认证别名 （选刚新建的别名）
- 映射配置别名 （无）
- 容器管理的认证别名 （选刚新建的别名）  
  确定，保存

1. 定制属性  
   选择刚创建的数据源  
   资源2 JDBC2 数据源2 数据源名（vastdata)  
   右侧其他属性2 定制属性  
   显示过滤器功能（填充值问DBA要）

- serverName
- portNumber
- databaseName

## 重启Server测试连接

1. 重启应用Server
2. 选中数据源，点击测试连接  
   资源2 JDBC2 数据源2 数据源名（vastdata)2 测试连接
