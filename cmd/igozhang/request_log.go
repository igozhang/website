package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// shouldSkipHTTPRequestLog 运行日志中跳过高频、低价值的探针与静态资源请求。
func shouldSkipHTTPRequestLog(path string) bool {
	if path == "/health" {
		return true
	}
	if strings.HasPrefix(path, "/favicon") {
		return true
	}
	return false
}

// formatRequestLogDuration 使用纯 ASCII 时长，避免终端编码将 µ 显示为乱码。
func formatRequestLogDuration(d time.Duration) string {
	us := float64(d) / float64(time.Microsecond)
	if us < 1000 {
		return fmt.Sprintf("%.3fus", us)
	}
	ms := us / 1000
	if ms < 1000 {
		return fmt.Sprintf("%.3fms", ms)
	}
	return fmt.Sprintf("%.3fs", ms/1000)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		skip := shouldSkipHTTPRequestLog(path)
		start := time.Now()
		next.ServeHTTP(w, r)
		if skip {
			return
		}
		ip, geo := visitorGeoLogEN(r)
		log.Printf("%s %s %s ip=%s geo=%s", r.Method, path, formatRequestLogDuration(time.Since(start)), ip, geo)
	})
}
