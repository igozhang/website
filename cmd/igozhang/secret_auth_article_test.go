package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestSecretAuthArticle_gateAndAPI(t *testing.T) {
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	authSessions = map[string]authSession{}
	igoKeyMu.Unlock()

	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	opsCurrent.authEnabled = true
	opsCurrent.igokeyEnabled = true

	slug := "secret_notes"
	if err := os.WriteFile(linkConfPath(dir), []byte("[secret-auth]\n"+slug+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	secretBody := "# Secret\n\nTOP_SECRET_BODY_CONTENT\n"
	_ = os.WriteFile(filepath.Join(assets, "md", slug+".md"), []byte(secretBody), 0o644)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: slug, Title: "Secret Notes", Categories: "private",
		Author: "igo", PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
	}, []byte(secretBody)); err != nil {
		t.Fatal(err)
	}
	docStore = st
	defer func() { docStore = nil }()

	mux := http.NewServeMux()
	registerAuthHandlers(mux, dir)
	registerDocHandlers(mux, dir)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		seg := strings.Trim(r.URL.Path, "/")
		if tryServeDocArticle(w, r, "", seg, dir) {
			return
		}
		http.NotFound(w, r)
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+slug, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("gate status=%d", rec.Code)
	}
	html := rec.Body.String()
	if !strings.Contains(html, "IGO密钥认证") {
		t.Fatal("gate page missing auth panel")
	}
	if !strings.Contains(html, "/assets/doc/code-chrome.js") {
		t.Fatal("gate page must include code-chrome.js")
	}
	if !strings.Contains(html, "initDocCodeChrome") {
		t.Fatal("gate boot must re-init code chrome after inject")
	}
	if strings.Contains(html, "TOP_SECRET_BODY_CONTENT") {
		t.Fatal("gate page must not leak article body")
	}
	if !strings.Contains(html, "Secret Notes") {
		t.Fatal("gate page should show title")
	}

	recAPI := httptest.NewRecorder()
	mux.ServeHTTP(recAPI, httptest.NewRequest(http.MethodGet, "/api/doc/secret-article?slug="+slug, nil))
	if recAPI.Code != http.StatusUnauthorized {
		t.Fatalf("unauth API want 401, got %d", recAPI.Code)
	}

	k1, exp1 := obtainIgoKey(dir, false)
	tok := "secret-auth-tok"
	authSessionMu.Lock()
	authSessions[tok] = authSession{igoKey: k1, igoExpiresAt: exp1}
	authSessionMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/doc/secret-article?slug="+slug, nil)
	req.Header.Set("X-Igo-Token", tok)
	recOK := httptest.NewRecorder()
	mux.ServeHTTP(recOK, req)
	if recOK.Code != http.StatusOK {
		t.Fatalf("auth API status=%d %s", recOK.Code, recOK.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(recOK.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true {
		t.Fatalf("ok=%v", resp["ok"])
	}
	bodyHTML, _ := resp["bodyHTML"].(string)
	if !strings.Contains(bodyHTML, "TOP_SECRET_BODY_CONTENT") {
		t.Fatalf("body missing content: %s", bodyHTML)
	}
}

func TestSecretAuthArticle_publicWhenAuthDisabled(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initOpsConfig(dir)
	opsCurrent.authEnabled = false
	opsCurrent.igokeyEnabled = true

	slug := "open_when_auth_off"
	if err := os.WriteFile(linkConfPath(dir), []byte("[secret-auth]\n"+slug+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	body := "# Open\n\nVISIBLE_WITHOUT_AUTH\n"
	_ = os.WriteFile(filepath.Join(assets, "md", slug+".md"), []byte(body), 0o644)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: slug, Title: "Open", Categories: "x",
		PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
	}, []byte(body)); err != nil {
		t.Fatal(err)
	}
	docStore = st
	defer func() { docStore = nil }()

	rec := httptest.NewRecorder()
	if !tryServeDocArticle(rec, httptest.NewRequest(http.MethodGet, "/"+slug, nil), "", slug, dir) {
		t.Fatal("expected serve")
	}
	if !strings.Contains(rec.Body.String(), "VISIBLE_WITHOUT_AUTH") {
		t.Fatal("when auth disabled, article should be public")
	}
	if strings.Contains(rec.Body.String(), "IGO密钥认证") {
		t.Fatal("should not show auth gate when auth disabled")
	}
}
