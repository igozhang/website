package docindex

import (
	"github.com/igokorea/igozhang/internal/markdown"
)

// ExcerptMarkdownSkipTitle 生成摘要并跳过与文章标题重复的首段。
func ExcerptMarkdownSkipTitle(body []byte, maxLines int, title string) string {
	return markdown.PlainTextExcerpt(body, maxLines, title)
}
