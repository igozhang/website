package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestZipNormalizeEntry_rejectsTraversal(t *testing.T) {
	bad := []string{"../x", "md/../../etc/passwd", "/md/a.md", "md/../x.md", `md\..\x.md`}
	for _, name := range bad {
		if _, ok := zipNormalizeEntry(name); ok {
			t.Fatalf("expected reject %q", name)
		}
	}
	if got, ok := zipNormalizeEntry(`md\foo.md`); !ok || got != "md/foo.md" {
		t.Fatalf("md/foo.md got %q ok=%v", got, ok)
	}
}

func TestDocExportImport_roundtrip(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	if _, err := os.Stat(root); err != nil {
		t.Skip("no repo assets")
	}
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := copyTreeForTest(root, assets); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Reconcile(); err != nil {
		t.Fatal(err)
	}
	// Seed metadata for roundtrip (repo ships md only).
	if err := st.ImportArticle(docindex.Meta{
		Slug: "redis_cfg", Title: "redis配置文件", Categories: "devops",
		Tags: "redis504配置文件, redis配置文件, Config", Author: "igozhang",
		PublishedAt: time.Now(), Views: 42,
	}, mustRead(t, filepath.Join(assets, "md", "redis_cfg.md"))); err != nil {
		t.Fatal(err)
	}

	zipBytes, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatal(err)
	}
	var hasMD, hasMeta bool
	for _, f := range zr.File {
		n, ok := zipNormalizeEntry(f.Name)
		if !ok {
			continue
		}
		if strings.HasPrefix(n, "md/") {
			hasMD = true
		}
		if strings.HasPrefix(n, "meta/") && strings.HasSuffix(n, ".json") {
			hasMeta = true
		}
		if n == "doc_info.md" || n == "doc_views.md" {
			t.Fatalf("legacy file must not be exported: %s", n)
		}
	}
	if !hasMD || !hasMeta {
		t.Fatalf("export missing md or meta (md=%v meta=%v)", hasMD, hasMeta)
	}

	initSecret07Files(dir)
	zipBytes2, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	zr2, err := zip.NewReader(bytes.NewReader(zipBytes2), int64(len(zipBytes2)))
	if err != nil {
		t.Fatal(err)
	}
	var hasLinkConf, hasIgoDiff bool
	for _, f := range zr2.File {
		switch f.Name {
		case "secret07/link.conf":
			hasLinkConf = true
		case "secret07/igo.conf":
			hasIgoDiff = true
		}
	}
	if hasLinkConf {
		t.Fatal("default link.conf should not export")
	}
	if hasIgoDiff {
		t.Fatal("default igo.conf should not export diff")
	}
	linkPath := filepath.Join(dir, secret07RelDir, linkConfFileName)
	linkRaw, _ := os.ReadFile(linkPath)
	customLink := strings.Replace(string(linkRaw), "chinesewedding", "redis_cfg", 1)
	if err := os.WriteFile(linkPath, []byte(customLink), 0o644); err != nil {
		t.Fatal(err)
	}
	zipBytesLink, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	zrLink, err := zip.NewReader(bytes.NewReader(zipBytesLink), int64(len(zipBytesLink)))
	if err != nil {
		t.Fatal(err)
	}
	hasLinkConf = false
	for _, f := range zrLink.File {
		if f.Name == "secret07/link.conf" {
			hasLinkConf = true
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(rc)
			_ = rc.Close()
			if !strings.Contains(string(body), "redis_cfg") {
				t.Fatalf("link.conf export: %s", body)
			}
		}
	}
	if !hasLinkConf {
		t.Fatal("custom link.conf should export")
	}
	confPath := filepath.Join(dir, secret07RelDir, opsConfigName)
	raw, _ := os.ReadFile(confPath)
	custom := strings.Replace(string(raw), "home_latest = 5", "home_latest = 9", 1)
	if err := os.WriteFile(confPath, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	zipBytes3, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	zr3, err := zip.NewReader(bytes.NewReader(zipBytes3), int64(len(zipBytes3)))
	if err != nil {
		t.Fatal(err)
	}
	hasIgoDiff = false
	for _, f := range zr3.File {
		if f.Name == "secret07/igo.conf" {
			hasIgoDiff = true
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(rc)
			_ = rc.Close()
			if !strings.Contains(string(body), "home_latest = 9") {
				t.Fatalf("igo.conf export: %s", body)
			}
			if isIgoConfExportDiff(body) {
				t.Fatalf("igo.conf should export full file, got diff:\n%s", body)
			}
		}
	}
	if !hasIgoDiff {
		t.Fatal("export missing igo.conf after change")
	}

	dir2 := t.TempDir()
	assets2 := filepath.Join(dir2, "assets")
	_ = os.MkdirAll(filepath.Join(assets2, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets2, "doc-pic"), 0o755)
	st2, err := docindex.Open(docDBPath(dir2), assets2)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	// 本地独有文章：导入后应保留并纳入索引。
	if err := st2.WriteArticle(docindex.Meta{
		Slug: "orphan_extra", Title: "Extra", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Extra\n")); err != nil {
		t.Fatal(err)
	}
	count, _, err := importDocZip(dir, assets2, st2, zr)
	if err != nil {
		t.Fatal(err)
	}
	if count < 200 {
		t.Fatalf("import count %d", count)
	}
	if ok, _ := st2.Exists("orphan_extra"); !ok {
		t.Fatal("import must keep local-only article orphan_extra")
	}
	slug := "redis_cfg"
	body, err := st2.ReadBody(slug)
	if err != nil || len(body) == 0 {
		t.Fatalf("read %s: %v", slug, err)
	}
	got, err := st2.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "redis配置文件" || got.Views != 42 {
		t.Fatalf("meta restore: %+v", got)
	}

	zrFull, err := zip.NewReader(bytes.NewReader(zipBytes3), int64(len(zipBytes3)))
	if err != nil {
		t.Fatal(err)
	}
	dir3 := t.TempDir()
	initSecret07Files(dir3)
	assets3 := filepath.Join(dir3, "assets")
	_ = os.MkdirAll(filepath.Join(assets3, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets3, "doc-pic"), 0o755)
	st3, err := docindex.Open(docDBPath(dir3), assets3)
	if err != nil {
		t.Fatal(err)
	}
	defer st3.Close()
	if _, _, err := importDocZip(dir3, assets3, st3, zrFull); err != nil {
		t.Fatal(err)
	}
	secret3 := filepath.Join(dir3, secret07RelDir)
	linkBody, _ := os.ReadFile(filepath.Join(secret3, linkConfFileName))
	if !strings.Contains(string(linkBody), "redis_cfg") {
		t.Fatalf("link.conf not restored: %s", linkBody)
	}
	igoBody, _ := os.ReadFile(filepath.Join(secret3, opsConfigName))
	if !strings.Contains(string(igoBody), "home_latest = 9") {
		t.Fatalf("igo.conf not restored: %s", igoBody)
	}
}

func TestDocExportZip_secret07Optional(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: "demo", Title: "Demo", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	initSecret07Files(dir)
	secret := filepath.Join(dir, secret07RelDir)

	zipNoExtra, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"secret07/ssh.pass", "secret07/banner.png", "secret07/favicon.png", "secret07/texas.png"} {
		if zipHasEntry(zipNoExtra, name) {
			t.Fatalf("default export should not include %s", name)
		}
	}

	if err := os.WriteFile(filepath.Join(secret, "ssh.pass"), []byte("10.0.0.1 22 root secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "banner.png"), []byte("PNG"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "favicon.png"), []byte("ICO"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "texas.png"), []byte("TEXAS"), 0o644); err != nil {
		t.Fatal(err)
	}
	zipWithExtra, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"secret07/ssh.pass", "secret07/banner.png", "secret07/favicon.png", "secret07/texas.png"} {
		if !zipHasEntry(zipWithExtra, name) {
			t.Fatalf("export missing %s", name)
		}
	}

	dir3 := t.TempDir()
	assets3 := filepath.Join(dir3, "assets")
	_ = os.MkdirAll(filepath.Join(assets3, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets3, "doc-pic"), 0o755)
	st3, err := docindex.Open(docDBPath(dir3), assets3)
	if err != nil {
		t.Fatal(err)
	}
	defer st3.Close()
	if err := st3.WriteArticle(docindex.Meta{
		Slug: "demo", Title: "Demo", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	zrImport, err := zip.NewReader(bytes.NewReader(zipWithExtra), int64(len(zipWithExtra)))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir3, assets3, st3, zrImport); err != nil {
		t.Fatal(err)
	}
	secret3 := filepath.Join(dir3, secret07RelDir)
	for _, name := range []string{"ssh.pass", "banner.png", "favicon.png", "texas.png"} {
		if _, err := os.Stat(filepath.Join(secret3, name)); err != nil {
			t.Fatalf("import missing secret07/%s: %v", name, err)
		}
	}
	if resolveSecret07Banner(dir3) == nil || resolveSecret07Favicon(dir3) == nil || resolveSecret07TexasCardBack(dir3) == nil {
		t.Fatal("imported banner/favicon/texas must resolve")
	}
}

func TestImportDocZip_igoConfLegacyDiff(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: "demo", Title: "Demo", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	initSecret07Files(dir)

	diff := buildIgoConfExportDiff([]byte(strings.Replace(string(opsConfigExample), "home_latest = 5", "home_latest = 11", 1)))
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := zipAddBytes(zw, "meta/demo.json", []byte(`{"slug":"demo","title":"Demo","author":"a"}`)); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "md/demo.md", []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "secret07/igo.conf", diff); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir, assets, st, zr); err != nil {
		t.Fatal(err)
	}
	igoBody, _ := os.ReadFile(opsConfigPath(dir))
	if !strings.Contains(string(igoBody), "home_latest = 11") {
		t.Fatalf("legacy diff igo.conf not merged: %s", igoBody)
	}
	if isIgoConfExportDiff(igoBody) {
		t.Fatal("imported igo.conf should be full config not diff header")
	}
}

