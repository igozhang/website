# keepalived_cfg配置文件

```
CentOS Linux release 7.6.1810 (Core)
nginx/1.20.1
Keepalived v1.3.5
```

```
192.168.3.2-3 两台nginx做互备,虚拟ip192.168.3.112
以下为第一台(主)192.168.3.2配置;
第二台192.168.3.3(备)配置只需要修改3处:
1. router_id igoZhang_LB_112_2 改为router_id igoZhang_LB_112_3
2. state MASTER  改为 BACKUP
3. priority 200  改为 priority 110
```

```
cat /etc/keepalived/chk_nginx.sh 
#!/bin/bash
# description: nginx check script
# check nginx is running or not. if not, start nginx
# if start fail, then stop keepalived
status=$(ps -C nginx --no-heading|wc -l)
echo $status
if [ "${status}" = "0" ]; then
	/usr/sbin/nginx
	sleep 3
	status2=$(ps -C nginx --no-heading|wc -l)
	if [ "${status2}" = "0"  ]; then
		/etc/init.d/keepalived stop
	fi
fi
```

```
cat /etc/keepalived/keepalived.conf
global_defs {
   router_id igoZhang_LB_112_2
   vrrp_skip_check_adv_addr
   #vrrp_strict
   vrrp_garp_interval 0
   vrrp_gna_interval 0
}

vrrp_script chk_nginx {
    script "/etc/keepalived/chk_nginx.sh"
    interval 2
    weight 2
}

vrrp_instance igoZhang_LB_112 {
    state MASTER
    interface eth0
    virtual_router_id 112
    priority 200
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 112-igoZhang
    }
	track_script {
        chk_nginx
    }
    virtual_ipaddress {
		192.168.3.112/24 dev eth0
    }
}
```
