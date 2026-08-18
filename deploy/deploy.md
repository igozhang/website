# igozhang 部署手册

面向 **Ubuntu 20.04** 宿主机与已有 Kubernetes 集群；镜像为自包含静态二进制，无宿主机语言运行时依赖。

> **与旧版 `fs` 并存**：同一 **`igo`** 命名空间内，旧站点 **`deployment/fs`**（域名 **`igozhang.cc`**、端口 **3002**、数据 **`/data/igozhang.cc`**）可继续运行；本项目使用 **`deployment/igozhang`**、**`igozhang.cn`**、端口 **3003**、容器内数据根 **`/igozhang.cn`**（宿主机默认 **`/data/igozhang.cn`**），资源名与端口均不冲突。

> **从旧项目名 `igo-doc` 升级**：K8s 资源已统一为 **`igozhang`**（Deployment / Service / Ingress、容器名、标签 `app=igozhang`）。若集群中仍有 **`deployment/igo-doc`** 等旧资源，发版时执行 `kubectl -n igo delete deployment,service,ingress igo-doc --ignore-not-found`，再 `kubectl apply -f deploy/k8s/`。运行日志文件由 **`secret-log/igo-doc.log`** 自动迁移为 **`igozhang.log`**（首次启动时重命名，若新文件已存在则跳过）。

### 与项目根文档的配套关系

- 项目能力、规则表、接口与环境变量：**[README.md](../README.md)**  
- 不变量、易回归点、发版/真机自检（含 Media Session / Android）：**[checklist.md](../checklist.md)**  
- **维护约定**：变更部署步骤、反代约定、HTTPS、探针或上线检查项时，请**同步更新本文件**；若同时影响业务规则或前端契约，请一并更新 **`README.md`** 与 **`checklist.md`**。

## 1. 前置条件

- 集群已安装 **Ingress NGINX**（或兼容 `ingressClassName: nginx` 的控制器）。
- 命名空间 **`igo`** 已存在（与现有 trans 等服务一致）；若不存在：`kubectl apply -f` 自行创建 Namespace。
- 镜像仓库可推送：`krccr.ccs.tencentyun.com/igokorea`。
- 对外域名：**`igozhang.cn`**（Ingress `host` 与 TLS SAN 一致）。

## 2. 数据目录

**两层路径**（见 **`deploy/k8s/deployment.yaml`**）：

| 位置 | 路径 | 说明 |
|------|------|------|
| **容器内** | **`/igozhang.cn`** | 环境变量 **`DATA_ROOT`** 与 **`volumeMounts.mountPath`**，进程读写均在此 |
| **宿主机** | **`/data/igozhang.cn`** | **`volumes.hostPath.path`**；迁移数据时**通常只改这一行** |

清单使用 **`hostPath`** 将宿主机目录挂载到容器 **`/igozhang.cn`**：

- 下载范围即为该目录树（根列表仅展示一级目录名以 **`f`** 开头的项，如 **`fs`**；**`assets`**、**`secret-*`** 等通过专用路由或后台访问）。
- URL 中转写入容器内 **`/igozhang.cn/fs`**（可通过环境变量 `TRANSFER_SUBDIR` 修改子目录名）。
- TLS 证书放在宿主机 **`/data/igozhang.cn/tls/`**（**`igozhang.cn.pem`**、**`igozhang.cn.key`**；供 Ingress Secret 与宝塔 Nginx 共用）。

**`DATA_ROOT`（容器内 `/igozhang.cn`）下主要子目录**（首次启动由进程自动创建缺失目录；**`assets/md/`** 为空时会从镜像内 **`assets/`** 种子复制，并写入内置演示文章 **`docker_ts`**）：

