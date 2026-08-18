package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	authReasonKeyRevoked = "key_revoked"
	authReasonExpired    = "key_expired"
	authReasonInvalid    = "invalid_token"
)

type authTokenState int

const (
	authTokenOK authTokenState = iota
	authTokenMissing
	authTokenExpired
	authTokenKeyRevoked
)

type authSession struct {
	igoKey       string
	igoExpiresAt time.Time
}

var (
	authSessionMu sync.RWMutex
	authSessions  = map[string]authSession{}
)

func authFeatureEnabled() bool {
	s := getOpsSnapshot()
	return s.authEnabled && s.igokeyEnabled
}

func syncIgoKeyForAuth(dataRoot string) {
	if dataRoot == "" {
		return
	}
	now := time.Now()
	igoKeyMu.Lock()
	syncIgoKeyActiveFromDiskLocked(dataRoot, now)
	igoKeyMu.Unlock()
}

func authTokenStateOf(dataRoot, tok string) (authTokenState, authSession) {
	if tok == "" {
		return authTokenMissing, authSession{}
	}
	syncIgoKeyForAuth(dataRoot)
	now := time.Now()
	authSessionMu.RLock()
	s, ok := authSessions[tok]
	authSessionMu.RUnlock()
	if !ok {
		return authTokenKeyRevoked, authSession{}
	}
	if !s.igoExpiresAt.After(now) {
		authSessionMu.Lock()
		delete(authSessions, tok)
		authSessionMu.Unlock()
		return authTokenExpired, s
	}
	if !validateIgoKey(s.igoKey) {
		authSessionMu.Lock()
		delete(authSessions, tok)
		authSessionMu.Unlock()
		return authTokenKeyRevoked, s
	}
	return authTokenOK, s
}

func authTokenValid(dataRoot, tok string) bool {
	st, _ := authTokenStateOf(dataRoot, tok)
	return st == authTokenOK
}

func authUnauthorized(w http.ResponseWriter, reason, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":     false,
		"reason": reason,
		"error":  msg,
	})
}

func authFailMessage(reason string) string {
	switch reason {
	case authReasonKeyRevoked:
		return "旧 key 不可用，请重新验证"
	case authReasonExpired:
		return "key 已过期，请重新验证"
	default:
		return "未认证，请先验证 key"
	}
}

// requireAuthToken 校验 X-Igo-Token 或 ?token=（与 /api/auth/login 签发的会话一致）。
func requireAuthToken(w http.ResponseWriter, r *http.Request, dataRoot string) bool {
	if !authFeatureEnabled() {
		http.NotFound(w, r)
		return false
	}
	tok := strings.TrimSpace(r.Header.Get("X-Igo-Token"))
	if tok == "" {
		tok = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	st, _ := authTokenStateOf(dataRoot, tok)
	if st != authTokenOK {
		reason := authReasonInvalid
		if st == authTokenKeyRevoked {
			reason = authReasonKeyRevoked
		} else if st == authTokenExpired {
			reason = authReasonExpired
		}
		authUnauthorized(w, reason, authFailMessage(reason))
		return false
	}
	return true
}

func registerAuthHandlers(mux *http.ServeMux, dataRoot string) {
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		authLoginHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/auth/session", func(w http.ResponseWriter, r *http.Request) {
		authSessionHandler(w, r, dataRoot)
	})
}

func authLoginHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !authFeatureEnabled() {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !checkAuthFailRL(w, r, dataRoot) {
		return
	}
	syncIgoKeyForAuth(dataRoot)
	key := strings.TrimSpace(body.Key)
	exp, ok := igoKeyExpiry(key)
	if !ok {
		if locked, retrySec := recordAuthFail(r, dataRoot); locked {
			writeAuthRateLimited(w, retrySec)
			return
		}
		jsonError(w, http.StatusUnauthorized, "invalid or expired key")
		return
	}
	clearAuthFail(r, dataRoot)
	tok, err := newAuthSessionToken()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "session error")
		return
	}
	authSessionMu.Lock()
	pruneAuthSessionsLocked(time.Now())
	authSessions[tok] = authSession{igoKey: key, igoExpiresAt: exp}
	authSessionMu.Unlock()

	clearLegacyAuthSessionCookie(w, r)

	jsonWrite(w, map[string]any{
		"ok":           true,
		"token":        tok,
		"expiresAt":    exp.Unix(),
		"expiresInSec": int(time.Until(exp).Seconds()),
	})
}

func authSessionHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !authFeatureEnabled() {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	st, sess := authTokenStateOf(dataRoot, body.Token)
	switch st {
	case authTokenOK:
		jsonWrite(w, map[string]any{
			"ok":           true,
			"expiresInSec": int(time.Until(sess.igoExpiresAt).Seconds()),
		})
	case authTokenKeyRevoked:
		authUnauthorized(w, authReasonKeyRevoked, authFailMessage(authReasonKeyRevoked))
	case authTokenExpired:
		authUnauthorized(w, authReasonExpired, authFailMessage(authReasonExpired))
	default:
		authUnauthorized(w, authReasonInvalid, authFailMessage(authReasonInvalid))
	}
}

func newAuthSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func pruneAuthSessionsLocked(now time.Time) {
	for k, s := range authSessions {
		if !s.igoExpiresAt.After(now) {
			delete(authSessions, k)
		}
	}
}

// revokeAuthSessionsForIgoKey 新密钥签发后，使用该旧 key 认证的会话立即失效。
func revokeAuthSessionsForIgoKey(igoKey string) {
	if igoKey == "" {
		return
	}
	now := time.Now()
	authSessionMu.Lock()
	for tok, s := range authSessions {
		if s.igoKey == igoKey {
			delete(authSessions, tok)
		}
	}
	pruneAuthSessionsLocked(now)
	authSessionMu.Unlock()
}

// clearLegacyAuthSessionCookie 清除旧版 Cookie 会话，刷新后不再自动登录。
func clearLegacyAuthSessionCookie(w http.ResponseWriter, r *http.Request) {
	pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
	if pub == "" {
		if p := strings.TrimSpace(getenv("BASE_PATH", "")); p != "" {
			pub = strings.TrimSuffix(p, "/")
		}
	}
	path := "/"
	if pub != "" {
		path = pub
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "fs_ops_sess",
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
