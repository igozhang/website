package markdown

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	rePlainLink    = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	rePlainImage   = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
	rePlainEm      = regexp.MustCompile(`[*_]{1,3}([^*_]+)[*_]{1,3}`)
	rePlainStrike  = regexp.MustCompile(`~~([^~]+)~~`)
	rePlainInline  = regexp.MustCompile("`([^`]+)`")
	rePlainHTMLTag = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	rePlainEscape  = regexp.MustCompile(`\\([!"#$%&'()*+,./:;<=>?@[\\\]^_` + "`" + `{|}~-])`)
	rePlainListNum = regexp.MustCompile(`^(\d+)\\\.(\s*)`)
	rePlainOrdList = regexp.MustCompile(`^\d+[.)]\s+`)
	rePlainAutolink = regexp.MustCompile(`<((?:https?|mailto):[^>]+)>`)
	rePlainHR      = regexp.MustCompile(`^[-*_]{3,}\s*$`)
	rePlainTable   = regexp.MustCompile(`^\|`)
	rePlainHTMLCmt = regexp.MustCompile(`<!--.*?-->`)
	rePlainSpaces  = regexp.MustCompile(`\s+`)
)

// PlainTextExcerpt 从 Markdown 正文提取可读纯文本摘要（非 HTML 源码）。
// skipTitle 非空时跳过与标题相同的首段（避免卡片摘要重复标题）。
func PlainTextExcerpt(body []byte, maxLines int, skipTitle string) string {
	if maxLines <= 0 {
		maxLines = 3
	}
	skipTitle = normalizePlainCompare(skipTitle)
	s := stripFencedCodeBlocks(string(body))
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if rePlainHR.MatchString(line) || rePlainTable.MatchString(line) {
			continue
		}
		line = stripLinePrefix(line)
		if line == "" {
			continue
		}
		line = plainTextLine(line)
		if line == "" {
			continue
		}
		if skipTitle != "" && normalizePlainCompare(line) == skipTitle {
			skipTitle = "" // 只跳过一次
			continue
		}
		lines = append(lines, line)
		if len(lines) >= maxLines {
			break
		}
	}
	out := rePlainSpaces.ReplaceAllString(strings.Join(lines, " "), " ")
	return trimPlainRunes(strings.TrimSpace(out), 400)
}

func stripFencedCodeBlocks(s string) string {
	for {
		i := strings.Index(s, "```")
		if i < 0 {
			return s
		}
		rest := s[i+3:]
		j := strings.Index(rest, "```")
		if j < 0 {
			return s[:i]
		}
		s = s[:i] + " " + rest[j+3:]
	}
}

func stripLinePrefix(line string) string {
	if strings.HasPrefix(line, "#") {
		line = strings.TrimLeft(line, "#")
		line = strings.TrimSpace(line)
		return line
	}
	if strings.HasPrefix(line, ">") {
		line = strings.TrimSpace(strings.TrimPrefix(line, ">"))
	}
	for _, p := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(line, p) {
			return strings.TrimSpace(line[len(p):])
		}
	}
	if rePlainOrdList.MatchString(line) {
		return strings.TrimSpace(rePlainOrdList.ReplaceAllString(line, ""))
	}
	return line
}

func plainTextLine(s string) string {
	s = rePlainHTMLCmt.ReplaceAllString(s, "")
	s = rePlainHTMLTag.ReplaceAllString(s, "")
	s = rePlainImage.ReplaceAllStringFunc(s, func(m string) string {
		sub := rePlainImage.FindStringSubmatch(m)
		if len(sub) < 2 {
			return ""
		}
		return strings.TrimSpace(sub[1])
	})
	s = rePlainLink.ReplaceAllString(s, "$1")
	s = rePlainStrike.ReplaceAllString(s, "$1")
	s = rePlainEm.ReplaceAllString(s, "$1")
	s = rePlainInline.ReplaceAllString(s, "$1")
	s = rePlainAutolink.ReplaceAllString(s, "$1")
	s = rePlainListNum.ReplaceAllString(s, "$1.$2")
	s = unescapeMarkdownPlain(s)
	s = rePlainSpaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func unescapeMarkdownPlain(s string) string {
	return rePlainEscape.ReplaceAllString(s, "$1")
}

func normalizePlainCompare(s string) string {
	s = plainTextLine(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

func trimPlainRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	var n int
	for i := range s {
		if n == max {
			return strings.TrimSpace(s[:i]) + "…"
		}
		n++
	}
	return s
}
