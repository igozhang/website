package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSecret07Banner(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret07")
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, n int) {
		b := make([]byte, n)
		for i := range b {
			b[i] = 'x'
		}
		if err := os.WriteFile(filepath.Join(secret, name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("banner.jpeg", 64)
	write("banner.png", 32)

	c := resolveSecret07Banner(dir)
	if c == nil {
		t.Fatal("expected banner")
	}
	if c.Ext != "png" {
		t.Fatalf("ext=%q want png", c.Ext)
	}
}

func TestResolveSecret07Banner_Oversized(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret07")
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	big := make([]byte, siteBannerMaxBytes+1)
	if err := os.WriteFile(filepath.Join(secret, "banner.png"), big, 0o644); err != nil {
		t.Fatal(err)
	}
	if c := resolveSecret07Banner(dir); c != nil {
		t.Fatal("oversized banner must be ignored")
	}
}

func TestSplitBannerName(t *testing.T) {
	if _, ok := splitBannerName("favicon.png"); ok {
		t.Fatal("favicon must not match banner")
	}
	if ext, ok := splitBannerName("banner.JPG"); !ok || ext != "jpg" {
		t.Fatalf("banner.JPG ok=%v ext=%q", ok, ext)
	}
}
