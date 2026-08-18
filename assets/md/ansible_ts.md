# ansible排错,ansible_ts

```
issue01

server
ansible 2.6.20
centos  6.9

client
centos  6.9

问题描述:
ansible连接不上对应机器,连接超时或失败
# ansible -i ./igo.hosts ames   -m shell -a 'ls /tmp' 
172.16.217.246 | UNREACHABLE! => {
    "changed": false, 
    "msg": "Failed to connect to the host via ssh: ", 
    "unreachable": true
}

# cat igo.hosts 
[ames]
172.16.217.246    ansible_port=22 ansible_user=mes-deploy ansible_ssh_pass='_i,PBDZ@t1E~m#c#' ansible_sudo_pass='_i,PBDZ@t1E~m#c#' ansible_become=yes  ansible_become_method=sudo

# ansible -i ./igo.hosts ames   -m shell -a 'ls /tmp'   -vvvv
ansible 2.6.20
  config file = /etc/ansible/ansible.cfg
  configured module search path = [u'/root/.ansible/plugins/modules', u'/usr/share/ansible/plugins/modules']
  ansible python module location = /usr/lib/python2.6/site-packages/ansible
  executable location = /usr/bin/ansible
  python version = 2.6.6 (r266:84292, Aug 18 2016, 15:13:37) [GCC 4.4.7 20120313 (Red Hat 4.4.7-17)]
Using /etc/ansible/ansible.cfg as config file
setting up inventory plugins
Parsed /data/tools-developing/deployjavaapp/inventory/uat/igo.hosts inventory source with ini plugin
Loading callback plugin minimal of type stdout, v2.0 from /usr/lib/python2.6/site-packages/ansible/plugins/callback/minimal.pyc
META: ran handlers
<172.16.217.246> ESTABLISH SSH CONNECTION FOR USER: mes-deploy
<172.16.217.246> SSH: EXEC sshpass -d12 ssh -vvv -C -o ControlMaster=auto -o ControlPersist=60s -o StrictHostKeyChecking=no -o Port=22 -o User=mes-deploy -o ConnectTimeout=10 -o ControlPath=/root/.ansible/cp/46e725bd3a 172.16.217.246 '/bin/sh -c '"'"'echo ~mes-deploy && sleep 0'"'"''
<172.16.217.246> (255, '', 'OpenSSH_5.3p1, OpenSSL 1.0.1e-fips 11 Feb 2013\ndebug1: Reading configuration data /etc/ssh/ssh_config\r\ndebug1: Applying options for *\r\ndebug1: auto-mux: Trying existing master\r\ndebug2: fd 3 setting O_NONBLOCK\r\ndebug2: mux_client_hello_exchange: master version 4\r\ndebug3: mux_client_request_forwards: requesting forwardings: 0 local, 0 remote\r\ndebug3: mux_client_request_session: entering\r\ndebug3: mux_client_request_alive: entering\r\ndebug3: mux_client_request_alive: done pid = 129804\r\ndebug3: mux_client_request_session: session request sent\r\ndebug1: mux_client_request_session: master session id: 12\r\ndebug3: mux_client_read_packet: read header failed: Broken pipe\r\ndebug2: Control master terminated unexpectedly\r\n')
172.16.217.246 | UNREACHABLE! => {
    "changed": false, 
    "msg": "Failed to connect to the host via ssh: OpenSSH_5.3p1, OpenSSL 1.0.1e-fips 11 Feb 2013\ndebug1: Reading configuration data /etc/ssh/ssh_config\r\ndebug1: Applying options for *\r\ndebug1: auto-mux: Trying existing master\r\ndebug2: fd 3 setting O_NONBLOCK\r\ndebug2: mux_client_hello_exchange: master version 4\r\ndebug3: mux_client_request_forwards: requesting forwardings: 0 local, 0 remote\r\ndebug3: mux_client_request_session: entering\r\ndebug3: mux_client_request_alive: entering\r\ndebug3: mux_client_request_alive: done pid = 129804\r\ndebug3: mux_client_request_session: session request sent\r\ndebug1: mux_client_request_session: master session id: 12\r\ndebug3: mux_client_read_packet: read header failed: Broken pipe\r\ndebug2: Control master terminated unexpectedly", 
    "unreachable": true
}

solution01:
ansible_sftp 传输失败导致无法连接,检查客户端发现/usr/lib/openssh/sftp-server不存在;
添加软连接问题解决
ln -s /usr/libexec/openssh/sftp-server /usr/lib/openssh/sftp-server

# ll /usr/lib/openssh/sftp-server 
lrwxrwxrwx 1 root root 32 Jun  8 15:32 /usr/lib/openssh/sftp-server -> /usr/libexec/openssh/sftp-server

solution02:
将ansible升级到2.9可以解决问题;
```
