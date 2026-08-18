package main

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed site-pub-base.js
var sitePubBaseJS []byte

// sitePageLayout 描述子页如何挂载站点 chrome（head / foot / mp3）。
type sitePageLayout struct {
	HeadCSS string // 空则按 MP3CSS 选 siteSubpageHeadCSS 或 siteSubpageWithMp3CSS
	MP3     bool
	MP3CSS  bool
	// MP3ScriptsOverride 非空时替代默认 siteMp3InitScripts（如文章页外置 + 懒列表）。
	MP3ScriptsOverride string
}

func siteFootHrefs(pub string) (gomoku, xiangqi, texas, zhijian string) {
	return opsFooterGomokuHref(pub), opsFooterXiangqiHref(pub), opsFooterTexasHref(pub), pubPathZhijian(pub)
}

func siteChromeHeadCSS(layout sitePageLayout) string {
	if layout.HeadCSS != "" {
		return layout.HeadCSS
	}
	if layout.MP3CSS {
		return siteSubpageWithMp3CSS()
	}
	return siteSubpageHeadCSS()
}

// sitePublicBaseBootScript 在 <head> 注入全站路径解析（须早于 footer/mp3/fs 脚本）。
func sitePublicBaseBootScript(pub string) string {
	baseJS, err := json.Marshal(strings.TrimSuffix(strings.TrimSpace(pub), "/"))
	if err != nil {
		baseJS = []byte(`""`)
	}
	return `<script>` + string(sitePubBaseJS) + `</script>
<script>window.__fsPublicBase=` + string(baseJS) + `;</script>`
}

func applySitePage(tpl string, pub string, layout sitePageLayout, extra map[string]string, visitorLabel string) []byte {
	home := baseHref(pub)
	ph := siteHeadPhrasesJSON()
	g, x, t, z := siteFootHrefs(pub)
	repl := map[string]string{
		"{{PAGE_HEAD}}":         pageHeadTags(pub),
		"{{SITE_HEAD_CSS}}":     siteChromeHeadCSS(layout),
		"{{SITE_HEAD}}":         siteHeadBlock(home, pub, ph),
		"{{FS_PUB_ATTR}}":       fsPubAttr(pub),
		"{{SITE_SUBPAGE_TAIL}}": siteSubpageTail(pub, ""),
		"{{SITE_FOOTER}}":       siteFootBlock(pub, g, x, t, z, visitorLabel),
	}
	if layout.MP3 {
		repl["{{MP3_SECTION}}"] = siteMp3Section()
		if layout.MP3ScriptsOverride != "" {
			repl["{{MP3_SCRIPTS}}"] = layout.MP3ScriptsOverride
		} else {
			repl["{{MP3_SCRIPTS}}"] = siteMp3InitScripts(pub)
		}
	} else {
		repl["{{MP3_SECTION}}"] = ""
		repl["{{MP3_SCRIPTS}}"] = ""
	}
	for k, v := range extra {
		repl[k] = v
	}
	out := tpl
	for k, v := range repl {
		out = strings.ReplaceAll(out, k, v)
	}
	return []byte(out)
}
