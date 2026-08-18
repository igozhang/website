package main

import (
	"path/filepath"
	"strings"
)

// previewKind 描述 /preview/ 如何展示该文件。
// kind: "md" 走 Markdown 渲染；"code" 走代码壳纯文本。
// 后缀列表须与 ui/ui-shared.js 的 isPreviewable 保持同步。
func previewKind(name string) (kind, lang string, ok bool) {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "" || base == "." || base == ".." {
		return "", "", false
	}
	lower := strings.ToLower(base)

	// 无标准后缀或带变体后缀的常见文本文件名。
	switch {
	case lower == "dockerfile" || strings.HasPrefix(lower, "dockerfile."):
		return "code", "shell", true
	case lower == "makefile" || lower == "gnumakefile" || strings.HasPrefix(lower, "makefile."):
		return "code", "shell", true
	case lower == "gemfile" || lower == "procfile" || lower == "jenkinsfile" || lower == "cmakelists.txt":
		return "code", "shell", true
	}

	ext := strings.ToLower(filepath.Ext(base))
	if ext == "" || ext == "." {
		return "", "", false
	}
	switch ext {
	case ".md", ".markdown":
		return "md", "", true

	case ".yaml", ".yml":
		return "code", "yaml", true
	case ".json", ".jsonc", ".json5":
		return "code", "json", true
	case ".go":
		return "code", "go", true
	case ".py", ".pyw", ".pyi":
		return "code", "python", true
	case ".xml", ".svg", ".html", ".htm", ".xhtml", ".xsl", ".xsd":
		return "code", "xml", true

	case ".sh", ".bash", ".zsh", ".fish", ".ksh",
		".conf", ".cfg", ".ini", ".env", ".properties", ".service", ".timer",
		".toml", ".tf", ".hcl",
		".js", ".mjs", ".cjs", ".ts", ".jsx", ".tsx", ".vue", ".svelte",
		".css", ".scss", ".sass", ".less",
		".sql", ".graphql", ".gql",
		".rs", ".java", ".kt", ".kts", ".scala",
		".c", ".h", ".cpp", ".cc", ".cxx", ".hpp", ".hh",
		".rb", ".php", ".lua", ".pl", ".pm", ".r", ".swift", ".m", ".mm",
		".ps1", ".psm1", ".bat", ".cmd",
		".txt", ".text", ".log", ".csv", ".tsv", ".lst", ".out", ".err",
		".gitignore", ".dockerignore", ".editorconfig", ".npmrc", ".nvmrc",
		".eslintrc", ".prettierrc", ".babelrc":
		// 无专用高亮器的语言走 shell/默认，仍有 Dark+ 代码壳与复制。
		lang = "shell"
		switch ext {
		case ".txt", ".text", ".log", ".csv", ".tsv", ".lst", ".out", ".err":
			lang = ""
		}
		return "code", lang, true
	default:
		return "", "", false
	}
}

func isPreviewableFilename(name string) bool {
	_, _, ok := previewKind(name)
	return ok
}

// previewableExtsForTest 返回可预览后缀（小写含点），供与前端列表对拍。
func previewableExtsForTest() []string {
	return []string{
		".md", ".markdown",
		".txt", ".text", ".log", ".csv", ".tsv", ".lst", ".out", ".err",
		".yaml", ".yml", ".json", ".jsonc", ".json5", ".toml", ".tf", ".hcl",
		".conf", ".cfg", ".ini", ".env", ".properties", ".service", ".timer",
		".js", ".mjs", ".cjs", ".ts", ".jsx", ".tsx", ".vue", ".svelte",
		".css", ".scss", ".sass", ".less",
		".go", ".py", ".pyw", ".pyi",
		".sh", ".bash", ".zsh", ".fish", ".ksh", ".ps1", ".psm1", ".bat", ".cmd",
		".xml", ".svg", ".html", ".htm", ".xhtml", ".xsl", ".xsd",
		".sql", ".graphql", ".gql",
		".rs", ".java", ".kt", ".kts", ".scala",
		".c", ".h", ".cpp", ".cc", ".cxx", ".hpp", ".hh",
		".rb", ".php", ".lua", ".pl", ".pm", ".r", ".swift", ".m", ".mm",
		".gitignore", ".dockerignore", ".editorconfig", ".npmrc", ".nvmrc",
		".eslintrc", ".prettierrc", ".babelrc",
	}
}
