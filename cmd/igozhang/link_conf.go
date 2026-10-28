package main

import (
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/igokorea/igozhang/internal/docindex"
)

const (
	linkConfFileName      = "link.conf"
	linkConfMaxLines      = 5
	linkConfSecretAuthMax = 50
)

const secret07LinkConfTemplate = `[external]
阿里镜像站 | https://developer.aliyun.com/mirror/

[internal]
chinesewedding

[secret-auth]

# === 说明与案例（均为注释，不生效；生效配置见文件头部）===
# 首页与文章页右侧链接胶囊与密钥保护列表
# [external]/[internal] 胶囊约 60s 缓存；[secret-auth] 每次请求读盘，改完即生效（无需重启）
#
# [external] 外链胶囊：无标题，一行一条，最多 5 条
# 格式：显示文字 | URL（仅 http/https）
# 以 # 开头的行不生效
# 示例（不生效）：# 示例站 | https://example.com
#
# [internal] 内链胶囊：标题 TeaCpu，一行一个 slug，最多 5 条
# 文章 slug：展示文章标题，链接 /slug
# 站点页 slug：fs（文件）、gomoku、xiangqi、texas、zhijian、mermaid、ops 等，展示固定标题
# 示例（不生效）：# example-slug
#
# [secret-auth] 需「IGO密钥认证」才能阅读的文章 slug（不展示胶囊，最多 50 条）
# 一行一个文章 slug；访问 /slug 时先验证密钥，通过后才加载正文
# 示例（不生效）：# private_notes
`

type homeExternalLink struct {
	Text string
	Href string
}

type homeInternalLink struct {
	Title string
	Href  string
}

// homeInternalRouteTitles 内链可引用的站点页 slug → 展示标题（非文章路由）。
var homeInternalRouteTitles = map[string]string{
	"fs":      "文件服务",
	"gomoku":  "五子棋",
	"xiangqi": "象棋",
	"texas":   "德扑",
	"zhijian": "知键",
	"mermaid": "Mermaid",
	"ops":     "运维",
}

func homeInternalRouteTitle(slug string) string {
	return homeInternalRouteTitles[strings.ToLower(strings.TrimSpace(slug))]
}

func homeInternalLinkAllowed(slug string) bool {
	if slug == "" {
		return false
	}
	if homeInternalRouteTitle(slug) != "" {
		return true
	}
	return docindex.ValidSlug(slug) && !docReservedSlug(slug)
}

func homeInternalLinkTitle(slug string) string {
	if docStore != nil {
		if meta, err := docStore.Get(slug); err == nil && strings.TrimSpace(meta.Title) != "" {
			return meta.Title
		}
	}
	if t := homeInternalRouteTitle(slug); t != "" {
		return t
	}
	return slug
}

func linkConfPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), linkConfFileName)
}

func readLinkConfBytes(dataRoot string) []byte {
	if dataRoot == "" {
		return []byte(secret07LinkConfTemplate)
	}
	raw, err := os.ReadFile(linkConfPath(dataRoot))
	if err != nil {
		return []byte(secret07LinkConfTemplate)
	}
	if !secret07HasEffectiveLines(raw) {
		return []byte(secret07LinkConfTemplate)
	}
	return raw
}

func parseHomeLinkConf(raw []byte) (external []homeExternalLink, internalSlugs, secretAuthSlugs []string) {
	section := ""
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}
		switch section {
		case "external":
			if len(external) >= linkConfMaxLines {
				continue
			}
			parts := strings.SplitN(line, "|", 2)
			if len(parts) < 2 {
				continue
			}
			text := strings.TrimSpace(parts[0])
			href := strings.TrimSpace(parts[1])
			if text == "" || href == "" || !isSafeFooterHref(href) {
				continue
			}
			external = append(external, homeExternalLink{Text: text, Href: href})
		case "internal":
			if len(internalSlugs) >= linkConfMaxLines {
				continue
			}
			slug := strings.TrimSpace(strings.TrimPrefix(line, "/"))
			if !homeInternalLinkAllowed(slug) {
				continue
			}
			internalSlugs = append(internalSlugs, slug)
		case "secret-auth":
			if len(secretAuthSlugs) >= linkConfSecretAuthMax {
				continue
			}
			slug := strings.TrimSpace(strings.TrimPrefix(line, "/"))
			if !docindex.ValidSlug(slug) || docReservedSlug(slug) {
				continue
			}
			secretAuthSlugs = append(secretAuthSlugs, slug)
		}
	}
	return external, internalSlugs, secretAuthSlugs
}

func listHomeExternalLinks(dataRoot string) []homeExternalLink {
	ext, _, _ := parseHomeLinkConf(readLinkConfBytes(dataRoot))
	return ext
}

func listHomeInternalLinks(dataRoot, pub string) []homeInternalLink {
	_, slugs, _ := parseHomeLinkConf(readLinkConfBytes(dataRoot))
	out := make([]homeInternalLink, 0, len(slugs))
	for _, slug := range slugs {
		out = append(out, homeInternalLink{
			Title: homeInternalLinkTitle(slug),
			Href:  pubPath(pub, slug),
		})
	}
	return out
}

func listSecretAuthSlugs(dataRoot string) []string {
	_, _, slugs := parseHomeLinkConf(readLinkConfBytes(dataRoot))
	return slugs
}

// isSecretAuthSlug 判断文章 slug 是否在 link.conf [secret-auth] 中（须认证后才可读正文）。
func isSecretAuthSlug(dataRoot, slug string) bool {
	slug = strings.TrimSpace(slug)
	if slug == "" || !docindex.ValidSlug(slug) {
		return false
	}
	for _, s := range listSecretAuthSlugs(dataRoot) {
		if s == slug {
			return true
		}
	}
	return false
}

func buildDocLinkAsideHTML(pub, dataRoot string) string {
	ext := listHomeExternalLinks(dataRoot)
	in := listHomeInternalLinks(dataRoot, pub)
	if len(ext) == 0 && len(in) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<aside class="doc-link-aside" aria-label="快捷链接">`)
	if len(ext) > 0 {
		b.WriteString(`<div class="doc-link-capsule doc-link-capsule-ext">`)
		b.WriteString(`<ul class="doc-link-list">`)
		for _, l := range ext {
			b.WriteString(`<li><a href="`)
			b.WriteString(html.EscapeString(l.Href))
			b.WriteString(`" target="_blank" rel="noopener noreferrer">`)
			b.WriteString(html.EscapeString(l.Text))
			b.WriteString(`</a></li>`)
		}
		b.WriteString(`</ul></div>`)
	}
	if len(in) > 0 {
		b.WriteString(`<div class="doc-link-capsule doc-link-capsule-int">`)
		b.WriteString(`<h3 class="doc-link-capsule-title">TeaCpu</h3>`)
		b.WriteString(`<ul class="doc-link-list">`)
		for _, l := range in {
			b.WriteString(`<li><a href="`)
			b.WriteString(html.EscapeString(l.Href))
			b.WriteString(`">`)
			b.WriteString(html.EscapeString(l.Title))
			b.WriteString(`</a></li>`)
		}
		b.WriteString(`</ul></div>`)
	}
	b.WriteString(`</aside>`)
	return b.String()
}
