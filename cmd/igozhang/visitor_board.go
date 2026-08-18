package main

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	visitorOnlineWindow = 5 * time.Minute
	visitorMaxEntries   = 500
	visitorRecentCount  = 5
	visitorEntryTTL     = 24 * time.Hour
	visitorTrailMax     = 12
)

var visitorHTTPBasePath string

func setVisitorHTTPBasePath(prefix string) {
	visitorHTTPBasePath = strings.TrimSuffix(strings.TrimSpace(prefix), "/")
}

func visitorNormalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if base := visitorHTTPBasePath; base != "" && strings.HasPrefix(path, base) {
		path = strings.TrimPrefix(path, base)
		if path == "" {
			return "/"
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func visitorSessionIdleTimeout() time.Duration {
	d := getOpsSnapshot().visitorSessionIdleTimeout
	if d <= 0 {
		return 30 * time.Minute
	}
	return d
}

type visitorActionStep struct {
	TSUnix int64
	Action string
	Target string
}

type visitorEntry struct {
	ip           string
	geo          string
	client       string
	trail        []visitorActionStep
	sessionStart time.Time
	lastSeen     time.Time
}

var (
	visitorBoardMu sync.Mutex
	visitorBoard   = make(map[string]*visitorEntry)
)

func resetVisitorBoardForImport() {
	visitorBoardMu.Lock()
	defer visitorBoardMu.Unlock()
	visitorBoard = make(map[string]*visitorEntry)
}

func requestGate(dataRoot string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := visitorClientIP(r)
		if ip != "" && isIPBlacklisted(ip) {
			writeIPForbidden(w)
			return
		}
		if ip != "" && !checkPathProbeGate(w, r, dataRoot) {
			return
		}
		path := visitorNormalizePath(r.URL.Path)
		if isOpsIgoAdminPath(path) && isIPWhitelistEnforced() && (ip == "" || !isIPWhitelisted(ip)) {
			writeIPForbidden(w)
			return
		}
		recordVisitor(dataRoot, r)
		next.ServeHTTP(w, r)
	})
}

func isOpsIgoAdminPath(path string) bool {
	if strings.HasPrefix(path, "/api/auth/") || path == "/api/auth" {
		return true
	}
	if strings.HasPrefix(path, "/api/ops/") || path == "/api/ops" {
		return true
	}
	if path == "/ops" || strings.HasPrefix(path, "/ops/") {
		return true
	}
	if strings.HasPrefix(path, "/api/igoadmin/") || path == "/api/igoadmin" {
		return true
	}
	if path == "/igoadmin" || strings.HasPrefix(path, "/igoadmin/") {
		return true
	}
	if path == "/fs" || strings.HasPrefix(path, "/fs/") {
		return true
	}
	return false
}

func shouldSkipVisitorRecord(path string) bool {
	if path == "/health" {
		return true
	}
	if strings.HasPrefix(path, "/favicon") {
		return true
	}
	if strings.HasPrefix(path, "/api/ops/") {
		return true
	}
	if strings.HasPrefix(path, "/api/auth/") {
		return true
	}
	switch path {
	case "/api/ops-config", "/api/footer":
		return true
	}
	if strings.HasPrefix(path, "/api/doc/") {
		return true
	}
	if path == "/api/browse" || path == "/api/fs/usage" {
		return true
	}
	return false
}

func parseUserAgentClient(ua string) string {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return ""
	}
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "edg/"):
		return "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		return "Opera"
	case strings.Contains(lower, "firefox/"):
		return "Firefox"
	case strings.Contains(lower, "chrome/") || strings.Contains(lower, "crios/"):
		return "Chrome"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome/") && !strings.Contains(lower, "crios/"):
		return "Safari"
	case strings.Contains(lower, "msie") || strings.Contains(lower, "trident/"):
		return "IE"
	default:
		return ""
	}
}

func visitorPageLabel(path string) string {
	path = strings.TrimSuffix(strings.TrimSpace(path), "/")
	if path == "" || path == "/" {
		return "首页"
	}
	seg := strings.Trim(strings.TrimPrefix(path, "/"), "/")
	if strings.Contains(seg, "/") {
		if strings.HasPrefix(seg, "stream/mp3") || strings.HasPrefix(seg, "api/mp3") {
			return ""
		}
		if strings.HasPrefix(seg, "files/") || strings.HasPrefix(seg, "preview/") {
			return ""
		}
		return ""
	}
	switch strings.ToLower(seg) {
	case "fs":
		return "文件中转"
	case "ops":
		return "运维"
	case "gomoku":
		return "五子棋"
	case "xiangqi":
		return "象棋"
	case "texas":
		return "德州扑克"
	case "zhijian":
		return "知键"
	case "igokey":
		return "igokey"
	case "igoadmin":
		return "文章后台"
	default:
		if docReservedSlug(seg) {
			return ""
		}
		return "/" + seg
	}
}

func visitorFeatureRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	path := visitorNormalizePath(r.URL.Path)
	switch {
	case strings.HasPrefix(path, "/stream/mp3/"), strings.HasPrefix(path, "/api/mp3/"):
		return "mp3"
	case strings.HasPrefix(path, "/api/stash"):
		if r.Method == http.MethodPost {
			return "暂存"
		}
		return ""
	case strings.HasPrefix(path, "/api/upload"):
		if r.Method == http.MethodPost {
			return "上传"
		}
		return ""
	case strings.HasPrefix(path, "/api/fetch"):
		if r.Method == http.MethodPost {
			return "URL拉取"
		}
		return ""
	case path == "/api/file":
		if r.Method == http.MethodDelete {
			return "删除"
		}
		return ""
	default:
		return ""
	}
}

// visitorFeatureLabel 供测试与路径推断；记录访客时须用 visitorFeatureRequest 区分 HTTP 方法。
func visitorFeatureLabel(path string) string {
	path = strings.ToLower(visitorNormalizePath(path))
	switch {
	case strings.HasPrefix(path, "/stream/mp3/"), strings.HasPrefix(path, "/api/mp3/"):
		return "mp3"
	case strings.HasPrefix(path, "/api/stash"):
		return "暂存"
	case strings.HasPrefix(path, "/api/upload"):
		return "上传"
	case strings.HasPrefix(path, "/api/fetch"):
		return "URL拉取"
	case path == "/api/file":
		return "删除"
	default:
		return ""
	}
}

func isVisitorPageRequest(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	path := visitorNormalizePath(r.URL.Path)
	label := visitorPageLabel(path)
	if label == "" {
		return false
	}
	if label == "igokey" {
		return isIgoKeyDirectPageVisit(r)
	}
	return true
}

func visitorActionFromRequest(r *http.Request) (action, target string, ok bool) {
	if r == nil {
		return "", "", false
	}
	if isVisitorPageRequest(r) {
		return accessLogActionVisit, visitorPageLabel(visitorNormalizePath(r.URL.Path)), true
	}
	if feature := visitorFeatureRequest(r); feature != "" {
		return accessLogActionFeature, feature, true
	}
	return "", "", false
}

func appendVisitorTrail(e *visitorEntry, now time.Time, action, target string) bool {
	if e == nil || action == "" || target == "" {
		return false
	}
	n := len(e.trail)
	if n > 0 {
		last := e.trail[n-1]
		if last.Action == action && last.Target == target {
			return false
		}
	}
	e.trail = append(e.trail, visitorActionStep{
		TSUnix: now.Unix(),
		Action: action,
		Target: target,
	})
	if len(e.trail) > visitorTrailMax {
		e.trail = e.trail[len(e.trail)-visitorTrailMax:]
	}
	return true
}

func splitBeijingTs(ts string) (date, timePart string) {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return "", ""
	}
	if i := strings.IndexByte(ts, ' '); i >= 0 {
		return ts[:i], strings.TrimSpace(ts[i+1:])
	}
	return "", ts
}

func formatVisitorTrailCompactStep(s visitorActionStep, prevDate *string) string {
	label := formatAccessActionLabel(s.Action, s.Target)
	if label == "" {
		return ""
	}
	ts := formatBeijingUnix(s.TSUnix)
	if ts == "" {
		return label
	}
	date, timePart := splitBeijingTs(ts)
	if *prevDate == "" || *prevDate != date {
		*prevDate = date
		return ts + " " + label
	}
	if timePart != "" {
		return timePart + " " + label
	}
	return ts + " " + label
}

func formatVisitorTrail(steps []visitorActionStep) string {
	if len(steps) == 0 {
		return ""
	}
	var prevDate string
	parts := make([]string, 0, len(steps))
	for _, s := range steps {
		if step := formatVisitorTrailCompactStep(s, &prevDate); step != "" {
			parts = append(parts, step)
		}
	}
	return strings.Join(parts, " -> ")
}

func visitorTrailToJSON(steps []visitorActionStep) []map[string]any {
	if len(steps) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(steps))
	for _, s := range steps {
		row := map[string]any{
			"action": s.Action,
			"target": s.Target,
			"label":  formatAccessActionLabel(s.Action, s.Target),
		}
		if s.TSUnix > 0 {
			row["tsUnix"] = s.TSUnix
			row["ts"] = formatBeijingUnix(s.TSUnix)
		}
		out = append(out, row)
	}
	return out
}

