# igozhang

基于 Go 的单二进制站点：在挂载的数据目录上提供**文章站**（Markdown + SQLite FTS）、**目录浏览与下载**、**HTTP(S) 拉取与浏览器上传到中转目录**、**文本暂存**、**MP3 流媒体与封面元数据**，以及五子棋/象棋/德扑/知键等子页。默认监听 **`:3003`**，无额外运行时依赖；前端页面与脚本嵌入二进制，便于容器镜像部署。

**代码规模（约，不含测试）**：`cmd/igozhang/` 主程序 **~1.6 万行 Go**、**~1.1 万行嵌入 JS**、**~4.3k 行 HTML**（`//go:embed` 页面、MP3、棋类等）；`internal/`（文档索引、Markdown、IP 地理）**~2.4k 行 Go**。无 Node 构建链；镜像为 **CGO=0 静态二进制**（不含 `assets/`，内容走 hostPath）。Go **1.22**；CI：`go vet` + `go test`。

站点标题默认为 **igozhang**（`igo.conf` `[site] title` 可改）。静态资源与 API 使用**相对路径**；生产部署为站点根路径（如 **`https://igozhang.cn/`**），一般**无需**设置 **`BASE_PATH`**。

### 文档三件套（变更时请同步维护）

| 文档 | 说明 |
|------|------|
| **[checklist.md](checklist.md)** | 不变量、易回归点（含 Media Session / Android）、发版与真机自检清单 |
| **[deploy/deploy.md](deploy/deploy.md)** | 镜像、Kubernetes、Ingress TLS、宝塔 Nginx、上线检查 |
| **[README.en.md](README.en.md)** | 英文概要（能力与构建；细则以本文件为准） |

凡影响**功能说明、业务限制、部署方式或前端契约**的改动，请更新 **`README.md`** 中对应章节，并视情况更新 **`checklist.md`** 与 **`deploy/deploy.md`**，避免文档漂移与同一逻辑反复修复。

---

## 功能模块概览

### 1. 文章站（`DATA_ROOT/assets`）

- **数据路径**：`assets/md/{slug}.md` 正文、`assets/doc-pic/` 图片、`assets/mp3/` 音频、`assets/eng/phrases.md` 页眉引语、`assets/eng/word3k-full.md` 知键词库；SQLite 索引在 `secret07/sqlitedb/doc.db`（元数据 + FTS 搜索 + 浏览量，含 `-wal`/`-shm` 运行时文件）。**备份/迁移**走后台 **导出 ZIP → 导入 ZIP**（**合并导入**：保留 ZIP 未包含的本地文章；`md` 与 ZIP 字节一致则不覆盖；ZIP 外且无索引的 `md` 按导入时刻补登记，浏览量 1–5 随机）（含 `meta/*.json`、`md/`、`doc-pic/`；有定制时另含 `secret07/` 下 `igo.conf`、`link.conf`、`ssh.pass`、`common-commands.md`、`banner.*`、`favicon.*`、`texas.*`；**黑名单非空**时含 `ip-blacklist.conf`；**白名单非空**时含 `ip-whitelist.conf`（导入后立即重载）；**存在访客数据**时含 **`secret07/user-access-bundle.json`（v3）**：导出前自动**月压实**；**近 1 年**月 summary 全量；**尚无月 summary 的日归档**全量（不受 30 天窗口限制）；明细**瘦身**（保留全部 `session`、每会话 path 最多 **12** 条）；含 **`statsManifest`**（day/week/month 预计算）与 **`insightsCache`**；`manifest.json` 另含 **`accessBundleStats`** 摘要）；文章变更后约 **3 秒**防抖预生成至 `secret-tempfile/doc-export.zip` 并**预热进内存**（生成完成前仍可用上一版；单实例 ≤28MB），导出优先 **`hit-mem`** 直出，否则 **`hit-disk`** 流式；下载文件名时间为**北京时间**（`igo.conf` `[site] beijing_offset_hours` 修正）。**导入**写回 log/归档、**立即恢复**看板 cache、**清空内存看板**；统计与 manifest 不一致时在响应 **`warnings`** 中提示（仍视为导入成功）。仍兼容导入 **v2** bundle。
- **首页 `/`**：分类导航 + 搜索（slug / 标题 / 标签）；列表按分类**分页**，每页 **`[doc] home_latest`** 条（默认 5，`?category=` / `?page=`）。**MP3 在 `main` 与页脚之间**（紧贴页脚上方）。样式走外置 **`/assets/doc/home.css?v=…`**（immutable）；MP3 逻辑与子页共用外置 **`/assets/doc/mp3-init.js?v=…`**。页眉引语 JSON 仅注入一次。应用对 HTML/CSS/JS/JSON 等文本响应做 **gzip**（客户端声明 `Accept-Encoding: gzip` 时）。**圆形导航球**（`[doc] nav_ball`，默认 on）：文档页桌面侧栏 / 其余页固定浮球（上/下半圆翻页、中心回首页）；`off` 关闭。
- **文章页** `GET /{slug}`：Goldmark 渲染；图片 URL 为 `/assets/doc-pic/…`。页顶与首页共用**分类 + 搜索**模块。标题自动生成 **Unicode 友好锚点**（中文等保留，便于 `[文字](#中文标题)` 页内跳转）。支持 fenced **`mermaid`** 代码块（flowchart 等），正文含 Mermaid 时按需加载嵌入的 Mermaid.js 渲染为 SVG；后台 `/igoadmin` 预览同样按需加载（首屏不预挂大库）。
- **密钥保护文章**：`secret07/link.conf` 的 **`[secret-auth]`** 段列出文章 slug（一行一个，最多 50）；访问对应 `/{slug}` 时先出「IGO密钥认证」面板，验证通过后经 **`GET /api/doc/secret-article?slug=`**（须 **`X-Igo-Token`**）加载正文；首页摘要不展示这些文章的 excerpt。须 **`[auth]`** 与 **`[igokey]`** 均开启；关闭认证时按普通文章公开。改 `link.conf`：**`[secret-auth]`** 每次请求读盘即生效；侧栏 **`[external]`/`[internal]`** 胶囊服务端约 **60s** 缓存；文章页 HTML 带 **`ETag`（含 link.conf mtime）** 与 **`private, max-age=60`**（与胶囊 TTL 对齐；首页为 **`no-store`**），改胶囊后文章页最多约 **60s** 内可能仍见旧侧栏，超时再访或强刷后会刷新。
- **模糊路径**：`GET /{前缀}` 若无精确 slug 且长度 ≥3，**302** 跳转到以该前缀开头的最短 slug（如 `/squid` → `/squid_ins`）。
- **搜索 API**：`GET /api/doc/search?q=`。
- **后台** `GET /igoadmin`：与 `/ops`、`/fs` 共用统一认证（`POST /api/auth/login` → `X-Igo-Token`）；新建 / 编辑 / **DELETE** 文章；**分类**列表与重命名/删除；**批量**改分类/标签（可 dry-run）；Markdown 内光标处粘贴图片（≤200KB，png/jpeg/jpg，**保存前仅浏览器缓存、保存时落盘**）；上传 `.md` / ZIP 导入导出；删除文章时清理未被其他文章引用的 `doc-pic` 图片。
- **首次部署**：将仓库 `assets/` 同步到 `DATA_ROOT/assets/`；本地开发可设 **`DOC_ASSETS_SEED=assets`**。若 `md/` 为空且无种子，自动内置演示文章 **`docker_ts`**。

