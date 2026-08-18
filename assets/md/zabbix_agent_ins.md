# zabbix-agent安装脚本，兼容Ubuntu16,18,20;CentOS7,8

```
zabbix_agent4.2.1
下载，校验，安装zabbix_agent并配置
```

```
#!/bin/bash
url='http://192.168.3.133'
zbxser='192.168.3.168'
md5str='e55ba94060ba2548ae8a1c29fd7cb7dd'
os='unknown'
ver='4.2.1'
log='/tmp/zabbix_agent_install.log'
zbxcfg='/etc/zabbix/zabbix_agentd.conf'
d=`date +%Y%m%d%H%M%S`

homedir=`env |grep ^HOME |awk -F '=' '{print $NF}'`
if [ ${homedir}x != '/rootx' ]
  then
     echo -e '\033[31mplease run with root,exit\033[0m'
     exit
fi

if [ $# == '1' ]
   then
       meta=$1
   else
      #meta='linux'
      echo -e '\033[31mInput error,Usage example:bash '$0' front|storage|phynode|vm|backend\033[0m'
      exit
fi

if [ "`grep -i ubuntu /etc/issue`" ]
   then
       os='ubuntu'
fi

if [ -e /etc/redhat-release ]
   then
      if [ "`grep 'CentOS Linux release 7' /etc/redhat-release`" ]
        then
            os='centos7'
      fi
fi

if [ $os == 'unknown' ]
   then
       echo 'Current system is not supporte,exit'
       exit     
fi

if [ -e /usr/sbin/zabbix_agentd ] && [ -e /etc/zabbix/zabbix_agentd.d/proc.conf ]
    then
        echo 'zabbix is exist'
        exit
fi

if [ ! -e /usr/sbin/zabbix_agentd ] 
   then     
       [ `cat /etc/passwd | grep zabbix` ] || useradd zabbix -s /sbin/nologin
       
       echo -ne '\033[32minstallation dependency...\033[0m'
       if [ $os == 'ubuntu' ]
          then
             cp /etc/apt/sources.list /etc/apt/sources.list_$d
             if [ "`grep 'archive.ubuntu.com' /etc/apt/sources.list`" ]  
               then
                  src=`grep 'archive.ubuntu.com' /etc/apt/sources.list |grep -v deb-src |head -1 |awk -F '/' '{print $3}'`
                  sed -i "s/$src/mirrors.aliyun.com/g" /etc/apt/sources.list
             fi
             sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list     
             apt-get update  >> $log 2>&1
             apt-get -y install libpcre3-dev supervisor gcc make iotop iftop fio sysstat lrzsz wget libssl-dev ipmitool fping >> $log 2>&1
          else
              yum -y install gcc gcc-c++ wget supervisor make unzip openssl-devel pcre-devel wget ipmitool fping >> $log 2>&1
       fi
       echo -e '\033[32mOK\033[0m'
       
       echo -ne '\033[32mdownload zabbix...\033[0m'
       wget $url/zabbix/zabbix-${ver}.tar.gz -o /dev/null -O /tmp/zabbix-${ver}.tar.gz
       echo -e '\033[32mOK\033[0m'
       
       zbxmd5=`md5sum /tmp/zabbix-${ver}.tar.gz |awk '{print $1}'`
       if [ ${md5str}x != ${zbxmd5}x ]
          then
              echo -e '\033[31mmd5str is wrong,exit\033[0m'
              exit
       fi
       
       echo -ne '\033[32minstall zabbix-agent...\033[0m'
       cd /tmp/
       tar xf zabbix-${ver}.tar.gz >> $log 2>&1
       cd zabbix-${ver}
       ./configure --enable-agent --sysconfdir=/etc/zabbix --sbindir=/usr/sbin >> $log 2>&1
       if [ $? != '0' ]
         then
            echo -e '\033[31mzabbix agent configure fail,exit\033[0m'
            exit
       fi
       make install >> $log 2>&1
       
       if [ -e /usr/sbin/zabbix_agentd ]
         then
           echo -e '\033[32mOK\033[0m'
         else
           echo -e '\033[31mfail,exit\033[0m'
           exit
       fi
  fi    
  
echo -ne '\033[32msetup zabbix...\033[0m'
wget $url/zabbix/zabbix.tar.gz -o /dev/null -O /tmp/zabbix.tar.gz
cd /tmp/
tar xf zabbix.tar.gz
[ ! -e /etc/zabbix ] || mv /etc/zabbix /etc/zabbix_$d
cp -a zabbix /etc/
chown -R zabbix:zabbix /etc/zabbix

[ -e /data ] || mkdir /data
touch /data/zbx.chk
chmod 666 /data/zbx.chk

chmod 640 /etc/sudoers
echo 'zabbix ALL=(ALL) NOPASSWD: /etc/zabbix/script/zbx_chk_ipmi.sh' >> /etc/sudoers
echo 'zabbix ALL=NOPASSWD:ALL' >> /etc/sudoers
chmod 440 /etc/sudoers

nm=`hostname`
sed -i "s/ServerActive=127.0.0.1/ServerActive=$zbxser/g" $zbxcfg
sed -i "s/Server=127.0.0.1/Server=$zbxser/g" $zbxcfg
sed -i "s/Hostname=HOSTNAME/Hostname=$nm/" $zbxcfg
sed -i "s/METADATA/$meta/" $zbxcfg

if [ $os == 'ubuntu' ]
   then
       \cp /etc/zabbix/start/zabbix-agent /etc/init.d/zabbix-agent
	   if [ ! -e /usr/bin/supervisorctl ]
	      then
		  apt -y install supervisor sysstat  >> $log 2>&1
                  systemctl enable supervisor >> $log 2>&1
                  systemctl start supervisor >> $log 2>&1
	   fi
       \cp /etc/zabbix/start/iostat.conf  /etc/supervisor/conf.d/iostat.conf
       chmod 755 /etc/init.d/zabbix-agent
fi

if [ $os == 'centos7' ]
   then
      setenforce 0 > /dev/null 2>&1
	   if [ ! -e /usr/bin/supervisorctl ]
	      then
		  yum -y install epel-release >> $log 2>&1
		  yum -y install supervisor sysstat >> $log 2>&1
                  systemctl enable supervisord >> $log 2>&1
                  systemctl start supervisord >> $log 2>&1
	   fi
       yum -y install nc >> $log 2>&1
      \cp /etc/zabbix/start/iostat.conf /etc/supervisord.d/iostat.ini
      \cp /etc/zabbix/start/zabbix-agent.service /usr/lib/systemd/system/zabbix-agent.service
fi

supervisorctl update >> $log 2>&1
systemctl daemon-reload >> $log 2>&1
systemctl start zabbix-agent >> $log 2>&1
systemctl enable zabbix-agent >> $log 2>&1

echo -e '\033[32mOK\033[0m'
echo -e "\033[32mzabbix-agent $ver install success\033[0m"
```
