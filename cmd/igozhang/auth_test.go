package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAuthSession_KeyRevokedAfterNewIgoKey(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "sess-revoke-test"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	mux := http.NewServeMux()
	registerAuthHandlers(mux, dir)

	body, _ := json.Marshal(map[string]string{"token": tok})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/session", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session check: status=%d %s", rec.Code, rec.Body.String())
	}

	_, _ = obtainIgoKey(dir, true)

	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/session", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 after key rotate, got %d", rec2.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["reason"] != authReasonKeyRevoked {
		t.Fatalf("reason=%v want %s", resp["reason"], authReasonKeyRevoked)
	}
}

func TestAuthFailRL_accumulates(t *testing.T) {
	authFailRLMu.Lock()
	authFailRL = make(map[string]*authFailState)
	authRLRoot = ""
	authFailRLMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initAuthFailRL(dir)

	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	r.RemoteAddr = "203.0.113.9:12345"
	w := httptest.NewRecorder()

	for i := 0; i < 4; i++ {
		if !checkAuthFailRL(w, r, dir) {
			t.Fatalf("attempt %d should not block", i+1)
		}
		locked, _ := recordAuthFail(r, dir)
		if locked {
			t.Fatalf("attempt %d should not lock yet", i+1)
		}
		authFailRLMu.Lock()
		n := len(authFailRL["203.0.113.9"].failTimes)
		authFailRLMu.Unlock()
		if n != i+1 {
			t.Fatalf("after %d fails, want %d recorded, got %d", i+1, i+1, n)
		}
	}

	if !checkAuthFailRL(w, r, dir) {
		t.Fatal("5th check should not block before recording")
	}
	locked, retrySec := recordAuthFail(r, dir)
	if !locked || retrySec < 1 {
		t.Fatalf("5th fail should lock, locked=%v retrySec=%d", locked, retrySec)
	}

	w2 := httptest.NewRecorder()
	if checkAuthFailRL(w2, r, dir) {
		t.Fatal("6th check should block while locked")
	}
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("blocked status=%d want 429", w2.Code)
	}

	raw, err := os.ReadFile(authRLPath(dir))
	if err != nil {
		t.Fatalf("persist file: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("rate limit state should persist to disk")
	}
}

func TestAuthFailRateLimit(t *testing.T) {
	authFailRLMu.Lock()
	authFailRL = make(map[string]*authFailState)
	authRLRoot = ""
	authFailRLMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	initAuthFailRL(dir)
	key, _ := obtainIgoKey(dir, false)

	mux := http.NewServeMux()
	registerAuthHandlers(mux, dir)

	doAuth := func(wantKey string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]string{"key": wantKey})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "203.0.113.9:12345"
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	for i := 0; i < 4; i++ {
		rec := doAuth("badkey00")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("fail %d: status=%d want 401", i+1, rec.Code)
		}
	}
	rec := doAuth("badkey00")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("5th fail: status=%d want 429", rec.Code)
	}
	var blocked map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &blocked); err != nil {
		t.Fatal(err)
	}
	if blocked["retryAfterSec"] == nil {
		t.Fatalf("429 response missing retryAfterSec: %v", blocked)
	}

	recOK := doAuth(key)
	if recOK.Code != http.StatusTooManyRequests {
		t.Fatalf("auth during lockout: status=%d want 429", recOK.Code)
	}

	authFailRLMu.Lock()
	if st := authFailRL["203.0.113.9"]; st != nil {
		st.lockedUntil = time.Now().Add(-time.Second)
	}
	persistAuthFailRLLocked(dir)
	authFailRLMu.Unlock()

	rec2 := doAuth(key)
	if rec2.Code != http.StatusOK {
		t.Fatalf("valid key after lock expired: status=%d %s", rec2.Code, rec2.Body.String())
	}
}

func TestAuthLogin_noTargetsField(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	key, _ := obtainIgoKey(dir, false)

	mux := http.NewServeMux()
	registerAuthHandlers(mux, dir)

	body, _ := json.Marshal(map[string]string{"key": key})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("auth without ssh.pass: status=%d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true || resp["token"] == "" {
		t.Fatalf("unexpected auth response: %v", resp)
	}
	if _, hasTargets := resp["targets"]; hasTargets {
		t.Fatalf("login must not include targets field: %v", resp)
	}
}

func TestOpsTargets_withToken(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	key, exp := obtainIgoKey(dir, false)
	tok := "ops-targets-tok"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: key, igoExpiresAt: exp}
	authSessionMu.Unlock()

	mux := http.NewServeMux()
	registerOpsHandlers(mux, dir)

	body, _ := json.Marshal(map[string]string{"token": tok})
	req := httptest.NewRequest(http.MethodPost, "/api/ops/targets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("targets: status=%d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true {
		t.Fatalf("resp=%v", resp)
	}
	targets, ok := resp["targets"].([]any)
	if !ok {
		t.Fatalf("targets missing or wrong type: %v", resp["targets"])
	}
	if len(targets) != 0 {
		t.Fatalf("targets=%v want empty without ssh.pass", targets)
	}
}

func TestOpsPageEnabled_independentOfIgokey(t *testing.T) {
	opsMu.Lock()
	opsCurrent.opsEnabled = true
	opsCurrent.igokeyEnabled = false
	opsMu.Unlock()
	if !opsPageEnabled() {
		t.Fatal("opsPageEnabled should follow ops only")
	}
	opsMu.Lock()
	opsCurrent.opsEnabled = false
	opsCurrent.igokeyEnabled = true
	opsMu.Unlock()
	if opsPageEnabled() {
		t.Fatal("opsPageEnabled should be false when ops off")
	}
}

func TestAuthFeatureEnabled(t *testing.T) {
	opsMu.Lock()
	opsCurrent.authEnabled = true
	opsCurrent.igokeyEnabled = true
	opsMu.Unlock()
	if !authFeatureEnabled() {
		t.Fatal("want auth on")
	}
	opsMu.Lock()
	opsCurrent.authEnabled = false
	opsMu.Unlock()
	if authFeatureEnabled() {
		t.Fatal("auth off should disable")
	}
	opsMu.Lock()
	opsCurrent.authEnabled = true
	opsCurrent.igokeyEnabled = false
	opsMu.Unlock()
	if authFeatureEnabled() {
		t.Fatal("igokey off should disable auth feature")
	}
}

func TestParseOpsConfig_AuthEnabled(t *testing.T) {
	s, err := parseOpsConfig("[auth]\nenabled = off\n")
	if err != nil {
		t.Fatal(err)
	}
	if s.authEnabled {
		t.Fatal("expected auth off")
	}
	s2, err := parseOpsConfig(string(opsConfigExample))
	if err != nil || !s2.authEnabled {
		t.Fatalf("example auth=%v err=%v", s2.authEnabled, err)
	}
}
