# 速查百宝箱

```
zabbix批量安装
获取行号：
$ grep -n '</host>' zbx_import_hosts.xml | cut -d ":" -f 1
grep -vE '^#|^$' /usr/local/zabbix/conf/zabbix_agentd.conf

cpu个数，cpu数量,mem数量
nproc
grep 'processor' /proc/cupinfo | wc -l
lscpu
lsmem, free -g

chrome截图 screenshot
ctrl+shift+i
ctrl+shift+p
full

检查所有接口的物理链路状态
for i in $(ip link show | awk -F: '/^[0-9]+:/ {print $2}'); do 
  echo -n "$i: "; 
  sudo ethtool $i | grep "Link detected"
done

获取行号，并在该行下插入一个host段，然后替换成需要的IP：
$ point=`grep -n '</host>' zbx_import_hosts.xml | cut -d ":" -f 1|tail -1`    //最后一个标签所在行；
$ sed -i "$point  r host.txt" zbx_import_hosts.xml                //host代码块文件：host.txt；
$ sed -i 's/point4ip/8.8.8.8/g' zbx_import_hosts.xml              //将host代码块中ip位置替换成指定ip；

找到nginx.conf文件包含'currentVer'的文本行,将该行中'was1'替换成'was2'并将该行中'asy_sc_175'替换成'asy_sc_184'
egrep 'lastVer|currentVer' nginx.conf
sed -i '/currentVer/s/was1/was2/g;/currentVer/s/asy_sc_175/asy_sc_184/g' nginx.conf
sed -i '/lastVer/s/was2/was1/g;/lastVer/s/asy_sc_184/asy_sc_175/g' nginx.conf

整体封装，制作zabbix导入模板：
$ cat insert.and.replace.sh
#!/bin/bash
for ip in $(cat sit1.last_version.ip); do
echo $ip
point=`grep -n '</host>' zbx_import_hosts.xml | cut -d ":" -f 1|tail -1`
sed -i "$point r host.txt" zbx_import_hosts.xml
sed -i "s/point4ip/$ip/g" zbx_import_hosts.xml;
done
//涉及到的文件host.block.txt,zbx_import_hosts.xml，一个是zabbix机器的导出文件，另一个属于导出文件里重复代码块<host>

9立刻强制，15 优雅退出
pkill -9 -f java
pkill -15 -f java
ps -ef|grep java | grep -v grep | awk ‘{print $2}’ | xargs kill -9

tar --exclude={'logs','g-ofp-app.log','nohup.out'} -cvf dep0813.tar deploy

查看线程
top -H -p $pid  ##H显示线程信息
ls /proc/$pid/task/
ps -T -p $pid
htop -> F2 -> 打开线程树
线程数
top; f(打开功能界面,按d选择nTH)
ps -o thcount $pid
cat /proc/$pid/status | grep Threads

查看CPU占用高线程,再结合threadump查看具体内容,容器包括内置tomcat父进程ppid都是1
ps H -eo user,pid,ppid,tid,time,%cpu,cmd --sort=%cpu |grep $pid
printf %x $tid (10转16进制再jsatck文档中搜索)
printf %o 25433 10转8进制
printf %d 0xF   16转10进制

查当前cpu,mem高消耗进程
ps -eo pid,ppid,cmd,%cpu --sort=-%cpu | head
ps -eo pid,ppid,cmd,%mem --sort=-%mem | head

查CPU频率
cat /proc/cpuinfo |grep MHz|uniq
查27号cpu队列情况
sar -q -f /var/log/sa/sa27
查27号cpu/mem/swap/disk使用情况
sar -u -f /var/log/sa/sa27
sar -r -f /var/log/sa/sa27
sar -S -f /var/log/sa/sa27
sar -d -f /var/log/sa/sa27
查当前间隔1秒查3次
sar -q 1 3

添加免密
ssh keygen
ssh-copy-id ops@$dst_ip

字符
locale
vim :set encoding=utf-8
LANG=en_US.UTF-8
LANG=zh_CN.UTF-8

中文乱码
vim ~/.vimrc
set termencoding=utf-8
set encoding=utf8
set fileencodings=utf8,ucs-bom,gbk,cp936,gb2312,gb18030

粘贴问题，粘贴补齐，乱码
:set paste

nmon -ft -s5 -c 720  (会自动后台运行)
间隔5秒,采样720次(1小时),-f标准格式输出,-t包含高占用进程

alias ll='ls -l --color=auto'
alias mv='mv -i'

wget http://172.22.240.207/zabbix/zabbix_agent_install.sh

文件描述符的限制(连接数,文件打开数都是这个)
ulimit -n 用户级别/etc/security/limits.conf
/proc/$pid/limits 进程级别
/proc/sys/fs/file-max 系统级别
/proc/sys/fs/nr_open  内核级别

后台监听7001端口，无需打印nohup日志
nc -l -p 7001 &

统计tcp连接及状态
netstat -n | awk '/^tcp/ {++S[$NF]} END {for(a in S) print a, S[a]}'
当前系统连接数
ss -s
netstat -an | awk '/^tcp/ {++s[$NF]} END {for(a in s ) print a,s[a]}'
各端口连接数并排序取前10
netstat -an | grep ESTABLISHED | awk '{print $4}' | awk -F':' '{print $NF}' | sort | uniq -c | sort -nr |head -10
对端IP连接数排序取前
ss -t -a | awk '{print $5}' | cut -d: -f1 | sort | uniq -c | sort -nr |head
netstat -nt | awk '{print $5}' | cut -d: -f1 | sort | uniq -c | sort -nr |head

windows查端口
netstat -ano |findstr :9100
Get-NetTCPConnection -State Listen

cat -v igo.sh (将win格式换行符转为unix格式换行符)
dos2unix igo.sh

进程打开文件
lsof -p $pid
进程打开文件数
lsof -p $pid |wc -l
所有文件打开数
lsof |wc -l
所有文件打开数并排序(打开数,命令,进程号)
lsof -n | awk '{print $1, $2}' | sort | uniq -c | sort -nr |head -15
端口连接数
lsof -ni tcp:29200 |awk '{print $1,$2}' |sort |uniq -c |sort -nr

lsof +w /igofile 查文件占用进程
strace -p $pid 查进程访问文件
ls -l /proc/$pid/exe 根据进程号查可执行文件位置

文件统计去重排序
sort igofile.txt | uniq -c | sort -nr
(先sort排序,再uniq统计去重,再sort排序)

tee >/etc/hostname <<EOF
igo_222
EOF

磁盘格式化挂载fdisk lsblk
parted -s  /dev/sda mklabel gpt mkpart primary ext4 0% 100%
mkfs.ext4 /dev/sda1
echo "/dev/sda1  /data  ext4  defaults,noatime  0 0" >> /etc/fstab
mount -a

logger -p auth.notice "This is a test message with priority"

带账密限速下载
wget --http-user=igo  --http-password=igo@123 --limit-rate=10M http://172.24.240.83:8080/kvm-template/es-template.qcow2
curl -OL -uigo:igo@123 --limit-rate=20M http://igoZhang.cn:8080/igo/4.tar

ps -eo pid,lstart,cmd |grep java

获取mac地址：
cat /sys/class/net/eth0/address
获取ip地址
pp=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')
获取公网IP
curl ifconfig.me

获取宿主机上虚机资源信息
# virsh list|awk '/running/{print $2}'|xargs -t -i virsh dominfo {} |grep -iE 'id|name|cpu|mem'
虚机CPU求和：
virsh list|awk '/running/{print $2}'|xargs -i virsh dominfo {} |grep  CPU\(s\)|awk '{print $2}'|awk '{sum +=$1};END {print sum}'
虚机内存求和：
expr `virsh list|awk '/running/{print $2}'|xargs -i virsh dominfo {} |grep "Used\ memory"|awk '{print $3}'|awk '{sum +=$1};END {print sum}'` / 1024 / 1024

虚机资源，虚机CPU，虚机内存
vm=nsq
virsh list |awk '/running/{print$2}'|grep $vm; hostname;
echo `virsh list |awk '/running/{print$2}'|grep $vm|xargs -i virsh dominfo {} |grep CPU\(s\)|awk '{print $2}'`C;
expr `virsh list |awk '/running/{print$2}'|grep $vm|xargs -i virsh dominfo {} |grep Used|awk '{print $3}'` / 1024 / 1024;

扩容
virsh setvcpus 4010-network-monitor-01 2 --live --config
virsh setmem  4010-network-monitor-01 4G --live --config
# 创建
qemu-img create -f qcow2 vm_name.qcow2 300G
# 添加
virsh attach-disk vm_name --source /data/vm_name.qcow2 --target vdb --persistent --subdriver qcow2
# 移除
virsh detach-disk ks-k8s-node-3 vdb
resize2fs /dev/vda1  看不出区别，去运行这条命令的话磁盘跑满报错

debian磁盘扩容

debian 安装显卡驱动
apt install nvidia-settings

pvcreate /dev/vda
vgs
vgextend centos /dev/vda
lvs
lvextend -l +100%FREE /dev/centos/root
df -Th
resize2fs /dev/centos/root
xfs_growfs /dev/centos/root

10G大文件
dd if=/dev/zero of=$(date +%H%M%S).10gfile bs=1M count=10240

windows查激活
slmgr.vbs -xpr  查激活到期
slmgr.vbs -dlv  查激活详细
slmgr /ato   激活

时区
# rm -rf /etc/localtime
# ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
timedatectl set-timezone Asia/Shanghai

逐行读文件
cat igo1113.txt |xargs -n 1 ping -c2;
for dom in $(cat igo1113.txt);do echo $dom;done
while IFS= read -r dom; do echo "$dom"; done < igo1113.txt

ipmi配置IP
ipmitool lan set 1 ipsrc static
ipmitool lan set 1 ipaddr 172.17.242.209
ipmitool lan set 1 netmask 255.255.255.0
ipmitool lan set 1 defgw ipaddr 172.17.242.254
ipmitool user list 1

ipmitool lan print  查看IP
ipmitool user list 1  列出ipmi用户
ipmitool sel list命令可以获取ipmi event log

删除重复行，去重：
sort -u sit1.last.ip.3 >s3.ip      //去重并排序
sort sit1.last.ip.2|uniq >s2.ip      //去重并排序

java环境变量配置
vim /etc/profile       # 文本末尾追加两行
export JAVA_HOME=/usr/local/jdk1.8.0_131
export PATH=.:$PATH:$JAVA_HOME/bin

ssh-copy-id
# cat ~/.ssh/id_*.pub|ssh root@10.1.69.65 'cat>>.ssh/authorized_keys'
批量公钥，批量下发公钥
#for i in $(cat push_pub_key.ip); do cat ~/.ssh/id_*.pub|ssh root@$i 'cat>>.ssh/authorized_keys'; done

使用sshpass工具批量下发公钥
# for i in $(cat sit1.enable.ip); do echo $i; cat ~/.ssh/id_*.pub|sshpass -p "root@12345" timeout 1 ssh -o stricthostkeychecking=no root@$i 'cat>>.ssh/authorized_keys'; done

不用文件的话，这样写：
# for i in {1..5};do ip="10.7.36.$i"; echo $ip; cat ~/.ssh/id_*.pub|sshpass -p "root@12345" ssh -o stricthostkeychecking=no root@$ip 'cat>>.ssh/authorized_keys';done

ssh远程执行命令，ssh远程下发命令：
ssh 10.7.36.19 "su -  sxappopt -c /sxapp/sxappopt/bin/start.sh"

sed清除“()[]”，sed清除乱码：
 ##echo "(([10(.1(6.8]]2.[33)"|sed -e 's/['\]'()\[]//g'
效果： 10.16.82.33

sed批量替换
sed -i 's/10.1.128.46/10.7.36.50/g' `grep 10.1.128.46 -rl *
sed匹配行头，替换整行内容
sed -i '/^cloud_server_ip/ccloud_server_ip = update.skyeye.360safe.com' name.txt

