package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const pathProbeBanFile = "path_probe"

var (
	pathProbeBanMu   sync.Mutex
	pathProbeBanRoot string
	pathProbeStates  = make(map[string]*pathProbeState)
	pathProbeBans    = make(map[string]pathProbeBanEntry)
)

type pathProbeEvent struct {
	path string
	at   time.Time
}

type pathProbeState struct {
	events []pathProbeEvent
}

type pathProbeBanEntry struct {
	until    time.Time
	reason   string
	paths    int
	bannedAt time.Time
}

type pathProbeBanDiskEntry struct {
	Until    int64  `json:"until"`
	Reason   string `json:"reason,omitempty"`
	Paths    int    `json:"paths,omitempty"`
	BannedAt int64  `json:"bannedAt,omitempty"`
}

func initPathProbeBan(dataRoot string) {
	pathProbeBanMu.Lock()
	defer pathProbeBanMu.Unlock()
	pathProbeBanRoot = dataRoot
	pathProbeStates = make(map[string]*pathProbeState)
	pathProbeBans = make(map[string]pathProbeBanEntry)
	loadPathProbeBansLocked(dataRoot)
}

func pathProbeBanPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), "ip-auto-ban.json")
}

func loadPathProbeBansLocked(dataRoot string) {
	if dataRoot == "" {
		return
	}
	raw, err := os.ReadFile(pathProbeBanPath(dataRoot))
	if err != nil {
		return
	}
	var disk map[string]pathProbeBanDiskEntry
	if err := json.Unmarshal(raw, &disk); err != nil {
		return
	}
	now := time.Now()
	for ip, e := range disk {
		ip, ok := normalizeBlacklistIP(ip)
		if !ok || e.Until <= 0 {
			continue
		}
		until := time.Unix(e.Until, 0)
		if !until.After(now) {
			continue
		}
		entry := pathProbeBanEntry{
			until:  until,
			reason: strings.TrimSpace(e.Reason),
			paths:  e.Paths,
		}
		if e.BannedAt > 0 {
			entry.bannedAt = time.Unix(e.BannedAt, 0)
		}
		if entry.reason == "" {
			entry.reason = pathProbeBanFile
		}
		pathProbeBans[ip] = entry
	}
}

