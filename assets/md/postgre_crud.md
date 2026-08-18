# postgre_crud

## 连接，检索基础信息

```
psql -U postgres -h 127.0.0.1 –port 8432
psql -U postgres -h 10.10.8.211 --port 8433
  本机127.0.0.1 地址连接可以不需要输入密码
select VERSION();
  查看版本
\s
  查看最近history命令记录
\l  或者 \list
  查看数据库
\c igodb
  连接igodb数据库
\d 或 \dt
或
SELECT tablename FROM pg_tables;
WHERE tablename NOT LIKE 'pg%'
AND tablename NOT LIKE 'sql_%'
ORDER BY tablename;
  列出当前数据库的所有数据表
\du
select usename from pg_user; // 获取系统用户信息
  查看系统用户信息

\d igotb
  检查igotb表的所有字段
SELECT * from pg_stat_activity;
  查看当前执行的事务
SELECT client_addr ,client_port,query_start FROM pg_stat_activity;
  查看当前连接

postgres=# \encoding utf-8       // 设置客户端的字元集
postgres=# \encoding                  // 显示客户端的字元集
postgres=# show client_encoding;      // 显示客户端的字元集
postgres=# show server_encoding;      // 显示服务器的字元集
  字符集

help
  帮助信息
\q 或 quit 或 exit
  退出数据库
```

## 用户及授权

引自阿里云[Postgresql 数据库用户权限授权]<https://developer.aliyun.com/article/815729?accounttraceid=86ba2549fbfb45638379ee159a8874c7ojnu>

```
-- 创建测试库
CREATE DATABASE testdb;
-- 创建测试表
CREATE TABLE "test_1" ("id" bigserial NOT NULL, "txt" text NOT NULL,   PRIMARY KEY ("id") );

-- 需要切换到 testdb 库执行相关的指令，否则无效
-- \c testdb
-- 创建角色并授予登陆的权限，默认拥有public的部分权限，只能查看数据库名词和表名，其他不能查阅
CREATE ROLE db_role1 WITH 
  LOGIN
  NOSUPERUSER
  NOINHERIT
  NOCREATEDB
  NOCREATEROLE
  NOREPLICATION
  ENCRYPTED PASSWORD '123456';    
-- 添加注释 
COMMENT ON ROLE db_role1 IS '只读用户';

-- 授予只读权限
GRANT select ON all tables in schema public TO db_role1;
-- 查看权限授权，发现目前库里的所有表的权限都已授权，但是重新创建的表，则没有对应的权限，还需要执行 ALTER...GRANT... 才能动态构建新创建表的权限授权
select * from information_schema.table_privileges where grantee='db_role1';

新表自动授权
-- 为后续创建的表赋予相关权限，则必须使用 ALTER  ... GRANT ... 的语法授权
-- 函数授权
GRANT USAGE ON SCHEMA public TO db_role1;
-- 每创建一个新表，则赋予对象查询的权限
ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public
GRANT select ON TABLES TO db_role1;

角色授权
-- 创建用户，默认可查看到所有的数据库名和表名
CREATE USER pguser_1 WITH PASSWORD '123456';
-- 将角色1授权给用户1
grant db_role1 to pguser_1; 

-- 回收所有权限
revoke all on database "testdb" from db_role1;
revoke all on all tables in schema public from db_role1;
revoke all ON SCHEMA public from db_role1;
-- 回收默认权限
alter DEFAULT PRIVILEGES for role postgres IN SCHEMA public
revoke all on tables from db_role1;
drop role db_role1; -- 删除角色
drop user pguser_1; -- 删除用户
```

## 库表操作

引自[psql基本命令]<https://www.cnblogs.com/my-blogs-for-everone/p/10226473.html>

```
列出数据库名
\l
或
SELECT datname FROM pg_database;

切换数据库
\c 数据库名

列出当前数据库所有表
\dt
列出表名
SELECT tablename FROM pg_tables;
WHERE tablename NOT LIKE 'pg%'
AND tablename NOT LIKE 'sql_%'
ORDER BY tablename;

其它SQL命令通用如(标准化SQL语句):
*创建数据库：
create database [数据库名];

*删除数据库：
drop database [数据库名]; 

*创建表：
create table ([字段名1] [类型1] ;,[字段名2] [类型2],......<,primary key (字段名m,字段名n,...)>;);

*在表中插入数据：
insert into 表名 ([字段名m],[字段名n],......) values ([列m的值],[列n的值],......);

*显示表内容:
select * from student;

*重命名一个表：
alter table [表名A] rename to [表名B];

*删除一个表：
drop table [表名]; 

*在已有的表里添加字段：
alter table [表名] add column [字段名] [类型];

*删除表中的字段：
alter table [表名] drop column [字段名];

*重命名一个字段： 
alter table [表名] rename column [字段名A] to [字段名B];

*给一个字段设置缺省值： 
alter table [表名] alter column [字段名] set default [新的默认值];

*去除缺省值： 
alter table [表名] alter column [字段名] drop default;

*修改表中的某行某列的数据：
update [表名] set [目标字段名]=[目标值] where [该行特征];
```

## 报错

```
1. ERROR:  cannot execute CREATE DATABASE in a read-only transaction
连接到从节点，无建数据库权限，需要重新连接到主节点；
```
