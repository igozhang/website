package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	docExportCacheName     = "doc-export.zip"
	docExportCacheMetaName = "doc-export.meta.json"
	docBackupDebounce      = 3 * time.Second
	docExportMemMaxBytes   = 28 << 20 // 导出 ZIP 内存预热上限（超过则仅 hit-disk）
	docExportWaitPoll      = 200 * time.Millisecond
	docExportWaitTimeout   = 2 * time.Minute
)

type docExportCacheMeta struct {
	ExportedAt string `json:"exportedAt"`
	Count      int    `json:"count"`
	Bytes      int    `json:"bytes"`
}

var docBackupMu sync.Mutex
var docBackupTimer *time.Timer
var docBackupDataRoot string

var (
	docExportMemMu sync.RWMutex
	docExportMem   struct {
		dataRoot string
		bytes    []byte
	}
)

var (
	docExportBuildMu sync.Mutex
	docExportBuild   = make(map[string]bool)
)

func docExportDownloadName() string {
	return "doc-export-" + formatBeijingCompact(time.Now().Unix()) + ".zip"
}

func docExportCachePath(dataRoot string) string {
	return filepath.Join(dataRoot, stashRelDir, docExportCacheName)
}

func docExportCacheMetaPath(dataRoot string) string {
	return filepath.Join(dataRoot, stashRelDir, docExportCacheMetaName)
}

func clearDocExportMemCache(dataRoot string) {
	docExportMemMu.Lock()
	defer docExportMemMu.Unlock()
	if dataRoot == "" || docExportMem.dataRoot == dataRoot {
		docExportMem.dataRoot = ""
		docExportMem.bytes = nil
	}
}

func setDocExportMemCache(dataRoot string, zipBytes []byte) {
	if dataRoot == "" || len(zipBytes) == 0 || len(zipBytes) > docExportMemMaxBytes {
		return
	}
	dup := make([]byte, len(zipBytes))
	copy(dup, zipBytes)
	docExportMemMu.Lock()
	docExportMem.dataRoot = dataRoot
	docExportMem.bytes = dup
	docExportMemMu.Unlock()
}

func getDocExportMemCache(dataRoot string) ([]byte, bool) {
	docExportMemMu.RLock()
	defer docExportMemMu.RUnlock()
	if dataRoot == "" || docExportMem.dataRoot != dataRoot || len(docExportMem.bytes) == 0 {
		return nil, false
	}
	return docExportMem.bytes, true
}

func invalidateDocExportCache(dataRoot string) {
	clearDocExportMemCache(dataRoot)
	_ = os.Remove(docExportCachePath(dataRoot))
	_ = os.Remove(docExportCacheMetaPath(dataRoot))
}

func docExportCacheFileReady(dataRoot string) bool {
	fi, err := os.Stat(docExportCachePath(dataRoot))
	return err == nil && fi.Size() > 0
}

func docExportBuildTryStart(dataRoot string) bool {
	if dataRoot == "" {
		return false
	}
	docExportBuildMu.Lock()
	defer docExportBuildMu.Unlock()
	if docExportBuild[dataRoot] {
		return false
	}
	docExportBuild[dataRoot] = true
	return true
}

func docExportBuildEnd(dataRoot string) {
	if dataRoot == "" {
		return
	}
	docExportBuildMu.Lock()
	delete(docExportBuild, dataRoot)
	docExportBuildMu.Unlock()
}

func docExportBuildActive(dataRoot string) bool {
	docExportBuildMu.Lock()
	defer docExportBuildMu.Unlock()
	return docExportBuild[dataRoot]
}

func waitForDocExportCache(dataRoot string, timeout time.Duration) bool {
	if dataRoot == "" {
		return false
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if docExportCacheFileReady(dataRoot) && !docExportBuildActive(dataRoot) {
			return true
		}
		time.Sleep(docExportWaitPoll)
	}
	return docExportCacheFileReady(dataRoot)
}