### 2. 目录浏览与下载

- Web 表格列出 `DATA_ROOT` 下指定子路径内的文件与文件夹，支持面包屑进入子目录（控制台 **`/fs`** 页内操作须先通过「IGO密钥认证」；见 §8）。
- 文件通过 **`GET /files/…`** 下载（附件）；常见文本类型（`.txt` / `.md` / `.js` / `.yaml` / `.toml` / `.conf` / `.json` / `.go` 等）可通过 **`GET /preview/…`** 在新标签页预览（Markdown 渲染；其余走文章同款代码壳；含空字节的二进制会拒绝；体积上限 **`[preview] max_size`**，默认 **4 MiB**）。**直链下载/预览不要求登录**（可单独分享 URL）。
- **删除文件**：控制台列表可删可浏览目录下的**普通文件**（不可删目录）；**`DELETE /api/file?path=`** 须会话 token。
- **目录白名单**：浏览与 **`/files`/`/preview`** 仅允许路径中每一级目录名以字母 **`f`** 开头（不区分大小写，如 **`fs`**）；**`assets`**、**`secret-*`** 等不以 **`f`** 开头的目录不出现在列表中。MP3 列表走专用 **`GET /api/browse?path=assets/mp3`**（**不需 igokey token**，但须 **MP3 同源门禁 Cookie**，见 §5）；播放走 **`/stream/mp3/…`**（磁盘 **`assets/mp3/`**）；文章图片走 **`/assets/doc-pic/…`**。
- **上传权限**：经 **`/api/upload`**、**`/api/fetch`** 及暂存导出写入中转目录的文件，保存为 **`0600`**（这些写接口须有效会话 token，见 §8）。

### 3. 文件中转（默认子目录 `fs`）

- **控制台**：一级路径 **`GET /fs`**（中转 / 暂存 / 浏览）；入口为统一「IGO密钥认证」面板，验证通过后方可操作。
- **链接拉取**：`POST /api/fetch`，请求体 JSON `{"url":"https://…"}`，请求头 **`X-Igo-Token`**（或与登录同源会话）；服务端下载远程文件到 **`DATA_ROOT`/`TRANSFER_SUBDIR`**（默认 **`fs`**）。
- **浏览器上传**：`POST /api/upload`，`multipart/form-data` 字段名 **`file`**（可多文件）；须带 token。
- 成功保存时若同名文件已存在，会自动追加时间戳后缀以避免覆盖。
- **拉取接口**：成功时为 **`application/x-ndjson`** 流式进度（`start` / `progress` / `done`）；上游失败或非流式错误仍为 JSON **`{"error":"…"}`**。
- **体积与超时**：单次拉取/上传写入磁盘上限 **2 GiB**（与请求体总上限一致量级）；拉取请求 JSON 体限制 **1 MiB**；客户端上传请求体总上限 **2 GiB**。

### 4. 文本暂存（`secret-tempfile/stash.json`）

- **`GET/POST/DELETE /api/stash`**：读写或清空服务端暂存文本（须会话 token）；**`GET`** 返回 **`updatedAt`**（Unix 秒），并带 **`ETag`**；客户端可发 **`If-None-Match`**，未变化时服务端返回 **304**（省流量）。**`POST`** 可带 **`ifMatchUpdatedAt`**，与服务器当前版本不一致时返回 **409**（正文含 **`serverUpdatedAt`**、**`serverText`**），供多设备冲突时选择覆盖或加载服务端版本。
- **`POST /api/stash/export`**：将当前暂存写入 **`DATA_ROOT/fs`**（或 `TRANSFER_SUBDIR`）下 **`stash-export-<时间戳>.txt`**（须 token），作为应急备份通道。
- 单份内容不超过 **200 KiB**（按 UTF-8 字节计）；服务端保留约 **12 小时**（超过 TTL 或清理任务会删除文件）。
- 存储路径位于 **`DATA_ROOT/secret-tempfile/`**（不以 **`f`** 开头，不出现在目录浏览列表中）。启动时若存在旧版 **`secret01/stash.json`** 会自动迁入新目录。
- **多端同步（控制台）**：其它已打开的页面在前台用 **`ETag` + `If-None-Match`** 轮询（未变化多为 **304**）；轮询间隔带**退避**（约 **3s** 起，最高约 **30s**）。标签页不可见时暂停定时拉取，切回前台或网络恢复时立即补拉。本机正在保存/清空、或暂存冲突条展开时暂停轮询，避免与 **`ifMatch`** 流程打架。若服务端版本已变而本地编辑相对「上次从服务端应用下来的内容」有改动，**不会覆盖文本框**，并在行内提示；与服务器一致时才会自动写入并 toast「已由服务端更新」。