| 路径 | 说明 |
|------|------|
| **`assets/md/`** | 文章 Markdown 正文 |
| **`assets/doc-pic/`** | 文章图片 |
| **`assets/mp3/`** | MP3 音频（体积大，请单独备份；不进文章 ZIP） |
| **`assets/eng/phrases.md`** | 页眉引语（可选；不足 150 条有效行时用内嵌样本） |
| **`assets/eng/word3k-full.md`** | 知键词库（可选；不足 50 条有效行时用内嵌样本） |
| **`secret07/igo.conf`** | 运维配置（首次从内嵌复制；改后须重启 Pod） |
| **`secret07/sqlitedb/doc.db`** | 文章 SQLite 索引 |
| **`secret07/ip-blacklist.conf`** | IP 黑名单（可选；非空时随 `/igoadmin` ZIP 备份） |
| **`secret07/user-access.log`** | 访客访问日志（JSON 行；ZIP 导出为 **`user-access-bundle.json` v3**，含月 summary、未压实日归档、statsManifest、insightsCache） |
| **`secret-tempfile/`** | 文本暂存、文章导出缓存 **`doc-export.zip`** |
| **`secret-log/`** | 运行日志（`igozhang.log`；不进文章 ZIP） |

**迁移提示**：若旧环境 MP3 在 **`DATA_ROOT/mp3`**、引语/词库在 **`secret07/`**，请手动移至上表 **`assets/`** 对应路径；文章备份统一走后台 **`/igoadmin`** 导出 ZIP → 导入 ZIP（含 `meta/*.json`、`md/`、`doc-pic/`；有定制时另含 `secret07/` 配置与图片；**黑名单**、**`user-access-bundle.json` v3** 亦在 ZIP 内）。导入为**合并**：不删除 ZIP 外的本地文章；`md` 与 ZIP 内容一致则跳过覆盖；黑名单导入后**立即生效**；访客 bundle 写回 log/归档、恢复看板 cache、清空内存看板；`manifest.json` 含 **`accessBundleStats`**；统计偏差时响应 **`warnings`**。

若集群禁止 `hostPath` 或节点不共享磁盘，请改为 **PVC** 并同步修改 `deployment.yaml` 中的 `volumes`。

请提前在目标节点创建目录并确认权限；容器进程需能对该目录读写（镜像默认 root）。

## 3. 构建与推送镜像

在项目根目录（含 `Dockerfile`）执行。Dockerfile 使用 BuildKit cache mount（`GOCACHE` / module cache），**请开启 BuildKit**（Docker 23+ 默认开启；旧版：`export DOCKER_BUILDKIT=1`）。二次构建在依赖未变时通常明显快于全量编译。

```bash
TAG=$(date +%Y%m%d-%H%M%S)
IMAGE=krccr.ccs.tencentyun.com/igokorea/igozhang:${TAG}

DOCKER_BUILDKIT=1 docker build -t "${IMAGE}" .
docker push "${IMAGE}"
```
示例（Docker）：
```bash
TAG=$(date +%Y%m%d-%H%M%S) && DOCKER_BUILDKIT=1 docker build -t krccr.ccs.tencentyun.com/igokorea/igozhang:${TAG} . && docker push krccr.ccs.tencentyun.com/igokorea/igozhang:${TAG}
```

示例（containerd / `nerdctl`；低配机建议限制 BuildKit 并行度）：
```bash
TAG=$(date +%Y%m%d-%H%M%S) && BUILDKIT_MAX_PARALLELISM=1 nerdctl build -t krccr.ccs.tencentyun.com/igokorea/igozhang:${TAG} . && nerdctl push krccr.ccs.tencentyun.com/igokorea/igozhang:${TAG}
```

标签格式：**年月日-时分秒**（示例：`20260509-143052`）。

将 `deploy/k8s/deployment.yaml` 中 `spec.template.spec.containers[0].image` 改为本次构建的 **`${IMAGE}`**，勿长期使用悬空 `:latest` 作为线上唯一版本（可减少回滚歧义）。

若 **`igozhang` Deployment 已在集群中**，只想换新镜像 tag，不必每次改 YAML，可直接：

```bash
kubectl -n igo set image deployment/igozhang igozhang=krccr.ccs.tencentyun.com/igokorea/igozhang:20260509-115923
```

说明：**`deployment/igozhang`** 为 Deployment 名称；**`igozhang=`** 后为容器名（须与 `deployment.yaml` 里 `containers[].name` 一致，本项目为 **`igozhang`**）；镜像地址换成你刚 **`docker push`** 的 tag。

随后执行 **`kubectl -n igo rollout status deployment/igozhang`** 确认滚动完成。

## 4. Kubernetes 部署

