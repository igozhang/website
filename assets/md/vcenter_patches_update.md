# VCenter补丁升级

Env

Vcenter6.7

步骤：

1. 打个快照稳稳哒

2. 下载VmWare官网最新补丁包

<https://my.vmware.com/cn/group/vmware/patch#search>

vcenter6.7补丁升级

3. 所在Esxi主机上传vcenter补丁包

存储->datastore1->数据存储浏览器->上载

4. 挂载iso

5. https://vcip:5480 左侧更新->检查更新

6. 安装更新

可用更新存在包含关系，选择最新更新即可 转储并安装

简述命令行升级：

```
需要先将iso挂载到vcenter
software-packages stage --iso     #转储 ISO
software-packages list --staged   #查看已转储的内容
software-packages install --staged  #安装已转储的 RPM
```
