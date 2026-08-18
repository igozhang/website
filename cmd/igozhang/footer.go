package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const secret07RelDir = "secret07"

//go:embed gomoku.html
var gomokuHTML []byte

//go:embed zhijian.html
var zhijianHTML []byte

//go:embed xiangqi.html
var xiangqiHTML []byte

//go:embed xiangqi-rules.js
var xiangqiRulesJS []byte

//go:embed xiangqi-board.js
var xiangqiBoardJS []byte

//go:embed xiangqi-page.js
var xiangqiPageJS []byte

//go:embed xiangqi-ai-worker.js
var xiangqiAIWorkerJS []byte

//go:embed texas.html
var texasHTML []byte

//go:embed texas-rules.js
var texasRulesJS []byte

//go:embed texas-board.js
var texasBoardJS []byte

//go:embed texas-summary.js
var texasSummaryJS []byte

//go:embed texas-page.js
var texasPageJS []byte

//go:embed texas-ai-worker.js
var texasAIWorkerJS []byte

type footerLink struct {
	Text string `json:"text"`
	Href string `json:"href"`
}

type footerConfig struct {
	Gomoku  string       `json:"gomoku"`
	Xiangqi string       `json:"xiangqi"`
	Texas   string       `json:"texas"`
	Links   []footerLink `json:"links"`
}

func secret07Dir(dataRoot string) string {
	return filepath.Join(dataRoot, secret07RelDir)
}

func gomokuHTMLPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), "gomoku.html")
}

func isSafeFooterHref(href string) bool {
	low := strings.ToLower(href)
	return strings.HasPrefix(low, "https://") || strings.HasPrefix(low, "http://") ||
		(strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//"))
}

// renderGomokuHTML 始终使用内嵌页面（避免 DATA_ROOT/secret07/gomoku.html 旧文件覆盖导致逻辑过期）。
func renderGomokuHTML(pub string, r *http.Request) []byte {
	return applySitePage(string(gomokuHTML), pub, sitePageLayout{MP3: true, MP3CSS: true}, nil, visitorGeoLabel(r))
}

func xiangqiScriptsBlock() string {
	rulesLit, _ := json.Marshal(string(xiangqiRulesJS))
	aiLit, _ := json.Marshal(string(xiangqiAIWorkerJS))
	return `<script>window.__fsXiangqiRulesSrc=` + string(rulesLit) + `;window.__fsXiangqiAiWorkerSrc=` + string(aiLit) + `;</script>
<script>` + string(xiangqiRulesJS) + `</script>
<script>` + string(xiangqiBoardJS) + `</script>
<script>` + string(xiangqiPageJS) + `</script>`
}

func renderXiangqiHTML(pub string, r *http.Request) []byte {
	return applySitePage(string(xiangqiHTML), pub, sitePageLayout{MP3: true, MP3CSS: true}, map[string]string{
		"{{XIANGQI_OPS_SCRIPT}}": opsXiangqiInjectScript(),
		"{{XIANGQI_SCRIPTS}}":    xiangqiScriptsBlock(),
	}, visitorGeoLabel(r))
}

func texasScriptsBlock() string {
	rulesLit, _ := json.Marshal(string(texasRulesJS))
	aiLit, _ := json.Marshal(string(texasAIWorkerJS))
	return `<script>window.__fsTexasRulesSrc=` + string(rulesLit) + `;window.__fsTexasAiWorkerSrc=` + string(aiLit) + `;</script>
<script>` + string(texasRulesJS) + `</script>
<script>` + string(texasSummaryJS) + `</script>
<script>` + string(texasBoardJS) + `</script>
<script>` + string(texasPageJS) + `</script>`
}

func renderTexasHTML(pub string, r *http.Request) []byte {
	return applySitePage(string(texasHTML), pub, sitePageLayout{MP3: true, MP3CSS: true}, map[string]string{
		"{{TEXAS_OPS_SCRIPT}}":     opsTexasInjectScript(),
		"{{TEXAS_SCRIPTS}}":        texasScriptsBlock(),
		"{{TEXAS_CARD_BACK_META}}": pageHeadTexasCardBackMeta(pub),
	}, visitorGeoLabel(r))
}

func renderZhijianHTML(pub string, r *http.Request) []byte {
	return applySitePage(string(zhijianHTML), pub, sitePageLayout{MP3: true, MP3CSS: true}, map[string]string{
		"{{SITE_SUBPAGE_TAIL}}": siteSubpageTail(pub, opsZjInjectScript()),
	}, visitorGeoLabel(r))
}

func registerFooterHandlers(mux *http.ServeMux, dataRoot string) {
	mux.HandleFunc("/api/footer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		jsonWrite(w, loadFooterConfigFromOps())
	})
	mux.HandleFunc("/api/zhijian/word3k", word3kHandler())
	mux.HandleFunc("/api/zhijian/word3k/", word3kHandler())
	mux.HandleFunc("/api/ops-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(opsClientJSON())
	})
	serveGomoku := func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = w.Write(renderGomokuHTML(pub, r))
	}
	mux.HandleFunc("/gomoku", serveGomoku)
	mux.HandleFunc("/gomoku/", serveGomoku)
	serveXiangqi := func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = w.Write(renderXiangqiHTML(pub, r))
	}
	mux.HandleFunc("/xiangqi", serveXiangqi)
	mux.HandleFunc("/xiangqi/", serveXiangqi)
	serveTexas := func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = w.Write(renderTexasHTML(pub, r))
	}
	mux.HandleFunc("/texas", serveTexas)
	mux.HandleFunc("/texas/", serveTexas)
	serveZhijian := func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = w.Write(renderZhijianHTML(pub, r))
	}
	mux.HandleFunc("/zhijian", serveZhijian)
	mux.HandleFunc("/zhijian/", serveZhijian)
}

// removeDeprecatedGomokuOverride 若存在旧版 secret07/gomoku.html 则提示删除（不再使用覆盖）。
func removeDeprecatedGomokuOverride(dataRoot string) {
	p := gomokuHTMLPath(dataRoot)
	if _, err := os.Stat(p); err == nil {
		log.Printf("ops-config: deprecated %s — remove this file; gomoku is embedded in binary", p)
	}
}
