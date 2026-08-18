# python排错

```
env
CentOS Linux release 7.6.1810 (Core) 
Python 3.6.8
```

```
1.
python文本不能打印中文:
此问题python2.7有，python3.6没有
# python t.py
  File "t.py", line 2
SyntaxError: Non-ASCII character '\xe4' in file t.py on line 2, but no encoding declared; see http://www.python.org/peps/pep-0263.html for details

solution:
顶格添加一行
# -*- coding: utf-8 -*-
```

```
2.
文件名与导入包名相同导致报错：
AttributeError: module 'multiprocessing' has no attribute 'Process'

solusion:
修改python文件名解决
mv multiprocessing.py  multitask.py
```
