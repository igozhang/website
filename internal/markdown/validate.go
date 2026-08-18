package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationIssue is a single format check result.
type ValidationIssue struct {
	Level   string `json:"level"` // "error" | "warn"
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

// ValidationResult aggregates markdown format checks before save.
type ValidationResult struct {
	OK       bool              `json:"ok"`
	Errors   []ValidationIssue   `json:"errors"`
	Warnings []ValidationIssue   `json:"warnings"`
}

var (
	docPicRefRE     = regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)
	unclosedImageRE = regexp.MustCompile(`!\[[^\]]*\]\([^)\n]*$`)
	unsafePathRE    = regexp.MustCompile(`(?i)(^|/)\.\.(/|$)|\s`)
)

// Validate checks article markdown; errors block save, warnings are advisory.
func Validate(src []byte) ValidationResult {
	var res ValidationResult
	res.Errors = []ValidationIssue{}
	res.Warnings = []ValidationIssue{}
	text := string(src)

	checkFencedCodeBlocks(text, &res)
	checkImageLinks(text, &res)
	checkDocPicRefs(text, &res)
	checkLongLines(text, &res)
	checkGoldmarkParse(src, &res)

	res.OK = len(res.Errors) == 0
	if res.Warnings == nil {
		res.Warnings = []ValidationIssue{}
	}
	if res.Errors == nil {
		res.Errors = []ValidationIssue{}
	}
	return res
}

func checkFencedCodeBlocks(text string, res *ValidationResult) {
	open := false
	openLine := 0
	for i, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			if open {
				open = false
			} else {
				open = true
				openLine = i + 1
			}
		}
	}
	if open {
		res.Errors = append(res.Errors, ValidationIssue{
			Level: "error", Line: openLine,
			Message: "代码块未闭合（缺少结束 ```）",
		})
	}
}

func checkImageLinks(text string, res *ValidationResult) {
	for i, line := range strings.Split(text, "\n") {
		if unclosedImageRE.FindString(line) != "" {
			res.Errors = append(res.Errors, ValidationIssue{
				Level: "error", Line: i + 1,
				Message: "图片语法未闭合（缺少 ））",
			})
		}
	}
}

func checkDocPicRefs(text string, res *ValidationResult) {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		for _, m := range docPicRefRE.FindAllStringSubmatch(line, -1) {
			if len(m) < 2 {
				continue
			}
			path := strings.TrimSpace(m[1])
			if strings.HasPrefix(path, "__draft__:") {
				continue
			}
			if !strings.HasPrefix(path, "doc-pic/") && !strings.Contains(path, "://") {
				res.Warnings = append(res.Warnings, ValidationIssue{
					Level: "warn", Line: i + 1,
					Message: "图片路径建议使用 doc-pic/ 或完整 URL",
				})
			}
			if unsafePathRE.MatchString(path) {
				res.Warnings = append(res.Warnings, ValidationIssue{
					Level: "warn", Line: i + 1,
					Message: "图片路径含空格或 ..，前台可能无法加载",
				})
			}
		}
	}
}

func checkLongLines(text string, res *ValidationResult) {
	const max = 4000
	for i, line := range strings.Split(text, "\n") {
		if len(line) > max {
			res.Warnings = append(res.Warnings, ValidationIssue{
				Level: "warn", Line: i + 1,
				Message: fmt.Sprintf("单行超过 %d 字符，建议换行", max),
			})
		}
	}
}

func checkGoldmarkParse(src []byte, res *ValidationResult) {
	if _, err := convertMarkdown(src); err != nil {
		res.Errors = append(res.Errors, ValidationIssue{
			Level: "error", Message: "Markdown 解析失败：" + err.Error(),
		})
	}
}

// FormatValidationSummary returns a short line for admin status text.
func FormatValidationSummary(v ValidationResult) string {
	if v.OK && len(v.Warnings) == 0 {
		return ""
	}
	var parts []string
	if len(v.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("%d 个错误", len(v.Errors)))
	}
	if len(v.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("%d 条提示", len(v.Warnings)))
	}
	if len(parts) == 0 {
		return ""
	}
	s := strings.Join(parts, "，")
	if len(v.Errors) > 0 && len(v.Errors) <= 2 {
		var msgs []string
		for _, e := range v.Errors {
			msg := e.Message
			if e.Line > 0 {
				msg = fmt.Sprintf("第 %d 行：%s", e.Line, e.Message)
			}
			msgs = append(msgs, msg)
		}
		s += "（" + strings.Join(msgs, "；") + "）"
	} else if len(v.Warnings) > 0 && len(v.Errors) == 0 && len(v.Warnings) <= 2 {
		var msgs []string
		for _, w := range v.Warnings {
			msg := w.Message
			if w.Line > 0 {
				msg = fmt.Sprintf("第 %d 行：%s", w.Line, w.Message)
			}
			msgs = append(msgs, msg)
		}
		s += "（" + strings.Join(msgs, "；") + "）"
	}
	return s
}
