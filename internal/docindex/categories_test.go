package docindex

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEscapeLikePattern(t *testing.T) {
	if got := escapeLikePattern(`a%b_c\d`); got != `a\%b\_c\\d` {
		t.Fatalf("escapeLikePattern: got %q", got)
	}
}

func TestPrimaryCategory(t *testing.T) {
	if got := PrimaryCategory("devops, k8s"); got != "devops" {
		t.Fatalf("PrimaryCategory: got %q", got)
	}
	if got := PrimaryCategory(""); got != "" {
		t.Fatalf("PrimaryCategory empty: got %q", got)
	}
}

func TestListCategoriesAndByCategory(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := Open(filepath.Join(dir, "doc.db"), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	now := time.Now()
	articles := []Meta{
		{Slug: "k8s_a", Title: "K8s A", Categories: "k8s, Git", Author: "a", PublishedAt: now},
		{Slug: "k8s_b", Title: "K8s B", Categories: "k8s", Author: "a", PublishedAt: now},
		{Slug: "devops_x", Title: "DevOps X", Categories: "devops", Author: "a", PublishedAt: now},
	}
	for _, m := range articles {
		if err := st.WriteArticle(m, []byte("# "+m.Title+"\n")); err != nil {
			t.Fatal(err)
		}
	}
	cats, err := st.ListCategories()
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) < 3 {
		t.Fatalf("expected categories, got %v", cats)
	}
	hits, err := st.ListByCategory("k8s", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected k8s articles")
	}
	for _, m := range hits {
		found := false
		for _, c := range SplitCategories(m.Categories) {
			if c == "k8s" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("slug %s categories %q not k8s", m.Slug, m.Categories)
		}
	}
}

func TestRenameAndRemoveCategory(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := Open(filepath.Join(dir, "doc.db"), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	m := Meta{Slug: "cat_test", Title: "t", Categories: "foo, bar", Author: "a", PublishedAt: time.Now()}
	if err := st.WriteArticle(m, []byte("# t\n")); err != nil {
		t.Fatal(err)
	}
	if err := st.RenameCategory("foo", "baz"); err != nil {
		t.Fatal(err)
	}
	got, err := st.Get("cat_test")
	if err != nil {
		t.Fatal(err)
	}
	if got.Categories != "baz, bar" {
		t.Fatalf("rename: got %q", got.Categories)
	}
	if err := st.RemoveCategory("bar"); err != nil {
		t.Fatal(err)
	}
	got, err = st.Get("cat_test")
	if err != nil {
		t.Fatal(err)
	}
	if got.Categories != "baz" {
		t.Fatalf("remove: got %q", got.Categories)
	}
}