### 5. MP3

- **同源门禁**（`igo.conf` `[mp3]`，默认 on）：列表 / 流 / 封面须 **HttpOnly Cookie** `igozhang_mp3`（HMAC 签名，TTL 默认 **12h**；HTTPS 或 `X-Forwarded-Proto: https` 时带 `Secure`）。带 MP3 的页面响应会签发；前端亦可 **`GET /api/mp3/gate`** 刷新。有 **Referer/Origin** 时须与本站 Host 同源（空 Referer 放行，兼容部分移动端）。关 **`gate_enabled`** 则不校验。同 IP 并发流默认 **4** 路（含预取），超额 **429**；每分钟不同曲目默认 **40**，超额 **429**。
- **`GET /stream/mp3/…`**：对 `DATA_ROOT/assets/mp3` 内文件以音频流形式播放（便于 `<audio>`）；须门禁 Cookie。
- **`GET /api/mp3/cover?path=`**：返回内嵌 APIC 封面原始图片（无封面时 **404**）；须门禁 Cookie；带 **`ETag`** / **`If-None-Match`**（**304**）减轻重复传输。页面与 **Media Session** 锁屏/通知栏优先用该 URL，避免超长 `data:` 图在移动端后台被回收；无封面时前端会用 **`/favicon.svg`** 作占位图。
- 页面内：**歌名/歌手搜索**（选中后从该曲起**随机连播**）、**随机 / 艺人标签 / 连播**、**`one` 单曲循环**、**锁屏与通知栏** 的播放/上一曲/下一曲（**上一曲**按本页播放历史栈，无历史时等同「当前曲从头播」）；**双 `<audio>` 槽位**：副槽预缓冲、在副槽 `playing` 后再与主槽交换（`dualSlotPlaySecondaryBuffer`），随机/锁屏「下一首」等与曲终连播走同一套逻辑，避免单槽直接改主槽 `src` 导致 Android 通知/锁屏元数据「先空后显」；失败再回退主槽直链。**不再整首 Blob 下载**：主槽以 **`/stream/mp3`** 流式播放；**首次**可适度预缓冲，**切歌/后台**走 fast attach。**锁屏/通知「下一首」**预缓存下一首歌名与封面（无封面时页内用 **favicon.svg**、锁屏/通知用 **favicon.png**），避免媒体卡片闪没；与其它 App 共存时采用 **ambient** 音频会话，被微信/录音等打断后不会抢播。**后台省电连播**为默认（无开关）：页隐藏时主槽连播、Media Session 保活、曲末前约 10s 预取下一首元数据与音频（弱网/省流模式跳过音频预取）、卡顿恢复；流响应 `Cache-Control: private, max-age=86400` 利于回访缓存；首次进后台 toast。验收：Android Chrome HTTPS 后台自动连播 **≥5 首**。逻辑见 **`mp3-init.js`**（外置 `/assets/doc/mp3-init.js`，须与 embed 同源）。**`navigator.vibrate`** 轻触反馈（支持时）等见该脚本。

### 6. 站点 chrome 模块（head / foot / mp3）

- **统一装配**：子页通过 `applySitePage` 注入 `{{SITE_HEAD}}`、`{{SITE_FOOTER}}`、可选 `{{MP3_*}}`；勿在各页重复拼装。
- **路径解析**：`site-pub-base.js` 提供全站唯一的 `window.__fsAbsRoot()`（写入于 `pageHeadTags`），footer / mp3 / fs / 知键 等脚本均委托它解析 `BASE_PATH`、文章 slug、`/igoadmin` 等路径。
- **表现一致**：凡带 footer 的页面，双击「专业化」复制 igokey、骰子、页脚链接等行为相同。

### 7. 前端体验（嵌入页）

- **文件中转**：控制台 **`/fs`** 须先「IGO密钥认证」；拉取/上传区主按钮略小；**拉取成功**后自动清空 URL 输入框；行内成功/失败/取消等提示约在数秒后自动消失（仍保留 toast）。
- **台词**：随机中英文电影台词（`assets/eng/phrases.md` 外部库或内嵌样本）；前台约每 **20 秒**自动换下一句（仅多于一条台词时；**`prefers-reduced-motion: reduce`** 下关闭自动轮换）；标签页不可见时暂停计时；台词行带 **`aria-live="polite"`** 便于读屏感知轮换。
- **文本暂存**：保存/清空/导出等操作后的行内提示会短时自动消失；多端同步见上文「文本暂存」（须已认证）。
- **护眼主题**：浅色护眼配色（见内嵌 CSS）。
- **站点名**：页眉站点链接与首页 **`<title>`** 由 **`igo.conf`** 的 **`[site] title`** 配置（`网站名 [短句]`，空格分隔；仅网站名时短句默认为 **Practice makes perfect**；默认网站名 **`igozhang`**）；修改后须 **重启 Pod** 生效。
- **页眉 banner**：`secret07/banner.{png,jpg,jpeg}`（≤200KB）；蒙版强度 **`[site] banner_mask`**（0–100，默认 **66**；改 conf 须重启）。详见 **`cmd/igozhang/pic.readme.md`**。
- **北京时间显示**：页眉时钟、文件列表时间等基于 **UTC+8**；若显示偏快/偏慢，可在 **`[site] beijing_offset_hours`** 用 **`±` 与 0–12** 修正（如快 1 小时设 **`-1`**，默认 **`0`** 不偏移）；仅影响展示，不改密钥校验用的服务器时钟。
- **页面路径**：面向用户的子功能页使用**一级路径**（如 **`/fs`**（文件中转/暂存/浏览）、**`/gomoku`**、**`/zhijian`**、**`/mermaid`**、**`/igokey`**）；首页仅 **MP3** 模块，其它子页亦挂载 MP3 播放器。
- **页脚**：五子棋（默认 **`/gomoku`**）、象棋（**`/xiangqi`**）、德扑（**`/texas`**）、知键、骰子动画、底部链接行等由 **`DATA_ROOT/secret07/igo.conf`** 配置（首次启动无文件时写入示例）；修改后须 **重启 Pod** 生效。

