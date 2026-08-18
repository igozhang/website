package main

import (
	"strings"
	"testing"
)

func TestParseOpsConfig_DocNavBall(t *testing.T) {
	s, err := parseOpsConfig("[doc]\nnav_ball = off\n")
	if err != nil {
		t.Fatal(err)
	}
	if s.docNavBallEnabled {
		t.Fatal("expected nav_ball off")
	}
	s2, err := parseOpsConfig(string(opsConfigExample))
	if err != nil {
		t.Fatal(err)
	}
	if !s2.docNavBallEnabled {
		t.Fatal("example nav_ball should default on")
	}
}

func TestDocNavBallScripts_enabled(t *testing.T) {
	applyOpsSnapshot(defaultOpsSnapshot())
	got := docNavBallScripts("/pub")
	if !strings.Contains(got, "window.__fsDocNavBallOps") {
		t.Fatal("missing ops inject")
	}
	if !strings.Contains(got, `"enabled":true`) {
		t.Fatalf("expected enabled: %s", got)
	}
	if !strings.Contains(got, "window.__fsDocNavBallInit") {
		t.Fatal("missing init boot")
	}
	if !strings.Contains(string(docNavBallJS), "__fsDocNavBallInit") {
		t.Fatal("missing embedded js export")
	}
}

func TestDocNavBallScripts_disabled(t *testing.T) {
	s := defaultOpsSnapshot()
	s.docNavBallEnabled = false
	applyOpsSnapshot(s)
	defer applyOpsSnapshot(defaultOpsSnapshot())
	if docNavBallScripts("") != "" {
		t.Fatal("disabled nav_ball should emit no scripts")
	}
	if docNavBallCSSIfEnabled() != "" {
		t.Fatal("disabled nav_ball should emit no css")
	}
}

func TestDocHomeSection_includesNavBall(t *testing.T) {
	applyOpsSnapshot(defaultOpsSnapshot())
	dir := t.TempDir()
	initSecret07Files(dir)
	html := string(renderArticleHTMLForTest("", dir))
	if !strings.Contains(html, "__fsDocNavBallInit") {
		t.Fatal("article page missing nav ball scripts")
	}
}

func TestApplySitePage_includesNavBall(t *testing.T) {
	applyOpsSnapshot(defaultOpsSnapshot())
	html := string(applySitePage(`{{SITE_SUBPAGE_TAIL}}`, "", sitePageLayout{}, nil, ""))
	if !strings.Contains(html, "__fsDocNavBallInit") {
		t.Fatal("subpage missing nav ball scripts")
	}
}

func TestHomePage_includesNavBallCSS(t *testing.T) {
	applyOpsSnapshot(defaultOpsSnapshot())
	if !strings.Contains(siteHomePageCSS(), ".doc-nav-ball{") {
		t.Fatal("home page CSS missing nav ball styles")
	}
}
