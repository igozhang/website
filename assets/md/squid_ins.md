# squid_ins

bash安装squid，ubuntu为例

## ubuntu

```
rm -f install-squid.sh && cat > install-squid.sh <<'EOF'
#!/bin/bash
# Author:运维工具
# System:Ubuntu20.04
# Squid 纯净无错乱脚本｜无转义｜无乱码｜生产可用
# 特性：默认无密码、安装强制清空残留、禁止多参数、健壮容错

# ========== 全局变量 ==========
SQUID_CONF="/etc/squid/squid.conf"
SQUID_PASS="/etc/squid/passwd"
DEFAULT_USER="proxy"
DEFAULT_PASS="123456"
DEFAULT_PORT="3128"
# 自动获取本机内网IP（用于输出客户端命令）
SERVER_IP=$(hostname -I | awk '{print $1}')

# ========== 帮助菜单 ==========
show_help(){
    SERVER_IP=$(hostname -I | awk '{print $1}')
    echo "============================================="
    echo "          Squid 运维脚本使用帮助"
    echo "============================================="
    echo "用法:  bash install-squid.sh [单个参数]   (每次仅允许一个参数)"
    echo ""
    echo "【命令参数】"
    echo "  install        一键重装(强制清空旧配置与数据后全新安装)"
    echo "  pass_on        开启密码认证(公网环境建议开启)"
    echo "  pass_off       关闭密码认证(恢复匿名访问)"
    echo "  remove         彻底卸载 Squid 及配置残留"
    echo "  passwd         修改代理认证密码(需已 pass_on)"
    echo "  status         查看 systemd 服务与监听端口"
    echo "  check          已安装时重新打印客户端配置与连通性测试命令"
    echo "  allow_all      允许全网 IP 连接代理"
    echo "  allow_lan      仅允许内网网段连接代理"
    echo "  help           查看本帮助(无参数时同 help)"
    echo ""
    echo "【代理与认证默认值】"
    echo "  监听端口:           ${DEFAULT_PORT}"
    echo "  代理账号(认证时):   ${DEFAULT_USER}"
    echo "  初始密码(仅服务端): ${DEFAULT_PASS}  (安装时写入 htpasswd，输出不向客户端展示)"
    echo "  密码认证:           关闭 (匿名直通，install 后默认状态)"
    echo "  认证方式:           Basic NCSA  (pass_on 后生效)"
    echo "  认证域(realm):      Proxy-Auth"
    echo "  认证子进程数:       5"
    echo "  密码文件:           ${SQUID_PASS}"
    echo "  本机 IP(自动检测):  ${SERVER_IP:-未检测到}"
    echo ""
    echo "【Squid 配置默认值(install 写入)】"
    echo "  主配置文件:         ${SQUID_CONF}"
    echo "  缓存目录:           /var/spool/squid  (ufs 1024MB, L1=16, L2=256)"
    echo "  内存缓存:           512 MB"
    echo "  单对象上限:         4096 MB"
    echo "  替换策略:           heap LFUDA"
    echo "  访问日志:           /var/log/squid/access.log"
    echo "  缓存日志:           /var/log/squid/cache.log"
    echo "  X-Forwarded-For:    关闭 (forwarded_for off)"
    echo ""
    echo "【访问控制默认值】"
    echo "  内网 ACL(localnet): 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16"
    echo "  http_access 顺序:   allow localnet -> allow all -> deny all"
    echo "  安装后访问范围:     允许所有来源 IP (等同 allow_all 未改前的默认行为)"
    echo ""
    echo "【系统与安装行为默认值】"
    echo "  适用系统:           Ubuntu 20.04 (apt 安装 squid、apache2-utils)"
    echo "  安装模式:           强制清空残留后重装 (remove_squid_force)"
    echo "  防火墙(ufw):        自动放行 tcp/${DEFAULT_PORT}"
    echo "  开机自启:           开启 (systemctl enable squid)"
    echo "  客户端代理协议:     HTTP  (export http_proxy/https_proxy=http://IP:PORT)"
    echo "  连通性测试域名:     www.qq.com, www.google.com  (见 check/install 输出)"
    echo "============================================="
}

# ========== 权限检测 ==========
check_root(){
    if [ $(id -u) -ne 0 ];then
        echo -e "\033[31m 请使用root权限执行！\033[0m"
        exit 1
    fi
}

# ========== 是否开启密码认证 ==========
is_auth_enabled(){
    [ -f ${SQUID_CONF} ] && grep -q "^auth_param basic program" ${SQUID_CONF} 2>/dev/null
}

# ========== 打印客户端配置与连通性测试 ==========
show_client_info(){
    SERVER_IP=$(hostname -I | awk '{print $1}')
    if is_auth_enabled; then
        local auth_status="开启(需账号密码)"
    else
        local auth_status="关闭(匿名无密码)"
    fi

    echo "============================================="
    echo " 代理地址: ${SERVER_IP}:${DEFAULT_PORT}"
    echo " 认证状态: ${auth_status}"
    echo " 开启密码: ./install-squid.sh pass_on"
    echo " 关闭密码: ./install-squid.sh pass_off"
    echo "============================================="
    echo -e "\033[34m 【Linux 客户端代理设置｜CentOS/Ubuntu 通用】\033[0m"
    echo -e "\033[36m ---------- 一、临时代理(仅当前终端) ----------\033[0m"
    echo "export http_proxy=http://${SERVER_IP}:${DEFAULT_PORT}"
    echo "export https_proxy=http://${SERVER_IP}:${DEFAULT_PORT}"
    echo "unset http_proxy https_proxy  # 取消代理"
    echo -e "\033[36m ---------- 二、永久代理(全局所有终端) ----------\033[0m"
    echo "# Ubuntu/Debian"
    echo "echo 'export http_proxy=http://${SERVER_IP}:${DEFAULT_PORT}' >> ~/.bashrc"
    echo "echo 'export https_proxy=http://${SERVER_IP}:${DEFAULT_PORT}' >> ~/.bashrc"
    echo "source ~/.bashrc"
    echo "# CentOS/RHEL"
    echo "echo 'export http_proxy=http://${SERVER_IP}:${DEFAULT_PORT}' >> ~/.bash_profile"
    echo "echo 'export https_proxy=http://${SERVER_IP}:${DEFAULT_PORT}' >> ~/.bash_profile"
    echo "source ~/.bash_profile"
    if is_auth_enabled; then
        echo -e "\033[36m ---------- 三、已开启密码认证时的客户端格式 ----------\033[0m"
        echo "# 将 <你的密码> 替换为实际密码(账号默认: ${DEFAULT_USER})"
        echo "export http_proxy=http://${DEFAULT_USER}:<你的密码>@${SERVER_IP}:${DEFAULT_PORT}"
        echo "export https_proxy=http://${DEFAULT_USER}:<你的密码>@${SERVER_IP}:${DEFAULT_PORT}"
        echo "# 或使用 curl -U 指定账号密码:"
        echo "curl -x http://${SERVER_IP}:${DEFAULT_PORT} -U ${DEFAULT_USER}:<你的密码> -I --connect-timeout 10 https://www.qq.com"
    else
        echo -e "\033[36m ---------- 三、如需密码认证 ----------\033[0m"
        echo "# 服务端执行 pass_on 后，客户端使用:"
        echo "export http_proxy=http://${DEFAULT_USER}:<你的密码>@${SERVER_IP}:${DEFAULT_PORT}"
    fi
    echo -e "\033[36m ---------- 四、使用前连通性测试(在 Linux 客户端执行) ----------\033[0m"
    if is_auth_enabled; then
        echo "curl -x http://${SERVER_IP}:${DEFAULT_PORT} -U ${DEFAULT_USER}:<你的密码> -I --connect-timeout 10 https://www.qq.com"
        echo "curl -x http://${SERVER_IP}:${DEFAULT_PORT} -U ${DEFAULT_USER}:<你的密码> -I --connect-timeout 10 https://www.google.com"
    else
        echo "curl -x http://${SERVER_IP}:${DEFAULT_PORT} -I --connect-timeout 10 https://www.qq.com"
        echo "curl -x http://${SERVER_IP}:${DEFAULT_PORT} -I --connect-timeout 10 https://www.google.com"
    fi
    echo "# 返回 HTTP/1.x 200 或 301/302 表示代理可用；超时或 403 请检查防火墙/ACL"
    echo "============================================="
}

# ========== 强制彻底清空 ==========
remove_squid_force(){
    check_root
    echo -e "\033[33m 正在强制清理残留...\033[0m"
    systemctl stop squid >/dev/null 2>&1
    systemctl disable squid >/dev/null 2>&1
    apt remove squid -y >/dev/null 2>&1
    apt autoremove -y >/dev/null 2>&1
    rm -rf /etc/squid /var/spool/squid /var/log/squid
    ufw delete allow ${DEFAULT_PORT}/tcp >/dev/null 2>&1
    echo -e "\033[32m 残留清理完成！\033[0m"
}

# ========== 安装重装函数 ==========
install_squid(){
    check_root
    remove_squid_force

    echo -e "\033[32m 1.安装Squid依赖...\033[0m"
    apt update -y >/dev/null 2>&1
    apt install squid apache2-utils -y >/dev/null 2>&1

    mkdir -p /etc/squid
    mkdir -p /var/spool/squid

    [ -f ${SQUID_CONF} ] && cp ${SQUID_CONF} ${SQUID_CONF}.bak

    echo -e "\033[32m 2.写入优化配置(默认匿名无密码)...\033[0m"
    tee ${SQUID_CONF} >/dev/null <<CONF
http_port ${DEFAULT_PORT}
cache_dir ufs /var/spool/squid 1024 16 256
cache_mem 512 MB
maximum_object_size 4096 MB
cache_replacement_policy heap LFUDA
cache_access_log /var/log/squid/access.log
cache_log /var/log/squid/cache.log

acl localnet src 10.0.0.0/8
acl localnet src 172.16.0.0/12
acl localnet src 192.168.0.0/16

# 默认关闭密码认证
# auth_param basic program /usr/lib/squid/basic_ncsa_auth ${SQUID_PASS}
# auth_param basic realm Proxy-Auth
# auth_param basic children 5
# acl authenticated proxy_auth REQUIRED

http_access allow localnet
http_access allow all
http_access deny all

forwarded_for off
request_header_access Allow allow all
CONF

    htpasswd -bc ${SQUID_PASS} ${DEFAULT_USER} ${DEFAULT_PASS} >/dev/null 2>&1
    squid -z >/dev/null 2>&1
    ufw allow ${DEFAULT_PORT}/tcp >/dev/null 2>&1
    systemctl enable squid >/dev/null 2>&1
    systemctl restart squid >/dev/null 2>&1

    echo -e "\033[32m 3.Squid 安装完成！\033[0m"
    show_client_info
}

# ========== 交互式卸载 ==========
remove_squid(){
    check_root
    read -p "确定卸载?(y/n):" op
    [ "${op}" != "y" ] && echo "已取消" && exit 0
    remove_squid_force
}

# ========== 修改密码 ==========
change_passwd(){
    check_root
    [ ! -d /etc/squid ] && mkdir -p /etc/squid
    read -p "输入新密码:" newpass
    htpasswd -bc ${SQUID_PASS} ${DEFAULT_USER} ${newpass}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 密码修改成功(新密码未在此显示，请自行记录)\033[0m"
    echo "客户端格式: export http_proxy=http://${DEFAULT_USER}:<你的密码>@${SERVER_IP}:${DEFAULT_PORT}"
}

# ========== 状态查看 ==========
status_squid(){
    echo "========== Squid 运行信息 =========="
    systemctl status squid | head -15
    echo ""
    echo "监听端口: $(ss -tulpn | grep squid 2>/dev/null || echo '未检测到')"
    if is_auth_enabled; then
        echo -e "认证状态: \033[31m 开启(需密码)\033[0m"
    else
        echo -e "认证状态: \033[32m 关闭(匿名)\033[0m"
    fi
}

# ========== 已安装时重新打印客户端信息 ==========
check_squid(){
    if ! command -v squid >/dev/null 2>&1 || [ ! -f ${SQUID_CONF} ]; then
        echo -e "\033[31m Squid 未安装，请先执行: bash install-squid.sh install\033[0m"
        exit 1
    fi
    echo "========== Squid 安装检查 =========="
    if systemctl is-active squid >/dev/null 2>&1; then
        echo -e "服务状态: \033[32m 运行中\033[0m"
    else
        echo -e "服务状态: \033[31m 未运行(可执行: systemctl start squid)\033[0m"
    fi
    echo "监听端口: $(ss -tulpn | grep squid 2>/dev/null || echo '未检测到')"
    echo ""
    show_client_info
}

# ========== 开启密码认证 ==========
pass_on(){
    check_root
    sed -i 's/^#auth_param/auth_param/g' ${SQUID_CONF}
    sed -i 's/^#acl authenticated/acl authenticated/g' ${SQUID_CONF}
    sed -i 's/http_access allow all/#http_access allow all/g' ${SQUID_CONF}
    sed -i '/http_access allow localnet/a\http_access allow authenticated' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 密码认证已开启｜账号: ${DEFAULT_USER} (密码见 passwd 命令修改，不在此显示)\033[0m"
    show_client_info
}

# ========== 关闭密码认证 ==========
pass_off(){
    check_root
    sed -i 's/^auth_param/#auth_param/g' ${SQUID_CONF}
    sed -i 's/^acl authenticated/#acl authenticated/g' ${SQUID_CONF}
    sed -i 's/#http_access allow all/http_access allow all/g' ${SQUID_CONF}
    sed -i '/http_access allow authenticated/d' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 密码认证关闭｜匿名直通\033[0m"
    show_client_info
}

# ========== 开放全网 ==========
allow_all(){
    check_root
    sed -i 's/acl localnet src 192.168.0.0\/16/#&/g' ${SQUID_CONF}
    sed -i '/http_access allow all/i\acl all_ip src 0.0.0.0/0\nhttp_access allow all_ip' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 已开放全网访问\033[0m"
}

# ========== 仅内网 ==========
allow_lan(){
    check_root
    sed -i '/all_ip/d' ${SQUID_CONF}
    sed -i 's/#acl localnet src 192.168.0.0\/16/&/g' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 已限制仅内网访问\033[0m"
}

# ========== 参数校验｜禁止多参数 ==========
if [ $# -gt 1 ];then
    echo -e "\033[31m 错误：只能传入单个参数！\033[0m"
    exit 1
fi

# ========== 逻辑入口 ==========
case $1 in
install) install_squid;;
pass_on) pass_on;;
pass_off) pass_off;;
remove) remove_squid;;
passwd) change_passwd;;
status) status_squid;;
check) check_squid;;
allow_all) allow_all;;
allow_lan) allow_lan;;
help|"") show_help;;
*) echo -e "\033[31m 参数错误！输入 help 查看帮助\033[0m";;
esac
EOF
chmod +x install-squid.sh
echo -e "\033[32m ✅ 使用示例： bash install-squid.sh install | bash install-squid.sh check \033[0m"
```

