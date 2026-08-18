package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

//go:embed article.html
var articlePageTpl []byte

var docStore *docindex.Store

func docAssetsRoot(dataRoot string) string {
	return filepath.Join(dataRoot, "assets")
}

const docDBFileName = "doc.db"

// docDBDir is DATA_ROOT/secret07/sqlitedb (SQLite WAL files live here, not under assets/).
func docDBDir(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), "sqlitedb")
}

func docDBPath(dataRoot string) string {
	return filepath.Join(docDBDir(dataRoot), docDBFileName)
}

func initDocStore(dataRoot string) error {
	seedDocAssetsIfEmpty(dataRoot)
	assets := docAssetsRoot(dataRoot)
	dbPath := docDBPath(dataRoot)
	st, err := docindex.Open(dbPath, assets)
	if err != nil {
		return err
	}
	if err := st.Reconcile(); err != nil {
		_ = st.Close()
		return err
	}
	if err := installBuiltinDemoArticle(st, assets); err != nil {
		_ = st.Close()
		return err
	}
	docStore = st
	log.Printf("doc: loaded %d articles, assets=%s db=%s", docCount(), assets, dbPath)
	scheduleDocBackup(dataRoot)
	preloadDocExportMemCache(dataRoot)
	return nil
}

func docCount() int {
	if docStore == nil {
		return 0
	}
	all, _ := docStore.ListAll()
	return len(all)
}

func seedDocAssetsIfEmpty(dataRoot string) {
	mdDir := filepath.Join(docAssetsRoot(dataRoot), "md")
	entries, err := os.ReadDir(mdDir)
	if err == nil && len(entries) > 0 {
		return
	}
	seed := strings.TrimSpace(os.Getenv("DOC_ASSETS_SEED"))
	if seed == "" {
		seed = "assets"
	}
	srcMD := filepath.Join(seed, "md")
	if st, err := os.Stat(srcMD); err != nil || !st.IsDir() {
		return
	}
	log.Printf("doc: seeding assets from %s -> %s", seed, docAssetsRoot(dataRoot))
	_ = copyTree(seed, docAssetsRoot(dataRoot))
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
}

func docHomeLatestN() int {
	n := getOpsSnapshot().docHomeLatest
	if n <= 0 {
		return 5
	}
	if n > 50 {
		return 50
	}
	return n
}

func docPicPublicBase(pub string) string {
	return pubPath(pub, "assets/doc-pic")
}

func registerDocHandlers(mux *http.ServeMux, dataRoot string) {
	mux.HandleFunc("/api/doc/search", docSearchHandler)
	mux.HandleFunc("/api/doc/home", func(w http.ResponseWriter, r *http.Request) {
		docHomeHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/doc/secret-article", func(w http.ResponseWriter, r *http.Request) {
		docSecretArticleHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/assets/doc-pic/", docPicFileHandler(dataRoot))
	mux.HandleFunc("/assets/doc/article.css", serveDocArticleCSS)
	mux.HandleFunc("/assets/doc/code-chrome.js", serveDocCodeChromeJS)
	mux.HandleFunc("/assets/doc/mermaid.min.js", serveDocMermaidLibJS)
	mux.HandleFunc("/assets/doc/mermaid.js", serveDocMermaidJS)
	mux.HandleFunc("/assets/doc/mp3.css", serveDocMp3CSS)
	mux.HandleFunc("/assets/doc/mp3-init.js", serveDocMp3InitJS)
	mux.HandleFunc("/assets/doc/home.css", serveDocHomeCSS)
	registerIgoAdminHandlers(mux, dataRoot)
}

func docHomeHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if docStore == nil {
		jsonWrite(w, map[string]any{"items": []any{}, "total": 0, "page": 1, "pageSize": 5, "totalPages": 1})
		return
	}
	cat := strings.TrimSpace(r.URL.Query().Get("category"))
	page := docParsePage(r.URL.Query().Get("page"))
	pageSize := docHomeLatestN()
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 50 {
			pageSize = n
		}
	}
	payload, err := docHomePayload(dataRoot, cat, page, pageSize)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "list failed")
		return
	}
	jsonWrite(w, payload)
}

func docParsePage(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 {
		return 1
	}
	return n
}

