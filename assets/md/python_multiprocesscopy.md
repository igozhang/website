# python多进程并发拷贝器

```
env
cent76
python36
初始源文件
mkdir -p /igo/source/
# for i in {4..15}; do echo $i;echo $i >/igo/source/$i.txt; done
```

```
# cat multiProcessCopy.py 
import os
import multiprocessing

def copy_file(file_name,source_dir,dest_dir):
# 1. 拼接文件路径
  source_path = source_dir + "/" + file_name
  dest_path = dest_dir + "/" + file_name

# 2. 读打开源文件，写打开目标文件
  with open(source_path, "rb") as source_file:
    with open(dest_path, "wb") as dest_file:

# 3. 循环读取源文件到目标路径
      while True:
        data = source_file.read(1024)
        if data:
          dest_file.write(data)
        else:
          break

if __name__ == '__main__':
  source_dir = "/igo/source"
  dest_dir = "/igo/dest"

  try:
    os.mkdir(dest_dir)
  except:
    print("dest_dir already exist")

  file_list = os.listdir(source_dir)
  for file_name in file_list:
    # 单进程
    # copy_file(file_name,source_dir,dest_dir)
    # 多进程
    sub_process = multiprocessing.Process(target=copy_file,args=(file_name,source_dir,dest_dir))
    sub_process.start()
```
