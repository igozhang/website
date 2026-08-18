# Esxi补丁升级

Env  
Esxi6.7

1.VmWare官网下载补丁包

https://my.vmware.com/cn/group/vmware/patch#search

下载对应补丁包

2.

2.上传补丁包到主机

存储->datastore1->数据存储浏览器->上载

3.SSH升级补丁包

```
3.1
开启esxi主机ssh
主机->管理->服务->ssh->开启

3.2
ssh安装补丁包
确认当前版本：
[root@localhost:~] vmware -vl
VMware ESXi 6.7.0 build-15160138
VMware ESXi 6.7.0 Update 3

Esxi进入维护模式
esxcli system maintenanceMode set --enable true

查询补丁包信息，检查offline bundle中的可用的profile：
[root@localhost:~] esxcli software sources profile list -d /vmfs/volumes/datastore1/ESXi670-202103001.zip
Name                              Vendor        Acceptance Level  Creation Time        Modification Time
--------------------------------  ------------  ----------------  -------------------  -------------------
ESXi-6.7.0-20210301001s-no-tools  VMware, Inc.  PartnerSupported  2021-03-04T10:17:40  2021-03-04T10:17:40
ESXi-6.7.0-20210304001-standard   VMware, Inc.  PartnerSupported  2021-03-04T10:17:40  2021-03-04T10:17:40
ESXi-6.7.0-20210301001s-standard  VMware, Inc.  PartnerSupported  2021-03-04T10:17:40  2021-03-04T10:17:40
ESXi-6.7.0-20210304001-no-tools   VMware, Inc.  PartnerSupported  2021-03-04T10:17:40  2021-03-04T10:17:40

试运行：看看哪些包会被新增，哪些VIB包会被清除，是否有报错
[root@localhost:~] esxcli software profile update -p ESXi-6.7.0-20210304001-standard -d /vmfs/volumes/datastore1/ESXi670-202103001.zip --dry-run

正式升级：
[root@localhost:~] esxcli software profile update -p ESXi-6.7.0-20210304001-standard -d /vmfs/volumes/datastore1/ESXi670-202103001
.zip

Esxi退出维护模式
esxcli system maintenanceMode set --enable false

重启完成升级

简述在线升级：
esxcli system maintenanceMode set --enable true
esxcli network firewall ruleset set -e true -r httpClient
esxcli software sources profile list -d https://hostupdate.vmware.com/software/VUM/PRODUCTION/main/vmw-depot-index.xml | grep ESXi-6.7
esxcli software profile update -d https://hostupdate.vmware.com/software/VUM/PRODUCTION/main/vmw-depot-index.xml -p ESXi-6.7U3d-18538813-standard --dry-run
esxcli software profile update -d https://hostupdate.vmware.com/software/VUM/PRODUCTION/main/vmw-depot-index.xml -p ESXi-6.7U3d-18538813-standard
esxcli network firewall ruleset set -e false -r httpClient
esxcli system maintenanceMode set --enable false
reboot

命令参数:  -p 代表 profile,  -d 代表depot
一般选择standard 结尾的那个profile， no-tools结尾的一般用于pxe
```
