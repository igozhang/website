package docindex

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSearchIncludesSlug(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "doc.db")
	assets := filepath.Join(dir, "assets")
	st, err := Open(dbPath, assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now()
	articles := []Meta{
		{Slug: "ts", Title: "速查百宝箱", Categories: "devops", PublishedAt: now.Add(-time.Hour)},
		{Slug: "linux_ts", Title: "linux_ts,排错", Categories: "devops", PublishedAt: now},
	}
	for _, m := range articles {
		if err := st.WriteArticle(m, []byte("# "+m.Title+"\n")); err != nil {
			t.Fatal(err)
		}
	}

	hits, err := st.Search("ts", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 2 {
		t.Fatalf("expected at least 2 hits, got %d", len(hits))
	}
	if hits[0].Slug != "ts" {
		t.Fatalf("exact slug match should rank first, got %q", hits[0].Slug)
	}
}

func TestMigrateFTSSlugIndexed(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "doc.db")
	assets := filepath.Join(dir, "assets")
	st, err := Open(dbPath, assets)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.WriteArticle(Meta{Slug: "demo", Title: "Demo", PublishedAt: time.Now()}, []byte("# Demo\n")); err != nil {
		st.Close()
		t.Fatal(err)
	}
	if _, err := st.db.Exec(`
DROP TABLE articles_fts;
CREATE VIRTUAL TABLE articles_fts USING fts5(
  slug UNINDEXED,
  title,
  tags,
  tokenize='unicode61 remove_diacritics 0'
);
INSERT INTO articles_fts(slug, title, tags) SELECT slug, title, tags FROM articles;`); err != nil {
		st.Close()
		t.Fatal(err)
	}
	st.Close()

	st, err = Open(dbPath, assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	var ddl string
	if err := st.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='articles_fts'`).Scan(&ddl); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToUpper(ddl), "UNINDEXED") {
		t.Fatalf("expected slug indexed FTS, got %q", ddl)
	}

	hits, err := st.Search("demo", 5)
	if err != nil || len(hits) != 1 || hits[0].Slug != "demo" {
		t.Fatalf("search after migration: hits=%v err=%v", hits, err)
	}
}
