# igozhang 约束与变更自检清单

本文与 **`README.md`**、**`deploy/deploy.md`** 配套：**凡影响功能说明、业务限制、部署方式、前端行为契约（含 Media Session）的改动**，请同步更新上述文件中对应段落，避免「代码已改、文档未改」或同一逻辑反复踩坑。

---

## 1. 文档三件套维护约定（必遵守）

| 文件 | 用途 | 变更时何时必须更新 |
|------|------|---------------------|
| **`README.md`** | 项目能力说明、规则对照表、环境变量、接口摘要、本地构建 | 增删 API、改限制常量、改环境变量、用户可见行为 |
| **`deploy/deploy.md`** | 镜像构建、K8s、Ingress TLS、宝塔 Nginx | 改端口、改探针、改证书/反代示例、改上线步骤 |
| **`checklist.md`**（本文件） | 不变量、真机/发版自检、易回归点 | 新增契约类逻辑、修过又易坏的点、部署检查项变更 |

合并请求或发版前建议：在描述中写明「已更新 README / checklist / deploy 中的哪些小节」。

---

## 2. 代码与 README「规则汇总」表一致

- 数值型限制（TTL、体积上限、并发、按 IP 频率等）以 **`igo.conf` + Go 默认值为真源**（启动时加载，**无热加载**）；改配置后须 **重启 Pod**。改代码默认值后须改 **`README.md`** 规则表及（若适用）**`checklist.md`** 中的阈值引用。
- 涉及文件：`stash.go`、`transfer_limit.go`、`light_rl.go`、`logging.go`、`fs_usage.go`、`main.go` 等。

---

## 3. Media Session / Android Chrome（易回归，重点自检）

以下约束对应唯一源码 **`cmd/igozhang/mp3-init.js`**（`site_mp3.go` //go:embed；首页/子页经外置 **`/assets/doc/mp3-init.js?v=…`** 加载，`{{.Mp3ScriptsHTML}}` / `{{MP3_SCRIPTS}}` 仅注入 boot）。**禁止**再在 HTML 内联整份 MP3 脚本；发版跑 **`go test ./cmd/igozhang -run 'TestMp3|TestHome|TestProjectTextFiles'`**。
- [ ] **源码 UTF-8**：`cmd/igozhang` 下 `.go/.js/.html/.md` 等文本须 **UTF-8**；禁止经 PowerShell 默认编码、`Out-File` 无 `-Encoding utf8`、或 GBK「另存为」写入。发版前跑 **`go test ./cmd/igozhang -run TestProjectTextFiles`**（全仓库文本扫描 + `mp3-init.js` UI 文案）。
- [ ] **首页首包**：首页 CSS 为 **`/assets/doc/home.css?v=…`**（immutable）；引语 JSON **只**由 `siteHeadBlock` 注入一次；应用层对 HTML/CSS/JS/JSON 做 **gzip**（`Accept-Encoding: gzip` 时响应含 `Content-Encoding: gzip`）。发版后 curl 首页确认体积与编码。

