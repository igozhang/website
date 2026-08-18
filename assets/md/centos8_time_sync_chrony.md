# Centos8时间同步，chrony

```
centos使用chrony进行时间同步

env
CentOS Linux release 8.2.2004

默认安装，也可以手动
yum install -y chrony
systemctl start chronyd
systemctl enable chronyd

时区设置
timedatectl
rm -rf /etc/localtime
ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

配置chrony
主 192.168.3.201
vim /etc/chrony.conf
pool ntp1.aliyun.com iburst
allow 192.168.0.0/16

从
pool 192.168.3.201 iburst
效果
[root@igo-k8s-3 ~]# vim /etc/chrony.conf 
[root@igo-k8s-3 ~]# systemctl restart chronyd
[root@igo-k8s-3 ~]# chronyc sources
210 Number of sources = 1
MS Name/IP address         Stratum Poll Reach LastRx Last sample               
===============================================================================
^? 192.168.3.201                 0   7     0     -     +0ns[   +0ns] +/-    0ns
[root@igo-k8s-3 ~]# chronyc tracking
Reference ID    : 00000000 ()
Stratum         : 0
Ref time (UTC)  : Thu Jan 01 00:00:00 1970
System time     : 0.000000000 seconds slow of NTP time
Last offset     : +0.000000000 seconds
RMS offset      : 0.000000000 seconds
Frequency       : 29.014 ppm slow
Residual freq   : +0.000 ppm
Skew            : 0.000 ppm
Root delay      : 1.000000000 seconds
Root dispersion : 1.000000000 seconds
Update interval : 0.0 seconds
Leap status     : Not synchronised
[root@igo-k8s-3 ~]# 

参考命令
chronyc tracking
chronyc sources
chronyc -a makestep
date -s "2021-08-07 10:35:35"
ntpdate ntp1.aliyun.com
```
