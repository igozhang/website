package main

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	usageCacheMu   sync.Mutex
	usageCacheSum  int64
	usageCacheAt   time.Time
	usageCachePath string
)

func sumTransferDirBytes(transferAbs string) int64 {
	usageCacheMu.Lock()
	defer usageCacheMu.Unlock()
	if usageCachePath == transferAbs && time.Since(usageCacheAt) < 45*time.Second {
		return usageCacheSum
	}
	var sum int64
	_ = filepath.Walk(transferAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		sum += info.Size()
		return nil
	})
	usageCachePath = transferAbs
	usageCacheSum = sum
	usageCacheAt = time.Now()
	return sum
}

func invalidateTransferUsageCache() {
	usageCacheMu.Lock()
	usageCacheAt = time.Time{}
	usageCacheMu.Unlock()
}

func fsUsageHandler(dataRoot, transferAbs string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !requireAuthToken(w, r, dataRoot) {
			return
		}
		used := sumTransferDirBytes(transferAbs)
		op := getOpsSnapshot()
		jsonWrite(w, map[string]any{
			"bytesUsed":       used,
			"baselineBytes":   usageBaselineBytes(),
			"capBytes":        transferDirMaxBytes,
			"fsFileMaxAgeSec": int64(fsFileMaxAge.Seconds()),
			"stashTTLHours":   int(stashTTL.Hours()),
			"stashMaxBytes":   maxStashBytes,
			"stashPollMs":     int(op.stashPollInterval.Milliseconds()),
			"stashPollMaxMs":  int(op.stashPollIntervalMax.Milliseconds()),
		})
	}
}
