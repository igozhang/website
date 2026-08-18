package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

const pathFs = "/fs"

//go:embed fs.html
var fsPageHTML []byte

//go:embed ui/mod-transfer.html
var modTransferHTML []byte

//go:embed ui/mod-stash.html
var modStashHTML []byte

//go:embed ui/mod-browse.html
var modBrowseHTML []byte

//go:embed ui/ui-shared.js
var uiSharedJS []byte

//go:embed ui/mod-transfer.js
var modTransferJS []byte

//go:embed ui/mod-stash.js
var modStashJS []byte

//go:embed ui/mod-browse.js
var modBrowseJS []byte

const siteFsModuleCSS = `
:root{--bg:#e8efe8;--panel:#dde8dd;--text:#2d3b2d;--muted:#5a6b5a;--accent:#3d6b4a;--line:#b8c9b8;--input:#f4faf4;--shell-max:52rem}
body{margin:0;font-family:system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;
  background:var(--bg);color:var(--text);line-height:1.5;min-height:100vh;display:flex;flex-direction:column}
h2{font-size:.95rem;margin:0 0 .75rem;color:var(--muted);font-weight:600;text-transform:uppercase;letter-spacing:.06em}
.row{display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
input[type=url],input[type=text]{flex:1;min-width:12rem;padding:.45rem .6rem;border:1px solid var(--line);
  border-radius:6px;background:var(--input);color:var(--text)}
button{padding:.45rem .9rem;border-radius:6px;border:1px solid var(--accent);background:var(--accent);
  color:#f5faf5;cursor:pointer;font-size:.9rem}
button:disabled{opacity:.55;cursor:not-allowed}
.breadcrumb{font-size:.85rem;color:var(--muted);margin-bottom:.5rem;word-break:break-all}
.breadcrumb a{color:var(--accent);text-decoration:none}
.breadcrumb a:hover{text-decoration:underline}
.preview-tab{color:var(--accent);font-size:.85rem;text-decoration:none;white-space:nowrap}
.preview-tab:hover{text-decoration:underline}
button.list-del-btn{color:#8b4a3a;font-size:.85rem;background:none;border:none;border-radius:0;padding:0;margin:0;cursor:pointer;font:inherit;line-height:inherit;white-space:nowrap;box-shadow:none}
button.list-del-btn:hover{text-decoration:underline;color:#6e3a2e;background:none}
button.list-del-btn:disabled{opacity:.45;cursor:not-allowed;text-decoration:none}
.col-ops{text-align:right;white-space:nowrap;width:1%;padding-left:.55rem}
.list-name-link{color:var(--text);text-decoration:none}
.list-name-link:visited{color:var(--text)}
.list-name-link:hover{color:var(--accent);text-decoration:underline}
.table-scroll{width:100%;overflow-x:auto;-webkit-overflow-scrolling:touch;margin-top:.25rem}
table{width:100%;border-collapse:collapse;font-size:.9rem;min-width:0}
th,td{text-align:left;padding:.35rem .4rem;border-bottom:1px solid var(--line);vertical-align:top}
th{color:var(--muted);font-weight:600;font-size:.75rem;text-transform:uppercase}
tr:hover td{background:rgba(61,107,74,.06)}
.size{color:var(--muted);text-align:right;white-space:nowrap}
.col-time{white-space:nowrap;font-variant-numeric:tabular-nums}
@media (max-width:640px){
  table{font-size:.82rem;min-width:0}
  th,td{padding:.3rem .28rem}
  .size:not(.col-time){white-space:nowrap}
}
.msg{font-size:.85rem;margin-top:.5rem;color:var(--muted)}
.fs-transfer-section>div.row{gap:.45rem;align-items:center}
.transfer-progress{display:flex;align-items:center;gap:.55rem;margin-top:.35rem;flex-wrap:wrap;min-height:1.35rem}
.transfer-progress[hidden]{display:none!important}
.transfer-progress progress{width:min(100%,16rem);height:.55rem;accent-color:var(--accent)}
.transfer-progress .progress-label{font-size:.84rem;color:var(--muted);flex:1;min-width:10rem}
.stash-textarea{width:100%;min-height:calc(2 * 1.45em + 1rem);max-height:40vh;padding:.45rem .55rem;border:1px solid var(--line);border-radius:6px;background:var(--input);color:var(--text);font:inherit;font-size:.88rem;line-height:1.45;resize:vertical;box-sizing:border-box}
.stash-toolbar{display:flex;flex-wrap:wrap;gap:.5rem;margin-top:.5rem;align-items:center;justify-content:space-between;width:100%}
.stash-toolbar-left{display:flex;flex-wrap:wrap;gap:.5rem;align-items:center}
.stash-snippet{font-size:.82rem;color:var(--muted);margin-top:.45rem;padding:.4rem .55rem;background:rgba(255,255,255,.35);border:1px solid var(--line);border-radius:6px;word-break:break-all;line-height:1.45}
.stash-snippet[hidden]{display:none!important}
.phrase-next{border:1px solid var(--line);background:var(--panel);color:var(--accent);cursor:pointer;font-size:.8rem;padding:.28rem .65rem;border-radius:6px;white-space:nowrap}
.phrase-next:hover{background:rgba(61,107,74,.08)}
.err{color:#8b4a3a}
.ok{color:var(--accent)}
.fs-usage-block{margin-top:.5rem;width:100%;max-width:100%}
.fs-usage-block progress{width:100%;height:.5rem}
.fs-usage-block progress.fs-usage-bar-dgreen{accent-color:#156630}
.fs-usage-block progress.fs-usage-bar-lgreen{accent-color:#7ccb8a}
.fs-usage-block progress.fs-usage-bar-red{accent-color:#e53935}
.fs-usage-block progress.fs-usage-bar-dred{accent-color:#8b1818}
.browse-controls{margin:.35rem 0 0;align-items:center;gap:.45rem .55rem}
.browse-controls label.msg{margin-top:0;margin-bottom:0;padding:.45rem 0;line-height:1.45;display:inline-flex;align-items:center;white-space:nowrap}
.browse-controls select{
  font:inherit;font-size:.85rem;line-height:1.45;color:var(--muted);
  padding:.45rem 1.75rem .45rem .55rem;min-height:2.35rem;
  border:1px solid var(--line);border-radius:6px;background-color:var(--panel);cursor:pointer;
  appearance:none;-webkit-appearance:none;
  background-image:url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='none' stroke='%235a6b5a' stroke-width='1.4' stroke-linecap='round' stroke-linejoin='round' d='M3 4.25 L6 7.75 L9 4.25'/%3E%3C/svg%3E");
  background-repeat:no-repeat;background-position:right .45rem center;background-size:.65rem}
.browse-controls select:focus{outline:2px solid rgba(61,107,74,.35);outline-offset:1px}
.browse-controls select option{color:var(--text);background:var(--panel)}
.browse-controls input[type=text]{min-width:8rem;max-width:14rem;min-height:2.35rem}
.stash-conflict{margin-top:.6rem;padding:.55rem .65rem;background:rgba(255,248,220,.92);border:1px solid #c9b860;border-radius:6px;font-size:.86rem}
.stash-conflict[hidden]{display:none!important}
.stash-conflict-msg{margin:0 0 .5rem;line-height:1.45;color:var(--text)}
.stash-conflict-actions{display:flex;flex-wrap:wrap;gap:.45rem;align-items:center}
.hidden{display:none!important}
`