镜像在 **`krccr.ccs.tencentyun.com`** 等需登录的私有仓库时，节点拉镜像须带凭证，否则 Pod 会出现 **`ErrImagePull` / `401 Unauthorized`**。在已完成 **`docker login krccr.ccs.tencentyun.com`** 的宿主机上，用其 **`~/.docker/config.json`**（root 用户一般为 **`/root/.docker/config.json`**）在 **`igo`** 创建 pull Secret，并挂到该命名空间 **`default` ServiceAccount**，使未显式写 `imagePullSecrets` 的 Pod（含本 Deployment）默认带上凭证：

```bash
kubectl create secret generic tcr-pull \
  --from-file=.dockerconfigjson=/root/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson \
  -n igo
kubectl patch serviceaccount default -n igo --type merge -p \
  '{"imagePullSecrets":[{"name":"tcr-pull"}]}'
```

说明：Secret 已存在时需先 **`kubectl delete secret tcr-pull -n igo`** 再创建。若 `default` SA 里已有其它 `imagePullSecrets`，`merge` 会整段替换该字段，需 **`kubectl edit serviceaccount default -n igo`** 手动合并。凭证更新后若 Pod 仍卡在 **`ImagePullBackOff`**，执行 **`kubectl -n igo rollout restart deployment/igozhang`**。更多说明见 **[trans/deploy/deploy.md](../../trans/deploy/deploy.md)** §「Kubernetes 拉取公有云私有镜像」。

### 4.1 TLS 证书 Secret（宿主机 `/data/igozhang.cn/tls/`）

Ingress 使用 **`igo`** 命名空间内 Secret **`igozhang-cn-tls`**（与 `deploy/k8s/ingress.yaml` 中 `secretName` 一致）。证书文件位于宿主机数据目录（与 **`volumes.hostPath`** 一致）：

- **`/data/igozhang.cn/tls/igozhang.cn.pem`**
- **`/data/igozhang.cn/tls/igozhang.cn.key`**

集群中**尚无**该 Secret 时，在已放置上述文件的宿主机执行（会覆盖同名 Secret）：

```bash
kubectl create secret tls igozhang-cn-tls \
  --cert=/data/igozhang.cn/tls/igozhang.cn.pem \
  --key=/data/igozhang.cn/tls/igozhang.cn.key \
  -n igo
```

续期或更换证书后，删除并重建 Secret，再滚动 Ingress / 等待控制器热加载：

```bash
kubectl delete secret igozhang-cn-tls -n igo
# 再执行上面的 kubectl create secret tls ...
kubectl -n igo rollout restart deployment/igozhang   # 可选；一般改 Secret 后 Ingress 会自动拾取
```

Ingress 已设置 **`ssl-redirect: false`** 与 **`force-ssl-redirect: false`**：**HTTP 与 HTTPS 均可访问**，控制器不会把明文请求强制 301 到 HTTPS。

### 4.2 应用清单

```bash
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
kubectl apply -f deploy/k8s/ingress.yaml
```

检查：

```bash
kubectl -n igo rollout status deployment/igozhang
kubectl -n igo get svc,ingress,pods -l app=igozhang
kubectl -n igo get secret igozhang-cn-tls
```

- **Service**：集群内 **80** → Pod **3003**（端口名 `http`）。
- **Ingress**：`host: igozhang.cn`，路径 `/`，TLS **`igozhang-cn-tls`**，后端 Service `igozhang:80`。
- **压缩**：进程内对 HTML/CSS/JS/JSON 等文本响应做 **gzip**（不依赖 Ingress ConfigMap）；`/stream/mp3/`、`/files/` 不压缩。上线后可用 `curl -H "Accept-Encoding: gzip" -D- -o NUL https://igozhang.cn/` 确认 `Content-Encoding: gzip`。

## 5. 访问方式与环境变量

### 5.1 站点根路径（`https://igozhang.cn/`）

- **不要**设置 `BASE_PATH`（`deployment.yaml` 中保持默认即可）。
- 前端与 API 均以域名根为基准（`/api/...`、`/files/...` 等）。
- 对外入口可为：**Ingress Controller 的 HTTP/HTTPS NodePort**，和/或宿主机 **宝塔 Nginx** 反代（见 §6）。

### 5.2 运维参数 `igo.conf`

