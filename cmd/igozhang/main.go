package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/igokorea/igozhang/internal/markdown"
)

func main() {
	dataRoot := getenv("DATA_ROOT", "/igozhang.cn")
	closeLog := initAppLogging(dataRoot)
	defer func() { _ = closeLog() }()
	transferRel := getenv("TRANSFER_SUBDIR", "fs")
	basePath := strings.TrimSuffix(getenv("BASE_PATH", ""), "/")
	setVisitorHTTPBasePath(basePath)
	addr := getenv("LISTEN", ":3003")

	transferAbs := filepath.Join(dataRoot, transferRel)
	if err := os.MkdirAll(transferAbs, 0o755); err != nil {
		log.Fatalf("mkdir transfer: %v", err)
	}
	migrateLegacyStash(dataRoot)
	initSecret07Files(dataRoot)
	initAssetsContentFiles(dataRoot)
	initOpsConfig(dataRoot)
	initAuthFailRL(dataRoot)
	initIPBlacklist(dataRoot)
	initPathProbeBan(dataRoot)
	initIPWhitelist(dataRoot)
	initAccessLog(dataRoot)
	initPhrasesCache(dataRoot)
	initWord3kCache(dataRoot)
	initIPGeo()
	initTransferLimits()
	removeDeprecatedGomokuOverride(dataRoot)
	if err := initDocStore(dataRoot); err != nil {
		log.Printf("doc: init %v (articles disabled)", err)
	}

	tmpl := template.Must(template.New("index").Parse(homeIndexHTMLTemplate()))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" && path != "" {
			seg := strings.Trim(path, "/")
			if !strings.Contains(seg, "/") {
				pub := publicBaseFromRequest(r, basePath)
				if tryServeDocArticle(w, r, pub, seg, dataRoot) {
					return
				}
				if docStore != nil && !docReservedSlug(seg) {
					if resolved, ok := docStore.ResolveSlug(seg); ok && resolved != seg {
						http.Redirect(w, r, pubPath(pub, resolved), http.StatusFound)
						return
					}
				}
			}
			http.NotFound(w, r)
			return
		}
		pub := publicBaseFromRequest(r, basePath)
		activeCat := strings.TrimSpace(r.URL.Query().Get("category"))
		page := docParsePage(r.URL.Query().Get("page"))
		issueMp3GateCookie(w, r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, map[string]any{
			"PageHeadHTML":       template.HTML(pageHeadTags(pub)),
			"HomeCSSLinkHTML":    template.HTML(docHomeExternalCSSLink(pub)),
			"FsPubAttr":          fsPubAttr(pub),
			"SiteTitle":          siteDisplayName(),
			"SiteFootHTML":       template.HTML(siteFootBlock(pub, opsFooterGomokuHref(pub), opsFooterXiangqiHref(pub), opsFooterTexasHref(pub), pubPathZhijian(pub), visitorGeoLabel(r))),
			"SiteHeadHTML":       template.HTML(siteHeadBlock(baseHref(pub), pub, siteHeadPhrasesJSON())),
			"DocHomeHTML":        buildDocHomeSection(pub, dataRoot, activeCat, page),
			"Mp3SectionHTML":     template.HTML(siteMp3Section()),
			"Mp3ScriptsHTML":     template.HTML(siteMp3InitScripts(pub)),
			"NavBallScriptsHTML": template.HTML(docNavBallScripts(pub)),
		})
	})
	registerDocHandlers(mux, dataRoot)
	registerFaviconHandlers(mux, dataRoot)
	registerTexasCardBackHandlers(mux, dataRoot)
	registerBannerHandlers(mux, dataRoot)

	mux.HandleFunc("/api/browse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !allowBrowseRL(w, r) {
			return
		}
		q := filepath.ToSlash(r.URL.Query().Get("path"))
		qNorm := strings.Trim(filepath.ToSlash(filepath.Clean(strings.TrimSpace(q))), "/")
		if qNorm == "." {
			qNorm = ""
		}
		// MP3 播放器列表：门禁 Cookie；其余目录浏览需与 /fs 相同的会话 token。
		if qNorm == "assets/mp3" {
			if !requireMp3Gate(w, r) {
				return
			}
		} else if !requireAuthToken(w, r, dataRoot) {
			return
		}
		if !browsePathAllowed(q) {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		target, ok := safeJoin(dataRoot, q)
		if !ok {
			jsonError(w, http.StatusBadRequest, "invalid path")
			return
		}
		st, err := os.Stat(target)
		if err != nil {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		type item struct {
			Name    string `json:"name"`
			IsDir   bool   `json:"isDir"`
			Size    int64  `json:"size"`
			ModTime int64  `json:"modTime"`
		}
		if !st.IsDir() {
			jsonError(w, http.StatusBadRequest, "not a directory")
			return
		}
		entries, err := os.ReadDir(target)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "read dir")
			return
		}
		out := make([]item, 0, len(entries))
		extFilter := r.URL.Query().Get("ext")
		var exts []string
		if extFilter != "" {
			for _, p := range strings.Split(extFilter, ",") {
				p = strings.TrimSpace(strings.ToLower(p))
				if p == "" {
					continue
				}
				if !strings.HasPrefix(p, ".") {
					p = "." + p
				}
				exts = append(exts, p)
			}
		}
		for _, e := range entries {
			if e.IsDir() && !isBrowsableDirName(e.Name()) {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			if len(exts) > 0 && !e.IsDir() {
				ln := strings.ToLower(e.Name())
				ok := false
				for _, ex := range exts {
					if strings.HasSuffix(ln, ex) {
						ok = true
						break
					}
				}
				if !ok {
					continue
				}
			}
			out = append(out, item{
				Name:    e.Name(),
				IsDir:   e.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
			})
		}
		sortKey := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
		if sortKey == "" {
			sortKey = "name"
		}
		desc := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("order"))) == "desc"
		sort.Slice(out, func(i, j int) bool {
			a, b := out[i], out[j]
			if a.IsDir != b.IsDir {
				return a.IsDir
			}
			var less bool
			switch sortKey {
			case "time", "modtime", "mtime":
				less = a.ModTime < b.ModTime
			case "size":
				less = a.Size < b.Size
			default:
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
			if desc {
				return !less
			}
			return less
		})
		jsonWrite(w, out)
	})

	mux.HandleFunc("/api/file", fileDeleteHandler(dataRoot))

	mux.HandleFunc("/api/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !requireAuthToken(w, r, dataRoot) {
			return
		}
		tip := clientIP(r)
		if !acquireTransfer(tip) {
			jsonError(w, http.StatusTooManyRequests, "too many concurrent transfers for this client")
			return
		}
		defer releaseTransfer(tip)
		var body struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
			jsonError(w, http.StatusBadRequest, "bad json")
			return
		}
		raw := strings.TrimSpace(body.URL)
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			jsonError(w, http.StatusBadRequest, "invalid url")
			return
		}
		client := &http.Client{Timeout: getOpsSnapshot().fetchUpstreamTimeout}
		resp, err := client.Get(u.String())
		if err != nil {
			jsonError(w, http.StatusBadGateway, "fetch failed: "+err.Error())
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			jsonError(w, http.StatusBadGateway, "upstream "+resp.Status)
			return
		}
		name := filepath.Base(u.Path)
		if name == "" || name == "." || name == "/" {
			name = "download"
		}
		name = sanitizeFilename(name)
		dest := filepath.Join(transferAbs, name)
		if _, err := os.Stat(dest); err == nil {
			dest = filepath.Join(transferAbs, name+"."+time.Now().Format("20060102-150405"))
		}
		f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, uploadedFileMode)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "create file")
			return
		}

		total := resp.ContentLength
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		fl, canFlush := w.(http.Flusher)

		writeLine := func(v map[string]any) bool {
			b, err := json.Marshal(v)
			if err != nil {
				return false
			}
			if _, err := w.Write(append(b, '\n')); err != nil {
				return false
			}
			if canFlush {
				fl.Flush()
			}
			return true
		}

		start := map[string]any{"type": "start"}
		if total >= 0 {
			start["total"] = total
		}
		if !writeLine(start) {
			f.Close()
			_ = os.Remove(dest)
			return
		}

		bodyLR := io.LimitReader(resp.Body, getOpsSnapshot().uploadMaxBytes)
		buf := make([]byte, 32*1024)
		var written int64
		const progEvery = 128 * 1024
		var lastEmit int64
		for {
			if err := r.Context().Err(); err != nil {
				f.Close()
				_ = os.Remove(dest)
				_ = writeLine(map[string]any{"type": "error", "error": "cancelled"})
				return
			}
			nr, er := bodyLR.Read(buf)
			if nr > 0 {
				nw, ew := f.Write(buf[:nr])
				if ew != nil || nw != nr {
					f.Close()
					_ = os.Remove(dest)
					_ = writeLine(map[string]any{"type": "error", "error": "save failed"})
					return
				}
				written += int64(nw)
				if written-lastEmit >= progEvery {
					lastEmit = written
					pm := map[string]any{"type": "progress", "bytes": written}
					if total >= 0 {
						pm["total"] = total
					}
					if !writeLine(pm) {
						f.Close()
						_ = os.Remove(dest)
						return
					}
				}
			}
			if er == io.EOF {
				break
			}
			if er != nil {
				f.Close()
				_ = os.Remove(dest)
				_ = writeLine(map[string]any{"type": "error", "error": "read failed"})
				return
			}
		}
		pmEnd := map[string]any{"type": "progress", "bytes": written}
		if total >= 0 {
			pmEnd["total"] = total
		}
		_ = writeLine(pmEnd)

		if err := f.Close(); err != nil {
			_ = os.Remove(dest)
			_ = writeLine(map[string]any{"type": "error", "error": "save failed"})
			return
		}
		_ = chmodUploadedFile(dest)
		_ = writeLine(map[string]any{
			"type":  "done",
			"saved": filepath.ToSlash(filepath.Join(transferRel, filepath.Base(dest))),
			"bytes": written,
		})
		invalidateTransferUsageCache()
	})

	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !requireAuthToken(w, r, dataRoot) {
			return
		}
		uip := clientIP(r)
		if !acquireTransfer(uip) {
			jsonError(w, http.StatusTooManyRequests, "too many concurrent transfers for this client")
			return
		}
		defer releaseTransfer(uip)
		uploadMax := getOpsSnapshot().uploadMaxBytes
		r.Body = http.MaxBytesReader(w, r.Body, uploadMax)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			jsonError(w, http.StatusBadRequest, "invalid multipart")
			return
		}
		defer func() {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
		}()
		fhs := r.MultipartForm.File["file"]
		if len(fhs) == 0 {
			jsonError(w, http.StatusBadRequest, "no file")
			return
		}
		var saved []string
		var total int64
		for _, fh := range fhs {
			name := sanitizeFilename(fh.Filename)
			if name == "" {
				name = "upload"
			}
			dest := filepath.Join(transferAbs, name)
			if _, err := os.Stat(dest); err == nil {
				dest = filepath.Join(transferAbs, fmt.Sprintf("%s.%s", name, time.Now().Format("20060102-150405")))
			}
			src, err := fh.Open()
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "open upload")
				return
			}
			dst, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, uploadedFileMode)
			if err != nil {
				src.Close()
				jsonError(w, http.StatusInternalServerError, "create file")
				return
			}
			n, err := io.Copy(dst, io.LimitReader(src, uploadMax))
			src.Close()
			closeErr := dst.Close()
			if err != nil || closeErr != nil {
				_ = os.Remove(dest)
				jsonError(w, http.StatusInternalServerError, "save failed")
				return
			}
			_ = chmodUploadedFile(dest)
			total += n
			saved = append(saved, filepath.ToSlash(filepath.Join(transferRel, filepath.Base(dest))))
		}
		jsonWrite(w, map[string]any{
			"saved": saved,
			"bytes": total,
		})
		invalidateTransferUsageCache()
	})

	mux.HandleFunc("/files/", fileDownloadHandler(dataRoot))
	mux.HandleFunc("/preview/", previewHandler(dataRoot))
	mux.HandleFunc("/stream/mp3/", streamMP3Handler(dataRoot))
	mux.HandleFunc("/api/mp3/cover", mp3CoverHandler(dataRoot))
	mux.HandleFunc("/api/mp3/gate", mp3GateHandler())
	mux.HandleFunc("/api/stash", stashHandler(dataRoot))
	mux.HandleFunc("/api/stash/export", stashExportHandler(dataRoot, transferAbs, transferRel))
	mux.HandleFunc("/api/fs/usage", fsUsageHandler(dataRoot, transferAbs))
	registerFooterHandlers(mux, dataRoot)
	registerMermaidPageHandler(mux)
	registerFsPageHandler(mux)
	registerIgoKeyHandler(mux, dataRoot)
	registerAuthHandlers(mux, dataRoot)
	registerOpsHandlers(mux, dataRoot)
	opsSSHReachable(dataRoot)

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	maintCtx, maintCancel := context.WithCancel(rootCtx)
	defer maintCancel()
	startHourlyMaintenance(maintCtx, dataRoot, transferAbs)

	root := http.NewServeMux()
	root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})
	root.Handle("/", mountAt(basePath, mux))

	handler := recoverHandler(requestGate(dataRoot, logRequests(gzipHandler(root))))

	// ReadHeaderTimeout 防慢速头攻击；不设 WriteTimeout/ReadTimeout，以免打断 MP3 流与大文件传输。
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listen %s DATA_ROOT=%s BASE_PATH=%q", addr, dataRoot, basePath)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-rootCtx.Done():
		log.Printf("shutdown: signal received")
		maintCancel()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}
}

