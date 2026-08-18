package markdown

import (
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestUnicodeHeadingIDs(t *testing.T) {
	ids := newUnicodeHeadingIDs()
	cases := []struct {
		title string
		want  string
	}{
		{"Pod", "pod"},
		{"Role / ClusterRole", "role--clusterrole"},
		{"PersistentVolume (PV)", "persistentvolume-pv"},
		{"页内导航", "页内导航"},
		{"通用操作速查", "通用操作速查"},
		{"CPU/内存 requests 与 limits", "cpu内存-requests-与-limits"},
		{"HPA 自动扩缩", "hpa-自动扩缩"},
		{"命名空间配额（ResourceQuota / LimitRange）", "命名空间配额-resourcequota--limitrange"},
		{"副本数扩缩容", "副本数扩缩容"},
	}
	for _, tc := range cases {
		got := string(ids.Generate([]byte(tc.title), ast.KindHeading))
		if got != tc.want {
			t.Fatalf("title %q: got %q want %q", tc.title, got, tc.want)
		}
	}
	// duplicate
	got := string(ids.Generate([]byte("Pod"), ast.KindHeading))
	if got != "pod-1" {
		t.Fatalf("dup: got %q", got)
	}
}

func TestRender_unicodeHeadingAnchors(t *testing.T) {
	src := []byte(`## 页内导航

[跳转](#页内导航)

## Pod

[去 Pod](#pod)

### CPU/内存 requests 与 limits

[去 CPU](#cpu内存-requests-与-limits)

### HPA 自动扩缩

[去 HPA](#hpa-自动扩缩)
`)
	html, err := Render(src, RenderOpts{})
	if err != nil {
		t.Fatal(err)
	}
	out := string(html)
	for _, id := range []string{
		`id="页内导航"`,
		`id="pod"`,
		`id="cpu内存-requests-与-limits"`,
		`id="hpa-自动扩缩"`,
	} {
		if !strings.Contains(out, id) {
			t.Fatalf("missing %s in:\n%s", id, out)
		}
	}
	// must not fall back to generic heading for Chinese
	if strings.Contains(out, `id="heading"`) {
		t.Fatalf("unexpected generic heading id:\n%s", out)
	}
}

func TestRender_k8sArticleNavAnchors(t *testing.T) {
	// Spot-check: Chinese nav links used by k8s-resources-igo.md style articles.
	src := []byte(`## 页内导航

| x | [通用操作速查](#通用操作速查) · [副本数扩缩容](#副本数扩缩容) · [推荐最小组合](#推荐最小组合) |
|---|---|

## 通用操作速查

### 副本数扩缩容

## 推荐最小组合
`)
	html, err := Render(src, RenderOpts{})
	if err != nil {
		t.Fatal(err)
	}
	out := string(html)
	reH := regexp.MustCompile(`id="([^"]+)"`)
	ids := map[string]bool{}
	for _, m := range reH.FindAllStringSubmatch(out, -1) {
		ids[m[1]] = true
	}
	for _, want := range []string{"页内导航", "通用操作速查", "副本数扩缩容", "推荐最小组合"} {
		if !ids[want] {
			t.Fatalf("missing heading id %q; have %v\nout=\n%s", want, ids, out)
		}
	}
	reA := regexp.MustCompile(`href="#([^"]+)"`)
	for _, m := range reA.FindAllStringSubmatch(out, -1) {
		id := m[1]
		if strings.Contains(id, "%") {
			decoded, err := url.PathUnescape(id)
			if err != nil {
				t.Fatalf("bad href encoding %q: %v", id, err)
			}
			id = decoded
		}
		if !ids[id] {
			t.Fatalf("nav href #%s has no matching heading id; have %v", id, ids)
		}
	}
}