### 8. 统一密钥认证、运维与看板

- **统一认证（Auth）**：`/ops`、`/igoadmin`、`/fs` 及 `link.conf` **`[secret-auth]`** 保护的文章共用同一套「IGO密钥认证」面板与会话（前端 `__igoAuthInit` / `ui/igo-auth.js`）。
  - 取钥：`GET /igokey`（须 **`[igokey] enabled=on`**）。
  - 登录：`POST /api/auth/login`，body `{key}` → 内存 `token` + 剩余秒数（**不**返回 SSH targets；刷新页须重新验证，不写 Cookie / localStorage）。
  - 探活：`POST /api/auth/session`，body `{token}`。
  - 业务请求：请求头 **`X-Igo-Token`**（igoadmin / fs）或 JSON body **`token`**（ops）；401 时 `reason` 为 `invalid_token` / `key_expired` / `key_revoked`。
  - 开关：须 **`[auth] enabled=on`** 且 **`[igokey] enabled=on`**（`authFeatureEnabled`）；**关 `[ops]` 不影响** fs / igoadmin 登录。
  - 失败限流：同 IP 约 **5 次/分钟**失败后锁定约 **1 分钟**（`secret07/auth-rl.json`，可自旧名 `ops-auth-rl.json` 迁移）。
- **`GET /health`**：返回纯文本 **`ok`**（**不受 `BASE_PATH` 前缀影响**，便于探针）。
- **定时维护**：每小时执行一次（启动时也会立即执行一次），见下文「规则汇总」。
- **运维页 `/ops`**（`igo.conf` **`[ops] enabled=on`**）：页面开关仅看 `[ops]`；登录走上文统一认证；验证后另拉 **`POST /api/ops/targets`** 填充 SSH 目标；可在页内执行远程 Shell（须配置 **`secret07/ssh.pass`**）；Shell 区下方为**看板**。
- **看板（访客画像）**：服务端根据请求自动采集（无需前端信标）；展示**当前在线**（默认 5 分钟内活跃）与**近期访客**（离线最多 5 条）。每条仅显示已获取字段：IP、geo、客户端（如 Chrome）、**本次会话**浏览路径、**会话开始时间（UTC+8 展示）**。看板按 **`[ops] board_poll_interval`** 自动刷新（默认 **5 秒**）。另含**国别来源 Top 8**与 **IP 访问 Top 8**（日/周/月切换，按**会话数**计）；统计为内存增量 + 每小时校正，缓存于 **`user-access-board.cache.json`**，启动异步重建不阻塞监听。**ZIP 导入** bundle 时立即写回 cache 并清空内存看板，避免「在线人数」与「访问次数」错位。
- **访问量统计**：基于 **`secret07/user-access.log`** 及归档，看板顶部显示近一日 / 近一周 / 近一月**访问次数（会话数）**：同一 IP 在 **`session_idle_timeout`**（默认 **30 分钟**）内无新请求视为同一会话，仅计 **1 次**；超时后再次访问计为新会话。会话内换页/功能另写 path 事件但不重复计次。**悬停**数字可预览最近 **30 条会话**（含路径摘要，离开即收起）；**点击**固定展开，点击页面其他区域关闭。
- **IP 黑名单**：运维看板 **`/ops`** 内维护；仅支持**精确 IPv4 / IPv6**（不支持 CIDR）；持久化 **`secret07/ip-blacklist.conf`**（文本，`#` 注释，支持逗号/空格等分隔）；命中全站 **403**；导入 ZIP 后**立即重载**，无需重启 Pod。
- **IP 白名单**：持久化 **`secret07/ip-whitelist.conf`**（格式同黑名单，仅精确 IP）；**列表为空不限制**；**非空**时仅列出的 IP 可访问 **`/ops`、`/igoadmin`、`/fs`、`/api/auth/*`**（否则 **403**）。**不**限制首页、文章、`/files` 直链、MP3 流等。运维页无单独 UI；ZIP 导入后立即重载；直接改文件须 **重启 Pod**。非空时随文章 ZIP 备份。
- **路径探测自动封禁**：同一 IP 在 **1 分钟**内访问 **≥20** 个不同路径（含 404 等不可用路径）时，自动临时封禁 **1 小时**（全站 **403**）；持久化 **`secret07/ip-auto-ban.json`**，**不写入**永久黑名单；运维看板可查看剩余时间并手动解除。默认不计入 `/health`、`/favicon*`、`/api/ops/*`、`/api/auth/*`、`/stream/mp3/`、`/assets/doc-pic/`；阈值见 **`igo.conf` `[security]`**。
- **访客访问日志 `secret07/user-access.log`**：每行一条 JSON（UTF-8），记录 IP、时间、geo、客户端、页面、功能等；**独立轮转**（与 `igozhang.log` 无关）：**近 30 天**日归档全量；满 **30 天**后按月压缩为 **`user-access.YYYY-MM.summary.log`**（统计 + **200** 条样本）；月归档保留 **1 年**；单文件超 **5MB** 截断为最近 **300** 行（允许丢弃更早记录）。

---

## 规则汇总（业务与维护）

