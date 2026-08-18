package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestParseHomeLinkConf_defaults(t *testing.T) {
	ext, slugs, secret := parseHomeLinkConf([]byte(secret07LinkConfTemplate))
	if len(ext) != 1 || ext[0].Text != "阿里镜像站" {
		t.Fatalf("external: %+v", ext)
	}
	if ext[0].Href != "https://developer.aliyun.com/mirror/" {
		t.Fatalf("href: %q", ext[0].Href)
	}
	if len(slugs) != 1 || slugs[0] != "chinesewedding" {
		t.Fatalf("internal slugs: %v", slugs)
	}
	if len(secret) != 0 {
		t.Fatalf("secret-auth default should be empty, got %v", secret)
	}
}

func TestParseHomeLinkConf_skipsCommentsAndInvalid(t *testing.T) {
	raw := `[external]
# 示例站 | https://example.com
有效 | https://example.org/
坏行无分隔符
[internal]
# bad-slug
goodslug
api
fs
[secret-auth]
# private_notes
secret_doc
ops
`
	ext, slugs, secret := parseHomeLinkConf([]byte(raw))
	if len(ext) != 1 || ext[0].Text != "有效" {
		t.Fatalf("external: %+v", ext)
	}
	if len(slugs) != 2 || slugs[0] != "goodslug" || slugs[1] != "fs" {
		t.Fatalf("slugs: %v", slugs)
	}
	if len(secret) != 1 || secret[0] != "secret_doc" {
		t.Fatalf("secret-auth: %v", secret)
	}
}

func TestParseHomeLinkConf_internalSiteRoute(t *testing.T) {
	_, slugs, _ := parseHomeLinkConf([]byte("[internal]\nfs\ntexas\nMermaid\n"))
	if len(slugs) != 3 || slugs[0] != "fs" || slugs[1] != "texas" || slugs[2] != "Mermaid" {
		t.Fatalf("slugs: %v", slugs)
	}
}

func TestParseHomeLinkConf_secretAuth(t *testing.T) {
	_, _, secret := parseHomeLinkConf([]byte("[secret-auth]\nprivate_notes\n/also_secret\nfs\n"))
	if len(secret) != 2 || secret[0] != "private_notes" || secret[1] != "also_secret" {
		t.Fatalf("secret: %v", secret)
	}
}

func TestIsSecretAuthSlug(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	if err := os.WriteFile(linkConfPath(dir), []byte("[secret-auth]\nhidden_doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isSecretAuthSlug(dir, "hidden_doc") {
		t.Fatal("expected hidden_doc protected")
	}
	if isSecretAuthSlug(dir, "chinesewedding") {
		t.Fatal("chinesewedding should not be protected")
	}
}

func TestBuildDocLinkAsideHTML_siteRouteMermaid(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	if err := os.WriteFile(linkConfPath(dir), []byte("[external]\n阿里 | https://example.com/\n[internal]\nMermaid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	html := buildDocLinkAsideHTML("", dir)
	if !strings.Contains(html, "Mermaid") {
		t.Fatalf("missing mermaid title: %s", html)
	}
	if !strings.Contains(html, `href="/Mermaid"`) {
		t.Fatalf("missing mermaid href: %s", html)
	}
}

func TestParseHomeLinkConf_maxFive(t *testing.T) {
	var b strings.Builder
	b.WriteString("[external]\n")
	for i := 0; i < 8; i++ {
		b.WriteString("link")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(" | https://example.com/\n")
	}
	ext, _, _ := parseHomeLinkConf([]byte(b.String()))
	if len(ext) != linkConfMaxLines {
		t.Fatalf("want %d external, got %d", linkConfMaxLines, len(ext))
	}
}

func TestBuildDocLinkAsideHTML(t *testing.T) {
	html := buildDocLinkAsideHTML("", "")
	if !strings.Contains(html, `doc-link-aside`) {
		t.Fatal("missing aside")
	}
	if !strings.Contains(html, "阿里镜像站") {
		t.Fatal("missing external link")
	}
	if !strings.Contains(html, "TeaCpu") {
		t.Fatal("missing TeaCpu title")
	}
	if !strings.Contains(html, "chinesewedding") {
		t.Fatal("missing internal slug fallback title")
	}
}

func TestBuildDocLinkAsideHTML_siteRouteFs(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	if err := os.WriteFile(linkConfPath(dir), []byte("[external]\n阿里 | https://example.com/\n[internal]\nfs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	html := buildDocLinkAsideHTML("", dir)
	if !strings.Contains(html, "文件服务") {
		t.Fatalf("missing fs title: %s", html)
	}
	if !strings.Contains(html, `href="/fs"`) {
		t.Fatalf("missing fs href: %s", html)
	}
}

func TestInitSecret07Files_createsLinkConf(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	p := filepath.Join(dir, secret07RelDir, linkConfFileName)
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "阿里镜像站") {
		t.Fatal("link.conf missing default external")
	}
	if !strings.Contains(string(raw), "# 示例（不生效）") {
		t.Fatal("link.conf missing inactive example")
	}
	if !strings.Contains(string(raw), "[secret-auth]") {
		t.Fatal("link.conf missing secret-auth section")
	}
}

func TestBuildIgoConfExportDiff_onlyNonDefault(t *testing.T) {
	diff := buildIgoConfExportDiff(opsConfigExample)
	if len(diff) != 0 {
		t.Fatalf("default config should export nothing, got:\n%s", diff)
	}
	custom := strings.Replace(string(opsConfigExample), "home_latest = 5", "home_latest = 8", 1)
	diff = buildIgoConfExportDiff([]byte(custom))
	if !strings.Contains(string(diff), "home_latest = 8") {
		t.Fatalf("missing changed key:\n%s", diff)
	}
	if !strings.Contains(string(diff), "[doc]") {
		t.Fatalf("missing section header:\n%s", diff)
	}
	merged := mergeIgoConfExportDiff(opsConfigExample, diff)
	if !strings.Contains(string(merged), "home_latest = 8") {
		t.Fatalf("merge diff: %s", merged)
	}
}

func TestReadIgoConfExportBytes_fullWhenCustom(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	if raw := readIgoConfExportBytes(dir); len(raw) != 0 {
		t.Fatalf("default igo.conf should not export, got %d bytes", len(raw))
	}
	p := opsConfigPath(dir)
	custom := strings.Replace(string(opsConfigExample), "home_latest = 5", "home_latest = 7", 1)
	if err := os.WriteFile(p, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	raw := readIgoConfExportBytes(dir)
	if !strings.Contains(string(raw), "home_latest = 7") {
		t.Fatalf("custom igo.conf: %s", raw)
	}
	if isIgoConfExportDiff(raw) {
		t.Fatal("export should be full file not diff")
	}
}

func TestReadLinkConfExportBytes_skipsDefault(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	if raw := readLinkConfExportBytes(dir); len(raw) != 0 {
		t.Fatalf("default link.conf should not export, got %q", raw)
	}
	p := linkConfPath(dir)
	if err := os.WriteFile(p, []byte("[external]\n自定义 | https://example.com/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	raw := readLinkConfExportBytes(dir)
	if !strings.Contains(string(raw), "自定义") {
		t.Fatalf("custom link.conf: %q", raw)
	}
}
