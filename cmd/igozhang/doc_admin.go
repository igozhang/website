package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
	"github.com/igokorea/igozhang/internal/markdown"
)

//go:embed igoadmin.html
var igoadminHTML []byte

//go:embed igoadmin.css
var igoadminCSS []byte

//go:embed igoadmin.js
var igoadminJS []byte

const (
	docPasteImageMaxBytes = 200 * 1024
	igoadminArticleMaxBody = 8 << 20
)

type igoadminArticleBody struct {
	Slug        string          `json:"slug"`
	OldSlug     string          `json:"oldSlug"`
	Title       string          `json:"title"`
	Categories  string          `json:"categories"`
	Tags        string          `json:"tags"`
	Author      string          `json:"author"`
	PublishedAt string          `json:"publishedAt"`
	Markdown    string          `json:"markdown"`
	DraftImages []draftImageIn  `json:"draftImages"`
	AutoSave    bool            `json:"autoSave"`
}

func prepareArticleMarkdown(dataRoot, slug, md string, drafts []draftImageIn, oldMD string, isUpdate bool) (string, error) {
	assets := docAssetsRoot(dataRoot)
	picDir := filepath.Join(assets, "doc-pic")
	newMD, err := materializeDraftImages(picDir, slug, md, drafts)
	if err != nil {
		return "", err
	}
	if isUpdate {
		if err := syncArticleImages(assets, slug, oldMD, newMD); err != nil {
			return "", err
		}
	}
	return newMD, nil
}

func registerIgoAdminHandlers(mux *http.ServeMux, dataRoot string) {
	mux.HandleFunc("/igoadmin", func(w http.ResponseWriter, r *http.Request) {
		igoadminPageHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/igoadmin/", func(w http.ResponseWriter, r *http.Request) {
		igoadminPageHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/articles", func(w http.ResponseWriter, r *http.Request) {
		igoadminArticlesHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/articles/batch", func(w http.ResponseWriter, r *http.Request) {
		igoadminArticlesBatchHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/preview", func(w http.ResponseWriter, r *http.Request) {
		igoadminPreviewHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/article", func(w http.ResponseWriter, r *http.Request) {
		igoadminArticleHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/categories", func(w http.ResponseWriter, r *http.Request) {
		igoadminCategoriesHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/category", func(w http.ResponseWriter, r *http.Request) {
		igoadminCategoryHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/upload-md", func(w http.ResponseWriter, r *http.Request) {
		igoadminUploadMDHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/export", func(w http.ResponseWriter, r *http.Request) {
		igoadminExportHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/api/igoadmin/import", func(w http.ResponseWriter, r *http.Request) {
		igoadminImportHandler(w, r, dataRoot)
	})
	mux.HandleFunc("/igoadmin/igoadmin.css", func(w http.ResponseWriter, r *http.Request) {
		serveIgoAdminAsset(w, r, igoadminCSS, "text/css; charset=utf-8")
	})
	mux.HandleFunc("/igoadmin/igoadmin.js", func(w http.ResponseWriter, r *http.Request) {
		serveIgoAdminAsset(w, r, igoadminJS, "application/javascript; charset=utf-8")
	})
}