shell上一条命令执行结果
if [ $? -eq 0 ]; then
     echo "succeed"
else
     echo "failed"
fi
-eq  等于
-ne  不等于

制作光盘：
mkisofs -r -J -o conf.iso conf2020/

添加
iptables -A INPUT -p tcp --dport 80 -j DROP
iptables -A INPUT -p tcp --dport 80 -j REJECT –reject-with tcp-reset
-j REJECT –reject-with tcp-reset比 -j DROP好，干脆利落，节约时间，节约带宽

先DROP所有,再根据需要-I追加需要开放的网段
iptables -I INPUT -p tcp --dport 7003 -j REJECT --reject-with tcp-reset
iptables -I INPUT -s 172.18.14.0/24 -p tcp --dport 7003 -j ACCEPT
service iptables save
RHEL7,CENT7中报错如下两种方法解决:
yum -y install iptables-services
或者
/usr/libexec/iptables/iptables.init save

删除
iptables -D INPUT -p tcp --dport 80 -j DROP
或者先找到规则编号，然后删除规则编号(1)
iptables -L -n --line-number
iptables -D INPUT 1

firewall-cmd --zone=public --add-port=80/tcp --permanent
firewall-cmd --add-port=80/tcp --permanent
firewall-cmd --zone=public --list-ports
firewall-cmd --reload