const siteVisibilityHubJS = `(function(){
  if(!window.__fsVisibilityHub){
    window.__fsVisibilityHub=true;
    document.addEventListener("visibilitychange", function(){
      try{
        if(document.visibilityState==="hidden"){
          document.dispatchEvent(new CustomEvent("fs:pagehidden"));
        } else {
          document.dispatchEvent(new CustomEvent("fs:pagevisible"));
        }
      }catch(e){}
    });
  }
})();`

func siteFsPageHeadCSS() string {
	return siteSubpageWithMp3CSS() + siteFsModuleCSS
}

func fsModulesHTML() string {
	return string(modTransferHTML) + string(modStashHTML) + string(modBrowseHTML)
}

func fsBootScripts(pub string) string {
	baseJS, err := json.Marshal(pub)
	if err != nil {
		baseJS = []byte(`""`)
	}
	var b strings.Builder
	b.WriteString(`<script>window.__fsPublicBase=`)
	b.Write(baseJS)
	b.WriteString(`;</script>
<script>`)
	b.Write(uiSharedJS)
	b.WriteString(`</script>
`)
	b.WriteString(igoAuthScriptTag())
	b.WriteString(`
<script>`)
	b.Write(modTransferJS)
	b.WriteString(`</script>
<script>`)
	b.Write(modStashJS)
	b.WriteString(`</script>
<script>`)
	b.Write(modBrowseJS)
	b.WriteString(`</script>
<script>(function(){
  var modulesReady=false;
  function bootModules(){
    if(modulesReady){
      if(window.__fs&&window.__fs.bus) window.__fs.bus.emit("auth:ok");
      return;
    }
    modulesReady=true;
    if(window.__fsTransferInit) window.__fsTransferInit();
    if(window.__fsStashInit) window.__fsStashInit();
    if(window.__fsBrowseInit) window.__fsBrowseInit();
  }
  function boot(){
    window.__fs=window.__fsSharedInit(window.__fsPublicBase||"");
    var auth=window.__igoAuthInit({
      apiBase: window.__fsPublicBase||"",
      onReady: bootModules,
      onLogout: function(){
        if(window.__fs&&window.__fs.bus) window.__fs.bus.emit("auth:logout");
      }
    });
    if(window.__fs&&auth){
      window.__fs.authHeaders=auth.authHeaders;
      window.__fs.withAuthQuery=auth.withAuthQuery;
      window.__fs.requireAuth=auth.requireAuth;
      window.__fs.handleAuthFailure=auth.handleAuthFailure;
      window.__fs.getAuthToken=auth.getToken;
      window.__fs.logoutAuth=auth.logout;
    }
  }
  if(document.readyState==="loading") document.addEventListener("DOMContentLoaded", boot);
  else boot();
})();</script>`)
	return b.String()
}