func visitorGeoZH(ip string) string {
	if ipgeoDB == nil {
		return ""
	}
	res, ok := ipgeoDB.Lookup(ip)
	if !ok {
		return ""
	}
	parts := visitorGeoParts(res, "auto", false)
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func recordVisitor(dataRoot string, r *http.Request) {
	if r == nil {
		return
	}
	path := visitorNormalizePath(r.URL.Path)
	if shouldSkipVisitorRecord(path) {
		return
	}
	action, target, ok := visitorActionFromRequest(r)
	if !ok {
		return
	}
	ip := visitorClientIP(r)
	if ip == "" {
		return
	}
	now := time.Now()
	client := ""
	if ua := strings.TrimSpace(r.Header.Get("User-Agent")); ua != "" {
		client = parseUserAgentClient(ua)
	}

	if strings.TrimSpace(dataRoot) == "" {
		return
	}

	idle := visitorSessionIdleTimeout()
	var pending []accessLogEvent
	needGeo := false

	visitorBoardMu.Lock()
	pruneVisitorBoardLocked(now)

	e := visitorBoard[ip]
	newSession := e == nil || now.Sub(e.lastSeen) > idle
	if e == nil {
		e = &visitorEntry{
			ip:           ip,
			sessionStart: now,
			lastSeen:     now,
		}
		visitorBoard[ip] = e
	} else if newSession {
		e.sessionStart = now
		e.trail = nil
	}
	e.lastSeen = now
	trailAdded := appendVisitorTrail(e, now, action, target)
	if client != "" {
		e.client = client
	}
	if e.geo == "" {
		needGeo = true
	}
	sessionID := formatSessionID(ip, e.sessionStart)

	if newSession {
		pending = append(pending, buildAccessLogEvent(e, now, accessLogActionSession, target, sessionID))
	} else if trailAdded && shouldAppendPathAccessLog(sessionID, action, target, now) {
		pending = append(pending, buildAccessLogEvent(e, now, action, target, sessionID))
	}

	evictVisitorBoardIfNeededLocked(now, idle)
	visitorBoardMu.Unlock()

	if len(pending) == 0 {
		return
	}
	if needGeo {
		if g := visitorGeoZH(ip); g != "" {
			visitorBoardMu.Lock()
			if cur := visitorBoard[ip]; cur != nil && cur.geo == "" {
				cur.geo = g
			}
			visitorBoardMu.Unlock()
			for i := range pending {
				if strings.TrimSpace(pending[i].Geo) == "" {
					pending[i].Geo = g
				}
			}
		}
	}
	for _, ev := range pending {
		appendAccessLog(dataRoot, ev)
	}
}

func pruneVisitorBoardLocked(now time.Time) {
	cutoff := now.Add(-visitorEntryTTL)
	for ip, e := range visitorBoard {
		if e == nil || e.lastSeen.Before(cutoff) {
			delete(visitorBoard, ip)
		}
	}
}

func evictVisitorBoardIfNeededLocked(now time.Time, idle time.Duration) {
	if len(visitorBoard) <= visitorMaxEntries {
		return
	}
	type pair struct {
		ip string
		t  time.Time
	}
	pairs := make([]pair, 0, len(visitorBoard))
	for ip, e := range visitorBoard {
		if e == nil {
			continue
		}
		if now.Sub(e.lastSeen) <= idle {
			continue
		}
		pairs = append(pairs, pair{ip: ip, t: e.lastSeen})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].t.Before(pairs[j].t)
	})
	for len(visitorBoard) > visitorMaxEntries && len(pairs) > 0 {
		delete(visitorBoard, pairs[0].ip)
		pairs = pairs[1:]
	}
}

func visitorEntryToJSON(e *visitorEntry) map[string]any {
	if e == nil {
		return nil
	}
	out := map[string]any{"ip": e.ip}
	if g := strings.TrimSpace(e.geo); g != "" {
		out["geo"] = g
	}
	if c := strings.TrimSpace(e.client); c != "" {
		out["client"] = c
	}
	if !e.sessionStart.IsZero() {
		out["loginAt"] = formatBeijingUnix(e.sessionStart.Unix())
		out["loginAtUnix"] = e.sessionStart.Unix()
	}
	if trail := visitorTrailToJSON(e.trail); len(trail) > 0 {
		out["trail"] = trail
		if text := formatVisitorTrail(e.trail); text != "" {
			out["trailText"] = text
		}
	}
	out["lastSeen"] = e.lastSeen.Unix()
	return out
}

func visitorBoardSnapshot() (online, recent []map[string]any) {
	now := time.Now()
	onlineCutoff := now.Add(-visitorOnlineWindow)

	visitorBoardMu.Lock()
	defer visitorBoardMu.Unlock()
	pruneVisitorBoardLocked(now)

	entries := make([]*visitorEntry, 0, len(visitorBoard))
	for _, e := range visitorBoard {
		if e != nil {
			entries = append(entries, e)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].lastSeen.After(entries[j].lastSeen)
	})

	for _, e := range entries {
		if e.lastSeen.After(onlineCutoff) {
			online = append(online, visitorEntryToJSON(e))
		}
	}
	onlineSet := make(map[string]struct{}, len(online))
	for _, item := range online {
		if ip, _ := item["ip"].(string); ip != "" {
			onlineSet[ip] = struct{}{}
		}
	}
	for _, e := range entries {
		if _, on := onlineSet[e.ip]; on {
			continue
		}
		if !e.lastSeen.After(onlineCutoff) {
			recent = append(recent, visitorEntryToJSON(e))
			if len(recent) >= visitorRecentCount {
				break
			}
		}
	}
	if online == nil {
		online = []map[string]any{}
	}
	if recent == nil {
		recent = []map[string]any{}
	}
	return online, recent
}
