package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
	"github.com/igokorea/igozhang/internal/markdown"
)

const maxDraftImagesPerSave = 32

var draftImgRE = regexp.MustCompile(`!\[([^\]]*)\]\(__draft__:([a-zA-Z0-9_-]+)\)`)

type draftImageIn struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

// materializeDraftImages replaces __draft__ placeholders with doc-pic files on disk.
func materializeDraftImages(picDir, slug, md string, drafts []draftImageIn) (string, error) {
	if strings.TrimSpace(md) == "" {
		return md, nil
	}
	byID := make(map[string]draftImageIn, len(drafts))
	for _, d := range drafts {
		id := strings.TrimSpace(d.ID)
		if id == "" {
			continue
		}
		byID[id] = d
	}
	usedIDs := make(map[string]struct{})
	var err error
	out := draftImgRE.ReplaceAllStringFunc(md, func(match string) string {
		if err != nil {
			return match
		}
		sub := draftImgRE.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		alt, id := sub[1], sub[2]
		if len(usedIDs) >= maxDraftImagesPerSave {
			err = fmt.Errorf("too many draft images (max %d)", maxDraftImagesPerSave)
			return match
		}
		d, ok := byID[id]
		if !ok {
			err = fmt.Errorf("missing draft image %q", id)
			return match
		}
		raw, ext, decErr := decodePasteImage(d.Data)
		if decErr != nil {
			err = decErr
			return match
		}
		name := fmt.Sprintf("%s_%d.%s", slug, time.Now().UnixNano(), ext)
		name = sanitizeFilename(name)
		dest := filepath.Join(picDir, name)
		if writeErr := os.WriteFile(dest, raw, 0o644); writeErr != nil {
			err = writeErr
			return match
		}
		usedIDs[id] = struct{}{}
		return "![" + alt + "](doc-pic/" + name + ")"
	})
	if err != nil {
		return "", err
	}
	return out, nil
}

func docPicRefsUsedElsewhere(assetsDir, excludeSlug string) (map[string]bool, error) {
	used := make(map[string]bool)
	mdDir := filepath.Join(assetsDir, "md")
	entries, err := os.ReadDir(mdDir)
	if err != nil {
		if os.IsNotExist(err) {
			return used, nil
		}
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		if slug == excludeSlug {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(mdDir, e.Name()))
		if err != nil {
			continue
		}
		for _, ref := range markdown.ExtractDocPicRefs(string(raw)) {
			used[ref] = true
		}
	}
	return used, nil
}

func deleteDocPicFiles(picDir string, names []string, keep map[string]bool) error {
	for _, name := range names {
		if keep[name] {
			continue
		}
		if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
			continue
		}
		p := filepath.Join(picDir, name)
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func syncArticleImages(assetsDir, slug, oldMD, newMD string) error {
	oldRefs := markdown.ExtractDocPicRefs(oldMD)
	newSet := make(map[string]bool)
	for _, r := range markdown.ExtractDocPicRefs(newMD) {
		newSet[r] = true
	}
	var removed []string
	for _, r := range oldRefs {
		if !newSet[r] {
			removed = append(removed, r)
		}
	}
	if len(removed) == 0 {
		return nil
	}
	usedElsewhere, err := docPicRefsUsedElsewhere(assetsDir, slug)
	if err != nil {
		return err
	}
	picDir := filepath.Join(assetsDir, "doc-pic")
	return deleteDocPicFiles(picDir, removed, usedElsewhere)
}

func deleteArticleWithMedia(st *docindex.Store, assetsDir, slug string) error {
	raw, err := st.ReadBody(slug)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	refs := markdown.ExtractDocPicRefs(string(raw))
	usedElsewhere, err := docPicRefsUsedElsewhere(assetsDir, slug)
	if err != nil {
		return err
	}
	if err := st.DeleteArticle(slug); err != nil {
		return err
	}
	picDir := filepath.Join(assetsDir, "doc-pic")
	return deleteDocPicFiles(picDir, refs, usedElsewhere)
}