以下数值默认以 **`DATA_ROOT/secret07/igo.conf`** 为准（**进程启动时**读取，改配置后须重启 Pod）；仓库示例见 **`deploy/secret07/igo.conf`** 与内嵌 **`cmd/igozhang/igo.conf`**。环境变量 **`TRANSFER_MAX_CONCURRENT_PER_IP`**、**`FS_USAGE_BASELINE_BYTES`** 仍可覆盖对应项。

| 规则 | 说明 | 定义位置（参阅） |
|------|------|------------------|
| 中转目录文件时效 | **`DATA_ROOT/fs`**（或 `TRANSFER_SUBDIR`）内**普通文件**修改时间早于 **7 天**的，在维护任务中**删除**（仅文件，不按目录递归删除空文件夹）。 | `stash.go`：`fsFileMaxAge` |
| 中转目录总容量上限 | 同上目录内**全部文件字节数之和**若超过 **9 GiB**，按**文件大小从大到小**逐个删除，直至合计 **≤ 9 GiB**（同一维护周期内：先做 TTL 删除，再按容量删）。删除日志含 **`reason=`**（`file_ttl_age` / `dir_size_cap`）。 | `stash.go`：`transferDirMaxBytes` |
| 前端用量展示基线 | 控制台用量条以 **`transferDirMaxBytes`（默认 9 GiB）** 为满刻度参考，与清理上限一致；可用 **`FS_USAGE_BASELINE_BYTES`** 单独覆盖。 | `fs_usage.go` + `stash.go`：`transferDirMaxBytes` |
| 文本暂存 TTL | **`stash.json`** 的更新时间超过 **12 小时**则删除整份暂存文件（`reason=stash_ttl`）。 | `stash.go`：`stashTTL` |
| 文本暂存大小 | 单次保存文本不超过 **200 KiB**（UTF-8）。 | `stash.go`：`maxStashBytes` |
| 拉取/上传单文件上限 | 写入磁盘来自拉取或上传的单个文件内容，以 **2 GiB** 为上限（与 HTTP 请求体总上限一致）。 | `main.go`：`MaxBytesReader` / `2<<30` |
| 同一客户端并发中转 | 默认每个 **IP** 同时进行的上传/拉取合计不超过 **2** 路；超额返回 **HTTP 429**。 | `transfer_limit.go`：`TRANSFER_MAX_CONCURRENT_PER_IP`（默认 2，上限 32） |
| 浏览 API 按 IP 频率 | 每个 **IP** 对 **`GET /api/browse`** 每分钟不超过 **150** 次；超额 **429**（JSON **`error`**）。 | `light_rl.go`：`browseRLPerMinute` |
| 暂存 API 按 IP 频率 | 每个 **IP** 对 **`/api/stash`**（任意方法）每分钟不超过 **80** 次；超额 **429**。 | `light_rl.go`：`stashRLPerMinute` |
| 浏览路径安全 | API 路径必须在 `DATA_ROOT` 之下，禁止 `..` 逃逸；**`secret-tempfile`**、**`secret07`**、**`secret-log`** 及其子路径浏览被禁止。 | `main.go`：`browse` |
| 根目录列表过滤 | 一级目录名不以 **`f`** 开头的不出现在根列表（含 **`assets`**、**`secret-*`**、**`mp3`** 旧路径等）；MP3 播放器通过专用 browse 路径 **`assets/mp3`** 列目录（须门禁 Cookie）。 | `browse_rules.go`：浏览逻辑 |
| 中转文件删除 | **`DELETE /api/file?path=`** 仅删可浏览路径下的普通文件（不可删目录）；须 token。 | `browse_rules.go`：`fileDeleteHandler` |
| 文本预览体积 | **`GET /preview/…`** 上限 **`[preview] max_size`**（默认 **4 MiB**）。 | `main.go`：`previewHandler`；`igo.conf` `[preview]` |
| IP 黑名单 | 精确 IPv4/IPv6 写入 **`secret07/ip-blacklist.conf`**（文本）；全站请求在最早中间件拦截，命中返回 **403**。 | `ip_blacklist.go` |
| IP 白名单 | 精确 IPv4/IPv6 写入 **`secret07/ip-whitelist.conf`**；空列表不限制；非空时仅放行 **`/ops`、`/igoadmin`、`/fs`、`/api/auth/*`**，否则 **403**。 | `ip_whitelist.go`、`visitor_board.go`：`isOpsIgoAdminPath` |
| MP3 同源门禁 | Cookie `igozhang_mp3`（默认 TTL **12h**）+ Referer 同源；同 IP 并发流 **4**、每分钟不同曲目 **40**（均可在 `[mp3]` 改）。 | `mp3_gate.go` |
| 路径探测自动封禁 | 滑动 **60s** 窗口内同一 IP 访问 **≥20** 个不同路径（含 404；默认排除 `/health`、`/favicon*`、`/api/ops/*`、`/stream/mp3/`、`/assets/doc-pic/`）→ 临时封禁 **1h**（**403**）；持久化 **`secret07/ip-auto-ban.json`**，到期或维护任务清理；**不写入**永久黑名单。 | `path_probe_ban.go` |
| 访客看板与 user-access.log | 每次可识别动作写原子事件：`session`（新会话，计访问次数）、`visit`/`feature`（会话内路径，不计次）；**30 分钟**无操作后下次请求开新会话（`igo.conf` `[ops] session_idle_timeout`，**5m–6h**）。内存保留每 IP 当前会话最近 **12** 步 `trail`；同步追加 **`secret07/user-access.log`**。看板默认 **5s** 轮询（`board_poll_interval`）。同会话同页 **5 秒**内不重复写 path 日志。 | `visitor_board.go`、`access_log.go`、`access_insights.go` |
| 运行日志 `secret-log` | 应用 **`log`** 输出同时写入 **`stderr`** 与 **`DATA_ROOT/secret-log/igozhang.log`**。HTTP 请求行跳过 **`/health`**（K8s 探针）、**`/favicon*`**；耗时以纯 ASCII（`us`/`ms`/`s`）记录，避免终端乱码。维护（每小时）：归档保留 **`[log] file_max_age`**（默认 **15 天**）；跨日将非空当前文件轮转为 **`igozhang.YYYY-MM-DD.log`**；当前文件超 **`rotate_size`**（默认 **500K**）轮转为 **`igozhang.part.{unix}.log`**；超 **`truncate_size`**（默认 **1M**）则截断为最近 **100** 行并清空归档。 | `logging.go`、`request_log.go` |
| 访客日志 `secret07/user-access.log` | **独立轮转**：近 **30 天**日归档全量；满 **30 天**后月压缩（统计 + **200** 条）；月归档 **1 年**；单文件 **≤5MB**（超出截断最近 **300** 行，可丢旧记录）；仅按日轮转，无 part 分片。 | `access_log.go` |
| 访客日志 ZIP bundle | 导出前**月压实**；bundle **v3** 含近 **1 年**月 summary、未压实日归档、瘦身明细（全 session + 每会话 ≤12 path）、**statsManifest**、**insightsCache**；导入恢复 cache 并清空内存看板；兼容 **v2**。 | `access_log_bundle.go` |

