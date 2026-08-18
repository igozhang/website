package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestIgoAdminUploadMD_metaAndValidation(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "upload-md-test"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()

	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("slug", "upload_meta")
	_ = w.WriteField("categories", "devops")
	_ = w.WriteField("tags", "a,b")
	_ = w.WriteField("author", "writer")
	part, err := w.CreateFormFile("file", "upload_meta.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(part, "# Upload\n\nok\n"); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/igoadmin/upload-md", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-Igo-Token", tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", rec.Code, rec.Body.String())
	}

	got, err := st.Get("upload_meta")
	if err != nil {
		t.Fatal(err)
	}
	if got.Categories != "devops" || got.Tags != "a,b" || got.Author != "writer" {
		t.Fatalf("meta=%+v", got)
	}

	var bad bytes.Buffer
	wb := multipart.NewWriter(&bad)
	_ = wb.WriteField("slug", "bad_fence")
	part, err = wb.CreateFormFile("file", "bad_fence.md")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(part, "```bash\necho\n")
	_ = wb.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/igoadmin/upload-md", &bad)
	req.Header.Set("Content-Type", wb.FormDataContentType())
	req.Header.Set("X-Igo-Token", tok)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid md status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "markdown validation failed" {
		t.Fatalf("error=%v", resp["error"])
	}
}

func TestIgoAdminArticlesList_searchLimit(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "articles-search-test"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	assets := filepath.Join(dir, "assets")
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	now := time.Now()
	for _, slug := range []string{"ts", "linux_ts", "docker_ts"} {
		if err := st.WriteArticle(docindex.Meta{
			Slug: slug, Title: slug, PublishedAt: now,
		}, []byte("# "+slug+"\n")); err != nil {
			t.Fatal(err)
		}
	}
	docStore = st
	defer func() { docStore = nil }()

	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	req := httptest.NewRequest(http.MethodGet, "/api/igoadmin/articles?q=ts&limit=50", nil)
	req.Header.Set("X-Igo-Token", tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []docindex.Meta `json:"items"`
		Total int             `json:"total"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total < 3 || len(resp.Items) < 3 {
		t.Fatalf("want >=3 hits for ts, got total=%d items=%d", resp.Total, len(resp.Items))
	}
	foundTS := false
	for _, m := range resp.Items {
		if m.Slug == "ts" {
			foundTS = true
			break
		}
	}
	if !foundTS {
		t.Fatalf("slug ts missing from search results: %+v", resp.Items)
	}
}