history格式，history加时间，用户，IP： vim /etc/profile ，编辑完source生效
USER_IP=`who -u am i 2>/dev/null| awk '{print $NF}'|sed -e 's/[()]//g'`I
export HISTTIMEFORMAT="[%F %T][`whoami`][${USER_IP}] "
（CentOS7,测试通过）
export PATH=/bin:/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
清理：  history -c

vm开vnc
<graphics type='vnc' port='5916' autoport='no' listen='0.0.0.0'>
<listen type='address' address='0.0.0.0'/>
</graphics>
vnc连kvm物理机加端口

shell取ip，shell获取ip：
/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"

echo改密码
echo `echo '!QAZ2wsx' | passwd --stdin oemchk`

linux物理机查序列号
dmidecode -t 1

查数据库版本
su - oracle
sqlplus
which sqlplus | xargs -i file {};看 32/64位
sqlplus system/$pwd@$hostname:1521/$servicename

find ./ -type f -exec ls -l {} \;
find -name "*.log" -exec mv {} test \;
find / -size +1M -type f -print  查找超过1M的文件
find . -type f -size +100M 查大文件
find . -type f -size +100M  -print0 | xargs -0 du -h | sort -nr
find /path/to/directory -type f -mtime -3 -exec ls -l {} \; 查最近2天更新文件
du -hm --max-depth=2 | sort -nr | head -20
找出/data目录最大的8个文件并排序的shell命令
ls -lSr /data | head -8
递归找出/data目录最大的8个文件并排序的shell命令
find /data -type f -exec du -Sh {} + | sort -rh | head -8
crontab -e
每凌晨3.删除7天前日志文件
0 3 * * * find /data/logs -type f -mtime +6 -exec rm -f {} \;
每5分钟运行一次
*/5  * * * * command
每小时检查清理大于10G文件
* */1 * * * find $target_dir -type f -size +10G -exec truncate -s 0 {} \;

