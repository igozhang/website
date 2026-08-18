package main

import (
	"net/http"
	"strings"
)

// etagNotModified 判断客户端 If-None-Match 是否允许对带引号的强 ETag 返回 304。
func etagNotModified(r *http.Request, etag string) bool {
	if r == nil {
		return false
	}
	inm := strings.TrimSpace(r.Header.Get("If-None-Match"))
	if inm == "" {
		return false
	}
	if inm == "*" {
		return true
	}
	want := strings.Trim(strings.TrimSpace(etag), `"`)
	for _, raw := range strings.Split(inm, ",") {
		part := strings.TrimSpace(raw)
		part = strings.TrimPrefix(part, "W/")
		part = strings.TrimSpace(part)
		if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
			part = part[1 : len(part)-1]
		}
		if part == want {
			return true
		}
	}
	return false
}
