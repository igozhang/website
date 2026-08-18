# ubuntu20_docker_ins

ubuntu20.04安装及卸载docker20.10.24

```
# cat /etc/os-release
NAME="Ubuntu"
VERSION="20.04.6 LTS (Focal Fossa)"

# docker --version
Docker version 20.10.24, build 297e128
```

## 安装

```
安装依赖
apt install -y apt-transport-https ca-certificates curl gnupg lsb-release

添加阿里云Docker GPG密钥
curl -fsSL http://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

添加Docker源（指定Ubuntu 20.04/focal）
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] http://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

更新源
apt update

查看可安装的Docker版本（可选，验证版本存在）
apt-cache madison docker-ce | grep 20.10.24

安装指定版本Docker CE、containerd
apt install -y docker-ce=5:20.10.24~3-0~ubuntu-focal docker-ce-cli=5:20.10.24~3-0~ubuntu-focal containerd.io=1.6.20-1

锁定Docker版本（防止自动更新）
apt-mark hold docker-ce docker-ce-cli containerd.io

---验证
查看Docker版本
docker --version
预期输出：Docker version 20.10.24, build 297e128

启动并设置开机自启
systemctl enable --now docker
systemctl status docker
```

## 卸载

```
#!/bin/bash

# 定义需要处理的包列表
PACKAGES=(
    "docker-ce=5:20.10.24~3-0~ubuntu-focal"
    "docker-ce-cli=5:20.10.24~3-0~ubuntu-focal"
    "containerd.io=1.6.20-1"
)

# 提取纯包名（去掉版本号部分），用于 apt-mark 和 remove 命令
# 因为 apt-mark unhold 和 apt remove 通常只需要包名，不需要指定具体版本来解锁或删除
PKG_NAMES=("docker-ce" "docker-ce-cli" "containerd.io")

echo ">>> 步骤 1: 检查当前锁定状态..."
apt-mark showhold

echo ""
echo ">>> 步骤 2: 解除包锁定 (unhold)..."
# 必须先用纯包名解除锁定，否则删除可能会失败
sudo apt-mark unhold "${PKG_NAMES[@]}"

if [ $? -eq 0 ]; then
    echo ">>> 锁定已解除。"
else
    echo ">>> 警告: 解除锁定过程中出现错误，但仍尝试继续删除..."
fi

echo ""
echo ">>> 步骤 3: 删除软件包及其配置文件 (purge)..."
# 使用 purge 会同时删除配置文件，如果只想删二进制文件保留配置，改用 remove
# 这里指定版本号是为了确保删除的是你提到的特定版本（虽然通常卸载最新版即可）
# 注意：apt remove 后面跟 version 有时在某些 apt 版本中行为不同，最稳妥是直接删包名
sudo apt-get purge -y "${PKG_NAMES[@]}"

echo ""
echo ">>> 步骤 4: 清理不再需要的依赖包..."
sudo apt-get autoremove -y

echo ""
echo ">>> 步骤 5: 验证删除结果..."
dpkg -l | grep -E "docker-ce|containerd.io"

if [ $? -ne 0 ]; then
    echo ">>> 成功：相关包已完全移除。"
else
    echo ">>> 注意：上述列出的包可能仍存在于系统中（可能是其他版本或相关包如 docker-ce-rootless-extras）。"
fi

echo ""
echo ">>> 可选：清理残留的 Docker 数据目录 (慎用！这将删除所有容器和镜像)"
read -p "是否删除 /var/lib/docker 和 /var/lib/containerd？(y/N): " confirm
if [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]]; then
    sudo rm -rf /var/lib/docker
    sudo rm -rf /var/lib/containerd
    echo "数据目录已清理。"
else
    echo "保留了数据目录。"
fi
```
