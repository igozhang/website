# rook_ceph_troubleShooting

```
1.
新部署rook报
health: HEALTH_WARN
    clock skew detected on mon.b, mon.c

解决方法a
kubectl -n rook-ceph edit cm rook-config-override -o yaml
apiVersion: v1
data:
  config: |
    [global]
    mon clock drift allowed = 0.5
kind: ConfigMap
metadata:

删除重建mon节点（只要1分钟）
kubectl -n rook-ceph delete pod $(kubectl -n rook-ceph get pods -o custom-columns=NAME:.metadata.name --no-headers| grep mon)
再看：
  cluster:
    id:     f3d157d0-77d5-4648-96d0-6e7e56552767
    health: HEALTH_OK

方法b
dashboard操作configmap
[global]
#Default of 0.05 is too aggressive for my cluster. (seconds)
mon clock drift allowed = 0.1
#K8s image-gc-low-threshold is 80% - not much point warning
before that point. (percent)
mon data avail warn = 20
```

```
2.
集群全体挂机后osd始终不能running
OSD pod permissions broken, unable to open OSD superblock after node restart
mount错误

原因
当时加盘的时候自动加的，重启后盘符变了导致osd失败

解决
可以xml指定盘符
```
