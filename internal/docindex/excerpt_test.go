package docindex

import (
	"strings"
	"testing"
)

func TestExcerptMarkdown(t *testing.T) {
	body := []byte(`# squid_ins

bash安装squid，ubuntu为例

## ubuntu

` + "```\necho hi\n```\n\n" + `更多说明文字在这里。`)
	got := ExcerptMarkdownSkipTitle(body, 3, "")
	if got == "" {
		t.Fatal("empty excerpt")
	}
	if strings.Contains(got, "```") || strings.Contains(got, "echo hi") {
		t.Fatalf("code leaked: %q", got)
	}
	if !strings.Contains(got, "bash安装squid") {
		t.Fatalf("missing body text: %q", got)
	}
}

func TestExcerptMarkdownSkipTitle(t *testing.T) {
	body := []byte(`# Calico BIRD 报告

1\. 问题现象 环境：k8s\-master52`)
	got := ExcerptMarkdownSkipTitle(body, 5, "Calico BIRD 报告")
	if strings.Contains(got, `\`) {
		t.Fatalf("escapes: %q", got)
	}
	if strings.Contains(got, "Calico BIRD 报告") {
		t.Fatalf("duplicate title: %q", got)
	}
	if !strings.Contains(got, "k8s-master52") {
		t.Fatalf("got %q", got)
	}
}
