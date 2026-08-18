package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	maxStashBytes       = 200 * 1024
	stashTTL            = 12 * time.Hour
	fsFileMaxAge        = 7 * 24 * time.Hour
	transferDirMaxBytes = int64(9) * 1024 * 1024 * 1024 // 9 GiB
)

type transferFileEnt struct {
	path string
	size int64
}

// stashRelDir 位于 DATA_ROOT 下（目录名不以 f 开头，不出现在目录浏览白名单列表中）；仅供服务端文本暂存。
const stashRelDir = "secret-tempfile"

const legacyStashRelDir = "secret01"

type stashDoc struct {
	Text      string `json:"text"`
	UpdatedAt int64  `json:"updatedAt"`
}

var stashMu sync.Mutex

func jsonConflictStash(w http.ResponseWriter, updatedAt int64, text string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusConflict)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(map[string]any{
		"error":           "stash conflict",
		"serverUpdatedAt": updatedAt,
		"serverText":      text,
	})
}

func stashFilePath(dataRoot string) string {
	return filepath.Join(dataRoot, stashRelDir, "stash.json")
}

// migrateLegacyStash 将旧版暂存迁到 secret-tempfile/（仅当新路径尚不存在时）：优先 secret01/，其次 .stash/。
func migrateLegacyStash(dataRoot string) {
	newPath := stashFilePath(dataRoot)
	if _, err := os.Stat(newPath); err == nil {
		return
	}
	sources := []string{
		filepath.Join(dataRoot, legacyStashRelDir, "stash.json"),
		filepath.Join(dataRoot, ".stash", "stash.json"),
	}
	for _, oldPath := range sources {
		data, err := os.ReadFile(oldPath)
		if err != nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
			return
		}
		if err := os.WriteFile(newPath, data, 0o600); err != nil {
			return
		}
		_ = os.Remove(oldPath)
		if strings.Contains(oldPath, legacyStashRelDir) {
			legacyDir := filepath.Join(dataRoot, legacyStashRelDir)
			if ents, err := os.ReadDir(legacyDir); err == nil && len(ents) == 0 {
				_ = os.Remove(legacyDir)
			}
		} else {
			_ = os.Remove(filepath.Join(dataRoot, ".stash"))
		}
		return
	}
}

func readStashForAPI(dataRoot string) (stashDoc, error) {
	p := stashFilePath(dataRoot)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return stashDoc{}, nil
		}
		return stashDoc{}, err
	}
	var d stashDoc
	if err := json.Unmarshal(data, &d); err != nil {
		return stashDoc{}, err
	}
	if d.UpdatedAt > 0 && time.Since(time.Unix(d.UpdatedAt, 0)) > stashTTL {
		_ = os.Remove(p)
		return stashDoc{}, nil
	}
	return d, nil
}

func stashETag(d stashDoc) string {
	return `"stash-v1-` + strconv.FormatInt(d.UpdatedAt, 10) + `"`
}

