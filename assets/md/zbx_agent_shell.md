# zabbix_agent3.4批量安装并导入zabbix_server

```
env
cent74
zabbix3.4

早期脚本，比较手动
```

```
1，检测ip可连通性，ip可用性：ping -c -w  -c count 次数， -w timeout时间second

\\建议检查前先IP去重复,可以再做一步区分windows和linux
$ for i in $(cat sit1.ip); do ping -c 1 -w 2 $i; if [[ $? != 0 ]];then echo $i >>sit1.unreachable.ip; else echo $i >> sit1.enable.ip; fi; done
```

```
2，给可联通的ip配置互信：把常用密码都试一遍，后面的尝试如果前面已经成功互信了也会优先用密钥连接成功，timeout设置超时时间，避免在不成功的节点耗费时间

$ for i in $(cat sit1.enable.ip); do echo $i; cat ~/.ssh/id_*.pub|sshpass -p "igopwd@12345" timeout 1 ssh -o stricthostkeychecking=no root@$i 'cat>>.ssh/authorized_keys'; done
$ for i in $(cat sit1.enable.ip); do echo $i; cat ~/.ssh/id_*.pub|sshpass -p  'igopwd123!@#' timeout 1 ssh -o stricthostkeychecking=no root@$i 'cat>>.ssh/authorized_keys'; done
\\'igopwd123!@#'密码包含特殊字符，必须用''符号包含出来;
```

```
3，下发安装程序，并安装zabbix-agent；（通过配互信的输出，只针对互信成功的ip下发）下发前可以先确认一下是否已经有zabbix；

$ for i in $(cat sit1.key.success.ip); do echo $i; scp zabbix-agent-3.4.15-1.el7.x86_64.rpm root@$i:/tmp/; done
$ for i in $(cat sit2.round2.ip); do echo $i; ssh root@$i  'yum -y install /tmp/zabbix-agent-3.4.15-1.el7.x86_64.rpm'; done
下发安装完验证下是否成功，事前最好也验证一下，避免和别人已经监控的机器冲突：
$ for i in $(cat sit1.key.success.ip); do echo $i; ssh root@$i 'grep ^[a-Z] /etc/zabbix/zabbix_agentd.conf|grep Hostname';done
```

```
4，下发命令执行安装，配置，启动，删除脚本操作；
$ vi zAgent.sh_v1.1    //兼容rhel6,rhel7

#!/bin/bash

i=`/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`
sed -i 's/^Server=127.0.0.1/Server=10.7.36.38/g' /etc/zabbix/zabbix_agentd.conf    ##rhel7
sed -i 's/^Server=0.0.0.0\/0/Server=10.7.36.38/g' /etc/zabbix/zabbix_agentd.conf   ##rhel6
sed -i 's/^ServerActive=127.0.0.1/ServerActive=10.7.36.38/g' /etc/zabbix/zabbix_agentd.conf
sed -i "s/^Hostname=Zabbix[ ]server/Hostname=zClient_sit1_$i/g" /etc/zabbix/zabbix_agentd.conf
##rhel7
systemctl start zabbix-agent
systemctl enable zabbix-agent
##rhel6
service zabbix-agent start
chkconfig zabbix-agent on
```

```
5. 下发并执行

rhel7:
$ for i in $(cat sit1.last.ip ); do echo $i; scp zAgent.sh root@$i:/tmp/ && timeout 2 ssh root@$i 'sh /tmp/zAgent.sh && rm -rf /tmp/zAgent.sh';done
rhel6:
$ for i in $(cat rhel6.ip ); do echo $i;  timeout 2 ssh root@$i 'service zabbix-agent start && chkconfig zabbix-agent on';done
    rhel6发现配置项错误：Server=0.0.0.0/0
$ for i in $(cat rhel6.ip ); do echo $i;  timeout 2 ssh root@$i 'grep ^[a-Z] /etc/zabbix/zabbix_agentd.conf|grep Server=';done
    修改：
$ for i in $(cat rhel6.ip ); do echo $i;  timeout 2 ssh root@$i "sed -i 's/^Server=0.0.0.0\/0/Server=10.7.36.38/g' /etc/zabbix/zabbix_agentd.conf";done
    重启：
$ for i in $(cat rhel6.ip ); do echo $i;  timeout 2 ssh root@$i "service zabbix-agent restart";done
```

```
6. 制作zabbix导入模板（首先web手动添加一台机器，然后点导出，用脚本修改导出的xml文件后实现批量导入）：

$ cat insert.and.replace.sh
#!/bin/bash
for ip in $(cat sit1.last_version.ip); do
echo $ip
point=`grep -n '</host>' zbx_import_hosts.xml | cut -d ":" -f 1|tail -1`
sed -i "$point r host.txt" zbx_import_hosts.xml
sed -i "s/point4ip/$ip/g" zbx_import_hosts.xml;
done

//涉及到的文件host.block.txt,zbx_import_hosts.xml，一个是zabbix机器的导出文件，另一个属于导出文件里重复代码块<host>
```

注意：必须要选择监控模板之后ZBX方块才会显示绿色！！，否则一直显示灰色
