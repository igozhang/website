package docindex

import (
	"path/filepath"
	"testing"
	"time"
)

func TestBackfillDefaultAuthor(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "doc.db")
	assets := filepath.Join(dir, "assets")
	st, err := Open(dbPath, assets)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.db.Exec(`
INSERT INTO articles(slug,title,categories,tags,author,published_at,views,md_mtime,updated_at)
VALUES('blank','Blank','','','',?,0,0,0)`, time.Now().Unix()); err != nil {
		st.Close()
		t.Fatal(err)
	}
	st.Close()

	st, err = Open(dbPath, assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	m, err := st.Get("blank")
	if err != nil {
		t.Fatal(err)
	}
	if m.Author != DefaultAuthor {
		t.Fatalf("author=%q want %q", m.Author, DefaultAuthor)
	}
}
