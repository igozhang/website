package main

import (
	"os"
	"strings"
	"testing"
)

func mp3DockFollowsMain(html string) bool {
	html = strings.ReplaceAll(html, "\r\n", "\n")
	dock := strings.Index(html, `<div class="shell mp3-above-foot">`)
	mainEnd := strings.LastIndex(html, "</main>")
	return dock > 0 && mainEnd > 0 && mainEnd < dock
}

func TestHomeHTML_NoInlineMp3Script(t *testing.T) {
	s := homeIndexHTMLTemplate()
	if strings.Contains(s, "  var mp3List=[];") {
		t.Fatal("home page still embeds inline MP3 script; use {{.Mp3ScriptsHTML}} and mp3-init.js only")
	}
	if strings.Contains(s, "<style>") {
		t.Fatal("home page must not inline CSS; use {{.HomeCSSLinkHTML}}")
	}
	if !strings.Contains(s, "{{.HomeCSSLinkHTML}}") {
		t.Fatal("home page must reference HomeCSSLinkHTML")
	}
	if !strings.Contains(s, "{{.Mp3SectionHTML}}") || !strings.Contains(s, "{{.Mp3ScriptsHTML}}") {
		t.Fatal("home page must reference Mp3SectionHTML and Mp3ScriptsHTML templates")
	}
	if !strings.Contains(s, "{{.DocHomeHTML}}") {
		t.Fatal("home page must reference DocHomeHTML before MP3")
	}
	if strings.Index(s, "{{.DocHomeHTML}}") > strings.Index(s, "{{.Mp3SectionHTML}}") {
		t.Fatal("DocHomeHTML must appear before Mp3SectionHTML (MP3 nearest footer)")
	}
	if !mp3DockFollowsMain(s) {
		t.Fatal("home MP3 dock must sit between </main> and footer")
	}
	if strings.Contains(s, "{{.Mp3ScriptsHTML}}\n(function(){") {
		t.Fatal("Mp3ScriptsHTML must not be inside another script block")
	}
	if strings.Contains(s, "PhrasesJSON") {
		t.Fatal("home template must not re-inject PhrasesJSON; siteHeadBlock already inits phrases")
	}
}

func TestHomeHTML_VisibilityHubSelfClosing(t *testing.T) {
	s := homeIndexHTMLTemplate()
	if !strings.Contains(s, siteVisibilityHubJS) {
		t.Fatal("home page must include visibility hub script")
	}
}

func TestHomeHTML_ExternalAssets(t *testing.T) {
	html := renderHomeHTMLForTest(t, "")
	if !strings.Contains(html, "/assets/doc/home.css?v=") {
		t.Fatal("home HTML must link home.css with version")
	}
	if !strings.Contains(html, "/assets/doc/mp3-init.js?v=") {
		t.Fatal("home HTML must load external mp3-init.js with version")
	}
	if strings.Contains(html, string(mp3InitJS)) {
		t.Fatal("home HTML must not inline mp3-init.js body")
	}
	if strings.Count(html, `"en":`) > 250 {
		// phrase bank ~200 pairs once; double inject would be ~400
		t.Fatalf("phrase JSON appears duplicated in home HTML (en count=%d)", strings.Count(html, `"en":`))
	}
	if !strings.Contains(html, "window.__fsPhraseInit&&window.__fsPhraseInit(") {
		t.Fatal("home HTML must init phrases via siteHeadBlock")
	}
	// Plan A: shell without inlined CSS/mp3-init should stay well under pre-opt ~185KB.
	if n := len(html); n > 120_000 {
		t.Fatalf("home HTML too large after Plan A externalization: %d bytes", n)
	}
}

func TestFsPage_HasFileModulesAndMp3(t *testing.T) {
	out := string(renderFsHTML("", nil))
	if !strings.Contains(out, `class="shell mp3-above-foot"`) {
		t.Fatal("fs page MP3 must be in mp3-above-foot above footer")
	}
	if !mp3DockFollowsMain(out) {
		t.Fatal("fs MP3 dock must sit between </main> and footer")
	}
	for _, needle := range []string{"文件中转", "文本暂存", "目录浏览与下载", `id="mp3Audio"`, "window.__fsSharedInit", "window.__fsTransferInit"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("fs page missing %q", needle)
		}
	}
}

func TestMp3InitJS_HasPageHandoff(t *testing.T) {
	s := string(mp3InitJS)
	for _, want := range []string{
		"fsMp3Handoff",
		"writeMp3Handoff",
		"restoreMp3Handoff",
		"flushMp3Handoff",
		"pickMp3HandoffCandidate",
		"bindHandoffGestureResume",
		"abortHandoffAutoResume",
		"clearHandoffResumeAttempt",
		"MP3_HANDOFF_RESUME_MAX_MS",
		"__fsMp3FlushHandoff",
		"跨页续播：轻触页面继续",
		"urlMp3Gate",
		"ensureMp3Gate",
		`addEventListener("pagehide"`,
		`addEventListener("beforeunload"`,
		`addEventListener("touchend"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("mp3-init.js: missing handoff marker %q", want)
		}
	}
}

func TestMp3InitJS_IsEmbeddedSource(t *testing.T) {
	if len(mp3InitJS) < 1000 {
		t.Fatal("mp3InitJS embed is empty or too small")
	}
	if !strings.Contains(string(mp3InitJS), "window.__fsMp3Init") {
		t.Fatal("mp3-init.js must export __fsMp3Init")
	}
	if !strings.Contains(string(mp3InitJS), "advanceToTrack") {
		t.Fatal("mp3-init.js missing advanceToTrack (sync check)")
	}
	if !strings.Contains(string(mp3InitJS), "trackHasEmbeddedCover(primaryFn)") {
		t.Fatal("mp3-init.js: mergeSessionArtwork must only prepend prev cover when primary has embedded cover")
	}
}

func TestMp3InitJS_HasLazyListLoad(t *testing.T) {
	s := string(mp3InitJS)
	for _, want := range []string{
		"__fsMp3LazyList",
		"bootMp3List",
		"ensureMp3ListLoaded",
		"withMp3ListReady",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("mp3-init.js: missing lazy-load marker %q", want)
		}
	}
}

func TestSubpages_IncludeMp3Placeholders(t *testing.T) {
	for _, p := range []string{"article.html", "gomoku.html", "zhijian.html", "xiangqi.html", "texas.html"} {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s := string(raw)
		if !strings.Contains(s, "{{MP3_SECTION}}") || !strings.Contains(s, "{{MP3_SCRIPTS}}") {
			t.Fatalf("%s must include MP3 placeholders", p)
		}
	}
}
