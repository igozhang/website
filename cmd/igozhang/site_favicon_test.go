package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSecret07Favicon_Priority(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, secret07RelDir)
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(secret, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("fav.jpg")
	write("favicon.jpeg")
	write("favicon.png")

	c := resolveSecret07Favicon(dir)
	if c == nil {
		t.Fatal("expected favicon")
	}
	if c.Base != "favicon" || c.Ext != "png" {
		t.Fatalf("got %s.%s want favicon.png", c.Base, c.Ext)
	}
}

func TestResolveSecret07Favicon_FaviconOverFav(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, secret07RelDir)
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "fav.png"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "favicon.jpg"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := resolveSecret07Favicon(dir)
	if c == nil || c.Base != "favicon" {
		t.Fatalf("got %+v", c)
	}
}

func TestResolveSecret07Favicon_SkipsOversized(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, secret07RelDir)
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	big := make([]byte, siteFaviconMaxBytes+1)
	if err := os.WriteFile(filepath.Join(secret, "favicon.png"), big, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "fav.jpg"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := resolveSecret07Favicon(dir)
	if c == nil || c.Base != "fav" || c.Ext != "jpg" {
		t.Fatalf("got %+v, want fav.jpg", c)
	}
}

func TestSiteFaviconIconLink_Default(t *testing.T) {
	siteFaviconDataRoot = t.TempDir()
	href, mime := siteFaviconIconLink("")
	if mime != "image/svg+xml" || !strings.HasPrefix(href, "data:image/svg+xml") {
		t.Fatalf("href=%q mime=%q", href, mime)
	}
}
