# sshd

```
rockey9 系统关防火墙，关selinux，sshd监听5008，开启root登录及密码认证，并清理history的命令

sed -i.bak -E 's/^#?PermitRootLogin[[:space:]]+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
sed -i.bak -E 's/^#?Port[[:space:]]+[0-9]+/Port 5008/' /etc/ssh/sshd_config
dnf install -y policycoreutils-python-utils
semanage port -l | grep ssh
semanage port -a -t ssh_port_t -p tcp 5008
semanage port -l | grep ssh
setenforce 0
sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
systemctl stop firewalld
systemctl status firewalld
systemctl restart sshd
systemctl status sshd
history -c

openEuler22 系统关防火墙，关selinux，sshd监听5008，开启root登录及密码认证，并清理history的命令

sed -i.bak -E 's/^#?PermitRootLogin[[:space:]]+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
sed -i.bak -E 's/^#?PasswordAuthentication[[:space:]]+.*/PasswordAuthentication yes/' /etc/ssh/sshd_config
sed -i.bak -E 's/^#?Port[[:space:]]+[0-9]+/Port 5008/' /etc/ssh/sshd_config
dnf install -y policycoreutils-python-utils
semanage port -a -t ssh_port_t -p tcp 5008
setenforce 0
sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
systemctl stop firewalld
systemctl disable firewalld
systemctl restart sshd
history -c

ubuntu22 系统关防火墙，关selinux，sshd监听5008，开启root登录及密码认证，并清理history的命令

sed -i.bak -E 's/^#?PermitRootLogin[[:space:]]+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
sed -i.bak -E 's/^#?PasswordAuthentication[[:space:]]+.*/PasswordAuthentication yes/' /etc/ssh/sshd_config
sed -i.bak -E 's/^#?Port[[:space:]]+[0-9]+/Port 5008/' /etc/ssh/sshd_config
grep -qE '^[[:space:]]*Port[[:space:]]+' /etc/ssh/sshd_config || echo 'Port 5008' >> /etc/ssh/sshd_config
ufw --force disable
systemctl stop ufw 2>/dev/null || true
systemctl disable ufw 2>/dev/null || true
systemctl stop firewalld 2>/dev/null || true
systemctl disable firewalld 2>/dev/null || true
if command -v getenforce >/dev/null 2>&1; then setenforce 0; [ -f /etc/selinux/config ] && sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config; fi
systemctl restart ssh
history -c
history -w
```