func renderFsHTML(pub string, r *http.Request) []byte {
	ph := siteHeadPhrasesJSON()
	return applySitePage(string(fsPageHTML), pub, sitePageLayout{HeadCSS: siteFsPageHeadCSS(), MP3: true, MP3CSS: true}, map[string]string{
		"{{SITE_TITLE}}":      siteDisplayName(),
		"{{FS_MODULES}}":      fsModulesHTML(),
		"{{STASH_OPS}}":       opsStashInjectScript(),
		"{{VISIBILITY_HUB}}":  siteVisibilityHubJS,
		"{{PHRASES_JSON}}":    string(ph),
		"{{FS_BOOT_SCRIPTS}}": fsBootScripts(pub),
		"{{IGO_AUTH_PANEL}}":  igoAuthPanelHTML,
		"{{IGO_AUTH_CSS}}":    igoAuthCSS,
	}, visitorGeoLabel(r))
}

func registerFsPageHandler(mux *http.ServeMux) {
	serve := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
		if pub == "" {
			if p := strings.TrimSpace(getenv("BASE_PATH", "")); p != "" {
				pub = strings.TrimSuffix(p, "/")
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		issueMp3GateCookie(w, r)
		_, _ = w.Write(renderFsHTML(pub, r))
	}
	mux.HandleFunc(pathFs, serve)
	mux.HandleFunc(pathFs+"/", serve)
}
