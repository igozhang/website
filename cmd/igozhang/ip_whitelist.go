package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const ipWhitelistFile = "ip-whitelist.conf"

const ipWhitelistFileTemplate = `# IP 白名单：每行一个 IPv4/IPv6；也支持逗号、空格、分号、- 分隔多个 IP
# 以 # 开头为注释；仅精确 IP，不支持 CIDR
# 仅对 /ops 与 /igoadmin 生效；列表非空时，仅白名单内 IP 可访问（否则 403）
# 列表为空时不限制（与黑名单「空列表不拦截」对称）
# 示例：203.0.113.50
# 示例：1.2.3.4, 5.6.7.8
`

var (
	ipWhitelistMu   sync.RWMutex
	ipWhitelistSet  = make(map[string]struct{})
	ipWhitelistRoot string
)

func initIPWhitelist(dataRoot string) {
	ipWhitelistMu.Lock()
	defer ipWhitelistMu.Unlock()
	ipWhitelistRoot = dataRoot
	ipWhitelistSet = make(map[string]struct{})
	ensureIPWhitelistFile(dataRoot)
	loadIPWhitelistLocked(dataRoot)
}

func ensureIPWhitelistFile(dataRoot string) {
	if dataRoot == "" {
		return
	}
	confPath := ipWhitelistPath(dataRoot)
	if _, err := os.Stat(confPath); err == nil {
		return
	}
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	if err := os.WriteFile(confPath, []byte(ipWhitelistFileTemplate), 0o600); err != nil {
		log.Printf("secret07: write %s: %v", ipWhitelistFile, err)
		return
	}
	log.Printf("secret07: created %s (edit on server as needed)", ipWhitelistFile)
}

func ipWhitelistPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), ipWhitelistFile)
}

func loadIPWhitelistLocked(dataRoot string) {
	if dataRoot == "" {
		return
	}
	raw, err := os.ReadFile(ipWhitelistPath(dataRoot))
	if err != nil {
		return
	}
	for _, ip := range parseIPBlacklistBytes(raw) {
		ipWhitelistSet[ip] = struct{}{}
	}
}

func formatIPWhitelistFile(ips []string) []byte {
	var b strings.Builder
	b.WriteString(ipWhitelistFileTemplate)
	for _, ip := range ips {
		b.WriteString(ip)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

func listIPWhitelist() []string {
	ipWhitelistMu.RLock()
	defer ipWhitelistMu.RUnlock()
	out := make([]string, 0, len(ipWhitelistSet))
	for ip := range ipWhitelistSet {
		out = append(out, ip)
	}
	sortStrings(out)
	return out
}

func isIPWhitelistEnforced() bool {
	ipWhitelistMu.RLock()
	defer ipWhitelistMu.RUnlock()
	return len(ipWhitelistSet) > 0
}

// isIPWhitelisted 白名单为空时不限制；非空时仅列出的 IP 通过。
func isIPWhitelisted(ip string) bool {
	ip, ok := normalizeBlacklistIP(ip)
	if !ok {
		return false
	}
	ipWhitelistMu.RLock()
	defer ipWhitelistMu.RUnlock()
	if len(ipWhitelistSet) == 0 {
		return true
	}
	_, ok = ipWhitelistSet[ip]
	return ok
}

// readIPWhitelistExportBytes 白名单非空时导出 conf 全文。
func readIPWhitelistExportBytes(dataRoot string) []byte {
	if dataRoot == "" || len(listIPWhitelist()) == 0 {
		return nil
	}
	raw, err := os.ReadFile(ipWhitelistPath(dataRoot))
	if err != nil || len(raw) == 0 {
		return nil
	}
	return raw
}

// normalizeIPWhitelistImportBytes 校验导入白名单，规范为 conf 文本。
func normalizeIPWhitelistImportBytes(raw []byte) ([]byte, error) {
	ips := parseIPBlacklistBytes(raw)
	return formatIPWhitelistFile(ips), nil
}

func reloadIPWhitelist(dataRoot string) {
	initIPWhitelist(dataRoot)
}
