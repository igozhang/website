package docindex

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMergeTagsField_and_Remove(t *testing.T) {
	if got := MergeTagsField("a, b", "b, c"); got != "a, b, c" {
		t.Fatalf("merge=%q", got)
	}
	if got := RemoveTagsField("a, b, c", "b"); got != "a, c" {
		t.Fatalf("remove=%q", got)
	}
	if got := ApplyTagsPatch("x", "y", "replace"); got != "y" {
		t.Fatalf("replace=%q", got)
	}
	if got := ApplyTagsPatch("a", "b", "append"); got != "a, b" {
		t.Fatalf("append=%q", got)
	}
}

func TestBatchPatchMeta_updatesDB(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	st, err := Open(filepath.Join(dir, "doc.db"), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	for _, m := range []Meta{
		{Slug: "a1", Title: "A1", Categories: "devops", Tags: "go", PublishedAt: mustTime()},
		{Slug: "a2", Title: "A2", Categories: "k8s", Tags: "k8s", PublishedAt: mustTime()},
	} {
		if err := st.WriteArticle(m, []byte("# "+m.Title)); err != nil {
			t.Fatal(err)
		}
	}
	cat := "sre"
	ok, failed := st.BatchPatchMeta([]string{"a1", "a2", "missing"}, PatchMetaOpts{
		Categories: &cat,
		Tags:       strPtr("newtag"),
		TagsMode:   "append",
	})
	if ok != 2 || len(failed) != 1 {
		t.Fatalf("ok=%d failed=%v", ok, failed)
	}
	m1, _ := st.Get("a1")
	if m1.Categories != "sre" || m1.Tags != "go, newtag" {
		t.Fatalf("a1: cat=%q tags=%q", m1.Categories, m1.Tags)
	}
	hits, err := st.Search("newtag", 5)
	if err != nil || len(hits) < 2 {
		t.Fatalf("fts search newtag: %v %d", err, len(hits))
	}
}

func strPtr(s string) *string { return &s }

func mustTime() time.Time { return time.Now() }
