# k8s_集群恢复_排错

```
env
kubectl v1.19.16
CentOS Linux release 8.2.2004
```

```
问题1：机器恢复，集群未恢复

机房故障集群全军覆没后，来电了，机器都恢复起来，但是k8s没起来
# kubectl get node
The connection to the server k8s-api.ilinux.io:6443 was refused - did you specify the right host or port?

1.
先检查docker和kubelet进程状态是否正常
异常的话修复问题启动进程

2.
如果进程正常但k8s集群没起来，需要统一重启一下两个进程
# ansible hostAll -m shell -a "systemctl restart docker" -f 6
# ansible hostAll -m shell -a "systemctl restart kubelet" -f 6

3.
再看，就正常了
# kubectl get node
NAME        STATUS   ROLES    AGE   VERSION
igo-k8s-1   Ready    master   13d   v1.19.16
igo-k8s-2   Ready    master   13d   v1.19.16
igo-k8s-3   Ready    master   13d   v1.19.16
igo-k8s-4   Ready    <none>   13d   v1.19.16
igo-k8s-5   Ready    <none>   13d   v1.19.16
igo-k8s-6   Ready    <none>   13d   v1.19.16
```

```
问题2：pod has unbound immediate PersistentVolumeClaims
部署报错
describe
Events:
  Type     Reason            Age   From               Message
  ----     ------            ----  ----               -------
  Warning  FailedScheduling  60s   default-scheduler  0/6 nodes are available: 6 pod has unbound immediate PersistentVolumeClaims.
  Warning  FailedScheduling  60s   default-scheduler  0/6 nodes are available: 6 pod has unbound immediate PersistentVolumeClaims.

两种可能原因
a. storageClass 配错了，改过来
storageClass: csi-rbd-sc
b. pvc被占用，比如改完后删了deployment，没删pvc；
注意：pvc不会自动删除
```

```
问题3：the provided range does not match the current range
部署报错
Error: INSTALLATION FAILED: Service "area-kafka-0-external" is invalid: spec.ports[0].nodePort: Invalid value: 29093: the provided range does not match the current range

vim /etc/kubernetes/manifests/kube-apiserver.yaml
所有master都添加行
    - --service-node-port-range=10000-39000
无需重启，会自行重启kubelet服务生效
```
