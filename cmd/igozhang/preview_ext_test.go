package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestPreviewKind(t *testing.T) {
	cases := []struct {
		name     string
		wantKind string
		wantLang string
		wantOK   bool
	}{
		{"notes.txt", "code", "", true},
		{"readme.md", "md", "", true},
		{"a.markdown", "md", "", true},
		{"app.js", "code", "shell", true},
		{"cfg.yaml", "code", "yaml", true},
		{"cfg.yml", "code", "yaml", true},
		{"app.toml", "code", "shell", true},
		{"nginx.conf", "code", "shell", true},
		{"data.json", "code", "json", true},
		{"main.go", "code", "go", true},
		{"x.py", "code", "python", true},
		{"Dockerfile", "code", "shell", true},
		{"Dockerfile.dev", "code", "shell", true},
		{"Makefile", "code", "shell", true},
		{"Makefile.linux", "code", "shell", true},
		{".gitignore", "code", "shell", true},
		{"photo.png", "", "", false},
		{"archive.zip", "", "", false},
		{"binary.exe", "", "", false},
		{"file.", "", "", false},
	}
	for _, tc := range cases {
		kind, lang, ok := previewKind(tc.name)
		if ok != tc.wantOK || kind != tc.wantKind || lang != tc.wantLang {
			t.Fatalf("%s: got kind=%q lang=%q ok=%v want kind=%q lang=%q ok=%v",
				tc.name, kind, lang, ok, tc.wantKind, tc.wantLang, tc.wantOK)
		}
	}
}

func TestIsPreviewableFilename(t *testing.T) {
	if !isPreviewableFilename("fs/a.yaml") {
		t.Fatal("yaml should be previewable")
	}
	if isPreviewableFilename("fs/a.bin") {
		t.Fatal("bin should not be previewable")
	}
}

func TestPreviewExtListSyncedWithFrontend(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	jsPath := filepath.Join(filepath.Dir(thisFile), "ui", "ui-shared.js")
	raw, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`return /\^\(\?:([^)]+)\)\$/`)
	m := re.FindSubmatch(raw)
	if m == nil {
		t.Fatal("isPreviewable extension regex not found in ui-shared.js")
	}
	jsSet := map[string]bool{}
	for _, part := range strings.Split(string(m[1]), "|") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		part = strings.ReplaceAll(part, `\.`, ".")
		jsSet[part] = true
	}
	for _, ext := range previewableExtsForTest() {
		if !jsSet[ext] {
			t.Fatalf("ui-shared.js isPreviewable missing %q", ext)
		}
		sample := "file" + ext
		switch ext {
		case ".gitignore", ".dockerignore", ".editorconfig", ".npmrc", ".nvmrc",
			".eslintrc", ".prettierrc", ".babelrc":
			sample = ext
		}
		if !isPreviewableFilename(sample) {
			t.Fatalf("previewKind should accept %q", sample)
		}
		delete(jsSet, ext)
	}
	if len(jsSet) > 0 {
		extra := make([]string, 0, len(jsSet))
		for k := range jsSet {
			extra = append(extra, k)
		}
		t.Fatalf("ui-shared.js has extra extensions not in Go list: %v", extra)
	}
}