- 路径：**`DATA_ROOT/secret07/igo.conf`**（仓库示例 **`deploy/secret07/igo.conf`**，由 **`cmd/igozhang/igo.conf`** 经 **`go generate ./cmd/igozhang/...`** 同步，勿单独改 deploy 副本）。
- 全站展示名（页眉与首页 `<title>`）：**`[site] title`**。
- 进程**仅在启动时**读取；修改文件后须**重启 Pod** 才生效，运行中不会监视文件变更。
- 改配置后滚动重启示例：

```bash
kubectl -n igo rollout restart deployment/igozhang
kubectl -n igo rollout status deployment/igozhang
```

- 已写入 PVC/宿主机上的旧版若仍含 **`reload_interval`**，会被忽略，可删除该行。

### 5.3 统一认证、运维页 `/ops`、看板与 SSH（`ssh.pass`）

- **统一认证**：`/ops`、`/igoadmin`、`/fs` 及 `link.conf` **`[secret-auth]`** 文章共用「IGO密钥认证」（须 **`[auth] enabled=on`** 且 **`[igokey] enabled=on`**）。
  - 取钥：访问 **`/igokey`** 取得 8 位字母 key（TTL 约 1h）。
  - 登录：`POST /api/auth/login` → 内存会话 token（**刷新/重开须重新验证**，不保留 Cookie）。
  - 探活：`POST /api/auth/session`；业务 API 带 **`X-Igo-Token`** 或 body **`token`**。
- **运维页入口**：直接访问 **`/ops`**（一级路径；无页脚按钮；仅须 **`[ops] enabled=on`**；关 ops **不影响** fs / igoadmin）。
- **流程**：输入有效 key → 验证通过 → 页面另请求 **`POST /api/ops/targets`** 填充目标机 → 执行 Shell；Shell 下方为**看板**（在线/近期访客、近一日/周/月访问量、黑名单维护）。
- 访客日志：**`DATA_ROOT/secret07/user-access.log`**（JSON 行；近 30 天日归档全量，月压缩统计 + 200 条，保留 1 年；单文件 ≤5MB）。
- 黑名单：**`DATA_ROOT/secret07/ip-blacklist.conf`**（精确 IPv4/IPv6）；命中全站 **403**；非空时随 `/igoadmin` ZIP 导出，导入后立即重载。
- Pod 内 Go 使用 **`golang.org/x/crypto/ssh`** 连接目标 Linux；**复用 SSH 连接**，空闲 **5 分钟**无命令自动断开。
- 首次启动若无 **`ssh.pass`**，会自动生成仅含一行说明注释的示例文件，须改为真实记录后重启 Pod。
- 凭证文件：**`DATA_ROOT/secret07/ssh.pass`**（可参考 **`deploy/secret07/ssh.pass.example`**），每行：

  ```text
  主机 端口 用户名 密码
  ```

  密码可含空格（写在第四段起）。**最多 10 行**，展开后 **最多 10 台**；IPv4 末段可用 **`10.1.1.2-5`** 表示 `.2`～`.5`。**勿提交 Git**；权限建议 `600`。

- 「执行」旁显示 **北京时间到期与剩余分钟数**（按浏览器时钟 + 服务端下发的剩余秒数计算，避免宿主机时区/时钟偏差）。

- 网络：确保 **Pod 出网** 能访问目标机的 SSH 端口（安全组 / 防火墙放行）。
- 关闭运维页： **`[ops] enabled = off`** 后重启 Pod（`/ops` 返回 404；**不**影响 `/fs`、`/igoadmin` 认证）。
- 关闭统一认证： **`[auth] enabled = off`**（或关 **`[igokey]`**）后重启 Pod（`/api/auth/*` 与依赖 token 的接口 404）。
## 6. 宿主机 Nginx（宝塔：`igozhang.cn` 反代，可选）

目标：浏览器访问 **`http://igozhang.cn/`** 或 **`https://igozhang.cn/`**，由本机 Nginx 反代到集群 **Ingress**（`Host: igozhang.cn`）。仓库参考：**`deploy/nginx/igozhang.cn.conf`**。

