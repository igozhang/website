package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed ops.html
var opsPageTpl []byte

var (
	opsExecRLMu        sync.Mutex
	opsExecRL          = make(map[string]*minuteHits)
	opsPoolJanitorOnce sync.Once
)

const opsExecRLPerMinute = 30

func opsPageEnabled() bool {
	return getOpsSnapshot().opsEnabled
}

func registerOpsHandlers(mux *http.ServeMux, dataRoot string) {
	opsPoolJanitorOnce.Do(func() { go opsSSHPool.runJanitor() })
	mux.HandleFunc("/ops", func(w http.ResponseWriter, r *http.Request) {
		opsPageHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/ops/", func(w http.ResponseWriter, r *http.Request) {
		opsPageHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/ops/targets", func(w http.ResponseWriter, r *http.Request) {
		opsTargetsHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/ops/exec", func(w http.ResponseWriter, r *http.Request) {
		opsExecHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/ops/board", func(w http.ResponseWriter, r *http.Request) {
		opsBoardHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/ops/stats-detail", func(w http.ResponseWriter, r *http.Request) {
		opsStatsDetailHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/ops/blacklist", func(w http.ResponseWriter, r *http.Request) {
		opsBlacklistHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/ops/auto-ban", func(w http.ResponseWriter, r *http.Request) {
		opsAutoBanHandler(w, r, dataRoot)
	})
}

func opsPageHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !opsPageEnabled() {
		http.NotFound(w, r)
		return
	}
	pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
	if pub == "" {
		if p := strings.TrimSpace(getenv("BASE_PATH", "")); p != "" {
			pub = strings.TrimSuffix(p, "/")
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(renderOpsHTML(pub, dataRoot, r))
}

func renderOpsHTML(pub, dataRoot string, r *http.Request) []byte {
	sshReady := "false"
	if tgts, err := listSSHTargets(dataRoot); err == nil && len(tgts) > 0 {
		sshReady = "true"
	}
	apiPub := strings.TrimSuffix(pub, "/")
	baseJS, err := json.Marshal(apiPub)
	if err != nil {
		baseJS = []byte(`""`)
	}
	cmds := listCommonCommands(dataRoot)
	if cmds == nil {
		cmds = []string{}
	}
	cmdsJS, err := json.Marshal(cmds)
	if err != nil {
		cmdsJS = []byte("[]")
	}
	return applySitePage(string(opsPageTpl), pub, sitePageLayout{}, map[string]string{
		"{{OPS_SSH_READY}}":   sshReady,
		"{{OPS_BASE}}":        string(baseJS),
		"{{OPS_COMMON_CMDS}}": string(cmdsJS),
		"{{IGO_AUTH_PANEL}}":  igoAuthPanelHTML,
		"{{IGO_AUTH_CSS}}":    igoAuthCSS,
		"{{IGO_AUTH_JS}}":     igoAuthScriptTag(),
	}, visitorGeoLabel(r))
}

func opsTargetsJSON(dataRoot string) []map[string]any {
	tgts, err := listSSHTargets(dataRoot)
	if err != nil {
		return nil
	}
	out := make([]map[string]any, 0, len(tgts))
	for _, t := range tgts {
		out = append(out, map[string]any{
			"id":    t.ID,
			"label": t.label(),
		})
	}
	return out
}

func opsRequireAuthToken(w http.ResponseWriter, r *http.Request, dataRoot, bodyToken string) bool {
	if !authFeatureEnabled() {
		http.NotFound(w, r)
		return false
	}
	tok := strings.TrimSpace(bodyToken)
	if tok == "" {
		tok = strings.TrimSpace(r.Header.Get("X-Igo-Token"))
	}
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

func opsTargetsHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !opsPageEnabled() {
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
	if !opsRequireAuthToken(w, r, dataRoot, body.Token) {
		return
	}
	targets := opsTargetsJSON(dataRoot)
	if targets == nil {
		targets = []map[string]any{}
	}
	jsonWrite(w, map[string]any{
		"ok":      true,
		"targets": targets,
	})
}

func opsExecHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !opsPageEnabled() {
		http.NotFound(w, r)
		return
	}
	if !allowMinuteRL(w, r, &opsExecRLMu, opsExecRL, opsExecRLPerMinute) {
		return
	}
	var body struct {
		Token   string `json:"token"`
		Command string `json:"command"`
		Target  int    `json:"target"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, opsSSHMaxCmdBytes+512)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !opsRequireAuthToken(w, r, dataRoot, body.Token) {
		return
	}
	stdout, stderr, code, err := runSSHCommand(dataRoot, body.Target, body.Command)
	if err != nil {
		jsonWrite(w, map[string]any{
			"ok":        false,
			"error":     err.Error(),
			"stdout":    stdout,
			"stderr":    stderr,
			"exitCode":  code,
			"stdoutLen": len(stdout),
			"stderrLen": len(stderr),
		})
		return
	}
	jsonWrite(w, map[string]any{
		"ok":        true,
		"stdout":    stdout,
		"stderr":    stderr,
		"exitCode":  code,
		"stdoutLen": len(stdout),
		"stderrLen": len(stderr),
	})
}

func opsBoardHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !opsPageEnabled() {
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
	if !opsRequireAuthToken(w, r, dataRoot, body.Token) {
		return
	}
	online, recent := visitorBoardSnapshot()
	stats := countAccessLogStats(dataRoot, time.Now())
	pollMs := int(getOpsSnapshot().boardPollInterval / time.Millisecond)
	jsonWrite(w, map[string]any{
		"ok":                  true,
		"online":              online,
		"recent":              recent,
		"blacklist":           listIPBlacklist(),
		"autoBans":            listAutoBansJSON(),
		"onlineWindowSec":     int(visitorOnlineWindow.Seconds()),
		"boardPollIntervalMs": pollMs,
		"stats":               stats,
		"insights":            accessBoardInsightsSnapshot(),
		"logFiles":            siteLogBoardFiles(dataRoot),
	})
}

func opsStatsDetailHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !opsPageEnabled() {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Token  string `json:"token"`
		Period string `json:"period"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !opsRequireAuthToken(w, r, dataRoot, body.Token) {
		return
	}
	period := strings.TrimSpace(strings.ToLower(body.Period))
	since, ok := accessLogPeriodCutoff(period, time.Now())
	if !ok {
		jsonError(w, http.StatusBadRequest, "invalid period")
		return
	}
	events := listRecentAccessLogSessionDetails(dataRoot, since, accessLogDetailLimit)
	labels := map[string]string{"day": "近一日", "week": "近一周", "month": "近一月"}
	jsonWrite(w, map[string]any{
		"ok":      true,
		"period":  period,
		"label":   labels[period],
		"records": events,
	})
}

func opsBlacklistHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !opsPageEnabled() {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Token  string `json:"token"`
		Action string `json:"action"`
		IP     string `json:"ip"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !opsRequireAuthToken(w, r, dataRoot, body.Token) {
		return
	}
	action := strings.TrimSpace(strings.ToLower(body.Action))
	ip, list, err := mutateIPBlacklist(dataRoot, action, body.IP)
	if err != nil {
		if err == errInvalidBlacklistAction {
			jsonError(w, http.StatusBadRequest, "invalid action")
			return
		}
		jsonError(w, http.StatusBadRequest, "无效 IP（仅支持精确 IPv4 或 IPv6）")
		return
	}
	jsonWrite(w, map[string]any{"ok": true, "ip": ip, "blacklist": list})
}

func opsAutoBanHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !opsPageEnabled() {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Token  string `json:"token"`
		Action string `json:"action"`
		IP     string `json:"ip"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !opsRequireAuthToken(w, r, dataRoot, body.Token) {
		return
	}
	action := strings.TrimSpace(strings.ToLower(body.Action))
	if action != "remove" {
		jsonError(w, http.StatusBadRequest, "invalid action")
		return
	}
	ip, err := removeAutoBan(dataRoot, body.IP)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "无效 IP（仅支持精确 IPv4 或 IPv6）")
		return
	}
	jsonWrite(w, map[string]any{"ok": true, "ip": ip, "autoBans": listAutoBansJSON()})
}
