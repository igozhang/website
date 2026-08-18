package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewHandler_txtUsesDocCodeChrome(t *testing.T) {
	dir := t.TempDir()
	fsDir := filepath.Join(dir, "fs")
	if err := os.MkdirAll(fsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	txt := filepath.Join(fsDir, "notes.txt")
	if err := os.WriteFile(txt, []byte("echo hello\n# comment\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/preview/", previewHandler(dir))

	req := httptest.NewRequest(http.MethodGet, "/preview/fs/notes.txt", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`class="doc-md"`,
		`<pre><code>`,
		`echo hello`,
		`/assets/doc/code-chrome.js`,
		`.doc-code-wrap`,
		`.doc-copy-btn`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("preview html missing %q", want)
		}
	}
	if strings.Contains(body, `preview-txt`) {
		t.Fatal("txt preview should not use legacy preview-txt class")
	}
}

func TestPreviewHandler_mdCodeFenceAndMermaid(t *testing.T) {
	dir := t.TempDir()
	fsDir := filepath.Join(dir, "fs")
	if err := os.MkdirAll(fsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := filepath.Join(fsDir, "demo.md")
	content := "# Keep Title\n\n```go\npackage main\n```\n\n```mermaid\ngraph TD; A-->B;\n```\n"
	if err := os.WriteFile(md, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/preview/", previewHandler(dir))

	req := httptest.NewRequest(http.MethodGet, "/preview/fs/demo.md", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`class="doc-md"`,
		`Keep Title`,
		`language-go`,
		`language-mermaid`,
		`/assets/doc/code-chrome.js`,
		`/assets/doc/mermaid`,
		`.doc-code-wrap`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("md preview missing %q\n%s", want, body[:min(500, len(body))])
		}
	}
}

func TestPreviewHandler_yamlCodeLang(t *testing.T) {
	dir := t.TempDir()
	fsDir := filepath.Join(dir, "fs")
	if err := os.MkdirAll(fsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fsDir, "app.yaml")
	if err := os.WriteFile(path, []byte("name: demo\nenabled: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/preview/", previewHandler(dir))
	req := httptest.NewRequest(http.MethodGet, "/preview/fs/app.yaml", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`class="doc-md"`,
		`language-yaml`,
		`name: demo`,
		`/assets/doc/code-chrome.js`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("yaml preview missing %q", want)
		}
	}
}

func TestPreviewHandler_rejectsBinary(t *testing.T) {
	dir := t.TempDir()
	fsDir := filepath.Join(dir, "fs")
	if err := os.MkdirAll(fsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fsDir, "fake.txt")
	if err := os.WriteFile(path, []byte("a\x00b"), 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/preview/", previewHandler(dir))
	req := httptest.NewRequest(http.MethodGet, "/preview/fs/fake.txt", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("want 415 got %d", rec.Code)
	}
}

func TestPreviewHandler_rejectsZip(t *testing.T) {
	dir := t.TempDir()
	fsDir := filepath.Join(dir, "fs")
	if err := os.MkdirAll(fsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fsDir, "a.zip")
	if err := os.WriteFile(path, []byte("PK"), 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/preview/", previewHandler(dir))
	req := httptest.NewRequest(http.MethodGet, "/preview/fs/a.zip", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rec.Code)
	}
}

func TestFilePreviewPageCSS_includesCodeChrome(t *testing.T) {
	css := filePreviewPageCSS()
	for _, want := range []string{`.doc-code-wrap`, `.doc-copy-btn`, `--out-bg:#0d1117`} {
		if !strings.Contains(css, want) {
			t.Fatalf("filePreviewPageCSS missing %q", want)
		}
	}
	if strings.Contains(css, `pre.preview-txt`) {
		t.Fatal("legacy preview-txt css should be removed")
	}
}
