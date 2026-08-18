package docindex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidSlug(t *testing.T) {
	if !ValidSlug("k8s_nginx") || ValidSlug("K8s") || ValidSlug("") || ValidSlug("a/b") {
		t.Fatal("slug validation")
	}
}

func TestReconcileFromRepoAssets(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	if _, err := os.Stat(root); err != nil {
		t.Skip("no repo assets")
	}
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "doc.db")
	assets := filepath.Join(dir, "assets")
	if err := copyDir(root, assets); err != nil {
		t.Fatal(err)
	}
	st, err := Open(dbPath, assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Reconcile(); err != nil {
		t.Fatal(err)
	}
	n, err := st.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(n) < 200 {
		t.Fatalf("expected many articles, got %d", len(n))
	}
	hits, err := st.Search("redis", 5)
	if err != nil || len(hits) == 0 {
		t.Fatalf("search redis: %v len=%d", err, len(hits))
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, info.Mode())
	})
}