type docHomePayloadJSON struct {
	Items      []docHomeItemJSON `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}

func docHomePayload(dataRoot, category string, page, pageSize int) (docHomePayloadJSON, error) {
	if pageSize <= 0 {
		pageSize = docHomeLatestN()
	}
	if docStore == nil {
		return docHomePayloadJSON{Page: 1, PageSize: pageSize, TotalPages: 1}, nil
	}
	total, err := docStore.CountByCategory(category)
	if err != nil {
		return docHomePayloadJSON{}, err
	}
	totalPages := docTotalPages(total, pageSize)
	if page > totalPages {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	metas, err := docStore.ListByCategoryPage(category, pageSize, offset)
	if err != nil {
		return docHomePayloadJSON{}, err
	}
	items, err := docMetasToHomeItems(dataRoot, metas)
	if err != nil {
		return docHomePayloadJSON{}, err
	}
	return docHomePayloadJSON{
		Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages,
	}, nil
}

func docTotalPages(total, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 1
	}
	return (total + pageSize - 1) / pageSize
}

type docHomeItemJSON struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	PublishedAt int64  `json:"publishedAt"`
	Excerpt     string `json:"excerpt"`
}

const docCapsuleTitleMaxRunes = 22

func truncateDocCapsuleTitle(s string) string {
	rs := []rune(s)
	if len(rs) <= docCapsuleTitleMaxRunes {
		return s
	}
	return string(rs[:docCapsuleTitleMaxRunes]) + "…"
}

func docMetasToHomeItems(dataRoot string, metas []docindex.Meta) ([]docHomeItemJSON, error) {
	out := make([]docHomeItemJSON, 0, len(metas))
	for _, m := range metas {
		excerpt := ""
		if !isSecretAuthSlug(dataRoot, m.Slug) {
			if body, err := docStore.ReadBody(m.Slug); err == nil {
				excerpt = docindex.ExcerptMarkdownSkipTitle(body, 4, m.Title)
			}
		}
		out = append(out, docHomeItemJSON{
			Slug: m.Slug, Title: m.Title,
			Category: docindex.PrimaryCategory(m.Categories),
			PublishedAt: m.PublishedAt.Unix(), Excerpt: excerpt,
		})
	}
	return out, nil
}

func docSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if docStore == nil {
		jsonWrite(w, []any{})
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 20
	hits, err := docStore.Search(q, limit)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "search failed")
		return
	}
	type row struct {
		Slug        string `json:"slug"`
		Title       string `json:"title"`
		Categories  string `json:"categories"`
		Tags        string `json:"tags"`
		PublishedAt int64  `json:"publishedAt"`
		Views       int64  `json:"views"`
	}
	out := make([]row, 0, len(hits))
	for _, m := range hits {
		out = append(out, row{
			Slug: m.Slug, Title: m.Title, Categories: m.Categories, Tags: m.Tags,
			PublishedAt: m.PublishedAt.Unix(), Views: m.Views,
		})
	}
	jsonWrite(w, out)
}

func docPicFileHandler(dataRoot string) http.HandlerFunc {
	root := filepath.Join(docAssetsRoot(dataRoot), "doc-pic")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/assets/doc-pic/")
		rel = filepath.ToSlash(rel)
		full, ok := safeJoin(root, rel)
		if !ok {
			http.NotFound(w, r)
			return
		}
		switch strings.ToLower(filepath.Ext(full)) {
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		default:
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, full)
	}
}

func buildDocHomeSection(pub string, dataRoot string, activeCat string, page int) template.HTML {
	if docStore == nil {
		return ""
	}
	cats, _ := docStore.ListCategories()
	limit := docHomeLatestN()
	var b strings.Builder
	b.WriteString(`<div class="doc-home-grid">`)
	b.WriteString(buildDocNavHTML(cats, limit, activeCat, false))
	b.WriteString(buildDocFeedHTML(pub, dataRoot, activeCat, page))
	b.WriteString(buildDocLinkAsideHTML(pub, dataRoot))
	b.WriteString(`</div>`)
	b.WriteString(docModuleScript(pub, false))
	return template.HTML(b.String())
}

// buildDocNavHTML 分类 + 搜索；articlePage 为 true 时仅导航（无列表），分类点击回首页筛选。
func buildDocNavHTML(cats []string, limit int, activeCat string, articlePage bool) string {
	activeCat = strings.TrimSpace(activeCat)
	var b strings.Builder
	b.WriteString(`<section class="doc-nav" id="docNav" data-doc-limit="`)
	b.WriteString(strconv.Itoa(limit))
	b.WriteString(`"`)
	if articlePage {
		b.WriteString(` data-doc-article="1"`)
	}
	b.WriteString(`>`)
	b.WriteString(`<div class="doc-toolbar">`)
	b.WriteString(`<nav class="doc-cats" id="docCats" role="tablist" aria-label="分类">`)
	allOn := activeCat == ""
	b.WriteString(`<button type="button" class="doc-cat`)
	if allOn {
		b.WriteString(` is-active`)
	}
	b.WriteString(`" data-cat="" role="tab" aria-selected="`)
	if allOn {
		b.WriteString(`true`)
	} else {
		b.WriteString(`false`)
	}
	b.WriteString(`">全部</button>`)
	for _, c := range cats {
		on := c == activeCat
		b.WriteString(`<button type="button" class="doc-cat`)
		if on {
			b.WriteString(` is-active`)
		}
		b.WriteString(`" data-cat="`)
		b.WriteString(html.EscapeString(c))
		b.WriteString(`" role="tab" aria-selected="`)
		if on {
			b.WriteString(`true`)
		} else {
			b.WriteString(`false`)
		}
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(c))
		b.WriteString(`</button>`)
	}
	b.WriteString(`</nav>`)
	b.WriteString(`<div class="doc-search-row">`)
	b.WriteString(`<span class="doc-search-icon" aria-hidden="true">`)
	b.WriteString(docSearchIconSVG())
	b.WriteString(`</span>`)
	b.WriteString(`<input type="search" id="docSearchInput" class="doc-search-input" placeholder="搜索 slug、标题、标签…" autocomplete="off" aria-label="搜索文章">`)
	b.WriteString(`<ul class="doc-search-results" id="docSearchResults" aria-live="polite"></ul>`)
	b.WriteString(`</div></div></section>`)
	return b.String()
}

