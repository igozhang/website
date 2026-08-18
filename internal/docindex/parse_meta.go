package docindex

import "time"

// DefaultAuthor is used when article author is unset.
const DefaultAuthor = "igozhang"

// Meta is article metadata stored in SQLite; exported as meta/{slug}.json in ZIP backups.
type Meta struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Categories  string    `json:"categories"`
	Tags        string    `json:"tags"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"publishedAt"`
	Views       int64     `json:"views"`
}
