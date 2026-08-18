package main

import "testing"

func TestIsBrowsableDirName(t *testing.T) {
	show := []string{"fs", "foo", "Ftp", "files"}
	hide := []string{"assets", "secret-tempfile", "public", "data", ""}
	for _, n := range show {
		if !isBrowsableDirName(n) {
			t.Fatalf("want show %q", n)
		}
	}
	for _, n := range hide {
		if isBrowsableDirName(n) {
			t.Fatalf("want hide %q", n)
		}
	}
}

func TestBrowsePathAllowed(t *testing.T) {
	if !browsePathAllowed("") || !browsePathAllowed("fs/foo") {
		t.Fatal("f-prefixed path segments should be allowed")
	}
	if browsePathAllowed("fs/data") {
		t.Fatal("non-f segment should be blocked")
	}
	if !browsePathAllowed("assets/mp3") {
		t.Fatal("assets/mp3 path should be allowed for player list API")
	}
	if browsePathAllowed("assets") || browsePathAllowed("assets/mp3/extra") {
		t.Fatal("assets root and nested mp3 paths should be blocked for generic browse")
	}
}

func TestBrowseRelBlockedForFiles(t *testing.T) {
	if !browseRelBlockedForFiles("assets/md/foo.md") {
		t.Fatal("assets path should block files download")
	}
	if browseRelBlockedForFiles("fs/readme.txt") {
		t.Fatal("file under fs should be allowed")
	}
}
