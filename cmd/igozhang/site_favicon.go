package main

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"html"
	"strings"
)

// secret07 可放置站点 favicon：favicon.{png,jpg,jpeg} 或 fav.{png,jpg,jpeg}（大小写不敏感）。
// 单文件不得超过 siteFaviconMaxBytes（100K），超出则忽略、回退内嵌默认图标。
// 多个同时存在时按 (stem 优先 favicon>fav, 扩展名优先 png>jpeg>jpg) 取一条。

const siteFaviconMaxBytes = 100 * 1024

var (
	siteFaviconDataRoot string
	faviconDataHref     template.URL
)

func init() {
	faviconDataHref = template.URL("data:image/svg+xml;base64," +
		base64.StdEncoding.EncodeToString([]byte(faviconSVG)))
}

func initSiteFavicon(dataRoot string) {
	siteFaviconDataRoot = dataRoot
	if c := resolveSecret07Favicon(dataRoot); c != nil {
		log.Printf("site-favicon: using secret07/%s (%s)", c.FileName, c.MIME)
		return
	}
	for _, name := range secret07OversizedFaviconNames(dataRoot) {
		log.Printf("site-favicon: ignored secret07/%s (>100K)", name)
	}
}

type secret07Favicon struct {
	Path     string
	FileName string
	Base     string
	Ext      string
	MIME     string
}

func registerFaviconHandlers(mux *http.ServeMux, dataRoot string) {
	initSiteFavicon(dataRoot)
	mux.HandleFunc("/favicon.svg", siteFaviconSVGHandler)
	mux.HandleFunc("/favicon.png", func(w http.ResponseWriter, r *http.Request) {
		serveSiteFaviconRaster(w, r, "png")
	})
	mux.HandleFunc("/favicon.jpg", func(w http.ResponseWriter, r *http.Request) {
		serveSiteFaviconRaster(w, r, "jpg")
	})
	mux.HandleFunc("/favicon.jpeg", func(w http.ResponseWriter, r *http.Request) {
		serveSiteFaviconRaster(w, r, "jpeg")
	})
}

func siteFaviconSVGHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(faviconSVG))
}

func serveSiteFaviconRaster(w http.ResponseWriter, r *http.Request, wantExt string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	wantExt = strings.ToLower(wantExt)
	if c := resolveSecret07Favicon(siteFaviconDataRoot); c != nil && strings.EqualFold(c.Ext, wantExt) {
		serveSecret07FaviconFile(w, r, c)
		return
	}
	if wantExt == "png" {
		writeSiteFaviconPNG(w)
		return
	}
	http.NotFound(w, r)
}

func serveSecret07FaviconFile(w http.ResponseWriter, r *http.Request, c *secret07Favicon) {
	fi, err := os.Stat(c.Path)
	if err != nil || fi.Size() > siteFaviconMaxBytes {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", c.MIME)
	w.Header().Set("Cache-Control", "public, max-age=300")
	if r.Method == http.MethodHead {
		return
	}
	f, err := os.Open(c.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	http.ServeContent(w, r, c.FileName, fi.ModTime(), f)
}

func resolveSecret07Favicon(dataRoot string) *secret07Favicon {
	if dataRoot == "" {
		return nil
	}
	dir := secret07Dir(dataRoot)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var best *secret07Favicon
	bestStem, bestExt := 99, 99
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		stem, ext, ok := splitFaviconName(e.Name())
		if !ok {
			continue
		}
		sr, er := faviconStemRank(stem), faviconExtRank(ext)
		if sr > 1 || er > 2 {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if !faviconFileSizeOK(e, path) {
			continue
		}
		if best == nil || sr < bestStem || (sr == bestStem && er < bestExt) {
			bestStem, bestExt = sr, er
			best = &secret07Favicon{
				Path:     path,
				FileName: e.Name(),
				Base:     stem,
				Ext:      strings.ToLower(ext),
				MIME:     faviconMIME(ext),
			}
		}
	}
	return best
}

func faviconFileSizeOK(ent os.DirEntry, path string) bool {
	fi, err := ent.Info()
	if err != nil {
		fi, err = os.Stat(path)
		if err != nil {
			return false
		}
	}
	return fi.Size() > 0 && fi.Size() <= siteFaviconMaxBytes
}

func secret07OversizedFaviconNames(dataRoot string) []string {
	if dataRoot == "" {
		return nil
	}
	dir := secret07Dir(dataRoot)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if _, _, ok := splitFaviconName(e.Name()); !ok {
			continue
		}
		path := filepath.Join(dir, e.Name())
		fi, err := e.Info()
		if err != nil {
			fi, err = os.Stat(path)
			if err != nil {
				continue
			}
		}
		if fi.Size() > siteFaviconMaxBytes {
			out = append(out, e.Name())
		}
	}
	return out
}

func splitFaviconName(name string) (stem, ext string, ok bool) {
	low := strings.ToLower(name)
	for _, e := range []string{".png", ".jpg", ".jpeg"} {
		if strings.HasSuffix(low, e) {
			stem = strings.ToLower(name[:len(name)-len(e)])
			ext = strings.TrimPrefix(e, ".")
			break
		}
	}
	if ext == "" {
		return "", "", false
	}
	switch stem {
	case "favicon", "fav":
		return stem, ext, true
	default:
		return "", "", false
	}
}

func faviconStemRank(stem string) int {
	switch strings.ToLower(stem) {
	case "favicon":
		return 0
	case "fav":
		return 1
	default:
		return 99
	}
}

func faviconExtRank(ext string) int {
	switch strings.ToLower(ext) {
	case "png":
		return 0
	case "jpeg":
		return 1
	case "jpg":
		return 2
	default:
		return 99
	}
}

func faviconMIME(ext string) string {
	switch strings.ToLower(ext) {
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

// siteFaviconIconLink 返回 <link rel="icon"> 的 href 与 type。
func siteFaviconIconLink(pub string) (href string, mimeType string) {
	if c := resolveSecret07Favicon(siteFaviconDataRoot); c != nil {
		return pubPath(pub, "favicon."+c.Ext), c.MIME
	}
	return string(faviconDataHref), "image/svg+xml"
}

// siteFaviconFetchHref 返回可供 fetch / Media Session 使用的图标 URL（非 data:）。
func siteFaviconFetchHref(pub string) string {
	if c := resolveSecret07Favicon(siteFaviconDataRoot); c != nil {
		return pubPath(pub, "favicon."+c.Ext)
	}
	return pubPath(pub, "favicon.svg")
}

// siteFaviconMediaHref 锁屏/通知栏占位图：定制 raster 用原文件，否则 /favicon.png 生成位图。
func siteFaviconMediaHref(pub string) string {
	if c := resolveSecret07Favicon(siteFaviconDataRoot); c != nil {
		return pubPath(pub, "favicon."+c.Ext)
	}
	return pubPath(pub, "favicon.png")
}

func pageHeadFaviconMeta(pub string) string {
	icon := siteFaviconFetchHref(pub)
	media := siteFaviconMediaHref(pub)
	return `<meta name="fs-favicon-href" content="` + html.EscapeString(icon) +
		`"><meta name="fs-favicon-media-href" content="` + html.EscapeString(media) + `">`
}