func serveIgoAdminAsset(w http.ResponseWriter, r *http.Request, body []byte, contentType string) {
	if !authFeatureEnabled() {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(body)
}

func igoadminAssetQuery(body []byte) string {
	h := fnv.New32a()
	_, _ = h.Write(body)
	return fmt.Sprintf("?v=%x", h.Sum32())
}

func igoadminRequireToken(w http.ResponseWriter, r *http.Request, dataRoot string) bool {
	return requireAuthToken(w, r, dataRoot)
}

func igoadminPageHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !authFeatureEnabled() {
		http.NotFound(w, r)
		return
	}
	pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
	if pub == "" {
		if p := strings.TrimSpace(getenv("BASE_PATH", "")); p != "" {
			pub = strings.TrimSuffix(p, "/")
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(renderIgoAdminHTML(pub, dataRoot, r))
}

func renderIgoAdminHTML(pub, dataRoot string, r *http.Request) []byte {
	apiPub := strings.TrimSuffix(pub, "/")
	baseJS, err := json.Marshal(apiPub)
	if err != nil {
		baseJS = []byte(`""`)
	}
	cssURL := apiPub + "/igoadmin/igoadmin.css" + igoadminAssetQuery(igoadminCSS)
	jsURL := apiPub + "/igoadmin/igoadmin.js" + igoadminAssetQuery(igoadminJS)
	libURL, err := json.Marshal(docMermaidLibJSAssetURL(apiPub))
	if err != nil {
		libURL = []byte(`""`)
	}
	thinURL, err := json.Marshal(docMermaidJSAssetURL(apiPub))
	if err != nil {
		thinURL = []byte(`""`)
	}
	return applySitePage(string(igoadminHTML), pub, sitePageLayout{}, map[string]string{
		"{{IGOADMIN_BASE}}":    string(baseJS),
		"{{IGOADMIN_CSS_URL}}": cssURL,
		"{{IGOADMIN_JS_URL}}":  jsURL,
		"{{IGO_MERMAID_LIB}}":  string(libURL),
		"{{IGO_MERMAID_JS}}":   string(thinURL),
		"{{DOC_CODE_CSS}}":     docCodeBlockCSS(),
		"{{DOC_CODE_JS}}":      docCodeChromeScript(),
		"{{IGO_AUTH_PANEL}}":   igoAuthPanelHTML,
		"{{IGO_AUTH_CSS}}":     igoAuthCSS,
		"{{IGO_AUTH_JS}}":      igoAuthScriptTag(),
	}, visitorGeoLabel(r))
}

func igoadminCategoriesHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if docStore == nil {
		jsonError(w, http.StatusServiceUnavailable, "doc store unavailable")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) {
		return
	}
	cats, err := docStore.ListCategories()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "list categories failed")
		return
	}
	if cats == nil {
		cats = []string{}
	}
	jsonWrite(w, cats)
}

func igoadminCategoryHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if docStore == nil {
		jsonError(w, http.StatusServiceUnavailable, "doc store unavailable")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) {
		return
	}
	switch r.Method {
	case http.MethodPut:
		var body struct {
			Old string `json:"old"`
			New string `json:"new"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
			jsonError(w, http.StatusBadRequest, "bad json")
			return
		}
		if err := docStore.RenameCategory(body.Old, body.New); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		invalidateDocChromeCache()
		scheduleDocBackup(dataRoot)
		jsonWrite(w, map[string]any{"ok": true})
	case http.MethodDelete:
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		if err := docStore.RemoveCategory(name); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		invalidateDocChromeCache()
		scheduleDocBackup(dataRoot)
		jsonWrite(w, map[string]any{"ok": true})
	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func igoadminArticlesHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if docStore == nil {
		jsonError(w, http.StatusServiceUnavailable, "doc store unavailable")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if q != "" || category != "" || r.URL.Query().Has("limit") || r.URL.Query().Has("offset") {
		if limit <= 0 {
			limit = 50
		}
		items, total, truncated, err := docStore.ListArticlesAdmin(category, q, limit, offset)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "list failed")
			return
		}
		if items == nil {
			items = []docindex.Meta{}
		}
		jsonWrite(w, map[string]any{
			"items": items, "total": total, "limit": limit, "offset": offset, "truncated": truncated,
		})
		return
	}
	all, err := docStore.ListAll()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "list failed")
		return
	}
	jsonWrite(w, all)
}

type igoadminBatchBody struct {
	Slugs            []string `json:"slugs"`
	UpdateCategories bool     `json:"updateCategories"`
	SetCategories    string   `json:"setCategories"`
	UpdateTags       bool     `json:"updateTags"`
	SetTags          string   `json:"setTags"`
	TagsMode         string   `json:"tagsMode"`
	DryRun           bool     `json:"dryRun"`
}

func igoadminArticlesBatchHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPatch {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if docStore == nil {
		jsonError(w, http.StatusServiceUnavailable, "doc store unavailable")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) {
		return
	}
	var body igoadminBatchBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !body.UpdateCategories && !body.UpdateTags {
		jsonError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	var slugs []string
	for _, s := range body.Slugs {
		s = strings.TrimSpace(s)
		if s != "" && docindex.ValidSlug(s) {
			slugs = append(slugs, s)
		}
	}
	if len(slugs) == 0 {
		jsonError(w, http.StatusBadRequest, "no valid slugs")
		return
	}
	opts := docindex.PatchMetaOpts{TagsMode: body.TagsMode}
	if body.UpdateCategories {
		c := strings.TrimSpace(body.SetCategories)
		opts.Categories = &c
	}
	if body.UpdateTags {
		t := strings.TrimSpace(body.SetTags)
		opts.Tags = &t
	}
	if body.DryRun {
		type preview struct {
			Slug             string `json:"slug"`
			CategoriesBefore string `json:"categoriesBefore"`
			CategoriesAfter  string `json:"categoriesAfter"`
			TagsBefore       string `json:"tagsBefore"`
			TagsAfter        string `json:"tagsAfter"`
		}
		var items []preview
		var skipped []docindex.BatchPatchFailure
		for _, slug := range slugs {
			m, err := docStore.Get(slug)
			if err != nil {
				skipped = append(skipped, docindex.BatchPatchFailure{Slug: slug, Error: "not found"})
				continue
			}
			p := preview{Slug: slug, CategoriesBefore: m.Categories, TagsBefore: m.Tags}
			p.CategoriesAfter, p.TagsAfter = m.Categories, m.Tags
			if opts.Categories != nil {
				p.CategoriesAfter = *opts.Categories
			}
			if opts.Tags != nil {
				p.TagsAfter = docindex.ApplyTagsPatch(m.Tags, *opts.Tags, opts.TagsMode)
			}
			items = append(items, p)
		}
		if skipped == nil {
			skipped = []docindex.BatchPatchFailure{}
		}
		jsonWrite(w, map[string]any{"dryRun": true, "count": len(items), "items": items, "skipped": skipped})
		return
	}
	ok, failed := docStore.BatchPatchMeta(slugs, opts)
	if ok > 0 {
		invalidateAllDocRenderCache()
		invalidateDocChromeCache()
		scheduleDocBackup(dataRoot)
	}
	if failed == nil {
		failed = []docindex.BatchPatchFailure{}
	}
	jsonWrite(w, map[string]any{"ok": ok, "failed": failed})
}

func igoadminArticleHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if docStore == nil {
		jsonError(w, http.StatusServiceUnavailable, "doc store unavailable")
		return
	}
	switch r.Method {
	case http.MethodPost:
		if !igoadminRequireToken(w, r, dataRoot) {
			return
		}
		var body igoadminArticleBody
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, igoadminArticleMaxBody)).Decode(&body); err != nil {
			jsonError(w, http.StatusBadRequest, "bad json")
			return
		}
		slug, ok := docindex.NormalizeSlug(body.Slug)
		if !ok {
			jsonError(w, http.StatusBadRequest, "invalid slug")
			return
		}
		exists, err := docStore.Exists(slug)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "check slug")
			return
		}
		if exists {
			jsonError(w, http.StatusConflict, "slug already exists")
			return
		}
		if docReservedSlug(slug) {
			jsonError(w, http.StatusBadRequest, "slug reserved")
			return
		}
		author := strings.TrimSpace(body.Author)
		if author == "" {
			author = docindex.DefaultAuthor
		}
		m := docindex.Meta{
			Slug: slug, Title: markdown.ResolveArticleTitle("", slug, []byte(body.Markdown)),
			Categories: strings.TrimSpace(body.Categories),
			Tags: strings.TrimSpace(body.Tags), Author: author, PublishedAt: parsePublishedAt(body.PublishedAt),
		}
		if m.Title == "" {
			m.Title = slug
		}
		v := markdown.Validate([]byte(body.Markdown))
		if !v.OK {
			igoadminWriteValidationFailed(w, v)
			return
		}
		md, err := prepareArticleMarkdown(dataRoot, slug, body.Markdown, body.DraftImages, "", false)
		if err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(md) == "" {
			md = "# " + m.Title + "\n\n"
		}
		if err := docStore.WriteArticle(m, []byte(md)); err != nil {
			jsonError(w, http.StatusInternalServerError, "save failed")
			return
		}
		noteDocArticleChanged(slug, dataRoot)
		igoadminArticleSavedResponse(w, v, map[string]any{"ok": true, "slug": slug, "markdown": md, "title": m.Title})

	case http.MethodPut:
		if !igoadminRequireToken(w, r, dataRoot) {
			return
		}
		var body igoadminArticleBody
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, igoadminArticleMaxBody)).Decode(&body); err != nil {
			jsonError(w, http.StatusBadRequest, "bad json")
			return
		}
		newSlug, ok := docindex.NormalizeSlug(body.Slug)
		if !ok {
			newSlug = strings.TrimSpace(body.Slug)
			if !docindex.ValidSlug(newSlug) {
				jsonError(w, http.StatusBadRequest, "invalid slug")
				return
			}
		}
		oldSlug := strings.TrimSpace(body.OldSlug)
		if oldSlug == "" {
			oldSlug = newSlug
		}
		if !docindex.ValidSlug(oldSlug) {
			jsonError(w, http.StatusBadRequest, "invalid old slug")
			return
		}
		existing, err := docStore.Get(oldSlug)
		if err != nil {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		oldRaw, _ := docStore.ReadBody(oldSlug)
		oldMD := string(oldRaw)
		v := markdown.Validate([]byte(body.Markdown))
		if !v.OK {
			igoadminWriteValidationFailed(w, v)
			return
		}
		if oldSlug != newSlug {
			if docReservedSlug(newSlug) {
				jsonError(w, http.StatusBadRequest, "slug reserved")
				return
			}
			taken, err := docStore.Exists(newSlug)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "check slug")
				return
			}
			if taken {
				jsonError(w, http.StatusConflict, "slug already exists")
				return
			}
			if err := docStore.RenameArticle(oldSlug, newSlug); err != nil {
				if strings.Contains(err.Error(), "not found") {
					jsonError(w, http.StatusNotFound, "not found")
					return
				}
				if strings.Contains(err.Error(), "already exists") {
					jsonError(w, http.StatusConflict, "slug already exists")
					return
				}
				jsonError(w, http.StatusInternalServerError, "rename failed")
				return
			}
		}
		slug := newSlug
		title := strings.TrimSpace(body.Title)
		if title == "" {
			title = existing.Title
		}
		author := strings.TrimSpace(body.Author)
		if author == "" {
			author = strings.TrimSpace(existing.Author)
		}
		if author == "" {
			author = docindex.DefaultAuthor
		}
		m := docindex.Meta{
			Slug: slug, Title: title, Categories: strings.TrimSpace(body.Categories),
			Tags: strings.TrimSpace(body.Tags), Author: author, PublishedAt: parsePublishedAtForUpdate(body.PublishedAt, existing.PublishedAt),
			Views: existing.Views,
		}
		md, err := prepareArticleMarkdown(dataRoot, slug, body.Markdown, body.DraftImages, oldMD, true)
		if err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := docStore.WriteArticle(m, []byte(md)); err != nil {
			jsonError(w, http.StatusInternalServerError, "save failed")
			return
		}
		if oldSlug != slug {
			invalidateDocRenderCache(oldSlug)
		}
		noteDocArticleChanged(slug, dataRoot)
		igoadminArticleSavedResponse(w, v, map[string]any{"ok": true, "slug": slug, "markdown": md, "meta": m})

	case http.MethodDelete:
		if !igoadminRequireToken(w, r, dataRoot) {
			return
		}
		slug := strings.TrimSpace(r.URL.Query().Get("slug"))
		if !docindex.ValidSlug(slug) {
			jsonError(w, http.StatusBadRequest, "invalid slug")
			return
		}
		if _, err := docStore.Get(slug); err != nil {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		if err := deleteArticleWithMedia(docStore, docAssetsRoot(dataRoot), slug); err != nil {
			jsonError(w, http.StatusInternalServerError, "delete failed")
			return
		}
		noteDocArticleChanged(slug, dataRoot)
		jsonWrite(w, map[string]any{"ok": true})

	case http.MethodGet:
		slug := strings.TrimSpace(r.URL.Query().Get("slug"))
		if slug == "" {
			jsonError(w, http.StatusBadRequest, "slug required")
			return
		}
		if !igoadminRequireToken(w, r, dataRoot) {
			return
		}
		m, err := docStore.Get(slug)
		if err != nil {
			jsonError(w, http.StatusNotFound, "not found")
			return
		}
		raw, _ := docStore.ReadBody(slug)
		jsonWrite(w, map[string]any{
			"meta": m, "markdown": string(raw),
		})
	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func igoadminWriteValidationFailed(w http.ResponseWriter, v markdown.ValidationResult) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":      "markdown validation failed",
		"validation": v,
		"hint":       markdown.FormatValidationSummary(v),
	})
}

func igoadminArticleSavedResponse(w http.ResponseWriter, v markdown.ValidationResult, body map[string]any) {
	body["validation"] = v
	if hint := markdown.FormatValidationSummary(v); hint != "" {
		body["validationHint"] = hint
	}
	jsonWrite(w, body)
}

func igoadminPreviewHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) {
		return
	}
	var body struct {
		Markdown string `json:"markdown"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, igoadminArticleMaxBody)).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	v := markdown.Validate([]byte(body.Markdown))
	pub := publicBaseFromRequest(r, getenv("BASE_PATH", ""))
	src := markdown.StripLeadingH1([]byte(body.Markdown))
	html, err := markdown.Render(src, markdown.RenderOpts{
		PicBaseURL: docPicPublicBase(pub) + "/",
	})
	if err != nil {
		jsonError(w, http.StatusBadRequest, "render failed")
		return
	}
	jsonWrite(w, map[string]any{
		"html": string(html), "validation": v, "hint": markdown.FormatValidationSummary(v),
	})
}

