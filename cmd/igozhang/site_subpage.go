package main

// 子页面共用样式；知键/五子棋/象棋等页可挂载 MP3（siteSubpageMp3Block）。

const (
	pathGomoku  = "/gomoku"
	pathXiangqi = "/xiangqi"
	pathTexas   = "/texas"
)

const siteSubpageCSS = `
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
main.shell.fs-page-body{display:block}
main.shell > section{width:100%;box-sizing:border-box}
.fs-page-body .gomoku-page{padding:1rem 1.1rem;margin:0 0 1rem;width:100%;flex:0 0 auto}
.fs-page-body .xiangqi-page{padding:1rem 1.1rem;margin:0 0 1rem;width:100%;flex:0 0 auto}
.fs-page-body .texas-page{padding:1rem 1.1rem;margin:0 0 1rem;width:100%;flex:0 0 auto}
.fs-page-body .zhijian-page{padding:1rem 1.1rem;margin:0 0 1rem;width:100%;flex:0 0 auto}
.fs-page-body .ops-page{padding:1rem 1.1rem;margin:0 0 1rem;width:100%;flex:0 0 auto}
.fs-page-body .igoadmin-page{padding:1rem 1.1rem;margin:0 0 1rem;width:100%;flex:0 0 auto}
section{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:1rem 1.1rem;margin-bottom:1rem}
.toast-wrap{position:fixed;right:1rem;bottom:1rem;z-index:10000;max-width:min(96vw,22rem);display:flex;flex-direction:column;gap:.4rem;pointer-events:none}
.toast-wrap .toast{margin:0;padding:.5rem .65rem;border-radius:8px;font-size:.86rem;line-height:1.45;box-shadow:0 2px 10px rgba(0,0,0,.12);pointer-events:auto}
.toast-ok{background:rgba(220,240,220,.95);border:1px solid var(--line);color:var(--text)}
.toast-err{background:rgba(255,230,220,.95);border:1px solid #c9a090;color:#5a2a1a}
`

func siteSubpageHeadCSS() string {
	return siteHeadCSS() + siteUIGuardCSS + siteSubpageCSS + siteFootCSS + docNavBallCSSIfEnabled()
}

func siteSubpageTail(pub, extra string) string {
	return `<div class="toast-wrap" id="toastWrap" aria-live="polite"></div>` + extra + docNavBallScripts(pub)
}

func pubPathZhijian(pub string) string {
	return pubPath(pub, "zhijian")
}
