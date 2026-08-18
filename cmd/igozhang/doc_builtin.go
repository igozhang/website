package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

//go:embed builtin/docker_ts.md
var builtinDockerTSMD []byte

const builtinDemoSlug = "docker_ts"

func mdDirHasArticles(assets string) bool {
	entries, err := os.ReadDir(filepath.Join(assets, "md"))
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			return true
		}
	}
	return false
}

// installBuiltinDemoArticle seeds docker_ts when assets/md is empty (display sample).
func installBuiltinDemoArticle(st *docindex.Store, assets string) error {
	if mdDirHasArticles(assets) {
		return nil
	}
	all, err := st.ListAll()
	if err != nil {
		return err
	}
	if len(all) > 0 {
		return nil
	}
	body := append([]byte(nil), builtinDockerTSMD...)
	m := docindex.Meta{
		Slug:        builtinDemoSlug,
		Title:       "docker_ts,docker排错",
		Categories:  "devops, k8s",
		Tags:        "docker, Troubleshooting",
		Author:      "igozhang",
		PublishedAt: time.Date(2022, 6, 1, 12, 0, 0, 0, time.Local),
	}
	if err := st.ImportArticle(m, body); err != nil {
		return err
	}
	log.Printf("doc: installed builtin demo article %s (md dir was empty)", builtinDemoSlug)
	return nil
}
