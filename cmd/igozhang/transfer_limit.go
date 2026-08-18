package main

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	transferSlotMu sync.Mutex
	transferSlots  map[string]int
	maxTransferConc int
)

func initTransferLimits() {
	maxTransferConc = getOpsSnapshot().transferMaxConcurrent
	if s := getenv("TRANSFER_MAX_CONCURRENT_PER_IP", ""); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 32 {
			maxTransferConc = n
		}
	}
	transferSlots = make(map[string]int)
}

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	ff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if ff != "" {
		if i := strings.IndexByte(ff, ','); i >= 0 {
			ff = strings.TrimSpace(ff[:i])
		}
		return ff
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func acquireTransfer(ip string) bool {
	transferSlotMu.Lock()
	defer transferSlotMu.Unlock()
	n := transferSlots[ip]
	if n >= maxTransferConc {
		return false
	}
	transferSlots[ip] = n + 1
	return true
}

func releaseTransfer(ip string) {
	transferSlotMu.Lock()
	defer transferSlotMu.Unlock()
	n := transferSlots[ip]
	if n <= 1 {
		delete(transferSlots, ip)
	} else {
		transferSlots[ip] = n - 1
	}
}
