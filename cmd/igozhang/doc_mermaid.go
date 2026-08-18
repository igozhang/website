package main

import (
	"bytes"
	"net/http"

	_ "embed"

	"github.com/igokorea/igozhang/internal/docmermaid"
)

// Vendor lib lives in internal/docmermaid so package main rebuilds without re-packing 2.5MB.
var docMermaidLibJS = docmermaid.LibJS

//go:embed doc-mermaid.js
var docMermaidJS []byte

// docMermaidCSS styles rendered diagrams and parse-error fallback on article / admin preview.
func docMermaidCSS() string {
	return `
/* Scheme B: transparent frame; light node fill only */
.doc-mermaid-wrap{margin:.85rem 0;overflow-x:auto;max-width:100%;text-align:center;
  background:transparent;border:none;padding:.25rem 0}
.doc-mermaid-wrap svg{max-width:100%;height:auto;background:transparent}
.doc-mermaid-wrap .cluster rect{fill:transparent!important;stroke:var(--line,#b8c9b8)!important}
.doc-mermaid-wrap .edgeLabel,.doc-mermaid-wrap .edgeLabel p,.doc-mermaid-wrap .labelBkg{
  background-color:transparent!important;background:transparent!important}
.doc-mermaid-wrap .node .label-container,.doc-mermaid-wrap .node rect.basic,
.doc-mermaid-wrap .node circle,.doc-mermaid-wrap .node ellipse,.doc-mermaid-wrap .node polygon{
  fill:var(--doc-surface,#d8e4d8)!important;stroke:var(--line,#b8c9b8)!important}
.doc-mermaid-err{margin:0 0 .55rem;padding:.35rem .55rem;text-align:left;font-size:.82rem;line-height:1.4;
  color:#8a1f1f;background:rgba(180,40,40,.08);border:1px solid rgba(180,40,40,.22);border-radius:6px}
.doc-mermaid-wrap pre{margin:0!important;text-align:left}
`
}

func docMermaidLibJSAssetURL(pub string) string {
	return pubPath(pub, "/assets/doc/mermaid.min.js?v="+docStaticAssetVersion())
}

func docMermaidJSAssetURL(pub string) string {
	return pubPath(pub, "/assets/doc/mermaid.js?v="+docStaticAssetVersion())
}

func serveDocMermaidLibJS(w http.ResponseWriter, r *http.Request) {
	serveImmutableAsset(w, r, "application/javascript; charset=utf-8", docMermaidLibJS)
}

func serveDocMermaidJS(w http.ResponseWriter, r *http.Request) {
	serveImmutableAsset(w, r, "application/javascript; charset=utf-8", docMermaidJS)
}

// docMermaidScriptTags loads mermaid lib then the thin renderer (defer, order preserved).
func docMermaidScriptTags(pub string) string {
	return `<script src="` + docMermaidLibJSAssetURL(pub) + `" defer></script>` +
		`<script src="` + docMermaidJSAssetURL(pub) + `" defer></script>`
}

func htmlHasMermaidFence(body []byte) bool {
	return bytes.Contains(bytes.ToLower(body), []byte("language-mermaid"))
}

// docArticleBodyScripts returns code-chrome scripts, plus Mermaid when the HTML has a mermaid fence.
func docArticleBodyScripts(pub string, htmlBody []byte) string {
	s := docArticleExternalScriptTag(pub)
	if htmlHasMermaidFence(htmlBody) {
		s += docMermaidScriptTags(pub)
	}
	return s
}
