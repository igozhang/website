package docindex

import "strings"

const slugFuzzyMinLen = 3

// ResolveSlug returns an existing slug: exact match first, else the best prefix extension (e.g. squid → squid_ins).
func (s *Store) ResolveSlug(query string) (string, bool) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" || !ValidSlug(query) {
		return "", false
	}
	ok, err := s.Exists(query)
	if err == nil && ok {
		return query, true
	}
	if len(query) < slugFuzzyMinLen {
		return "", false
	}
	all, err := s.ListAll()
	if err != nil {
		return "", false
	}
	var best string
	bestExtra := int(^uint(0) >> 1)
	for _, m := range all {
		slug := strings.ToLower(m.Slug)
		if slug == query {
			return m.Slug, true
		}
		if strings.HasPrefix(slug, query) {
			extra := len(slug) - len(query)
			if best == "" || extra < bestExtra {
				best = m.Slug
				bestExtra = extra
			}
		}
	}
	if best != "" {
		return best, true
	}
	return "", false
}
