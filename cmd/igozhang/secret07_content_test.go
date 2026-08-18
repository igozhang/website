package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecret07TreatAsMissing(t *testing.T) {
	tpl := []byte(assetsPhrasesTemplate)
	if !secret07TreatAsMissing(tpl, tpl) {
		t.Fatal("template alone should be missing")
	}
	if !secret07TreatAsMissing([]byte("# only\n# comment\n"), tpl) {
		t.Fatal("comments only should be missing")
	}
	if secret07TreatAsMissing([]byte("Hello｜你好\n"), tpl) {
		t.Fatal("effective line should not be missing")
	}
	if secret07TreatAsMissing([]byte("Hello｜你好\n"), nil) {
		t.Fatal("effective line without template compare should not be missing")
	}
}

func TestSecret07ReadFileOrMissing(t *testing.T) {
	dir := t.TempDir()
	p := docAssetsPhrasesPath(dir)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(assetsPhrasesTemplate), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := secret07ReadFileOrMissing(p, []byte(assetsPhrasesTemplate)); ok {
		t.Fatal("template file should not be readable as external")
	}
	if err := os.WriteFile(p, []byte("Real｜真实\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := secret07ReadFileOrMissing(p, []byte(assetsPhrasesTemplate)); !ok {
		t.Fatal("file with effective line should be readable")
	}
}

func TestListSSHTargets_templateAsMissing(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	_, err := listSSHTargets(dir)
	if err != os.ErrNotExist {
		t.Fatalf("template ssh.pass err=%v want ErrNotExist", err)
	}
}

func TestBuildPhraseServePool_templateUsesEmbedded(t *testing.T) {
	dir := t.TempDir()
	initAssetsContentFiles(dir)
	withFile, _ := buildPhraseServePool(dir)
	emptyDir := t.TempDir()
	withoutFile, _ := buildPhraseServePool(emptyDir)
	if len(withFile) != len(withoutFile) {
		t.Fatalf("template phrases pool len %d != no-file %d", len(withFile), len(withoutFile))
	}
}