### 目录列表 API 参数

- **`GET /api/browse`** 支持 **`sort=name|mtime|size`**、**`order=asc|desc`**、**`ext=`**（逗号分隔后缀，不含点）。

---

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `LISTEN` | `:3003` | 监听地址 |
| `DATA_ROOT` | `/igozhang.cn` | 容器内数据根目录；其下 **`assets/`**（文章、MP3、引语、词库）、**`fs/`** 中转、**`secret07/`** 配置与 SQLite 等（K8s 经 hostPath 挂载，见 **`deploy/deploy.md`**） |
| `TRANSFER_SUBDIR` | `fs` | 中转目录名（位于 `DATA_ROOT` 下） |
| `BASE_PATH` | （空） | HTTP 子路径前缀，如 `/fs`。独立域名根路径部署勿设置 |
| `DOC_ASSETS_SEED` | `assets` | 首次 **`assets/md/`** 为空时，从该路径（相对工作目录）整棵复制到 `DATA_ROOT/assets/` |
| `TRANSFER_MAX_CONCURRENT_PER_IP` | `2` | 每个 IP 同时进行的上传 + URL 拉取合计槽位数 |
| `FS_USAGE_BASELINE_BYTES` | （默认与 **9 GiB** 上限相同） | `/api/fs/usage` 与前端用量条的展示基线字节数 |

---

## HTTP 接口摘要

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 文章首页（分类 + 分页列表 + 搜索；MP3 在页脚上方） |
| GET | `/health` | 健康检查，正文 `ok`（不受 `BASE_PATH` 影响） |
| GET | `/favicon.svg` | 内嵌魂斗罗风格 favicon（若部署子路径需带前缀） |
| GET | `/api/browse?path=` | 列目录；`path` 相对 `DATA_ROOT`（非 `assets/mp3` 须 **`X-Igo-Token`**；MP3 列表须门禁 Cookie） |
| DELETE | `/api/file?path=` | 删除可浏览目录下的普通文件（须 token；不可删目录） |
| POST | `/api/fetch` | 拉取 URL 至中转目录（须 token；成功为 NDJSON 流） |
| POST | `/api/upload` | 表单上传至中转目录（须 token） |
| GET | `/api/fs/usage` | 中转目录用量（须 token） |
| GET | `/api/stash` | 读取文本暂存（须 token） |
| POST | `/api/stash` | 保存文本暂存（须 token） |
| DELETE | `/api/stash` | 清空暂存（须 token） |
| POST | `/api/stash/export` | 暂存导出到 fs（须 token） |
| GET | `/files/…` | 下载文件（**公开**） |
| GET | `/preview/…` | 预览常见文本文件（md 渲染；js/yaml/toml/conf 等代码壳；**公开**；默认 ≤4 MiB） |
| GET | `/stream/mp3/…` | MP3 流媒体（须门禁 Cookie） |
| GET | `/api/mp3/cover?path=` | MP3 内嵌封面原图（须门禁 Cookie；无则 404；服务端按文件 mtime 缓存 ID3 解析） |
| GET | `/api/mp3/gate` | 签发/刷新 MP3 门禁 Cookie |
| GET | `/favicon.png` | 锁屏/通知栏占位封面（无内嵌 APIC 时前端使用） |
| GET | `/banner.png` 等 | 页眉背景（`secret07/banner.*`，≤200K） |
| GET | `/api/footer` | 页脚配置 JSON（`gomoku`/`xiangqi`/`texas` 链接、`links` 数组，来自 `igo.conf` `[footer]`） |
| GET | `/api/ops-config` | 前端可调运维参数子集（暂存、MP3、知键、骰子等） |
| GET | `/fs` | 文件中转 / 暂存 / 浏览控制台（统一「IGO密钥认证」；直链 `/files` 不需登录） |
| GET | `/igoadmin` | 文章后台（统一认证；导出/导入 ZIP、新建编辑删除、分类批量） |
| GET | `/api/igoadmin/articles` | 文章列表（须 token；可选 `q`/`category`/`limit`/`offset`） |
| GET/POST/PUT/DELETE | `/api/igoadmin/article` | 读/新建/更新/删除单篇（须 token） |
| PATCH | `/api/igoadmin/articles/batch` | 批量改分类/标签（须 token；可 `dryRun`） |
| GET | `/api/igoadmin/categories` | 分类列表（须 token） |
| PUT/DELETE | `/api/igoadmin/category` | 重命名 / 删除分类（须 token） |
| POST | `/api/igoadmin/preview` | 后台 Markdown 预览（须 token） |
| POST | `/api/igoadmin/upload-md` | 上传 `.md`（须 token） |
| GET | `/api/igoadmin/export` | 导出文章 ZIP（须 token） |
| POST | `/api/igoadmin/import` | 导入文章 ZIP（须 token） |
| GET | `/api/doc/search?q=` | 文章搜索 |
| GET | `/api/doc/home` | 首页文章列表 JSON（`category`/`page`/`limit`） |
| GET | `/api/doc/secret-article?slug=` | 读取 `[secret-auth]` 保护文章正文（须 **`X-Igo-Token`**） |
| GET | `/api/zhijian/word3k` | 知键词库抽样（`assets/eng/word3k-full.md` 或内嵌） |
| GET | `/gomoku` | 五子棋人机对战页（内嵌） |
| GET | `/xiangqi` | 象棋人机对战页（内嵌） |
| GET | `/texas` | 德州扑克人机对战页（内嵌，1v3～1v5） |
| GET | `/texas-card-back.png` 等 | 德扑牌背图（`secret07/texas.*`，≤150K；无则前端 Canvas 绘制） |
| GET | `/zhijian` | 知键打字练习页 |
| GET | `/mermaid`、`/Mermaid` | Mermaid 源码编辑与即时预览 |
| GET | `/igokey` | 签发 8 位字母验证密钥（**`igo.conf` `[igokey] enabled=on`**，默认开启；**`off`** 时返回 **404**）；剩余 **>15 分钟** 时刷新复用，`?renew=1` 强制重新生成 |
| GET | `/ops` | 运维页：统一认证后远程 Shell + 看板（**`[ops] enabled=on`**；认证另须 **`[auth]`+`[igokey]`**；Shell 须 **`ssh.pass`**） |
| POST | `/api/auth/login` | 提交 `{key}` 验证 igokey，返回内存会话 `token`（刷新页须重新验证；ops / igoadmin / fs 共用） |
| POST | `/api/auth/session` | `{token}` 续期检查 |
| POST | `/api/ops/targets` | `{token}` 返回 SSH 目标列表 |
| POST | `/api/ops/exec` | 已认证下 `{token,command,target}`，经 SSH 执行（超时 60s，输出上限约 256KB） |
| POST | `/api/ops/board` | `{token}` 返回在线/近期访客、黑名单、访问量统计、日志清单 |
| POST | `/api/ops/stats-detail` | `{token,period:"day"|"week"|"month"}` 返回对应时段最近 30 条访问明细 |
| POST | `/api/ops/blacklist` | `{token,action:"add"|"remove",ip}` 维护黑名单 |
| POST | `/api/ops/auto-ban` | `{token,action:"remove",ip}` 提前解除路径探测自动封禁 |

