# vcenter_scp传文件

```
env
vcenter6.7
```

```
1.
打开ssh
在vCenter Server Appliance管理界面中，单击访问，然后单击编辑。（5480）
编辑vCenter Server Appliance的访问设置。 选项 描述 启用SSH 登录 启用对vCenter Server Appliance的SSH 访问。 启用DCUI. ...
单击确定以保存设置

2.
vcenter改shell模式
解决vcenter_scp传文件传文件报错问题
chsh -s /bin/bash //切换到bash shell
chsh -s /bin/appliancesh //切回到Appliance Shell
```
