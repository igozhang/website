# hint

## helm

```
1. 部署命令使用helm upgrade –install，兼容两种场景
2. namespace等其他参数都尽量在helm安装命令里指定，其中namespace在命令里创建
3. 避免使用最新版本，应该使用较新的稳定版本
5. 使用nodeport模式
6. 使用"--set"指定参数
7. 使用域名"xx.igozhang.cn"
```

## trouble-shooting

```
1. 给出方案前先确认必要信息；
2. 每次不超过一条命令，可以是聚合命令，也可以是多行，实现一次对话一次复制，根据输出再进行下一步，需要说明命令用途；
3. 如需排查日志，每次不要取超过60行,多个日志可以各取60行，帮助快速定位原因；
4. 默认有必要服务器的ssh权限，多台服务器的情况需说明在哪台服务器执行；
```

## 日常

```
1. 评估当前环境版本等必要参数，提供对应的完整准确下载链接，不是下载方法；
2. 服务端无该文件时：“生成文件，简洁说明内容格式及一条样例，均保持注释'#'状态”；
```

## 问题报告

```
写一篇md格式的问题报告，包含：
1. 问题现象；
2. 问题原因；
3. 解决方案；
4. 后续预防；
要简洁，不要太冗长，可以包含必要环境信息
```

## file-server

```
功能：文件服务器，fs服务
1. 文件下载： 将宿主机的/data/dls目录挂载到容器内，提供下载
2. 文件中转： 在输入框输入URL，下载对应文件到/data/dls/fs目录

绘制16*16 svg格式魂斗罗风格的favicon
网站title “file-server”
整体护眼风格
文件路径需要使用相对路径；
整个项目统一字符编码
使用模块化分体式设计，所有功能模块考虑后续扩展和运维需求

1. 使用go语言实现需求，监听3002端口，不依赖服务器环境；
2. 一个简单的前端,使用精简的代码量提供易用的服务；
3. 目标服务器系统ubuntu20.04
4. 目标部署形式（容器，K8S）
5. 考虑桌面端和移动端两种访问需求
6. 目标ingress的域名使用“dl.igozhang.cn”
7. 在项目目录编写readme.md，在deploy目录编写部署手册 deploy.md，约束条件写入checklist.md，这几个文件每次变更都保持更新
8. 项目根目录添加 .gitattributes，强制文本用 LF，Windows 常用 CRLF，Linux/macOS 用 LF

上线流程
1. 先编写代码
2. 再打包成镜像
3. 然后上传镜像到krccr.ccs.tencentyun.com/igokorea仓库；tag格式：年月日-时分秒；
4. k8s部署服务 （yaml文件放在项目deploy子目录的下级k8s目录里面）
5. 最后根据需要添加一个nginx-conf文件，利用wp服务器反向代理提供服务（也可能直接提供服务不共用wp服务器）
nginx-conf参考trans服务的：
tee /www/server/panel/vhost/nginx/k8s-includes/trans.conf <<'EOF'
# trans：igozhang.cn/trans → Ingress(trans.igozhang.cn)
location = /trans {
    return 302 /trans/;
}
location ^~ /trans/ {
    proxy_pass http://127.0.0.1:30080/;
    proxy_set_header Host trans.igozhang.cn;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_connect_timeout 7200s;
    proxy_read_timeout 7200s;
    proxy_send_timeout 7200s;
    client_max_body_size 0;
}
EOF
```
