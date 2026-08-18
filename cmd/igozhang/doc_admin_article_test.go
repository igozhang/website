package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestIgoAdminArticlePUT_updatesMeta(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "put-article-test"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	slug := "edit_put_test"
	_ = os.WriteFile(filepath.Join(assets, "md", slug+".md"), []byte("# Old\n\nbody"), 0o644)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	pubAt := time.Date(2024, 3, 15, 10, 0, 0, 0, time.Local)
	if err := st.WriteArticle(docindex.Meta{
		Slug: slug, Title: "Old", Categories: "devops", Tags: "a",
		Author: "alice", PublishedAt: pubAt,
	}, []byte("# Old\n\nbody")); err != nil {
		t.Fatal(err)
	}
	docStore = st
	defer func() { docStore = nil }()

	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	body, _ := json.Marshal(map[string]any{
		"slug":        slug,
		"title":       "New Title",
		"categories":  "k8s",
		"tags":        "b,c",
		"publishedAt": "2024-06-01 08:30:00",
		"markdown":    "# New Title\n\nupdated body",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/igoadmin/article", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Igo-Token", tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true {
		t.Fatalf("ok=%v", resp["ok"])
	}

	got, err := st.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "New Title" {
		t.Fatalf("title=%q", got.Title)
	}
	if got.Categories != "k8s" {
		t.Fatalf("categories=%q", got.Categories)
	}
	if got.Tags != "b,c" {
		t.Fatalf("tags=%q", got.Tags)
	}
	if got.Author != "alice" {
		t.Fatalf("author=%q want alice preserved", got.Author)
	}
	wantPub := time.Date(2024, 6, 1, 8, 30, 0, 0, time.Local)
	if !got.PublishedAt.Equal(wantPub) {
		t.Fatalf("publishedAt=%v want %v", got.PublishedAt, wantPub)
	}
	raw, err := st.ReadBody(slug)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "# New Title\n\nupdated body" {
		t.Fatalf("body=%q", string(raw))
	}
}

func TestIgoAdminArticlePUT_renamesSlug(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "rename-article-test"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	oldSlug := "rename_put_old"
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: oldSlug, Title: "Keep", Tags: "x", Views: 3, PublishedAt: time.Now(),
	}, []byte("# Keep\n\nv1")); err != nil {
		t.Fatal(err)
	}
	docStore = st
	defer func() { docStore = nil }()

	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	newSlug := "rename_put_new"
	body, _ := json.Marshal(map[string]any{
		"slug":       newSlug,
		"oldSlug":    oldSlug,
		"title":      "Keep",
		"tags":       "x",
		"markdown":   "# Keep\n\nv2",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/igoadmin/article", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Igo-Token", tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, err := st.Get(oldSlug); err == nil {
		t.Fatal("old slug should be gone")
	}
	got, err := st.Get(newSlug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Views != 3 {
		t.Fatalf("views=%d want 3", got.Views)
	}
	raw, err := st.ReadBody(newSlug)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "# Keep\n\nv2" {
		t.Fatalf("body=%q", string(raw))
	}
}

func TestIgoAdminArticlePUT_renameBlockedByValidation(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "rename-validate-test"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	oldSlug := "rename_validate_old"
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{Slug: oldSlug, Title: "T", PublishedAt: time.Now()}, []byte("# T\n")); err != nil {
		t.Fatal(err)
	}
	docStore = st
	defer func() { docStore = nil }()

	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	body, _ := json.Marshal(map[string]any{
		"slug":     "rename_validate_new",
		"oldSlug":  oldSlug,
		"markdown": "```bash\necho\n",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/igoadmin/article", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Igo-Token", tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want 400 body=%s", rec.Code, rec.Body.String())
	}
	if _, err := st.Get(oldSlug); err != nil {
		t.Fatal("old slug should remain")
	}
	if ok, _ := st.Exists("rename_validate_new"); ok {
		t.Fatal("new slug should not be created")
	}
}

func TestIgoAdminArticlePUT_validationBlocksUnclosedFence(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "validate-put-test"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	slug := "validate_put"
	_ = os.WriteFile(filepath.Join(assets, "md", slug+".md"), []byte("# T\n"), 0o644)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{Slug: slug, Title: "T", PublishedAt: time.Now()}, []byte("# T\n")); err != nil {
		t.Fatal(err)
	}
	docStore = st
	defer func() { docStore = nil }()

	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	body, _ := json.Marshal(map[string]any{
		"slug": slug, "markdown": "```bash\necho\n",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/igoadmin/article", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Igo-Token", tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestParsePublishedAtValue_dateOnlyFillsCurrentClock(t *testing.T) {
	before := time.Now()
	got, ok := parsePublishedAtValue("2018-06-23")
	after := time.Now()
	if !ok {
		t.Fatal("expected ok")
	}
	if y, m, d := got.Date(); y != 2018 || m != time.June || d != 23 {
		t.Fatalf("date part=%v", got)
	}
	gotTod := got.Hour()*3600 + got.Minute()*60 + got.Second()
	beforeTod := before.Hour()*3600 + before.Minute()*60 + before.Second()
	afterTod := after.Hour()*3600 + after.Minute()*60 + after.Second()
	if before.Day() == after.Day() && (gotTod < beforeTod || gotTod > afterTod) {
		t.Fatalf("clock=%02d:%02d:%02d not in [%v, %v]", got.Hour(), got.Minute(), got.Second(), before.Format("15:04:05"), after.Format("15:04:05"))
	}
	full, ok := parsePublishedAtValue("2024-06-01 08:30:00")
	if !ok {
		t.Fatal("expected full datetime ok")
	}
	want := time.Date(2024, 6, 1, 8, 30, 0, 0, time.Local)
	if !full.Equal(want) {
		t.Fatalf("full=%v want %v", full, want)
	}
	if _, ok := parsePublishedAtValue(""); ok {
		t.Fatal("empty should be false")
	}
}
