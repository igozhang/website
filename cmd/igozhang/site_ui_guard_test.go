package main

import (
	"os"
	"strings"
	"testing"
)

func TestSiteUIGuardCSS_InAllPageStyles(t *testing.T) {
	if strings.Contains(siteUIGuardCSS, "body{") && strings.Contains(siteUIGuardCSS, "user-select:none") {
		t.Fatal("site guard must not disable text selection on body")
	}
	if !strings.Contains(siteUIGuardCSS, "caret-color:transparent") {
		t.Fatal("site guard CSS must hide insertion caret on static UI")
	}
	if !strings.Contains(siteUIGuardCSS, "textarea,.stash-textarea") {
		t.Fatal("site guard CSS must allow caret in real text fields")
	}
	raw, err := os.ReadFile("site_home.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "siteUIGuardCSS") {
		t.Fatal("home page must include siteUIGuardCSS")
	}
	sub, err := os.ReadFile("site_subpage.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sub), "siteUIGuardCSS") {
		t.Fatal("subpages must include siteUIGuardCSS")
	}
	head, err := os.ReadFile("site_head.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(head), "siteUIGuardScript()") {
		t.Fatal("site head must inject siteUIGuardScript early")
	}
	if !strings.Contains(string(head), "siteUIGuardCSS") {
		t.Fatal("site head must include siteUIGuardCSS")
	}
	guardJS, err := os.ReadFile("site-ui-guard.js")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(guardJS), "preventDefault") {
		t.Fatal("site-ui-guard.js must not preventDefault on click (blocks text selection)")
	}
	if !strings.Contains(string(guardJS), "focusin") {
		t.Fatal("site-ui-guard.js must blur stray focus via focusin")
	}
	if !strings.Contains(string(guardJS), "__fsSiteUIGuardInit") {
		t.Fatal("site-ui-guard.js must be idempotent")
	}
}

func TestZhijianHTML_NoFuseSpark(t *testing.T) {
	raw, err := os.ReadFile("zhijian.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if strings.Contains(s, "zjFuseSpark") || strings.Contains(s, "zj-fuse-spark") {
		t.Fatal("zhijian must not include fuse spark (removed; was mistaken for text caret)")
	}
	if strings.Contains(s, `data-ch="`) || strings.Contains(s, `aria-live="polite">—</span>`) {
		t.Fatal("current letter must not use data-ch or aria-live on letter box (causes faux caret)")
	}
	if !strings.Contains(s, "zj-letter-svg") || !strings.Contains(s, `id="zjLetterBlock" inert`) {
		t.Fatal("zhijian must render letter via SVG and inert letter block")
	}
	if strings.Contains(s, "zjKeyCapture") {
		t.Fatal("zhijian must not use hidden key capture input (causes page-wide caret)")
	}
	if strings.Contains(s, "zjSuppressTextCaret") {
		t.Fatal("zhijian must rely on site-ui-guard.js not duplicate caret handlers")
	}
	if !strings.Contains(s, "transform-origin:left") || !strings.Contains(s, "scaleX") {
		t.Fatal("fuse ember must use scaleX with left origin")
	}
	if !strings.Contains(s, "zj-letter-block") || !strings.Contains(s, "align-items:flex-end") {
		t.Fatal("zhijian must use zj-letter-block with bottom-aligned letter row")
	}
	if !strings.Contains(s, `id="zjFuseOverlay"`) || !strings.Contains(s, "zj-fuse-overlay") {
		t.Fatal("zhijian fuse countdown must use overlay inside current letter box")
	}
	if strings.Contains(s, `id="zjFuse"`) {
		t.Fatal("zhijian must not use separate bottom fuse bar")
	}
	if strings.Contains(s, `class="phrase-next primary" id="zjStart"`) {
		t.Fatal("zhijian start button must use zhijian-btn not phrase-next")
	}
}