func buildDocFeedHTML(pub, dataRoot, activeCat string, page int) string {
	if docStore == nil {
		return `<section class="doc-feed" id="docFeed" aria-label="文章列表"><div class="doc-feed-panel"><ul class="doc-card-list" id="docCardList"></ul></div></section>`
	}
	limit := docHomeLatestN()
	payload, err := docHomePayload(dataRoot, activeCat, page, limit)
	if err != nil {
		payload = docHomePayloadJSON{Page: 1, PageSize: limit, TotalPages: 1}
	}
	var b strings.Builder
	b.WriteString(`<section class="doc-feed" id="docFeed" aria-label="文章列表">`)
	b.WriteString(`<div class="doc-feed-panel">`)
	b.WriteString(`<ul class="doc-card-list" id="docCardList" data-doc-ssr="1">`)
	writeDocHomeCards(&b, pub, payload.Items)
	b.WriteString(`</ul>`)
	b.WriteString(buildDocFeedPagerHTML(payload.Total, payload.Page, payload.PageSize, payload.TotalPages))
	b.WriteString(`</div></section>`)
	return b.String()
}

func buildDocFeedPagerHTML(total, page, pageSize, totalPages int) string {
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	prevDis := ""
	nextDis := ""
	if page <= 1 {
		prevDis = ` disabled`
	}
	if page >= totalPages {
		nextDis = ` disabled`
	}
	var b strings.Builder
	b.WriteString(`<div class="doc-feed-pager" id="docFeedPager">`)
	b.WriteString(`<span class="doc-feed-count" id="docFeedCount">共 `)
	b.WriteString(strconv.Itoa(total))
	b.WriteString(` 篇</span>`)
	b.WriteString(`<div class="doc-feed-pager-actions">`)
	b.WriteString(`<button type="button" class="doc-pager-btn" id="docPagerPrev"`)
	b.WriteString(prevDis)
	b.WriteString(`>上一页</button>`)
	b.WriteString(`<span class="doc-pager-info" id="docPagerInfo">第 `)
	b.WriteString(strconv.Itoa(page))
	b.WriteString(` / `)
	b.WriteString(strconv.Itoa(totalPages))
	b.WriteString(` 页</span>`)
	b.WriteString(`<button type="button" class="doc-pager-btn" id="docPagerNext"`)
	b.WriteString(nextDis)
	b.WriteString(`>下一页</button>`)
	b.WriteString(`</div></div>`)
	return b.String()
}

func writeDocHomeCards(b *strings.Builder, pub string, items []docHomeItemJSON) {
	for _, it := range items {
		href := pubPath(pub, it.Slug)
		date := ""
		if it.PublishedAt > 0 {
			date = time.Unix(it.PublishedAt, 0).Format("2006-01-02")
		}
		title := it.Title
		if title == "" {
			title = it.Slug
		}
		title = truncateDocCapsuleTitle(title)
		b.WriteString(`<li class="doc-capsule"><a class="doc-capsule-link" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`"><div class="doc-capsule-head"><span class="doc-capsule-title">`)
		b.WriteString(html.EscapeString(title))
		b.WriteString(`</span><span class="doc-capsule-head-meta">`)
		if it.Category != "" {
			b.WriteString(`<span class="doc-capsule-cat">`)
			b.WriteString(html.EscapeString(it.Category))
			b.WriteString(`</span>`)
		}
		b.WriteString(`<time class="doc-capsule-date" datetime="`)
		b.WriteString(html.EscapeString(date))
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(date))
		b.WriteString(`</time></span></div>`)
		if it.Excerpt != "" {
			b.WriteString(`<p class="doc-capsule-excerpt">`)
			b.WriteString(html.EscapeString(it.Excerpt))
			b.WriteString(`</p>`)
		}
		b.WriteString(`</a></li>`)
	}
}

func docSearchIconSVG() string {
	return `<svg viewBox="0 0 20 20" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="8.5" cy="8.5" r="5.5"/><path d="M13 13l4 4" stroke-linecap="round"/></svg>`
}

func formatDocMetaLine(m docindex.Meta) string {
	parts := []string{m.PublishedAt.Format("2006-01-02")}
	if m.Categories != "" {
		parts = append(parts, m.Categories)
	}
	if m.Views > 0 {
		parts = append(parts, formatInt64(m.Views)+" 次阅读")
	}
	return strings.Join(parts, " · ")
}

func formatDocAuthorLine(m docindex.Meta) string {
	author := strings.TrimSpace(m.Author)
	if author == "" {
		author = docindex.DefaultAuthor
	}
	return "作者 · " + author
}

func formatInt64(n int64) string {
	return strconvFormatInt(n)
}

