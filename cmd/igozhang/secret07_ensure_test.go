package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitSecret07Files_createsAllTemplates(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	secret := filepath.Join(dir, secret07RelDir)
	names := []string{
		opsConfigName,
		sshPassFileName,
		commonCommandsFileName,
		secret07PicReadmeName,
		linkConfFileName,
		ipBlacklistFile,
		ipWhitelistFile,
	}
	for _, name := range names {
		p := filepath.Join(secret, name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(secret, commonCommandsFileName))
	if err != nil {
		t.Fatal(err)
	}
	if cmds := parseCommonCommandsLines(raw); len(cmds) != 0 {
		t.Fatalf("common-commands default should have no active lines, got %v", cmds)
	}
	if !strings.Contains(string(raw), "# 示例：uname -a") {
		t.Fatal("common-commands missing commented example")
	}
	picReadme, err := os.ReadFile(filepath.Join(secret, secret07PicReadmeName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(picReadme), "Favicon") || !strings.Contains(string(picReadme), "Banner") || !strings.Contains(string(picReadme), "Texas") {
		t.Fatal("pic.readme.md should document favicon, banner and texas card back")
	}
	conf, err := os.ReadFile(filepath.Join(secret, opsConfigName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(conf), "[site]") || !strings.Contains(string(conf), "word3k_pool_size") {
		t.Fatal("igo.conf should be full embedded config")
	}
}

func TestInitSecret07Files_idempotent(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	p := filepath.Join(dir, secret07RelDir, commonCommandsFileName)
	if err := os.WriteFile(p, []byte("echo live\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initSecret07Files(dir)
	raw, _ := os.ReadFile(p)
	if string(raw) != "echo live\n" {
		t.Fatalf("should not overwrite existing file, got %q", raw)
	}
}
