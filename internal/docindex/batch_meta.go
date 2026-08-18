package docindex

import (
	"fmt"
	"strings"
	"time"
)

// SplitTags splits comma-separated tags into trimmed parts.
func SplitTags(raw string) []string {
	return SplitCategories(raw)
}

func joinTags(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

// MergeTagsField appends tags (comma-separated) with deduplication preserving order.
func MergeTagsField(old, add string) string {
	seen := make(map[string]struct{})
	var out []string
	for _, t := range append(SplitTags(old), SplitTags(add)...) {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return joinTags(out)
}

// RemoveTagsField removes listed tags from old field.
func RemoveTagsField(old, remove string) string {
	rm := make(map[string]struct{})
	for _, t := range SplitTags(remove) {
		rm[t] = struct{}{}
	}
	var out []string
	for _, t := range SplitTags(old) {
		if _, drop := rm[t]; drop {
			continue
		}
		out = append(out, t)
	}
	return joinTags(out)
}

// ApplyTagsPatch applies replace / append / remove to an existing tags field.
func ApplyTagsPatch(old, patch, mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	switch mode {
	case "append":
		return MergeTagsField(old, patch)
	case "remove":
		return RemoveTagsField(old, patch)
	default:
		return strings.TrimSpace(patch)
	}
}

// ArticleHasCategory reports whether article categories contain the given name (empty name matches all).
func ArticleHasCategory(articleCategories, category string) bool {
	category = strings.TrimSpace(category)
	if category == "" {
		return true
	}
	for _, c := range SplitCategories(articleCategories) {
		if c == category {
			return true
		}
	}
	return false
}

// PatchMetaOpts selects metadata fields to update on PatchArticleMeta.
type PatchMetaOpts struct {
	Categories *string
	Tags       *string
	TagsMode   string
}

// PatchArticleMeta updates categories/tags in SQLite and FTS only (markdown body unchanged).
func (s *Store) PatchArticleMeta(slug string, opts PatchMetaOpts) error {
	if !ValidSlug(slug) {
		return fmt.Errorf("invalid slug")
	}
	if opts.Categories == nil && opts.Tags == nil {
		return fmt.Errorf("nothing to update")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var title, categories, tags, author string
	var pub, views int64
	err := s.db.QueryRow(`
SELECT title, categories, tags, author, published_at, views
FROM articles WHERE slug = ?`, slug).Scan(&title, &categories, &tags, &author, &pub, &views)
	if err != nil {
		return fmt.Errorf("not found")
	}
	if opts.Categories != nil {
		cat := strings.TrimSpace(*opts.Categories)
		if cat != "" && !ValidCategoryName(cat) {
			return fmt.Errorf("invalid category")
		}
		categories = cat
	}
	if opts.Tags != nil {
		tags = ApplyTagsPatch(tags, *opts.Tags, opts.TagsMode)
	}
	now := time.Now().Unix()
	if _, err := s.db.Exec(`
UPDATE articles SET categories = ?, tags = ?, updated_at = ? WHERE slug = ?`,
		categories, tags, now, slug); err != nil {
		return err
	}
	return s.syncFTS(slug, title, tags)
}

// BatchPatchFailure is a per-slug error from BatchPatchMeta.
type BatchPatchFailure struct {
	Slug  string `json:"slug"`
	Error string `json:"error"`
}

// BatchPatchMeta applies PatchMetaOpts to each slug; returns success count and failures.
func (s *Store) BatchPatchMeta(slugs []string, opts PatchMetaOpts) (int, []BatchPatchFailure) {
	var failed []BatchPatchFailure
	ok := 0
	seen := make(map[string]struct{})
	for _, slug := range slugs {
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}
		if _, dup := seen[slug]; dup {
			continue
		}
		seen[slug] = struct{}{}
		if err := s.PatchArticleMeta(slug, opts); err != nil {
			failed = append(failed, BatchPatchFailure{Slug: slug, Error: err.Error()})
			continue
		}
		ok++
	}
	return ok, failed
}

// ListArticlesAdmin returns a page of articles filtered by category and/or search query.
// When truncated is true, keyword matches were capped (see searchCap) and total may be understated.
func (s *Store) ListArticlesAdmin(category, q string, limit, offset int) (items []Meta, total int, truncated bool, err error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	category = strings.TrimSpace(category)
	q = strings.TrimSpace(q)

	if q == "" {
		total, err = s.CountByCategory(category)
		if err != nil {
			return nil, 0, false, err
		}
		items, err = s.ListByCategoryPage(category, limit, offset)
		return items, total, false, err
	}

	const searchCap = 5000
	var hits []Meta
	hits, err = s.Search(q, searchCap)
	if err != nil {
		return nil, 0, false, err
	}
	truncated = len(hits) >= searchCap
	var filtered []Meta
	for _, m := range hits {
		if ArticleHasCategory(m.Categories, category) {
			filtered = append(filtered, m)
		}
	}
	total = len(filtered)
	if offset >= total {
		return []Meta{}, total, truncated, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	out := make([]Meta, end-offset)
	copy(out, filtered[offset:end])
	return out, total, truncated, nil
}