func docModuleScript(pub string, articlePage bool) string {
	articleFlag := "0"
	if articlePage {
		articleFlag = "1"
	}
	base := pubPath(pub, "")
	if base == "/" {
		base = ""
	}
	return `<script>(function(){
var nav=document.getElementById("docNav");
var input=document.getElementById("docSearchInput");
var list=document.getElementById("docSearchResults");
var cardList=document.getElementById("docCardList");
var catNav=document.getElementById("docCats");
var pager=document.getElementById("docFeedPager");
var btnPrev=document.getElementById("docPagerPrev");
var btnNext=document.getElementById("docPagerNext");
var countEl=document.getElementById("docFeedCount");
var infoEl=document.getElementById("docPagerInfo");
if(!nav)return;
var onArticle=nav.getAttribute("data-doc-article")==="1"||"`+articleFlag+`"==="1";
var limit=parseInt(nav.getAttribute("data-doc-limit")||"5",10)||5;
var activeCat="";
var currentPage=1;
var timer=null;
function homeURL(){
  var root="` + base + `";
  var u=root?(root.replace(/\/$/,"")+"/"):"/";
  if(activeCat) u+="?category="+encodeURIComponent(activeCat);
  return u;
}
function esc(s){var d=document.createElement("div");d.textContent=s||"";return d.innerHTML;}
function apiRoot(){
  if(typeof window.__fsAbsRoot==="function") return window.__fsAbsRoot("");
  return (window.__fsPublicBase||"").replace(/\/$/,"");
}
function fmtDate(ts){
  if(!ts)return"";
  try{
    var d=new Date(ts*1000);
    var y=d.getFullYear(),m=("0"+(d.getMonth()+1)).slice(-2),day=("0"+d.getDate()).slice(-2);
    return y+"-"+m+"-"+day;
  }catch(e){return"";}
}
function truncTitle(s){
  s=s||"";
  var arr=Array.from(s);
  if(arr.length<=22)return s;
  return arr.slice(0,22).join("")+"…";
}
function syncTabs(){
  if(!catNav)return;
  catNav.querySelectorAll(".doc-cat").forEach(function(b){
    var c=b.getAttribute("data-cat")||"";
    var on=(c===activeCat);
    b.classList.toggle("is-active",on);
    b.setAttribute("aria-selected",on?"true":"false");
  });
}
function syncURL(){
  var u=new URL(window.location.href);
  if(activeCat) u.searchParams.set("category",activeCat); else u.searchParams.delete("category");
  if(currentPage>1) u.searchParams.set("page",String(currentPage)); else u.searchParams.delete("page");
  history.replaceState(null,"",u.pathname+u.search);
}
function readURL(){
  var u=new URL(window.location.href);
  activeCat=u.searchParams.get("category")||"";
  currentPage=parseInt(u.searchParams.get("page")||"1",10)||1;
  syncTabs();
}
function renderCards(items){
  if(!cardList)return;
  if(!items||!items.length){cardList.innerHTML='<li class="doc-capsule doc-capsule-empty">暂无文章</li>';return;}
  cardList.innerHTML=items.map(function(it){
    var href="` + base + `"+(it.slug||"");
    var title=esc(truncTitle(it.title||it.slug));
    var date=esc(fmtDate(it.publishedAt));
    var cat=esc(it.category||"");
    var catHtml=cat?'<span class="doc-capsule-cat">'+cat+"</span>":"";
    var ex=it.excerpt?'<p class="doc-capsule-excerpt">'+esc(it.excerpt)+"</p>":"";
    return '<li class="doc-capsule"><a class="doc-capsule-link" href="'+href+'"><div class="doc-capsule-head"><span class="doc-capsule-title">'+title+'</span><span class="doc-capsule-head-meta">'+catHtml+'<time class="doc-capsule-date">'+date+'</time></span></div>'+ex+'</a></li>';
  }).join("");
}
function renderPager(data){
  if(!pager)return;
  var total=data.total||0,page=data.page||1,tp=data.totalPages||1;
  if(countEl) countEl.textContent="共 "+total+" 篇";
  if(infoEl) infoEl.textContent="第 "+page+" / "+tp+" 页";
  if(btnPrev) btnPrev.disabled=(page<=1);
  if(btnNext) btnNext.disabled=(page>=tp);
}
function clearSSR(){
  if(cardList) cardList.removeAttribute("data-doc-ssr");
  ssrFeed=false;
}
function loadCards(page){
  if(!cardList)return;
  if(page) currentPage=page;
  var url=apiRoot()+"/api/doc/home?limit="+limit+"&page="+currentPage;
  if(activeCat) url+="&category="+encodeURIComponent(activeCat);
  fetch(url).then(function(r){if(!r.ok)throw new Error("http");return r.json();}).then(function(data){
    if(!data||!data.items){renderCards([]);renderPager({total:0,page:1,totalPages:1});clearSSR();syncURL();return;}
    currentPage=data.page||currentPage;
    renderCards(data.items);
    renderPager(data);
    clearSSR();
    syncURL();
  }).catch(function(){
    if(ssrFeed)return;
    renderCards([]);
    renderPager({total:0,page:1,totalPages:1});
  });
}
var ssrFeed=cardList&&cardList.getAttribute("data-doc-ssr")==="1";
readURL();
if(cardList&&!ssrFeed) loadCards(currentPage);
if(catNav){
  catNav.addEventListener("click",function(e){
    var btn=e.target.closest(".doc-cat");
    if(!btn)return;
    activeCat=btn.getAttribute("data-cat")||"";
    syncTabs();
    if(onArticle){if(window.__fsMp3FlushHandoff) window.__fsMp3FlushHandoff(); location.href=homeURL();return;}
    clearSSR();
    loadCards(1);
  });
}
if(btnPrev) btnPrev.addEventListener("click",function(){if(currentPage>1){clearSSR();loadCards(currentPage-1);}});
if(btnNext) btnNext.addEventListener("click",function(){clearSSR();loadCards(currentPage+1);});
if(input&&list){
  var searchRow=input.closest(".doc-search-row");
  function closeSearch(){
    list.innerHTML="";
    list.classList.remove("is-open");
  }
  function renderSearch(items){
    var q=(input.value||"").trim();
    if(!q.length||!items||!items.length){closeSearch();return;}
    list.classList.add("is-open");
    list.innerHTML=items.map(function(it){
      var href="` + base + `"+(it.slug||"");
      return "<li><a href=\""+href+"\"><strong>"+esc(it.title||it.slug)+"</strong><span>"+esc((it.categories||"")+" "+(it.tags||""))+"</span></a></li>";
    }).join("");
  }
  function goSearch(){
    var q=(input.value||"").trim();
    if(q.length<1){closeSearch();return;}
    fetch(apiRoot()+"/api/doc/search?q="+encodeURIComponent(q))
      .then(function(r){return r.json();})
      .then(renderSearch)
      .catch(function(){closeSearch();});
  }
  input.addEventListener("input",function(){
    clearTimeout(timer);
    timer=setTimeout(goSearch,280);
  });
  input.addEventListener("keydown",function(e){
    if(e.key==="Escape") closeSearch();
  });
  document.addEventListener("click",function(e){
    if(!searchRow||searchRow.contains(e.target))return;
    closeSearch();
  });
}
})();</script>`
}

