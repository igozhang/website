package main

import (
	"bytes"
	"os"
	"strings"
)

type opsConfigLine struct {
	section string
	key     string
	value   string
	rawLine string
}

func parseOpsConfigEffectiveLines(text string) []opsConfigLine {
	var out []opsConfigLine
	section := ""
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.ToLower(strings.Trim(trimmed, "[]"))
			continue
		}
		key, val, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		out = append(out, opsConfigLine{
			section: section,
			key:     strings.TrimSpace(strings.ToLower(key)),
			value:   strings.TrimSpace(val),
			rawLine: trimmed,
		})
	}
	return out
}

func opsConfigLineKey(l opsConfigLine) string {
	return l.section + "\x00" + l.key + "\x00" + l.value
}

// buildIgoConfExportDiff 从当前 igo.conf 提取与内嵌默认不一致的生效配置行。
func buildIgoConfExportDiff(currentRaw []byte) []byte {
	cur := parseOpsConfigEffectiveLines(string(currentRaw))
	def := parseOpsConfigEffectiveLines(string(opsConfigExample))
	defSet := make(map[string]struct{}, len(def))
	for _, l := range def {
		defSet[opsConfigLineKey(l)] = struct{}{}
	}
	var (
		diffLines    []string
		seenSections = make(map[string]struct{})
		lastSection  string
	)
	for _, l := range cur {
		if _, ok := defSet[opsConfigLineKey(l)]; ok {
			continue
		}
		if l.section != lastSection {
			if _, ok := seenSections[l.section]; !ok && l.section != "" {
				diffLines = append(diffLines, "["+l.section+"]")
				seenSections[l.section] = struct{}{}
			}
			lastSection = l.section
		}
		diffLines = append(diffLines, l.rawLine)
	}
	if len(diffLines) == 0 {
		return nil
	}
	const header = "# igo.conf 非默认生效项（从运行中 secret07/igo.conf 导出）\n" +
		"# 合并到目标 secret07/igo.conf 后重启 Pod 生效\n\n"
	return []byte(header + strings.Join(diffLines, "\n") + "\n")
}

func isIgoConfExportDiff(raw []byte) bool {
	return bytes.HasPrefix(raw, []byte("# igo.conf 非默认生效项"))
}

// mergeIgoConfExportDiff 将旧版 ZIP 中的 diff 片段合并进内嵌默认 igo.conf，得到可写回的完整配置。
func mergeIgoConfExportDiff(baseRaw, diffRaw []byte) []byte {
	overlay := parseOpsConfigEffectiveLines(string(diffRaw))
	if len(overlay) == 0 {
		return append([]byte(nil), baseRaw...)
	}
	pending := make(map[string]opsConfigLine, len(overlay))
	for _, l := range overlay {
		pending[l.section+"\x00"+l.key] = l
	}
	var out []string
	section := ""
	for _, line := range strings.Split(string(baseRaw), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.ToLower(strings.Trim(trimmed, "[]"))
			out = append(out, line)
			continue
		}
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			if key, _, ok := strings.Cut(trimmed, "="); ok {
				k := strings.TrimSpace(strings.ToLower(key))
				if l, found := pending[section+"\x00"+k]; found {
					out = append(out, l.rawLine)
					delete(pending, section+"\x00"+k)
					continue
				}
			}
		}
		out = append(out, line)
	}
	if len(pending) > 0 {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		lastSec := ""
		for _, l := range overlay {
			key := l.section + "\x00" + l.key
			if _, ok := pending[key]; !ok {
				continue
			}
			if l.section != lastSec {
				out = append(out, "["+l.section+"]")
				lastSec = l.section
			}
			out = append(out, l.rawLine)
			delete(pending, key)
		}
	}
	return []byte(strings.Join(out, "\n"))
}

func normalizeIgoConfImportBytes(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, os.ErrInvalid
	}
	if isIgoConfExportDiff(raw) {
		return mergeIgoConfExportDiff(opsConfigExample, raw), nil
	}
	return raw, nil
}

func readIgoConfExportBytes(dataRoot string) []byte {
	if dataRoot == "" {
		return nil
	}
	raw, ok := secret07ReadFileOrMissing(opsConfigPath(dataRoot), opsConfigExample)
	if !ok {
		return nil
	}
	if len(buildIgoConfExportDiff(raw)) == 0 {
		return nil
	}
	return raw
}

func readLinkConfExportBytes(dataRoot string) []byte {
	if dataRoot == "" {
		return nil
	}
	raw, ok := secret07ReadFileOrMissing(linkConfPath(dataRoot), []byte(secret07LinkConfTemplate))
	if !ok {
		return nil
	}
	return raw
}

func readSSHPassExportBytes(dataRoot string) []byte {
	if dataRoot == "" {
		return nil
	}
	raw, ok := secret07ReadFileOrMissing(sshPassPath(dataRoot), []byte(secret07SSHPassTemplate))
	if !ok {
		return nil
	}
	if _, err := listSSHTargets(dataRoot); err != nil {
		return nil
	}
	return raw
}