func TestSecret07ZipExportImportParity(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: "demo", Title: "Demo", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	initSecret07Files(dir)
	secret := filepath.Join(dir, secret07RelDir)
	_ = os.WriteFile(filepath.Join(secret, "ssh.pass"), []byte("10.0.0.1 22 root secret\n"), 0o600)
	_ = os.WriteFile(filepath.Join(secret, "banner.png"), []byte("PNG"), 0o644)
	_ = os.WriteFile(filepath.Join(secret, "favicon.png"), []byte("ICO"), 0o644)
	_ = os.WriteFile(filepath.Join(secret, "texas.png"), []byte("TEXAS"), 0o644)
	linkCustom := "[external]\n自定义 | https://example.com/\n[internal]\ndemo\n"
	_ = os.WriteFile(linkConfPath(dir), []byte(linkCustom), 0o644)
	customIgo := strings.Replace(string(opsConfigExample), "home_latest = 5", "home_latest = 12", 1)
	_ = os.WriteFile(opsConfigPath(dir), []byte(customIgo), 0o644)
	_ = os.WriteFile(commonCommandsPath(dir), []byte("kubectl get pods\n"), 0o644)

	zipBytes, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		entry, ok := zipNormalizeEntry(f.Name)
		if !ok || !strings.HasPrefix(entry, "secret07/") {
			continue
		}
		if _, ok := secret07ZipImportTarget(dir, entry); !ok {
			t.Fatalf("export entry %q has no import handler", entry)
		}
	}

	dir2 := t.TempDir()
	assets2 := filepath.Join(dir2, "assets")
	_ = os.MkdirAll(filepath.Join(assets2, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets2, "doc-pic"), 0o755)
	st2, err := docindex.Open(docDBPath(dir2), assets2)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	if err := st2.WriteArticle(docindex.Meta{
		Slug: "other", Title: "Other", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Other\n")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir2, assets2, st2, zr); err != nil {
		t.Fatal(err)
	}
	secret2 := filepath.Join(dir2, secret07RelDir)
	for _, name := range []string{"ssh.pass", "banner.png", "favicon.png", "texas.png", linkConfFileName, opsConfigName, commonCommandsFileName} {
		if _, err := os.Stat(filepath.Join(secret2, name)); err != nil {
			t.Fatalf("import missing %s: %v", name, err)
		}
	}
	cmds := listCommonCommands(dir2)
	if len(cmds) != 1 || cmds[0] != "kubectl get pods" {
		t.Fatalf("common-commands not restored: %v", cmds)
	}
	linkBody, _ := os.ReadFile(filepath.Join(secret2, linkConfFileName))
	if !strings.Contains(string(linkBody), "自定义") {
		t.Fatalf("link.conf not restored: %s", linkBody)
	}
	igoBody, _ := os.ReadFile(filepath.Join(secret2, opsConfigName))
	if !strings.Contains(string(igoBody), "home_latest = 12") {
		t.Fatalf("igo.conf not restored: %s", igoBody)
	}
}