func tryServeDocArticle(w http.ResponseWriter, r *http.Request, pub, slug, dataRoot string) bool {
	if docStore == nil || docReservedSlug(slug) || !docindex.ValidSlug(slug) {
		return false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return true
	}
	meta, err := docStore.Get(slug)
	if err != nil {
		return false
	}
	if isSecretAuthSlug(dataRoot, slug) && authFeatureEnabled() {
		serveSecretAuthArticleGate(w, r, pub, slug, dataRoot, meta)
		return true
	}
	mdMtime := articleMdMtimeUnix(slug)
	etag := docArticleETag(slug, mdMtime, linkConfMtimeUnix(dataRoot))
	if inm := strings.TrimSpace(r.Header.Get("If-None-Match")); inm != "" && inm == etag {
		issueMp3GateCookie(w, r)
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", docArticleCacheControl())
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	body, err := docStore.ReadBody(slug)
	if err != nil {
		return false
	}
	views, _ := recordDocView(w, r, slug)

	htmlBody, err := renderDocArticleHTML(pub, slug, body)
	if err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
		return true
	}

	title := html.EscapeString(meta.Title)
	metaForLine := meta
	metaForLine.Views = views
	metaLine := html.EscapeString(formatDocMetaLine(metaForLine))
	if meta.Tags != "" {
		metaLine += " · " + html.EscapeString(meta.Tags)
	}
	cats, _ := docStore.ListCategories()
	limit := docHomeLatestN()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", docArticleCacheControl())
	issueMp3GateCookie(w, r)
	out := applySitePage(string(articlePageTpl), pub, sitePageLayout{
		HeadCSS:            siteSubpageHeadCSS(),
		MP3:                true,
		MP3CSS:             false,
		MP3ScriptsOverride: siteMp3ArticleInitScripts(pub),
	}, map[string]string{
		"{{DOC_NAV}}":            buildDocNavHTMLCached(cats, limit, "", true),
		"{{DOC_LINK_ASIDE}}":     buildDocLinkAsideHTMLCached(pub, dataRoot),
		"{{DOC_MODULE_JS}}":      docModuleScript(pub, true),
		"{{DOC_ARTICLE_ASSETS}}": docArticleExternalCSSLink(pub) + docArticleExternalMp3CSSLink(pub) + docArticleBodyScripts(pub, htmlBody),
		"{{ARTICLE_TITLE}}":      title,
		"{{ARTICLE_META}}":       metaLine,
		"{{ARTICLE_AUTHOR}}":     html.EscapeString(formatDocAuthorLine(meta)),
		"{{ARTICLE_BODY}}":       `<div class="md doc-md">` + string(htmlBody) + `</div>`,
		"{{ARTICLE_CSS}}":        "",
	}, visitorGeoLabel(r))
	if r.Method == http.MethodHead {
		return true
	}
	_, _ = w.Write(out)
	return true
}

func serveSecretAuthArticleGate(w http.ResponseWriter, r *http.Request, pub, slug, dataRoot string, meta docindex.Meta) {
	cats, _ := docStore.ListCategories()
	limit := docHomeLatestN()
	title := html.EscapeString(meta.Title)
	assets := docArticleExternalCSSLink(pub) + docArticleExternalMp3CSSLink(pub) +
		docArticleExternalScriptTag(pub) + igoAuthScriptTag()
	if bodyPeek, err := docStore.ReadBody(slug); err == nil {
		lower := bytes.ToLower(bodyPeek)
		if bytes.Contains(lower, []byte("```mermaid")) {
			assets += docMermaidScriptTags(pub)
		}
	}
	assets += secretAuthArticleBootScript(pub, slug, meta.Title)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	issueMp3GateCookie(w, r)
	out := applySitePage(string(articlePageTpl), pub, sitePageLayout{
		HeadCSS:            siteSubpageHeadCSS(),
		MP3:                true,
		MP3CSS:             false,
		MP3ScriptsOverride: siteMp3ArticleInitScripts(pub),
	}, map[string]string{
		"{{DOC_NAV}}":            buildDocNavHTMLCached(cats, limit, "", true),
		"{{DOC_LINK_ASIDE}}":     buildDocLinkAsideHTMLCached(pub, dataRoot),
		"{{DOC_MODULE_JS}}":      docModuleScript(pub, true),
		"{{DOC_ARTICLE_ASSETS}}": assets,
		"{{ARTICLE_TITLE}}":      title,
		"{{ARTICLE_META}}":       "需通过 IGO密钥认证 后阅读",
		"{{ARTICLE_AUTHOR}}":     "",
		"{{ARTICLE_BODY}}":       secretAuthArticleGateBodyHTML(),
		"{{ARTICLE_CSS}}":        igoAuthCSS,
	}, visitorGeoLabel(r))
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(out)
}