错误响应多为 JSON：`{"error":"…"}`。

**`DATA_ROOT` 数据目录**（摘要）：

| 路径 | 说明 |
|------|------|
| **`assets/md/`** | 文章 Markdown 正文 |
| **`assets/doc-pic/`** | 文章图片 |
| **`assets/mp3/`** | MP3 音频（体积大，独立备份；不进 ZIP） |
| **`assets/eng/phrases.md`** | 页眉引语外部库（≥150 条有效行启用，否则内嵌） |
| **`assets/eng/word3k-full.md`** | 知键词库外部库（≥50 条有效行启用，否则内嵌） |
| **`secret07/igo.conf`** | 站点运维配置（见下） |
| **`secret07/sqlitedb/doc.db`** | 文章 SQLite 索引（ZIP 导入时自动重建） |
| **`secret07/ssh.pass`** | 运维 SSH 凭证 |
| **`secret07/common-commands.md`** | 运维页常用命令（有生效行时随文章 ZIP 备份/恢复） |
| **`secret07/ip-blacklist.conf`** | IP 黑名单（非空时随文章 ZIP 备份/恢复；导入后立即生效） |
| **`secret07/ip-whitelist.conf`** | IP 白名单（非空时限制 `/ops` `/igoadmin` `/fs` `/api/auth`；随 ZIP 备份；导入后立即生效） |
| **`secret07/auth-rl.json`** | 统一认证失败限流状态（可由旧 **`ops-auth-rl.json`** 自动迁移） |
| **`secret07/ip-auto-ban.json`** | 路径探测自动临时封禁（重启保留；不进文章 ZIP） |
| **`secret07/user-access.log`** | 访客访问日志（当前 + 日/月归档；ZIP 导出为 **`user-access-bundle.json` v3**，导入写回 log/归档、恢复看板 cache、清空内存看板；见 §1 备份说明） |
| **`secret07/favicon.*` / `banner.*` / `texas.*`** | 自定义图标、页眉背景、德扑牌背 |
| **`secret-tempfile/`** | 文本暂存、`doc-export.zip` 导出缓存 |
| **`fs/`** | 文件中转（或 `TRANSFER_SUBDIR`） |
| **`secret-log/`** | 运行日志（`igozhang.log` 及归档；**不进文章 ZIP**） |

**`secret07` 运维配置**（位于 `DATA_ROOT/secret07/`，不出现在根目录列表，且禁止 `/api/browse` 列出其下路径）：

