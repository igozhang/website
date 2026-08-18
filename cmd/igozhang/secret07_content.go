package main

import (
	"bytes"
	"os"
	"strings"
)

// secret07HasEffectiveLines 是否存在非空且非 # 注释行。
func secret07HasEffectiveLines(b []byte) bool {
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return true
		}
	}
	return false
}

func secret07NormalizeContent(b []byte) []byte {
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	return []byte(strings.TrimSpace(s))
}

// secret07TreatAsMissing 自动生成的说明模板、或与模板完全一致、或无任何生效行时，
// 等同「无有效外部文件」，调用方应回退内嵌参数/素材。
func secret07TreatAsMissing(b []byte, template []byte) bool {
	if len(b) == 0 {
		return true
	}
	if !secret07HasEffectiveLines(b) {
		return true
	}
	if len(template) > 0 && bytes.Equal(secret07NormalizeContent(b), secret07NormalizeContent(template)) {
		return true
	}
	return false
}

// secret07ReadFileOrMissing 读 secret07 下文件；不存在或视为无效模板时 ok=false。
func secret07ReadFileOrMissing(path string, template []byte) (raw []byte, ok bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	if secret07TreatAsMissing(b, template) {
		return nil, false
	}
	return b, true
}
