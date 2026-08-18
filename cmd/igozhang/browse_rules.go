package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// isBrowsableDirName：目录浏览白名单，仅展示名称以字母 f 开头的文件夹（不区分大小写）。
func isBrowsableDirName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return n != "" && strings.HasPrefix(n, "f")
}

// browsePathAllowed 为 true 时允许列出该相对路径（路径中每一级目录名均须满足白名单）。
func browsePathAllowed(rel string) bool {
	rel = filepath.ToSlash(filepath.Clean(strings.TrimSpace(rel)))
	rel = strings.Trim(rel, "/")
	if rel == "" || rel == "." {
		return true
	}
	// MP3 播放器专用：仅允许列出 DATA_ROOT/assets/mp3（不在 /fs 文件浏览中展示）。
	if rel == "assets/mp3" {
		return true
	}
	for _, seg := range strings.Split(rel, "/") {
		if !isBrowsableDirName(seg) {
			return false
		}
	}
	return true
}

// browseRelBlockedForFiles 为 true 时禁止通过 /files、/preview 访问（路径中任一层目录不在白名单内）。
func browseRelBlockedForFiles(rel string) bool {
	rel = filepath.ToSlash(filepath.Clean(rel))
	rel = strings.Trim(rel, "/")
	if rel == "" || rel == "." {
		return false
	}
	parts := strings.Split(rel, "/")
	for i := 0; i < len(parts)-1; i++ {
		if !isBrowsableDirName(parts[i]) {
			return true
		}
	}
	return false
}

const uploadedFileMode = 0o600

func chmodUploadedFile(path string) error {
	return os.Chmod(path, uploadedFileMode)
}

// fileDeleteHandler 删除可浏览目录下的普通文件（需认证；不可删目录）。
func fileDeleteHandler(dataRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !requireAuthToken(w, r, dataRoot) {
			return
		}
		q := filepath.ToSlash(r.URL.Query().Get("path"))
		q = strings.Trim(filepath.ToSlash(filepath.Clean(strings.TrimSpace(q))), "/")
		if q == "" || q == "." {
			jsonError(w, http.StatusBadRequest, "path required")
			return
		}
		if browseRelBlockedForFiles(q) {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		parent := path.Dir(q)
		if parent == "." {
			parent = ""
		}
		if !browsePathAllowed(parent) {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		target, ok := safeJoin(dataRoot, q)
		if !ok {
			jsonError(w, http.StatusBadRequest, "invalid path")
			return
		}
		st, err := os.Lstat(target)
		if err != nil {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		// 仅删普通文件或符号链接本身；目录与其它特殊节点一律拒绝。
		if st.IsDir() || (!st.Mode().IsRegular() && st.Mode().Type()&os.ModeSymlink == 0) {
			jsonError(w, http.StatusBadRequest, "not a file")
			return
		}
		if err := os.Remove(target); err != nil {
			jsonError(w, http.StatusInternalServerError, "delete failed")
			return
		}
		invalidateTransferUsageCache()
		jsonWrite(w, map[string]any{"ok": true, "deleted": q})
	}
}
