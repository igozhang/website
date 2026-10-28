package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMermaidPage_embedded(t *testing.T) {
	html := string(mermaidPageTpl)
	for _, frag := range []string{
		"id=\"mermaidSrc\"",
		"id=\"mermaidPreview\"",
		"id=\"mmdDrawer\"",
		"id=\"btnMmdDrawer\"",
		"大图预览",
		"btnMmdDrawerPng",
		"btnMmdZoomFit",
		"igo.mermaidDrawerPx",
		"is-mmd-drawer-open",
		"renderMermaidSource",
		"data-mmd-example",
		"{{MERMAID_CSS}}",
		"{{MERMAID_SCRIPTS}}",
	} {
		if !strings.Contains(html, frag) {
			t.Fatalf("mermaid_page.html missing %q", frag)
		}
	}
}

func TestRenderMermaidPageHTML(t *testing.T) {
	html := string(renderMermaidPageHTML("", nil))
	for _, frag := range []string{
		"Mermaid 预览",
		"id=\"mermaidSrc\"",
		"id=\"mmdDrawer\"",
		"大图预览",
		"/assets/doc/mermaid.min.js",
		"/assets/doc/mermaid.js",
		".doc-mermaid-wrap",
		"renderMermaidSource",
		"exportMermaidPngFromSource",
	} {
		if !strings.Contains(html, frag) {
			t.Fatalf("renderMermaidPageHTML missing %q", frag)
		}
	}
	for _, bad := range []string{"{{MERMAID_CSS}}", "{{MERMAID_SCRIPTS}}"} {
		if strings.Contains(html, bad) {
			t.Fatalf("renderMermaidPageHTML still has placeholder %q", bad)
		}
	}
}

func TestMermaidPageHandler(t *testing.T) {
	mux := http.NewServeMux()
	registerMermaidPageHandler(mux)
	for _, path := range []string{"/mermaid", "/Mermaid"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status=%d", path, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Fatalf("%s: Content-Type=%q", path, ct)
		}
		if !strings.Contains(rec.Body.String(), "Mermaid 预览") {
			t.Fatalf("%s: missing page title", path)
		}
	}
}

func TestDocReservedSlug_mermaid(t *testing.T) {
	if !docReservedSlug("mermaid") || !docReservedSlug("Mermaid") {
		t.Fatal("mermaid slug must be reserved")
	}
}
