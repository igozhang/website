package main

import (
	"strings"
	"testing"
)

func TestApplySitePage_sharedFooterAndHead(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	pub := ""
	html := string(applySitePage(`{{PAGE_HEAD}}<body>{{SITE_HEAD}}{{SITE_FOOTER}}</body>`, pub, sitePageLayout{}, nil, ""))
	if !strings.Contains(html, `name="fs-public-base"`) {
		t.Fatal("missing fs-public-base meta")
	}
	if !strings.Contains(html, "window.__fsAbsRoot") {
		t.Fatal("missing site-pub-base.js in PAGE_HEAD")
	}
	if !strings.Contains(html, `id="footIgoHit"`) {
		t.Fatal("missing footer igo hit")
	}
	if !strings.Contains(html, "class=\"site-head\"") {
		t.Fatal("missing site head")
	}
}

func TestApplySitePage_mp3Slots(t *testing.T) {
	html := string(applySitePage(`{{MP3_SECTION}}{{MP3_SCRIPTS}}`, "", sitePageLayout{MP3: true, MP3CSS: true}, nil, ""))
	if !strings.Contains(html, `id="mp3Audio"`) {
		t.Fatal("missing mp3 section when MP3 enabled")
	}
}

func TestApplySitePage_replacesAllPlaceholders(t *testing.T) {
	html := string(applySitePage(
		`<title>{{ARTICLE_TITLE}}</title><h1>{{ARTICLE_TITLE}}</h1>`,
		"",
		sitePageLayout{},
		map[string]string{"{{ARTICLE_TITLE}}": "redis排错"},
		"",
	))
	if strings.Contains(html, "{{ARTICLE_TITLE}}") {
		t.Fatalf("unreplaced placeholder: %s", html)
	}
	n := strings.Count(html, "redis排错")
	if n != 2 {
		t.Fatalf("want 2 replacements, got %d in %s", n, html)
	}
}
