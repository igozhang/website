package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/dhowden/tag"
)

var maxCoverBytes = 8 << 20

type mp3CoverCacheEntry struct {
	mime string
	data []byte
	etag string
	ok   bool // false = probed, no APIC
}

var (
	mp3CoverCacheMu sync.RWMutex
	mp3CoverCache   = map[string]mp3CoverCacheEntry{}
)

func mp3CoverCacheKey(full string, modUnix, size int64) string {
	return full + "|" + strconv.FormatInt(modUnix, 10) + "|" + strconv.FormatInt(size, 10)
}

// readMp3APICPicture returns embedded ID3 cover bytes and MIME type, or empty if none.
func readMp3APICPicture(full string) (mime string, data []byte) {
	f, err := os.Open(full)
	if err != nil {
		return "", nil
	}
	defer f.Close()

	md, err := tag.ReadFrom(f)
	if err != nil || md == nil {
		return "", nil
	}

	pic := md.Picture()
	if pic == nil || len(pic.Data) == 0 || len(pic.Data) > maxCoverBytes {
		return "", nil
	}
	data = pic.Data
	mime = strings.TrimSpace(pic.MIMEType)
	if mime == "" {
		switch strings.ToLower(pic.Ext) {
		case "jpg", "jpeg":
			mime = "image/jpeg"
		case "png":
			mime = "image/png"
		case "gif":
			mime = "image/gif"
		case "webp":
			mime = "image/webp"
		default:
			mime = "image/jpeg"
		}
	}
	return mime, data
}

func cachedMp3Cover(full string, st os.FileInfo) (mime string, data []byte, etag string, ok bool) {
	key := mp3CoverCacheKey(full, st.ModTime().Unix(), st.Size())
	mp3CoverCacheMu.RLock()
	if e, hit := mp3CoverCache[key]; hit {
		mp3CoverCacheMu.RUnlock()
		if !e.ok || len(e.data) == 0 {
			return "", nil, "", false
		}
		return e.mime, e.data, e.etag, true
	}
	mp3CoverCacheMu.RUnlock()

	mime, data = readMp3APICPicture(full)
	entry := mp3CoverCacheEntry{ok: len(data) > 0}
	if entry.ok {
		entry.mime = mime
		entry.data = data
		entry.etag = fmt.Sprintf(`"%d-%d-%d"`, st.ModTime().Unix(), st.Size(), len(data))
	}
	mp3CoverCacheMu.Lock()
	mp3CoverCache[key] = entry
	mp3CoverCacheMu.Unlock()
	if !entry.ok {
		return "", nil, "", false
	}
	return entry.mime, entry.data, entry.etag, true
}

// mp3CoverHandler serves raw embedded cover image for MediaSession / lock screen (HTTPS URL; data: URLs are unreliable on Android Chrome).
func mp3CoverHandler(dataRoot string) http.HandlerFunc {
	root := docAssetsMp3Dir(dataRoot)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !requireMp3Gate(w, r) {
			return
		}
		q := filepath.ToSlash(strings.TrimSpace(r.URL.Query().Get("path")))
		full, ok := safeJoin(root, q)
		if !ok {
			jsonError(w, http.StatusBadRequest, "invalid path")
			return
		}
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			http.NotFound(w, r)
			return
		}
		if strings.ToLower(filepath.Ext(full)) != ".mp3" {
			jsonError(w, http.StatusBadRequest, "not an mp3 file")
			return
		}

		mime, data, etag, has := cachedMp3Cover(full, st)
		if !has {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("ETag", etag)
		if etagNotModified(r, etag) {
			w.Header().Set("Content-Type", mime)
			w.Header().Set("Cache-Control", "private, max-age=86400")
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "private, max-age=86400")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(data)
	}
}
