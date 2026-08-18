# jekins

```
Jenkins 2.319.1
CentOS Linux release 7.9.2009 (Core)
# docker version
Client:
 Version:         1.13.1
 API version:     1.26
 Package version: docker-1.13.1-209.git7d71120.el7.centos.x86_64
 Go version:      go1.10.3
 Git commit:      7d71120/1.13.1

Server:
 Version:         1.13.1
 API version:     1.26 (minimum version 1.12)
 Package version: docker-1.13.1-209.git7d71120.el7.centos.x86_64
 Go version:      go1.10.3
```

```
docker pull jenkins/jenkins:lts
docker run -d -p 8080:8080 -p 50000:50000 --restart=on-failure -v jenkins_home:/var/jenkins_home jenkins/jenkins:lts-jdk11

看到jenkins is fully up and running就可以访问jenkins并配置了
配置密码看容器输出或者 cat /var/jenkins_home/secrets/initialAdminPassword

http://192.168.3.211:8080/

安装推荐的插件
```

```
添加用户
修改简体中文：
安装：主界面-->系统管理-->插件管理（locale插件）
重启：http://192.168.3.211:8080/restart
```

