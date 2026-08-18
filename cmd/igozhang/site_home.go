package main

func siteHomePageCSS() string {
	return siteHeadCSS() + siteUIGuardCSS + `
:root{
  --bg:#e8efe8;--panel:#dde8dd;--text:#2d3b2d;--muted:#5a6b5a;
  --accent:#3d6b4a;--line:#b8c9b8;--input:#e0eae0;--shell-max:52rem;
}
*{box-sizing:border-box}
body{margin:0;font-family:system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;
  background:var(--bg);color:var(--text);line-height:1.5;min-height:100vh;display:flex;flex-direction:column}
.shell{width:100%;max-width:var(--shell-max);margin-left:auto;margin-right:auto;padding-left:1.25rem;padding-right:1.25rem;box-sizing:border-box}
.site-head .shell,.site-foot .shell{width:100%;max-width:var(--shell-max);box-sizing:border-box}
main.shell{width:100%;max-width:var(--shell-max);padding-top:1.25rem;padding-bottom:1.5rem;flex:1 0 auto;box-sizing:border-box}
main.shell.fs-home{overflow:visible}
main.shell > section{width:100%;box-sizing:border-box}
section{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:1rem 1.1rem;margin-bottom:1rem}
.phrase-next{border:1px solid var(--line);background:var(--panel);color:var(--accent);cursor:pointer;font-size:.8rem;padding:.28rem .65rem;border-radius:6px;white-space:nowrap}
.phrase-next:hover{background:rgba(61,107,74,.08)}
.toast-wrap{position:fixed;right:1rem;bottom:1rem;z-index:10000;max-width:min(96vw,22rem);display:flex;flex-direction:column;gap:.4rem;pointer-events:none}
.toast-wrap .toast{margin:0;padding:.5rem .65rem;border-radius:8px;font-size:.86rem;line-height:1.45;box-shadow:0 2px 10px rgba(0,0,0,.12);pointer-events:auto}
.toast-ok{background:rgba(220,240,220,.95);border:1px solid var(--line);color:var(--text)}
.toast-err{background:rgba(255,230,220,.95);border:1px solid #c9a090;color:#5a2a1a}
` + docHomeCSS() + docNavBallCSSIfEnabled() + siteMp3StylesOnly() + siteFootCSS
}

// homeIndexHTML 首页：文章模块在 main 内；MP3 在 main 与页脚之间（紧贴页脚上方）。
// 与 applySitePage 子页一致：PAGE_HEAD（含 __fsAbsRoot）、SITE_HEAD 仅 header 块、统一 footer。
// 引语仅由 SiteHeadHTML（siteHeadBlock）初始化，勿在此重复注入 PhrasesJSON。
func homeIndexHTMLTemplate() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
{{.PageHeadHTML}}
<title>{{.SiteTitle}}</title>
{{.HomeCSSLinkHTML}}
</head>
<body>
{{.SiteHeadHTML}}
<main class="shell fs-home" data-fs-pub="{{.FsPubAttr}}">
{{.DocHomeHTML}}
</main>
<div class="shell mp3-above-foot">{{.Mp3SectionHTML}}</div>
{{.SiteFootHTML}}
<div class="toast-wrap" id="toastWrap" aria-live="polite"></div>
<script>` + siteVisibilityHubJS + `</script>
{{.Mp3ScriptsHTML}}
{{.NavBallScriptsHTML}}
</body>
</html>`
}
