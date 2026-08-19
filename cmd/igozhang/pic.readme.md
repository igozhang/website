# 站点图标、页眉背景与德扑牌背（favicon / banner / texas）

本文件仅为说明，**不参与程序解析**。将图片放在 **`DATA_ROOT/secret07/`**（与 `igo.conf` 同级），替换后**刷新页面**即可，**无需重启 Pod**。有定制时 favicon / banner / texas 随 `/igoadmin` 导出 ZIP 一并备份；**黑名单**、**白名单**（各自非空时）与 **`user-access-bundle.json`**（访客明细 + 统计 + 看板 cache）亦在 ZIP 内。

## Favicon（站点图标）

| 项 | 说明 |
|----|------|
| **文件名** | `favicon.png` / `favicon.jpg` / `favicon.jpeg`，或 `fav.png` / `fav.jpg` / `fav.jpeg`（大小写不敏感） |
| **大小** | 单文件 **≤ 100KB**，超出则忽略，回退内嵌默认图标 |
| **用途** | 浏览器标签页图标；MP3 无内嵌封面时锁屏 / 通知栏占位图（经 `/favicon.png` 等路径提供） |

**多文件同时存在时的选取优先级**：

1. stem：`favicon` 优先于 `fav`
2. 扩展名：`png` > `jpeg` > `jpg`

**未放置有效文件时**：使用内嵌魂斗罗风格 SVG（`/favicon.svg`）及程序生成的 `/favicon.png`。

## Banner（页眉背景）

| 项 | 说明 |
|----|------|
| **文件名** | `banner.png` / `banner.jpg` / `banner.jpeg`（大小写不敏感，仅上述三种） |
| **大小** | 单文件 **≤ 200KB**，超出则忽略，页眉不显示背景图 |
| **效果** | 首页与文章页顶栏背景，带半透明渐变蒙版；无有效文件时不启用 banner 样式 |

**蒙版强度**：在 `igo.conf` 的 `[site]` 段设置 `banner_mask`（**0**=全透明，**100**=全遮挡；默认 **66**）。修改 `igo.conf` 后须**重启 Pod**。

## Texas（德扑牌背）

| 项 | 说明 |
|----|------|
| **文件名** | `texas.png` / `texas.jpg` / `texas.jpeg`（大小写不敏感，仅上述三种） |
| **大小** | 单文件 **≤ 150KB**，超出则忽略，页面回退 Canvas 绘制花纹牌背 |
| **用途** | `/texas` 德扑人机对战页中未亮明的手牌背面；经 `/texas-card-back.png` 等路径提供 |
| **扩展名优先级** | 多文件同时存在时：`png` > `jpeg` > `jpg` |

**未放置有效文件时**：前端用 Canvas 绘制默认牌背，无需额外配置。

## 目录示例

```text
secret07/
  igo.conf
  favicon.png      ← 站点图标（≤100KB）
  banner.jpg       ← 页眉背景（≤200KB）
  texas.png        ← 德扑牌背（≤150KB）
  pic.readme.md    ← 本说明（可删，不影响运行）
```

## ZIP 备份 / 恢复

- **导出**：存在有效 favicon / banner / texas 时，ZIP 内包含 `secret07/favicon.*`、`secret07/banner.*` 或 `secret07/texas.*`；**黑名单非空**时含 `secret07/ip-blacklist.conf`；**白名单非空**时含 `secret07/ip-whitelist.conf`；**存在访客数据**时含 `secret07/user-access-bundle.json`（**v3**：导出前月压实、statsManifest、瘦身明细、insightsCache）
- **导入**：写回 `secret07/`；图片类清除同类型旧文件后再写入；黑名单 / 白名单导入后**立即生效**；`user-access-bundle.json` 写回 log/归档、恢复看板 cache、清空内存看板；统计偏差时导入响应含 **`warnings`**
