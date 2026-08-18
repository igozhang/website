package markdown

import (
	"strings"
	"testing"
)

func TestRewriteFragmentHrefs(t *testing.T) {
	in := []byte(`<p><a href="#pod">Pod</a> <a href="#页内导航">nav</a> <a href="/other">x</a> <a href="#">bare</a></p>`)
	got := string(RewriteFragmentHrefs(in, "/k8s_crud"))
	for _, want := range []string{
		`href="/k8s_crud#pod"`,
		`href="/k8s_crud#页内导航"`,
		`href="/other"`,
		`href="#"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s in %s", want, got)
		}
	}
}

func TestRewriteFragmentHrefs_withBasePath(t *testing.T) {
	in := []byte(`<a href="#cpu内存-requests-与-limits">x</a>`)
	got := string(RewriteFragmentHrefs(in, "/fs/k8s_crud"))
	if !strings.Contains(got, `href="/fs/k8s_crud#cpu内存-requests-与-limits"`) {
		t.Fatalf("got %s", got)
	}
}

func TestRewriteFragmentHrefs_percentEncoded(t *testing.T) {
	in := []byte(`<a href="#%E9%A1%B5%E5%86%85%E5%AF%BC%E8%88%AA">nav</a>`)
	got := string(RewriteFragmentHrefs(in, "/k8s_crud"))
	want := `href="/k8s_crud#%E9%A1%B5%E5%86%85%E5%AF%BC%E8%88%AA"`
	if !strings.Contains(got, want) {
		t.Fatalf("got %s", got)
	}
}

func TestRewriteFragmentHrefs_noop(t *testing.T) {
	in := []byte(`<p>no anchors</p>`)
	if string(RewriteFragmentHrefs(in, "/x")) != string(in) {
		t.Fatal("should leave html unchanged")
	}
	if string(RewriteFragmentHrefs([]byte(`<a href="#a">`), "")) != `<a href="#a">` {
		t.Fatal("empty pagePath should not rewrite")
	}
}
