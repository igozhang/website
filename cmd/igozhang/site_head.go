package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"strconv"
	"strings"
)

//go:embed site-phrase-init.js
var sitePhraseInitJS []byte

//go:embed site-beijing-time.js
var siteBeijingTimeJS []byte

const (
	// banner_mask：0=全透明，100=全遮挡；66 对应当前默认蒙版（顶/中/底约 .30/.48/.66）。
	siteBannerMaskDefault         = 66
	siteBannerMaskPresentTop      = 0.3
	siteBannerMaskPresentMid      = 0.48
	siteBannerMaskPresentBottom   = 0.66
	siteBannerMaskAt100Top        = siteBannerMaskPresentTop / siteBannerMaskPresentBottom
	siteBannerMaskAt100Mid        = siteBannerMaskPresentMid / siteBannerMaskPresentBottom
	siteBannerMaskAt100Bottom     = 1.0
)

const siteHeadCSSBase = `
.site-head{background:var(--panel);border-bottom:1px solid var(--line);padding:1rem 0}
.site-head.has-banner{padding:0}
.site-head .shell{width:100%;max-width:var(--shell-max,52rem);margin:0 auto;padding-left:1.25rem;padding-right:1.25rem;box-sizing:border-box}
.site-head.has-banner .shell{position:relative;isolation:isolate;--site-head-banner-inset-x:.625rem;padding-top:1rem;padding-bottom:1rem}
.site-head.has-banner .shell::before{content:"";position:absolute;inset:0 var(--site-head-banner-inset-x);z-index:0;border-radius:8px;background-image:var(--site-head-banner);background-size:cover;background-position:center 28%;pointer-events:none}
.site-head.has-banner .shell::after{content:"";position:absolute;inset:0 var(--site-head-banner-inset-x);z-index:0;border-radius:8px;background:{{BANNER_MASK_GRADIENT}};pointer-events:none}
.site-head.has-banner .shell>*{position:relative;z-index:1}
.site-head.has-banner h1 a.site-title-link,.site-head.has-banner .site-title-tagline{color:var(--text);text-shadow:0 0 1px rgba(232,239,232,.95),0 1px 3px rgba(221,232,221,.78)}
.site-head.has-banner .site-beijing-time{text-shadow:0 0 1px rgba(232,239,232,.9),0 1px 2px rgba(221,232,221,.65)}
.site-head.has-banner .phrase-tagline{border-top-color:rgba(184,201,184,.38)}
.site-head.has-banner .phrase-tagline .phrase-en{color:var(--text);text-shadow:0 0 1px rgba(232,239,232,.92),0 1px 2px rgba(221,232,221,.72)}
.site-head.has-banner .phrase-tagline .phrase-dot,.site-head.has-banner .phrase-tagline .phrase-zh{text-shadow:none}
.site-head.has-banner .phrase-tagline .phrase-next{background:transparent;border-color:rgba(61,107,74,.32);color:var(--accent);backdrop-filter:blur(2px);-webkit-backdrop-filter:blur(2px)}
.site-head.has-banner .phrase-tagline .phrase-next:hover{background:rgba(221,232,221,.22);border-color:rgba(61,107,74,.42)}
.site-head h1{font-size:1.2rem;font-weight:600;margin:0;letter-spacing:.02em;line-height:1.35;display:flex;flex-wrap:nowrap;align-items:baseline;gap:.5em;min-width:0}
@media (min-width:640px){.site-head h1{gap:.95em}}
@media (min-width:900px){.site-head h1{gap:1.35em}}
.site-title-tagline{font-size:.88rem;font-weight:400;color:var(--muted);margin:0;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:min(58vw,20rem);flex:0 1 auto;min-width:0}
.site-head h1 a.site-title-link{color:inherit;text-decoration:none;font:inherit;font-weight:inherit;letter-spacing:inherit}
.site-head h1 a.site-title-link:hover,.site-head h1 a.site-title-link:focus,.site-head h1 a.site-title-link:active,.site-head h1 a.site-title-link:visited{color:inherit;text-decoration:none;background:transparent;outline:none}
.phrase-tagline{margin-top:.75rem;padding-top:.85rem;border-top:1px solid var(--line)}
.phrase-tagline .phrase-row{display:grid;grid-template-columns:1fr auto;column-gap:.65rem;row-gap:.5rem;align-items:end;margin-top:0}
.phrase-tagline .phrase-line{font-size:.95rem;line-height:1.6;min-height:calc(2 * 1.6em);margin:0;min-width:0;grid-column:1;grid-row:1;display:flex;align-items:center;overflow:hidden}
.phrase-tagline .phrase-line-inner{display:-webkit-box;-webkit-box-orient:vertical;-webkit-line-clamp:2;line-clamp:2;overflow:hidden;text-overflow:ellipsis;line-height:1.6;flex:1;min-width:0}
.phrase-tagline .phrase-next{font-size:.84rem;padding:.32rem .7rem;grid-column:2;grid-row:1;justify-self:end;align-self:end;margin:0 .12rem .1rem 0;position:relative}
.phrase-tagline .phrase-next::after{content:"";position:absolute;right:4px;bottom:4px;width:5px;height:5px;border-radius:50%;pointer-events:none}
.phrase-tagline .phrase-next.phrase-src-embed::after{background:rgba(201,158,72,.9);box-shadow:0 0 3px rgba(201,158,72,.5),0 0 7px rgba(201,158,72,.22)}
.phrase-tagline .phrase-next.phrase-src-external::after{background:rgba(92,138,104,.88);box-shadow:0 0 3px rgba(92,138,104,.48),0 0 7px rgba(92,138,104,.2)}
@media (max-width:639px){
  .phrase-tagline .phrase-line{grid-column:1/-1;grid-row:1}
  .phrase-tagline .phrase-next{grid-column:2;grid-row:2;justify-self:end}
}
.phrase-row{display:flex;flex-wrap:wrap;align-items:flex-start;gap:.65rem;margin-top:.35rem}
.phrase-line{flex:1;min-width:min(100%,12rem);font-size:.88rem;line-height:1.55;margin:0}
.phrase-en{font-style:italic;color:var(--text)}
.phrase-dot{color:var(--line)}
.phrase-zh{color:var(--muted)}
.phrase-next{border:1px solid var(--line);background:var(--panel);color:var(--accent);cursor:pointer;font-size:.8rem;padding:.28rem .65rem;border-radius:6px;white-space:nowrap}
.phrase-next:hover{background:rgba(61,107,74,.08)}
.site-beijing-time{display:block;font-size:.8rem;color:var(--muted);margin:.4rem 0 0;font-variant-numeric:tabular-nums}
`

