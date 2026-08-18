package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestRecordDocView_dedupWithin12h(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	slug := "view_test"
	_ = os.WriteFile(filepath.Join(assets, "md", slug+".md"), []byte("# t\n\nbody"), 0o644)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Reconcile(); err != nil {
		t.Fatal(err)
	}
	docStore = st
	defer func() { docStore = nil }()

	req := httptest.NewRequest(http.MethodGet, "/"+slug, nil)
	rec := httptest.NewRecorder()
	v1, err := recordDocView(rec, req, slug)
	if err != nil {
		t.Fatal(err)
	}
	if v1 != 1 {
		t.Fatalf("first view want 1, got %d", v1)
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("expected view cookie")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/"+slug, nil)
	for _, c := range rec.Result().Cookies() {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	v2, err := recordDocView(rec2, req2, slug)
	if err != nil {
		t.Fatal(err)
	}
	if v2 != 1 {
		t.Fatalf("second view same cookie want 1, got %d", v2)
	}
}
