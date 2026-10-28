package main

import (
	"strings"
	"testing"
)

func TestDocMermaidAssets_embedded(t *testing.T) {
	if len(docMermaidLibJS) < 100000 {
		t.Fatalf("mermaid.min.js too small: %d", len(docMermaidLibJS))
	}
	head := string(docMermaidLibJS[:500])
	if !strings.Contains(head, "use strict") {
		t.Fatal("mermaid.min.js does not look like JS")
	}
	js := string(docMermaidJS)
	for _, frag := range []string{
		"function initDocMermaid",
		"function renderMermaidSource",
		"global.renderMermaidSource",
		"exportMermaidPngFromPreview",
		"exportMermaidPngFromSource",
		"htmlLabels: forPreview",
		"useMaxWidth: forPreview",
		"sanitizeSvgForExport",
		"svgExportSizeFromStringRegex",
		"clampExportScale",
		"EXPORT_MAX_EDGE",
		"svgElementToExportString",
		"renderMermaidSvgForExport",
		"language-mermaid",
		"doc-mermaid-wrap",
		"mermaid.render",
		".render(id, text)",
		"Mermaid 渲染失败",
		`clusterBkg: "transparent"`,
		`theme: "base"`,
	} {
		if frag == "mermaid.render" {
			if !strings.Contains(js, "global.mermaid") || !strings.Contains(js, ".render(id, text)") {
				t.Fatalf("doc-mermaid.js missing mermaid.render call")
			}
			continue
		}
		if !strings.Contains(js, frag) {
			t.Fatalf("doc-mermaid.js missing %q", frag)
		}
	}
}

func TestDocMermaidCSS(t *testing.T) {
	css := docMermaidCSS()
	for _, frag := range []string{
		".doc-mermaid-wrap", ".doc-mermaid-err", "svg",
		"background:transparent", ".cluster rect", "fill:transparent!important",
	} {
		if !strings.Contains(css, frag) {
			t.Fatalf("docMermaidCSS missing %q", frag)
		}
	}
	if !strings.Contains(docCodeBlockCSS(), ".doc-mermaid-wrap") {
		t.Fatal("docCodeBlockCSS must include mermaid styles")
	}
}

func TestHtmlHasMermaidFence(t *testing.T) {
	if !htmlHasMermaidFence([]byte(`<pre><code class="language-mermaid">flowchart TB
A-->B
</code></pre>`)) {
		t.Fatal("expected detect language-mermaid")
	}
	if htmlHasMermaidFence([]byte(`<p>hello</p><pre><code class="language-bash">echo hi</code></pre>`)) {
		t.Fatal("must not false-positive on bash")
	}
}

func TestDocArticleBodyScripts_conditionalMermaid(t *testing.T) {
	plain := docArticleBodyScripts("", []byte(`<p>hi</p>`))
	if !strings.Contains(plain, "/assets/doc/code-chrome.js") {
		t.Fatal("plain article must include code-chrome")
	}
	if strings.Contains(plain, "mermaid.min.js") || strings.Contains(plain, "/assets/doc/mermaid.js") {
		t.Fatal("plain article must not load mermaid")
	}
	with := docArticleBodyScripts("", []byte(`<code class="language-mermaid">x</code>`))
	if !strings.Contains(with, "mermaid.min.js") || !strings.Contains(with, "/assets/doc/mermaid.js") {
		t.Fatal("mermaid article must load mermaid scripts")
	}
	if !strings.Contains(docMermaidScriptTags(""), "defer") {
		t.Fatal("mermaid script tags must use defer")
	}
}