- [ ] **锁屏/通知「无缝切歌」**：点「下一首」或连播切歌时，Android Chrome 的媒体卡片**不得先消失再出现**。须**预缓存下一首**歌名与封面（`warmNextTrackMeta` / `mp3NextMetaCache`）；切歌瞬间用 `applyMediaSessionFromCache`（`mergeSessionArtwork` 重建 artwork，**禁止**用陈旧 `cache.artwork`）更新 `MediaMetadata` 且 **`playbackState` 保持 `playing`**；无内嵌封面时锁屏/通知 artwork **仅**用 **`fs-favicon-media-href`**（定制 raster 或 `/favicon.png`，**禁止**仅用 SVG）；**页内无封面则不显示封面区**（勿用 favicon.svg 占位）。**有内嵌封面时 artwork 仅封面 URL，禁止 favicon 排在封面之前**（否则 Android 常只显示 favicon）。**当前曲无封面时禁止把上一首封面排在 favicon 前**（`trackHasEmbeddedCover(primaryFn)`，否则 Android 只显示旧封面、看不到站点 favicon）。`msMetaFingerprint` 须含 `artwork[0].src`，勿仅用 `art.length`。副槽可播时：仅**当前曲已有内嵌封面**时 `mergeSessionArtwork` 才把上一首封面排在前面（`extraArtworkTrack`）。
- [ ] **与其它 App 共存**：`navigator.audioSession.type` 为 **`ambient`**；系统打断（`interrupted`）或被动 `pause` 时设 `mp3DeferToOtherAudio`，**禁止**后台 stall/恢复定时器抢播；用户点本页或锁屏 **▶** 后再 `clearMp3AudioDefer()`。
- [ ] **双槽状态表**（`mp3-init.js` 文件头注释）：可见+预取 → `tryPlayPrefetched`/dualSlot；可见无预取 → `playTrackWithDualBuffer`；隐藏 → 主槽 `playSpecificMp3OnMain(fast)`。切歌入口统一 **`advanceToTrack`**。
- [ ] **双槽无缝切歌**：自动连播、随机下一首、锁屏/通知「下一首」等**不得**在已解码的副槽可用时仍走「单槽改主 `<audio>.src`」路径。应优先 **`tryPlayPrefetched()`**（副槽已缓冲且与目标曲一致）或 **`playTrackWithDualBuffer` → `dualSlotPlaySecondaryBuffer`**（在副槽 `load` 就绪后 `play()`，于 **`playing`** 再 `swapMp3Outputs`）；**禁止**恢复已删除的「hidden 时把副槽 URL 偷到主槽同一元素」逻辑（会破坏 Media Session 无缝）。失败超时（如 12s）再回退 **`playSpecificMp3OnMain`**。
- [ ] **锁屏 / 页签后台（`document.hidden` / `visibilityState==="hidden"`）自动连播**：系统常**拒绝副槽** `HTMLMediaElement.play()`（无延续的用户激活链），会导致**无音轨、锁屏与通知栏媒体卡片消失且不再出现**。此时**必须**让下一首在**逻辑主槽**上通过 **`playSpecificMp3OnMain`** 起播（含 **`tryPlayPrefetched`** 命中预取时的主槽分支；**`playTrackWithDualBuffer`** 在 hidden 下须直接 `return false` 交由主槽）。**`dualSlotPlaySecondaryBuffer`** 里副槽 `play()` 的 `catch` 须优先 **`playSpecificMp3OnMain(mp3Current, …)`** 重试当前目标曲，勿在未恢复播放前把会话标成已结束。
- [ ] **曲终 `pause` 早于 `ended`**：自然播完时须在清 `setPositionState` / 停 keepalive **之前**把 **`mp3LoadingTrack`** 置真并同步 **`playbackState`/`metadata`**（见 `wireMp3Both("pause")`）；**`syncMediaSessionPlaybackState`** 须对「连播中且主槽 `ended` 仍有 `src`」兜底为 **`playing`**，避免 Chrome 将通知误判为用户暂停后收掉。
- [ ] **切歌加载中**：对外展示的曲目名/元数据应优先保持「**上一首**」直到新歌稳定进入 `playing`（通过 `mp3LockScreenTrack` 与 `mediaSessionTrackName()` 等逻辑）；**禁止**在加载间隙把 `navigator.mediaSession.metadata` 置为 **`null`**（会话仍视为播放中时用占位元数据）。
- [ ] **`playing` 事件**：更新 `mp3LockScreenTrack` 之前保留 `prevLock`，向 `syncMediaSessionMetadata` 传入 **`extraArtworkTrack`**，使 **`artwork`** 中在**新歌封面 URL 之前**仍带有**上一首封面 URL**（减轻 Android 先卸旧图、新图未加载完时的空白闪烁）。
- [ ] **`playSpecificMp3OnMain`（主槽直链）**：仅在双槽不可用时回退；在改写 `mp3Current` 为新文件名之前，若主 `<audio>` 已有 `src` 且为自动播放，将 **`mp3LockScreenTrack`** 设为 outgoing 文件名。对外入口 **`playSpecificMp3`** 须先尝试 **`playTrackWithDualBuffer`**。
- [ ] **visibility**：页面可见性应通过 **`fs:pagevisible` / `fs:pagehidden`** 统一派发（单一 `visibilitychange` 桥），避免多处重复监听导致顺序不可控。
- [ ] **HTTPS**：锁屏封面 URL、Media Session 能力在生产环境依赖 **HTTPS**；部署检查见 **`deploy/deploy.md`** §7。

---

## 4. 文本暂存与浏览防刷