sysctl -w net.core.rmem_max=212992 设置值
vi /etc/sysctl.conf  配置文件设置值
sysctl -p  生效(从配置文件“/etc/sysctl.conf”加载内核参数)
所有sysctl内核参数都在目录/proc/sys中；
记录net.core.rmem_max的文件就是/proc/sys/net/core/rmem_max；
查看值：
sysctl net.core.rmem_max
cat /proc/sys/net/core/rmem_max
sysctl -a | grep rmem_max

linux文件释放
lsof | grep deleted
然后根据进程号
cd /proc/进程号/fd/；
echo>2；

java -jar -Xms28g -Xmx28g C:\Users\Administrator\Desktop\ha457.jar

带时间戳，ping
ping 172.27.240.168  | awk '{ print $0"\t" strftime("%Y-%m-%d %H:%M:%S",systime())}'

添加sudo，5/6版本
chattr -i /etc/sudoers; chmod u+w /etc/sudoers ;echo "oemchk ALL=NOPASSWD:/bin/su - " >>/etc/sudoers ; chmod u-w /etc/sudoers ; chattr +i /etc/sudoers
添加sudo，7版本
chmod -v u+w /etc/sudoers
vim /etc/sudoers； test ALL=(ALL)  ALL； 如果不想需要输入密码，将最后一个 ALL 修改成 NOPASSWD: ALL 。
chmod -v u-w /etc/sudoers
或者：
usermod -aG wheel ops

rhel7配dns,centos7配dns，resolv.conf配置，rhel7双网卡
# vim /etc/sysconfig/network-scripts/ifcfg-ens37      //在不需要DNS的网卡上关闭PREDNS，系统就会自动获取指定网卡的DNS配置；
PEERDNS=no

