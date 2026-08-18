package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
	"github.com/igokorea/igozhang/internal/markdown"
)

func tinyPNGDataURL() string {
	// 1x1 PNG
	raw := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
}

func TestMaterializeDraftImages(t *testing.T) {
	dir := t.TempDir()
	picDir := filepath.Join(dir, "doc-pic")
	if err := os.MkdirAll(picDir, 0o755); err != nil {
		t.Fatal(err)
	}
	id := "draft1"
	md := "hello ![](__draft__:" + id + ") world"
	out, err := materializeDraftImages(picDir, "demo", md, []draftImageIn{{ID: id, Data: tinyPNGDataURL()}})
	if err != nil {
		t.Fatal(err)
	}
	refs := markdown.ExtractDocPicRefs(out)
	if len(refs) != 1 || refs[0] == "" {
		t.Fatalf("out=%q refs=%v", out, refs)
	}
	if _, err := os.Stat(filepath.Join(picDir, refs[0])); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteArticleWithMedia_removesUnsharedPic(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	mdDir := filepath.Join(assets, "md")
	picDir := filepath.Join(assets, "doc-pic")
	for _, p := range []string{mdDir, picDir} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	picName := "only_one.png"
	if err := os.WriteFile(filepath.Join(picDir, picName), []byte{0x89, 0x50}, 0o644); err != nil {
		t.Fatal(err)
	}
	body := []byte("# t\n\n![](doc-pic/" + picName + ")\n")
	if err := st.WriteArticle(docindex.Meta{
		Slug: "only_one", Title: "t", Author: "igozhang", PublishedAt: time.Now(),
	}, body); err != nil {
		t.Fatal(err)
	}
	if err := deleteArticleWithMedia(st, assets, "only_one"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(picDir, picName)); !os.IsNotExist(err) {
		t.Fatalf("pic should be removed, err=%v", err)
	}
}
