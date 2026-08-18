package docindex

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

// PrimaryCategory returns the first category in a comma-separated list.
func PrimaryCategory(raw string) string {
	parts := SplitCategories(raw)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// SplitCategories splits a comma-separated category field into trimmed parts.
func SplitCategories(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// categoryMatchToken normalizes categories for substring matching: ",a,b,".
func categoryMatchToken(categories string) string {
	parts := SplitCategories(categories)
	if len(parts) == 0 {
		return ","
	}
	return "," + strings.Join(parts, ",") + ","
}

// ListCategories returns distinct category names sorted ascending.
func (s *Store) ListCategories() ([]string, error) {
	rows, err := s.db.Query(`SELECT categories FROM articles WHERE categories != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	set := make(map[string]struct{})
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		for _, c := range SplitCategories(raw) {
			set[c] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Strings(out)
	return out, nil
}

// ValidCategoryName reports whether a category label is allowed.
func ValidCategoryName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 64 {
		return false
	}
	return !strings.Contains(name, ",")
}

func joinCategories(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

func replaceCategoryInField(raw, oldName, newName string) (string, bool) {
	parts := SplitCategories(raw)
	if len(parts) == 0 {
		return raw, false
	}
	changed := false
	for i, p := range parts {
		if p == oldName {
			parts[i] = newName
			changed = true
		}
	}
	if !changed {
		return raw, false
	}
	return joinCategories(parts), true
}

func removeCategoryFromField(raw, name string) (string, bool) {
	parts := SplitCategories(raw)
	if len(parts) == 0 {
		return raw, false
	}
	var out []string
	changed := false
	for _, p := range parts {
		if p == name {
			changed = true
			continue
		}
		out = append(out, p)
	}
	if !changed {
		return raw, false
	}
	return joinCategories(out), true
}

// RenameCategory renames a category across all articles.
func (s *Store) RenameCategory(oldName, newName string) error {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if !ValidCategoryName(oldName) || !ValidCategoryName(newName) {
		return fmt.Errorf("invalid category name")
	}
	if oldName == newName {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT slug, categories FROM articles`)
	if err != nil {
		return err
	}
	type row struct {
		slug, cats string
	}
	var list []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.slug, &r.cats); err != nil {
			rows.Close()
			return err
		}
		list = append(list, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	now := nowUnix()
	for _, r := range list {
		next, ok := replaceCategoryInField(r.cats, oldName, newName)
		if !ok {
			continue
		}
		if _, err := s.db.Exec(`UPDATE articles SET categories = ?, updated_at = ? WHERE slug = ?`, next, now, r.slug); err != nil {
			return err
		}
	}
	return nil
}

// RemoveCategory removes a category label from all articles.
func (s *Store) RemoveCategory(name string) error {
	name = strings.TrimSpace(name)
	if !ValidCategoryName(name) {
		return fmt.Errorf("invalid category name")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT slug, categories FROM articles`)
	if err != nil {
		return err
	}
	type row struct {
		slug, cats string
	}
	var list []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.slug, &r.cats); err != nil {
			rows.Close()
			return err
		}
		list = append(list, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	now := nowUnix()
	for _, r := range list {
		next, ok := removeCategoryFromField(r.cats, name)
		if !ok {
			continue
		}
		if _, err := s.db.Exec(`UPDATE articles SET categories = ?, updated_at = ? WHERE slug = ?`, next, now, r.slug); err != nil {
			return err
		}
	}
	return nil
}

func nowUnix() int64 {
	return time.Now().Unix()
}
func escapeLikePattern(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '%', '_':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func categoryListSQL(category string) (where string, arg any) {
	category = strings.TrimSpace(category)
	if category == "" {
		return "", nil
	}
	needle := "," + escapeLikePattern(category) + ","
	return `WHERE (',' || replace(categories, ' ', '') || ',') LIKE ? ESCAPE '\'`, "%" + needle + "%"
}

// CountByCategory returns article count; empty category means all.
func (s *Store) CountByCategory(category string) (int, error) {
	where, arg := categoryListSQL(category)
	var n int
	var err error
	if where == "" {
		err = s.db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&n)
	} else {
		err = s.db.QueryRow(`SELECT COUNT(*) FROM articles `+where, arg).Scan(&n)
	}
	return n, err
}

// ListByCategory returns latest articles; empty category means all articles.
func (s *Store) ListByCategory(category string, limit int) ([]Meta, error) {
	return s.ListByCategoryPage(category, limit, 0)
}

// ListByCategoryPage returns a page of articles ordered by published_at desc.
func (s *Store) ListByCategoryPage(category string, limit, offset int) ([]Meta, error) {
	if limit <= 0 {
		limit = 5
	}
	if offset < 0 {
		offset = 0
	}
	where, arg := categoryListSQL(category)
	var rows *sql.Rows
	var err error
	if where == "" {
		rows, err = s.db.Query(`
SELECT slug, title, categories, tags, author, published_at, views
FROM articles ORDER BY published_at DESC LIMIT ? OFFSET ?`, limit, offset)
	} else {
		rows, err = s.db.Query(`
SELECT slug, title, categories, tags, author, published_at, views
FROM articles `+where+`
ORDER BY published_at DESC LIMIT ? OFFSET ?`, arg, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetaRows(rows)
}