func secretAuthArticleGateBodyHTML() string {
	return igoAuthPanelHTML + `
<div id="igoMainSec" class="hidden">
<p class="igo-authed-hint">已验证，可阅读本文。<span class="igo-auth-expiry" id="igoMainKeyExpiry" aria-live="polite"></span></p>
<div id="secretAuthMd" class="md doc-md"></div>
</div>`
}

func secretAuthArticleBootScript(pub, slug, gateTitle string) string {
	baseJS, err := json.Marshal(pub)
	if err != nil {
		baseJS = []byte(`""`)
	}
	slugJS, err := json.Marshal(slug)
	if err != nil {
		slugJS = []byte(`""`)
	}
	titleJS, err := json.Marshal(gateTitle)
	if err != nil {
		titleJS = []byte(`""`)
	}
	return `<script>(function(){
  var apiBase=` + string(baseJS) + `;
  var slug=` + string(slugJS) + `;
  var gateTitle=` + string(titleJS) + `;
  var loaded=false;
  function api(p){ return apiBase+p; }
  function setMeta(title, meta, author){
    var t=document.getElementById("docArticleTitle");
    var m=document.getElementById("docArticleMeta");
    var a=document.getElementById("docArticleAuthor");
    if(t&&title) t.textContent=title;
    if(m) m.textContent=meta||"";
    if(a) a.textContent=author||"";
    if(title) document.title=title;
  }
  function showLoadErr(msg){
    var md=document.getElementById("secretAuthMd");
    if(!md) return;
    md.textContent="";
    var p=document.createElement("p");
    p.className="igo-auth-msg is-err";
    p.textContent=msg||"加载失败";
    md.appendChild(p);
  }
  function loadArticle(auth){
    if(loaded||!auth||!auth.requireAuth()) return;
    fetch(api("/api/doc/secret-article?slug="+encodeURIComponent(slug)),{
      credentials:"same-origin",
      headers:auth.authHeaders({"Accept":"application/json"})
    }).then(function(r){
      return r.json().then(function(j){ return {r:r,j:j}; }).catch(function(){ return {r:r,j:null}; });
    }).then(function(x){
      if(auth.handleAuthFailure(x.r.status,x.j)) return;
      if(!x.r.ok||!x.j||!x.j.ok){
        showLoadErr((x.j&&x.j.error)||"加载失败");
        return;
      }
      loaded=true;
      setMeta(x.j.title||"", x.j.meta||"", x.j.author||"");
      var md=document.getElementById("secretAuthMd");
      if(md){
        md.innerHTML=x.j.bodyHTML||"";
        if(window.__fsFixFragmentHrefs) try{ window.__fsFixFragmentHrefs(md); }catch(e){}
        if(window.initDocCodeChrome) try{ window.initDocCodeChrome(md); }catch(e){}
        if(window.initDocMermaid) try{ window.initDocMermaid(md); }catch(e){}
      }
    }).catch(function(){ showLoadErr("请求失败"); });
  }
  function boot(){
    var auth=window.__igoAuthInit({
      apiBase: apiBase,
      onReady: function(){ loadArticle(auth); },
      onLogout: function(){
        loaded=false;
        setMeta(gateTitle, "需通过 IGO密钥认证 后阅读", "");
        var md=document.getElementById("secretAuthMd");
        if(md) md.innerHTML="";
      }
    });
  }
  if(document.readyState==="loading") document.addEventListener("DOMContentLoaded", boot);
  else boot();
})();</script>`
}

func docSecretArticleHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !requireAuthToken(w, r, dataRoot) {
		return
	}
	slug := strings.TrimSpace(r.URL.Query().Get("slug"))
	if !docindex.ValidSlug(slug) || docReservedSlug(slug) {
		jsonError(w, http.StatusBadRequest, "bad slug")
		return
	}
	if !isSecretAuthSlug(dataRoot, slug) {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	if docStore == nil {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	meta, err := docStore.Get(slug)
	if err != nil {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	body, err := docStore.ReadBody(slug)
	if err != nil {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
	htmlBody, err := renderDocArticleHTML(pub, slug, body)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "render failed")
		return
	}
	views, _ := recordDocView(w, r, slug)
	metaForLine := meta
	metaForLine.Views = views
	metaLine := formatDocMetaLine(metaForLine)
	if meta.Tags != "" {
		metaLine += " · " + meta.Tags
	}
	jsonWrite(w, map[string]any{
		"ok":       true,
		"slug":     slug,
		"title":    meta.Title,
		"meta":     metaLine,
		"author":   formatDocAuthorLine(meta),
		"bodyHTML": string(htmlBody),
	})
}

func docModuleCSS() string {
	return `
.doc-home-grid{position:relative;margin-bottom:0;display:grid;grid-template-columns:1fr}
.doc-home-grid .doc-nav{grid-column:1;grid-row:1;margin:0 0 1rem}
.doc-home-grid .doc-feed,.doc-home-grid .doc-article-block{grid-column:1;grid-row:2;margin:0 0 1rem}
.doc-link-aside{grid-column:1;grid-row:1/3;justify-self:end;align-self:start;position:sticky;top:1.25rem;width:9rem;max-width:10.5rem;min-width:6.5rem;transform:translateX(calc(100% + .65rem));display:flex;flex-direction:column;gap:.65rem;z-index:3}
.doc-link-capsule{background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:.55rem .65rem;box-shadow:0 1px 3px rgba(45,59,45,.06)}
.doc-link-capsule-title{margin:0 0 .4rem;font-size:.78rem;font-weight:600;color:var(--muted);letter-spacing:.02em}
.doc-link-list{list-style:none;margin:0;padding:0;display:flex;flex-direction:column;gap:.3rem}
.doc-link-list a{display:block;font-size:.78rem;line-height:1.35;color:var(--accent);text-decoration:none;padding:.15rem 0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.doc-link-list a:hover,.doc-link-list a:focus{text-decoration:none;animation:doc-link-breathe 1.4s ease-in-out infinite}
@keyframes doc-link-breathe{0%,100%{opacity:1}50%{opacity:.5}}
@media(max-width:42rem){
.doc-home-grid{display:block}
.doc-home-grid .doc-nav,.doc-home-grid .doc-feed,.doc-home-grid .doc-article-block{grid-column:auto;grid-row:auto}
.doc-link-aside{grid-column:auto;grid-row:auto;justify-self:auto;position:static;transform:none;width:100%;max-width:none;margin:1rem 0 0}
}
.doc-nav,.doc-feed{margin:0 0 1rem;padding:0;background:transparent;border:0;box-shadow:none}
.doc-nav .doc-toolbar{background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:.65rem .85rem .75rem;box-shadow:0 1px 3px rgba(45,59,45,.06)}
.doc-cats{display:flex;flex-wrap:nowrap;gap:.15rem;overflow-x:auto;margin:0 0 .6rem;padding:0 0 .15rem;-webkit-overflow-scrolling:touch;scrollbar-width:thin}
.doc-cats::-webkit-scrollbar{height:4px}
.doc-cat{flex:0 0 auto;border:0;background:transparent;color:var(--muted);font:inherit;font-size:.84rem;padding:.35rem .55rem .4rem;cursor:pointer;border-bottom:2px solid transparent;border-radius:0;white-space:nowrap}
.doc-cat:hover{color:var(--text)}
.doc-cat.is-active{color:var(--accent);border-bottom-color:var(--accent);font-weight:600}
.doc-search-row{position:relative;margin:0}
.doc-search-icon{position:absolute;left:.55rem;top:50%;transform:translateY(-50%);color:var(--muted);pointer-events:none;display:flex}
.doc-search-input{width:100%;padding:.5rem .65rem .5rem 2.05rem;border:1px solid var(--line);border-radius:10px;background:var(--doc-surface);font:inherit;font-size:.88rem}
.doc-search-input:focus{outline:none;border-color:var(--accent);box-shadow:0 0 0 2px rgba(61,107,74,.15)}
.doc-search-results{position:absolute;left:0;right:0;top:100%;z-index:20;margin:.25rem 0 0;padding:.2rem 0;list-style:none;border:1px solid var(--line);border-radius:10px;max-height:14rem;overflow:auto;box-shadow:0 4px 12px rgba(45,59,45,.12);background:var(--doc-search-drop);pointer-events:none;opacity:0;visibility:hidden;transition:opacity .15s ease,visibility .15s}
.doc-search-results.is-open{opacity:1;visibility:visible;pointer-events:auto}
.doc-search-results li{border-bottom:1px solid rgba(184,201,184,.45)}
.doc-search-results li:last-child{border-bottom:0}
.doc-search-results a{display:block;padding:.45rem .6rem;text-decoration:none;color:var(--text);font-size:.88rem;background:transparent}
.doc-search-results a:hover{background:rgba(61,107,74,.1)}
.doc-search-results strong{display:block}
.doc-search-results span{display:block;font-size:.78rem;color:var(--muted)}
.doc-feed{margin-bottom:1.25rem}
.doc-feed-panel{background:var(--panel);border:1px solid var(--line);border-radius:14px;padding:1rem 1.1rem 1rem;box-shadow:0 1px 4px rgba(45,59,45,.07)}
.doc-card-list{list-style:none;margin:0;padding:0;display:flex;flex-direction:column;gap:.75rem}
.doc-capsule{background:var(--doc-surface);border:1px solid var(--line);border-radius:12px;padding:.85rem 1rem .9rem;transition:border-color .15s,box-shadow .15s,transform .15s}
.doc-capsule:hover{border-color:rgba(61,107,74,.45);box-shadow:0 2px 10px rgba(61,107,74,.1);transform:translateY(-1px)}
.doc-capsule-empty{color:var(--muted);font-size:.88rem;background:transparent;border:0;box-shadow:none;transform:none}
.doc-feed-pager{display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:.65rem;margin-top:1rem;padding-top:.85rem;border-top:1px solid var(--line)}
.doc-feed-count{font-size:.84rem;color:var(--muted)}
.doc-feed-pager-actions{display:flex;flex-wrap:wrap;align-items:center;gap:.5rem}
.doc-pager-info{font-size:.84rem;color:var(--muted);min-width:5.5rem;text-align:center}
.doc-pager-btn{border:1px solid var(--line);background:var(--doc-surface);color:var(--accent);font:inherit;font-size:.84rem;padding:.32rem .75rem;border-radius:8px;cursor:pointer}
.doc-pager-btn:hover:not(:disabled){background:rgba(61,107,74,.08)}
.doc-pager-btn:disabled{opacity:.45;cursor:not-allowed;color:var(--muted)}
.doc-capsule-link{display:block;text-decoration:none;color:var(--text);border-radius:10px}
.doc-capsule-link:hover .doc-capsule-title{color:var(--accent)}
.doc-capsule-link:focus-visible{outline:2px solid rgba(61,107,74,.35);outline-offset:2px}
.doc-capsule-head{display:flex;align-items:baseline;gap:.65rem;margin-bottom:.2rem}
.doc-capsule-title{font-weight:600;font-size:.94rem;line-height:1.4;flex:0 1 auto;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.doc-capsule-head-meta{display:flex;align-items:baseline;gap:.55rem;flex:0 0 auto;margin-left:auto}
.doc-capsule-cat{font-size:.76rem;color:var(--muted);white-space:nowrap}
.doc-capsule-date{font-size:.76rem;color:var(--muted);white-space:nowrap;font-variant-numeric:tabular-nums}
.doc-capsule-excerpt{margin:.1rem 0 0;font-size:.84rem;line-height:1.62;color:var(--muted);display:-webkit-box;-webkit-line-clamp:4;-webkit-box-orient:vertical;overflow:hidden;text-wrap:pretty;word-break:break-word}
`
}

// docThemeVars 文档模块用色（与全站 --bg/--panel 同系，避免刺眼纯白）。
func docThemeVars() string {
	return `
:root{
  --bg:#e8efe8;
  --panel:#dde8dd;
  --text:#2d3b2d;
  --muted:#5a6b5a;
  --line:#b8c9b8;
  --accent:#3d6b4a;
  --doc-surface:#d8e4d8;
  --doc-search-drop:rgba(216,228,216,.88);
}
`
}

// docHomeCSS 供首页样式表引用（与 docModuleCSS 相同）。
func docHomeCSS() string { return docThemeVars() + docModuleCSS() }

// filePreviewPageCSS 文件预览页（/preview/）：布局与文章同系，代码块复用 docCodeBlockCSS。
func filePreviewPageCSS() string {
	return docThemeVars() + `
body{font-family:system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;background:var(--bg);color:var(--text);margin:0;padding:1rem 1rem 1.5rem;line-height:1.6}
.wrap{max-width:52rem;margin:0 auto;background:var(--panel);border:1px solid var(--line);border-radius:14px;padding:1rem 1.2rem 1.15rem;box-shadow:0 1px 4px rgba(45,59,45,.07)}
.preview-fname{font-size:.82rem;margin:0 0 .85rem;color:var(--muted);font-weight:500;word-break:break-all}
.doc-md{font-size:.95rem;line-height:1.62;overflow-wrap:break-word;word-wrap:break-word;word-break:break-word;max-width:100%}
` + docCodeBlockCSS() + `
.doc-md img{max-width:100%;height:auto;border-radius:6px}
.doc-md table{border-collapse:collapse;width:100%;max-width:100%;display:block;overflow-x:auto}
.doc-md th,.doc-md td{border:1px solid var(--line);padding:.35rem .5rem;word-break:break-word}
.doc-md blockquote{margin:.75rem 0;padding:.4rem .75rem;border-left:3px solid var(--line);background:var(--doc-surface);color:var(--muted)}
.doc-md p,.doc-md li,.doc-md ul,.doc-md ol{overflow-wrap:break-word}
`
}

func docArticlePageCSS() string {
	return docThemeVars() + `
.doc-page-shell{padding-bottom:1.5rem}
.doc-article-block{margin-bottom:1rem}
.doc-article{background:var(--panel);border:1px solid var(--line);border-radius:14px;padding:1.05rem 1.15rem;box-shadow:0 1px 4px rgba(45,59,45,.07);overflow-x:hidden;max-width:100%}
.doc-article .doc-article-head{margin-bottom:1rem;padding-bottom:.65rem;border-bottom:1px solid var(--line)}
.doc-article h1{font-size:1.25rem;margin:0 0 .35rem;line-height:1.35;overflow-wrap:break-word}
.doc-article-meta{font-size:.85rem;color:var(--muted);margin:.2rem 0;overflow-wrap:break-word}
.doc-article-foot{margin-top:1.25rem;padding-top:.65rem;border-top:1px solid var(--line)}
.doc-article-foot .doc-article-meta{margin:0}
.doc-md{font-size:.95rem;line-height:1.6;overflow-wrap:break-word;word-wrap:break-word;word-break:break-word;max-width:100%}
` + docCodeBlockCSS() + `
.doc-md blockquote{margin:.75rem 0;padding:.35rem .75rem;border-left:3px solid var(--line);background:var(--doc-surface);color:var(--muted);overflow-wrap:break-word}
.doc-md img{max-width:100%;height:auto}
.doc-md table{border-collapse:collapse;width:100%;max-width:100%;display:block;overflow-x:auto}
.doc-md th,.doc-md td{border:1px solid var(--line);padding:.35rem .5rem;word-break:break-word}
.doc-md p,.doc-md li,.doc-md ul,.doc-md ol{overflow-wrap:break-word}
`
}

// strconvFormatInt avoids importing strconv in template helpers only used here.
func strconvFormatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