| 组件 | 配置要点 |
|------|----------|
| **Ingress** | `host: igozhang.cn`，TLS Secret **`igozhang-cn-tls`**；不强制 HTTP→HTTPS |
| **Deployment** | 不设 **`BASE_PATH`** |
| **宝塔 Nginx** | `proxy_pass http://127.0.0.1:节点口/`（常见 **30080**）；**`proxy_set_header Host igozhang.cn`** |
| **80 端口** | **不**做全站 301 到 HTTPS（与 Ingress 一致，明文可访问） |

**说明**：`listen` 行勿写 `http2` 参数（会触发 `the "listen ... http2" directive is deprecated`）；应使用独立指令 **`http2 on;`**（Nginx 1.25.1+）。

安装示例（按需将 **`30080`** 换成本机 Ingress HTTP NodePort）：

```bash
sudo install -m 644 deploy/nginx/igozhang.cn.conf /www/server/panel/vhost/nginx/igozhang.cn.conf
sudo nginx -t && sudo nginx -s reload
```

请确认 DNS **`igozhang.cn`** 已解析到本机，且宿主机 **`/data/igozhang.cn/tls/`** 下存在 **`igozhang.cn.pem`**、**`igozhang.cn.key`**（与 §4.1 同步到集群 Secret 的为同一套文件）。

若仅需 **Ingress 直接对外暴露** NodePort、不经宝塔，可跳过本节；此时 HTTPS 由 Ingress Controller 使用 **`igozhang-cn-tls`** 终结。

## 7. 上线检查清单

**完整约束与易漏项**（含 Media Session / Android 真机项）见项目根 **[checklist.md](../checklist.md)**；本节为部署侧操作清单，发版建议两文件对照执行。

- [ ] `deployment.yaml` 中 **image** 已改为本次 **TAG**。
- [ ] 节点上宿主机 **`/data/igozhang.cn`**（或 PVC）存在且可读写；其下 **`tls/`** 证书文件齐全。
- [ ] **未设置** `BASE_PATH`（根路径部署）。
- [ ] **`igo` 命名空间** 存在 Secret **`igozhang-cn-tls`**，且 SAN 含 **`igozhang.cn`**。
- [ ] Ingress **`ssl-redirect` / `force-ssl-redirect`** 为 **false**；**HTTP** 与 **HTTPS** 均可访问。
- [ ] **媒体能力**：MP3 封面与 **Media Session** 在 **HTTPS** 下更可靠；生产建议提供 HTTPS 入口，但不强制关闭 HTTP。
- [ ] 若使用宝塔：**`igozhang.cn.conf`** 已写入，**80 未强制跳转 443**，`nginx -t` 通过。
- [ ] Ingress **host**（**`igozhang.cn`**）与 DNS / 证书一致；**`proxy_set_header Host`** 与之一致。
- [ ] Pod **`readinessProbe`** 访问 **`/health`** 成功（3003 端口）。

## 8. 常见问题

**Pod 无法拉起或探针失败**

- 确认容器监听 **`LISTEN=:3003`**，且 Service `targetPort` 指向端口名 **`http`**。
- `kubectl -n igo logs deploy/igozhang` 查看启动日志。

**中转保存失败或权限错误**

- 检查 **`hostPath`** 目录权限与节点是否与调度 Pod 一致。
- 确认磁盘剩余空间。

**Ingress HTTPS 不可用或证书错误**

- `kubectl -n igo describe ingress igozhang` 查看 Events。
- 确认 Secret **`igozhang-cn-tls`** 存在且 `tls.crt` / `tls.key` 与宿主机 **`/data/igozhang.cn/tls/`** 下证书一致。
- 证书续期后须按 §4.1 重建 Secret。

**页面能打开但目录列表显示「请求失败」**

- 确认未误设 **`BASE_PATH`**；API 应为 **`https://igozhang.cn/api/browse`**（或当前访问域名的 `/api/...`）。
- 开发者工具 → Network 检查 **`browse`** 请求 URL 与状态码。

**请求 URL 出现 `%22`（引号）或异常路径前缀**

- 多为旧版页面模板注入 **`BasePath`** 错误；请使用当前镜像（`BasePathJS` 注入）。仍异常时核对是否混用了旧 Nginx **`X-Forwarded-Prefix`** 配置（根路径部署不需要该头）。
