package main

import (
	"strings"
	"testing"
)

func TestSiteFootQuoteHTML_HiddenHits(t *testing.T) {
	h := siteFootQuoteHTML()
	if strings.Contains(h, `id="footOpsHit"`) || strings.Contains(h, `id="footAdminHit"`) {
		t.Fatal("ops/admin footer hits should be removed")
	}
	if !strings.Contains(h, `id="footIgoHit"`) {
		t.Fatal("missing footIgoHit span")
	}
	if strings.Count(h, "死去") != 1 || strings.Count(h, "昆虫") != 1 || strings.Count(h, "专业化") != 1 {
		t.Fatalf("死去=%d 昆虫=%d 专业化=%d", strings.Count(h, "死去"), strings.Count(h, "昆虫"), strings.Count(h, "专业化"))
	}
	want := "只有昆虫才专业化。"
	if !strings.Contains(siteFootQuote, want) {
		t.Fatal("quote fixture must contain 只有昆虫才专业化。")
	}
}

func TestSiteFootInit_safariIgoKeyCopy(t *testing.T) {
	js := string(siteFootInitJS)
	for _, needle := range []string{
		"footPrefetchIgoKey",
		"igoKeyCache",
		"document.execCommand(\"copy\")",
	} {
		if !strings.Contains(js, needle) {
			t.Fatalf("site-foot-init.js missing %q", needle)
		}
	}
	for _, banned := range []string{
		"footIgoCopyToast",
		"验证密钥已复制到剪贴板",
		"无法自动复制",
		"showToast",
	} {
		if strings.Contains(js, banned) {
			t.Fatalf("site-foot-init.js should not prompt on copy: %q", banned)
		}
	}
}
