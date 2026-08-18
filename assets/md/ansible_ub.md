# ansible_ub

搭建一套ansible

1. 纳管10.80.238.31-39
2. 批量执行命令
3. 批量分发文件

env  
ubuntu20.04  
ansible [core 2.12.10]

## 安装验证

```
apt install -y software-properties-common
add-apt-repository --yes --update ppa:ansible/ansible
apt install -y ansible
ansible --version
```

## 添加主机

```
先添加root，再批量创建igo用户，最后批量发key到igo：
ansible 'k8s*' -m user -a "name=igo state=present shell=/bin/bash"
ansible 'k8s*' -m authorized_key -a "user=igo key='{{ lookup('file', '/root/.ssh/id_rsa.pub') }}'"

tee <<EOF >/etc/ansible/hosts 
[k8s:children]
k8s_pro
k8s_pre
k8s_nfs
k8s_midware
[k8s:vars]
ansible_ssh_user=igo
ansible_ssh_port=5022

[k8s_pro]
10.80.222.[51:59]

[k8s_pre]
10.80.222.[31:37]

[k8s_nfs]
10.80.222.39
10.80.222.60

[k8s_midware]
10.80.222.[71:77]

[morocco_k8s]
10.85.222.[221:226]
ansible_ssh_user=igo
ansible_ssh_pass=igo@2026
ansible_ssh_port=5008
ansible_ssh_common_args='-o StrictHostKeyChecking=no'
ansible_host_key_checking=False
EOF

ansible-inventory --list -y
ansible k8s_pre --list-hosts

建议不设置明文密码，而是ssh免密通信
ssh-keygen -t rsa -b 2048 -N "" -f ~/.ssh/id_rsa
for i in {31..36}; do ssh-copy-id root@10.80.238.$i; done
for i in {31..36}; do sshpass -p 'Password@1' ssh-copy-id -o StrictHostKeyChecking=no -p 5008 root@10.80.238.$i; done
```

## 命令使用

```
ansible k8s_pre -m shell -a "uptime"
1. apt安装
ansible k8s_pre_nfs -m apt -a "name=nginx state=present update_cache=yes" --become
2. 服务重启
ansible k8s_pre_nfs -m service -a "name=nginx state=restarted" --become
3. 拷贝文件
ansible k8s_pre_nfs -m copy -a "src=igo.txt dest=/tmp/ mode=0644"
4. 追加文件内容
ansible k8s_pre_nfs -m blockinfile -a "path=/etc/hosts block='10.80.238.31  k8s-pre-31\n10.80.238.32  k8s-pre-32' marker='# ANSIBLE MANAGED BLOCK'" --become
5. 批量修改127主机名，避免普通用户的sudo报错
ansible all -b -m shell -a "sed -i 's/^127.0.1.1.*/127.0.1.1 '\$(hostname)'/' /etc/hosts"
ansible k8s_pre:k8s_pre_nfs -b -m shell -a "sed -i 's/^127.0.1.1.*/127.0.1.1 '\$(hostname)'/' /etc/hosts"
```

## playbook样例

```
cat <<'YAML' | ansible-playbook /dev/stdin -f 1
---
- hosts: "10.80.238.77"
  become: true
  gather_facts: false
  tasks:
    - ansible.builtin.shell: |
        set -euo pipefail
        Hostname=$(ip -4 addr show scope global | awk '/inet / {print $2}' | cut -d/ -f1 | head -n1)
        test -n "$Hostname"
        cp -a /usr/local/zabbix/conf/zabbix_agentd.conf "/usr/local/zabbix/conf/zabbix_agentd.conf.bak.$(date +%Y%m%d%H%M%S)"
        tee /usr/local/zabbix/conf/zabbix_agentd.conf >/dev/null <<EOF
        LogFile=/tmp/zabbix_agentd.log
        Server=10.80.238.38
        ServerActive=10.80.238.38:31051
        Hostname=${Hostname}
        Timeout=20
        ListenPort=10050
        EOF
        systemctl restart zabbix-agent
        systemctl --no-pager -l status zabbix-agent || true
      args:
        executable: /bin/bash
YAML

sys_init.yaml

---
- name: 初始化 31-39 服务器
  hosts: my_servers
  become: yes
  gather_facts: no

  tasks:
    - name: 更新 apt 缓存
      apt:
        update_cache: yes
        cache_valid_time: 3600

    - name: 安装基础工具
      apt:
        name:
          - vim
          - curl
          - wget
          - htop
        state: present

    - name: 关闭防火墙 (ufw)
      service:
        name: ufw
        state: stopped
        enabled: no
```