func TestDocExportZip_engAssets(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	engDir := filepath.Join(assets, assetsEngRelDir)
	if err := os.MkdirAll(engDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	phrases := "Hello world\uFF5C你好\n"
	word3k := "hello /h\u0259\u02C8lo\u028A/ 你好\n"
	if err := os.WriteFile(filepath.Join(engDir, phrasesFileName), []byte(phrases), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(engDir, word3kFileName), []byte(word3k), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: "demo", Title: "Demo", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}

	zipBytes, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"eng/phrases.md", "eng/word3k-full.md"} {
		if !zipHasEntry(zipBytes, name) {
			t.Fatalf("export missing %s", name)
		}
	}

	dir2 := t.TempDir()
	assets2 := filepath.Join(dir2, "assets")
	_ = os.MkdirAll(filepath.Join(assets2, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets2, "doc-pic"), 0o755)
	st2, err := docindex.Open(docDBPath(dir2), assets2)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	if err := st2.WriteArticle(docindex.Meta{
		Slug: "other", Title: "Other", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Other\n")); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		entry, ok := zipNormalizeEntry(f.Name)
		if !ok || !strings.HasPrefix(entry, "eng/") {
			continue
		}
		if dest, ok := zipAssetsDest(assets2, entry); !ok || dest == "" {
			t.Fatalf("eng entry %q has no import dest", entry)
		}
	}
	if _, _, err := importDocZip(dir2, assets2, st2, zr); err != nil {
		t.Fatal(err)
	}
	gotPhrases, err := os.ReadFile(filepath.Join(assets2, assetsEngRelDir, phrasesFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(gotPhrases) != phrases {
		t.Fatalf("phrases restore: %q", gotPhrases)
	}
	gotWord3k, err := os.ReadFile(filepath.Join(assets2, assetsEngRelDir, word3kFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(gotWord3k) != word3k {
		t.Fatalf("word3k restore: %q", gotWord3k)
	}
}

func TestImportDocZip_skipsIdenticalMd(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	mdDir := filepath.Join(assets, "md")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	const slug = "keep_me"
	body := []byte("# Keep\n\nlocal unchanged\n")
	mdPath := filepath.Join(mdDir, slug+".md")
	if err := os.WriteFile(mdPath, body, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	meta, _ := json.Marshal(docindex.Meta{
		Slug: slug, Title: "From Zip", Author: "igozhang", PublishedAt: time.Now(), Views: 99,
	})
	if err := zipAddBytes(zw, "meta/"+slug+".json", meta); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "md/"+slug+".md", body); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir, assets, st, zr); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("md should be unchanged: %q", got)
	}
	m, err := st.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "From Zip" || m.Views != 99 {
		t.Fatalf("meta from zip: %+v", m)
	}
}

func TestImportDocZip_indexesLocalOnlyMd(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	mdDir := filepath.Join(assets, "md")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	const slug = "local_only"
	if err := os.WriteFile(filepath.Join(mdDir, slug+".md"), []byte("# Local Only\n\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := zipAddBytes(zw, "meta/other.json", []byte(`{"slug":"other","title":"Other","author":"a","publishedAt":"2026-01-01T00:00:00Z"}`)); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "md/other.md", []byte("# Other\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir, assets, st, zr); err != nil {
		t.Fatal(err)
	}
	if ok, _ := st.Exists(slug); !ok {
		t.Fatal("local_only should be indexed")
	}
	m, err := st.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	if m.Views < 1 || m.Views > 5 {
		t.Fatalf("views want 1-5 got %d", m.Views)
	}
}

func TestDocExportZip_skipsMetaWhenMdMissing(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	mdDir := filepath.Join(assets, "md")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: "has_md", Title: "Has MD", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Has MD\n")); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteArticle(docindex.Meta{
		Slug: "ghost", Title: "Ghost", Author: "a", PublishedAt: time.Now(), Views: 7,
	}, []byte("# Ghost\n")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(mdDir, "ghost.md")); err != nil {
		t.Fatal(err)
	}

	zipBytes, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatal(err)
	}
	if zipHasEntry(zipBytes, "meta/ghost.json") {
		t.Fatal("orphan meta must not export without md")
	}
	if !zipHasEntry(zipBytes, "meta/has_md.json") {
		t.Fatal("paired meta must export")
	}
	var manifest map[string]any
	for _, f := range zr.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(rc)
		_ = rc.Close()
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
	}
	if manifest["count"].(float64) != 1 {
		t.Fatalf("manifest count want 1 got %v", manifest["count"])
	}
	missing, ok := manifest["missingMd"].([]any)
	if !ok || len(missing) != 1 || missing[0] != "ghost" {
		t.Fatalf("missingMd: %v", manifest["missingMd"])
	}

	dir2 := t.TempDir()
	assets2 := filepath.Join(dir2, "assets")
	_ = os.MkdirAll(filepath.Join(assets2, "md"), 0o755)
	st2, err := docindex.Open(docDBPath(dir2), assets2)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	if _, _, err := importDocZip(dir, assets2, st2, zr); err != nil {
		t.Fatal(err)
	}
	if ok, _ := st2.Exists("ghost"); ok {
		t.Fatal("ghost should not be indexed from export without meta")
	}
}

