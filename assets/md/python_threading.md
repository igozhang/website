# python多线程并发

```
env
cent76
python36

进程是分配资源的最小单位
线程是程序执行的最小单位,是进程的一个实体
进程初始化会生成一个主线程，主线程默认顺序执行程序函数
线程并发可以在不增加资源的情况下，并发执行多个函数，实现提高效率的目的

进程可以使用多核
线程只能轮询使用单核
```

```
# cat multiThread.py
import threading
import time

def sing():
    for i in range(3):
        print("sing...")
        time.sleep(1)

def dance():
    for i in range(3):
        print("dance...")
        time.sleep(1)

if __name__ == "__main__":
    # sing()
    # dance()
    thread_sing = threading.Thread(target=sing)
    thread_dance = threading.Thread(target=dance)
    thread_sing.start()
    thread_dance.start()
	
结果:三秒执行完成6个动作
# python3 multiThread.py
sing...
dance...
sing...
dance...
dance...
sing...
```
