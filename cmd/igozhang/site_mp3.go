package main

import (
	_ "embed"
	"encoding/json"
	"html"
	"strings"
)

// mp3-init.js：首页与子页唯一 MP3 逻辑源（//go:embed）；勿用非 UTF-8 工具保存（见 encoding_test.go）。
//
//go:embed mp3-init.js
var mp3InitJS []byte

//go:embed ui/mod-mp3.html
var modMp3HTML []byte

// siteMp3PageCSS：首页 shell / section / MP3 / toast（子页请用 siteSubpageCSS，不含 MP3）。
const siteMp3PageCSS = siteSubpageCSS + `
.shell.mp3-above-foot{width:100%;max-width:var(--shell-max,52rem);margin-left:auto;margin-right:auto;padding:0 1.25rem .5rem;box-sizing:border-box;flex-shrink:0}
.shell.mp3-above-foot > .mp3-section{margin-top:0;margin-bottom:0}
main.shell.fs-home:has(+ .mp3-above-foot),main.shell.fs-page-body:has(+ .mp3-above-foot){flex:1 0 auto;padding-bottom:.75rem}
.mp3-section{margin-top:0;margin-bottom:0;width:100%;container-type:inline-size;container-name:mp3sec}
.fs-page-body .mp3-section{margin-bottom:0}
.mp3-head-row{display:flex;flex-wrap:wrap;align-items:center;gap:.45rem .55rem;margin-bottom:.35rem;width:100%;min-width:0}
.mp3-heading{font-size:.95rem;margin:0;color:var(--muted);font-weight:600;letter-spacing:.05em;text-transform:none;white-space:nowrap;flex:0 0 auto}
.mp3-main-row{display:flex;flex:0 1 auto;align-items:center;gap:.35rem .5rem;min-width:0;flex-wrap:wrap}
.mp3-tags-inline{display:flex;flex-wrap:wrap;align-items:center;justify-content:flex-start;gap:.28rem .4rem;min-width:0;min-height:1.6rem;flex:0 1 auto}
.mp3-controls{display:flex;align-items:center;gap:.4rem .5rem;flex:0 0 auto;margin-left:auto;min-width:0}
.mp3-search-wrap{position:relative;flex:0 1 9rem;width:9rem;max-width:10.5rem;min-width:6.5rem}
.mp3-search-input{width:100%;box-sizing:border-box;font:inherit;font-size:.78rem;line-height:1.35;padding:.24rem .52rem;border:1px solid var(--line);border-radius:6px;background:var(--panel);color:var(--text);min-height:1.85rem}
.mp3-search-input:disabled{opacity:.55;cursor:not-allowed}
.mp3-search-input::placeholder{color:var(--muted);opacity:.88}
.mp3-search-input:focus{outline:2px solid rgba(61,107,74,.35);outline-offset:1px}
.mp3-search-results{position:absolute;left:0;right:0;top:calc(100% + .18rem);margin:0;padding:0;list-style:none;max-height:10.5rem;overflow-y:auto;background:var(--panel);border:1px solid var(--line);border-radius:6px;box-shadow:0 4px 14px rgba(0,0,0,.1);z-index:24}
.mp3-search-results[hidden]{display:none!important}
.mp3-search-results li{margin:0;padding:0}
.mp3-search-results button{display:block;width:100%;text-align:left;font:inherit;font-size:.76rem;line-height:1.35;padding:.36rem .52rem;border:none;background:transparent;cursor:pointer;color:var(--text)}
.mp3-search-results button:disabled{cursor:default;color:var(--muted)}
.mp3-search-results button:not(:disabled):hover,.mp3-search-results button:not(:disabled):focus{background:rgba(61,107,74,.1);outline:none}
.mp3-search-results button.mp3-search-hit-on{background:rgba(61,107,74,.16)}
.mp3-search-results button strong{display:block;font-weight:600;color:var(--text)}
.mp3-search-results button span{display:block;font-size:.72rem;color:var(--muted);margin-top:.08rem}
.mp3-row{display:flex;flex-wrap:nowrap;gap:.5rem;align-items:center;flex:0 0 auto}
@container mp3sec (min-width:42.01rem){
  @media (min-width:64.01rem){
    .mp3-main-row{flex:1 1 12rem}
    .mp3-tags-inline{flex:1 1 auto}
  }
}
@container mp3sec (max-width:42rem){
  .mp3-head-row{flex-direction:column;align-items:stretch;gap:.55rem}
  .mp3-main-row{flex-direction:column;align-items:stretch;width:100%;gap:.45rem}
  .mp3-controls{flex-direction:column;align-items:stretch;margin-left:0;width:100%;gap:.38rem}
  .mp3-search-wrap{width:100%;max-width:none;min-width:0;flex:0 0 auto}
  .mp3-row{width:100%;justify-content:flex-start;flex-wrap:wrap;flex:0 0 auto}
}
@media (max-width:64rem){
  .mp3-head-row{flex-direction:column;align-items:stretch;gap:.55rem}
  .mp3-main-row{flex-direction:column;align-items:stretch;width:100%;gap:.45rem}
  .mp3-controls{flex-direction:column;align-items:stretch;margin-left:0;width:100%;gap:.38rem}
  .mp3-search-wrap{width:100%;max-width:none;min-width:0;flex:0 0 auto}
  .mp3-row{width:100%;justify-content:flex-start;flex-wrap:wrap;flex:0 0 auto}
}
.phrase-next.mp3-active{border-color:var(--accent);background:rgba(61,107,74,.12);color:#1e4028}
#mp3Loop.phrase-next{min-width:2.1rem;padding-left:.48rem;padding-right:.48rem;border-style:solid;font-size:.78rem;font-weight:600;letter-spacing:.02em}
#mp3Loop.phrase-next:not(.mp3-loop-on){opacity:.88;color:var(--muted)}
#mp3Loop.phrase-next.mp3-loop-on{border-color:var(--accent);background:rgba(61,107,74,.28);color:#0f3018;font-weight:700;box-shadow:inset 0 0 0 1px rgba(61,107,74,.32)}
button.mp3-tag{font:inherit;font-size:.72rem;padding:.2rem .48rem;border-radius:5px;line-height:1.35}
.mp3-tag.mp3-tag-on{border-color:var(--accent);background:rgba(61,107,74,.12);color:#1e4028}
.mp3-title{font-size:.88rem;margin:.2rem 0 0;color:var(--text);word-break:break-all;display:none}
.mp3-title.mp3-title-visible{display:block}
.mp3-meta{display:none;margin-top:.45rem;--mp3-cover-side:calc(4 * 0.82rem * 1.55)}
.mp3-meta.mp3-meta-visible{display:block}
.mp3-cover{display:block;box-sizing:border-box;width:var(--mp3-cover-side);height:var(--mp3-cover-side);max-width:min(100%,var(--mp3-cover-side));border-radius:8px;border:1px solid var(--line);object-fit:contain;background:var(--panel)}
#mp3Audio,#mp3Prefetch{display:none}
`

