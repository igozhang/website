# fastapi_ins01

```
搭建一个fastapi服务器
监听本地8888端口
GET访问服务器/file01路径,返回"D:\cursor\ex.txt"文件内容
```

```
ENV:
Python 3.13.4
Windows11
```

```
1. 先安装服务器包：pip install fastapi uvicorn

2. 创建main.py
from fastapi import FastAPI
from fastapi.responses import PlainTextResponse

app = FastAPI()

@app.get("/file01")
async def read_file():
    try:
        with open("D:/cursor/ex.txt", "r", encoding="utf-8") as file:
            content = file.read()
        return PlainTextResponse(content)
    except FileNotFoundError:
        return PlainTextResponse("文件未找到", status_code=404)
    except Exception as e:
        return PlainTextResponse(f"发生错误: {str(e)}", status_code=500)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8888)

3. 启动服务器
python main.py
```

```
脚本后台运行：
先创建批处理脚本

@echo off
start /B python main.py
echo 服务器已在后台启动，监听端口 8888

然后双击运行
```

```
查看效果BYcursor
```
