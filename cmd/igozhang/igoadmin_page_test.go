package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestRenderIgoAdminHTML_siteChrome(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	html := string(renderIgoAdminHTML("", dir, nil))
	if !strings.Contains(html, "class=\"site-head\"") || !strings.Contains(html, "class=\"site-foot\"") {
		t.Fatal("igoadmin page must include site head and footer")
	}
	if strings.Contains(html, "认证方式与") {
		t.Fatal("igoadmin must not show ops auth hint text")
	}
	if strings.Contains(html, "<style><style>") {
		t.Fatal("duplicate nested style tags")
	}
	if !strings.Contains(html, "id=\"btnIgoAuth\"") {
		t.Fatal("auth button missing")
	}
	if !strings.Contains(html, "IGO密钥认证") || !strings.Contains(html, `id="igoAuthSec"`) {
		t.Fatal("unified IGO auth panel missing")
	}
	if !strings.Contains(html, `id="igoMainSec"`) || !strings.Contains(html, "<h2>后台</h2>") {
		t.Fatal("igoadmin main section with 后台 heading missing")
	}
	if !strings.Contains(html, `class="igoadmin-page fs-module"`) {
		t.Fatal("igoadmin outer section missing")
	}
	if !strings.Contains(html, `href="`) || !strings.Contains(html, "/igoadmin/igoadmin.css") {
		t.Fatal("igoadmin must link external stylesheet")
	}
	if !strings.Contains(html, "/igoadmin/igoadmin.js") || !strings.Contains(html, "window.__IGOADMIN_BASE__") {
		t.Fatal("igoadmin must load external script with API base bootstrap")
	}
	if !strings.Contains(string(igoadminCSS), ".igoadmin-block{background:var(--panel)") {
		t.Fatal("igoadmin inner blocks must keep panel chrome in CSS")
	}
	if !strings.Contains(html, `id="tabNew"`) || !strings.Contains(html, `id="tabEdit"`) || !strings.Contains(html, `id="tabBatch"`) {
		t.Fatal("igoadmin article mode tabs missing")
	}
	if !strings.Contains(html, `id="panelNew"`) || !strings.Contains(html, `id="panelEdit"`) || !strings.Contains(html, `id="panelBatch"`) {
		t.Fatal("igoadmin article mode panels missing")
	}
	if !strings.Contains(html, `id="panelEdit" class="igoadmin-mode-panel hidden"`) {
		t.Fatal("edit panel must be hidden by default")
	}
	if !strings.Contains(html, `id="panelBatch" class="igoadmin-mode-panel hidden"`) {
		t.Fatal("batch panel must be hidden by default")
	}
	if !strings.Contains(html, `id="btnBatchApply"`) {
		t.Fatal("batch apply button missing")
	}
	if !strings.Contains(html, `id="btnEditPreview"`) || !strings.Contains(html, ".doc-code-wrap") {
		t.Fatal("igoadmin preview or doc code chrome missing")
	}
	if strings.Contains(html, "{{DOC_MERMAID_SCRIPTS}}") || strings.Contains(html, "{{IGO_MERMAID_LIB}}") || strings.Contains(html, "{{IGO_MERMAID_JS}}") {
		t.Fatal("igoadmin left unreplaced mermaid template placeholders")
	}
	if regexp.MustCompile(`(?i)<script[^>]+src=["'][^"']*mermaid[^"']*["']`).MatchString(html) {
		t.Fatal("igoadmin must not preload mermaid scripts on first paint")
	}
	if !strings.Contains(html, "__IGO_MERMAID_LIB__") || !strings.Contains(html, "__IGO_MERMAID_JS__") {
		t.Fatal("igoadmin must expose mermaid asset URLs for on-demand preview load")
	}
	if !strings.Contains(html, "/assets/doc/mermaid.min.js") || !strings.Contains(html, "/assets/doc/mermaid.js") {
		t.Fatal("igoadmin bootstrap must include versioned mermaid asset URLs")
	}
	if !strings.Contains(string(igoadminJS), "ensureMermaidBundle") || !strings.Contains(string(igoadminJS), "previewEpochByPanel") || !strings.Contains(string(igoadminJS), "initDocMermaid") {
		t.Fatal("igoadmin preview must lazy-load mermaid then call initDocMermaid")
	}
	if !strings.Contains(html, `id="btnBatchCatAdd"`) || !strings.Contains(html, "igoadmin-authed-block") {
		t.Fatal("igoadmin batch cat add or authed block missing")
	}
	if !strings.Contains(string(igoadminJS), "function api(p){ return apiBase+p; }") {
		t.Fatal("igoadmin must use api() as URL builder like ops page")
	}
	if !strings.Contains(string(igoadminJS), `addEventListener("beforeunload"`) {
		t.Fatal("igoadmin must warn on page leave when unsaved")
	}
	if !strings.Contains(string(igoadminJS), "runNewAutoSave") || !strings.Contains(string(igoadminJS), "runEditAutoSave") {
		t.Fatal("igoadmin must auto-save on new and edit tabs")
	}
	if strings.Contains(string(igoadminJS), "return fetch(api(p),") && !strings.Contains(string(igoadminJS), "function apiFetchJSON") {
		t.Fatal("igoadmin must use apiFetchJSON helper for authenticated requests")
	}
	if strings.Contains(html, "localStorage.getItem(\"igoadminToken\")") || strings.Contains(html, "localStorage.setItem(\"igoadminToken\"") {
		t.Fatal("igoadmin must not persist auth token in localStorage")
	}
	if !strings.Contains(html, "__igoAuthInit") && !strings.Contains(string(igoAuthJS), "__igoAuthInit") {
		t.Fatal("igoadmin must use shared __igoAuthInit")
	}
	if !strings.Contains(string(igoadminJS), "__igoAuthInit") {
		t.Fatal("igoadmin.js must call __igoAuthInit")
	}
	if !strings.Contains(string(igoadminJS), "function editCategoriesPayload") {
		t.Fatal("igoadmin must preserve multi-category on edit save")
	}
	if !strings.Contains(string(igoadminJS), "/api/igoadmin/articles?q=") {
		t.Fatal("igoadmin edit search must use admin articles API")
	}
	if !strings.Contains(string(igoadminJS), "newCreatedSlug=uploadSlug") &&
		!strings.Contains(string(igoadminJS), `newCreatedSlug=(x.j&&x.j.slug)||uploadSlug`) {
		t.Fatal("igoadmin upload-md must sync newCreatedSlug after success")
	}
	if !strings.Contains(string(igoadminJS), `function newArticleSaveMethod`) ||
		!strings.Contains(string(igoadminJS), `return newCreatedSlug?"PUT":"POST"`) {
		t.Fatal("igoadmin save-new must use PUT whenever article already created (incl. slug rename)")
	}
	if !strings.Contains(string(igoadminJS), "attachNewArticleOldSlug") {
		t.Fatal("igoadmin save-new must attach oldSlug when renaming after create/upload")
	}
	if !strings.Contains(string(igoadminJS), "已载入正文 · 当前没有合法 slug") {
		t.Fatal("igoadmin upload-md must load body without API when filename slug is invalid")
	}
	if !strings.Contains(string(igoadminJS), `function apiErrText`) {
		t.Fatal("igoadmin must map API slug errors to Chinese")
	}
	if !strings.Contains(html, "保留文章原有的多分类") {
		t.Fatal("igoadmin edit panel must hint about multi-category preservation")
	}
	if strings.Contains(string(igoadminJS), "localStorage.getItem(\"igoadminToken\")") {
		t.Fatal("igoadmin must not auto-login from localStorage")
	}
	if strings.Contains(html, "ssh.pass") {
		t.Fatal("igoadmin must not depend on ssh.pass")
	}
}
