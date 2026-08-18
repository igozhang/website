package main

import (
	_ "embed"
	"net/http"
	"strings"
)

//go:embed mermaid_page.html
var mermaidPageTpl []byte

func renderMermaidPageHTML(pub string, r *http.Request) []byte {
	return applySitePage(string(mermaidPageTpl), pub, sitePageLayout{}, map[string]string{
		"{{MERMAID_CSS}}":     docMermaidCSS(),
		"{{MERMAID_SCRIPTS}}": docMermaidScriptTags(pub),
	}, visitorGeoLabel(r))
}

func registerMermaidPageHandler(mux *http.ServeMux) {
	serve := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
		if pub == "" {
			if p := strings.TrimSpace(getenv("BASE_PATH", "")); p != "" {
				pub = strings.TrimSuffix(p, "/")
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		_, _ = w.Write(renderMermaidPageHTML(pub, r))
	}
	mux.HandleFunc("/mermaid", serve)
	mux.HandleFunc("/mermaid/", serve)
	mux.HandleFunc("/Mermaid", serve)
	mux.HandleFunc("/Mermaid/", serve)
}
