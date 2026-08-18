package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const mp3GateCookieName = "igozhang_mp3"

var (
	mp3GateKeyOnce sync.Once
	mp3GateKey     []byte

	mp3GateEnabled        = true
	mp3GateTTL            = 12 * time.Hour
	mp3RefererCheck       = true
	mp3MaxConcurrentPerIP = 4
	mp3MaxTracksPerMinute = 40

	mp3StreamSlotMu sync.Mutex
	mp3StreamSlots  = map[string]int{}

	mp3TrackRLMu sync.Mutex
	mp3TrackRL   = map[string]*mp3TrackHits{}
)

type mp3TrackHits struct {
	minute int64
	names  map[string]struct{}
}

func ensureMp3GateKey() {
	mp3GateKeyOnce.Do(func() {
		mp3GateKey = make([]byte, 32)
		if _, err := rand.Read(mp3GateKey); err != nil {
			// 极罕见：退化为固定派生，仍优于明文 "1"
			sum := sha256.Sum256([]byte("igozhang-mp3-gate-fallback"))
			mp3GateKey = sum[:]
		}
	})
}

func mp3GateCookieSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

func signMp3GateValue(expUnix int64) string {
	ensureMp3GateKey()
	msg := strconv.FormatInt(expUnix, 10)
	mac := hmac.New(sha256.New, mp3GateKey)
	_, _ = mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	if len(sig) > 24 {
		sig = sig[:24]
	}
	return msg + "." + sig
}

func verifyMp3GateValue(val string) bool {
	ensureMp3GateKey()
	val = strings.TrimSpace(val)
	i := strings.IndexByte(val, '.')
	if i <= 0 || i >= len(val)-1 {
		return false
	}
	expUnix, err := strconv.ParseInt(val[:i], 10, 64)
	if err != nil || expUnix < time.Now().Unix() {
		return false
	}
	want := signMp3GateValue(expUnix)
	return hmac.Equal([]byte(val), []byte(want))
}

// issueMp3GateCookie 在带 MP3 的页面响应或 /api/mp3/gate 上签发短时 HttpOnly Cookie。
func issueMp3GateCookie(w http.ResponseWriter, r *http.Request) {
	if !mp3GateEnabled {
		return
	}
	ttl := mp3GateTTL
	if ttl < time.Minute {
		ttl = time.Minute
	}
	exp := time.Now().Add(ttl).Unix()
	http.SetCookie(w, &http.Cookie{
		Name:     mp3GateCookieName,
		Value:    signMp3GateValue(exp),
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   mp3GateCookieSecure(r),
	})
}

func mp3GateCookieOK(r *http.Request) bool {
	if !mp3GateEnabled {
		return true
	}
	c, err := r.Cookie(mp3GateCookieName)
	if err != nil || c == nil {
		return false
	}
	return verifyMp3GateValue(c.Value)
}

// mp3RefererOK：有 Referer/Origin 时须与本站 Host 同源；空 Referer 放行（兼容部分移动端）。
func mp3RefererOK(r *http.Request) bool {
	if !mp3GateEnabled || !mp3RefererCheck {
		return true
	}
	reqHost := mp3NormalizeHost(r.Host)
	check := func(raw string) bool {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return true
		}
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return false
		}
		return mp3NormalizeHost(u.Host) == reqHost
	}
	ref := strings.TrimSpace(r.Header.Get("Referer"))
	ori := strings.TrimSpace(r.Header.Get("Origin"))
	if ref == "" && ori == "" {
		return true
	}
	if ref != "" && !check(ref) {
		return false
	}
	if ori != "" && ori != "null" && !check(ori) {
		return false
	}
	return true
}

func mp3NormalizeHost(hostport string) string {
	hostport = strings.ToLower(strings.TrimSpace(hostport))
	if hostport == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		return h
	}
	return hostport
}

// requireMp3Gate 校验 Cookie + Referer；失败写 403。
func requireMp3Gate(w http.ResponseWriter, r *http.Request) bool {
	if !mp3GateEnabled {
		return true
	}
	if !mp3GateCookieOK(r) {
		http.Error(w, "mp3 gate required", http.StatusForbidden)
		return false
	}
	if !mp3RefererOK(r) {
		http.Error(w, "invalid referer", http.StatusForbidden)
		return false
	}
	return true
}

func acquireMp3StreamSlot(ip string) bool {
	if ip == "" || mp3MaxConcurrentPerIP <= 0 {
		return true
	}
	mp3StreamSlotMu.Lock()
	defer mp3StreamSlotMu.Unlock()
	n := mp3StreamSlots[ip]
	if n >= mp3MaxConcurrentPerIP {
		return false
	}
	mp3StreamSlots[ip] = n + 1
	return true
}

func releaseMp3StreamSlot(ip string) {
	if ip == "" {
		return
	}
	mp3StreamSlotMu.Lock()
	defer mp3StreamSlotMu.Unlock()
	n := mp3StreamSlots[ip]
	if n <= 1 {
		delete(mp3StreamSlots, ip)
	} else {
		mp3StreamSlots[ip] = n - 1
	}
}

// allowMp3TrackRL 限制单 IP 每分钟不同曲目数（同曲 Range/重试不计多次）。
func allowMp3TrackRL(w http.ResponseWriter, r *http.Request, track string) bool {
	if !mp3GateEnabled || mp3MaxTracksPerMinute <= 0 {
		return true
	}
	ip := clientIP(r)
	if ip == "" {
		return true
	}
	track = strings.TrimSpace(track)
	if track == "" {
		return true
	}
	now := time.Now().Unix() / 60
	mp3TrackRLMu.Lock()
	defer mp3TrackRLMu.Unlock()
	h := mp3TrackRL[ip]
	if h == nil {
		h = &mp3TrackHits{names: make(map[string]struct{})}
		mp3TrackRL[ip] = h
	}
	if h.minute != now {
		h.minute = now
		h.names = make(map[string]struct{})
	}
	if _, ok := h.names[track]; ok {
		return true
	}
	if len(h.names) >= mp3MaxTracksPerMinute {
		http.Error(w, "too many mp3 tracks from this client", http.StatusTooManyRequests)
		return false
	}
	h.names[track] = struct{}{}
	if len(mp3TrackRL) > 8192 {
		pruneMp3TrackRLLocked(now)
	}
	return true
}

func pruneMp3TrackRLLocked(nowMinute int64) {
	for ip, h := range mp3TrackRL {
		if h == nil || h.minute != nowMinute {
			delete(mp3TrackRL, ip)
		}
	}
}

// mp3GateHandler：刷新门禁 Cookie（供 mp3-init.js 在拉列表/播流前调用）。
func mp3GateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !mp3RefererOK(r) {
			jsonError(w, http.StatusForbidden, "invalid referer")
			return
		}
		issueMp3GateCookie(w, r)
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodHead {
			return
		}
		jsonWrite(w, map[string]any{"ok": true, "enabled": mp3GateEnabled})
	}
}
