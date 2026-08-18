# linux启用telnet登陆

```
Cent76环境验证通过
yum install xinetd telnet-server -y
vim /etc/securetty
底部添加四行：
pts/0
pts/1
pts/2
pts/3
启动服务
systemctl enable xinetd
systemctl enable telnet.socket
systemctl start telnet.socket
systemctl start xinetd
```
