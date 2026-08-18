package main

import (
	"hash/fnv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/igokorea/igozhang/internal/markdown"
)

const (
	docChromeCacheTTL = 60 * time.Second
	// 文章页浏览器新鲜窗口：与侧栏/导航 chrome 服务端 TTL 对齐；过短则每次导航都回源，过长则改 link.conf 等嵌入内容迟迟不更新。
	docArticleBrowserMaxAgeSec = 60
	docArticleCacheMax         = 256
	// Bump when article HTML shape changes independently of md bytes / static assets
	// (heading-id algorithm, fragment href rewrite vs <base>, StripLeadingH1, etc.).
	docHTMLRenderRev = "heading-id-unicode-v2-fraghref"
)

func docArticleCacheControl() string {
	return fmt.Sprintf("private, max-age=%d", docArticleBrowserMaxAgeSec)
}

type docRenderEntry struct {
	bodyHash uint64
	html     []byte
}

var docRenderCache struct {
	mu    sync.RWMutex
	bySlug map[string]docRenderEntry
	order  []string
}

var docNavCache struct {
	mu       sync.RWMutex
	html     string
	key      string
	expires  time.Time
}

var docLinkAsideCache struct {
	mu      sync.RWMutex
	html    string
	key     string
	expires time.Time
}

func docBodyHash(body []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(body)
	return h.Sum64()
}

func docRenderCacheGet(slug string, body []byte) ([]byte, bool) {
	hash := docBodyHash(body)
	docRenderCache.mu.RLock()
	e, ok := docRenderCache.bySlug[slug]
	docRenderCache.mu.RUnlock()
	if !ok || e.bodyHash != hash || len(e.html) == 0 {
		return nil, false
	}
	dup := make([]byte, len(e.html))
	copy(dup, e.html)
	return dup, true
}

func docRenderCacheSet(slug string, body, html []byte) {
	if slug == "" || len(html) == 0 {
		return
	}
	dup := make([]byte, len(html))
	copy(dup, html)
	hash := docBodyHash(body)

	docRenderCache.mu.Lock()
	defer docRenderCache.mu.Unlock()
	if docRenderCache.bySlug == nil {
		docRenderCache.bySlug = make(map[string]docRenderEntry)
	}
	if _, had := docRenderCache.bySlug[slug]; !had {
		docRenderCache.order = append(docRenderCache.order, slug)
	}
	docRenderCache.bySlug[slug] = docRenderEntry{bodyHash: hash, html: dup}
	for len(docRenderCache.order) > docArticleCacheMax {
		old := docRenderCache.order[0]
		docRenderCache.order = docRenderCache.order[1:]
		delete(docRenderCache.bySlug, old)
	}
}

func invalidateDocRenderCache(slug string) {
	docRenderCache.mu.Lock()
	defer docRenderCache.mu.Unlock()
	delete(docRenderCache.bySlug, slug)
	for i, s := range docRenderCache.order {
		if s == slug {
			docRenderCache.order = append(docRenderCache.order[:i], docRenderCache.order[i+1:]...)
			break
		}
	}
}

func invalidateAllDocRenderCache() {
	docRenderCache.mu.Lock()
	docRenderCache.bySlug = nil
	docRenderCache.order = nil
	docRenderCache.mu.Unlock()
}

func invalidateDocChromeCache() {
	docNavCache.mu.Lock()
	docNavCache.expires = time.Time{}
	docNavCache.html = ""
	docNavCache.key = ""
	docNavCache.mu.Unlock()

	docLinkAsideCache.mu.Lock()
	docLinkAsideCache.expires = time.Time{}
	docLinkAsideCache.html = ""
	docLinkAsideCache.key = ""
	docLinkAsideCache.mu.Unlock()
}

func noteDocArticleChanged(slug, dataRoot string) {
	if slug != "" {
		invalidateDocRenderCache(slug)
	} else {
		invalidateAllDocRenderCache()
	}
	invalidateDocChromeCache()
	if dataRoot != "" {
		scheduleDocBackup(dataRoot)
	}
}

func articleMdMtimeUnix(slug string) int64 {
	if docStore == nil {
		return 0
	}
	st, err := os.Stat(filepath.Join(docStore.AssetsDir(), "md", slug+".md"))
	if err != nil {
		return 0
	}
	return st.ModTime().Unix()
}

func linkConfMtimeUnix(dataRoot string) int64 {
	if dataRoot == "" {
		return 0
	}
	st, err := os.Stat(linkConfPath(dataRoot))
	if err != nil {
		return 0
	}
	return st.ModTime().Unix()
}

func docArticleETag(slug string, mdMtime, linkConfMtime int64) string {
	h := fnv.New32a()
	// linkConfMtime：文章 HTML 嵌入侧栏胶囊，改 link.conf 须打穿浏览器 304。
	// docHTMLRenderRev busts browser 304 when Markdown→HTML rules change (e.g. heading ids)
	// without touching article md mtime or static CSS/JS hashes.
	_, _ = fmt.Fprintf(h, "%s:%d:%d:%s:%s", slug, mdMtime, linkConfMtime, docStaticAssetVersion(), docHTMLRenderRev)
	return fmt.Sprintf(`"%x"`, h.Sum32())
}

func renderDocArticleHTML(pub, slug string, body []byte) ([]byte, error) {
	pagePath := pubPath(pub, slug)
	var html []byte
	if cached, ok := docRenderCacheGet(slug, body); ok {
		html = cached
	} else {
		src := markdown.StripLeadingH1(body)
		rendered, err := markdown.Render(src, markdown.RenderOpts{
			PicBaseURL: docPicPublicBase(pub) + "/",
		})
		if err != nil {
			return nil, err
		}
		// Cache Goldmark output without path-pinning so BASE_PATH variants stay correct.
		docRenderCacheSet(slug, body, rendered)
		html = rendered
	}
	// <base href="/"> would turn href="#id" into origin/#id (home). Pin to article path.
	return markdown.RewriteFragmentHrefs(html, pagePath), nil
}

func buildDocNavHTMLCached(cats []string, limit int, activeCat string, articlePage bool) string {
	activeCat = strings.TrimSpace(activeCat)
	key := fmt.Sprintf("%d|%t|%s|%v", limit, articlePage, activeCat, cats)
	now := time.Now()

	docNavCache.mu.RLock()
	if now.Before(docNavCache.expires) && docNavCache.key == key && docNavCache.html != "" {
		html := docNavCache.html
		docNavCache.mu.RUnlock()
		return html
	}
	docNavCache.mu.RUnlock()

	html := buildDocNavHTML(cats, limit, activeCat, articlePage)
	docNavCache.mu.Lock()
	docNavCache.html = html
	docNavCache.key = key
	docNavCache.expires = now.Add(docChromeCacheTTL)
	docNavCache.mu.Unlock()
	return html
}

func buildDocLinkAsideHTMLCached(pub, dataRoot string) string {
	key := pub + "|" + dataRoot
	now := time.Now()

	docLinkAsideCache.mu.RLock()
	if now.Before(docLinkAsideCache.expires) && docLinkAsideCache.key == key && docLinkAsideCache.html != "" {
		html := docLinkAsideCache.html
		docLinkAsideCache.mu.RUnlock()
		return html
	}
	docLinkAsideCache.mu.RUnlock()

	html := buildDocLinkAsideHTML(pub, dataRoot)
	docLinkAsideCache.mu.Lock()
	docLinkAsideCache.html = html
	docLinkAsideCache.key = key
	docLinkAsideCache.expires = now.Add(docChromeCacheTTL)
	docLinkAsideCache.mu.Unlock()
	return html
}
