# KQL,kibana查询语法

```
env
cent79
es7.9.3

note: KQL不区分大小写
https://www.elastic.co/guide/en/kibana/7.9/kuery-query.html
```

```
1. message:"error"
查询message中包含error的日志条目;

2. response:200
查询响应请求为200的日志条目;

3. 与或(and,all)
3.1 response:200 or extension:php
3.2 response:(200 or 404)
3.3 response:200 and extension:php
3.4 response:200 and (extension:php or extension:css)

4. 非(not)
4.1 not response:200
4.2 response:200 and not (extension:php or extension:css)

5. 范围(range)
5.1 account_number >= 100 and items_sold <= 200

6. 通配符
6.1 machine.os:win*

7. 嵌套查询
7.1
items:{ name:banana and stock > 10 }  //同一个文档
items:{ name:banana and stock:9 }
items:{ name:banana } and items:{ stock:9 } //多个文档
```
