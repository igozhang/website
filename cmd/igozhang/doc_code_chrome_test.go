package main

import (
	"strings"
	"testing"
)

func TestDocCodeBlockCSS_shellPalette(t *testing.T) {
	css := docCodeBlockCSS()
	for _, frag := range []string{
		"--hl-comment:#6a9955",
		"--hl-cmd:#dcdcaa",
		"--hl-string:#ce9178",
		".doc-code-wrap .doc-hl-keyword",
		".doc-code-wrap .doc-hl-key",
		".doc-code-wrap .doc-hl-tag",
		".doc-code-wrap .doc-hl-path",
		".doc-code-wrap .doc-hl-err",
	} {
		if !strings.Contains(css, frag) {
			t.Fatalf("docCodeBlockCSS missing %q", frag)
		}
	}
}

func TestDocCodeChromeJS_highlightAndInit(t *testing.T) {
	js := string(docCodeChromeJS)
	for _, frag := range []string{
		"function highlightDocShell",
		"function highlightDocCode",
		"highlightDocYAML",
		"highlightDocJSON",
		"highlightDocGo",
		"highlightDocPython",
		"highlightDocXML",
		"detectCodeLang",
		"initDocCodeChrome",
		"doc-hl-comment",
		"doc-hl-flag",
		"systemctl",
		"data-lang",
		"expandPlaceholders",
		"requestIdleCallback",
		"escHtml(t.raw)",
		`if (lang === "mermaid") return`,
		`return "\x00h" + i + "\x00"`,
		"Never rematch inside placeholders",
	} {
		if !strings.Contains(js, frag) {
			t.Fatalf("doc-code-chrome.js missing %q", frag)
		}
	}
	if !strings.Contains(docCodeChromeScript(), "<script>") {
		t.Fatal("docCodeChromeScript must wrap JS")
	}
}
