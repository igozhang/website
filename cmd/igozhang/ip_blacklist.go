package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

const ipBlacklistFile = "ip-blacklist.conf"

const ipBlacklistFileTemplate = `# IP 黑名单：每行一个 IPv4/IPv6；也支持逗号、空格、分号、- 分隔多个 IP
# 以 # 开头为注释；仅精确 IP，不支持 CIDR；命中后全站返回 403
# 示例：203.0.113.50
# 示例：1.2.3.4, 5.6.7.8
`

var (
	ipBlacklistMu   sync.RWMutex
	ipBlacklistSet  = make(map[string]struct{})
	ipBlacklistRoot string
)

func initIPBlacklist(dataRoot string) {
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()
	ipBlacklistRoot = dataRoot
	ipBlacklistSet = make(map[string]struct{})
	ensureIPBlacklistFile(dataRoot)
	loadIPBlacklistLocked(dataRoot)
}

func ensureIPBlacklistFile(dataRoot string) {
	if dataRoot == "" {
		return
	}
	confPath := ipBlacklistPath(dataRoot)
	if _, err := os.Stat(confPath); err == nil {
		return
	}
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	if err := os.WriteFile(confPath, []byte(ipBlacklistFileTemplate), 0o600); err != nil {
		log.Printf("secret07: write %s: %v", ipBlacklistFile, err)
		return
	}
	log.Printf("secret07: created %s (edit on server as needed)", ipBlacklistFile)
}

func ipBlacklistPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), ipBlacklistFile)
}

func loadIPBlacklistLocked(dataRoot string) {
	if dataRoot == "" {
		return
	}
	raw, err := os.ReadFile(ipBlacklistPath(dataRoot))
	if err != nil {
		return
	}
	for _, ip := range parseIPBlacklistBytes(raw) {
		ipBlacklistSet[ip] = struct{}{}
	}
}

func parseIPBlacklistBytes(raw []byte) []string {
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return nil
	}
	return normalizeBlacklistIPList(tokenizeIPBlacklistText(text))
}

func tokenizeIPBlacklistText(text string) []string {
	var tokens []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		tokens = append(tokens, splitBlacklistLineTokens(line)...)
	}
	return tokens
}

func splitBlacklistLineTokens(line string) []string {
	line = strings.ReplaceAll(line, "，", ",")
	line = strings.ReplaceAll(line, "；", ";")
	var parts []string
	var cur strings.Builder
	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" {
			parts = append(parts, s)
		}
		cur.Reset()
	}
	for _, r := range line {
		switch r {
		case ',', ';', ' ', '\t':
			flush()
		case '-':
			if cur.Len() == 0 {
				continue
			}
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return parts
}

func normalizeBlacklistIPList(in []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(in))
	for _, s := range in {
		ip, ok := normalizeBlacklistIP(s)
		if !ok {
			continue
		}
		if _, dup := seen[ip]; dup {
			continue
		}
		seen[ip] = struct{}{}
		out = append(out, ip)
	}
	sortStrings(out)
	return out
}

func formatIPBlacklistFile(ips []string) []byte {
	var b strings.Builder
	b.WriteString(ipBlacklistFileTemplate)
	for _, ip := range ips {
		b.WriteString(ip)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

func writeIPBlacklistFile(dataRoot string, ips []string) error {
	if dataRoot == "" {
		return nil
	}
	ips = normalizeBlacklistIPList(ips)
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	path := ipBlacklistPath(dataRoot)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, formatIPBlacklistFile(ips), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func persistIPBlacklistLocked(dataRoot string) {
	ips := make([]string, 0, len(ipBlacklistSet))
	for ip := range ipBlacklistSet {
		ips = append(ips, ip)
	}
	sortStrings(ips)
	if err := writeIPBlacklistFile(dataRoot, ips); err != nil {
		log.Printf("ip-blacklist: write %s: %v", ipBlacklistFile, err)
	}
}

// normalizeBlacklistIP 校验并规范化 IPv4/IPv6；不支持 CIDR 或主机名。
func normalizeBlacklistIP(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if strings.ContainsAny(s, "/%") {
		return "", false
	}
	if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
		return "", false
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return "", false
	}
	if ip.To4() == nil && ip.To16() == nil {
		return "", false
	}
	return ip.String(), true
}

func isIPBlacklisted(ip string) bool {
	ip, ok := normalizeBlacklistIP(ip)
	if !ok {
		return false
	}
	ipBlacklistMu.RLock()
	defer ipBlacklistMu.RUnlock()
	_, ok = ipBlacklistSet[ip]
	return ok
}

func listIPBlacklist() []string {
	ipBlacklistMu.RLock()
	defer ipBlacklistMu.RUnlock()
	out := make([]string, 0, len(ipBlacklistSet))
	for ip := range ipBlacklistSet {
		out = append(out, ip)
	}
	sortStrings(out)
	return out
}

func addIPBlacklist(dataRoot, rawIP string) (string, error) {
	ip, ok := normalizeBlacklistIP(rawIP)
	if !ok {
		return "", errInvalidBlacklistIP
	}
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()
	if ipBlacklistRoot == "" && dataRoot != "" {
		ipBlacklistRoot = dataRoot
		loadIPBlacklistLocked(dataRoot)
	}
	ipBlacklistSet[ip] = struct{}{}
	persistIPBlacklistLocked(dataRoot)
	return ip, nil
}

func removeIPBlacklist(dataRoot, rawIP string) (string, error) {
	ip, ok := normalizeBlacklistIP(rawIP)
	if !ok {
		return "", errInvalidBlacklistIP
	}
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()
	if ipBlacklistRoot == "" && dataRoot != "" {
		ipBlacklistRoot = dataRoot
		loadIPBlacklistLocked(dataRoot)
	}
	delete(ipBlacklistSet, ip)
	persistIPBlacklistLocked(dataRoot)
	return ip, nil
}

func mutateIPBlacklist(dataRoot, action, rawIP string) (ip string, list []string, err error) {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "add":
		ip, err = addIPBlacklist(dataRoot, rawIP)
	case "remove":
		ip, err = removeIPBlacklist(dataRoot, rawIP)
	default:
		return "", nil, errInvalidBlacklistAction
	}
	if err != nil {
		return "", nil, err
	}
	return ip, listIPBlacklist(), nil
}

var errInvalidBlacklistAction = &blacklistActionError{}

type blacklistActionError struct{}

func (e *blacklistActionError) Error() string {
	return "invalid action"
}

// readIPBlacklistExportBytes 黑名单非空时导出 conf 全文。
func readIPBlacklistExportBytes(dataRoot string) []byte {
	if dataRoot == "" || len(listIPBlacklist()) == 0 {
		return nil
	}
	raw, err := os.ReadFile(ipBlacklistPath(dataRoot))
	if err != nil || len(raw) == 0 {
		return nil
	}
	return raw
}

// normalizeIPBlacklistImportBytes 校验导入黑名单，规范为 conf 文本。
func normalizeIPBlacklistImportBytes(raw []byte) ([]byte, error) {
	ips := parseIPBlacklistBytes(raw)
	return formatIPBlacklistFile(ips), nil
}

func reloadIPBlacklist(dataRoot string) {
	initIPBlacklist(dataRoot)
}

var errInvalidBlacklistIP = &blacklistIPError{}

type blacklistIPError struct{}

func (e *blacklistIPError) Error() string {
	return "invalid ip"
}

func writeIPForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte("Forbidden"))
}

func sortStrings(ss []string) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0 && ss[j] < ss[j-1]; j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
}
