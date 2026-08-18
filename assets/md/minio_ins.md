# minio_ins

Ubuntu 20.04 MinIO（deb 包）完整部署文档  
适用于：minio\_20230804174021.0.0\_amd64.deb

## 一、环境信息

- 操作系统：Ubuntu 20.04 LTS amd64
- MinIO 安装包：minio\_20230804174021.0.0\_amd64.deb  
  https://dl.minio.io/server/minio/release/linux-amd64/archive/
- API 端口：9000
- Web 控制台端口：9001
- 数据目录：/data/minio
- 服务用户：minio-user（deb 包默认）

---

## 二、部署前准备

1. 将 MinIO deb 包上传至服务器目录（示例：/igo/soft/minio）
2. 使用 root 或具有 sudo 权限的账号操作
3. 确保防火墙/安全组可管理

---

## 三、安装 MinIO

```
cd /igo/soft/minio
dpkg -i minio_20230804174021.0.0_amd64.deb
```

安装后默认路径：

- 二进制：/usr/local/bin/minio
- 配置文件：/etc/default/minio
- 服务文件：/lib/systemd/system/minio.service

---

## 四、创建系统用户与数据目录

### 4.1 创建 MinIO 专用系统用户（必须）

```
useradd -r -s /sbin/nologin minio-user
```

### 4.2 创建数据目录

```
mkdir -p /data/minio
```

### 4.3 授权目录权限

```
chown -R minio-user:minio-user /data/minio
chmod -R 755 /data/minio
```

---

## 五、配置 MinIO 启动参数

```
cat > /etc/default/minio << 'EOF'
MINIO_VOLUMES="/data/minio"
MINIO_OPTS="--console-address :9001 --address :9000"
MINIO_ROOT_USER="minioadmin"
MINIO_ROOT_PASSWORD="Minio@2025#admin"
EOF
```

---

## 六、启动服务并设置开机自启

```
systemctl daemon-reload
systemctl start minio
systemctl enable minio
systemctl status minio
```

出现 `active (running)` 即为成功。

---

## 七、防火墙放行端口

```
ufw allow 9000/tcp
ufw allow 9001/tcp
ufw reload
```

---

## 八、Web 控制台访问

访问地址：

```Plain Text

http://服务器IP:9001

```
默认管理员账号：

- 用户：minioadmin

- 密码：Minio@2025#admin

---

# 九、访问控制与权限管理（IAM）

## 9.1 管理员账号说明

- 拥有全局所有权限

- 禁止直接用于业务程序接入

- 仅用于控制台管理与用户创建

## 9.2 创建业务子用户

1. 控制台 → Identity → Users

2. Create User

3. 设置用户名、密码

4. 绑定权限策略

## 9.3 常用内置策略

- consoleAdmin：控制台管理员

- diagnostics：诊断权限

- readonly：只读所有桶

- readwrite：读写所有桶

- writeonly：只写所有桶

## 9.4 自定义桶权限策略示例（test-bucket）
```

JSON

{  
"Version": "2012-10-17",  
"Statement": [  
{  
"Effect": "Allow",  
"Action": [  
"s3:*" ], "Resource": [ "arn:aws:s3:::test-bucket", "arn:aws:s3:::test-bucket/*"  
]  
}  
]  
}

```
---

# 十、用户使用指南（业务接入）

## 10.1 接入信息

- 协议：S3 兼容

- Endpoint：http://服务器IP:9000

- 认证：Access Key + Secret Key

## 10.2 创建 AccessKey

1. 控制台 → Access Keys → Create access key

2. 保存生成的 Access Key / Secret Key（仅显示一次）

## 10.3 支持客户端

- AWS S3 SDK

- MinIO Client (mc)

- Java/Python/Go/PHP/C++ 等

- 各种 OSS 管理工具

## 10.4 安全规范

1. 禁止使用 root 管理员账号接入业务

2. 每个业务独立用户

3. 最小权限原则

4. 定期轮换密钥

5. 密码复杂度：≥12 位，大小写+数字+特殊字符

---

# 十一、常用运维命令
```

Bash

systemctl status minio  
systemctl restart minio  
systemctl stop minio  
journalctl -u minio -f  
dpkg -r minio # 卸载

```
---

# 十二、常见问题排查

## 12.1 启动报错：用户不存在

日志：
```

Plain Text

Failed to determine user credentials

```
解决：
```

Bash

useradd -r -s /sbin/nologin minio-user  
chown -R minio-user:minio-user /data/minio  
systemctl restart minio  
```

## 12.2 启动报错：MINIO\_VOLUMES 未设置

解决：重新写入配置文件并重启。

## 12.3 控制台无法访问

- 检查服务状态
- 检查防火墙/安全组
- 检查端口是否冲突

---

# 十三、生产安全建议

1. 管理员密码复杂度 ≥12 位
2. 业务必须使用 AccessKey
3. 生产环境启用 HTTPS
4. 定期备份 /data/minio
5. 控制台仅内网访问，不暴露公网
