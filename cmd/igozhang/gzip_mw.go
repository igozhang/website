package main

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"
)

const gzipMinBytes = 1024

// gzipHandler 对文本类响应做 gzip（应用层，不依赖 Ingress ConfigMap）。
// 跳过：客户端不接受 gzip、已有 Content-Encoding、过小 body、音视频路径、不可压缩类型。
func gzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}
		if !acceptsGzip(r.Header.Get("Accept-Encoding")) {
			next.ServeHTTP(w, r)
			return
		}
		path := r.URL.Path
		if skipGzipPath(path) {
			next.ServeHTTP(w, r)
			return
		}
		bw := &gzipBufferWriter{ResponseWriter: w, status: http.StatusOK}
		defer bw.flush()
		next.ServeHTTP(bw, r)
	})
}

// skipGzipPath 跳过音视频/附件；兼容 BASE_PATH 前缀（如 /app/stream/mp3/…）。
func skipGzipPath(path string) bool {
	return strings.Contains(path, "/stream/mp3/") || strings.Contains(path, "/files/")
}

func acceptsGzip(ae string) bool {
	ae = strings.ToLower(ae)
	for _, part := range strings.Split(ae, ",") {
		part = strings.TrimSpace(part)
		if i := strings.IndexByte(part, ';'); i >= 0 {
			part = strings.TrimSpace(part[:i])
		}
		if part == "gzip" {
			return true
		}
	}
	return false
}

func gzipCompressibleType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	switch ct {
	case "text/html", "text/css", "text/plain", "text/javascript",
		"application/javascript", "application/json", "application/xml",
		"image/svg+xml", "application/manifest+json":
		return true
	default:
		return strings.HasPrefix(ct, "text/")
	}
}

type gzipBufferWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	status      int
	wroteHeader bool
}

func (g *gzipBufferWriter) WriteHeader(code int) {
	if g.wroteHeader {
		return
	}
	g.status = code
	g.wroteHeader = true
}

func (g *gzipBufferWriter) Write(p []byte) (int, error) {
	if !g.wroteHeader {
		g.WriteHeader(http.StatusOK)
	}
	return g.buf.Write(p)
}

func (g *gzipBufferWriter) flush() {
	body := g.buf.Bytes()
	hdr := g.ResponseWriter.Header()
	ct := hdr.Get("Content-Type")
	ce := hdr.Get("Content-Encoding")

	useGzip := ce == "" &&
		g.status >= 200 && g.status < 300 &&
		len(body) >= gzipMinBytes &&
		gzipCompressibleType(ct)

	if !useGzip {
		g.ResponseWriter.WriteHeader(g.status)
		if len(body) > 0 {
			_, _ = g.ResponseWriter.Write(body)
		}
		return
	}

	var zbuf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&zbuf, gzip.DefaultCompression)
	if err != nil {
		hdr.Del("Content-Encoding")
		g.ResponseWriter.WriteHeader(g.status)
		_, _ = g.ResponseWriter.Write(body)
		return
	}
	if _, err := zw.Write(body); err != nil {
		_ = zw.Close()
		hdr.Del("Content-Encoding")
		g.ResponseWriter.WriteHeader(g.status)
		_, _ = g.ResponseWriter.Write(body)
		return
	}
	if err := zw.Close(); err != nil {
		hdr.Del("Content-Encoding")
		g.ResponseWriter.WriteHeader(g.status)
		_, _ = g.ResponseWriter.Write(body)
		return
	}

	hdr.Set("Content-Encoding", "gzip")
	hdr.Del("Content-Length")
	hdr.Add("Vary", "Accept-Encoding")
	g.ResponseWriter.WriteHeader(g.status)
	_, _ = g.ResponseWriter.Write(zbuf.Bytes())
}
