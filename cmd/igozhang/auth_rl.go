package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	authFailMax       = 5
	authFailWindow    = time.Minute
	authFailLockAfter = time.Minute
	authRLFile        = "auth-rl.json"
	authRLLegacyFile  = "ops-auth-rl.json"
)

type authFailDiskEntry struct {
	FailAt      []int64 `json:"failAt,omitempty"`
	LockedUntil int64   `json:"lockedUntil,omitempty"`
}

type authFailState struct {
	failTimes   []time.Time
	lockedUntil time.Time
}

var (
	authFailRLMu sync.Mutex
	authFailRL   = make(map[string]*authFailState)
	authRLRoot   string
)

func initAuthFailRL(dataRoot string) {
	authFailRLMu.Lock()
	defer authFailRLMu.Unlock()
	authRLRoot = dataRoot
	loadAuthFailRLLocked(dataRoot)
}

func authRLPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), authRLFile)
}

func authRLLegacyPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), authRLLegacyFile)
}

func loadAuthFailRLLocked(dataRoot string) {
	if dataRoot == "" {
		return
	}
	path := authRLPath(dataRoot)
	raw, err := os.ReadFile(path)
	if err != nil {
		legacy := authRLLegacyPath(dataRoot)
		raw, err = os.ReadFile(legacy)
		if err != nil {
			return
		}
		// migrate: rewrite under new name when possible
		_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
		tmp := path + ".tmp"
		if werr := os.WriteFile(tmp, raw, 0o600); werr == nil {
			if rerr := os.Rename(tmp, path); rerr == nil {
				_ = os.Remove(legacy)
			} else {
				_ = os.Remove(tmp)
			}
		}
	}
	var disk map[string]authFailDiskEntry
	if err := json.Unmarshal(raw, &disk); err != nil {
		return
	}
	now := time.Now()
	cutoff := now.Add(-authFailWindow).Unix()
	for ip, e := range disk {
		st := &authFailState{}
		if e.LockedUntil > 0 {
			st.lockedUntil = time.Unix(e.LockedUntil, 0)
		}
		for _, ts := range e.FailAt {
			if ts >= cutoff {
				st.failTimes = append(st.failTimes, time.Unix(ts, 0))
			}
		}
		if st.lockedUntil.After(now) || len(st.failTimes) > 0 {
			authFailRL[ip] = st
		}
	}
}

func persistAuthFailRLLocked(dataRoot string) {
	if dataRoot == "" {
		return
	}
	now := time.Now()
	cutoff := now.Add(-authFailWindow).Unix()
	disk := make(map[string]authFailDiskEntry, len(authFailRL))
	for ip, st := range authFailRL {
		if st == nil {
			continue
		}
		e := authFailDiskEntry{}
		if st.lockedUntil.After(now) {
			e.LockedUntil = st.lockedUntil.Unix()
		}
		for _, t := range st.failTimes {
			if ts := t.Unix(); ts >= cutoff {
				e.FailAt = append(e.FailAt, ts)
			}
		}
		if e.LockedUntil > 0 || len(e.FailAt) > 0 {
			disk[ip] = e
		}
	}
	path := authRLPath(dataRoot)
	if len(disk) == 0 {
		_ = os.Remove(path)
		return
	}
	b, err := json.Marshal(disk)
	if err != nil {
		return
	}
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

func clientIPForRL(r *http.Request) string {
	ip := clientIP(r)
	if ip != "" {
		return ip
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}

func writeAuthRateLimited(w http.ResponseWriter, retrySec int) {
	if retrySec < 1 {
		retrySec = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(retrySec))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":         "验证尝试过于频繁，请稍后再试",
		"retryAfterSec": retrySec,
	})
}

func checkAuthFailRL(w http.ResponseWriter, r *http.Request, dataRoot string) bool {
	ip := clientIPForRL(r)
	now := time.Now()
	authFailRLMu.Lock()
	if authRLRoot == "" && dataRoot != "" {
		authRLRoot = dataRoot
		loadAuthFailRLLocked(dataRoot)
	}
	st := authFailRL[ip]
	if st != nil && !st.lockedUntil.IsZero() && now.Before(st.lockedUntil) {
		sec := int(time.Until(st.lockedUntil).Seconds())
		authFailRLMu.Unlock()
		writeAuthRateLimited(w, sec)
		return false
	}
	if st != nil && !st.lockedUntil.IsZero() && !now.Before(st.lockedUntil) {
		delete(authFailRL, ip)
	}
	authFailRLMu.Unlock()
	return true
}

// recordAuthFail 记录一次失败；若触发锁定则返回 true 与剩余秒数。
func recordAuthFail(r *http.Request, dataRoot string) (locked bool, retrySec int) {
	ip := clientIPForRL(r)
	now := time.Now()
	cutoff := now.Add(-authFailWindow)
	authFailRLMu.Lock()
	defer authFailRLMu.Unlock()
	if authRLRoot == "" && dataRoot != "" {
		authRLRoot = dataRoot
		loadAuthFailRLLocked(dataRoot)
	}
	st := authFailRL[ip]
	if st == nil {
		st = &authFailState{}
		authFailRL[ip] = st
	}
	ft := make([]time.Time, 0, len(st.failTimes))
	for _, t := range st.failTimes {
		if t.After(cutoff) {
			ft = append(ft, t)
		}
	}
	ft = append(ft, now)
	st.failTimes = ft
	if len(st.failTimes) >= authFailMax {
		st.lockedUntil = now.Add(authFailLockAfter)
		st.failTimes = nil
		retrySec = int(authFailLockAfter.Seconds())
		persistAuthFailRLLocked(dataRoot)
		return true, retrySec
	}
	persistAuthFailRLLocked(dataRoot)
	return false, 0
}

func clearAuthFail(r *http.Request, dataRoot string) {
	ip := clientIPForRL(r)
	authFailRLMu.Lock()
	delete(authFailRL, ip)
	persistAuthFailRLLocked(dataRoot)
	authFailRLMu.Unlock()
}