func TestImportDocZip_indexesZipMdWithoutMeta(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	mdDir := filepath.Join(assets, "md")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	const slug = "orphan_in_zip"
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := zipAddBytes(zw, "md/"+slug+".md", []byte("# Orphan Title\n\nbody\n")); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "meta/other.json", []byte(`{"slug":"other","title":"Other","author":"a","publishedAt":"2026-01-01T00:00:00Z"}`)); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "md/other.md", []byte("# Other\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir, assets, st, zr); err != nil {
		t.Fatal(err)
	}
	m, err := st.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Orphan Title" {
		t.Fatalf("title want Orphan Title got %q", m.Title)
	}
}

func TestImportDocZip_skipsOrphanMetaWithoutMd(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	mdDir := filepath.Join(assets, "md")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := zipAddBytes(zw, "meta/ghost.json", []byte(`{"slug":"ghost","title":"Ghost","author":"a","views":9}`)); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "meta/demo.json", []byte(`{"slug":"demo","title":"Demo","author":"a","publishedAt":"2026-01-01T00:00:00Z"}`)); err != nil {
		t.Fatal(err)
	}
	if err := zipAddBytes(zw, "md/demo.md", []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir, assets, st, zr); err != nil {
		t.Fatalf("legacy orphan meta must not fail import: %v", err)
	}
	if ok, _ := st.Exists("ghost"); ok {
		t.Fatal("ghost without md must not be indexed")
	}
	if ok, _ := st.Exists("demo"); !ok {
		t.Fatal("demo should be indexed")
	}
}

func zipHasEntry(zipBytes []byte, name string) bool {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return false
	}
	for _, f := range zr.File {
		if f.Name == name {
			return true
		}
	}
	return false
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func copyTreeForTest(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
}