func parsePublishedAt(s string) time.Time {
	if t, ok := parsePublishedAtValue(s); ok {
		return t
	}
	return time.Now()
}

func parsePublishedAtForUpdate(s string, fallback time.Time) time.Time {
	if t, ok := parsePublishedAtValue(s); ok {
		return t
	}
	return fallback
}

// parsePublishedAtValue parses admin publishedAt strings.
// Date-only values (YYYY-MM-DD) keep that calendar day and fill hour/minute/second from time.Now().
func parsePublishedAtValue(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		now := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local), true
	}
	return time.Time{}, false
}

func igoadminUploadMDHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) || docStore == nil {
		return
	}
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid multipart")
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	fh := r.MultipartForm.File["file"]
	if len(fh) == 0 {
		jsonError(w, http.StatusBadRequest, "no file")
		return
	}
	f, err := fh[0].Open()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "open file")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(f, 4<<20))
	f.Close()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "read file")
		return
	}
	slug, ok := docindex.NormalizeSlug(r.FormValue("slug"))
	if !ok {
		slug, ok = docindex.NormalizeSlug(strings.TrimSuffix(sanitizeFilename(fh[0].Filename), ".md"))
	}
	if !ok {
		jsonError(w, http.StatusBadRequest, "invalid slug")
		return
	}
	v := markdown.Validate(raw)
	if !v.OK {
		igoadminWriteValidationFailed(w, v)
		return
	}
	exists, _ := docStore.Exists(slug)
	if exists {
		jsonError(w, http.StatusConflict, "slug exists; use PUT to update")
		return
	}
	author := strings.TrimSpace(r.FormValue("author"))
	if author == "" {
		author = docindex.DefaultAuthor
	}
	title := markdown.ResolveArticleTitle("", slug, raw)
	m := docindex.Meta{
		Slug: slug, Title: title,
		Categories: strings.TrimSpace(r.FormValue("categories")),
		Tags:       strings.TrimSpace(r.FormValue("tags")),
		Author:     author,
		PublishedAt: parsePublishedAt(r.FormValue("publishedAt")),
	}
	if err := docStore.WriteArticle(m, raw); err != nil {
		jsonError(w, http.StatusInternalServerError, "save failed")
		return
	}
	noteDocArticleChanged(slug, dataRoot)
	jsonWrite(w, map[string]any{"ok": true, "slug": slug, "title": title})
}

func igoadminExportHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) || docStore == nil {
		return
	}
	if tryServeDocExportCache(w, r, dataRoot) {
		return
	}
	if !docExportBuildTryStart(dataRoot) {
		if waitForDocExportCache(dataRoot, docExportWaitTimeout) && tryServeDocExportCache(w, r, dataRoot) {
			return
		}
		if tryServeDocExportCacheStale(w, r, dataRoot) {
			return
		}
		jsonError(w, http.StatusServiceUnavailable, "export in progress, retry shortly")
		return
	}
	defer docExportBuildEnd(dataRoot)

	if err := streamDocExportZipMiss(w, dataRoot); err != nil {
		log.Printf("doc-export: stream: %v", err)
		if w.Header().Get("X-Doc-Export-Cache") == "" {
			jsonError(w, http.StatusInternalServerError, err.Error())
		}
	}
}

func igoadminImportHandler(w http.ResponseWriter, r *http.Request, dataRoot string) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !igoadminRequireToken(w, r, dataRoot) || docStore == nil {
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid multipart")
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	fhs := r.MultipartForm.File["file"]
	if len(fhs) == 0 {
		jsonError(w, http.StatusBadRequest, "no file")
		return
	}
	f, err := fhs[0].Open()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "open zip")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(f, zipImportArchiveMax))
	f.Close()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "read zip")
		return
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid zip")
		return
	}
	assets := docAssetsRoot(dataRoot)
	count, warnings, err := importDocZip(dataRoot, assets, docStore, zr)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	invalidateDocExportCache(dataRoot)
	noteDocArticleChanged("", dataRoot)
	triggerDocExportRebuild(dataRoot, true)
	out := map[string]any{"ok": true, "count": count}
	if len(warnings) > 0 {
		out["warnings"] = warnings
	}
	jsonWrite(w, out)
}