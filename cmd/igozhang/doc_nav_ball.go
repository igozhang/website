package main

import (
	_ "embed"
	"encoding/json"
)

//go:embed doc-nav-ball.js
var docNavBallJS []byte

const docNavBallCSS = `
.doc-nav-ball-wrap{display:flex;justify-content:center;margin:.15rem 0 0}
.doc-link-aside-nav-only{min-height:0}
.doc-nav-ball{position:relative;width:3.25rem;height:3.25rem;border-radius:50%;background:var(--panel);border:1px solid var(--line);box-shadow:0 1px 3px rgba(45,59,45,.06);flex-shrink:0;overflow:hidden}
.doc-nav-ball-half{position:absolute;left:0;width:100%;height:50%;border:0;background:transparent;cursor:pointer;padding:0;margin:0;font:inherit;color:var(--muted);-webkit-tap-highlight-color:transparent;touch-action:manipulation}
.doc-nav-ball-top{top:0;clip-path:path('M 0 0 h 3.25rem a 1.625rem 1.625rem 0 0 0 -3.25rem 0 Z')}
.doc-nav-ball-bottom{bottom:0;clip-path:path('M 0 0 a 1.625rem 1.625rem 0 0 1 3.25rem 0 v 1.625rem h -3.25rem Z')}
.doc-nav-ball-half:hover{background:rgba(61,107,74,.08);color:var(--accent)}
.doc-nav-ball-half:focus{outline:none}
.doc-nav-ball-half:focus-visible{background:rgba(61,107,74,.08);color:var(--accent)}
.doc-nav-ball-half:active{background:rgba(61,107,74,.12);color:var(--accent)}
.doc-nav-ball-icon{display:block;font-size:.72rem;line-height:1;pointer-events:none}
.doc-nav-ball-top .doc-nav-ball-icon{margin-top:.28rem}
.doc-nav-ball-bottom .doc-nav-ball-icon{margin-top:.62rem}
.doc-nav-ball-home{position:absolute;left:50%;top:50%;transform:translate(-50%,-50%);z-index:2;width:1.57rem;height:.76rem;border-radius:999px;display:flex;align-items:center;justify-content:center;font-size:.61rem;font-weight:600;letter-spacing:.01em;text-transform:lowercase;text-decoration:none;color:var(--accent);background:var(--doc-surface);border:1px solid var(--line);cursor:pointer;line-height:1;padding:0;-webkit-tap-highlight-color:transparent;touch-action:manipulation}
.doc-nav-ball-home:hover,.doc-nav-ball-home:focus-visible{background:rgba(61,107,74,.1);outline:2px solid rgba(61,107,74,.25);outline-offset:1px}
.doc-nav-ball-wrap-fixed{position:fixed;right:max(1rem,env(safe-area-inset-right));z-index:40;margin:0;pointer-events:none}
.doc-nav-ball-wrap-fixed .doc-nav-ball{pointer-events:auto}
@media(min-width:42.01rem){
.doc-nav-ball-wrap-fixed{top:max(1.25rem,env(safe-area-inset-top));bottom:auto}
}
@media(max-width:42rem){
.doc-nav-ball-wrap-fixed{bottom:max(5.25rem,env(safe-area-inset-bottom));top:auto}
}
`

func docNavBallCSSIfEnabled() string {
	if !docNavBallEnabled() {
		return ""
	}
	return docNavBallCSS
}

func docNavBallEnabled() bool {
	return getOpsSnapshot().docNavBallEnabled
}

func opsDocNavBallInjectScript() string {
	s := getOpsSnapshot()
	payload := map[string]any{"enabled": s.docNavBallEnabled}
	b, _ := json.Marshal(payload)
	return `<script>window.__fsDocNavBallOps=` + string(b) + `;</script>`
}

func docNavBallScripts(pub string) string {
	if !docNavBallEnabled() {
		return ""
	}
	baseJS, err := json.Marshal(pub)
	if err != nil {
		baseJS = []byte(`""`)
	}
	var b []byte
	b = append(b, opsDocNavBallInjectScript()...)
	b = append(b, `<script>`...)
	b = append(b, docNavBallJS...)
	b = append(b, `</script>
<script>(function(){
  function boot(){ if(window.__fsDocNavBallInit) window.__fsDocNavBallInit(`...)
	b = append(b, baseJS...)
	b = append(b, `); }
  if(document.readyState==="loading") document.addEventListener("DOMContentLoaded",boot);
  else boot();
})();</script>`...)
	return string(b)
}
