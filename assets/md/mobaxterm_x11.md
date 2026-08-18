# mobaxterm调取liunx图形化界面

```
redhat7.8
mobaxterm-PersonalEdition-v22.1

原理:
X11实现，大部分情况下Xserver和Xclient在同一台机器上，但Xserver和Xclient也可以不在同一个机器，彼此通过X协议通信即可。我们可以通SSH的X11 forwarding转发实现，远程Linux机器运行GUI程序（Xclient），在本地Windows显示（X server）
```

```
yum -y install xorg-x11-xauth xorg-x11-fonts-* xorg-x11-font-utils xorg-x11-fonts-Type1 xclock

如果依然无法打开图形化(有时候可以打开xclock,但是无法打开其他图形化):
yum -y install gnome-classic-session gnome-terminal nautilus-open-terminal control-center liberation-mono-fonts

vim /etc/ssh/sshd_config
X11Forwarding yes
X11UseLocalhost no

systemctl restart sshd

settings>edit session preset>SSH>AdvanceSSHsetting>X11Forwarding勾选
export DISPLAY=$clientWindowsIP:0.0
xclock
```

```
其他办法:
BMC,IPMI
CONSOLE
VNC&VNC_Server搭配使用
Xshell+Xmanager(要花钱)
```

```
报错: xclock Error: Can't open display
原因: 没有设置xserver地址
解决: export DISPLAY=$clientWindowsIP:0.0
```

```
xclock start with the error 'Missing charsets in String to FontSet conversion'

sulution:
$ export LC_ALL=C
```