// triggerDocExportRebuild 后台重建导出缓存；同一 dataRoot 同时只跑一个任务。
func triggerDocExportRebuild(dataRoot string, prepareAccessLogs bool) {
	if dataRoot == "" || docStore == nil {
		return
	}
	if !docExportBuildTryStart(dataRoot) {
		return
	}
	go func() {
		defer docExportBuildEnd(dataRoot)
		if err := writeDocExportCacheWithOpts(dataRoot, prepareAccessLogs); err != nil {
			log.Printf("doc-export: rebuild: %v", err)
		}
	}()
}

// preloadDocExportMemCache 进程启动后异步把磁盘缓存读入内存（已有 doc-export.zip 时）。
func preloadDocExportMemCache(dataRoot string) {
	if dataRoot == "" {
		return
	}
	go func() {
		raw, err := os.ReadFile(docExportCachePath(dataRoot))
		if err != nil || len(raw) == 0 {
			return
		}
		setDocExportMemCache(dataRoot, raw)
	}()
}

// scheduleDocBackup debounces background ZIP generation into secret-tempfile/.
// 不在调度时删除已有缓存：persist 原子替换磁盘前继续用旧包（内存+磁盘）。
func scheduleDocBackup(dataRoot string) {
	if dataRoot == "" || docStore == nil {
		return
	}
	docBackupMu.Lock()
	defer docBackupMu.Unlock()
	docBackupDataRoot = dataRoot
	if docBackupTimer != nil {
		docBackupTimer.Stop()
	}
	docBackupTimer = time.AfterFunc(docBackupDebounce, runDocBackup)
}

func runDocBackup() {
	docBackupMu.Lock()
	dataRoot := docBackupDataRoot
	docBackupMu.Unlock()
	if dataRoot == "" || docStore == nil {
		return
	}
	triggerDocExportRebuild(dataRoot, true)
}

func finalizeDocExportCacheTmp(dataRoot, zipTmp string, meta docExportCacheMeta) error {
	dir := filepath.Join(dataRoot, stashRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		_ = os.Remove(zipTmp)
		return err
	}
	mb, err := json.Marshal(meta)
	if err != nil {
		_ = os.Remove(zipTmp)
		return err
	}
	metaTmp := filepath.Join(dir, docExportCacheMetaName+".tmp")
	if err := os.WriteFile(metaTmp, mb, 0o600); err != nil {
		_ = os.Remove(zipTmp)
		return err
	}
	finalZip := docExportCachePath(dataRoot)
	if err := os.Rename(zipTmp, finalZip); err != nil {
		_ = os.Remove(zipTmp)
		_ = os.Remove(metaTmp)
		return err
	}
	if err := os.Rename(metaTmp, docExportCacheMetaPath(dataRoot)); err != nil {
		return err
	}
	if fi, err := os.Stat(finalZip); err == nil && fi.Size() > 0 && fi.Size() <= int64(docExportMemMaxBytes) {
		raw, err := os.ReadFile(finalZip)
		if err == nil {
			setDocExportMemCache(dataRoot, raw)
		}
	}
	return nil
}

func persistDocExportCache(dataRoot string, zipBytes []byte, meta docExportCacheMeta) error {
	dir := filepath.Join(dataRoot, stashRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	zipTmp := filepath.Join(dir, docExportCacheName+".tmp")
	if err := os.WriteFile(zipTmp, zipBytes, 0o600); err != nil {
		return err
	}
	meta.Bytes = len(zipBytes)
	return finalizeDocExportCacheTmp(dataRoot, zipTmp, meta)
}

func writeDocExportCache(dataRoot string) error {
	return writeDocExportCacheWithOpts(dataRoot, true)
}

func writeDocExportCacheWithOpts(dataRoot string, prepareAccessLogs bool) error {
	if dataRoot == "" || docStore == nil {
		return fmt.Errorf("doc export: no store")
	}
	dir := filepath.Join(dataRoot, stashRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	zipTmp := filepath.Join(dir, docExportCacheName+".tmp")
	f, err := os.Create(zipTmp)
	if err != nil {
		return err
	}
	assets := docAssetsRoot(dataRoot)
	buildErr := writeDocExportZip(f, dataRoot, assets, docStore, prepareAccessLogs)
	closeErr := f.Close()
	if buildErr != nil {
		_ = os.Remove(zipTmp)
		return buildErr
	}
	if closeErr != nil {
		_ = os.Remove(zipTmp)
		return closeErr
	}
	all, err := docStore.ListAll()
	if err != nil {
		_ = os.Remove(zipTmp)
		return err
	}
	fi, err := os.Stat(zipTmp)
	if err != nil {
		_ = os.Remove(zipTmp)
		return err
	}
	meta := docExportCacheMeta{
		ExportedAt: time.Now().Format(time.RFC3339),
		Count:      len(all),
		Bytes:      int(fi.Size()),
	}
	return finalizeDocExportCacheTmp(dataRoot, zipTmp, meta)
}

func readDocExportCache(dataRoot string) ([]byte, bool) {
	raw, err := os.ReadFile(docExportCachePath(dataRoot))
	if err != nil || len(raw) == 0 {
		return nil, false
	}
	return raw, true
}

func setDocExportResponseHeaders(w http.ResponseWriter, cacheStatus string) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+docExportDownloadName()+`"`)
	w.Header().Set("X-Doc-Export-Cache", cacheStatus)
}

