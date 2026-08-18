# RHEL7安装MQ8

```
ENV
RHEL7.3/JDK1.7.0_80/MQ8.0
配置java
export JAVA_HOME=/sxapp/sxappopt/mq8.0/jdk1.7.0_80
export JRE_HOME=$JAVA_HOME/jre
export CLASSPATH=.:$JAVA_HOME/jre/lib/rt.jar:$JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib/tools.jar
export PATH=$PATH:$JAVA_HOME/bin:$JRE_HOME/bin
需求：
队列管理器名：QM_MI
服务器连接通道：CHN_MI
端口：8927
接收队列名：QUEUE_RECV
发送队列名：QUEUE_REPLY
MQ用户名：mqm

安装：
最先安装证书「mqlicense.sh」, 否则后面的都安装不了，按1accept；
# mkdir /mqsoft;
# tar zxvf MQ_8.0.*.tar.gz -C /mqsoft
# cd /mqsoft
# ./mqlicense.sh
# 直接输入数字 1, 然后按回车即可同意.
# 出现下面这句话就算是接受协议了, 后面就可以开始安装软件了
...Agreement accepted: Proceed with install....

接下来安装的是 MQSeriesRuntime , 这个也要先安装
# rpm -ivh MQSeriesRuntime-8.0.0-0.x86_64.rpm

接下来就是安装剩下的了，实验证明在SUSE11SP4下只需要安装SDK，Server，Msg_Zh_CN,，Man就可以用；
# yum -y install *.rpm

安装完会自动添加用户mqm，我们给他设置密码
# passwd mqm;

配置：
配置mqm环境变量
vi /etc/profile
export PATH=/opt/mqm/samp/bin:/opt/mqm/bin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/X11R6/bin:$PATH
cd /opt/mqm/bin
1.创建缺省的的MQ队列管理器
./crtmqm -q QM_mi
2.启动队列管理器
./strmqm QM_mi
3.运行MQ命令
./runmqsc QM_mi
4.定义监听器(在控制台执行)
DEFINE LISTENER(MI) TRPTYPE(TCP) PORT(9827) CONTROL(QMGR) REPLACE
5.启动监听器
START LISTENER(MI)
6.定义服务器连接通道
DEFINE CHANNEL(CNN_MI) CHLTYPE(SVRCONN) MCAUSER('mqm') REPLACE
7.设置MCA用户标识
ALTER CHANNEL(CNN_MI) CHLTYPE(SVRCONN) MCAUSER('mqm')
8.启动通道
START CHANNEL(CNN_JACK)
9.定义队列
DEFINE QL(QUEUE_RECV) REPLACE
DEFINE QL(QUEUE_REPLY) REPLACE
10.关闭授权队列管理器（MQ8.0需要关闭）
ALTER QMGR CHLAUTH(DISABLED) CONNAUTH('')
11.刷新
REFRESH SECURITY TYPE(CONNAUTH)
12.退出
end
13.停止队列管理器
$ endmqm -i QM_mi

导出配置：
参考：
https://www.ibm.com/developerworks/cn/websphere/library/techarticles/1312_hub_mq/1312_hub_mq.html
https://www.cnblogs.com/lichmama/p/4272826.html
源端导出，mqm导出：
su - mqm
cd qmgrs/
dmpmqcfg -m QMEMBFE -a > QMEMBFE.dump
dmpmqcfg -m QMEMBFE -o setmqaut > QMEMBFE.aut
目标端导入，mqm导入：
crtmqm QMEMBFE
strmqm QMEMBFE
runmqsc QMEMBFE < QMEMBFE.dump
setmqaut -m QMEMBFE -n GSYHQ1 -t queue -g mqm +browse +chg +clr +dlt +dsp +get +inq +put +passall +passid +set +setall +setid

MQ常用命令:（）
dspmq
con>   dis qmgr
con>   display listener(*)
con>   display listener(listen_name)
con>   
con>   

https://blog.csdn.net/csdn_v_for_vendetta/article/details/84066351
查看版本
dspmqver

1、队列管理器
1.1创建队列管理器
命令：crtmqm <qmgrName>
示例：
crtmqm -ld /var/mqm/log -lc -lf 16384 -lp 25 -ls 5 QM_TEST
参数说明：
-ld 日志文件目录
-lc 使用循环日志
-lf 日志文件页大小
-lp 逐日之文件个数
-ls 辅助日志文件个数
QM_TEST 队列管理器名
1.2 启动队列管理器
命令：strmqm <qmgrName>
停止队列管理器
命令：endmqm -c|-w|-i|-p <qmgrName>

https://blog.csdn.net/weixin_33890499/article/details/87218224
```
