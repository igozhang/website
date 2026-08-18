# python多进程并发

```
env
cent76
python36
```

```
1.
单进程
# cat singletask.py 
import time

def sing():
  for i in range(3):
    print("唱歌...")
    time.sleep(0.5)

def dance():
  for i in range(3):
    print("跳舞...")
    time.sleep(0.5)

if __name__== '__main__':
  sing()
  dance()

结果：
# python3 singletask.py 
唱歌...
唱歌...
唱歌...
跳舞...
跳舞...
跳舞...
```

```
2.
python多进程
# cat multitask.py 
# 1. import multiprocessing
# 2. 指定对象名 “sing_process”以及对象所对应的函数名 “sing”
# 3. 启动多进程对象 “process.start()”

import time
import multiprocessing

def sing():
  for i in range(3):
    print("唱歌...")
    time.sleep(0.5)

def dance():
  for i in range(3):
    print("跳舞...")
    time.sleep(0.5)

if __name__== '__main__':
  sing_process = multiprocessing.Process(target=sing)
  dance_process = multiprocessing.Process(target=dance)

  sing_process.start()
  dance_process.start()

结果
# python3 multitask.py
唱歌...
跳舞...
唱歌...
跳舞...
唱歌...
跳舞...
```
