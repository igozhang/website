package docindex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSlug_prefix(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	for _, slug := range []string{"squid_ins", "redis_cfg"} {
		_ = os.WriteFile(filepath.Join(assets, "md", slug+".md"), []byte("# x\n"), 0o644)
	}
	st, err := Open(filepath.Join(assets, "doc.db"), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Reconcile(); err != nil {
		t.Fatal(err)
	}
	got, ok := st.ResolveSlug("squid")
	if !ok || got != "squid_ins" {
		t.Fatalf("squid -> %q ok=%v", got, ok)
	}
	if _, ok := st.ResolveSlug("sq"); ok {
		t.Fatal("too short query should not fuzzy match")
	}
}
