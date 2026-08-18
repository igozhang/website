# 使用dashboard部署k8s应用

```
env
CentOS Linux release 7.9.2009 (Core)
kubectl v1.19.16
3+3
NAME        STATUS   ROLES    AGE    VERSION
igo-k8s-1   Ready    master   3h5m   v1.19.16
igo-k8s-2   Ready    master   171m   v1.19.16
igo-k8s-3   Ready    master   171m   v1.19.16
igo-k8s-4   Ready    <none>   170m   v1.19.16
igo-k8s-5   Ready    <none>   170m   v1.19.16
igo-k8s-6   Ready    <none>   170m   v1.19.16
```

用户可以直接输入要部署应用的名字、镜像、副本数等信息；也可以上传 YAML配置文件。如果是上传配置文件，则可以创建任意类型的资源，不仅仅是 Deployment

dashboard作用相当于可视化的kubectl，支持json,yaml以及部分表格形式创建编辑各种资源，十分方便
