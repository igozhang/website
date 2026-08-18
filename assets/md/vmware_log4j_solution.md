# Vmware_log4j漏洞临时解决方案

```
env
vmware6.7
验证通过
漏洞编号 CVE-2021-44228
```

```
漏洞补丁
remove_log4j_class
vmsa-2021-0028-kb87081

上传到vcenter /tmp

# python /tmp/vmsa-2021-0028-kb87081.py

This script will help to automate the steps described in VMware KB https://kb.vmware.com/s/article/87081

All Services will be restarted by the script to mitigate the VMSA, Please enter YES to proceed further or NO to Exit [[Yes/No/Y/N]] ? y

Remediating vMon Config files
...Taking Backup of file /usr/lib/vmware-vmon/java-wrapper-vmon
...Successfully completed the backup - /usr/lib/vmware-vmon/java-wrapper-vmon_backup_24-Dec-21-10-28-12
...Updating Config file
...Completed Config file update
...Stopping all Services
...Starting all Services
...Successfully Started All Services
...Completed remediating vMon services

Remediating STSD Config files
...Taking Backup of file /etc/rc.d/init.d/vmware-stsd
...Successfully completed the backup - /root/vmware-stsd_backup_24-Dec-21-10-28-12
...Updating Config file
...Completed Config file update
...Restarting vmware-stsd Service
...Successfully restarted vmware-stsd Service
...Completed remediating vmware-stsd service

Remediating IDMD Config files
...Taking Backup of file /etc/rc.d/init.d/vmware-sts-idmd
...Successfully completed the backup - /root/vmware-sts-idmd_backup_24-Dec-21-10-28-12
...Updating Config file
...Completed Config file update
...Restarting vmware-stsd-idmd Service
...Successfully restarted vmware-sts-idmd Service
...Completed remediating vmware-sts-idmd service

Remediating Analytics Service Config files
...Taking Backup of file /usr/lib/vmware/common-jars/log4j-core-2.8.2.jar
...Successfully completed the backup - /usr/lib/vmware/common-jars/log4j-core-2.8.2.jar_backup_24-Dec-21-10-28-12
...Updating Config file
...Successfully updated the Jar file
...Restarting Analytics Service
...Successfully restarted Analytics Service
...Completed remediating Analytics service

Remediating CM Service Config files
...Taking Backup of file /usr/lib/vmware-cm/lib/log4j-core.jar
...Successfully completed the backup - /usr/lib/vmware-cm/lib/log4j-core.jar_backup_24-Dec-21-10-28-12
...Updating Config file
...Successfully updated the Jar file
...Restarting CM Service
...Successfully restarted CM Service
...Completed remediating CM service

Verifying the vulnerability status after applying the workaround :

..Verifying the status of vMon Services
....SUCCESS
..Verifying the status of vmware-stsd Service
....SUCCESS
..Verifying the status of vmware-sts-idmd Service
....SUCCESS
..Verifying the status of VMware Analytics Service
....SUCCESS
..Verifying the status of CM Service
....SUCCESS
Successfully applied the workaround steps in KB 87081 to mitigate the VMSA-2021-0028

# python /tmp/remove_log4j_class.py
A service stop and start is required to complete this operation.  Continue?[y]y
2021-12-24T10:41:03 INFO stop: stopping services
2021-12-24T10:42:59 INFO process_archive: VULNERABLE FILE: /usr/lib/vmware-dbcc/dbcc.jar backed up to /tmp/tmpqyy53er7/usr/lib/vmware-dbcc/dbcc.jar.bak
2021-12-24T10:43:06 INFO process_archive: VULNERABLE FILE: /usr/lib/vmware/common-jars/log4j-core-2.11.2.jar backed up to /tmp/tmpqyy53er7/usr/lib/vmware/common-jars/log4j-core-2.11.2.jar.bak
2021-12-24T10:43:15 INFO process_war: VULNERABLE FILE: /usr/lib/vmware-sso/vmware-sts/webapps/ROOT.war backed up to /tmp/tmpqyy53er7/usr/lib/vmware-sso/vmware-sts/webapps/ROOT.war.bak
2021-12-24T10:43:16 INFO process_archive: VULNERABLE FILE: /usr/lib/vmware-sso/vmware-sts/webapps/ROOT/WEB-INF/lib/log4j-core-2.11.2.jar backed up to /tmp/tmpqyy53er7/usr/lib/vmware-sso/vmware-sts/webapps/ROOT/WEB-INF/lib/log4j-core-2.11.2.jar.bak
2021-12-24T10:43:21 INFO start: starting services
2021-12-24T10:50:01 INFO main: Done.

方法适用于6.5,6.7,7.0
```
