# docker_ts,docker排错

```
1.
问题:文件过大导致宿主机磁盘被写满
ls /proc/mounts 可以发现容器OverlayFS 对应的 lowerdir 和 upperdir,lowerdir只读,upperdir可写,所有写入文件都会再upperdir;

解析: 
以最常用的文件系统xfs和ext4中的xfs举例;
XFS 文件系统缺省都有 Quota 的特性，这个特性可以为 Linux 系统里的一个用户（user），一个用户组（group）或者一个项目（project）来限制它们使用文件系统的额度（quota），也就是限制它们可以写入文件系统的文件总量;
首先我们要使用 XFS Quota 特性，必须在文件系统挂载的时候加上对应的 Quota 选项，比如我们目前需要配置 Project Quota，那么这个挂载参数就是"pquota";

第一点，新建的目录 /tmp/xfs_prjquota，我们想对它做 Quota 限制。所以在这里要对它打上一个 Project ID。第二点，通过 xfs_quota 这条命令，我们给 /tmp/xfs_prjquota 打上 Project ID 值 101，这个 101 是我随便选的一个数字，就是个 ID 标识，你先有个印象。在后面针对 Project 进行 Quota 限制的时候，我们还会用到这个 ID
# mkdir -p  /tmp/xfs_prjquota
# xfs_quota -x -c 'project -s -p /tmp/xfs_prjquota 101' /
Setting up project 101 (path /tmp/xfs_prjquota)...
Processed 1 (/etc/projects and cmdline) paths for project 101 with recursion depth infinite (-1).
最后，我们还是使用 xfs_quota 命令，对 101（我们刚才建立的这个 Project ID）做 Quota 限制。你可以执行下面这条命令，里面的"-p bhard=10m 101"就代表限制 101 这个 project ID，限制它的数据块写入量不能超过 10MB;
做好限制后我们dd写20M数据就会报错:"No space left on device";

解决: (docker 命令已经实现了利用quota原理的限制参数)
docker run --storage-opt size=10M -it centos bash
```
