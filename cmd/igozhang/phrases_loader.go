package main

import (
	_ "embed"
	"log"
	"math/rand"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const (
	phrasesFileName      = "phrases.md"
	phrasesMinExternal   = 150
	phrasesServePoolSize = 200
	phrasesMinValidRatio = 0.85 // 非注释非空行中，有效行占比下限
	phrasesMaxEnRunes    = 500
	phrasesMaxZhRunes    = 500
	phrasesMinEnRunes    = 2
	phrasesMinZhRunes    = 1
	phrasesPipe          = "\uFF5C" // 全角｜
)

// 内嵌 100 条；完整引语见 DATA_ROOT/assets/eng/phrases.md（≥150 条有效行时启用）。
// 通过校验后随机抽取 phrasesServePoolSize 条写入 phraseCache；不保留全集。
// 仅 Pod 重启或缓存为空时重新读盘并抽样。重新生成内嵌：node scripts/gen-phrases-embed.mjs
//
//go:embed phrases-embed.md
var phrasesEmbedded []byte

var (
	phraseCache         []phrasePair
	phraseCacheExternal bool
	phraseCacheMu       sync.RWMutex
	phraseRoot          string
)

func initPhrasesCache(dataRoot string) {
	phraseRoot = dataRoot
	phraseCacheMu.Lock()
	rebuildPhraseCacheLocked()
	phraseCacheMu.Unlock()
}

func reloadPhrasesCache() {
	if phraseRoot == "" {
		return
	}
	phraseCacheMu.Lock()
	rebuildPhraseCacheLocked()
	phraseCacheMu.Unlock()
}

func rebuildPhraseCacheLocked() {
	phraseCache, phraseCacheExternal = buildPhraseServePool(phraseRoot)
}

// phraseSourceIsExternal 当前页眉引语是否来自 DATA_ROOT/assets/eng/phrases.md 外部库。
func phraseSourceIsExternal() bool {
	phraseCacheMu.RLock()
	ext := phraseCacheExternal
	has := len(phraseCache) > 0
	phraseCacheMu.RUnlock()
	if !has {
		reloadPhrasesCache()
		phraseCacheMu.RLock()
		ext = phraseCacheExternal
		phraseCacheMu.RUnlock()
	}
	return ext
}

func phraseBankForUI() []phrasePair {
	phraseCacheMu.RLock()
	bank := phraseCache
	phraseCacheMu.RUnlock()
	if len(bank) == 0 {
		reloadPhrasesCache()
		phraseCacheMu.RLock()
		bank = phraseCache
		phraseCacheMu.RUnlock()
	}
	if len(bank) == 0 {
		return nil
	}
	out := make([]phrasePair, len(bank))
	copy(out, bank)
	return out
}

// buildPhraseServePool 读盘、校验、随机抽样；返回仅用于展示的固定池（最多 200 条）及是否启用外部库。
func buildPhraseServePool(dataRoot string) ([]phrasePair, bool) {
	embedded := parsePhrasesMarkdown(phrasesEmbedded, false)
	if dataRoot == "" {
		return samplePhrasePairs(embedded, phrasesServePoolSize), false
	}
	path := docAssetsPhrasesPath(dataRoot)
	b, ok := secret07ReadFileOrMissing(path, []byte(assetsPhrasesTemplate))
	if !ok {
		return samplePhrasePairs(embedded, phrasesServePoolSize), false
	}
	if !utf8.Valid(b) {
		log.Printf("phrases: %s: not valid UTF-8, using embedded", path)
		return samplePhrasePairs(embedded, phrasesServePoolSize), false
	}
	external, ok := parsePhrasesExternalMarkdown(b)
	if !ok {
		log.Printf("phrases: %s: format check failed, using embedded", path)
		return samplePhrasePairs(embedded, phrasesServePoolSize), false
	}
	return samplePhrasePairs(external, phrasesServePoolSize), true
}

// parsePhrasesExternalMarkdown 对外部 phrases.md 做严格校验；未通过则 ok=false。
func parsePhrasesExternalMarkdown(b []byte) ([]phrasePair, bool) {
	pairs, valid, examined := scanPhrasesMarkdown(b, true)
	if valid < phrasesMinExternal {
		return nil, false
	}
	if examined == 0 || float64(valid)/float64(examined) < phrasesMinValidRatio {
		return nil, false
	}
	return pairs, true
}

// parsePhrasesMarkdown 解析引语；strict 为 true 时逐行做格式校验（内嵌源亦适用）。
func parsePhrasesMarkdown(b []byte, strict bool) []phrasePair {
	pairs, _, _ := scanPhrasesMarkdown(b, strict)
	return pairs
}

func scanPhrasesMarkdown(b []byte, strict bool) (pairs []phrasePair, valid, examined int) {
	for _, line := range strings.Split(string(b), "\n") {
		raw := strings.TrimSpace(line)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		examined++
		p, ok := phraseFromLine(raw, strict)
		if !ok {
			continue
		}
		valid++
		pairs = append(pairs, p)
	}
	return pairs, valid, examined
}

func phraseFromLine(raw string, strict bool) (phrasePair, bool) {
	if !strict {
		return phraseFromLineLenient(raw)
	}
	if strings.Contains(raw, "|") && !strings.Contains(raw, phrasesPipe) {
		return phrasePair{}, false
	}
	if strings.Count(raw, phrasesPipe) != 1 {
		return phrasePair{}, false
	}
	i := strings.Index(raw, phrasesPipe)
	en := strings.TrimSpace(raw[:i])
	zh := strings.TrimSpace(raw[i+len(phrasesPipe):])
	if en == "" || zh == "" {
		return phrasePair{}, false
	}
	if !utf8.ValidString(en) || !utf8.ValidString(zh) {
		return phrasePair{}, false
	}
	if hasDisallowedControl(en) || hasDisallowedControl(zh) {
		return phrasePair{}, false
	}
	enRunes := utf8.RuneCountInString(en)
	zhRunes := utf8.RuneCountInString(zh)
	if enRunes < phrasesMinEnRunes || zhRunes < phrasesMinZhRunes {
		return phrasePair{}, false
	}
	if enRunes > phrasesMaxEnRunes || zhRunes > phrasesMaxZhRunes {
		return phrasePair{}, false
	}
	if !hasLatinLetter(en) || !hasCJK(zh) {
		return phrasePair{}, false
	}
	return phrasePair{En: en, Zh: zh}, true
}

func phraseFromLineLenient(raw string) (phrasePair, bool) {
	i := strings.Index(raw, phrasesPipe)
	if i <= 0 {
		return phrasePair{}, false
	}
	en := strings.TrimSpace(raw[:i])
	zh := strings.TrimSpace(raw[i+len(phrasesPipe):])
	if en == "" || zh == "" {
		return phrasePair{}, false
	}
	return phrasePair{En: en, Zh: zh}, true
}

func hasLatinLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func hasCJK(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func hasDisallowedControl(s string) bool {
	for _, r := range s {
		if r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

func samplePhrasePairs(src []phrasePair, n int) []phrasePair {
	if len(src) == 0 {
		return nil
	}
	if n <= 0 || n >= len(src) {
		out := make([]phrasePair, len(src))
		copy(out, src)
		return out
	}
	idx := rand.Perm(len(src))[:n]
	out := make([]phrasePair, n)
	for i, j := range idx {
		out[i] = src[j]
	}
	return out
}
