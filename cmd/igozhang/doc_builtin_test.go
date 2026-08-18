package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestInstallBuiltinDemoArticle_emptyMD(t *testing.T) {
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
	if err := installBuiltinDemoArticle(st, assets); err != nil {
		t.Fatal(err)
	}
	got, err := st.Get(builtinDemoSlug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title == "" {
		t.Fatal("expected title")
	}
	body, err := st.ReadBody(builtinDemoSlug)
	if err != nil || len(body) == 0 {
		t.Fatalf("body: %v len=%d", err, len(body))
	}
}

func TestDocExportDownloadName_hasMinute(t *testing.T) {
	name := docExportDownloadName()
	if len(name) < len("doc-export-20060102-1504.zip") {
		t.Fatalf("unexpected name %q", name)
	}
	if name[:10] != "doc-export" {
		t.Fatalf("prefix: %q", name)
	}
}
