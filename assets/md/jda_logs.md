# jda_logs

```
#!/bin/bash
#日志清理脚本
#author:igo

#集成日志
find /data/jda/jda2019_1/customize_scripts/integration/log/debug -name "*.log" -mtime +30 |xargs rm -rf {} \;
find /data/jda/jda2019_1/customize_scripts/integration/log/error -name "*.log" -mtime +30 |xargs rm -rf {} \;
find /data/jda/jda2019_1/customize_scripts/integration/log/temp -name "*.log" -mtime +30 |xargs rm -rf {} \;

#Weblogic日志
find /data/jda/jda2019_1/config/JDADomain/servers/*Server/logs/*.log*  -mtime +30 |xargs rm -rf {} \;

#appserver日志
find /data/jda/jda2019_1*/config/logs/appserver -name "*" -mtime +30 |xargs rm -rf {} \;

#node manager日志
find /data/jda/jda2019_1*/config/logs/NodePoolManager-*.log -mtime +30 |xargs rm -rf {} \;
find /data/jda/jda2019_1*/config/logs/NodePoolManager-*/*.log* -mtime +30 |xargs rm -rf {} \;

#node日志
find /data/jda/jda2019_1/config/logs/NodePool* -name "*.log*" -mtime +30 |xargs rm -rf {} \;
find /data/jda/jda2019_1_esp1/config/logs/NodePool* -name "*.log*" -mtime +30 |xargs rm -rf {} \;
find /data/jda/jda2019_1_atp1/config/logs/NodePool* -name "*.log*" -mtime +30 |xargs rm -rf {} \;

#node native日志
find /data/jda/jda2019_1*/config/logs/NodePool-MP*/*native.log* -mtime +30 |xargs rm -rf {} \;
find /data/jda/jda2019_1*/config/logs/NodePool-MP*/*native.log* -mtime +30 |xargs rm -rf {} \;

#SRE Process调度日志
find /data/jda/jda2019_1*/config/logs/*.log.xml -mtime +30 |xargs rm -rf {} \;

#sqlldr日志
find /data/jda/jda2019_1*/config/scpoweb/esp/instances/masterplanning/export-staging -mtime +30 |xargs rm -rf {} \;

#JVM crash日志
find /data/jda/jda2019_1*/config/bin/platform/hs_err* -mtime +30 |xargs rm -rf {} \;
find /data/jda/jda2019_1*/config/bin/platform/core.* -mtime +30 |xargs rm -rf {} \;
```
