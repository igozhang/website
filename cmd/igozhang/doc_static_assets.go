package main

import (
	"hash/fnv"
	"fmt"
	"net/http"
	"strconv"
)

func docArticleBundleCSS() string {
	return docModuleCSS() + docArticlePageCSS()
}

func docStaticAssetVersion() string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(docArticleBundleCSS()))
	_, _ = h.Write([]byte(siteHomePageCSS()))
	_, _ = h.Write(docCodeChromeJS)
	_, _ = h.Write(docMermaidLibJS)
	_, _ = h.Write(docMermaidJS)
	_, _ = h.Write([]byte(siteMp3StylesOnly()))
	_, _ = h.Write(mp3InitJS)
	return fmt.Sprintf("%08x", h.Sum32())
}

func docArticleCSSAssetURL(pub string) string {
	return pubPath(pub, "/assets/doc/article.css?v="+docStaticAssetVersion())
}

func docCodeChromeJSAssetURL(pub string) string {
	return pubPath(pub, "/assets/doc/code-chrome.js?v="+docStaticAssetVersion())
}

func docMp3CSSAssetURL(pub string) string {
	return pubPath(pub, "/assets/doc/mp3.css?v="+docStaticAssetVersion())
}

func docMp3InitJSAssetURL(pub string) string {
	return pubPath(pub, "/assets/doc/mp3-init.js?v="+docStaticAssetVersion())
}

func docHomeCSSAssetURL(pub string) string {
	return pubPath(pub, "/assets/doc/home.css?v="+docStaticAssetVersion())
}

func serveImmutableAsset(w http.ResponseWriter, r *http.Request, contentType string, body []byte) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h := fnv.New32a()
	_, _ = h.Write(body)
	etag := fmt.Sprintf(`"%x"`, h.Sum32())
	if inm := r.Header.Get("If-None-Match"); inm == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(body)
}

func serveDocArticleCSS(w http.ResponseWriter, r *http.Request) {
	serveImmutableAsset(w, r, "text/css; charset=utf-8", []byte(docArticleBundleCSS()))
}

func serveDocCodeChromeJS(w http.ResponseWriter, r *http.Request) {
	serveImmutableAsset(w, r, "application/javascript; charset=utf-8", docCodeChromeJS)
}

func serveDocMp3CSS(w http.ResponseWriter, r *http.Request) {
	serveImmutableAsset(w, r, "text/css; charset=utf-8", []byte(siteMp3StylesOnly()))
}

func serveDocMp3InitJS(w http.ResponseWriter, r *http.Request) {
	serveImmutableAsset(w, r, "application/javascript; charset=utf-8", mp3InitJS)
}

func serveDocHomeCSS(w http.ResponseWriter, r *http.Request) {
	serveImmutableAsset(w, r, "text/css; charset=utf-8", []byte(siteHomePageCSS()))
}

func docArticleExternalScriptTag(pub string) string {
	return `<script src="` + docCodeChromeJSAssetURL(pub) + `" defer></script>`
}

func docArticleExternalCSSLink(pub string) string {
	return `<link rel="stylesheet" href="` + docArticleCSSAssetURL(pub) + `">`
}

func docArticleExternalMp3CSSLink(pub string) string {
	return `<link rel="stylesheet" href="` + docMp3CSSAssetURL(pub) + `">`
}

func docHomeExternalCSSLink(pub string) string {
	return `<link rel="stylesheet" href="` + docHomeCSSAssetURL(pub) + `">`
}