func siteMp3Section() string {
	return string(modMp3HTML)
}

// siteMp3InitScripts 首页与默认子页：外置 mp3-init.js（与文章页同源缓存）；暂不懒加载列表。
func siteMp3InitScripts(pub string) string {
	return siteMp3BootScripts(pub, true, false)
}

// siteMp3ArticleInitScripts 文章页：外置 mp3-init.js + 懒加载列表（跨页续播 handoff 仍立即拉列表）。
func siteMp3ArticleInitScripts(pub string) string {
	return siteMp3BootScripts(pub, true, true)
}

func siteMp3BootScripts(pub string, externalJS, lazyList bool) string {
	baseJS, err := json.Marshal(pub)
	if err != nil {
		baseJS = []byte(`""`)
	}
	var b []byte
	b = append(b, opsMp3OnlyInjectScript()...)
	b = append(b, `<script>window.__fsPublicBase=`...)
	b = append(b, baseJS...)
	b = append(b, `;</script>`...)
	if lazyList {
		b = append(b, `<script>window.__fsMp3LazyList=true;</script>`...)
	}
	if externalJS {
		b = append(b, `<script src="`...)
		b = append(b, docMp3InitJSAssetURL(pub)...)
		b = append(b, `" defer></script>
<script defer>`...)
	} else {
		b = append(b, `<script>`...)
		b = append(b, mp3InitJS...)
		b = append(b, `</script>
<script>`...)
	}
	b = append(b, `(function(){
  function boot(){ if(window.__fsMp3Init) window.__fsMp3Init(window.__fsPublicBase||""); }
  if(document.readyState==="loading") document.addEventListener("DOMContentLoaded", boot);
  else boot();
})();</script>`...)
	return string(b)
}

func siteMp3Scripts(pub string) string {
	return `<div class="toast-wrap" id="toastWrap" aria-live="polite"></div>` + siteMp3InitScripts(pub)
}

func siteMp3Block(pub string) string {
	return siteMp3Section() + siteMp3Scripts(pub)
}

// siteMp3StylesOnly 为 MP3 区块样式（不含 siteSubpageCSS，避免与子页 head 重复）。
func siteMp3StylesOnly() string {
	if strings.HasPrefix(siteMp3PageCSS, siteSubpageCSS) {
		return siteMp3PageCSS[len(siteSubpageCSS):]
	}
	return siteMp3PageCSS
}

func siteSubpageWithMp3CSS() string {
	return siteSubpageHeadCSS() + siteMp3StylesOnly()
}

func siteSubpageMp3Block(pub string) string {
	return siteMp3Section() + siteMp3InitScripts(pub)
}

func fsPubAttr(pub string) string {
	return html.EscapeString(pub)
}
