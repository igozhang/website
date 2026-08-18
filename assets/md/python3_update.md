# CentOS 7 中 Python 3.11 无伤升级

你想在 CentOS 7 系统中将默认的 Python 3.6.8 升级到 3.11 版本，这个需求我完全理解。需要注意的是，CentOS 7 系统依赖 Python 2.7，直接替换系统默认 Python 可能导致 yum 等工具失效，因此我们采用**源码编译安装**的方式，并保留系统原有版本。

### 一、升级前准备（安装依赖）

首先安装编译 Python 所需的依赖包，避免编译过程中报错：

```
# 先更新系统包

sudo yum update -y

# 安装编译依赖

sudo yum install -y gcc openssl-devel bzip2-devel libffi-devel zlib-devel wget make
```

### 二、下载并编译 Python 3.11

1. 下载 Python 3.11 源码包（以 3.11.9 为例，可替换为最新的 3.11 版本）：

```
# 切换到临时目录

cd /usr/local/src

# 下载源码包（如果下载慢，可替换为国内镜像，如 https://mirrors.huaweicloud.com/python/3.11.9/Python-3.11.9.tgz）

sudo wget https://www.python.org/ftp/python/3.11.9/Python-3.11.9.tgz

# 解压源码包

sudo tar -zxvf Python-3.11.9.tgz

# 进入解压目录

cd Python-3.11.9
```

2. 配置编译参数（关键：指定安装路径，避免覆盖系统版本）：

```
# 配置编译选项：--prefix 指定安装路径，--enable-optimizations 优化性能，--with-ssl 解决 pip 报错问题

sudo ./configure --prefix=/usr/local/python311 --enable-optimizations --with-ssl
```

3. 编译并安装（`-j` 后接 CPU 核心数，加快编译速度，比如 `make -j4`）：

```
# 编译

sudo make -j $(nproc)

# 安装（不会覆盖系统原有 Python）

sudo make altinstall
```

### 三、配置环境变量（设置 python3.11 为默认）

1. 创建软链接（方便调用）：

```
# 为 python3.11 创建软链接（避免覆盖系统 python3）

sudo ln -sf /usr/local/python311/bin/python3.11 /usr/bin/python3.11

sudo ln -sf /usr/local/python311/bin/pip3.11 /usr/bin/pip3.11
```

2. 验证安装结果：

```
# 查看 Python 3.11 版本

python3.11 --version

# 查看 pip 3.11 版本

pip3.11 --version
```

正常输出应为`Python 3.11.9` 和对应的 pip 版本。

### 四、（可选）设置 python3 指向 3.11

如果希望输入 `python3` 时默认调用 3.11 版本（不影响系统原有工具）：

```
# 先备份原有软链接

sudo mv /usr/bin/python3 /usr/bin/python3.6.8.bak

# 创建新软链接

sudo ln -sf /usr/local/python311/bin/python3.11 /usr/bin/python3

# 验证

python3 --version  # 应输出 3.11.9
```

### 五、修复 yum 依赖（关键！）

如果修改了 `python3` 软链接，需确保 yum 仍使用系统 Python 2.7：

```
# 编辑 yum 配置文件

sudo vi /usr/bin/yum

# 将第一行的 #!/usr/bin/python 改为 #!/usr/bin/python2.7

# 同理修复 yum-config-manager

sudo vi /usr/libexec/urlgrabber-ext-down

# 将第一行的 #!/usr/bin/python 改为 #!/usr/bin/python2.7
```

### 总结

1. **核心方式**：通过源码编译安装 Python 3.11，避免覆盖系统默认版本导致工具失效；
2. **关键步骤**：安装编译依赖 → 源码编译 → 配置软链接 → 修复 yum 依赖；
3. **验证标准**：执行 `python3.11 --version` 输出 3.11.x，`yum` 命令能正常使用。

注意：CentOS 7 不建议完全替换系统 Python，使用`python3.11` 或 `python3`（软链接）调用新版本即可，原有 `python`（指向 2.7）和 `python3.6.8` 仍保留。

```
notes

碰到问题，回退重新编译：

# 进入Python 3.11源码目录
cd /data/python311/Python-3.11.9
# 清理编译残留
sudo make clean
sudo rm -rf ./build
```
