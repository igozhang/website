package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode"
)

func TestRandomIgoKeyString(t *testing.T) {
	k, err := randomIgoKeyString(igoKeyLen)
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != igoKeyLen {
		t.Fatalf("len=%d key=%q", len(k), k)
	}
	for _, c := range k {
		if !unicode.IsLetter(c) {
			t.Fatalf("non-letter %q in %q", c, k)
		}
	}
}

func TestValidateIgoKey(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	key, exp := obtainIgoKey(dir, false)
	if !validateIgoKey(key) {
		t.Fatal("expected valid key")
	}
	igoKeyMu.Lock()
	igoKeyStore[key] = igoKeyEntry{expiresAt: time.Now().Add(-time.Minute)}
	igoKeyMu.Unlock()
	if validateIgoKey(key) {
		t.Fatal("expected expired key invalid")
	}
	if validateIgoKey("short") || validateIgoKey("12345678") {
		t.Fatal("expected malformed invalid")
	}
	_ = exp
}

func TestIgokeyHTTPHandler_Disabled(t *testing.T) {
	opsMu.Lock()
	opsCurrent.igokeyEnabled = false
	opsMu.Unlock()
	defer func() {
		opsMu.Lock()
		opsCurrent.igokeyEnabled = true
		opsMu.Unlock()
	}()

	mux := http.NewServeMux()
	registerIgoKeyHandler(mux, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/igokey", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
}

func TestIgokeyHTTPHandler_JSON(t *testing.T) {
	mux := http.NewServeMux()
	registerIgoKeyHandler(mux, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/igokey?format=json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Key          string `json:"key"`
		ExpiresAt    int64  `json:"expiresAt"`
		ExpiresInSec int    `json:"expiresInSec"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Key) != igoKeyLen {
		t.Fatalf("key=%q", out.Key)
	}
	if out.ExpiresInSec < 3500 || out.ExpiresInSec > 3700 {
		t.Fatalf("expiresInSec=%d", out.ExpiresInSec)
	}
	if !validateIgoKey(out.Key) {
		t.Fatal("issued key not valid in store")
	}
}

func TestIgokeyHTTPHandler_HTMLClipboard(t *testing.T) {
	mux := http.NewServeMux()
	registerIgoKeyHandler(mux, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/igokey", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if !strings.Contains(body, "navigator.clipboard") {
		t.Fatal("missing clipboard script")
	}
	if !strings.Contains(body, "验证密钥已复制到剪贴板") {
		t.Fatal("missing toast copy hint")
	}
	if !strings.Contains(body, `id="btnIgoCopy"`) {
		t.Fatal("missing copy button")
	}
	if !strings.Contains(body, `id="btnIgoRenew"`) {
		t.Fatal("missing renew button")
	}
	if strings.Contains(body, `href="?renew=1"`) {
		t.Fatal("renew must use absolute /igokey path, not ?renew=1 alone")
	}
	if !strings.Contains(body, `href="/igokey?renew=1"`) {
		t.Fatal("missing renew href /igokey?renew=1")
	}
	if !strings.Contains(body, "var remainSec=") {
		t.Fatal("missing remainSec for client-relative expiry display")
	}
	if !strings.Contains(body, `id="igoRemMin"`) {
		t.Fatal("missing remaining minutes display")
	}
	if !strings.Contains(body, `class="site-head"`) {
		t.Fatal("missing site header")
	}
	if !strings.Contains(body, `class="site-foot"`) {
		t.Fatal("missing site footer")
	}
	if !strings.Contains(body, "image/svg+xml") {
		t.Fatal("missing favicon in page head")
	}
}

func TestIgokeyHTTPHandler_RenewRedirectsToCleanURL(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	igoKeyMu.Unlock()

	dir := t.TempDir()
	mux := http.NewServeMux()
	registerIgoKeyHandler(mux, dir)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("initial status=%d", rec.Code)
	}
	k1 := extractIgoKeyFromHTML(rec.Body.String())
	if k1 == "" {
		t.Fatal("missing key in html")
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey?renew=1", nil))
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("renew status=%d want 303", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/igokey" {
		t.Fatalf("Location=%q want /igokey", loc)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status=%d", rec.Code)
	}
	k2 := extractIgoKeyFromHTML(rec.Body.String())
	if k2 == "" || k2 == k1 {
		t.Fatalf("renew key=%q initial=%q", k2, k1)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey", nil))
	k3 := extractIgoKeyFromHTML(rec.Body.String())
	if k3 != k2 {
		t.Fatalf("second refresh changed key: %q vs %q", k3, k2)
	}
}

func TestIgokeyHTTPHandler_RenewJSONNoRedirect(t *testing.T) {
	mux := http.NewServeMux()
	registerIgoKeyHandler(mux, t.TempDir())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey?renew=1&format=json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func extractIgoKeyFromHTML(body string) string {
	const pre = `<p class="key" id="igoKey" aria-label="验证密钥">`
	i := strings.Index(body, pre)
	if i < 0 {
		return ""
	}
	rest := body[i+len(pre):]
	j := strings.Index(rest, "</p>")
	if j < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:j])
}

func TestObtainIgoKey_ReuseAndRenew(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	igoKeyMu.Unlock()

	dir := t.TempDir()
	k1, exp1 := obtainIgoKey(dir, false)
	k2, _ := obtainIgoKey(dir, false)
	if k1 != k2 {
		t.Fatalf("reuse: %q vs %q", k1, k2)
	}
	igoKeyMu.Lock()
	expReuse := time.Now().Add(20 * time.Minute)
	igoKeyStore[k1] = igoKeyEntry{expiresAt: expReuse}
	igoKeyActive.key = k1
	igoKeyActive.expiresAt = expReuse
	persistIgoKeyActiveLocked(dir, k1, expReuse)
	igoKeyMu.Unlock()

	k3, _ := obtainIgoKey(dir, false)
	if k3 != k1 {
		t.Fatalf("reuse within 15m window: %q vs %q", k3, k1)
	}

	igoKeyMu.Lock()
	expShort := time.Now().Add(1 * time.Minute)
	igoKeyStore[k1] = igoKeyEntry{expiresAt: expShort}
	igoKeyActive.expiresAt = expShort
	persistIgoKeyActiveLocked(dir, k1, expShort)
	igoKeyMu.Unlock()

	k4, _ := obtainIgoKey(dir, false)
	if k4 == k1 {
		t.Fatal("expected new key when remaining <= 15m")
	}

	k5, _ := obtainIgoKey(dir, true)
	if k5 == k4 {
		t.Fatal("renew should issue different key")
	}
	if validateIgoKey(k1) || validateIgoKey(k4) {
		t.Fatal("revoked keys must not validate")
	}
	if !validateIgoKey(k5) {
		t.Fatal("current active key must validate")
	}
	_ = exp1
}

func TestObtainIgoKey_RevokesAuthSessionsOnNewKey(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	k1, exp1 := obtainIgoKey(dir, false)
	tok := "testtoken"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()
	if !authTokenValid(dir, tok) {
		t.Fatal("auth session should be valid with active key")
	}

	igoKeyMu.Lock()
	expShort := time.Now().Add(1 * time.Minute)
	igoKeyStore[k1] = igoKeyEntry{expiresAt: expShort}
	igoKeyActive.expiresAt = expShort
	persistIgoKeyActiveLocked(dir, k1, expShort)
	igoKeyMu.Unlock()

	k2, _ := obtainIgoKey(dir, false)
	if k2 == k1 {
		t.Fatal("expected new key when remaining <= 15m")
	}
	if authTokenValid(dir, tok) {
		t.Fatal("auth session with old key must be revoked after new site key")
	}
	if !validateIgoKey(k2) {
		t.Fatal("new active key must validate")
	}
}