查命令安装包，查安装包，命令查包：
# yum provides "nc"
#yum install --downloadonly --downloaddir=/igo/elk elasticsearch
# rpm -qf /usr/bin/nc
# rpm -qa |grep docker
# which scp
apt-cache madison vim
yum list docker-ce --showduplicates
yum -y install docker-ce-18.06.3.ce-3.el7
yum -y install nc --nogpgcheck

tcpdump -S host 10.21.120.247 and 10.21.81.17  -w /tmp/igo0418.cap
tcpdump -i eth0 -w /tmp/igo0418.cap
过滤所有经过ens32上目的或者源地址为192.168.1.1上的数据
tcpdump -i ens32 host 192.168.1.1
过滤所有经过ens32上源地址为192.168.1.1的数据，只有源地址为192.168.1.1的单向数据
tcpdump -i ens32 src host 192.168.1.1
过滤所有经过ens32上目的地址为192.168.1.1的数据，只有目的地址为192.168.1.1的单向数据
tcpdump -i ens32 dst host 192.168.1.1
抓取与主机192.168.1.1通信的，且不包含192.168.1.2的数据包（.2和.1也通信）
tcpdump -i ens32 ip host 192.168.1.1 and !192.168.1.1
抓源端口或者目的端口为8080的所有数据包
tcpdump -i ens32 port 8080
抓源端口为22的数据包
tcpdump -i ens32 src port 22

抓119.29.121.116的80端口和110和25以外的其他端口的包
tcpdump -i eth0 host 119.29.121.116 and ! port 80 and ! port 25 and ! port 110 -w /tmp/xxx.cap
抓vlan 1的包
tcpdump -i eth0 port 80 and vlan 1 -w /tmp/xxx.cap
过滤arp包
tcpdump -i ens32 arp
过滤 ip包
tcpdump -i ens32 ip
过滤tcp包
tcpdump -i ens32 tcp
过滤udp包
tcpdump -i ens32 udp
https://www.cnblogs.com/wongbingming/p/13212306.html ‘全网最详细的 tcpdump 使用指南’
过滤来自进程名为 nc 发出的流经 en0 网卡的数据包，或者不流经 en0 的入方向数据包
$ tcpdump "( if=en0 and proc =nc ) || (if != en0 and dir=in)"

iostat -d -k 2 3
只显示设备信息-d,以kilobyte没单位,每2秒刷新一次,取值3次;
第一条表示从系统开机到当前执行时刻的统计信息,可以-y跳过;
-x参数可以统计一些io扩展信息;
iostat -c 1 3 查看cpu状态值

查日志,查文件,长日志,大日志
igofile=/home/ops/igo.log
找出日志文件中最长日志所在行号
awk '{ print length(), NR }' /data/logs/access.log | sort -nr | head -n1 | awk '{print $2}'
查指定行(第100行)字节数
sed -n '100{p;q;}' "$igofile" | wc -c
查指定行(第100行)内容并写到igo.txt
sed -n 100p $igofile >/tmp/igo.txt

增量备份：
# rsync -az --delete SRC DST
其中：
-z 压缩
-a 简单理解就是保持一致性
--delete 严格保证DST内容与SRC一致，即DST中SRC没有的文件会被删除掉   （--delete-before表示在进行同步之前，先将目标目录全部删除，然后再进行同步操作）

压缩解压缩
gzip filename  # 压缩 filename，生成 filename.gz，原文件删除
gzip -c filename > filename.gz  # 压缩后保留原文件
gzip -d filename.gz  # 解压为 filename，压缩包删除

tar -zcvf archive.tar.gz /path/to/dir_or_files  # -c 打包，-v 显示过程，-f 指定压缩包名 -z gzip压缩
tar -jcvf archive.tar.bz2 /path/to/dir_or_files -j 打包并以 bzip2 压缩

zip -r archive.zip /path/to/dir_or_files  # -r 递归压缩目录
zip document.zip document.txt
unzip archive.zip

# tar -g snapshot -zcf backup_incremental_1.tar.gz test
利用tar -g参数，在第一次备份时候生成时间戳文件，里面包含指定备份目录下的所有文件的一个时间戳，下次增量备份，tar会利用时间戳文件去比较，只有那些内容在这段时间有修改的文件，才会被打包

-bash-4.2$ 问题，找到家目录，复制bash文件到家目录，提示符错误；
cp /etc/skel/.bash_logout     /var/mqm/
cp /etc/skel/.bash_profile    /var/mqm/
cp /etc/skel/.bashrc          /var/mqm/

