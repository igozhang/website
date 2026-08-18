# pip_localnet

本地网络安装pip工具包，以pyzabbix举例

## 方法一 离线包安装

1. 在外网环境下载所需包

```
mkdir pyzabbix_packages
cd pyzabbix_packages
pip download pyzabbix -d .
```

2. 打包并上传

```
tar -cf pz.tar pyzabbix_packages
sz
rz
tar -xf pz.tar
```

3. 安装

```
pip3 install --no-index --find-links=. pyzabbix
```

## 方法二 搭建本地pypi镜像源

1. 搭建镜像源

```
pip install pypiserver
mkdir pypi_packages
pip download pyzabbix -d pypi_packages
启动server
pypi-server -p 8080 pypi_packages
```

2. 在内网服务器配置使用本地镜像源

```
pip install --index-url=http://<外网机器 IP>:8080 pyzabbix
```
