// Package markdown renders article Markdown to safe HTML for public pages.
package markdown

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var docPicRE = regexp.MustCompile(`!\[([^\]]*)\]\(doc-pic/([^)]+)\)`)

// RenderOpts configures image URL rewriting for doc-pic references.
type RenderOpts struct {
	// PicBaseURL is the public prefix for images, e.g. "/assets/doc-pic/" or "/fs/assets/doc-pic/".
	PicBaseURL string
}

// StripLeadingH1 removes the first ATX h1 line (# title). Article pages already show title in the page header.
func StripLeadingH1(src []byte) []byte {
	lines := strings.Split(string(src), "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i >= len(lines) {
		return src
	}
	line := strings.TrimSpace(lines[i])
	if !strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
		return src
	}
	rest := append(lines[:i], lines[i+1:]...)
	j := i
	for j < len(rest) && strings.TrimSpace(rest[j]) == "" {
		j++
	}
	return []byte(strings.Join(append(rest[:i], rest[j:]...), "\n"))
}

// ExtractLeadingH1Title returns the text of the first ATX h1 (# title), or "" if none.
func ExtractLeadingH1Title(src []byte) string {
	lines := strings.Split(string(src), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			return ""
		}
		return strings.TrimSpace(strings.TrimPrefix(line, "# "))
	}
	return ""
}

// TitleFromSlug derives a display title from slug (underscores to spaces).
func TitleFromSlug(slug string) string {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return ""
	}
	return strings.ReplaceAll(slug, "_", " ")
}

// ResolveArticleTitle picks title from explicit value, markdown h1, or slug.
func ResolveArticleTitle(explicit, slug string, md []byte) string {
	if t := strings.TrimSpace(explicit); t != "" {
		return t
	}
	if t := ExtractLeadingH1Title(md); t != "" {
		return t
	}
	if t := TitleFromSlug(slug); t != "" {
		return t
	}
	return strings.TrimSpace(slug)
}

// newGoldmark returns the site Markdown engine (GFM + Unicode heading ids).
func newGoldmark() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe(), html.WithHardWraps()),
	)
}

func convertMarkdown(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	ctx := parser.NewContext(parser.WithIDs(newUnicodeHeadingIDs()))
	if err := newGoldmark().Convert(src, &buf, parser.WithContext(ctx)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Render converts Markdown to an HTML fragment (no outer document).
func Render(src []byte, opts RenderOpts) ([]byte, error) {
	base := strings.TrimSuffix(strings.TrimSpace(opts.PicBaseURL), "/")
	if base != "" {
		base += "/"
	}
	rewritten := docPicRE.ReplaceAllStringFunc(string(src), func(m string) string {
		sub := docPicRE.FindStringSubmatch(m)
		if len(sub) != 3 {
			return m
		}
		alt, name := sub[1], sub[2]
		return "![" + alt + "](" + base + name + ")"
	})
	return convertMarkdown([]byte(rewritten))
}