- [ ] **`GET /api/stash`**：`ETag` + **`If-None-Match`** → **304** 行为与前端 `stashListETag`、脏检测、轮询退避一致；改服务端 JSON 形状时同步 README。
- [ ] **按 IP 限流**：`GET /api/browse`、`/api/stash` 的分钟窗阈值变更后更新 README 规则表与本文件 §2。
- [ ] **路径探测自动封禁**：改 `[security]` 阈值或排除前缀后更新 README 规则表；确认 `/api/auth/*`、`/stream/mp3/` 连播、文章多图加载不会误伤正常访客。

---

## 5. 统一认证、运维看板、黑名单与访客日志

- [ ] **IGO密钥认证**：`/ops`、`/igoadmin`、`/fs` 入口标题均为「IGO密钥认证」；`POST /api/auth/login` / `session`；token 仅存内存；失败限流与 **`secret07/auth-rl.json`** 行为与 README §8 一致。
- [ ] **`[secret-auth]` 文章**：`link.conf` 列出的 slug 访问时出认证面板、正文不进 HTML；`GET /api/doc/secret-article` 无 token 401、有 token 返回 bodyHTML；首页 excerpt 为空；改 conf **即时生效**（无需重启）。
- [ ] **开关解耦**：关 **`[ops]`** 后 `/ops` 404，但 `/fs`、`/igoadmin` 仍可登录；关 **`[auth]`** 或 **`[igokey]`** 后认证 API 404。
- [ ] **`/fs` 边界**：页内拉取/上传/暂存/浏览须 token；**`/files/…` 直链**与 **`/api/browse?path=assets/mp3`** **不**需登录。
- [ ] **`/ops` 看板**：统一认证后展示在线访客（默认 5 分钟窗）、近期 5 条、访问量统计；登录后另拉 **`/api/ops/targets`**；悬停/点击日/周/月数字可查看最近 30 条明细；字段未获取则不显示；**ZIP 导入 bundle 后**在线人数与访问次数应一致（内存看板已清空、cache 已恢复）。
- [ ] **黑名单**：仅精确 IPv4/IPv6；`secret07/ip-blacklist.conf` 持久化；命中全站 **403**；ZIP 导入后无需重启即生效。
- [ ] **`secret07/user-access.log`**：独立轮转；近 **30 天**日归档全量；月压缩（统计 + **200** 条）；月归档 **1 年**；单文件 **≤5MB** 截断。
- [ ] **文章 ZIP**：非空黑名单、访客 bundle 随 `/igoadmin` 导出；bundle **v3**（导出前月压实、`statsManifest`、瘦身明细、`insightsCache`）；导入写回 log/归档、恢复 cache、清空内存看板；`manifest.json` 含 **`accessBundleStats`**；导入响应 **`warnings`** 非空时核对统计是否完整。

---

## 6. 部署与反代（摘要 + 详见 deploy.md）

- [ ] **K8s 资源名**：使用 **`igozhang`** Deployment/Service/Ingress（**非**旧版 **`fs`** 或已废弃的 **`igo-doc`**）；容器端口 **3003**（旧版 **3002**）。
- [ ] **根路径部署**：Deployment **未设置** `BASE_PATH`；对外为 **`https://igozhang.cn/`**（或 HTTP 同源）。
- [ ] **hostPath**：宿主机数据目录统一为 **`/data/igozhang.cn`**（与 `deployment.yaml` / `deploy/deploy.md` / Nginx 证书路径一致）；容器内仍为 **`/igozhang.cn`**。
- [ ] **Ingress TLS**：`igo` 命名空间 Secret **`igozhang-cn-tls`** 有效；**不**强制 HTTP→HTTPS（见 **`deploy/deploy.md`** §4.1）。
- [ ] **健康检查**：探针请求 **`/health`**；滚动更新依赖进程 **graceful shutdown**（`terminationGracePeriodSeconds: 30`）。
- [ ] **镜像 tag**：生产勿长期仅依赖无版本 `:latest`；记录于 **`deploy/deploy.md`** §3。

完整上线步骤与常见问题：**[deploy/deploy.md](deploy/deploy.md)**。

---

## 7. 发版前快速自检（建议逐项打勾）

### 7.1 通用

- [ ] `go build ./cmd/igozhang` 通过；`go vet ./...` 与 `go test ./...`（或至少 `go test ./cmd/igozhang -run TestProjectTextFiles`）通过。
- [ ] **`README.md`** / **`checklist.md`** / **`deploy/deploy.md`** 中与本次改动相关段落已更新。

### 7.2 部署环境（生产或预发）

- [ ] **`deploy/deploy.md`** §7「上线检查清单」全部 relevant 项已确认。

