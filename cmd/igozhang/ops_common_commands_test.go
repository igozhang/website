package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCommonCommandsLines(t *testing.T) {
	raw := []byte("# comment\n\nuname -a\n  df -h  \n# tail\nfree -m\n")
	got := parseCommonCommandsLines(raw)
	want := []string{"uname -a", "df -h", "free -m"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] got %q want %q", i, got[i], want[i])
		}
	}
}

func TestParseCommonCommandsLinesMax20(t *testing.T) {
	lines := make([]string, 25)
	for i := range lines {
		lines[i] = "echo " + string(rune('a'+i))
	}
	got := parseCommonCommandsLines([]byte(strings.Join(lines, "\n")))
	if len(got) != commonCommandsMaxLines {
		t.Fatalf("len=%d want %d", len(got), commonCommandsMaxLines)
	}
}

func TestReadCommonCommandsExportBytes_skipsDefault(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	if raw := readCommonCommandsExportBytes(dir); len(raw) != 0 {
		t.Fatalf("default common-commands should not export, got %q", raw)
	}
}

func TestReadCommonCommandsExportBytes_custom(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	custom := "# notes\nkubectl get pods\nuname -a\n"
	if err := os.WriteFile(commonCommandsPath(dir), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	raw := readCommonCommandsExportBytes(dir)
	if !strings.Contains(string(raw), "kubectl get pods") {
		t.Fatalf("custom export: %q", raw)
	}
	got := listCommonCommands(dir)
	if len(got) != 2 || got[0] != "kubectl get pods" {
		t.Fatalf("list after custom: %v", got)
	}
}

func TestEnsureAndListCommonCommands(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	got := listCommonCommands(dir)
	if len(got) != 0 {
		t.Fatalf("default list=%v want no active commands", got)
	}
	p := commonCommandsPath(dir)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("stat: %v", err)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "# 示例：uname -a") {
		t.Fatal("missing commented example uname -a")
	}
	_ = os.WriteFile(p, []byte("echo one\necho two\n"), 0o644)
	got2 := listCommonCommands(dir)
	if len(got2) != 2 || got2[0] != "echo one" {
		t.Fatalf("after edit got %v", got2)
	}
	secret := filepath.Join(dir, secret07RelDir)
	if _, err := os.Stat(secret); err != nil {
		t.Fatalf("secret07 dir: %v", err)
	}
}
