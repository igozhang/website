# ssh_troubleShooting

```
systemctl start sshd
service start ssh
均报错
Dec 23 11:54:02 debian systemd[1]: Failed to start OpenBSD Secure Shell server

/usr/sbin/sshd -T
no hostkeys available— exiting，则错误原因是ssh密钥出现问题

solution

ssh-keygen -t dsa -f /etc/ssh/ssh_host_dsa_key
ssh-keygen -t rsa -f /etc/ssh/ssh_host_rsa_key

chmod 600 /etc/ssh/ssh_host_dsa_key
chmod 600 /etc/ssh/ssh_host_rsa_key

在启动ssh服务就可以了
```
