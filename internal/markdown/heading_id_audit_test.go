package markdown

import (
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestHeadingID_asciiParityAndUnicode(t *testing.T) {
	cases := []struct{ title, want string }{
		{"Hello World", "hello-world"},
		{"foo_bar", "foo-bar"},
		{"A--B", "a--b"},
		{"Role / ClusterRole", "role--clusterrole"},
		{"PersistentVolume (PV)", "persistentvolume-pv"},
		{"What's up?", "whats-up"},
		{"3.14 Pi", "314-pi"},
		{"", "heading"},
		{"!!!", "heading"},
		{"中文", "中文"},
		{"日本語の見出し", "日本語の見出し"},
		{"Привет", "привет"},
		{"emoji 😄 test", "emoji--test"},
		{"CPU/内存 requests 与 limits", "cpu内存-requests-与-limits"},
		{"命名空间配额（ResourceQuota / LimitRange）", "命名空间配额-resourcequota--limitrange"},
	}
	for _, tc := range cases {
		ids := newUnicodeHeadingIDs()
		got := string(ids.Generate([]byte(tc.title), ast.KindHeading))
		if got != tc.want {
			t.Fatalf("%q → %q want %q", tc.title, got, tc.want)
		}
	}
}

func TestRender_duplicateChineseHeadings(t *testing.T) {
	html, err := Render([]byte("## 甲\n\n## 乙\n\n## 甲\n"), RenderOpts{})
	if err != nil {
		t.Fatal(err)
	}
	out := string(html)
	for _, id := range []string{`id="甲"`, `id="乙"`, `id="甲-1"`} {
		if !strings.Contains(out, id) {
			t.Fatalf("missing %s\n%s", id, out)
		}
	}
}

func TestK8sArticle_allNavAnchorsMatch(t *testing.T) {
	path := "/Users/igozhang/Documents/cursor/0719-write-mds/k8s-resources-igo.md"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skip("sample article not present:", err)
	}
	html, err := Render(StripLeadingH1(raw), RenderOpts{})
	if err != nil {
		t.Fatal(err)
	}
	out := string(html)
	reH := regexp.MustCompile(`id="([^"]+)"`)
	ids := map[string]bool{}
	for _, m := range reH.FindAllStringSubmatch(out, -1) {
		ids[m[1]] = true
	}
	if ids["heading"] || ids["heading-1"] {
		t.Fatal("generic heading ids still used for Chinese titles")
	}
	reA := regexp.MustCompile(`href="#([^"]+)"`)
	seen := map[string]bool{}
	var miss []string
	okn := 0
	for _, m := range reA.FindAllStringSubmatch(out, -1) {
		id, err := url.PathUnescape(m[1])
		if err != nil {
			id = m[1]
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		if !ids[id] {
			miss = append(miss, id)
		} else {
			okn++
		}
	}
	if len(miss) > 0 {
		t.Fatalf("anchors miss=%v (ok=%d)", miss, okn)
	}
	if okn < 30 {
		t.Fatalf("expected >=30 unique nav anchors, got %d", okn)
	}
}
