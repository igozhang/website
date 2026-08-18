package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestFsPageHTML_HasAuthGate(t *testing.T) {
	html := string(renderFsHTML("", nil))
	for _, want := range []string{
		`IGO密钥认证`,
		`id="igoAuthSec"`,
		`验证密钥（key）`,
		`id="igoKey"`,
		`id="btnIgoAuth"`,
		`id="igoMainSec"`,
		`class="hidden"`,
		`__igoAuthInit`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("fs html missing %q", want)
		}
	}
	if !strings.Contains(html, `data-fs-module="transfer"`) {
		t.Fatal("fs modules should still be present inside main sec")
	}
}

func TestFsAPIs_RequireToken(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	opsMu.Lock()
	opsCurrent.opsEnabled = true
	opsCurrent.authEnabled = true
	opsCurrent.igokeyEnabled = true
	opsMu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/browse", func(w http.ResponseWriter, r *http.Request) {
		q := filepath.ToSlash(r.URL.Query().Get("path"))
		qNorm := strings.Trim(filepath.ToSlash(filepath.Clean(strings.TrimSpace(q))), "/")
		if qNorm == "." {
			qNorm = ""
		}
		if qNorm != "assets/mp3" && !requireAuthToken(w, r, dir) {
			return
		}
		jsonWrite(w, []any{})
	})
	mux.HandleFunc("/api/stash", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuthToken(w, r, dir) {
			return
		}
		jsonWrite(w, map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/fs/usage", fsUsageHandler(dir, dir))

	// unauthorized
	for _, path := range []string{"/api/browse?path=", "/api/stash", "/api/fs/usage"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s without token: want 401 got %d", path, rec.Code)
		}
	}

	// mp3 browse skips igokey token (gate cookie enforced by real handler)
	for _, mp3Path := range []string{"/api/browse?path=assets/mp3", "/api/browse?path=assets/mp3/", "/api/browse?path=./assets/mp3"} {
		req := httptest.NewRequest(http.MethodGet, mp3Path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("mp3 browse %s should be public, got %d", mp3Path, rec.Code)
		}
	}

	key, exp := obtainIgoKey(dir, false)
	tok := "fs-auth-test-tok"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: key, igoExpiresAt: exp}
	authSessionMu.Unlock()

	req2 := httptest.NewRequest(http.MethodGet, "/api/stash", nil)
	req2.Header.Set("X-Igo-Token", tok)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("stash with token: want 200 got %d %s", rec2.Code, rec2.Body.String())
	}
	var stashResp map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &stashResp); err != nil {
		t.Fatal(err)
	}
	if stashResp["ok"] != true {
		t.Fatalf("stash resp=%v", stashResp)
	}
}

func TestIsOpsIgoAdminPath_IncludesFs(t *testing.T) {
	if !isOpsIgoAdminPath("/fs") || !isOpsIgoAdminPath("/fs/") {
		t.Fatal("/fs should be IP-whitelist protected like /ops")
	}
	if isOpsIgoAdminPath("/api/browse") {
		t.Fatal("/api/browse stays off whitelist path list (token-gated instead)")
	}
	if isOpsIgoAdminPath("/files/fs/igozhang.cn.yaml") {
		t.Fatal("/files/* direct download must stay public (not whitelist-gated)")
	}
}
