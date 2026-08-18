package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// secret07 可放置页眉背景：banner.{png,jpg,jpeg}，单文件 ≤ siteBannerMaxBytes（200K）。

const siteBannerMaxBytes = 200 * 1024

var siteBannerDataRoot string

type secret07Banner struct {
	Path     string
	FileName string
	Ext      string
	MIME     string
}

func initSiteBanner(dataRoot string) {
	siteBannerDataRoot = dataRoot
	if c := resolveSecret07Banner(dataRoot); c != nil {
		log.Printf("site-banner: using secret07/%s (%s)", c.FileName, c.MIME)
		return
	}
	for _, name := range secret07OversizedBannerNames(dataRoot) {
		log.Printf("site-banner: ignored secret07/%s (>200K)", name)
	}
}

func registerBannerHandlers(mux *http.ServeMux, dataRoot string) {
	initSiteBanner(dataRoot)
	mux.HandleFunc("/banner.png", func(w http.ResponseWriter, r *http.Request) {
		serveSiteBannerRaster(w, r, "png")
	})
	mux.HandleFunc("/banner.jpg", func(w http.ResponseWriter, r *http.Request) {
		serveSiteBannerRaster(w, r, "jpg")
	})
	mux.HandleFunc("/banner.jpeg", func(w http.ResponseWriter, r *http.Request) {
		serveSiteBannerRaster(w, r, "jpeg")
	})
}

func serveSiteBannerRaster(w http.ResponseWriter, r *http.Request, wantExt string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	wantExt = strings.ToLower(wantExt)
	c := resolveSecret07Banner(siteBannerDataRoot)
	if c == nil || !strings.EqualFold(c.Ext, wantExt) {
		http.NotFound(w, r)
		return
	}
	fi, err := os.Stat(c.Path)
	if err != nil || fi.Size() > siteBannerMaxBytes || fi.Size() == 0 {
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

func resolveSecret07Banner(dataRoot string) *secret07Banner {
	if dataRoot == "" {
		return nil
	}
	dir := secret07Dir(dataRoot)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var best *secret07Banner
	bestExt := 99
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		ext, ok := splitBannerName(e.Name())
		if !ok {
			continue
		}
		er := rasterImageExtRank(ext)
		if er > 2 {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if !bannerFileSizeOK(e, path) {
			continue
		}
		if best == nil || er < bestExt {
			bestExt = er
			best = &secret07Banner{
				Path:     path,
				FileName: e.Name(),
				Ext:      strings.ToLower(ext),
				MIME:     rasterImageMIME(ext),
			}
		}
	}
	return best
}

func splitBannerName(name string) (ext string, ok bool) {
	low := strings.ToLower(name)
	if low != "banner.png" && low != "banner.jpg" && low != "banner.jpeg" {
		return "", false
	}
	if strings.HasSuffix(low, ".jpeg") {
		return "jpeg", true
	}
	if strings.HasSuffix(low, ".jpg") {
		return "jpg", true
	}
	if strings.HasSuffix(low, ".png") {
		return "png", true
	}
	return "", false
}

func bannerFileSizeOK(ent os.DirEntry, path string) bool {
	fi, err := ent.Info()
	if err != nil {
		fi, err = os.Stat(path)
		if err != nil {
			return false
		}
	}
	return fi.Size() > 0 && fi.Size() <= siteBannerMaxBytes
}

func secret07OversizedBannerNames(dataRoot string) []string {
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
		if _, ok := splitBannerName(e.Name()); !ok {
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
		if fi.Size() > siteBannerMaxBytes {
			out = append(out, e.Name())
		}
	}
	return out
}

func rasterImageExtRank(ext string) int {
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

func rasterImageMIME(ext string) string {
	switch strings.ToLower(ext) {
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func siteBannerBackgroundURL(pub string) string {
	if c := resolveSecret07Banner(siteBannerDataRoot); c != nil {
		return pubPath(pub, "banner."+c.Ext)
	}
	return ""
}
