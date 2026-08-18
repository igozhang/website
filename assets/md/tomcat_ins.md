# tomcat8.5_ins

tomcat8.5安装部署

## 环境

java1.8  
yum -y install java-1.8.0-openjdk java-1.8.0-openjdk-devel  
CentOS Linux release 7.8.2003 (Core)

## 安装部署

1. 创建用户  
   useradd -r -m -U -d /opt/tomcat -s /sbin/nologin tomcat  
   为 Tomcat 目录设置权限  
   chown -R tomcat:tomcat /opt/tomcat  
   chmod -R 755 /opt/tomcat  
   -r：创建一个系统用户，系统用户通常用于运行系统服务，其 UID 一般小于 1000。  
   -m：为用户创建家目录，这里家目录指定为 /opt/tomcat。  
   -U：创建与用户同名的组。  
   -d /opt/tomcat：指定用户的家目录为 /opt/tomcat。  
   -s /sbin/nologin：指定用户的登录 shell 为 /sbin/nologin，表示该用户不能登录系统
2. 安装

```
# 创建一个目录用于存放 Tomcat
mkdir -p /opt/tomcat
# 下载 Tomcat 8.5 压缩包
wget https://archive.apache.org/dist/tomcat/tomcat-8/v8.5.81/bin/apache-tomcat-8.5.81.tar.gz -P /opt
# 解压压缩包到 /opt/tomcat 目录
tar -zxvf /opt/apache-tomcat-8.5.81.tar.gz -C /opt/tomcat8 --strip-components=1
# --strip-components=1 表示在解压过程中，去除压缩包内文件和目录路径中的最顶层目录apache-tomcat-8.5.81/
```

3. 添加服务

```
tee >/etc/systemd/system/tomcat.service <<EOF
[Unit]
Description=Apache Tomcat Web Application Container
After=syslog.target network.target

[Service]
Type=forking

Environment=JAVA_HOME=/usr/lib/jvm/jre
Environment=CATALINA_PID=/opt/tomcat/temp/tomcat.pid
Environment=CATALINA_HOME=/opt/tomcat
Environment=CATALINA_BASE=/opt/tomcat
Environment='CATALINA_OPTS=-Xms512M -Xmx1024M -server -XX:+UseParallelGC'
Environment='JAVA_OPTS=-Djava.awt.headless=true -Djava.security.egd=file:/dev/./urandom'

ExecStart=/opt/tomcat/bin/startup.sh
ExecStop=/opt/tomcat/bin/shutdown.sh

User=tomcat
Group=tomcat
UMask=0007
RestartSec=10
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
```

4. 配置环境

```
tee >/etc/profile.d/tomcat.sh <<EOF
export CATALINA_HOME=/opt/tomcat
export PATH=$PATH:$CATALINA_HOME/bin
EOF

source /etc/profile.d/tomcat.sh
```

5. 启动管理服务  
   启动 Tomcat 服务  
   systemctl start tomcat  
   设置 Tomcat 服务开机自启  
   systemctl enable tomcat  
   检查 Tomcat 服务状态  
   systemctl status tomcat
6. 防火墙

```
firewall-cmd --zone=public --add-port=8080/tcp --permanent
firewall-cmd --reload
```