### 7.3 移动端（若本期改动了 MP3 / Media Session / 封面）

- [ ] **Chrome + Android**：锁屏与通知栏 **下一首 / 上一首 / 播放暂停**；观察切歌时**元数据与封面是否闪烁为空**。
- [ ] **后台连播**：切后台或锁屏后自动连播 **≥5 首**；中断时可点锁屏 **▶** 恢复。
- [ ] **流式缓冲**：Network 为 **`/stream/mp3`**（非整首 Blob）；**首次播放**可适度预缓冲（约 **512KB**）；**切歌/后台连播**须走 **fast attach**（`attachMp3StreamFast`），禁止每首再等 1MB。
- [ ] **亮屏恢复**：后台被系统掐断后，`fs:pagevisible` 在会话仍标记为播放时应尝试 **`schedulePlaybackResume`** 或 **`pickRandomMp3(true,true)`** 恢复。
- [ ] Network 中 API 是否为站点根路径（如 **`/api/browse`**，域名 **`igozhang.cn`**）。

---

## 8. 数据目录布局（与 README / deploy §2 一致）

- [ ] 文章正文 **`assets/md/`**、图片 **`assets/doc-pic/`**；SQLite **`secret07/sqlitedb/doc.db`**（不进文章 ZIP）。
- [ ] **Mermaid**：文章 fenced `mermaid` 块在含 `language-mermaid` 时按需加载 `/assets/doc/mermaid.min.js`；`doc-code-chrome.js` 须跳过 mermaid；后台 `/igoadmin` **首屏不预挂** Mermaid，预览 HTML 含 mermaid 块时再动态加载 `mermaid.min.js` + `doc-mermaid.js` 并调用 `initDocMermaid`；**`/mermaid`**（及 **`/Mermaid`**）工具页预挂载并调用 `renderMermaidSource` 即时预览。升级库时替换 **`internal/docmermaid/mermaid.min.js`**（当前 **11.4.1**）并跑 **`go test ./cmd/igozhang -run TestDocMermaid`**。
- [ ] MP3 **`assets/mp3/`**、引语 **`assets/eng/phrases.md`**、知键 **`assets/eng/word3k-full.md`**（旧 **`assets/phrases.md`** / **`assets/word3k-full.md`** 启动时自动迁入 **`eng/`**）。
- [ ] 导入 ZIP：**合并**模式（不删本地独有 md；与 ZIP 字节一致的 md 不覆盖；ZIP 外且无索引的 md 按导入时刻登记，浏览量 1–5 随机）。
- [ ] 文章备份/恢复仅走 **`/igoadmin`** 导出 ZIP → 导入 ZIP；缓存 **`secret-tempfile/doc-export.zip`**；可选含 **`secret07/ip-blacklist.conf`**、**`secret07/user-access-bundle.json` v3**（导出前月压实、`statsManifest`、瘦身明细、`insightsCache`）。
- [ ] **`igo.conf`** 唯一源码 **`cmd/igozhang/igo.conf`**；改后 **`go generate ./cmd/igozhang/...`** 同步 **`deploy/secret07/igo.conf`**。

---

## 9. 修订记录（可选）

团队可在合并时于本节追加一行：`YYYY-MM-DD — 摘要 — MR/链接`，便于审计文档是否随代码更新。

- 2026-07-30 — 健壮性 P0：graceful shutdown、panic recovery、Server ReadHeaderTimeout；Docker BuildKit 缓存；CI；hostPath 统一 `/data/igozhang.cn` — （本迭代）
- 2026-07-19 — 首页加载优化 A：gzip、去重引语、外置 home.css / mp3-init.js — （本迭代）
- 2026-07-19 — `/igoadmin` 首屏不再预挂 Mermaid；预览含 mermaid 块时再动态加载 — （本迭代）
- 2026-07-19 — `link.conf` 新增 `[secret-auth]`：列出的文章 slug 访问须「IGO密钥认证」；`GET /api/doc/secret-article` — （本迭代）
- 2026-07-19 — 统一 Auth 模块（`/api/auth/login|session`）；ops/igoadmin/fs 共用「IGO密钥认证」；ops targets 另拉；`[auth]` 与 `[ops]` 解耦 — （本迭代）
- 2026-07-14 — 文章 / igoadmin 预览支持 Mermaid（embed 11.4.1，按需加载）— （本迭代）
- 2026-07-02 — 运维看板、IP 黑名单、`secret07/access.log` 及 ZIP 导入导出 — （本迭代）
