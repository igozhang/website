# arthas_flame

## env

以es为例，生成arthas火焰图分析CPU时间占用

CentOS Linux release 7.6.1810 (Core)  
elasticsearch-7.17  
arthas-boot version: 3.7.2

## 快速安装启动

1. java

```
yum -y install java-11-openjdk java-11-openjdk-devel
```

2. es

```
curl -OL https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.17.22-x86_64.rpm
yum -y install ./elasticsearch-7.17.22-x86_64.rpm
systemctl start elasticsearch.service
```

3. arthas

```
curl -O https://arthas.aliyun.com/arthas-boot.jar
java -jar arthas-boot.jar -h
java -jar arthas-boot.jar
```

## 生成火焰图

1. 修复es可登录

```
systemctl stop elasticsearch.service
mkdir -p /home/es
usermod elasticsearch -s shell
chsh -s /bin/bash elasticsearch
usermod -d /home/es elasticsearch
cp /etc/skel/.bash_logout     /home/es
cp /etc/skel/.bash_profile    /home/es
cp /etc/skel/.bashrc          /home/es
chown -R elasticsearch.elasticsearch /home/es
systemctl start elasticsearch.service
```

2. 登录arthas监控es  
   需要满足，同用户，同java，且打开security permission  
   打开permission  
   在/usr/share/elasticsearch/jdk/conf/security/java.policy最后添加：

```
permission java.security.AllPermission;
```

并重启es；

```
su - elasticsearch  ##同用户
/usr/share/elasticsearch/jdk/bin/java -jar arthas-boot.jar --target-ip 0.0.0.0 --username arthas --password Ab1234  ##同java
```

3. 生成火焰图  
   先给文件夹授予es用户权限，否则profiler stop失败（stop需要保存火焰图文件）

```
$ profiler start 启动采集（默认是CPU）
$ profiler getSamples 查看采样的数量
$ profiler status  查看采样的状态（是否在运行，运行了多久）
$ profiler stop 停止并生成火焰图
```

可以通过网页cosole访问火焰图  
http://h1.igozhang.cn:3658/arthas-output/

绿色部分代表Java代码  
黄色部分代表JVM C++代码  
橙色部分代表内核态C语言代码  
红色代表用户态C语言代码
