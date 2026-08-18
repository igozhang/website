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

func TestIgoAdminArticlesList_filtered(t *testing.T) {
	setupIgoAdminTestToken(t)
	dir := t.TempDir()
	assets := setupIgoAdminAssets(t, dir)
	st := openIgoAdminStore(t, assets)
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()
	for _, m := range []docindex.Meta{
		{Slug: "batch_a", Title: "A", Categories: "ops", Tags: "redis", PublishedAt: time.Now()},
		{Slug: "batch_b", Title: "B", Categories: "k8s", Tags: "go", PublishedAt: time.Now()},
	} {
		if err := st.WriteArticle(m, []byte("# "+m.Title)); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	req := httptest.NewRequest(http.MethodGet, "/api/igoadmin/articles?category=ops&limit=10&offset=0", nil)
	req.Header.Set("X-Igo-Token", testIgoAdminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []docindex.Meta `json:"items"`
		Total int             `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].Slug != "batch_a" {
		t.Fatalf("ops filter: total=%d items=%+v", resp.Total, resp.Items)
	}
}

func TestIgoAdminArticlesBatch_PATCH(t *testing.T) {
	setupIgoAdminTestToken(t)
	dir := t.TempDir()
	assets := setupIgoAdminAssets(t, dir)
	st := openIgoAdminStore(t, assets)
	defer st.Close()
	docStore = st
	defer func() { docStore = nil }()
	slug := "batch_patch"
	if err := st.WriteArticle(docindex.Meta{
		Slug: slug, Title: "T", Categories: "old", Tags: "x", PublishedAt: time.Now(),
	}, []byte("# T")); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	registerIgoAdminHandlers(mux, dir)

	body, _ := json.Marshal(map[string]any{
		"slugs":            []string{slug},
		"updateCategories": true,
		"setCategories":    "newcat",
		"updateTags":       true,
		"setTags":          "y",
		"tagsMode":         "append",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/igoadmin/articles/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Igo-Token", testIgoAdminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH status=%d %s", rec.Code, rec.Body.String())
	}
	got, err := st.Get(slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Categories != "newcat" || got.Tags != "x, y" {
		t.Fatalf("meta: cat=%q tags=%q", got.Categories, got.Tags)
	}
	raw, _ := st.ReadBody(slug)
	if string(raw) != "# T" {
		t.Fatalf("body changed: %q", string(raw))
	}
}

func setupIgoAdminTestToken(t *testing.T) {
	t.Helper()
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	testIgoAdminToken = "batch-admin-tok"
	authSessionMu.Lock()
	authSessions[testIgoAdminToken] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()
}

var testIgoAdminToken string

func setupIgoAdminAssets(t *testing.T, dir string) string {
	t.Helper()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	return assets
}

func openIgoAdminStore(t *testing.T, assets string) *docindex.Store {
	t.Helper()
	st, err := docindex.Open(filepath.Join(filepath.Dir(assets), "doc.db"), assets)
	if err != nil {
		t.Fatal(err)
	}
	return st
}
