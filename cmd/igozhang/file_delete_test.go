package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAPIFileDelete(t *testing.T) {
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

	fsDir := filepath.Join(dir, "fs")
	if err := os.MkdirAll(fsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(fsDir, "hello.zip")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(dir, "assets", "mp3")
	if err := os.MkdirAll(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blocked, "song.mp3"), []byte("m"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(fsDir, "foo"), 0o755); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/file", fileDeleteHandler(dir))

	req := httptest.NewRequest(http.MethodDelete, "/api/file?path=fs/hello.zip", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rec.Code)
	}

	key, exp := obtainIgoKey(dir, false)
	tok := "file-del-tok"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: key, igoExpiresAt: exp}
	authSessionMu.Unlock()

	authReq := func(url string) *http.Request {
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		r.Header.Set("X-Igo-Token", tok)
		return r
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, authReq("/api/file?path=fs/foo"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("dir delete: want 400 got %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, authReq("/api/file?path=assets/mp3/song.mp3"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("blocked path: want 404 got %d %s", rec.Code, rec.Body.String())
	}

	for _, bad := range []string{
		"/api/file?path=",
		"/api/file?path=.",
		"/api/file?path=../hello.zip",
		"/api/file?path=fs/../../hello.zip",
	} {
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, authReq(bad))
		if rec.Code == http.StatusOK {
			t.Fatalf("bad path %s should not succeed: %s", bad, rec.Body.String())
		}
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, authReq("/api/file?path=fs/hello.zip"))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: want 200 got %d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true || resp["deleted"] != "fs/hello.zip" {
		t.Fatalf("resp=%v", resp)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("file should be gone")
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, authReq("/api/file?path=fs/hello.zip"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing file: want 404 got %d", rec.Code)
	}
}
