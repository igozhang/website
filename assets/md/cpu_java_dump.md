# cpu高则触发java_dump

```
crontab -e
* * * * * sh /data/igo/java_dump.sh >>/tmp/java_dump.log 2>&1
```

```
# cat java_dump.sh 
#!/bin/bash

a=$(uptime |awk '{print $10}'|awk -F',' '{print $1}'|awk -F. '{print $1}')
b=5
pid=$(ps -ef |grep -i sig1 |grep sig2 |awk '{print $2}')
echo "$a cpu used"

if [ `echo "$a < $b"|bc` -eq 1 ] ; then
echo  "health "
else
jmap -dump:format=b,file=/data/dump/$pid-$(date +%m%d%H).hprof $pid && chmod 644 /data/dump/$pid-$(date +%m%d%H).hprof
fi
```
