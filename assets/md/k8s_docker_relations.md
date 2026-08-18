# k8s,docker兼容关系说明

1. K8S v1.20之前版本完全兼容Docker。
2. K8S v1.20到v1.22版本，仍然支持Docker，但在kubelet启动时打印告警。
3. 自K8S v1.23版本开始，不再支持Docker引擎，但仍兼容原有的Docker镜像，并且用Docker build命令生成的镜像也能正常使用。
4. 自K8S v1.23版本开始，可使用containerd或CRI-O容器运行时。
