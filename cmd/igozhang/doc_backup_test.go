package main

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestDocBackup_writeCache(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()

	m := docindex.Meta{Slug: "cache_test", Title: "Cache Test", Author: "a", PublishedAt: time.Now()}
	if err := st.WriteArticle(m, []byte("# Cache Test\n\nbody")); err != nil {
		t.Fatal(err)
	}
	if err := writeDocExportCache(dir); err != nil {
		t.Fatal(err)
	}
	raw, ok := readDocExportCache(dir)
	if !ok || len(raw) == 0 {
		t.Fatal("expected cached zip")
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	var hasMeta bool
	for _, f := range zr.File {
		if f.Name == "meta/cache_test.json" {
			hasMeta = true
		}
	}
	if !hasMeta {
		t.Fatal("cache zip missing meta/cache_test.json")
	}
}

func TestScheduleDocBackup_debounced(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()

	docBackupMu.Lock()
	if docBackupTimer != nil {
		docBackupTimer.Stop()
		docBackupTimer = nil
	}
	docBackupMu.Unlock()
	docExportBuildEnd(dir)

	if err := writeDocExportCache(dir); err != nil {
		t.Fatal(err)
	}
	rawBefore, ok := readDocExportCache(dir)
	if !ok || len(rawBefore) == 0 {
		t.Fatal("expected cache before debounced backup")
	}
	scheduleDocBackup(dir)
	rawMid, ok := readDocExportCache(dir)
	if !ok || len(rawMid) == 0 {
		t.Fatal("scheduleDocBackup must not delete existing cache before rebuild")
	}
	runDocBackup()
	if !waitForDocExportCache(dir, 30*time.Second) {
		t.Fatal("timeout waiting for debounced backup")
	}
	raw, ok := readDocExportCache(dir)
	if !ok || len(raw) == 0 {
		t.Fatal("expected backup after runDocBackup")
	}
}

func TestTryServeDocExportCache_streams(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()
	if err := st.WriteArticle(docindex.Meta{Slug: "s", Title: "S", Author: "a", PublishedAt: time.Now()}, []byte("# S\n")); err != nil {
		t.Fatal(err)
	}
	if err := writeDocExportCache(dir); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/igoadmin/export", nil)
	rec := httptest.NewRecorder()
	if !tryServeDocExportCache(rec, req, dir) {
		t.Fatal("expected cache hit")
	}
	if got := rec.Header().Get("X-Doc-Export-Cache"); got != "hit-mem" {
		t.Fatalf("cache header: %q want hit-mem", got)
	}
	if cl := rec.Header().Get("Content-Length"); cl == "" {
		t.Fatal("missing Content-Length on mem hit")
	}
	cd := rec.Header().Get("Content-Disposition")
	if cd == "" || len(cd) < 20 {
		t.Fatalf("Content-Disposition: %q", cd)
	}
	if len(rec.Body.Bytes()) == 0 {
		t.Fatal("empty body")
	}
}

func TestDocExportMemCache_afterWrite(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()
	if err := st.WriteArticle(docindex.Meta{Slug: "m", Title: "M", Author: "a", PublishedAt: time.Now()}, []byte("# M\n")); err != nil {
		t.Fatal(err)
	}
	if err := writeDocExportCache(dir); err != nil {
		t.Fatal(err)
	}
	raw, ok := getDocExportMemCache(dir)
	if !ok || len(raw) == 0 {
		t.Fatal("expected in-memory cache after writeDocExportCache")
	}
	invalidateDocExportCache(dir)
	if _, ok := getDocExportMemCache(dir); ok {
		t.Fatal("mem cache should clear on invalidate")
	}
}

func TestDocExportBuild_singleflight(t *testing.T) {
	dir := t.TempDir()
	if docExportBuildTryStart(dir) != true {
		t.Fatal("expected first start")
	}
	if docExportBuildTryStart(dir) {
		t.Fatal("expected second start blocked")
	}
	docExportBuildEnd(dir)
	if !docExportBuildTryStart(dir) {
		t.Fatal("expected start after end")
	}
	docExportBuildEnd(dir)
}

func TestWriteDocExportCache_skipsLogPrepareOnFastPath(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()
	if err := st.WriteArticle(docindex.Meta{Slug: "fast", Title: "Fast", Author: "a", PublishedAt: time.Now()}, []byte("# Fast\n")); err != nil {
		t.Fatal(err)
	}
	if err := writeDocExportCacheWithOpts(dir, false); err != nil {
		t.Fatal(err)
	}
	if !docExportCacheFileReady(dir) {
		t.Fatal("expected cache file")
	}
}