func baseHref(basePath string) string {
	if basePath == "" {
		return "/"
	}
	return basePath + "/"
}

// pubPath builds a browser path under the public prefix (e.g. /fs/gomoku).
func pubPath(pub, rel string) string {
	rel = strings.TrimPrefix(strings.TrimSpace(rel), "/")
	if pub == "" {
		return "/" + rel
	}
	return strings.TrimSuffix(pub, "/") + "/" + rel
}

// mountAt serves inner at prefix (e.g. /fs); redirects prefix without slash to prefix/.
func mountAt(prefix string, inner http.Handler) http.Handler {
	if prefix == "" {
		return inner
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == prefix {
			http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
			return
		}
		http.StripPrefix(prefix, inner).ServeHTTP(w, r)
	})
}

func safeJoin(root, rel string) (string, bool) {
	root = filepath.Clean(root)
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || rel == "." {
		return root, true
	}
	rel = filepath.FromSlash(rel)
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", false
	}
	full := filepath.Join(root, rel)
	rc := filepath.Clean(root)
	fc := filepath.Clean(full)
	if fc == rc {
		return fc, true
	}
	sep := string(os.PathSeparator)
	if strings.HasPrefix(fc, rc+sep) {
		return fc, true
	}
	return "", false
}

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonWrite(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

func sanitizeFilename(s string) string {
	s = filepath.Base(s)
	if s == "." || s == ".." {
		return "file"
	}
	return s
}

// streamMP3Handler serves files under DATA_ROOT/assets/mp3 with inline audio/mpeg (for <audio> playback; /files/ forces attachment).
func streamMP3Handler(dataRoot string) http.HandlerFunc {
	root := docAssetsMp3Dir(dataRoot)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !requireMp3Gate(w, r) {
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/stream/mp3/")
		rel = filepath.ToSlash(rel)
		full, ok := safeJoin(root, rel)
		if !ok {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			http.NotFound(w, r)
			return
		}
		ext := strings.ToLower(filepath.Ext(full))
		if ext != ".mp3" {
			http.Error(w, "not an mp3 file", http.StatusBadRequest)
			return
		}
		track := filepath.Base(full)
		if !allowMp3TrackRL(w, r, track) {
			return
		}
		ip := clientIP(r)
		if !acquireMp3StreamSlot(ip) {
			http.Error(w, "too many concurrent mp3 streams", http.StatusTooManyRequests)
			return
		}
		defer releaseMp3StreamSlot(ip)
		w.Header().Set("Content-Type", "audio/mpeg")
		// private：门禁 Cookie 下不宜 public/CDN 共享缓存；拉长 max-age 利于同浏览器回访命中。
		w.Header().Set("Cache-Control", "private, max-age=86400")
		http.ServeFile(w, r, full)
	}
}

func fileDownloadHandler(dataRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/files/")
		rel = filepath.ToSlash(rel)
		if browseRelBlockedForFiles(rel) {
			http.NotFound(w, r)
			return
		}
		full, ok := safeJoin(dataRoot, rel)
		if !ok {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			http.NotFound(w, r)
			return
		}
		fn := filepath.Base(full)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="`+escapeFilename(fn)+`"`)
		http.ServeFile(w, r, full)
	}
}

func escapeFilename(s string) string {
	return strings.ReplaceAll(s, `"`, `'`)
}

var maxPreviewBytes int64 = 4 << 20 // 4MB

func previewHandler(dataRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/preview/")
		rel = filepath.ToSlash(rel)
		if browseRelBlockedForFiles(rel) {
			http.NotFound(w, r)
			return
		}
		full, ok := safeJoin(dataRoot, rel)
		if !ok {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			http.NotFound(w, r)
			return
		}
		kind, lang, ok := previewKind(filepath.Base(full))
		if !ok {
			http.Error(w, "preview not supported for this file type", http.StatusBadRequest)
			return
		}
		if st.Size() > maxPreviewBytes {
			http.Error(w, "file too large for preview (max 4MB)", http.StatusRequestEntityTooLarge)
			return
		}
		raw, err := os.ReadFile(full)
		if err != nil {
			http.Error(w, "read failed", http.StatusInternalServerError)
			return
		}
		if bytes.IndexByte(raw, 0) >= 0 {
			http.Error(w, "binary file cannot be previewed", http.StatusUnsupportedMediaType)
			return
		}

		pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
		var bodyHTML []byte
		if kind == "md" {
			// 预览页没有文章标题栏，保留正文首个 H1；代码块展示与文章一致。
			rendered, err := markdown.Render(raw, markdown.RenderOpts{
				PicBaseURL: docPicPublicBase(pub) + "/",
			})
			if err != nil {
				http.Error(w, "markdown render failed", http.StatusInternalServerError)
				return
			}
			var b strings.Builder
			b.Grow(len(rendered) + 40)
			b.WriteString(`<div class="doc-md">`)
			b.Write(rendered)
			b.WriteString(`</div>`)
			bodyHTML = []byte(b.String())
		} else {
			var b strings.Builder
			b.WriteString(`<div class="doc-md"><pre><code`)
			if lang != "" {
				b.WriteString(` class="language-`)
				b.WriteString(html.EscapeString(lang))
				b.WriteString(`"`)
			}
			b.WriteString(`>`)
			b.WriteString(html.EscapeString(string(raw)))
			b.WriteString(`</code></pre></div>`)
			bodyHTML = []byte(b.String())
		}

		fname := html.EscapeString(filepath.Base(full))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "private, max-age=60")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>%s</style>
</head>
<body>
<div class="wrap">
<p class="preview-fname">%s</p>
`, fname, filePreviewPageCSS(), fname)
		_, _ = w.Write(bodyHTML)
		io.WriteString(w, "\n</div>\n")
		io.WriteString(w, docArticleBodyScripts(pub, bodyHTML))
		io.WriteString(w, "\n</body>\n</html>")
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

// publicBaseFromRequest is the URL prefix shown in HTML/JS (e.g. /fs). Env BASE_PATH wins;
// else X-Forwarded-Prefix from the reverse proxy (strip-prefix mode: internal path is / but browser uses /fs/…).
func publicBaseFromRequest(r *http.Request, envBase string) string {
	if s := strings.TrimSuffix(strings.TrimSpace(envBase), "/"); s != "" {
		return s
	}
	h := strings.TrimSpace(r.Header.Get("X-Forwarded-Prefix"))
	if h == "" {
		return ""
	}
	if !strings.HasPrefix(h, "/") {
		return ""
	}
	h = path.Clean(h)
	if h == "/" || h == "." {
		return ""
	}
	return h
}

const faviconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" shape-rendering="crispEdges">
  <rect width="16" height="16" fill="#1a2d1a"/>
  <rect x="5" y="2" width="6" height="2" fill="#c45c3a"/>
  <rect x="4" y="4" width="8" height="3" fill="#4a7a4a"/>
  <rect x="5" y="7" width="2" height="4" fill="#3d5c8a"/>
  <rect x="9" y="7" width="2" height="4" fill="#3d5c8a"/>
  <rect x="6" y="7" width="4" height="2" fill="#5a8a5a"/>
  <rect x="6" y="9" width="4" height="3" fill="#2d4a2d"/>
  <rect x="3" y="5" width="2" height="1" fill="#8a5a3a"/>
  <rect x="11" y="5" width="3" height="1" fill="#5a5a5a"/>
  <rect x="12" y="6" width="2" height="1" fill="#3a3a3a"/>
  <rect x="6" y="12" width="1" height="2" fill="#2a3d5a"/>
  <rect x="9" y="12" width="1" height="2" fill="#2a3d5a"/>
  <rect x="5" y="14" width="2" height="1" fill="#1a1a1a"/>
  <rect x="9" y="14" width="2" height="1" fill="#1a1a1a"/>
</svg>`
