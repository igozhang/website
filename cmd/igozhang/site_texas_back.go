package main

import (
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// secret07 可放置德州扑克牌背：texas.{png,jpg,jpeg}（大小写不敏感）。
// 单文件不得超过 texasCardBackMaxBytes（150K），超出则忽略、回退 Canvas 绘制。
// 多个同时存在时扩展名优先 png > jpeg > jpg。

const texasCardBackMaxBytes = 150 * 1024

var texasCardBackDataRoot string

type secret07TexasCardBack struct {
	Path     string
	FileName string
	Ext      string
	MIME     string
}

func initTexasCardBack(dataRoot string) {
	texasCardBackDataRoot = dataRoot
	if c := resolveSecret07TexasCardBack(dataRoot); c != nil {
		log.Printf("texas-card-back: using secret07/%s (%s)", c.FileName, c.MIME)
		return
	}
	for _, name := range secret07OversizedTexasCardBackNames(dataRoot) {
		log.Printf("texas-card-back: ignored secret07/%s (>150K)", name)
	}
}

func registerTexasCardBackHandlers(mux *http.ServeMux, dataRoot string) {
	initTexasCardBack(dataRoot)
	mux.HandleFunc("/texas-card-back.png", func(w http.ResponseWriter, r *http.Request) {
		serveTexasCardBackRaster(w, r, "png")
	})
	mux.HandleFunc("/texas-card-back.jpg", func(w http.ResponseWriter, r *http.Request) {
		serveTexasCardBackRaster(w, r, "jpg")
	})
	mux.HandleFunc("/texas-card-back.jpeg", func(w http.ResponseWriter, r *http.Request) {
		serveTexasCardBackRaster(w, r, "jpeg")
	})
}

func serveTexasCardBackRaster(w http.ResponseWriter, r *http.Request, wantExt string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	wantExt = strings.ToLower(wantExt)
	if c := resolveSecret07TexasCardBack(texasCardBackDataRoot); c != nil && strings.EqualFold(c.Ext, wantExt) {
		serveSecret07TexasCardBackFile(w, r, c)
		return
	}
	http.NotFound(w, r)
}

func serveSecret07TexasCardBackFile(w http.ResponseWriter, r *http.Request, c *secret07TexasCardBack) {
	fi, err := os.Stat(c.Path)
	if err != nil || fi.Size() > texasCardBackMaxBytes {
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

func resolveSecret07TexasCardBack(dataRoot string) *secret07TexasCardBack {
	if dataRoot == "" {
		return nil
	}
	dir := secret07Dir(dataRoot)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var best *secret07TexasCardBack
	bestExt := 99
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		stem, ext, ok := splitTexasCardBackName(e.Name())
		if !ok || stem != "texas" {
			continue
		}
		er := faviconExtRank(ext)
		if er > 2 {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if !texasCardBackFileSizeOK(e, path) {
			continue
		}
		if best == nil || er < bestExt {
			bestExt = er
			best = &secret07TexasCardBack{
				Path:     path,
				FileName: e.Name(),
				Ext:      strings.ToLower(ext),
				MIME:     faviconMIME(ext),
			}
		}
	}
	return best
}

func texasCardBackFileSizeOK(ent os.DirEntry, path string) bool {
	fi, err := ent.Info()
	if err != nil {
		fi, err = os.Stat(path)
		if err != nil {
			return false
		}
	}
	return fi.Size() > 0 && fi.Size() <= texasCardBackMaxBytes
}

func secret07OversizedTexasCardBackNames(dataRoot string) []string {
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
		if _, _, ok := splitTexasCardBackName(e.Name()); !ok {
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
		if fi.Size() > texasCardBackMaxBytes {
			out = append(out, e.Name())
		}
	}
	return out
}

func splitTexasCardBackName(name string) (stem, ext string, ok bool) {
	low := strings.ToLower(name)
	for _, e := range []string{".png", ".jpg", ".jpeg"} {
		if strings.HasSuffix(low, e) {
			stem = strings.ToLower(name[:len(name)-len(e)])
			ext = strings.TrimPrefix(e, ".")
			return stem, ext, stem == "texas"
		}
	}
	return "", "", false
}

func pageHeadTexasCardBackMeta(pub string) string {
	if c := resolveSecret07TexasCardBack(texasCardBackDataRoot); c != nil {
		return `<meta name="fs-texas-card-back" content="` +
			html.EscapeString(pubPath(pub, "texas-card-back."+c.Ext)) + `">` + "\n"
	}
	return ""
}