func siteBannerMaskPercent() int {
	p := getOpsSnapshot().siteBannerMaskPercent
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

// siteBannerMaskGradient 按 igo.conf banner_mask（0=透明，100=全遮挡）线性插值蒙版。
func siteBannerMaskGradient() string {
	scale := float64(siteBannerMaskPercent()) / 100
	a0 := siteBannerMaskAt100Top * scale
	a1 := siteBannerMaskAt100Mid * scale
	a2 := siteBannerMaskAt100Bottom * scale
	return fmt.Sprintf("linear-gradient(180deg,rgba(232,239,232,%.4f) 0%%,rgba(221,232,221,%.4f) 68%%,rgba(221,232,221,%.4f) 100%%)", a0, a1, a2)
}

func siteHeadCSS() string {
	return strings.Replace(siteHeadCSSBase, "{{BANNER_MASK_GRADIENT}}", siteBannerMaskGradient(), 1)
}

func pageHeadTags(pub string) string {
	bh := baseHref(pub)
	escPub := html.EscapeString(strings.TrimSuffix(strings.TrimSpace(pub), "/"))
	var b []byte
	b = append(b, `<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
<meta name="fs-public-base" content="`...)
	b = append(b, escPub...)
	b = append(b, `">
`...)
	if bh != "" {
		b = append(b, `<base href="`...)
		b = append(b, html.EscapeString(bh)...)
		b = append(b, `">`...)
	} else {
		b = append(b, `<base href="/">`...)
	}
	iconHref, iconType := siteFaviconIconLink(pub)
	b = append(b, `
<link rel="icon" href="`...)
	b = append(b, html.EscapeString(iconHref)...)
	b = append(b, `" type="`...)
	b = append(b, html.EscapeString(iconType)...)
	b = append(b, `">
`...)
	b = append(b, pageHeadFaviconMeta(pub)...)
	b = append(b, "\n"...)
	b = append(b, sitePublicBaseBootScript(pub)...)
	return string(b)
}

func siteHeadOpenTag(pub string) string {
	cls := "site-head"
	var style string
	if u := siteBannerBackgroundURL(pub); u != "" {
		cls += " has-banner"
		style = `--site-head-banner:url('` + html.EscapeString(u) + `')`
	}
	if style != "" {
		return `<header class="` + cls + `" style="` + style + `">`
	}
	return `<header class="` + cls + `">`
}

func phraseNextSwitchMarkup() string {
	cls := "phrase-next phrase-src-embed"
	if phraseSourceIsExternal() {
		cls = "phrase-next phrase-src-external"
	}
	return `<button type="button" class="` + cls + `" id="phraseNext" title="Show another sentence">switch</button>`
}

func siteHeadBlock(homeHref, pub string, phrasesJSON template.JS) string {
	escHome := html.EscapeString(homeHref)
	escTagline := html.EscapeString(siteTagline())
	var b []byte
	b = append(b, siteHeadOpenTag(pub)...)
	b = append(b, `
<div class="shell">
<h1><a class="site-title-link" href="`...)
	b = append(b, escHome...)
	b = append(b, `" title="返回主站">`+html.EscapeString(siteDisplayName())+`</a><span class="site-title-tagline" aria-hidden="true">`...)
	b = append(b, escTagline...)
	b = append(b, `</span></h1>
<time class="site-beijing-time" id="siteBeijingTime" aria-label="北京时间"></time>
<div class="phrase-tagline">
<div class="phrase-row">
<p class="msg phrase-line" id="phraseLine" aria-live="polite"><span class="phrase-line-inner"><span class="phrase-en"></span><span class="phrase-dot"> —— </span><span class="phrase-zh"></span></span></p>
`...)
	b = append(b, phraseNextSwitchMarkup()...)
	b = append(b, `
</div>
</div>
</div>
</header>
<script>
if(!window.__fsVisibilityHub){
  window.__fsVisibilityHub=true;
  document.addEventListener("visibilitychange",function(){
    try{
      if(document.visibilityState==="hidden") document.dispatchEvent(new CustomEvent("fs:pagehidden"));
      else document.dispatchEvent(new CustomEvent("fs:pagevisible"));
    }catch(e){}
  });
}
</script>
<script>`...)
	b = append(b, sitePhraseInitJS...)
	b = append(b, `</script>
<script>window.__fsPhraseInit&&window.__fsPhraseInit(`...)
	b = append(b, []byte(phrasesJSON)...)
	b = append(b, `);</script>`...)
	b = append(b, siteUIGuardScript()...)
	b = append(b, siteBeijingTimeScript()...)
	return string(b)
}

func siteBeijingTimeScript() string {
	off := getOpsSnapshot().beijingOffsetHours
	if off > 12 {
		off = 12
	} else if off < -12 {
		off = -12
	}
	return "<script>window.__fsBeijingOffsetHours=" + strconv.Itoa(off) + ";</script><script>" + string(siteBeijingTimeJS) + "</script>"
}

func siteHeadPhrasesJSON() template.JS {
	raw, err := json.Marshal(phraseBankForUI())
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(string(raw))
}

func siteHeadHTML(homeHref, pub string, phrasesJSON template.JS) template.HTML {
	return template.HTML("<style>" + siteHeadCSS() + siteUIGuardCSS + "</style>" + siteHeadBlock(homeHref, pub, phrasesJSON))
}