## 优雅控制

```
chrome插件配置
SwitchyOmega (V3) https igozhang.cc 7029

启用squid
iptables -D INPUT -p tcp --dport 7028 -j DROP 2>/dev/null
iptables -D INPUT -p tcp --dport 7029 -j DROP 2>/dev/null
禁用squid
iptables -A INPUT -p tcp --dport 7028 -j DROP
iptables -A INPUT -p tcp --dport 7029 -j DROP
```

## 腾讯云OpenCloudOS 9.4

```
rm -f install-squid.sh && cat > install-squid.sh <<'EOF'
#!/bin/bash
# Author:运维工具
# System:OpenCloudOS 9.4
# Squid 纯净无错乱脚本｜无转义｜无乱码｜生产可用
# 特性：默认无密码、安装强制清空残留、禁止多参数、健壮容错
# 代理：7028 HTTP + 7029 HTTPS(宝塔证书 igozhang.cc)

# ========== 全局变量 ==========
SQUID_CONF="/etc/squid/squid.conf"
SQUID_PASS="/etc/squid/passwd"
DEFAULT_USER="proxy"
DEFAULT_PASS="123456"
DEFAULT_PORT_HTTP="7028"
DEFAULT_PORT_HTTPS="7029"
TLS_CERT="/www/server/panel/vhost/cert/igozhang.cc/igozhang.cc.pem"
TLS_KEY="/www/server/panel/vhost/cert/igozhang.cc/igozhang.cc.key"
TLS_DOMAIN="igozhang.cc"
# basic_ncsa_auth 路径(OpenCloudOS/RHEL9 多为 lib64)
if [ -f /usr/lib64/squid/basic_ncsa_auth ]; then
    NCSA_AUTH="/usr/lib64/squid/basic_ncsa_auth"
elif [ -f /usr/lib/squid/basic_ncsa_auth ]; then
    NCSA_AUTH="/usr/lib/squid/basic_ncsa_auth"
else
    NCSA_AUTH="/usr/lib64/squid/basic_ncsa_auth"
fi
# 自动获取本机内网IP（用于输出客户端命令）
SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
[ -z "${SERVER_IP}" ] && SERVER_IP=$(ip -4 route get 1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')

# ========== 证书检测与权限 ==========
check_cert_files(){
    if [ ! -f "${TLS_CERT}" ] || [ ! -f "${TLS_KEY}" ]; then
        echo -e "\033[31m 证书文件不存在，请确认宝塔证书路径：\033[0m"
        echo "  ${TLS_CERT}"
        echo "  ${TLS_KEY}"
        exit 1
    fi
}

fix_cert_perm(){
    chmod 644 "${TLS_CERT}" 2>/dev/null
    chmod 640 "${TLS_KEY}" 2>/dev/null
    chgrp squid "${TLS_KEY}" 2>/dev/null || true
    chgrp squid "${TLS_CERT}" 2>/dev/null || true
}

# ========== 防火墙放行/撤销 ==========
firewall_allow_port(){
    if systemctl is-active firewalld >/dev/null 2>&1; then
        firewall-cmd --permanent --add-port=${DEFAULT_PORT_HTTP}/tcp >/dev/null 2>&1
        firewall-cmd --permanent --add-port=${DEFAULT_PORT_HTTPS}/tcp >/dev/null 2>&1
        firewall-cmd --reload >/dev/null 2>&1
    fi
}

firewall_remove_port(){
    if systemctl is-active firewalld >/dev/null 2>&1; then
        firewall-cmd --permanent --remove-port=${DEFAULT_PORT_HTTP}/tcp >/dev/null 2>&1
        firewall-cmd --permanent --remove-port=${DEFAULT_PORT_HTTPS}/tcp >/dev/null 2>&1
        firewall-cmd --reload >/dev/null 2>&1
    fi
}

# ========== SELinux 自定义端口 ==========
selinux_allow_port(){
    if command -v getenforce >/dev/null 2>&1 && [ "$(getenforce 2>/dev/null)" != "Disabled" ]; then
        if command -v semanage >/dev/null 2>&1; then
            for p in ${DEFAULT_PORT_HTTP} ${DEFAULT_PORT_HTTPS}; do
                semanage port -a -t squid_port_t -p tcp ${p} >/dev/null 2>&1 \
                    || semanage port -m -t squid_port_t -p tcp ${p} >/dev/null 2>&1
            done
        fi
    fi
}

# ========== 帮助菜单 ==========
show_help(){
    SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
    [ -z "${SERVER_IP}" ] && SERVER_IP=$(ip -4 route get 1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
    echo "============================================="
    echo "          Squid 运维脚本使用帮助"
    echo "============================================="
    echo "用法:  bash install-squid.sh [单个参数]   (每次仅允许一个参数)"
    echo ""
    echo "【命令参数】"
    echo "  install        一键重装(强制清空旧配置与数据后全新安装)"
    echo "  pass_on        开启密码认证(公网环境建议开启)"
    echo "  pass_off       关闭密码认证(恢复匿名访问)"
    echo "  remove         彻底卸载 Squid 及配置残留"
    echo "  passwd         修改代理认证密码(需已 pass_on)"
    echo "  status         查看 systemd 服务与监听端口"
    echo "  check          已安装时重新打印客户端配置与连通性测试命令"
    echo "  allow_all      允许全网 IP 连接代理"
    echo "  allow_lan      仅允许内网网段连接代理"
    echo "  help           查看本帮助(无参数时同 help)"
    echo ""
    echo "【代理与认证默认值】"
    echo "  HTTP  代理端口:     ${DEFAULT_PORT_HTTP}  (明文，内网/SSH隧道)"
    echo "  HTTPS 代理端口:     ${DEFAULT_PORT_HTTPS} (TLS，公网客户端推荐)"
    echo "  TLS 证书:           ${TLS_CERT}"
    echo "  TLS 私钥:           ${TLS_KEY}"
    echo "  证书域名:           ${TLS_DOMAIN}"
    echo "  代理账号(认证时):   ${DEFAULT_USER}"
    echo "  初始密码(仅服务端): ${DEFAULT_PASS}  (安装时写入 htpasswd，输出不向客户端展示)"
    echo "  密码认证:           关闭 (匿名直通，install 后默认状态)"
    echo "  认证方式:           Basic NCSA  (pass_on 后生效)"
    echo "  认证域(realm):      Proxy-Auth"
    echo "  认证子进程数:       5"
    echo "  密码文件:           ${SQUID_PASS}"
    echo "  本机 IP(自动检测):  ${SERVER_IP:-未检测到}"
    echo ""
    echo "【Squid 配置默认值(install 写入)】"
    echo "  主配置文件:         ${SQUID_CONF}"
    echo "  缓存目录:           /var/spool/squid  (ufs 1024MB, L1=16, L2=256)"
    echo "  内存缓存:           512 MB"
    echo "  单对象上限:         4096 MB"
    echo "  替换策略:           heap LFUDA"
    echo "  访问日志:           /var/log/squid/access.log"
    echo "  缓存日志:           /var/log/squid/cache.log"
    echo "  X-Forwarded-For:    关闭 (forwarded_for off)"
    echo ""
    echo "【访问控制默认值】"
    echo "  内网 ACL(localnet): 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16"
    echo "  http_access 顺序:   allow localnet -> allow all -> deny all"
    echo "  安装后访问范围:     允许所有来源 IP (等同 allow_all 未改前的默认行为)"
    echo ""
    echo "【系统与安装行为默认值】"
    echo "  适用系统:           OpenCloudOS 9.4 (dnf 安装 squid、httpd-tools)"
    echo "  安装模式:           强制清空残留后重装 (remove_squid_force)"
    echo "  防火墙(firewalld):  运行中则放行 tcp/${DEFAULT_PORT_HTTP} tcp/${DEFAULT_PORT_HTTPS}"
    echo "  SELinux:            启用时注册 squid_port_t 上述两端口"
    echo "  开机自启:           开启 (systemctl enable squid)"
    echo "  公网推荐:           https://${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS} (需在客户端信任证书)"
    echo "  连通性测试域名:     www.qq.com, www.google.com  (见 check/install 输出)"
    echo "============================================="
}

# ========== 权限检测 ==========
check_root(){
    if [ $(id -u) -ne 0 ];then
        echo -e "\033[31m 请使用root权限执行！\033[0m"
        exit 1
    fi
}

# ========== 是否开启密码认证 ==========
is_auth_enabled(){
    [ -f ${SQUID_CONF} ] && grep -q "^auth_param basic program" ${SQUID_CONF} 2>/dev/null
}

# ========== 打印客户端配置与连通性测试 ==========
show_client_info(){
    SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
    [ -z "${SERVER_IP}" ] && SERVER_IP=$(ip -4 route get 1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
    if is_auth_enabled; then
        local auth_status="开启(需账号密码)"
    else
        local auth_status="关闭(匿名无密码)"
    fi

    echo "============================================="
    echo " HTTP  代理(内网/隧道): http://${SERVER_IP}:${DEFAULT_PORT_HTTP}"
    echo " HTTPS 代理(公网推荐):  https://${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS}"
    echo "                        https://${SERVER_IP}:${DEFAULT_PORT_HTTPS}"
    echo " 认证状态: ${auth_status}"
    echo " 开启密码: ./install-squid.sh pass_on"
    echo " 关闭密码: ./install-squid.sh pass_off"
    echo " 腾讯云安全组请放行: TCP ${DEFAULT_PORT_HTTP} ${DEFAULT_PORT_HTTPS}"
    echo "============================================="
    echo -e "\033[34m 【一、HTTP 代理 ${DEFAULT_PORT_HTTP}｜Linux】\033[0m"
    echo "export http_proxy=http://${SERVER_IP}:${DEFAULT_PORT_HTTP}"
    echo "export https_proxy=http://${SERVER_IP}:${DEFAULT_PORT_HTTP}"
    echo -e "\033[34m 【二、HTTPS 代理 ${DEFAULT_PORT_HTTPS}｜公网客户端推荐】\033[0m"
    echo "# 代理地址 scheme 必须为 https:// (加密 CONNECT)"
    echo "export HTTPS_PROXY=https://${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS}"
    echo "export HTTP_PROXY=https://${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS}"
    echo "# curl 测试(证书已由宝塔签发则无需 -k；否则加 -k)"
    echo "curl -x https://${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS} -I --connect-timeout 15 https://www.google.com"
    echo -e "\033[34m 【三、Windows PowerShell HTTPS 代理测试】\033[0m"
    echo "curl.exe -sS -x \"https://${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS}\" -o NUL -w \"http_code=%{http_code}\" https://www.google.com/"
    if is_auth_enabled; then
        echo -e "\033[36m 【已开启密码认证】\033[0m"
        echo "export HTTPS_PROXY=https://${DEFAULT_USER}:<你的密码>@${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS}"
        echo "curl -x https://${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS} -U ${DEFAULT_USER}:<你的密码> -I https://www.google.com"
    fi
    echo -e "\033[36m 【说明】境内公网直连 HTTP:${DEFAULT_PORT_HTTP} 可能被干扰；请优先 HTTPS:${DEFAULT_PORT_HTTPS}\033[0m"
    echo "============================================="
}

# ========== 强制彻底清空 ==========
remove_squid_force(){
    check_root
    echo -e "\033[33m 正在强制清理残留...\033[0m"
    systemctl stop squid >/dev/null 2>&1
    systemctl disable squid >/dev/null 2>&1
    dnf remove -y squid >/dev/null 2>&1
    rm -rf /etc/squid /var/spool/squid /var/log/squid
    firewall_remove_port
    echo -e "\033[32m 残留清理完成！\033[0m"
}

# ========== 安装重装函数 ==========
install_squid(){
    check_root
    remove_squid_force
    check_cert_files

    echo -e "\033[32m 1.安装Squid依赖...\033[0m"
    dnf makecache -y >/dev/null 2>&1
    dnf install -y squid httpd-tools policycoreutils-python-utils openssl >/dev/null 2>&1

    fix_cert_perm

    mkdir -p /etc/squid
    mkdir -p /var/spool/squid

    [ -f ${SQUID_CONF} ] && cp ${SQUID_CONF} ${SQUID_CONF}.bak

    echo -e "\033[32m 2.写入优化配置(7028 HTTP + 7029 HTTPS)...\033[0m"
    tee ${SQUID_CONF} >/dev/null <<CONF
# HTTP 明文代理(内网/本机)
http_port ${DEFAULT_PORT_HTTP}
# HTTPS 代理(TLS 终结，公网客户端推荐)
https_port ${DEFAULT_PORT_HTTPS} tls-cert=${TLS_CERT} tls-key=${TLS_KEY}

cache_dir ufs /var/spool/squid 1024 16 256
cache_mem 512 MB
maximum_object_size 4096 MB
cache_replacement_policy heap LFUDA
cache_access_log /var/log/squid/access.log
cache_log /var/log/squid/cache.log

acl localnet src 10.0.0.0/8
acl localnet src 172.16.0.0/12
acl localnet src 192.168.0.0/16

# 默认关闭密码认证
# auth_param basic program ${NCSA_AUTH} ${SQUID_PASS}
# auth_param basic realm Proxy-Auth
# auth_param basic children 5
# acl authenticated proxy_auth REQUIRED

http_access allow localnet
http_access allow all
http_access deny all

forwarded_for off
request_header_access Allow allow all
CONF

    htpasswd -bc ${SQUID_PASS} ${DEFAULT_USER} ${DEFAULT_PASS} >/dev/null 2>&1
    selinux_allow_port
    if ! squid -z >/dev/null 2>&1; then
        echo -e "\033[31m 缓存初始化失败，请检查配置\033[0m"
        exit 1
    fi
    if ! squid -k parse 2>/dev/null; then
        echo -e "\033[31m 配置校验失败，请检查证书路径与 https_port 语法：\033[0m"
        squid -k parse
        exit 1
    fi
    firewall_allow_port
    systemctl enable squid >/dev/null 2>&1
    systemctl restart squid >/dev/null 2>&1
    if ! systemctl is-active squid >/dev/null 2>&1; then
        echo -e "\033[31m Squid 启动失败，请执行: journalctl -u squid -n 30 --no-pager\033[0m"
        exit 1
    fi

    echo -e "\033[32m 3.Squid 安装完成！\033[0m"
    echo -e "\033[33m 监听: HTTP ${DEFAULT_PORT_HTTP} | HTTPS ${DEFAULT_PORT_HTTPS} | 证书 ${TLS_DOMAIN}\033[0m"
    show_client_info
}

# ========== 交互式卸载 ==========
remove_squid(){
    check_root
    read -p "确定卸载?(y/n):" op
    [ "${op}" != "y" ] && echo "已取消" && exit 0
    remove_squid_force
}

# ========== 修改密码 ==========
change_passwd(){
    check_root
    [ ! -d /etc/squid ] && mkdir -p /etc/squid
    read -p "输入新密码:" newpass
    htpasswd -bc ${SQUID_PASS} ${DEFAULT_USER} ${newpass}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 密码修改成功(新密码未在此显示，请自行记录)\033[0m"
    echo "HTTP:  http://${DEFAULT_USER}:<你的密码>@${SERVER_IP}:${DEFAULT_PORT_HTTP}"
    echo "HTTPS: https://${DEFAULT_USER}:<你的密码>@${TLS_DOMAIN}:${DEFAULT_PORT_HTTPS}"
}

# ========== 状态查看 ==========
status_squid(){
    echo "========== Squid 运行信息 =========="
    systemctl status squid | head -15
    echo ""
    echo "监听端口: $(ss -tulpn | grep squid 2>/dev/null || echo '未检测到')"
    echo "证书: ${TLS_CERT}"
    if is_auth_enabled; then
        echo -e "认证状态: \033[31m 开启(需密码)\033[0m"
    else
        echo -e "认证状态: \033[32m 关闭(匿名)\033[0m"
    fi
}

# ========== 已安装时重新打印客户端信息 ==========
check_squid(){
    if ! command -v squid >/dev/null 2>&1 || [ ! -f ${SQUID_CONF} ]; then
        echo -e "\033[31m Squid 未安装，请先执行: bash install-squid.sh install\033[0m"
        exit 1
    fi
    echo "========== Squid 安装检查 =========="
    if systemctl is-active squid >/dev/null 2>&1; then
        echo -e "服务状态: \033[32m 运行中\033[0m"
    else
        echo -e "服务状态: \033[31m 未运行(可执行: systemctl start squid)\033[0m"
    fi
    echo "监听端口: $(ss -tulpn | grep squid 2>/dev/null || echo '未检测到')"
    grep -E '^http_port|^https_port' ${SQUID_CONF} 2>/dev/null
    echo ""
    show_client_info
}

# ========== 开启密码认证 ==========
pass_on(){
    check_root
    sed -i 's/^#auth_param/auth_param/g' ${SQUID_CONF}
    sed -i 's/^#acl authenticated/acl authenticated/g' ${SQUID_CONF}
    sed -i 's/http_access allow all/#http_access allow all/g' ${SQUID_CONF}
    sed -i '/http_access allow localnet/a\http_access allow authenticated' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 密码认证已开启｜账号: ${DEFAULT_USER} (密码见 passwd 命令修改，不在此显示)\033[0m"
    show_client_info
}

# ========== 关闭密码认证 ==========
pass_off(){
    check_root
    sed -i 's/^auth_param/#auth_param/g' ${SQUID_CONF}
    sed -i 's/^acl authenticated/#acl authenticated/g' ${SQUID_CONF}
    sed -i 's/#http_access allow all/http_access allow all/g' ${SQUID_CONF}
    sed -i '/http_access allow authenticated/d' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 密码认证关闭｜匿名直通\033[0m"
    show_client_info
}

# ========== 开放全网 ==========
allow_all(){
    check_root
    sed -i 's/acl localnet src 192.168.0.0\/16/#&/g' ${SQUID_CONF}
    sed -i '/http_access allow all/i\acl all_ip src 0.0.0.0/0\nhttp_access allow all_ip' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 已开放全网访问\033[0m"
}

# ========== 仅内网 ==========
allow_lan(){
    check_root
    sed -i '/all_ip/d' ${SQUID_CONF}
    sed -i 's/#acl localnet src 192.168.0.0\/16/acl localnet src 192.168.0.0\/16/g' ${SQUID_CONF}
    systemctl reload squid >/dev/null 2>&1
    echo -e "\033[32m 已限制仅内网访问\033[0m"
}

# ========== 参数校验｜禁止多参数 ==========
if [ $# -gt 1 ];then
    echo -e "\033[31m 错误：只能传入单个参数！\033[0m"
    exit 1
fi

# ========== 逻辑入口 ==========
case $1 in
install) install_squid;;
pass_on) pass_on;;
pass_off) pass_off;;
remove) remove_squid;;
passwd) change_passwd;;
status) status_squid;;
check) check_squid;;
allow_all) allow_all;;
allow_lan) allow_lan;;
help|"") show_help;;
*) echo -e "\033[31m 参数错误！输入 help 查看帮助\033[0m";;
esac
EOF
chmod +x install-squid.sh
echo -e "\033[32m ✅ 使用示例： bash install-squid.sh install | bash install-squid.sh check \033[0m"
```

