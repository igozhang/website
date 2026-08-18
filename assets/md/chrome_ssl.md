# chrome_ssl

OpenCloudOS 9 上的远程 Chrome 完整部署记录  
✅ 远程 Chrome 完整部署记录（OpenCloudOS 9 + noVNC + SSH 隧道）

## 一、环境说明

- 服务器系统：OpenCloudOS 9
- 核心目标：搭建轻量远程图形桌面，通过浏览器访问 Chrome，充分利用显示空间，无冗余报错
- 访问方式：SSH 隧道本地访问（彻底规避 HTTPS 强制跳转、证书问题）
- <https://igozhang.cn:9222/vnc.html>

---

## 二、一、服务器端：安装依赖（仅执行一次）

```
dnf install -y xorg-x11-server-Xvfb openbox x11vnc novnc google-chrome-stable
```

---

## 三、服务器端：一键启动脚本（可重复执行）

### 1. 创建启动脚本 `start_chrome.sh`

```
#!/bin/bash
# 清理旧环境
pkill -9 -f "Xvfb|x11vnc|websockify|chrome|openbox"
rm -f /tmp/.X1-lock

# 1. 启动 1920x1080 宽屏虚拟显示器（可根据需求调整分辨率）
Xvfb :1 -screen 0 1920x1080x16 >/dev/null 2>&1 &
export DISPLAY=:1
sleep 1

# 2. 启动 openbox 窗口管理器（用于窗口最大化管理）
openbox --replace >/dev/null 2>&1 &
sleep 1

# 3. 启动 x11vnc 服务
x11vnc -display :1 -nopw -listen 127.0.0.1 -xkb -forever >/dev/null 2>&1 &

# 4. 启动 noVNC 代理（复用 9222 端口，无需额外开防火墙）
/usr/bin/websockify --web /usr/share/novnc 127.0.0.1:9222 127.0.0.1:5900 >/dev/null 2>&1 &

# 5. 启动 Chrome：最大化启动、禁用 DBus 报错、保留完整界面
su - myuser -c "DISPLAY=:1 google-chrome \
--no-sandbox \
--disable-gpu \
--no-dbus \
--autoplay-policy=no-user-gesture-required \
--start-maximized \
--disable-pinch \
--force-device-scale-factor=1 &" >/dev/null 2>&1 &

echo "✅ 服务启动成功！"
echo "Mac 端执行：ssh -L 9222:127.0.0.1:9222 root@igozhang.cn"
echo "浏览器访问：http://127.0.0.1:9222/vnc.html"
```

### 2. 创建停止脚本 `stop_chrome.sh`

```
#!/bin/bash
pkill -9 -f "Xvfb|x11vnc|websockify|chrome|openbox"
rm -f /tmp/.X1-lock
echo "✅ 服务已停止"
```

### 3. 赋予脚本执行权限

```
chmod +x start_chrome.sh stop_chrome.sh
```

---

## 四、Mac 端：连接与访问步骤

### 1. 建立 SSH 隧道（保持终端窗口打开）

```
ssh -L 9222:127.0.0.1:9222 root@igozhang.cn
```

输入服务器密码，**不要关闭终端**，否则隧道断开。

### 2. 浏览器访问

打开 Chrome/Safari，输入地址：

```Plain Text

http://127.0.0.1:9222/vnc.html

```
点击 **Connect** 进入远程桌面。

### 3. noVNC 客户端关键设置（首次访问配置）

1. 点击左侧⚙️ **设置图标**

2. **缩放模式**：选择「**缩放以适应 (Scale to fit)**」

3. 勾选「**限制/裁切窗口大小 (Clip to window)**」

4. 取消勾选「仅查看」，确保可操作

5. 点击 **Save** 保存（后续访问自动生效）

---

## 五、关键配置说明

### 1. 分辨率调整

如需更大显示空间，修改 `Xvfb` 启动参数即可：

- 超宽屏：`Xvfb :1 -screen 0 2560x1440x16`

- 标准宽屏：`Xvfb :1 -screen 0 1920x1080x16`（默认）

---

## 六、日常使用流程

1. 服务器端：执行 `./start_chrome.sh` 启动服务

2. Mac 端：终端执行 SSH 隧道命令

3. 浏览器访问 `http://127.0.0.1:9222/vnc.html`，使用远程 Chrome

4. 用完后：服务器端执行 `./stop_chrome.sh` 停止服务，关闭 Mac 终端

---

## 七、常见问题处理

### 1. 浏览器自动跳转 HTTPS

- 解决方案：使用 SSH 隧道本地访问 `127.0.0.1`，彻底规避 HTTPS 强制升级

- 若仍有问题：清除浏览器 HSTS 缓存（Chrome 访问 `chrome://net-internals/#hsts`，删除 `igozhang.cn` 策略）

### 2. Chrome 未最大化

- 解决方案：进入桌面后，按 `F11` 两次（第一次全屏，第二次恢复最大化，保留标签栏）

- 确保启动参数包含 `--start-maximized`

### 3. 连接重置/无法访问

- 解决方案：检查服务器服务状态，重新执行启动脚本；确认 SSH 隧道正常运行

---

## 八、方案优势

✅ 轻量低耗：总内存占用 < 500MB，不卡服务器

✅ 端口复用：仅用 9222 端口，无需额外开防火墙/安全组

✅ 稳定无坑：SSH 隧道本地访问，彻底解决 HTTPS、证书问题

✅ 界面完整：保留 Chrome 标签栏、地址栏，完全可用

✅ 充分利用空间：1920x1080 宽屏 + 自动最大化，无黑边无空白

---
```

## 九 直接使用证书，无需本地ssl转发方案

```
# 彻底清理残留
pkill -9 -f "Xvfb|x11vnc|websockify|chrome|openbox"
rm -rf /tmp/.X1-lock /tmp/.X11-unix/X1
sleep 1

# 启动虚拟桌面
Xvfb :1 -screen 0 1920x1080x16 >/dev/null 2>&1 &
export DISPLAY=:1
sleep 2

# 启动窗口管理器
openbox --replace >/dev/null 2>&1 &
sleep 1

# 启动x11vnc
x11vnc -display :1 -nopw -listen 127.0.0.1 -xkb -forever >/dev/null 2>&1 &
sleep 1

# 启动noVNC+HTTPS
websockify \
--web /usr/share/novnc \
--cert=/www/server/panel/vhost/cert/wordpress.local/igozhang.cn.pem \
--key=/www/server/panel/vhost/cert/wordpress.local/igozhang.cn.key \
0.0.0.0:9222 127.0.0.1:5900 >/dev/null 2>&1 &
sleep 2

# 启动Chrome
su - myuser -c "DISPLAY=:1 google-chrome \
--no-sandbox \
--disable-gpu \
--no-dbus \
--autoplay-policy=no-user-gesture-required \
--start-maximized &" >/dev/null 2>&1 &

验证访问
ps aux | grep -E "Xvfb|x11vnc|websockify|chrome|openbox"
https://igozhang.cn:9222/vnc.html
```
