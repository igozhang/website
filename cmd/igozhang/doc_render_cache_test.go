package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
	"github.com/igokorea/igozhang/internal/markdown"
)

func TestDocRenderCache_hitAfterFirstRender(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	body := []byte("# Title\n\nhello **world**")
	if err := st.WriteArticle(docindex.Meta{Slug: "cache", Title: "Title", Author: "a", PublishedAt: time.Now()}, body); err != nil {
		t.Fatal(err)
	}
	html1, err := renderDocArticleHTML("", "cache", body)
	if err != nil {
		t.Fatal(err)
	}
	html2, err := renderDocArticleHTML("", "cache", body)
	if err != nil {
		t.Fatal(err)
	}
	if string(html1) != string(html2) {
		t.Fatal("expected cache hit with identical output")
	}
	invalidateDocRenderCache("cache")
	html3, err := renderDocArticleHTML("", "cache", body)
	if err != nil {
		t.Fatal(err)
	}
	if string(html3) != string(html1) {
		t.Fatal("expected same render after invalidate+rebuild")
	}
}

func TestRenderDocArticleHTML_rewritesFragmentHrefs(t *testing.T) {
	invalidateAllDocRenderCache()
	body := []byte("# T\n\n## Pod\n\n[go](#pod)\n")
	html, err := renderDocArticleHTML("", "k8s_crud", body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(html)
	if !strings.Contains(s, `href="/k8s_crud#pod"`) {
		t.Fatalf("expected path-pinned fragment href, got:\n%s", s)
	}
	if strings.Contains(s, `href="#pod"`) {
		t.Fatal("bare href=#pod must be rewritten (breaks under <base href=/>)")
	}

	// Cache must not bake in pub=""; BASE_PATH render still gets correct prefix.
	htmlFS, err := renderDocArticleHTML("/fs", "k8s_crud", body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(htmlFS), `href="/fs/k8s_crud#pod"`) {
		t.Fatalf("BASE_PATH rewrite failed:\n%s", htmlFS)
	}
}

func TestRewriteFragmentHrefs_idempotent(t *testing.T) {
	in := []byte(`<a href="#pod">Pod</a>`)
	once := markdown.RewriteFragmentHrefs(in, "/k8s_crud")
	twice := markdown.RewriteFragmentHrefs(once, "/k8s_crud")
	if string(once) != string(twice) {
		t.Fatalf("rewrite should be idempotent:\n%s\nvs\n%s", once, twice)
	}
	if strings.Contains(string(twice), `/k8s_crud/k8s_crud#`) {
		t.Fatal("double-prefixed path")
	}
}

func TestDocNavCache_reusesHTML(t *testing.T) {
	invalidateDocChromeCache()
	cats := []string{"devops", "k8s"}
	a := buildDocNavHTMLCached(cats, 10, "", true)
	b := buildDocNavHTMLCached(cats, 10, "", true)
	if a != b {
		t.Fatal("expected cached nav html")
	}
}

func TestServeDocArticleCSS_immutable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/assets/doc/article.css", nil)
	rec := httptest.NewRecorder()
	serveDocArticleCSS(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc == "" || cc == "no-store" {
		t.Fatalf("Cache-Control: %q", cc)
	}
	if len(rec.Body.Bytes()) == 0 {
		t.Fatal("empty css body")
	}
}

func TestServeDocMp3CSS_immutable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/assets/doc/mp3.css", nil)
	rec := httptest.NewRecorder()
	serveDocMp3CSS(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc == "" || cc == "no-store" {
		t.Fatalf("Cache-Control: %q", cc)
	}
	if len(rec.Body.Bytes()) == 0 {
		t.Fatal("empty mp3 css body")
	}
}

func TestDocArticleETag_stable(t *testing.T) {
	if docHTMLRenderRev == "" {
		t.Fatal("docHTMLRenderRev must be set so heading-id / render rule changes bust ETag")
	}
	e1 := docArticleETag("squid_ins", 100, 50)
	e2 := docArticleETag("squid_ins", 100, 50)
	if e1 != e2 {
		t.Fatalf("etag unstable: %s vs %s", e1, e2)
	}
	e3 := docArticleETag("squid_ins", 101, 50)
	if e1 == e3 {
		t.Fatal("etag should change when md mtime changes")
	}
	e4 := docArticleETag("squid_ins", 100, 51)
	if e1 == e4 {
		t.Fatal("etag should change when link.conf mtime changes")
	}
}

func TestDocArticleCacheControl_alignedWithChromeTTL(t *testing.T) {
	if docArticleBrowserMaxAgeSec != int(docChromeCacheTTL/time.Second) {
		t.Fatalf("browser max-age %d should match chrome TTL %s", docArticleBrowserMaxAgeSec, docChromeCacheTTL)
	}
	want := fmt.Sprintf("private, max-age=%d", docArticleBrowserMaxAgeSec)
	if got := docArticleCacheControl(); got != want {
		t.Fatalf("Cache-Control: got %q want %q", got, want)
	}
}