func persistPathProbeBansLocked(dataRoot string) {
	if dataRoot == "" {
		return
	}
	now := time.Now()
	disk := make(map[string]pathProbeBanDiskEntry, len(pathProbeBans))
	for ip, e := range pathProbeBans {
		if !e.until.After(now) {
			continue
		}
		de := pathProbeBanDiskEntry{
			Until:  e.until.Unix(),
			Reason: e.reason,
			Paths:  e.paths,
		}
		if !e.bannedAt.IsZero() {
			de.BannedAt = e.bannedAt.Unix()
		}
		disk[ip] = de
	}
	path := pathProbeBanPath(dataRoot)
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

func pathProbeConfig() (enabled bool, window time.Duration, maxPaths int, banDur time.Duration, exclude []string) {
	s := getOpsSnapshot()
	enabled = s.pathProbeEnabled
	window = s.pathProbeWindow
	maxPaths = s.pathProbeMaxPaths
	banDur = s.pathProbeBanDuration
	exclude = append([]string(nil), s.pathProbeExcludePrefixes...)
	if window <= 0 {
		window = time.Minute
	}
	if maxPaths <= 0 {
		maxPaths = 20
	}
	if banDur <= 0 {
		banDur = time.Hour
	}
	return enabled, window, maxPaths, banDur, exclude
}

func pathProbeNormalizePath(path string) string {
	p := strings.ToLower(visitorNormalizePath(path))
	if p == "" {
		return "/"
	}
	return p
}

func pathProbeExcluded(path string, exclude []string) bool {
	p := pathProbeNormalizePath(path)
	if p == "/health" {
		return true
	}
	if strings.HasPrefix(p, "/favicon") {
		return true
	}
	if strings.HasPrefix(p, "/api/ops/") || p == "/api/ops" {
		return true
	}
	if strings.HasPrefix(p, "/api/auth/") || p == "/api/auth" {
		return true
	}
	for _, prefix := range exclude {
		prefix = strings.ToLower(strings.TrimSpace(prefix))
		if prefix == "" {
			continue
		}
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

func isIPAutoBanned(ip string) bool {
	ip, ok := normalizeBlacklistIP(ip)
	if !ok {
		return false
	}
	now := time.Now()
	pathProbeBanMu.Lock()
	defer pathProbeBanMu.Unlock()
	e, ok := pathProbeBans[ip]
	if !ok {
		return false
	}
	if e.until.After(now) {
		return true
	}
	delete(pathProbeBans, ip)
	return false
}

func recordPathProbe(ip, path, dataRoot string) (banned bool) {
	ip, ok := normalizeBlacklistIP(ip)
	if !ok {
		return false
	}
	enabled, window, maxPaths, banDur, exclude := pathProbeConfig()
	if !enabled {
		return false
	}
	if pathProbeExcluded(path, exclude) {
		return false
	}
	npath := pathProbeNormalizePath(path)
	now := time.Now()
	cutoff := now.Add(-window)

	pathProbeBanMu.Lock()
	defer pathProbeBanMu.Unlock()
	if pathProbeBanRoot == "" && dataRoot != "" {
		pathProbeBanRoot = dataRoot
		loadPathProbeBansLocked(dataRoot)
	}
	if e, ok := pathProbeBans[ip]; ok && e.until.After(now) {
		return true
	}

	st := pathProbeStates[ip]
	if st == nil {
		st = &pathProbeState{}
		pathProbeStates[ip] = st
	}
	kept := st.events[:0]
	seen := make(map[string]struct{}, len(st.events)+1)
	for _, ev := range st.events {
		if ev.at.Before(cutoff) {
			continue
		}
		kept = append(kept, ev)
		seen[ev.path] = struct{}{}
	}
	if _, dup := seen[npath]; !dup {
		kept = append(kept, pathProbeEvent{path: npath, at: now})
		seen[npath] = struct{}{}
	}
	st.events = kept

	if len(seen) < maxPaths {
		return false
	}

	entry := pathProbeBanEntry{
		until:    now.Add(banDur),
		reason:   pathProbeBanFile,
		paths:    len(seen),
		bannedAt: now,
	}
	pathProbeBans[ip] = entry
	delete(pathProbeStates, ip)
	persistPathProbeBansLocked(dataRoot)
	log.Printf("auto-ban: ip=%s reason=path_probe paths=%d until=%s", ip, entry.paths, entry.until.Format(time.RFC3339))
	return true
}

func maintainPathProbeBan(dataRoot string) {
	now := time.Now()
	cutoff := now.Add(-time.Hour)
	pathProbeBanMu.Lock()
	if pathProbeBanRoot == "" && dataRoot != "" {
		pathProbeBanRoot = dataRoot
		loadPathProbeBansLocked(dataRoot)
	}
	changed := false
	for ip, e := range pathProbeBans {
		if !e.until.After(now) {
			delete(pathProbeBans, ip)
			changed = true
		}
	}
	for ip, st := range pathProbeStates {
		if st == nil || len(st.events) == 0 {
			delete(pathProbeStates, ip)
			continue
		}
		last := st.events[len(st.events)-1].at
		if last.Before(cutoff) {
			delete(pathProbeStates, ip)
		}
	}
	if changed {
		persistPathProbeBansLocked(dataRoot)
	}
	pathProbeBanMu.Unlock()
}

func listAutoBansJSON() []map[string]any {
	now := time.Now()
	pathProbeBanMu.Lock()
	defer pathProbeBanMu.Unlock()
	out := make([]map[string]any, 0, len(pathProbeBans))
	for ip, e := range pathProbeBans {
		if !e.until.After(now) {
			continue
		}
		sec := int(e.until.Sub(now).Seconds())
		if sec < 1 {
			sec = 1
		}
		item := map[string]any{
			"ip":           ip,
			"untilUnix":    e.until.Unix(),
			"remainingSec": sec,
			"reason":       e.reason,
		}
		if e.paths > 0 {
			item["paths"] = e.paths
		}
		if !e.bannedAt.IsZero() {
			item["bannedAtUnix"] = e.bannedAt.Unix()
		}
		out = append(out, item)
	}
	sortAutoBanItems(out)
	return out
}

func sortAutoBanItems(items []map[string]any) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0; j-- {
			a, _ := items[j]["ip"].(string)
			b, _ := items[j-1]["ip"].(string)
			if a >= b {
				break
			}
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func removeAutoBan(dataRoot, rawIP string) (string, error) {
	ip, ok := normalizeBlacklistIP(rawIP)
	if !ok {
		return "", errInvalidBlacklistIP
	}
	pathProbeBanMu.Lock()
	defer pathProbeBanMu.Unlock()
	if pathProbeBanRoot == "" && dataRoot != "" {
		pathProbeBanRoot = dataRoot
		loadPathProbeBansLocked(dataRoot)
	}
	delete(pathProbeBans, ip)
	delete(pathProbeStates, ip)
	persistPathProbeBansLocked(dataRoot)
	return ip, nil
}

func checkPathProbeGate(w http.ResponseWriter, r *http.Request, dataRoot string) bool {
	ip := visitorClientIP(r)
	if ip == "" {
		return true
	}
	if isIPAutoBanned(ip) {
		writeIPForbidden(w)
		return false
	}
	path := visitorNormalizePath(r.URL.Path)
	if recordPathProbe(ip, path, dataRoot) {
		writeIPForbidden(w)
		return false
	}
	return true
}
