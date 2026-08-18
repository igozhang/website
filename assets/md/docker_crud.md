# docker_crud

```
https://www.runoob.com/docker/docker-command-manual.html
https://cloud.tencent.com/developer/article/1772136

想要安装指定镜像版本去hub搜索
https://hub.docker.com/
```

```
docker pull elasticsearch:7.17.2
docker pull mysql:5.7.39
停止所有容器运行： docker stop $(docker ps -a -q)
删除所有停止运行的容器： docker rm $(docker ps -a -q)

docker save pinpointdocker/pinpoint-mysql:2.3.3 >pinpoint-mysql.tar
docker load<pinpoint-mysql.tar
```

```
docker run -p 8000:80 -it ubuntu /bin/bash
使用宿主机8000端口映射容器80端口

docker run -itd --name igo-mysql -p 2206:3306 -e MYSQL_ROOT_PASSWORD=123456 mysql:5.7.39

docker run \
-d \
--name igo_es \
-p 9200:9200 \
-p 9300:9300 \
-e ES_JAVA_OPTS="-Xms512m -Xmx512m" \
-e "discovery.type=single-node" \
elasticsearch:7.17.2
```

```
docker exec -it igo-mysql /bin/bash
docker exec -it igo-mysql /bin/ls
```

```
容器生命周期管理
run
start/stop/restart
kill
rm
pause/unpause
create
exec
容器操作
ps
inspect
top
attach
events
logs
wait
export
port
容器rootfs命令
commit
cp
diff
镜像仓库
login
pull
push
search
本地镜像管理
images
rmi
tag
build
history
save
load
import
info|version
info
version
```