func stashHandler(dataRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireAuthToken(w, r, dataRoot) {
			return
		}
		if !allowStashRL(w, r) {
			return
		}
		switch r.Method {
		case http.MethodGet:
			stashMu.Lock()
			d, err := readStashForAPI(dataRoot)
			stashMu.Unlock()
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "read stash")
				return
			}
			etag := stashETag(d)
			w.Header().Set("ETag", etag)
			if etagNotModified(r, etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			jsonWrite(w, map[string]any{"text": d.Text, "updatedAt": d.UpdatedAt})
		case http.MethodPost:
			var body struct {
				Text             string `json:"text"`
				IfMatchUpdatedAt int64  `json:"ifMatchUpdatedAt"`
			}
			if err := json.NewDecoder(io.LimitReader(r.Body, int64(maxStashBytes)+65536)).Decode(&body); err != nil {
				jsonError(w, http.StatusBadRequest, "bad json")
				return
			}
			if len(body.Text) > maxStashBytes {
				jsonError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("text exceeds %d bytes", maxStashBytes))
				return
			}
			stashMu.Lock()
			cur, rerr := readStashForAPI(dataRoot)
			if rerr != nil {
				stashMu.Unlock()
				jsonError(w, http.StatusInternalServerError, "read stash")
				return
			}
			if body.IfMatchUpdatedAt != 0 && cur.UpdatedAt != 0 && body.IfMatchUpdatedAt != cur.UpdatedAt {
				jsonConflictStash(w, cur.UpdatedAt, cur.Text)
				stashMu.Unlock()
				return
			}
			err := writeStashFile(dataRoot, body.Text)
			stashMu.Unlock()
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "save stash")
				return
			}
			jsonWrite(w, map[string]any{"ok": true, "updatedAt": time.Now().Unix()})
		case http.MethodDelete:
			stashMu.Lock()
			err := os.Remove(stashFilePath(dataRoot))
			stashMu.Unlock()
			if err != nil && !os.IsNotExist(err) {
				jsonError(w, http.StatusInternalServerError, "clear stash")
				return
			}
			jsonWrite(w, map[string]any{"ok": true})
		default:
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func stashExportHandler(dataRoot, transferAbs, transferRel string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !requireAuthToken(w, r, dataRoot) {
			return
		}
		stashMu.Lock()
		d, err := readStashForAPI(dataRoot)
		stashMu.Unlock()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "read stash")
			return
		}
		if strings.TrimSpace(d.Text) == "" {
			jsonError(w, http.StatusBadRequest, "stash empty")
			return
		}
		name := fmt.Sprintf("stash-export-%d.txt", time.Now().Unix())
		dest := filepath.Join(transferAbs, name)
		if err := os.WriteFile(dest, []byte(d.Text), uploadedFileMode); err != nil {
			jsonError(w, http.StatusInternalServerError, "export failed")
			return
		}
		jsonWrite(w, map[string]any{
			"saved": filepath.ToSlash(filepath.Join(transferRel, name)),
			"bytes": len(d.Text),
		})
	}
}

func writeStashFile(dataRoot, text string) error {
	p := stashFilePath(dataRoot)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	d := stashDoc{Text: text, UpdatedAt: time.Now().Unix()}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

func startHourlyMaintenance(ctx context.Context, dataRoot, transferAbs string) {
	go func() {
		runMaintenance(dataRoot, transferAbs)
		for {
			interval := getOpsSnapshot().maintenanceInterval
			if interval < time.Hour {
				interval = time.Hour
			}
			timer := time.NewTimer(interval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				runMaintenance(dataRoot, transferAbs)
			}
		}
	}()
}

func runMaintenance(dataRoot, transferAbs string) {
	maintainPathProbeBan(dataRoot)

	stashMu.Lock()
	p := stashFilePath(dataRoot)
	if data, err := os.ReadFile(p); err == nil {
		var d stashDoc
		if json.Unmarshal(data, &d) == nil && d.UpdatedAt > 0 {
			if time.Since(time.Unix(d.UpdatedAt, 0)) > stashTTL {
				if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
					log.Printf("maintenance: stash remove: %v", err)
				} else {
					log.Printf("maintenance: delete reason=stash_ttl path=%s", p)
				}
			}
		}
	}
	stashMu.Unlock()

	cutoff := time.Now().Add(-fsFileMaxAge)
	var ents []transferFileEnt
	var sum int64
	_ = filepath.Walk(transferAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				log.Printf("maintenance: remove %s: %v", path, err)
			} else {
				log.Printf("maintenance: delete reason=file_ttl_age path=%s size=%d", path, info.Size())
			}
			return nil
		}
		sz := info.Size()
		ents = append(ents, transferFileEnt{path: path, size: sz})
		sum += sz
		return nil
	})
	if sum <= transferDirMaxBytes {
		maintainSecretLogDir(dataRoot)
		maintainAccessLogDir(dataRoot)
		return
	}
	sort.Slice(ents, func(i, j int) bool { return ents[i].size > ents[j].size })
	for _, e := range ents {
		if sum <= transferDirMaxBytes {
			break
		}
		if err := os.Remove(e.path); err != nil {
			log.Printf("maintenance: size cap remove %s: %v", e.path, err)
			continue
		}
		sum -= e.size
		log.Printf("maintenance: delete reason=dir_size_cap path=%s size=%d bytes_remaining=%d", e.path, e.size, sum)
	}
	maintainSecretLogDir(dataRoot)
	maintainAccessLogDir(dataRoot)
	invalidateTransferUsageCache()
}
