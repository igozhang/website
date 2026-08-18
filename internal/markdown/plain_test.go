package markdown

import (
	"strings"
	"testing"
)

func TestPlainTextExcerpt_basic(t *testing.T) {
	body := []byte(`# squid_ins

bash安装squid，ubuntu为例

## ubuntu

` + "```\necho hi\n```\n\n" + `更多说明文字在这里。`)
	got := PlainTextExcerpt(body, 3, "squid_ins")
	if got == "" {
		t.Fatal("empty excerpt")
	}
	if strings.Contains(got, "```") || strings.Contains(got, "echo hi") {
		t.Fatalf("code leaked: %q", got)
	}
	if !strings.Contains(got, "bash安装squid") {
		t.Fatalf("missing body: %q", got)
	}
}

func TestPlainTextExcerpt_unescapesAndSkipsTitle(t *testing.T) {
	body := []byte(`# Calico BIRD 报告

1\. 问题现象 环境：k8s\-master52，free \-g 显示
![图示](doc-pic/x.png)
| col | val |
|---|---|
| a | b |`)
	got := PlainTextExcerpt(body, 5, "Calico BIRD 报告")
	if strings.Contains(got, `\`) || strings.Contains(got, "|") || strings.Contains(got, "doc-pic") {
		t.Fatalf("markdown source leaked: %q", got)
	}
	if strings.Contains(got, "Calico BIRD 报告") {
		t.Fatalf("title should be skipped: %q", got)
	}
	for _, want := range []string{"1. 问题现象", "k8s-master52", "free -g"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestPlainTextExcerpt_strikeAndLink(t *testing.T) {
	body := []byte(`正文 ~~旧~~ 见 [文档](https://example.com/a) 结束。`)
	got := PlainTextExcerpt(body, 2, "")
	if strings.Contains(got, "~~") || strings.Contains(got, "](http") {
		t.Fatalf("raw md: %q", got)
	}
	if !strings.Contains(got, "旧") || !strings.Contains(got, "文档") {
		t.Fatalf("got %q", got)
	}
}