- **`igo.conf`**：INI 风格分段配置（站点名、暂存、中转、限流、日志、文章、MP3、知键、象棋、页脚链接等）。大小用 **`200K`/`9G`**，时长用 **`4s`/`12h`/`30m`** 等带单位写法。**修改后须重启 Pod**，运行中不热加载。**唯一源码**：**`cmd/igozhang/igo.conf`**（`go:embed` 内置）；改完后执行 **`go generate ./cmd/igozhang/...`** 同步到 **`deploy/secret07/igo.conf`**（测试会校验二者一致）。
- **`[site]`**：`title`（`网站名 [短句]`，空格分隔；首页 `<title>` 用网站名）；`beijing_offset_hours`；`banner_mask`（页眉蒙版 0–100，默认 66）。
- **`[igokey]`**：`enabled`（`on` / `off`，默认 **`on`**；`off` 时 **`/igokey`** 不可访问，且统一认证不可用）。
- **`[auth]`**：`enabled`（`on` / `off`，默认 **`on`**）；与 **`[igokey]`** 同时为 on 时提供 **`/api/auth/login|session`** 与 token 守卫（**`/ops`、`/igoadmin`、`/fs` 共用**）。
- **`[ops]`**：`enabled`（`on` / `off`，默认 **`on`**）；仅控制运维页 **`/ops`** 与 **`/api/ops/*`** 业务；登录仍走 **`[auth]`**；Shell 须 **`secret07/ssh.pass`**；验证后展示 Shell 与**看板**。
- **`ssh.pass`**：每行 `主机 端口 用户名 密码`（最多 10 行/10 台；IPv4 可用 `10.1.1.2-5` 展开）；SSH 连接复用、空闲 5 分钟断开；运维会话绑定统一 Auth 的 **key** 过期时间（页内显示剩余分钟），刷新 `/ops` 须重新验证。
- **`ip-blacklist.conf`**：精确 IPv4/IPv6 列表（文本，`#` 注释）；运维看板或 ZIP 导入维护；命中全站 **403**。
- **`ip-whitelist.conf`**：精确 IPv4/IPv6；空不限制；非空时仅放行 `/ops`、`/igoadmin`、`/fs`、`/api/auth/*`。ZIP 导入后立即重载；直接改文件须重启 Pod。
- **`user-access.log`**：访客 JSON 行日志；独立轮转（见上表）。
- **Favicon（文件）**：在 `secret07/` 放置 **`favicon.png`** / **`favicon.jpg`** / **`favicon.jpeg`** 或 **`fav.*`**（同上后缀，文件名大小写不敏感）。单文件 **≤ 100K**，超出则**忽略**并回退默认图标。全站 `<link rel="icon">` 与 MP3 锁屏占位图会改用该文件（经 **`/favicon.png`** 等路径提供）；**无可用文件时**仍用内嵌 SVG + 生成的 **`/favicon.png`**。多文件同时存在时：**`favicon` 优先于 `fav`**，同 stem 下 **`png` > `jpeg` > `jpg`**。替换文件后刷新页面即可，无需重启 Pod。
- **`[footer]`**：`gomoku`、`xiangqi`、`texas`、`max_links`、`link = 文字 | URL`（可重复多行）。
- **德扑牌背（文件）**：在 `secret07/` 放置 **`texas.png`** / **`texas.jpg`** / **`texas.jpeg`**（单文件 **≤ 150K**）。经 **`/texas-card-back.*`** 提供；无可用文件时页面用 Canvas 花纹牌背。
- 五子棋页面由内嵌 **`gomoku.html`** 提供（随二进制更新）；若曾放置 **`secret07/gomoku.html`** 旧版覆盖，请删除以免沿用过期逻辑。

拉取流式进度建议关闭缓冲或转发 **`X-Accel-Buffering: no`**（应用已对 NDJSON 响应设置该头）。

---

## 本地构建与运行

```bash
go build -o igozhang ./cmd/igozhang
```

重复编译时 Go 会使用本机 **`GOCACHE`**（默认 `~/go/cache` / `%LocalAppData%\go-build`），无需额外配置；清理缓存后首次会变慢：`go clean -cache`。

Windows 示例：

```bat
set DATA_ROOT=E:\temp\igozhang.cn
igozhang.exe
```

Linux / macOS：

```bash
DATA_ROOT=/tmp/igozhang.cn ./igozhang
```

浏览器访问 `http://127.0.0.1:3003/`。

本地自检（与 CI 一致，见 **`.github/workflows/ci.yml`**）：

```bash
go vet ./...
go test ./...
```

改 **`cmd/igozhang/igo.conf`** 后须同步部署副本：

```bash
go generate ./cmd/igozhang/...
```

---

## 容器与部署

Dockerfile 使用 BuildKit 缓存挂载，请开启 BuildKit：

```bash
DOCKER_BUILDKIT=1 docker build -t krccr.ccs.tencentyun.com/igokorea/igozhang:$(date +%Y%m%d-%H%M%S) .
```

镜像仓库、Kubernetes、Ingress（TLS）、宝塔 Nginx 反代等步骤见 **[deploy/deploy.md](deploy/deploy.md)**；不变量与发版自检见 **[checklist.md](checklist.md)**。镜像**不含** `assets/`：首次请把仓库 `assets/` 同步到宿主机 **`/data/igozhang.cn/assets/`**。

---

## 仓库目录说明

| 路径 | 说明 |
|------|------|
| `assets/` | 站点内容与媒体：`md/`、`doc-pic/`、`mp3/`、`eng/phrases.md`、`eng/word3k-full.md`（**不打进镜像**，运行时在 `DATA_ROOT`） |
| `cmd/igozhang/` | 服务入口、HTTP 路由、内嵌前端与 `igo.conf` |
| `cmd/igozhang/igo.conf` | 运维配置唯一源码（`go:embed`）；`go generate` 同步到 `deploy/secret07/` |
| `cmd/igozhang/pic.readme.md` | favicon / banner / 德扑牌背放置说明 |
| `internal/` | `docindex`（SQLite/FTS）、`markdown`（Goldmark）、`ipgeo`、`docmermaid`（Mermaid 11.4.1） |
| `scripts/` | `sync_igo_conf`、词库/引语/地理库生成脚本 |
| `deploy/secret07/igo.conf` | 部署用配置副本（须与内嵌一致） |
| `checklist.md` | 约束与发版自检（与 README、deploy 同步维护） |
| `deploy/deploy.md` | 部署手册 |
| `deploy/k8s/` | Deployment / Service / Ingress |
| `deploy/nginx/` | 宝塔 Nginx 反代示例（`igozhang.cn.conf`） |
| `.github/workflows/ci.yml` | `go vet` + `go test` |
