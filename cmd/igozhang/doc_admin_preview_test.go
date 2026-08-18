package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestIgoAdminPreview_stripsLeadingH1(t *testing.T) {
	setupIgoAdminTestToken(t)
	dir := t.TempDir()
	assets := setupIgoAdminAssets(t, dir)
	st := openIgoAdminStore(t, assets)
	defer st.Close()

	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	body, _ := json.Marshal(map[string]string{
		"markdown": "# Page Title\n\nHello **world**",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/igoadmin/preview", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Igo-Token", testIgoAdminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		HTML string `json:"html"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(resp.HTML, "Page Title") {
		t.Fatalf("preview should strip leading h1, got %q", resp.HTML)
	}
	if !strings.Contains(resp.HTML, "world") {
		t.Fatalf("preview body missing: %q", resp.HTML)
	}
}

func TestIgoAdminArticlePOST_validationBlocks(t *testing.T) {
	setupIgoAdminTestToken(t)
	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
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

	body, _ := json.Marshal(map[string]any{
		"slug": "new_bad", "markdown": "```\nopen\n",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/igoadmin/article", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Igo-Token", testIgoAdminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d %s", rec.Code, rec.Body.String())
	}
	_ = time.Now()
}
