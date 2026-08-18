package markdown

import (
	"strings"
	"testing"
)

func TestStripLeadingH1(t *testing.T) {
	src := []byte("# redis排错\n\nhello\n")
	got := string(StripLeadingH1(src))
	if strings.Contains(got, "redis排错") {
		t.Fatalf("h1 not stripped: %q", got)
	}
	if !strings.Contains(got, "hello") {
		t.Fatalf("body lost: %q", got)
	}
}

func TestStripLeadingH1_keepsH2(t *testing.T) {
	src := []byte("## section\n\nbody\n")
	got := string(StripLeadingH1(src))
	if !strings.HasPrefix(got, "## section") {
		t.Fatalf("h2 changed: %q", got)
	}
}

func TestResolveArticleTitle(t *testing.T) {
	md := []byte("# Redis 排错\n\nbody")
	if got := ResolveArticleTitle("", "redis_debug", md); got != "Redis 排错" {
		t.Fatalf("h1: got %q", got)
	}
	if got := ResolveArticleTitle("", "k8s_nginx", []byte("no h1")); got != "k8s nginx" {
		t.Fatalf("slug: got %q", got)
	}
}
