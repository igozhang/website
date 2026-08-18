# VSphere6.7 添加SNMP监控

vmware图形界面->配置-->设置-->常规-->SNMP接收方

esxi开启snmp服务

```
命令行方式添加snmp配置
vcsa
Command> snmp.set --users test/-/-/none
snmp.enable
Command> snmp.get

esxi
esxcli system snmp set --communities public,eastnoc,westnoc
esxcli system snmp set --targets=192.168.99.99@161/public
esxcli system snmp set --enable true
esxcli system snmp test

esxcli network firewall ruleset set --ruleset-id snmp --allowed-all true
esxcli network firewall ruleset set --ruleset-id snmp --enabled true
```
