package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	igoKeyLen        = 8
	igoKeyTTL        = time.Hour
	igoKeyReuseAfter = 15 * time.Minute
	igoKeyAlphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	igoKeyMaxEntries = 4096
)

//go:embed igokey.html
var igokeyPageTpl []byte

type igoKeyEntry struct {
	expiresAt time.Time
}

var (
	igoKeyMu     sync.RWMutex
	igoKeyStore  = map[string]igoKeyEntry{}
	igoKeyActive struct {
		key       string
		expiresAt time.Time
	}
)

func registerIgoKeyHandler(mux *http.ServeMux, dataRoot string) {
	mux.HandleFunc("/igokey", func(w http.ResponseWriter, r *http.Request) {
		igokeyHTTPHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/igokey/", func(w http.ResponseWriter, r *http.Request) {
		igokeyHTTPHandler(w, r, dataRoot)
	})
}

func igoKeyEnabled() bool {
	return getOpsSnapshot().igokeyEnabled
}

func igokeyHTTPHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !igoKeyEnabled() {
		http.NotFound(w, r)
		return
	}
	// 关闭 HTML 页面访问；仅保留 format=plain/json（页脚复制与程序取钥）。
	if !wantsIgoKeyJSON(r) && !wantsIgoKeyPlain(r) {
		http.NotFound(w, r)
		return
	}
	forceNew := igoKeyForceRenew(r)
	key, expiresAt := obtainIgoKey(dataRoot, forceNew)
	if r.Method == http.MethodHead {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Igo-Key-Expires", expiresAt.UTC().Format(time.RFC3339))
		return
	}
	if wantsIgoKeyJSON(r) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"key":          key,
			"expiresAt":    expiresAt.Unix(),
			"expiresInSec": int(time.Until(expiresAt).Seconds()),
		})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(key))
}

func wantsIgoKeyJSON(r *http.Request) bool {
	if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "json") {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json") &&
		!strings.Contains(accept, "text/html")
}

func wantsIgoKeyPlain(r *http.Request) bool {
	f := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	return f == "text" || f == "plain"
}

// isIgoKeyDirectPageVisit 仅用户直接打开 /igokey HTML 页时记访客；页脚预取 format=plain/json 等不计。
func isIgoKeyDirectPageVisit(r *http.Request) bool {
	if r == nil || r.Method != http.MethodGet {
		return false
	}
	path := strings.TrimSuffix(strings.TrimSpace(r.URL.Path), "/")
	if path != "/igokey" {
		return false
	}
	if wantsIgoKeyPlain(r) || wantsIgoKeyJSON(r) {
		return false
	}
	return true
}

func igoKeyForceRenew(r *http.Request) bool {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("renew")))
	return q == "1" || q == "true" || q == "yes"
}

// obtainIgoKey 签发或复用站点密钥：剩余 >15min 的普通刷新复用；≤15min 刷新或 renew=1 时签发新密钥并作废旧密钥。
func obtainIgoKey(dataRoot string, forceNew bool) (key string, expiresAt time.Time) {
	now := time.Now()
	var revokeKey string

	igoKeyMu.Lock()
	syncIgoKeyActiveFromDiskLocked(dataRoot, now)
	pruneIgoKeyStoreLocked(now)

	if !forceNew {
		if k, exp, ok := activeIgoKeyReusableLocked(now); ok {
			ensureIgoKeyPlainMatchesLocked(dataRoot, k)
			igoKeyMu.Unlock()
			return k, exp
		}
	}

	if igoKeyActive.key != "" {
		revokeKey = igoKeyActive.key
		delete(igoKeyStore, revokeKey)
		igoKeyActive.key = ""
		igoKeyActive.expiresAt = time.Time{}
	}

	key, expiresAt = issueNewIgoKeyLocked(now)
	igoKeyActive.key = key
	igoKeyActive.expiresAt = expiresAt
	persistIgoKeyActiveLocked(dataRoot, key, expiresAt)
	igoKeyMu.Unlock()

	if revokeKey != "" && revokeKey != key {
		revokeAuthSessionsForIgoKey(revokeKey)
	}
	return key, expiresAt
}