// tryServeDocExportCache 优先内存预热包（hit-mem），否则流式读磁盘（hit-disk）。
func tryServeDocExportCache(w http.ResponseWriter, r *http.Request, dataRoot string) bool {
	if raw, ok := getDocExportMemCache(dataRoot); ok {
		setDocExportResponseHeaders(w, "hit-mem")
		w.Header().Set("Content-Length", strconv.Itoa(len(raw)))
		_, _ = w.Write(raw)
		return true
	}
	path := docExportCachePath(dataRoot)
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil || fi.Size() == 0 {
		return false
	}
	setDocExportResponseHeaders(w, "hit-disk")
	http.ServeContent(w, r, filepath.Base(path), fi.ModTime(), f)
	return true
}

// tryServeDocExportCacheStale 返回磁盘旧包并触发后台重建（hit-disk-stale）。
func tryServeDocExportCacheStale(w http.ResponseWriter, r *http.Request, dataRoot string) bool {
	path := docExportCachePath(dataRoot)
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil || fi.Size() == 0 {
		return false
	}
	setDocExportResponseHeaders(w, "hit-disk-stale")
	http.ServeContent(w, r, filepath.Base(path), fi.ModTime(), f)
	triggerDocExportRebuild(dataRoot, false)
	return true
}

// streamDocExportZipMiss 无缓存时流式生成 ZIP（同时写入临时文件供落盘），不阻塞其它请求整包进内存。
func streamDocExportZipMiss(w http.ResponseWriter, dataRoot string) error {
	assets := docAssetsRoot(dataRoot)
	dir := filepath.Join(dataRoot, stashRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	zipTmp := filepath.Join(dir, docExportCacheName+".tmp")
	tmpFile, err := os.Create(zipTmp)
	if err != nil {
		return err
	}

	setDocExportResponseHeaders(w, "miss")
	mw := io.MultiWriter(w, tmpFile)
	buildErr := writeDocExportZip(mw, dataRoot, assets, docStore, false)
	closeErr := tmpFile.Close()
	if buildErr != nil {
		_ = os.Remove(zipTmp)
		return buildErr
	}
	if closeErr != nil {
		_ = os.Remove(zipTmp)
		return closeErr
	}

	go func() {
		all, listErr := docStore.ListAll()
		if listErr != nil {
			log.Printf("doc-export: list: %v", listErr)
			_ = os.Remove(zipTmp)
			return
		}
		fi, statErr := os.Stat(zipTmp)
		if statErr != nil {
			_ = os.Remove(zipTmp)
			return
		}
		meta := docExportCacheMeta{
			ExportedAt: time.Now().Format(time.RFC3339),
			Count:      len(all),
			Bytes:      int(fi.Size()),
		}
		if err := finalizeDocExportCacheTmp(dataRoot, zipTmp, meta); err != nil {
			log.Printf("doc-export: cache finalize: %v", err)
			_ = os.Remove(zipTmp)
		}
	}()
	return nil
}
