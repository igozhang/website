package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLoadWord3kSource_Embedded(t *testing.T) {
	src := loadWord3kSource(t.TempDir())
	if len(src.lines) == 0 {
		t.Fatal("expected embedded word3k source lines")
	}
	if src.external {
		t.Fatal("temp dir should use embedded corpus")
	}
}

func TestWord3kHandler_GET(t *testing.T) {
	initWord3kCache(t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/zhijian/word3k", nil)
	rr := httptest.NewRecorder()
	word3kHandler()(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if strings.Count(rr.Body.String(), word3kPipe) < 10 {
		t.Fatalf("expected pooled corpus, pipes=%d", strings.Count(rr.Body.String(), word3kPipe))
	}
	if got := rr.Header().Get(word3kSourceHeader); got != word3kSourceEmbed {
		t.Fatalf("source header=%q want %q", got, word3kSourceEmbed)
	}
	lines := countNonEmptyLines(rr.Body.Bytes())
	if lines > word3kPoolSize() {
		t.Fatalf("response lines=%d want <= %d", lines, word3kPoolSize())
	}
}

func TestWord3kHandler_ResamplesPerRequest(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(docAssetsEngDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; i < 200; i++ {
		b.WriteString("word")
		b.WriteString(strings.Repeat("a", i%9))
		b.WriteString(" /wɜːd/ 词")
		b.WriteString(word3kPipe)
		b.WriteString("Def")
		b.WriteString(word3kPipe)
		b.WriteString("Example text.")
		b.WriteString(word3kPipe)
		b.WriteString("例")
		b.WriteByte('\n')
	}
	if err := os.WriteFile(docAssetsWord3kPath(dir), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	s := defaultOpsSnapshot()
	s.zhijianWord3kPoolSize = 30
	applyOpsSnapshot(s)
	initWord3kCache(dir)

	h := word3kHandler()
	bodyA := word3kGETBody(t, h)
	bodyB := word3kGETBody(t, h)
	if bodyA == bodyB {
		t.Fatal("expected different pools on consecutive requests (page refresh)")
	}
}

func TestLoadWord3kSource_External(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(docAssetsEngDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; i < 120; i++ {
		b.WriteString("word")
		b.WriteString(strings.Repeat("x", i%5))
		b.WriteString(" /wɜːd/ 词")
		b.WriteString(strings.Repeat("x", i%3))
		b.WriteString(word3kPipe)
		b.WriteString("Def line")
		b.WriteString(word3kPipe)
		b.WriteString("Example here.")
		b.WriteString(word3kPipe)
		b.WriteString("例句")
		b.WriteByte('\n')
	}
	if err := os.WriteFile(docAssetsWord3kPath(dir), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	src := loadWord3kSource(dir)
	if !src.external {
		t.Fatal("expected external source")
	}
	if len(src.lines) != 120 {
		t.Fatalf("source lines=%d want all 120 (< cap)", len(src.lines))
	}
}

func TestWord3kSourceCap_LimitsWorkset(t *testing.T) {
	if got := word3kSourceCap(80); got != 800 {
		t.Fatalf("cap=%d want 800", got)
	}
	if got := word3kSourceCap(500); got != word3kSourceMaxLines {
		t.Fatalf("cap=%d want max %d", got, word3kSourceMaxLines)
	}
}

func TestSampleWord3kSource_CapsAt1000(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 1500; i++ {
		b.WriteString("word")
		b.WriteString(strings.Repeat("a", i%7))
		b.WriteString(" /wɜːd/ 词")
		b.WriteString(word3kPipe)
		b.WriteString("Def")
		b.WriteString(word3kPipe)
		b.WriteString("Ex.")
		b.WriteString(word3kPipe)
		b.WriteString("例")
		b.WriteByte('\n')
	}
	lines, ok := sampleWord3kSourceFromMarkdown([]byte(b.String()), word3kSourceMaxLines, false)
	if !ok {
		t.Fatal("expected ok")
	}
	if len(lines) != word3kSourceMaxLines {
		t.Fatalf("workset len=%d want %d", len(lines), word3kSourceMaxLines)
	}
}

func TestParseOpsConfig_ZhijianWord3kPoolSize(t *testing.T) {
	s, err := parseOpsConfig(`
[zhijian]
word3k_pool_size = 120
`)
	if err != nil {
		t.Fatal(err)
	}
	if s.zhijianWord3kPoolSize != 120 {
		t.Fatalf("word3k_pool_size=%d", s.zhijianWord3kPoolSize)
	}
}

func word3kGETBody(t *testing.T, h http.HandlerFunc) string {
	t.Helper()
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/api/zhijian/word3k", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	return rr.Body.String()
}

func countNonEmptyLines(b []byte) int {
	n := 0
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}