// activeIgoKeyReusableLocked 当前活跃密钥仍有效且剩余 >15min 时可复用（调用方已持 igoKeyMu）。
func activeIgoKeyReusableLocked(now time.Time) (key string, expiresAt time.Time, ok bool) {
	if igoKeyActive.key == "" {
		return "", time.Time{}, false
	}
	e, found := igoKeyStore[igoKeyActive.key]
	if !found || !e.expiresAt.After(now) {
		return "", time.Time{}, false
	}
	if e.expiresAt.Sub(now) <= igoKeyReuseAfter {
		return "", time.Time{}, false
	}
	igoKeyActive.expiresAt = e.expiresAt
	return igoKeyActive.key, e.expiresAt, true
}

func issueNewIgoKeyLocked(now time.Time) (key string, expiresAt time.Time) {
	expiresAt = now.Add(igoKeyTTL)
	for i := 0; i < 8; i++ {
		k, err := randomIgoKeyString(igoKeyLen)
		if err != nil {
			log.Printf("igokey: random: %v", err)
			continue
		}
		if len(igoKeyStore) >= igoKeyMaxEntries {
			log.Printf("igokey: store full, dropping new key issuance")
			igoKeyStore[k] = igoKeyEntry{expiresAt: expiresAt}
			return k, expiresAt
		}
		if _, exists := igoKeyStore[k]; exists {
			continue
		}
		igoKeyStore[k] = igoKeyEntry{expiresAt: expiresAt}
		return k, expiresAt
	}
	k, _ := randomIgoKeyString(igoKeyLen)
	igoKeyStore[k] = igoKeyEntry{expiresAt: expiresAt}
	return k, expiresAt
}

func randomIgoKeyString(n int) (string, error) {
	var b strings.Builder
	b.Grow(n)
	max := big.NewInt(int64(len(igoKeyAlphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(igoKeyAlphabet[idx.Int64()])
	}
	return b.String(), nil
}

func pruneIgoKeyStoreLocked(now time.Time) {
	for k, e := range igoKeyStore {
		if !e.expiresAt.After(now) {
			delete(igoKeyStore, k)
			if igoKeyActive.key == k {
				igoKeyActive.key = ""
				igoKeyActive.expiresAt = time.Time{}
			}
		}
	}
}

// igoKeyExpiry 返回当前活跃密钥的过期时刻（须与 key 一致且仍有效）。
func igoKeyExpiry(key string) (time.Time, bool) {
	key = strings.TrimSpace(key)
	if len(key) != igoKeyLen {
		return time.Time{}, false
	}
	now := time.Now()
	igoKeyMu.RLock()
	defer igoKeyMu.RUnlock()
	if igoKeyActive.key != key {
		return time.Time{}, false
	}
	e, ok := igoKeyStore[key]
	if !ok || !e.expiresAt.After(now) {
		return time.Time{}, false
	}
	return e.expiresAt, true
}

// validateIgoKey 校验是否为当前活跃且未过期的站点密钥（供 ops 等使用）。
func validateIgoKey(key string) bool {
	key = strings.TrimSpace(key)
	if len(key) != igoKeyLen {
		return false
	}
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	now := time.Now()
	igoKeyMu.RLock()
	defer igoKeyMu.RUnlock()
	if igoKeyActive.key != key {
		return false
	}
	e, ok := igoKeyStore[key]
	return ok && e.expiresAt.After(now)
}

func renderIgoKeyHTML(pub, key string, expiresAt time.Time, r *http.Request) []byte {
	escKey := html.EscapeString(key)
	remainSec := int(time.Until(expiresAt).Seconds())
	if remainSec < 0 {
		remainSec = 0
	}
	return applySitePage(string(igokeyPageTpl), pub, sitePageLayout{}, map[string]string{
		"{{IGO_KEY}}":            escKey,
		"{{IGO_KEY_JS}}":         templateJSEscape(key),
		"{{IGO_EXPIRES_IN_SEC}}": fmt.Sprintf("%d", remainSec),
		"{{IGO_RENEW_HREF}}":     html.EscapeString(pubPath(pub, "igokey") + "?renew=1"),
	}, visitorGeoLabel(r))
}

func templateJSEscape(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}
