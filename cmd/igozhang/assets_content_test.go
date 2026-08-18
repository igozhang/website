package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitAssetsContentFiles_createsTemplates(t *testing.T) {
	dir := t.TempDir()
	initAssetsContentFiles(dir)
	eng := docAssetsEngDir(dir)
	for _, name := range []string{phrasesFileName, word3kFileName} {
		if _, err := os.Stat(filepath.Join(eng, name)); err != nil {
			t.Fatalf("missing assets/eng/%s: %v", name, err)
		}
	}
	if st, err := os.Stat(docAssetsMp3Dir(dir)); err != nil || !st.IsDir() {
		t.Fatalf("mp3 dir: %v", err)
	}
}
