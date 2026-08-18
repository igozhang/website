package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseUserAgentClient(t *testing.T) {
	cases := map[string]string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36":          "Chrome",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36 Edg/120.0": "Edge",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15":                "Safari",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0":                    "Firefox",
		"": "",
		"curl/8.0": "",
	}
	for ua, want := range cases {
		if got := parseUserAgentClient(ua); got != want {
			t.Fatalf("ua=%q got %q want %q", ua, got, want)
		}
	}
}

func TestVisitorPageAndFeatureLabels(t *testing.T) {
	if visitorPageLabel("/") != "首页" {
		t.Fatal("home")
	}
	if visitorPageLabel("/fs") != "文件中转" {
		t.Fatal("fs")
	}
	if visitorPageLabel("/hello-world") != "/hello-world" {
		t.Fatal("article slug")
	}
	if visitorPageLabel("/api/browse") != "" {
		t.Fatal("api")
	}
	if visitorFeatureLabel("/stream/mp3/song.mp3") != "mp3" {
		t.Fatal("mp3 stream")
	}
	if visitorFeatureLabel("/api/stash") != "暂存" {
		t.Fatal("stash")
	}
	if visitorFeatureLabel("/api/file") != "删除" {
		t.Fatal("file delete")
	}
}

func TestVisitorFeatureRequest_method(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/stash", nil)
	if visitorFeatureRequest(r) != "" {
		t.Fatal("get stash should not count")
	}
	r.Method = http.MethodPost
	if visitorFeatureRequest(r) != "暂存" {
		t.Fatal("post stash")
	}
	r = httptest.NewRequest(http.MethodDelete, "/api/file?path=fs/a.zip", nil)
	if visitorFeatureRequest(r) != "删除" {
		t.Fatal("delete file")
	}
	r.Method = http.MethodGet
	if visitorFeatureRequest(r) != "" {
		t.Fatal("get file should not count")
	}
}

func TestShouldSkipVisitorRecord(t *testing.T) {
	skip := []string{"/health", "/favicon.ico", "/api/ops/board", "/api/auth/login", "/api/footer", "/api/ops-config", "/api/browse"}
	for _, p := range skip {
		if !shouldSkipVisitorRecord(p) {
			t.Fatalf("should skip %s", p)
		}
	}
	record := []string{"/igokey", "/ops", "/igoadmin"}
	for _, p := range record {
		if shouldSkipVisitorRecord(p) {
			t.Fatalf("should record %s", p)
		}
	}
}

func TestIsIgoKeyDirectPageVisit(t *testing.T) {
	if !isIgoKeyDirectPageVisit(httptest.NewRequest(http.MethodGet, "/igokey", nil)) {
		t.Fatal("html page visit")
	}
	if isIgoKeyDirectPageVisit(httptest.NewRequest(http.MethodGet, "/igokey?format=plain", nil)) {
		t.Fatal("plain prefetch should not count")
	}
	if isIgoKeyDirectPageVisit(httptest.NewRequest(http.MethodGet, "/igokey?format=json", nil)) {
		t.Fatal("json api should not count")
	}
	if isIgoKeyDirectPageVisit(httptest.NewRequest(http.MethodHead, "/igokey", nil)) {
		t.Fatal("head should not count")
	}
}

