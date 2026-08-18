package docindex

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Store holds article metadata in SQLite; bodies live in assets/md.
type Store struct {
	db        *sql.DB
	assetsDir string
	mu        sync.RWMutex
}

// Open opens or creates the SQLite file at dbPath; article files remain under assetsDir.
func Open(dbPath, assetsDir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(assetsDir, "md"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(assetsDir, "doc-pic"), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, assetsDir: assetsDir}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS articles (
  slug TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  categories TEXT NOT NULL DEFAULT '',
  tags TEXT NOT NULL DEFAULT '',
  author TEXT NOT NULL DEFAULT '',
  published_at INTEGER NOT NULL,
  views INTEGER NOT NULL DEFAULT 0,
  md_mtime INTEGER NOT NULL DEFAULT 0,
  updated_at INTEGER NOT NULL DEFAULT 0
);
CREATE VIRTUAL TABLE IF NOT EXISTS articles_fts USING fts5(
  slug,
  title,
  tags,
  tokenize='unicode61 remove_diacritics 0'
);
`)
	if err != nil {
		return err
	}
	if err := s.migrateFTSSlugIndexed(); err != nil {
		return err
	}
	return s.backfillDefaultAuthor()
}

func (s *Store) backfillDefaultAuthor() error {
	_, err := s.db.Exec(`UPDATE articles SET author = ? WHERE author IS NULL OR trim(author) = ''`, DefaultAuthor)
	return err
}

// migrateFTSSlugIndexed rebuilds articles_fts when slug was UNINDEXED (legacy schema).
func (s *Store) migrateFTSSlugIndexed() error {
	var ddl string
	err := s.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='articles_fts'`).Scan(&ddl)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToUpper(ddl), "UNINDEXED") {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`
CREATE VIRTUAL TABLE articles_fts_new USING fts5(
  slug,
  title,
  tags,
  tokenize='unicode61 remove_diacritics 0'
);
INSERT INTO articles_fts_new(slug, title, tags) SELECT slug, title, tags FROM articles_fts;
DROP TABLE articles_fts;
ALTER TABLE articles_fts_new RENAME TO articles_fts;`); err != nil {
		return err
	}
	return tx.Commit()
}

// AssetsDir returns the assets root (…/assets).
func (s *Store) AssetsDir() string { return s.assetsDir }

func (s *Store) mdPath(slug string) string {
	return filepath.Join(s.assetsDir, "md", slug+".md")
}

func (s *Store) syncFTS(slug, title, tags string) error {
	_, err := s.db.Exec(`DELETE FROM articles_fts WHERE slug = ?`, slug)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO articles_fts(slug, title, tags) VALUES(?,?,?)`, slug, title, tags)
	return err
}

// Reconcile scans md/*.md and ensures each file has a SQLite row (preserves existing metadata).
func (s *Store) Reconcile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mdDir := filepath.Join(s.assetsDir, "md")
	entries, err := os.ReadDir(mdDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	existing := make(map[string]Meta)
	rows, err := s.db.Query(`
SELECT slug, title, categories, tags, author, published_at, views FROM articles`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var m Meta
		var pub int64
		if err := rows.Scan(&m.Slug, &m.Title, &m.Categories, &m.Tags, &m.Author, &pub, &m.Views); err != nil {
			rows.Close()
			return err
		}
		m.PublishedAt = time.Unix(pub, 0)
		existing[m.Slug] = m
	}
	if err := rows.Close(); err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		m, ok := existing[slug]
		if !ok {
			m = Meta{
				Slug:        slug,
				Title:       slug,
				Author:      DefaultAuthor,
				PublishedAt: time.Now(),
			}
		} else if strings.TrimSpace(m.Author) == "" {
			m.Author = DefaultAuthor
		}
		st, err := os.Stat(s.mdPath(slug))
		mdMtime := int64(0)
		if err == nil {
			mdMtime = st.ModTime().Unix()
		}
		pub := m.PublishedAt.Unix()
		if pub == 0 {
			pub = now
		}
		_, err = s.db.Exec(`
INSERT INTO articles(slug,title,categories,tags,author,published_at,views,md_mtime,updated_at)
VALUES(?,?,?,?,?,?,?,?,?)
ON CONFLICT(slug) DO UPDATE SET
  md_mtime=excluded.md_mtime,
  updated_at=excluded.updated_at`,
			slug, m.Title, m.Categories, m.Tags, m.Author, pub, m.Views, mdMtime, now,
		)
		if err != nil {
			return err
		}
		if err := s.syncFTS(slug, m.Title, m.Tags); err != nil {
			return err
		}
	}
	return nil
}

// ListLatest returns articles ordered by published_at desc.
func (s *Store) ListLatest(limit int) ([]Meta, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := s.db.Query(`
SELECT slug, title, categories, tags, author, published_at, views
FROM articles ORDER BY published_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetaRows(rows)
}

