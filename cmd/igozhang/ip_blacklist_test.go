package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeBlacklistIP(t *testing.T) {
	cases := []struct {
		in     string
		ok     bool
		normal string
	}{
		{"1.2.3.4", true, "1.2.3.4"},
		{"  1.2.3.4  ", true, "1.2.3.4"},
		{"2001:db8::1", true, "2001:db8::1"},
		{"10.0.0.0/8", false, ""},
		{"evil.com", false, ""},
		{"", false, ""},
	}
	for _, c := range cases {
		got, ok := normalizeBlacklistIP(c.in)
		if ok != c.ok || got != c.normal {
			t.Fatalf("in=%q got=%q ok=%v want=%q ok=%v", c.in, got, ok, c.normal, c.ok)
		}
	}
}

func TestParseIPBlacklistText_flexible(t *testing.T) {
	raw := `# header
1.2.3.4, 5.6.7.8
9.9.9.9 - 8.8.8.8
2001:db8::5
`
	ips := parseIPBlacklistBytes([]byte(raw))
	if len(ips) != 5 {
		t.Fatalf("ips=%v", ips)
	}
}

func TestEnsureIPBlacklistFile_createsTemplate(t *testing.T) {
	dir := t.TempDir()
	ensureIPBlacklistFile(dir)
	raw, err := os.ReadFile(ipBlacklistPath(dir))
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

func TestIPBlacklistPersist(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initIPBlacklist(dir)

	if _, err := addIPBlacklist(dir, "1.2.3.4"); err != nil {
		t.Fatal(err)
	}
	if _, err := addIPBlacklist(dir, "2001:db8::5"); err != nil {
		t.Fatal(err)
	}
	if !isIPBlacklisted("1.2.3.4") {
		t.Fatal("missing v4")
	}
	if !isIPBlacklisted("2001:db8::5") {
		t.Fatal("missing v6")
	}

	raw, err := os.ReadFile(filepath.Join(secret07Dir(dir), ipBlacklistFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "1.2.3.4") {
		t.Fatalf("file=%q", raw)
	}
	trim := strings.TrimSpace(string(raw))
	if !strings.HasPrefix(trim, "1.2.3.4") && !strings.HasPrefix(trim, "2001:db8::5") {
		t.Fatalf("effective IPs should be at file head, got %q", trim)
	}
	if !strings.Contains(string(raw), "# === 说明与案例") {
		t.Fatal("docs trailer missing")
	}
	if strings.HasPrefix(trim, "[") {
		t.Fatal("should not be json")
	}

	initIPBlacklist(dir)
	if !isIPBlacklisted("1.2.3.4") {
		t.Fatal("reload v4")
	}

	if _, err := removeIPBlacklist(dir, "1.2.3.4"); err != nil {
		t.Fatal(err)
	}
	if isIPBlacklisted("1.2.3.4") {
		t.Fatal("still blocked")
	}
	if !isIPBlacklisted("2001:db8::5") {
		t.Fatal("v6 should remain")
	}

	raw, err = os.ReadFile(filepath.Join(secret07Dir(dir), ipBlacklistFile))
	if err != nil {
		t.Fatal(err)
	}
	trim = strings.TrimSpace(string(raw))
	if !strings.HasPrefix(trim, "2001:db8::5") {
		t.Fatalf("remaining IP should stay at head, got %q", trim)
	}
	if !strings.Contains(string(raw), "# === 说明与案例") {
		t.Fatal("docs trailer should remain after effective IPs")
	}

	if _, err := removeIPBlacklist(dir, "2001:db8::5"); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(filepath.Join(secret07Dir(dir), ipBlacklistFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(raw)), "#") {
		t.Fatal("empty list should keep comment-only template")
	}
}
