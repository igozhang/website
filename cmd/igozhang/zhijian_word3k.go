package main

import (
	_ "embed"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	word3kFileName          = "word3k-full.md"
	word3kSourceHeader      = "X-Fs-Word3k-Source"
	word3kSourceEmbed       = "embedded"
	word3kSourceExternal    = "external"
	word3kPipe              = "\uFF5C"
	word3kMinExternalLines  = 50
	word3kSourceMaxLines    = 1000 // 工作集上限，防止 word3k_pool_size 误配
	word3kSourcePoolFactor  = 10   // 工作集 = pool_size * 10（不超过 word3kSourceMaxLines）
)

// 内嵌精简副本（约 50 条）；完整词库见 DATA_ROOT/assets/eng/word3k-full.md。
// Pod 启动时从文件随机蓄水池抽样至多 min(1000, pool×10) 条入内存（不足则全读）；不载入全集。
// 知键每次刷新页面请求 API 时，再从该工作集随机抽 word3k_pool_size 条返回。
// 重新生成内嵌：node scripts/gen-word3k-embed.mjs
//
//go:embed word3k-embed.md
var word3kEmbedded []byte

var word3kHeadRE = regexp.MustCompile(`^(\S+)\s+(\/.+?\/)\s+(.+)$`)

type word3kPayload struct {
	body     []byte
	external bool
}

type word3kSource struct {
	lines    []string
	external bool
}

var (
	word3kSourceCache word3kSource
	word3kCacheMu     sync.RWMutex
	word3kRoot        string
)

func initWord3kCache(dataRoot string) {
	word3kRoot = dataRoot
	reloadWord3kSource()
}

func reloadWord3kCache() {
	reloadWord3kSource()
}

func reloadWord3kSource() {
	if word3kRoot == "" {
		return
	}
	word3kCacheMu.Lock()
	word3kSourceCache = loadWord3kSource(word3kRoot)
	word3kCacheMu.Unlock()
}

func word3kPoolSize() int {
	n := getOpsSnapshot().zhijianWord3kPoolSize
	if n < 10 {
		return 80
	}
	if n > 500 {
		return 500
	}
	return n
}

// word3kSourceCap 内存工作集行数：pool_size×10，至多 1000；文件有效行不足 cap 时全部保留。
func word3kSourceCap(poolN int) int {
	if poolN < 1 {
		poolN = 1
	}
	cap := poolN * word3kSourcePoolFactor
	if cap > word3kSourceMaxLines {
		cap = word3kSourceMaxLines
	}
	return cap
}

// loadWord3kSource 读盘一次并蓄水池随机抽样入内存工作集（非全集）。
func loadWord3kSource(dataRoot string) word3kSource {
	poolN := word3kPoolSize()
	cap := word3kSourceCap(poolN)
	if dataRoot != "" {
		path := docAssetsWord3kPath(dataRoot)
		if b, ok := secret07ReadFileOrMissing(path, []byte(assetsWord3kTemplate)); ok {
			if !utf8.Valid(b) {
				log.Printf("word3k: %s: not valid UTF-8, using embedded", path)
			} else if lines, ok := sampleWord3kSourceFromMarkdown(b, cap, true); ok {
				return word3kSource{lines: lines, external: true}
			} else {
				log.Printf("word3k: %s: insufficient valid lines, using embedded", path)
			}
		}
	}
	lines, _ := sampleWord3kSourceFromMarkdown(word3kEmbedded, cap, false)
	return word3kSource{lines: lines, external: false}
}

// sampleWord3kSourceFromMarkdown 单次扫描文件，蓄水池随机保留至多 cap 条有效行。
func sampleWord3kSourceFromMarkdown(b []byte, cap int, external bool) ([]string, bool) {
	if cap < 1 {
		cap = 1
	}
	reservoir := make([]string, 0, cap)
	validTotal := 0
	for _, line := range strings.Split(string(b), "\n") {
		raw := strings.TrimSpace(line)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		if !isValidWord3kLine(raw) {
			continue
		}
		validTotal++
		reservoirAppend(&reservoir, raw, validTotal, cap)
	}
	if external && validTotal < word3kMinExternalLines {
		return nil, false
	}
	if validTotal == 0 {
		return nil, false
	}
	return reservoir, true
}

func reservoirAppend(reservoir *[]string, item string, seen, cap int) {
	if cap <= 0 {
		return
	}
	if seen <= cap {
		*reservoir = append(*reservoir, item)
		return
	}
	j := rand.Intn(seen)
	if j < cap {
		(*reservoir)[j] = item
	}
}

func word3kSampleForRequest() word3kPayload {
	word3kCacheMu.RLock()
	src := word3kSourceCache
	word3kCacheMu.RUnlock()
	if len(src.lines) == 0 {
		reloadWord3kSource()
		word3kCacheMu.RLock()
		src = word3kSourceCache
		word3kCacheMu.RUnlock()
	}
	if len(src.lines) == 0 {
		return word3kPayload{}
	}
	picked := sampleStrings(src.lines, word3kPoolSize())
	return word3kPayload{body: encodeWord3kLines(picked), external: src.external}
}

func encodeWord3kLines(lines []string) []byte {
	if len(lines) == 0 {
		return nil
	}
	var sb strings.Builder
	for i, line := range lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
	}
	sb.WriteByte('\n')
	return []byte(sb.String())
}

func isValidWord3kLine(raw string) bool {
	parts := strings.Split(raw, word3kPipe)
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			return false
		}
	}
	if !word3kHeadRE.MatchString(strings.TrimSpace(parts[0])) {
		return false
	}
	return word3kHasPracticeLetters(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]))
}

func word3kHasPracticeLetters(parts ...string) bool {
	n := 0
	for _, s := range parts {
		for _, r := range s {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				n++
			}
		}
	}
	return n > 0
}

func sampleStrings(src []string, n int) []string {
	if len(src) == 0 {
		return nil
	}
	if n <= 0 || n >= len(src) {
		out := make([]string, len(src))
		copy(out, src)
		return out
	}
	idx := rand.Perm(len(src))[:n]
	out := make([]string, n)
	for i, j := range idx {
		out[i] = src[j]
	}
	return out
}

func word3kHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		p := word3kSampleForRequest()
		if len(p.body) == 0 {
			http.Error(w, "word3k corpus unavailable", http.StatusServiceUnavailable)
			return
		}
		src := word3kSourceEmbed
		if p.external {
			src = word3kSourceExternal
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set(word3kSourceHeader, src)
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Length", strconv.Itoa(len(p.body)))
			return
		}
		_, _ = w.Write(p.body)
	}
}
