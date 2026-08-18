package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestExpandHostPattern_ipv4Range(t *testing.T) {
	got, err := expandHostPattern("10.1.1.2-5")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"10.1.1.2", "10.1.1.3", "10.1.1.4", "10.1.1.5"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestExpandHostPattern_hostnameDash(t *testing.T) {
	got, err := expandHostPattern("my-server-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "my-server-01" {
		t.Fatalf("got %v", got)
	}
}

func TestListSSHTargets_maxTen(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret07")
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	var lines []string
	for i := 2; i <= 12; i++ {
		lines = append(lines, "10.1.1."+strconv.Itoa(i)+" 22 root pass")
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(filepath.Join(secret, "ssh.pass"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	tgts, err := listSSHTargets(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tgts) != 10 {
		t.Fatalf("len(targets)=%d want 10", len(tgts))
	}
}

func TestListSSHTargets_ipRangeInOneLine(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret07")
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "ssh.pass"), []byte("10.1.1.2-5 22 root secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tgts, err := listSSHTargets(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tgts) != 4 {
		t.Fatalf("len=%d want 4", len(tgts))
	}
	if tgts[0].Host != "10.1.1.2" || tgts[3].Host != "10.1.1.5" {
		t.Fatalf("hosts: %s %s", tgts[0].Host, tgts[3].Host)
	}
}
