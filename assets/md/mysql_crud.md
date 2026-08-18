# mysql_crud

```
v5.7.27

1、查看最大连接数
show variables like '%max_connections%';
2、修改最大连接数
set GLOBAL max_connections = 200;

mysql -uroot -p123456 -hlocalhost
```

```
数据库：
CREATE DATABASE 数据库名;
CREATE DATABASE  d_paas_mgr_release DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;drop database <数据库名>;
SHOW DATABASES;
use database_RUNOOB;

数据表：
SHOW TABLES；
SHOW COLUMNS FROM  tableA;  ==  desc tableA;
SHOW TABLE STATUS  FROM RUNOOB;
SHOW TABLE STATUS from RUNOOB LIKE 'runoob%'\G;   # 加上 \G，查询结果按列打印
CREATE TABLE table_name (column_name column_type);
SHOW CREATE TABLE table_name\G
DROP TABLE table_name ;
CREATE TABLE tableA (
   id INT NOT NULL AUTO_INCREMENT,
   name VARCHAR(40) NOT NULL,
   url VARCHAR(40) NOT NULL,
   alexa INT,
   country VARCHAR(40) NOT NULL,
   PRIMARY KEY ( id ))ENGINE=InnoDB DEFAULT CHARSET=utf8;
数据记录：
DELETE FROM table_name [WHERE Clause]
    DELETE FROM runoob_tbl WHERE runoob_id=3;
SELECT column_name,column_name
    FROM table_name
    [WHERE Clause][LIMIT N][ OFFSET M]
  你可以使用 LIMIT 属性来设定返回的记录数。
  你可以通过OFFSET指定SELECT语句开始查询的数据偏移量。默认情况下偏移量为0。
INSERT INTO table_name ( field1, field2,...fieldN )
                       VALUES
                       ( value1, value2,...valueN )，
                       ( value1, value2,...valueN );

INSERT INTO tableA (name, url, alexa, country)
    VALUES 
    ('Google','https//www.google.com/','1','usa'),
    ('淘宝','http://wwww.taobao.com/','13','cn'),
    ('菜鸟教程','http://www.runoob.com/','4689','cn'),
    ('weibo','http://weibo.com/','20','cn'),
    ('Facebook','https://www.facebook.com/','3','USA'),
    ('baidu','https://www.baidu.com/','4','CN');

UPDATE table_name SET field1=new-value1, field2=new-value2 [WHERE Clause]
UPDATE t_game_virtual_machine_info  SET sub_area_type=89 WHERE id BETWEEN 16 and 20;
update t_game_virtual_machine_info SET sub_area_type=3 where vmid BETWEEN 171 and 240;

索引：
CREATE INDEX indexName ON mytable(username(length));
ALTER table tableName ADD INDEX indexName(columnName)
DROP INDEX [indexName] ON mytable;
SHOW INDEX FROM runoob_tbl;

授权：
GRANT SELECT,INSERT,UPDATE,DELETE,CREATE,DROP ON mi.* TO 'igor'@'%' IDENTIFIED BY 'zara123' WITH GRANT OPTION;
WITH GRANT OPTION 这个选项表示该用户可以将自己拥有的权限授权给别人。

mysql> GRANT SELECT,INSERT,UPDATE,DELETE,CREATE,DROP
    -> ON TUTORIALS.*
    -> TO 'zara'@'localhost'
    -> IDENTIFIED BY 'zara123';

grant all privileges on *.* to igo_rw@"%" identified by 'mi';
grant select on *.* to  igo_ro@'%' identified by 'mi';

use mysql;
set password for root@localhost = password('igozhang');

FLUSH   PRIVILEGES;
```