func TestRecordVisitor_igokeyPlainSkipped(t *testing.T) {
	dir := t.TempDir()
	initAccessLog(dir)
	visitorBoardMu.Lock()
	visitorBoard = make(map[string]*visitorEntry)
	visitorBoardMu.Unlock()
	accessLogDedupMu.Lock()
	accessLogDedup = make(map[string]time.Time)
	accessLogDedupMu.Unlock()

	r := httptest.NewRequest(http.MethodGet, "/igokey?format=plain", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36")
	r.RemoteAddr = "203.0.113.10:12345"
	recordVisitor(dir, r)

	raw, err := os.ReadFile(accessLogPath(dir))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(raw) != 0 {
		t.Fatalf("plain igokey should not log: %q", raw)
	}

	r2 := httptest.NewRequest(http.MethodGet, "/igokey", nil)
	r2.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36")
	r2.RemoteAddr = "203.0.113.10:12345"
	recordVisitor(dir, r2)
	raw, err = os.ReadFile(accessLogPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"action":"session"`) || !strings.Contains(string(raw), `"target":"igokey"`) {
		t.Fatalf("direct igokey should log session: %q", raw)
	}
}

func TestRecordVisitor_opsAndAdmin(t *testing.T) {
	dir := t.TempDir()
	initAccessLog(dir)
	visitorBoardMu.Lock()
	visitorBoard = make(map[string]*visitorEntry)
	visitorBoardMu.Unlock()
	accessLogDedupMu.Lock()
	accessLogDedup = make(map[string]time.Time)
	accessLogDedupMu.Unlock()

	for _, tc := range []struct {
		path   string
		target string
	}{
		{"/ops", "运维"},
		{"/igoadmin", "文章后台"},
	} {
		r := httptest.NewRequest(http.MethodGet, tc.path, nil)
		r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36")
		r.RemoteAddr = "203.0.113.11:12345"
		recordVisitor(dir, r)
		raw, err := os.ReadFile(accessLogPath(dir))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), `"target":"`+tc.target+`"`) {
			t.Fatalf("%s should log target %s: %q", tc.path, tc.target, raw)
		}
	}
}

func TestRecordVisitor_skipsNoise(t *testing.T) {
	dir := t.TempDir()
	visitorBoardMu.Lock()
	visitorBoard = make(map[string]*visitorEntry)
	visitorBoardMu.Unlock()

	paths := []string{"/api/footer", "/api/ops-config", "/api/browse"}
	for _, p := range paths {
		r := httptest.NewRequest(http.MethodGet, p, nil)
		r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36")
		r.RemoteAddr = "203.0.113.10:12345"
		recordVisitor(dir, r)
	}
	online, _ := visitorBoardSnapshot()
	if len(online) != 0 {
		t.Fatalf("noise recorded %d entries", len(online))
	}
}

func TestRecordVisitorAndSnapshot(t *testing.T) {
	dir := t.TempDir()
	initAccessLog(dir)
	visitorBoardMu.Lock()
	visitorBoard = make(map[string]*visitorEntry)
	visitorBoardMu.Unlock()
	accessLogDedupMu.Lock()
	accessLogDedup = make(map[string]time.Time)
	accessLogDedupMu.Unlock()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36")
	r.RemoteAddr = "203.0.113.10:12345"
	recordVisitor(dir, r)

	r2 := httptest.NewRequest(http.MethodGet, "/stream/mp3/a.mp3", nil)
	r2.RemoteAddr = "203.0.113.10:12345"
	recordVisitor(dir, r2)

	r3 := httptest.NewRequest(http.MethodGet, "/gomoku", nil)
	r3.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36")
	r3.RemoteAddr = "203.0.113.10:12345"
	recordVisitor(dir, r3)

	online, recent := visitorBoardSnapshot()
	if len(online) != 1 {
		t.Fatalf("online=%d want 1", len(online))
	}
	if online[0]["ip"] != "203.0.113.10" {
		t.Fatalf("ip=%v", online[0]["ip"])
	}
	if online[0]["client"] != "Chrome" {
		t.Fatalf("client=%v", online[0]["client"])
	}
	if _, ok := online[0]["loginAt"]; !ok {
		t.Fatalf("loginAt missing: %+v", online[0])
	}
	trail, _ := online[0]["trail"].([]map[string]any)
	if len(trail) != 3 {
		t.Fatalf("trail=%v", online[0]["trail"])
	}
	for i, step := range trail {
		if step["ts"] == nil || step["ts"] == "" {
			t.Fatalf("trail[%d] missing ts: %+v", i, step)
		}
	}
	trailText, _ := online[0]["trailText"].(string)
	if !strings.Contains(trailText, "首页") || !strings.Contains(trailText, "mp3") || !strings.Contains(trailText, "五子棋") {
		t.Fatalf("trailText=%q", trailText)
	}
	if len(recent) != 0 {
		t.Fatalf("recent=%d", len(recent))
	}

	visitorBoardMu.Lock()
	visitorBoard["203.0.113.10"].lastSeen = time.Now().Add(-10 * time.Minute)
	visitorBoardMu.Unlock()
	online, recent = visitorBoardSnapshot()
	if len(online) != 0 {
		t.Fatalf("online after idle=%d", len(online))
	}
	if len(recent) != 1 {
		t.Fatalf("recent=%d want 1", len(recent))
	}
	recentTrail, _ := recent[0]["trail"].([]map[string]any)
	if len(recentTrail) != 3 {
		t.Fatalf("recent trail=%v", recent[0]["trail"])
	}
}

func TestRecordVisitor_sessionAndPath(t *testing.T) {
	dir := t.TempDir()
	initAccessLog(dir)
	visitorBoardMu.Lock()
	visitorBoard = make(map[string]*visitorEntry)
	visitorBoardMu.Unlock()
	accessLogDedupMu.Lock()
	accessLogDedup = make(map[string]time.Time)
	accessLogDedupMu.Unlock()

	ip := "203.0.113.50"
	ua := "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36"

	r1 := httptest.NewRequest(http.MethodGet, "/", nil)
	r1.Header.Set("User-Agent", ua)
	r1.RemoteAddr = ip + ":12345"
	recordVisitor(dir, r1)

	r2 := httptest.NewRequest(http.MethodGet, "/gomoku", nil)
	r2.Header.Set("User-Agent", ua)
	r2.RemoteAddr = ip + ":12345"
	recordVisitor(dir, r2)

	raw, err := os.ReadFile(accessLogPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want session+path lines, got %d: %q", len(lines), raw)
	}
	if !strings.Contains(lines[0], `"action":"session"`) || !strings.Contains(lines[0], `"target":"首页"`) {
		t.Fatalf("first line should be session: %q", lines[0])
	}
	if !strings.Contains(lines[1], `"action":"visit"`) || !strings.Contains(lines[1], `"target":"五子棋"`) {
		t.Fatalf("second line should be path visit: %q", lines[1])
	}

	st := countAccessLogStats(dir, time.Now())
	if st.Day != 1 {
		t.Fatalf("stats day=%d want 1 session", st.Day)
	}
}

func TestRecordVisitor_newSessionAfterIdle(t *testing.T) {
	dir := t.TempDir()
	initAccessLog(dir)
	visitorBoardMu.Lock()
	visitorBoard = make(map[string]*visitorEntry)
	visitorBoardMu.Unlock()
	accessLogDedupMu.Lock()
	accessLogDedup = make(map[string]time.Time)
	accessLogDedupMu.Unlock()

	ip := "203.0.113.51"
	ua := "Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36"

	r1 := httptest.NewRequest(http.MethodGet, "/", nil)
	r1.Header.Set("User-Agent", ua)
	r1.RemoteAddr = ip + ":12345"
	recordVisitor(dir, r1)

	visitorBoardMu.Lock()
	visitorBoard[ip].lastSeen = time.Now().Add(-31 * time.Minute)
	visitorBoardMu.Unlock()

	r2 := httptest.NewRequest(http.MethodGet, "/fs", nil)
	r2.Header.Set("User-Agent", ua)
	r2.RemoteAddr = ip + ":12345"
	recordVisitor(dir, r2)

	raw, err := os.ReadFile(accessLogPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(raw), `"action":"session"`) != 2 {
		t.Fatalf("want 2 sessions: %q", raw)
	}

	online, _ := visitorBoardSnapshot()
	trail, _ := online[0]["trail"].([]map[string]any)
	if len(trail) != 1 || trail[0]["target"] != "文件中转" {
		t.Fatalf("trail should reset for new session: %+v", online[0]["trail"])
	}

	st := countAccessLogStats(dir, time.Now())
	if st.Day != 2 {
		t.Fatalf("stats day=%d want 2 sessions", st.Day)
	}
}

func TestVisitorNormalizePath(t *testing.T) {
	setVisitorHTTPBasePath("/fs")
	defer setVisitorHTTPBasePath("")
	if got := visitorNormalizePath("/fs/gomoku"); got != "/gomoku" {
		t.Fatalf("gomoku=%q", got)
	}
	if got := visitorNormalizePath("/fs"); got != "/" {
		t.Fatalf("fs root=%q want /", got)
	}
	if got := visitorPageLabel(visitorNormalizePath("/fs/gomoku")); got != "五子棋" {
		t.Fatalf("label=%q", got)
	}
	setVisitorHTTPBasePath("")
	if got := visitorNormalizePath("/gomoku"); got != "/gomoku" {
		t.Fatalf("plain=%q", got)
	}
}

func TestEvictVisitorBoard_skipsActiveSession(t *testing.T) {
	visitorBoardMu.Lock()
	visitorBoard = make(map[string]*visitorEntry)
	for i := 0; i < visitorMaxEntries; i++ {
		ip := fmt.Sprintf("10.0.0.%d", i)
		visitorBoard[ip] = &visitorEntry{
			ip:       ip,
			lastSeen: time.Now().Add(-2 * time.Hour),
		}
	}
	activeIP := "203.0.113.99"
	visitorBoard[activeIP] = &visitorEntry{
		ip:       activeIP,
		lastSeen: time.Now().Add(-5 * time.Minute),
	}
	now := time.Now()
	evictVisitorBoardIfNeededLocked(now, visitorSessionIdleTimeout())
	if _, ok := visitorBoard[activeIP]; !ok {
		t.Fatal("active session IP should not be evicted")
	}
	visitorBoardMu.Unlock()
}

func TestRequestGateBlocksBlacklist(t *testing.T) {
	ipBlacklistMu.Lock()
	prev := ipBlacklistSet
	ipBlacklistSet = map[string]struct{}{"203.0.113.99": {}}
	ipBlacklistMu.Unlock()
	t.Cleanup(func() {
		ipBlacklistMu.Lock()
		ipBlacklistSet = prev
		ipBlacklistMu.Unlock()
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := requestGate("", next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.99:12345"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "203.0.113.1:12345"
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("allowed status=%d", rec2.Code)
	}
}

func TestRequestGateOpsWhitelist(t *testing.T) {
	ipWhitelistMu.Lock()
	prev := ipWhitelistSet
	ipWhitelistSet = map[string]struct{}{"203.0.113.10": {}}
	ipWhitelistMu.Unlock()
	t.Cleanup(func() {
		ipWhitelistMu.Lock()
		ipWhitelistSet = prev
		ipWhitelistMu.Unlock()
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := requestGate("", next)

	req := httptest.NewRequest(http.MethodGet, "/ops", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("listed ops status=%d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/igoadmin", nil)
	req2.RemoteAddr = "203.0.113.99:12345"
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("unlisted igoadmin status=%d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "203.0.113.99:12345"
	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("public path should not be restricted status=%d", rec3.Code)
	}
}