```
针对低配置squid优化（效果明显）

# 备份原配置，文件名包含当前日期，方便恢复
cp /etc/squid/squid.conf /etc/squid/squid.conf.bak.$(date +%Y%m%d_%H%M%S)
cat > /etc/squid/squid.conf <<'EOF' | tee /etc/squid/squid.conf
# ====================== 网络端口配置 ======================
# HTTP 明文代理(已按要求注释关闭)
# http_port 7028

# HTTPS 代理(TLS 终结，公网客户端推荐)
# 修正：恢复你原来的宝塔证书文件名
https_port 7029 tls-cert=/www/server/panel/vhost/cert/igozhang.cc/igozhang.cc.pem tls-key=/www/server/panel/vhost/cert/igozhang.cc/igozhang.cc.key

# 关闭所有不必要的UDP端口和集群协议
icp_port 0
htcp_port 0
snmp_port 0

# ====================== 内存与缓存优化（核心） ======================
# 内存缓存上限：1C1G服务器建议256MB（预留系统和磁盘索引内存）
cache_mem 256 MB

# 内存中单个对象最大大小：只缓存小文件到内存
maximum_object_size_in_memory 4 MB

# 磁盘缓存配置：512M磁盘≈16M内存索引
cache_dir ufs /var/spool/squid 512 8 128

# 磁盘缓存单个对象最大大小：避免大文件挤占缓存空间
maximum_object_size 64 MB

# 缓存替换策略：LFUDA更适合低内存环境
cache_replacement_policy heap LFUDA

# 缓存高低水位：提前开始清理，避免内存突然暴涨
cache_swap_low 80
cache_swap_high 90

# DNS缓存优化：减小缓存大小，加快过期
ipcache_size 256
ipcache_low 90
ipcache_high 95
fqdncache_size 128

# ====================== 进程与CPU优化 ======================
# 1核CPU强制使用1个worker进程
workers 1

# 限制最大文件描述符
max_filedesc 8192

# 限制每个客户端最大连接数（修正：Squid 6.5正确参数名）
client_ip_max_connections 32

# ====================== 日志优化 ======================
# 关闭访问日志，大幅减少磁盘IO和CPU占用
cache_access_log none

# 错误日志只记录严重错误
cache_log /var/log/squid/cache.log 2

# ====================== 访问控制 ======================
acl localnet src 10.0.0.0/8
acl localnet src 172.16.0.0/12
acl localnet src 192.168.0.0/16

# HTTPS代理必需的ACL
acl SSL_ports port 443
acl CONNECT method CONNECT
http_access allow CONNECT SSL_ports

# 默认关闭密码认证
# auth_param basic program /usr/lib64/squid/basic_ncsa_auth /etc/squid/passwd
# auth_param basic realm Proxy-Auth
# auth_param basic children 5
# acl authenticated proxy_auth REQUIRED

http_access allow localnet
http_access allow all
http_access deny all

# ====================== 安全与性能优化 ======================
forwarded_for off
request_header_access Allow allow all

# 隐藏Squid版本号
via off
httpd_suppress_version_string on

# 关闭不必要的请求头处理
request_header_access X-Forwarded-For deny all
request_header_access Via deny all

# 优化连接超时，快速释放空闲连接
connect_timeout 1 minute
peer_connect_timeout 30 seconds
read_timeout 2 minutes
write_timeout 2 minutes
client_lifetime 1 hour
half_closed_clients off

# 持久连接优化
client_persistent_connections on
server_persistent_connections on
persistent_request_timeout 60 seconds

# ====================== 缓存策略优化 ======================
# 不缓存动态内容和查询请求
refresh_pattern -i (/cgi-bin/|\?|.php|.asp|.jsp) 0 0% 0

# 静态文件缓存时间适当延长
refresh_pattern -i \.(jpg|jpeg|png|gif|css|js|ico)$ 1440 80% 10080

# 默认缓存策略
refresh_pattern ^ftp:           1440    20%     10080
refresh_pattern ^gopher:        1440    0%      1440
refresh_pattern .               0       20%     4320
EOF

优化内存参数：
# 临时生效
sysctl -w net.ipv4.tcp_keepalive_time=300
sysctl -w net.ipv4.tcp_keepalive_intvl=30
sysctl -w net.ipv4.tcp_keepalive_probes=3
sysctl -w net.ipv4.tcp_fin_timeout=30
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.ipv4.tcp_max_tw_buckets=5000
sysctl -w net.ipv4.tcp_nodelay=1

# 永久生效
cat >> /etc/sysctl.conf << EOF
net.ipv4.tcp_keepalive_time=300
net.ipv4.tcp_keepalive_intvl=30
net.ipv4.tcp_keepalive_probes=3
net.ipv4.tcp_fin_timeout=30
net.ipv4.tcp_tw_reuse=1
net.ipv4.tcp_max_tw_buckets=5000
net.ipv4.tcp_nodelay=1
EOF

# 清理旧缓存+初始化新缓存
rm -rf /var/spool/squid/* && squid -z
# 检查配置语法
squid -k parse
# 启动服务
systemctl start squid
```
