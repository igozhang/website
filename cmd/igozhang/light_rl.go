package main

import (
	"net/http"
	"sync"
	"time"
)

// 轻量按 IP 滑动分钟窗，防刷 /api/browse 与 /api/stash（不含鉴权场景下的滥用）。

var (
	browseRLPerMinute = 150
	stashRLPerMinute  = 80
)

type minuteHits struct {
	minute int64
	n      int
}

var (
	browseRLMu sync.Mutex
	browseRL   = make(map[string]*minuteHits)
	stashRLMu  sync.Mutex
	stashRL    = make(map[string]*minuteHits)
)

func allowBrowseRL(w http.ResponseWriter, r *http.Request) bool {
	return allowMinuteRL(w, r, &browseRLMu, browseRL, browseRLPerMinute)
}

func allowStashRL(w http.ResponseWriter, r *http.Request) bool {
	return allowMinuteRL(w, r, &stashRLMu, stashRL, stashRLPerMinute)
}

func allowMinuteRL(w http.ResponseWriter, r *http.Request, mu *sync.Mutex, m map[string]*minuteHits, max int) bool {
	ip := clientIP(r)
	if ip == "" {
		return true
	}
	now := time.Now().Unix() / 60
	mu.Lock()
	defer mu.Unlock()
	h := m[ip]
	if h == nil {
		h = &minuteHits{}
		m[ip] = h
	}
	if h.minute != now {
		h.minute = now
		h.n = 0
	}
	if h.n >= max {
		jsonError(w, http.StatusTooManyRequests, "too many requests from this client")
		return false
	}
	h.n++
	return true
}
