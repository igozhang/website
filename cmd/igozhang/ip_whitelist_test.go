package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureIPWhitelistFile_createsTemplate(t *testing.T) {
	dir := t.TempDir()
	ensureIPWhitelistFile(dir)
	raw, err := os.ReadFile(ipWhitelistPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "#") {
		t.Fatalf("want comment header: %q", raw)
	}
	if len(parseIPBlacklistBytes(raw)) != 0 {
		t.Fatal("template should have no effective ips")
	}
}

func TestIPWhitelistPersistAndReload(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initIPWhitelist(dir)

	raw := []byte("# test\n203.0.113.10\n1.2.3.4\n")
	norm, err := normalizeIPWhitelistImportBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ipWhitelistPath(dir), norm, 0o600); err != nil {
		t.Fatal(err)
	}
	reloadIPWhitelist(dir)

	if !isIPWhitelistEnforced() {
		t.Fatal("should enforce when non-empty")
	}
	if !isIPWhitelisted("203.0.113.10") {
		t.Fatal("listed ip should pass")
	}
	if isIPWhitelisted("9.9.9.9") {
		t.Fatal("unlisted ip should fail when enforced")
	}

	initIPWhitelist(dir)
	if !isIPWhitelisted("203.0.113.10") {
		t.Fatal("reload listed ip")
	}

	if err := os.WriteFile(ipWhitelistPath(dir), []byte(ipWhitelistFileTemplate), 0o600); err != nil {
		t.Fatal(err)
	}
	reloadIPWhitelist(dir)
	if isIPWhitelistEnforced() {
		t.Fatal("empty list should not enforce")
	}
	if !isIPWhitelisted("9.9.9.9") {
		t.Fatal("empty list should allow any ip")
	}

	got, err := os.ReadFile(filepath.Join(secret07Dir(dir), ipWhitelistFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(got), "#") {
		t.Fatal("cleared list should keep comment template")
	}
}