端口检测
for pp in $(cat ip.list);do echo $pp; nc -vz $pp 9080;done
while read rows;do echo $rows; nc -nvz $rows;done <ip.list
telnet ip port
wget ip:port
ssh -v -p port root@ip
curl -v ip:port

nginx_accesslog
统计url访问量并统计排序
awk '{print $11}' /var/log/nginx/access.log|sort | uniq -c |sort -n -k 1 -r|more
统计IP访问量并统计排序
awk '{print $1}' /var/log/nginx/access.log|sort | uniq -c |sort -n -k 1 -r|more
找出最近5000条包含的报错日志
tail -5000 /var/log/nginx/access.log| awk '$(NF-4)>400 {print $0}' |more
统计访问时间并倒序排列(从大到小)
sort /var/log/nginx/access.log |uniq -c |sort -n -k 6 -r|more
统计访问时间并统计排序
awk '$5>4 {print $0}' /var/log/nginx/access.log |more

返回head标头
curl -I $url
检查代理连通性
curl -I -xnetproxy.igozhang.cn:3128 www.baidu.com 

java
查heap
bin/jhsdb jmap --heap --pid 255873

hosts
C:\Windows\System32\drivers\etc\hosts

时间
ntpdate ntp1.aliyun.com
chronyc -a makestep
date -s "2021-08-07 10:35:35"
date +%Y%m%d-%H%M%S
cp /etc/ntp.conf{,.$(date +%Y%m%d)}

~/.vimrc
syntax on
set showmode
set encoding=utf-8 
set autoindent
set noautoindent

lsof -n |awk '{print $2}'|sort|uniq -c|sort -nr|head
结果第一列为打开文件数,第二列为PID
lsof -p <pid>| awk '{print $4}' |grep "^[0-9]" |wc -l
查看实际打开FD数,第一个命令炒出来的可能是因为多进程共用文件导致的重复计数
ls /proc/$pid/fd  第二种方式确认pid打开fd
# file /proc/4065/fd/3
/proc/4065/fd/3: symbolic link to `/var/log/zabbix/zabbix_agent2.log'
# file /proc/4065/fd/1
/proc/4065/fd/1: broken symbolic link to `socket:[19975]'

useradd -M -s /sbin/nologin -g mysql mysql
useradd –g ftp –s /sbin/nologin ftpuser
useradd -d /data/igosmb -s /sbin/nologin smbuser
useradd deploy && echo 'dep@123' |passwd --stdin deploy
加sudo
usermod -aG wheel deploy
或者免密 /etc/sudoers
deploy ALL=(ALL) NOPASSWD:ALL
usermod -s /sbin/nologin trans_user
添加igo并sudo免密
useradd -m -s /bin/bash igo 2>/dev/null || true; echo 'igo:igo@123' | chpasswd; printf '%s\n' 'igo ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/igo; chmod 440 /etc/sudoers.d/igo
```

```
#!/bin/bash
echo "Hi, I'm sleeping for 5 seconds..."
sleep 5
echo "Hi, I'm sleeping for 2 mins..."
sleep 2m
```

```
时间
一月 January 二月 February 三月 March 四月 April
五月 May 六月 June 七月 July 八月 August 九月 September
十月 October  十一月 November  十二月 December

星期一：Monday 二Tuesday 三 Wednesday 四Thursday
星期五：Friday 六Saturday 日Sunday
```

```
journalctl is used to print the log entries stored in the journal by systemd-journald.service(8) and systemd-journal-remote.service(8)
journalctl -p err..alert
journalctl -f
jurnalctl -b
journalctl -u nginx.service
journalctl -u sshd
journalctl --since "today" | grep 'sshd'
journalctl --since "2023-03-03" --until "2023-03-03 03:00"
journalctl --since "2 hours ago" --until "1 hour ago"

grepm 
grep -A 10 '2024.06.18 17:54:00' igo.log 前10行
grep -B 10 '2024.06.18 17:54:00' igo.log 后10行
grep -C 10 '2024.06.18 17:54:00' igo.log 前后10行
grep -w '2024.06.18 17:54:00' 完整匹配

/var/log/messages  主日志
/var/log/dmesg  启动日志
tail ~/.bash_history  history文件位置
```
