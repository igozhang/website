package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// 历史上 UTF-8 中文被误按 GBK/Latin-1 写入时常见的乱码片段（测试源文件仅用 \u 转义）。
var historicalMojibakeFragments = []string{
	"\u92b3\u4e56",      // 銆乗
	"\u9366\u3115\u60c8", // 鍦ㄥ惈
	"\u9357\u66f2\u6d93", // 鍗曟洸
}

var projectTextExtensions = map[string]bool{
	".go": true, ".js": true, ".html": true, ".md": true,
	".json": true, ".yaml": true, ".yml": true, ".tpl": true,
	".txt": true, ".conf": true, ".svg": true,
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestProjectTextFiles_UTF8NoMojibake(t *testing.T) {
	root := moduleRoot(t)
	skipDir := map[string]bool{
		".git": true, "node_modules": true,
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !projectTextExtensions[ext] {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if strings.HasPrefix(rel, "..") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: read: %v", path, err)
			continue
		}
		if !utf8.Valid(raw) {
			t.Errorf("%s: not valid UTF-8", path)
			continue
		}
		s := string(raw)
		if strings.ContainsRune(s, '\uFFFD') {
			t.Errorf("%s: contains U+FFFD replacement character", path)
		}
		for _, frag := range historicalMojibakeFragments {
			if strings.Contains(s, frag) {
				t.Errorf("%s: contains historical mojibake fragment %q", path, frag)
			}
		}
	}
}

func TestMp3InitJS_UICopy(t *testing.T) {
	if !utf8.Valid(mp3InitJS) {
		t.Fatal("mp3-init.js: invalid UTF-8")
	}
	s := string(mp3InitJS)
	for _, want := range []string{
		"播放", "暂停", "one", "从此曲开始随机连播", "无匹配", "rebuildMp3SearchIndex",
		"mp3-search-hit-on", "ArrowDown",
		"在含「", "的歌曲里随机一首并持续连播",
		"无法开始播放（请再点一次播放）",
		"MP3 列表加载失败",
		`[\u3001\uFF0C,、\/\|]+`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("mp3-init.js: missing %q", want)
		}
	}
	if !utf8.Valid(modMp3HTML) {
		t.Fatal("mod-mp3.html: invalid UTF-8")
	}
	mod := string(modMp3HTML)
	for _, want := range []string{"搜歌名/歌手", "mp3Search", "mp3Loop", "mp3-main-row", "mp3-controls", "单曲循环：仅重复当前这首；换歌或点随机后自动关闭"} {
		if !strings.Contains(mod, want) {
			t.Errorf("mod-mp3.html: missing %q", want)
		}
	}
}
