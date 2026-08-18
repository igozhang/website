# py-redis

```
env
CentOS Linux release 7.9.2009 (Core)
Docker version 1.13.1, build 7d71120/1.13.1
Python 3.6.8 (default, Nov 16 2020, 16:55:22)

https://github.com/WoLpH/redis-py
```

```
docker run --name redis507 -p 6379:6379 -d redis:5.0.7
docker run -it --network host --rm redis:5.0.7 redis-cli
pip3 install redis
```

```
# cat redi.py
import redis   # 导入redis 模块

r = redis.Redis(host='localhost', port=6379, decode_responses=True)

if __name__== '__main__':
  r.set('name', 'igo')  # 设置 name 对应的值
  print(r['name'])
  print(r.get('name'))  # 取出键 name 对应的值
  print(type(r.get('name')))  # 查看类型
```
