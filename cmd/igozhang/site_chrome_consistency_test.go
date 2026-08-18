package main

import (
	"html/template"
	"strings"
	"testing"
	"time"
)

// chromeMarkers 为所有带站点 chrome 的页面应具备的公共片段。
func chromeMarkers() []string {
	return []string{
		`name="fs-public-base"`,
		"window.__fsAbsRoot",
		"window.__fsFixFragmentHrefs",
		`id="footIgoHit"`,
		`class="site-head"`,
		`class="site-foot"`,
		"window.__fsSiteFootInit",
	}
}

func assertChrome(t *testing.T, name, html string, wantMP3 bool) {
	t.Helper()
	for _, m := range chromeMarkers() {
		if !strings.Contains(html, m) {
			t.Fatalf("%s: missing chrome marker %q", name, m)
		}
	}
	if wantMP3 {
		if !strings.Contains(html, `id="mp3Audio"`) {
			t.Fatalf("%s: missing mp3 section", name)
		}
		if !strings.Contains(html, "window.__fsMp3Init") {
			t.Fatalf("%s: missing mp3 init", name)
		}
	} else {
		if strings.Contains(html, `id="mp3Audio"`) {
			t.Fatalf("%s: unexpected mp3 section", name)
		}
	}
}

func renderHomeHTMLForTest(t *testing.T, pub string) string {
	t.Helper()
	tmpl, err := template.New("home").Parse(homeIndexHTMLTemplate())
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	err = tmpl.Execute(&b, map[string]any{
		"PageHeadHTML":       template.HTML(pageHeadTags(pub)),
		"HomeCSSLinkHTML":    template.HTML(docHomeExternalCSSLink(pub)),
		"FsPubAttr":          fsPubAttr(pub),
		"SiteTitle":          siteDisplayName(),
		"SiteFootHTML":       template.HTML(siteFootBlock(pub, opsFooterGomokuHref(pub), opsFooterXiangqiHref(pub), opsFooterTexasHref(pub), pubPathZhijian(pub), "")),
		"SiteHeadHTML":       template.HTML(siteHeadBlock(baseHref(pub), pub, siteHeadPhrasesJSON())),
		"DocHomeHTML":        buildDocHomeSection(pub, "", "", 1),
		"Mp3SectionHTML":     template.HTML(siteMp3Section()),
		"Mp3ScriptsHTML":     template.HTML(siteMp3InitScripts(pub)),
		"NavBallScriptsHTML": template.HTML(docNavBallScripts(pub)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func renderArticleHTMLForTest(pub, dataRoot string) []byte {
	return applySitePage(string(articlePageTpl), pub, sitePageLayout{
		HeadCSS:            siteSubpageHeadCSS(),
		MP3:                true,
		MP3CSS:             false,
		MP3ScriptsOverride: siteMp3ArticleInitScripts(pub),
	}, map[string]string{
		"{{DOC_NAV}}":            buildDocNavHTML(nil, 10, "", true),
		"{{DOC_LINK_ASIDE}}":     buildDocLinkAsideHTML(pub, dataRoot),
		"{{DOC_MODULE_JS}}":      docModuleScript(pub, true),
		"{{DOC_ARTICLE_ASSETS}}": docArticleExternalCSSLink(pub) + docArticleExternalMp3CSSLink(pub) + docArticleBodyScripts(pub, []byte("<p>hi</p>")),
		"{{ARTICLE_TITLE}}":      "test",
		"{{ARTICLE_META}}":       "meta",
		"{{ARTICLE_AUTHOR}}":     "作者 · igozhang",
		"{{ARTICLE_BODY}}":       `<div class="md doc-md"><p>hi</p></div>`,
		"{{ARTICLE_CSS}}":        "",
	}, "")
}

func TestChromeConsistency_allPages(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	pub := ""

	assertChrome(t, "home", renderHomeHTMLForTest(t, pub), true)
	homeHTML := renderHomeHTMLForTest(t, pub)
	if !strings.Contains(homeHTML, `data-fs-pub="`) {
		t.Fatal("home: main missing data-fs-pub")
	}
	if !strings.Contains(homeHTML, "/assets/doc/home.css") {
		t.Fatal("home: must link external home.css")
	}
	if !strings.Contains(homeHTML, "/assets/doc/mp3-init.js") {
		t.Fatal("home: must link external mp3-init.js")
	}
	if strings.Contains(homeHTML, "__fsMp3LazyList") {
		t.Fatal("home: must not enable lazy mp3 list yet (plan B)")
	}
	assertChrome(t, "fs", string(renderFsHTML(pub, nil)), true)
	assertChrome(t, "ops", string(renderOpsHTML(pub, dir, nil)), false)
	assertChrome(t, "igoadmin", string(renderIgoAdminHTML(pub, dir, nil)), false)
	assertChrome(t, "igokey", string(renderIgoKeyHTML(pub, "abcdEFGH", time.Now().Add(time.Hour), nil)), false)
	assertChrome(t, "gomoku", string(renderGomokuHTML(pub, nil)), true)
	assertChrome(t, "xiangqi", string(renderXiangqiHTML(pub, nil)), true)
	assertChrome(t, "texas", string(renderTexasHTML(pub, nil)), true)
	assertChrome(t, "zhijian", string(renderZhijianHTML(pub, nil)), true)
	assertChrome(t, "mermaid", string(renderMermaidPageHTML(pub, nil)), false)
	articleHTML := string(renderArticleHTMLForTest(pub, dir))
	assertChrome(t, "article", articleHTML, true)
	if !strings.Contains(articleHTML, "doc-link-aside") || !strings.Contains(articleHTML, "doc-link-capsule") {
		t.Fatal("article page must include link.conf capsules")
	}
	if !strings.Contains(articleHTML, "/assets/doc/article.css") {
		t.Fatal("article page must link external doc CSS")
	}
	if !strings.Contains(articleHTML, "/assets/doc/mp3.css") {
		t.Fatal("article page must link external mp3 CSS")
	}
	if !strings.Contains(articleHTML, "/assets/doc/code-chrome.js") {
		t.Fatal("article page must link external code highlighter JS")
	}
	if strings.Contains(articleHTML, "mermaid.min.js") {
		t.Fatal("plain article must not load mermaid unless body has language-mermaid")
	}
	if !strings.Contains(articleHTML, "/assets/doc/mp3-init.js") {
		t.Fatal("article page must link external mp3-init JS")
	}
	if !strings.Contains(articleHTML, "__fsMp3LazyList") {
		t.Fatal("article page must enable lazy mp3 list loading")
	}
	if !mp3DockFollowsMain(articleHTML) {
		t.Fatal("article MP3 dock must sit between </main> and footer")
	}
	articleIdx := strings.Index(articleHTML, `class="doc-article-block`)
	asideIdx := strings.Index(articleHTML, `class="doc-link-aside`)
	if articleIdx < 0 || asideIdx < 0 || asideIdx < articleIdx {
		t.Fatal("article page: link aside must follow article body in DOM (mobile order)")
	}
}
