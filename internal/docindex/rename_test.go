package docindex

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRenameArticle(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	st, err := Open(filepath.Join(dir, "doc.db"), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	oldSlug, newSlug := "rename_old", "rename_new"
	body := []byte("# T\n\nbody")
	if err := st.WriteArticle(Meta{
		Slug: oldSlug, Title: "T", Tags: "a", Views: 7,
		PublishedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}, body); err != nil {
		t.Fatal(err)
	}
	if err := st.RenameArticle(oldSlug, newSlug); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Get(oldSlug); err == nil {
		t.Fatal("old slug should be gone")
	}
	got, err := st.Get(newSlug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "T" || got.Tags != "a" || got.Views != 7 {
		t.Fatalf("meta=%+v", got)
	}
	raw, err := st.ReadBody(newSlug)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(body) {
		t.Fatalf("body=%q", raw)
	}
	if _, err := os.Stat(filepath.Join(assets, "md", oldSlug+".md")); !os.IsNotExist(err) {
		t.Fatal("old md file should be removed")
	}
	hits, err := st.Search("rename_new", 5)
	if err != nil || len(hits) == 0 {
		t.Fatalf("fts search: %v len=%d", err, len(hits))
	}
}

func TestRenameArticle_conflict(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	st, err := Open(filepath.Join(dir, "doc.db"), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	now := time.Now()
	for _, slug := range []string{"a_slug", "b_slug"} {
		if err := st.WriteArticle(Meta{Slug: slug, Title: slug, PublishedAt: now}, []byte("# x\n")); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.RenameArticle("a_slug", "b_slug"); err == nil {
		t.Fatal("expected conflict")
	}
}
