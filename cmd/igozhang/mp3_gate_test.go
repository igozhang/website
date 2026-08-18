package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMp3Gate_SignVerify(t *testing.T) {
	ensureMp3GateKey()
	exp := time.Now().Add(time.Hour).Unix()
	val := signMp3GateValue(exp)
	if !verifyMp3GateValue(val) {
		t.Fatal("fresh gate value should verify")
	}
	if verifyMp3GateValue("1") || verifyMp3GateValue(val+"x") {
		t.Fatal("forged values must fail")
	}
	expired := signMp3GateValue(time.Now().Add(-time.Minute).Unix())
	if verifyMp3GateValue(expired) {
		t.Fatal("expired gate value must fail")
	}
}

func TestMp3GateHandler_SetsCookie(t *testing.T) {
	mp3GateEnabled = true
	mp3RefererCheck = true
	req := httptest.NewRequest(http.MethodGet, "/api/mp3/gate", nil)
	req.Host = "igozhang.cn"
	rec := httptest.NewRecorder()
	mp3GateHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("gate: %d %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != mp3GateCookieName || !verifyMp3GateValue(cookies[0].Value) {
		t.Fatalf("cookies=%v", cookies)
	}
}

func TestStreamMP3_RequiresGate(t *testing.T) {
	mp3GateEnabled = true
	mp3RefererCheck = false
	mp3MaxConcurrentPerIP = 4
	mp3MaxTracksPerMinute = 40

	dir := t.TempDir()
	mp3Dir := filepath.Join(dir, "assets", "mp3")
	if err := os.MkdirAll(mp3Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	song := filepath.Join(mp3Dir, "a.mp3")
	if err := os.WriteFile(song, []byte("ID3fake"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := streamMP3Handler(dir)
	req := httptest.NewRequest(http.MethodGet, "/stream/mp3/a.mp3", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("without cookie want 403 got %d", rec.Code)
	}

	gateRec := httptest.NewRecorder()
	mp3GateHandler().ServeHTTP(gateRec, httptest.NewRequest(http.MethodGet, "/api/mp3/gate", nil))
	var c *http.Cookie
	for _, ck := range gateRec.Result().Cookies() {
		if ck.Name == mp3GateCookieName {
			c = ck
			break
		}
	}
	if c == nil {
		t.Fatal("missing gate cookie")
	}
	req2 := httptest.NewRequest(http.MethodGet, "/stream/mp3/a.mp3", nil)
	req2.AddCookie(c)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("with cookie want 200 got %d %s", rec2.Code, rec2.Body.String())
	}
	if ct := rec2.Header().Get("Content-Type"); !strings.HasPrefix(ct, "audio/mpeg") {
		t.Fatalf("content-type=%q", ct)
	}
}

func TestMp3RefererCheck(t *testing.T) {
	mp3GateEnabled = true
	mp3RefererCheck = true
	req := httptest.NewRequest(http.MethodGet, "/stream/mp3/a.mp3", nil)
	req.Host = "igozhang.cn"
	req.Header.Set("Referer", "https://evil.example/page")
	if mp3RefererOK(req) {
		t.Fatal("foreign referer should fail")
	}
	req.Header.Set("Referer", "https://igozhang.cn/")
	if !mp3RefererOK(req) {
		t.Fatal("same-host referer should pass")
	}
	req.Header.Del("Referer")
	if !mp3RefererOK(req) {
		t.Fatal("empty referer should pass")
	}
}

func TestParseOpsConfig_Mp3Gate(t *testing.T) {
	s, err := parseOpsConfig(`
[mp3]
gate_enabled = off
gate_ttl = 2h
referer_check = off
max_concurrent_per_ip = 6
max_tracks_per_minute = 12
`)
	if err != nil {
		t.Fatal(err)
	}
	if s.mp3GateEnabled || s.mp3RefererCheck {
		t.Fatalf("enabled/referer: %v %v", s.mp3GateEnabled, s.mp3RefererCheck)
	}
	if s.mp3GateTTL != 2*time.Hour || s.mp3MaxConcurrentPerIP != 6 || s.mp3MaxTracksPerMinute != 12 {
		t.Fatalf("ttl=%v conc=%d tracks=%d", s.mp3GateTTL, s.mp3MaxConcurrentPerIP, s.mp3MaxTracksPerMinute)
	}
}