// ListAll returns all articles sorted by published_at desc.
func (s *Store) ListAll() ([]Meta, error) {
	rows, err := s.db.Query(`
SELECT slug, title, categories, tags, author, published_at, views
FROM articles ORDER BY published_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetaRows(rows)
}

func scanMetaRows(rows *sql.Rows) ([]Meta, error) {
	var out []Meta
	for rows.Next() {
		var m Meta
		var pub int64
		if err := rows.Scan(&m.Slug, &m.Title, &m.Categories, &m.Tags, &m.Author, &pub, &m.Views); err != nil {
			return nil, err
		}
		m.PublishedAt = time.Unix(pub, 0)
		out = append(out, m)
	}
	return out, rows.Err()
}

// Get returns metadata for slug.
func (s *Store) Get(slug string) (Meta, error) {
	var m Meta
	var pub int64
	err := s.db.QueryRow(`
SELECT slug, title, categories, tags, author, published_at, views
FROM articles WHERE slug = ?`, slug).Scan(
		&m.Slug, &m.Title, &m.Categories, &m.Tags, &m.Author, &pub, &m.Views,
	)
	if err != nil {
		return m, err
	}
	m.PublishedAt = time.Unix(pub, 0)
	return m, nil
}

// Exists reports whether slug is taken.
func (s *Store) Exists(slug string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT 1 FROM articles WHERE slug = ?`, slug).Scan(&n)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ReadBody loads markdown body from disk.
func (s *Store) ReadBody(slug string) ([]byte, error) {
	return os.ReadFile(s.mdPath(slug))
}

// ImportArticleMeta updates SQLite/FTS from m without writing the markdown file (body already on disk).
func (s *Store) ImportArticleMeta(m Meta) error {
	if !ValidSlug(m.Slug) {
		return fmt.Errorf("invalid slug")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := os.Stat(s.mdPath(m.Slug))
	if err != nil {
		return err
	}
	mdMtime := st.ModTime().Unix()
	pub := m.PublishedAt.Unix()
	if pub == 0 {
		pub = time.Now().Unix()
	}
	now := time.Now().Unix()
	_, err = s.db.Exec(`
INSERT INTO articles(slug,title,categories,tags,author,published_at,views,md_mtime,updated_at)
VALUES(?,?,?,?,?,?,?,?,?)
ON CONFLICT(slug) DO UPDATE SET
  title=excluded.title, categories=excluded.categories, tags=excluded.tags,
  author=excluded.author, published_at=excluded.published_at,
  views=excluded.views,
  md_mtime=excluded.md_mtime, updated_at=excluded.updated_at`,
		m.Slug, m.Title, m.Categories, m.Tags, m.Author, pub, m.Views, mdMtime, now,
	)
	if err != nil {
		return err
	}
	return s.syncFTS(m.Slug, m.Title, m.Tags)
}

// ImportArticle sets metadata and body exactly (for ZIP restore; views are not merged).
func (s *Store) ImportArticle(m Meta, body []byte) error {
	if !ValidSlug(m.Slug) {
		return fmt.Errorf("invalid slug")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.WriteFile(s.mdPath(m.Slug), body, 0o644); err != nil {
		return err
	}
	st, _ := os.Stat(s.mdPath(m.Slug))
	mdMtime := time.Now().Unix()
	if st != nil {
		mdMtime = st.ModTime().Unix()
	}
	pub := m.PublishedAt.Unix()
	if pub == 0 {
		pub = time.Now().Unix()
	}
	now := time.Now().Unix()
	_, err := s.db.Exec(`
INSERT INTO articles(slug,title,categories,tags,author,published_at,views,md_mtime,updated_at)
VALUES(?,?,?,?,?,?,?,?,?)
ON CONFLICT(slug) DO UPDATE SET
  title=excluded.title, categories=excluded.categories, tags=excluded.tags,
  author=excluded.author, published_at=excluded.published_at,
  views=excluded.views,
  md_mtime=excluded.md_mtime, updated_at=excluded.updated_at`,
		m.Slug, m.Title, m.Categories, m.Tags, m.Author, pub, m.Views, mdMtime, now,
	)
	if err != nil {
		return err
	}
	return s.syncFTS(m.Slug, m.Title, m.Tags)
}

// WriteArticle saves metadata and markdown body; creates or updates.
func (s *Store) WriteArticle(m Meta, body []byte) error {
	if !ValidSlug(m.Slug) {
		return fmt.Errorf("invalid slug")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.WriteFile(s.mdPath(m.Slug), body, 0o644); err != nil {
		return err
	}
	st, _ := os.Stat(s.mdPath(m.Slug))
	mdMtime := time.Now().Unix()
	if st != nil {
		mdMtime = st.ModTime().Unix()
	}
	pub := m.PublishedAt.Unix()
	if pub == 0 {
		pub = time.Now().Unix()
	}
	now := time.Now().Unix()
	views := m.Views
	_, err := s.db.Exec(`
INSERT INTO articles(slug,title,categories,tags,author,published_at,views,md_mtime,updated_at)
VALUES(?,?,?,?,?,?,?,?,?)
ON CONFLICT(slug) DO UPDATE SET
  title=excluded.title, categories=excluded.categories, tags=excluded.tags,
  author=excluded.author, published_at=excluded.published_at,
  views=CASE WHEN excluded.views > articles.views THEN excluded.views ELSE articles.views END,
  md_mtime=excluded.md_mtime, updated_at=excluded.updated_at`,
		m.Slug, m.Title, m.Categories, m.Tags, m.Author, pub, views, mdMtime, now,
	)
	if err != nil {
		return err
	}
	if err := s.syncFTS(m.Slug, m.Title, m.Tags); err != nil {
		return err
	}
	return nil
}

// RenameArticle moves metadata and markdown from oldSlug to newSlug.
func (s *Store) RenameArticle(oldSlug, newSlug string) error {
	if !ValidSlug(oldSlug) || !ValidSlug(newSlug) {
		return fmt.Errorf("invalid slug")
	}
	if oldSlug == newSlug {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var n int
	if err := s.db.QueryRow(`SELECT 1 FROM articles WHERE slug = ?`, oldSlug).Scan(&n); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}
	if err := s.db.QueryRow(`SELECT 1 FROM articles WHERE slug = ?`, newSlug).Scan(&n); err == nil {
		return fmt.Errorf("slug already exists")
	} else if err != sql.ErrNoRows {
		return err
	}

	oldPath, newPath := s.mdPath(oldSlug), s.mdPath(newSlug)
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	now := time.Now().Unix()
	if _, err := s.db.Exec(`UPDATE articles SET slug = ?, updated_at = ? WHERE slug = ?`, newSlug, now, oldSlug); err != nil {
		_ = os.Rename(newPath, oldPath)
		return err
	}
	_, _ = s.db.Exec(`DELETE FROM articles_fts WHERE slug = ?`, oldSlug)
	var title, tags string
	if err := s.db.QueryRow(`SELECT title, tags FROM articles WHERE slug = ?`, newSlug).Scan(&title, &tags); err != nil {
		return err
	}
	return s.syncFTS(newSlug, title, tags)
}

// DeleteArticle removes metadata, FTS row, and markdown file for slug.
func (s *Store) DeleteArticle(slug string) error {
	if !ValidSlug(slug) {
		return fmt.Errorf("invalid slug")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.db.Exec(`DELETE FROM articles WHERE slug = ?`, slug); err != nil {
		return err
	}
	_, _ = s.db.Exec(`DELETE FROM articles_fts WHERE slug = ?`, slug)
	if err := os.Remove(s.mdPath(slug)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IncrViews increments view count for slug.
func (s *Store) IncrViews(slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE articles SET views = views + 1, updated_at = ? WHERE slug = ?`, time.Now().Unix(), slug)
	return err
}

// Search finds articles by slug/title/tags (FTS prefix + slug LIKE, merged).
func (s *Store) Search(q string, limit int) ([]Meta, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	ftsOut, err := s.searchFTS(q, limit)
	if err != nil {
		ftsOut = nil
	}
	likeOut, err := s.searchLike(q, limit)
	if err != nil {
		if len(ftsOut) > 0 {
			return ftsOut, nil
		}
		return nil, err
	}
	return mergeSearchResults(ftsOut, likeOut, q, limit), nil
}

func (s *Store) searchFTS(q string, limit int) ([]Meta, error) {
	terms := strings.Fields(strings.ToLower(q))
	var ftsParts []string
	for _, t := range terms {
		t = strings.Trim(t, `"'`)
		if t == "" {
			continue
		}
		ftsParts = append(ftsParts, fmt.Sprintf(`"%s"*`, strings.ReplaceAll(t, `"`, "")))
	}
	if len(ftsParts) == 0 {
		return nil, nil
	}
	match := strings.Join(ftsParts, " AND ")
	rows, err := s.db.Query(`
SELECT a.slug, a.title, a.categories, a.tags, a.author, a.published_at, a.views
FROM articles_fts f
JOIN articles a ON a.slug = f.slug
WHERE articles_fts MATCH ?
ORDER BY a.published_at DESC
LIMIT ?`, match, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetaRows(rows)
}

func mergeSearchResults(fts, like []Meta, q string, limit int) []Meta {
	qLower := strings.ToLower(q)
	seen := make(map[string]struct{})
	var exact, slugHit, other []Meta
	add := func(m Meta, bucket *[]Meta) {
		if _, ok := seen[m.Slug]; ok {
			return
		}
		seen[m.Slug] = struct{}{}
		*bucket = append(*bucket, m)
	}
	for _, m := range like {
		slugLower := strings.ToLower(m.Slug)
		switch {
		case slugLower == qLower:
			add(m, &exact)
		case strings.Contains(slugLower, qLower):
			add(m, &slugHit)
		default:
			add(m, &other)
		}
	}
	for _, m := range fts {
		slugLower := strings.ToLower(m.Slug)
		switch {
		case slugLower == qLower:
			add(m, &exact)
		case strings.Contains(slugLower, qLower):
			add(m, &slugHit)
		default:
			add(m, &other)
		}
	}
	out := append(exact, slugHit...)
	out = append(out, other...)
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *Store) searchLike(q string, limit int) ([]Meta, error) {
	pat := "%" + strings.ToLower(q) + "%"
	rows, err := s.db.Query(`
SELECT slug, title, categories, tags, author, published_at, views
FROM articles
WHERE lower(slug) LIKE ? OR lower(title) LIKE ? OR lower(tags) LIKE ?
ORDER BY published_at DESC LIMIT ?`, pat, pat, pat, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetaRows(rows)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// ValidSlug checks slug format.
func ValidSlug(slug string) bool {
	if slug == "" || len(slug) > 64 {
		return false
	}
	for i, c := range slug {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' {
			continue
		}
		if i > 0 {
			return false
		}
		return false
	}
	return slug[0] >= 'a' && slug[0] <= 'z'
}
