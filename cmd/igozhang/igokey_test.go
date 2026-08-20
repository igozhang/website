package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestIgokeyHTTPHandler_HTMLPageClosed(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	igoKeyMu.Unlock()

	dir := t.TempDir()
	mux := http.NewServeMux()
	registerIgoKeyHandler(mux, dir)
	for _, path := range []string{"/igokey", "/igokey/", "/igokey?renew=1"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status=%d want 404", path, rec.Code)
		}
		body := rec.Body.String()
		if strings.Contains(body, "验证密钥已复制到剪贴板") || strings.Contains(body, `id="igoKey"`) {
			t.Fatalf("%s should not serve igokey HTML: %s", path, body)
		}
	}
	if _, err := os.Stat(igoKeyPlainPath(dir)); !os.IsNotExist(err) {
		t.Fatal("HTML visit should not write secret07/igokey")
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey?format=plain", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("plain status=%d", rec.Code)
	}
	key := strings.TrimSpace(rec.Body.String())
	if len(key) != igoKeyLen {
		t.Fatalf("plain key=%q", key)
	}
	if got := readPlainIgoKeyFile(t, dir); got != key {
		t.Fatalf("secret07/igokey=%q want %q", got, key)
	}
}

func TestIgokeyHTTPHandler_RenewJSON(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	igoKeyMu.Unlock()

	dir := t.TempDir()
	mux := http.NewServeMux()
	registerIgoKeyHandler(mux, dir)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey?format=json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("initial status=%d", rec.Code)
	}
	var first struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if first.Key == "" {
		t.Fatal("missing key in json")
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey?renew=1&format=json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("renew status=%d want 200", rec.Code)
	}
	var second struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	if second.Key == "" || second.Key == first.Key {
		t.Fatalf("renew key=%q initial=%q", second.Key, first.Key)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/igokey?format=json", nil))
	var third struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &third); err != nil {
		t.Fatal(err)
	}
	if third.Key != second.Key {
		t.Fatalf("second refresh changed key: %q vs %q", third.Key, second.Key)
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
