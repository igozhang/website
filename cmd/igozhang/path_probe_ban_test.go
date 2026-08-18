package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func resetPathProbeBanState() {
	pathProbeBanMu.Lock()
	pathProbeBanRoot = ""
	pathProbeStates = make(map[string]*pathProbeState)
	pathProbeBans = make(map[string]pathProbeBanEntry)
	pathProbeBanMu.Unlock()
	opsMu.Lock()
	opsCurrent = defaultOpsSnapshot()
	opsMu.Unlock()
}

func enablePathProbeForTest(window time.Duration, maxPaths int) {
	opsMu.Lock()
	s := opsCurrent
	s.pathProbeEnabled = true
	s.pathProbeWindow = window
	s.pathProbeMaxPaths = maxPaths
	s.pathProbeBanDuration = time.Hour
	s.pathProbeExcludePrefixes = []string{"/stream/mp3/", "/assets/doc-pic/"}
	opsCurrent = s
	opsMu.Unlock()
}

func TestPathProbeBan_triggersAtThreshold(t *testing.T) {
	resetPathProbeBanState()
	enablePathProbeForTest(time.Minute, 5)

	dir := t.TempDir()
	initPathProbeBan(dir)
	ip := "203.0.113.55"

	for i := 0; i < 4; i++ {
		if recordPathProbe(ip, "/scan-"+string(rune('a'+i)), dir) {
			t.Fatalf("path %d should not ban yet", i+1)
		}
	}
	if !recordPathProbe(ip, "/scan-e", dir) {
		t.Fatal("5th distinct path should trigger ban")
	}
	if !isIPAutoBanned(ip) {
		t.Fatal("ip should be auto-banned")
	}
	if !recordPathProbe(ip, "/scan-f", dir) {
		t.Fatal("already banned should stay banned")
	}
}

func TestPathProbeBan_dedupSamePath(t *testing.T) {
	resetPathProbeBanState()
	enablePathProbeForTest(time.Minute, 3)

	dir := t.TempDir()
	initPathProbeBan(dir)
	ip := "203.0.113.56"

	for i := 0; i < 10; i++ {
		recordPathProbe(ip, "/same", dir)
	}
	if isIPAutoBanned(ip) {
		t.Fatal("same path repeated should not ban")
	}
}

func TestPathProbeBan_excludePrefix(t *testing.T) {
	resetPathProbeBanState()
	enablePathProbeForTest(time.Minute, 3)

	dir := t.TempDir()
	initPathProbeBan(dir)
	ip := "203.0.113.57"

	for i := 0; i < 10; i++ {
		recordPathProbe(ip, "/stream/mp3/song-"+string(rune('a'+i))+".mp3", dir)
	}
	if isIPAutoBanned(ip) {
		t.Fatal("excluded mp3 paths should not ban")
	}
}

func TestPathProbeBan_persistAndReload(t *testing.T) {
	resetPathProbeBanState()
	enablePathProbeForTest(time.Minute, 2)

	dir := t.TempDir()
	initPathProbeBan(dir)
	ip := "203.0.113.58"

	if recordPathProbe(ip, "/a", dir) {
		t.Fatal("first path should not ban yet")
	}
	if !recordPathProbe(ip, "/b", dir) {
		t.Fatal("second path should ban")
	}

	raw, err := os.ReadFile(pathProbeBanPath(dir))
	if err != nil {
		t.Fatalf("persist file: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("expected persisted ban")
	}

	resetPathProbeBanState()
	initPathProbeBan(dir)
	if !isIPAutoBanned(ip) {
		t.Fatal("ban should survive reload")
	}
}

func TestPathProbeBan_requestGate(t *testing.T) {
	resetPathProbeBanState()
	enablePathProbeForTest(time.Minute, 3)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := requestGate("", next)

	req := func(path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.RemoteAddr = "203.0.113.59:12345"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}

	if req("/p1").Code != http.StatusOK {
		t.Fatal("first path")
	}
	if req("/p2").Code != http.StatusOK {
		t.Fatal("second path")
	}
	if req("/p3").Code != http.StatusForbidden {
		t.Fatalf("third distinct path should ban, got %d", req("/p3").Code)
	}
	if req("/").Code != http.StatusForbidden {
		t.Fatal("subsequent requests should stay banned")
	}
}

func TestRemoveAutoBan(t *testing.T) {
	resetPathProbeBanState()
	enablePathProbeForTest(time.Minute, 1)

	dir := t.TempDir()
	initPathProbeBan(dir)
	ip := "203.0.113.60"
	recordPathProbe(ip, "/only", dir)
	if !isIPAutoBanned(ip) {
		t.Fatal("should be banned")
	}
	if _, err := removeAutoBan(dir, ip); err != nil {
		t.Fatal(err)
	}
	if isIPAutoBanned(ip) {
		t.Fatal("should be unbanned")
	}
	if _, err := os.Stat(pathProbeBanPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("persist file should be removed: %v", err)
	}
}

func TestMaintainPathProbeBan_prunesExpired(t *testing.T) {
	resetPathProbeBanState()
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "secret07"), 0o755)

	pathProbeBanMu.Lock()
	pathProbeBanRoot = dir
	pathProbeBans["203.0.113.61"] = pathProbeBanEntry{
		until:  time.Now().Add(-time.Minute),
		reason: pathProbeBanFile,
	}
	pathProbeBanMu.Unlock()

	maintainPathProbeBan(dir)
	if isIPAutoBanned("203.0.113.61") {
		t.Fatal("expired ban should be pruned")
	}
}
