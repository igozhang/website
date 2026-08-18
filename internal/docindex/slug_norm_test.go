package docindex

import "testing"

func TestNormalizeSlug(t *testing.T) {
	cases := []struct {
		in  string
		out string
		ok  bool
	}{
		{"K8s-Nginx", "k8s_nginx", true},
		{"K8s", "k8s", true},
		{"  My  Slug  ", "my_slug", true},
		{"a__b---c", "a_b_c", true},
		{"-leading", "leading", true},
		{"9start", "", false},
		{"", "", false},
		{"valid_slug", "valid_slug", true},
	}
	for _, c := range cases {
		got, ok := NormalizeSlug(c.in)
		if ok != c.ok || got != c.out {
			t.Fatalf("NormalizeSlug(%q) = %q, %v; want %q, %v", c.in, got, ok, c.out, c.ok)
		}
	}
}
