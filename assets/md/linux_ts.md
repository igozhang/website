# linux_ts,排错

```
运行命令报错
# ls
bash: fork: retry: Resource temporarily unavailable

原因: 进程数pids被占满
1 个 init 进程 +1000 个僵尸进程 +1 个 bash 进程 ，总共就是 1002 个进程。
如果 pids Cgroup 也限制了这个容器的最大进程号的数量，限制为 1002 的话，我们在 pids Cgroup 里可以看到，pids.current == pids.max，也就是已经达到了容器进程号数的上限。这时候，如果我们在容器里想再启动一个进程，例如运行一下 ls 命令，就会看到 Resource temporarily unavailable 的错误消息。已经退出的无用进程，却阻碍了有用进程的启动

### On host
# docker ps
CONTAINER ID        IMAGE                      COMMAND             CREATED             STATUS              PORTS               NAMES
09e6e8e16346        registry/zombie-proc:v1   "/app-test 1000"    29 minutes ago      Up 29 minutes                           peaceful_ritchie

# pwd
/sys/fs/cgroup/pids/system.slice/docker-09e6e8e1634612580a03dd3496d2efed2cf2a510b9688160b414ce1d1ea3e4ae.scope

# cat pids.max
1002
# cat pids.current
1002

### On Container
[root@09e6e8e16346 /]# ls
bash: fork: retry: Resource temporarily unavailable
bash: fork: retry: Resource temporarily unavailable

解决办法:
1. 增加pids.max
echo '1024'>pids.max
egrep -v "^$|^#" /etc/security/limits.d/20-nproc.conf  虚机改这并重启,参考值cpu*1024

2. 回收僵尸进程
父进程(容器init,虚机systemd)调用wait()方法回收僵尸进程,缺点wait()是阻塞调用,如果没有zombie进程,进程一直不返回就会阻塞住干不了别的;
父进程(容器init,虚机systemd)调用waitpid()方法,如果没有zombie进程就会直接返回;

每一个 Linux 进程在退出的时候都会进入一个僵尸状态（EXIT_ZOMBIE）；
僵尸进程如果不清理，就会消耗系统中的进程数资源，最坏的情况是导致新的进程无法启动；
僵尸进程一定需要父进程调用 wait() 或者 waitpid() 系统调用来清理，这也是容器中 init 进程必须具备的一个功能。
```

```
2.安装rpm包失败
# yum -y install /tmp/filebeat-7.17.1-x86_64.rpm
Downloading packages:
warning: /tmp/filebeat-7.17.1-x86_64.rpm: Header V4 RSA/SHA512 Signature, key ID d88e42b4: NOKEY

Public key for filebeat-7.17.1-x86_64.rpm is not installed

解决
yum -y install --nogpgcheck /tmp/filebeat-7.17.1-x86_64.rpm
```

```
3. 文件目录只读 
bash: igo.txt: Read-only file system
/data: File cannot be written normally

排错
vim /var/log/kern
May 12 01:32:39 [CRIT] : cndgdlbvdc01-14-150 - [57401848.355202] EXT4-fs error (device dm-2): ext4_journal_check_start:56: Detected aborted journal
May 12 01:32:39 [CRIT] : cndgdlbvdc01-14-150 - [57401848.355823] EXT4-fs (dm-2): Remounting filesystem read-only

ll /dev/mapper  (确认dm-2就是/data)
lsblk

解决
mount | grep data
umount /data; 报(device is busy)
fuser -cu /data (看啥进程啥用户在用)
lsof /data (看啥进程啥用户在用)
fuser -ck /data (如果进程可以干掉,-k)

mount -o remount,rw /data
```
