package main

import (
	"strings"
	"testing"
)

func TestTruncateDocCapsuleTitle(t *testing.T) {
	short := "Redis 从节点 CPU 过高"
	if got := truncateDocCapsuleTitle(short); got != short {
		t.Fatalf("short title unchanged: got %q", got)
	}
	long := "这是一段超过二十二个字符的文章标题用于测试截断效果"
	got := truncateDocCapsuleTitle(long)
	if len([]rune(got)) != 23 {
		t.Fatalf("truncated len: got %q (%d runes)", got, len([]rune(got)))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("truncated must end with ellipsis: got %q", got)
	}
}
