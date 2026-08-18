# shells;shell函数

```
查询当前服务器和ntp服务器的timeoffset,超过1s就提示
timeRangeValidation(){
echo "checking timeoffset....."
timeoffset=$(ntpdate -q ntp.myoas.com | head -1 | cut -d " " -f 6 | sed "s/.$//")
echo "timeoffset is:"$timeoffset
echo $(date) >>/tmp/timeoffset
echo $timeoffset >>/tmp/timeoffset
number=$(echo $timeoffset|awk -F"-" '{print $NF}')
echo $number
#if [ $number -ge 1.0 ]; then
if [ $(echo "$number < 1.0" | bc) -eq 1 ];then
    echo "valid timeoffset"
else
    echo "timeoffset is too large, check again"
fi
}
```
