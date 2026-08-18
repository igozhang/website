package markdown

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// unicodeHeadingIDs generates heading ids similar to Goldmark's default,
// but keeps Unicode letters/digits (e.g. Chinese) so anchors like #页内导航 work.
// ASCII punctuation is skipped (so "Role / ClusterRole" stays role--clusterrole).
// When punctuation sits between two word characters with no surrounding spaces
// (e.g. 配额（ResourceQuota), a single "-" is inserted.
type unicodeHeadingIDs struct {
	values map[string]bool
}

func newUnicodeHeadingIDs() parser.IDs {
	return &unicodeHeadingIDs{values: map[string]bool{}}
}

func (s *unicodeHeadingIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	value = trimSpaceBytes(value)
	result := make([]byte, 0, len(value))
	for i := 0; i < len(value); {
		r, size := utf8.DecodeRune(value[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		switch {
		case r <= 0x7f && isASCIIAlphaNum(byte(r)):
			if 'A' <= r && r <= 'Z' {
				r += 'a' - 'A'
			}
			result = append(result, byte(r))
			i += size
		case r <= 0x7f && (r == ' ' || r == '\t' || r == '-' || r == '_'):
			result = append(result, '-')
			i += size
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], unicode.ToLower(r))
			result = append(result, buf[:n]...)
			i += size
		case unicode.IsSpace(r):
			result = append(result, '-')
			i += size
		default:
			// Skip punctuation/symbols. For non-ASCII punct between word chars
			// (e.g. 配额（ResourceQuota), insert "-". ASCII punct like "/" stays
			// Goldmark-compatible (CPU/内存 → cpu内存, Role / X → role--x).
			if r > 0x7f && endsWithWord(result) && nextKeptIsWord(value[i+size:]) {
				result = append(result, '-')
			}
			i += size
		}
	}
	if len(result) == 0 {
		if kind == ast.KindHeading {
			result = []byte("heading")
		} else {
			result = []byte("id")
		}
	}
	key := string(result)
	if !s.values[key] {
		s.values[key] = true
		return result
	}
	for n := 1; ; n++ {
		next := fmt.Sprintf("%s-%d", key, n)
		if !s.values[next] {
			s.values[next] = true
			return []byte(next)
		}
	}
}

func (s *unicodeHeadingIDs) Put(value []byte) {
	s.values[string(value)] = true
}

func isASCIIAlphaNum(v byte) bool {
	return ('a' <= v && v <= 'z') || ('A' <= v && v <= 'Z') || ('0' <= v && v <= '9')
}

func endsWithWord(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	r, _ := utf8.DecodeLastRune(b)
	if r == utf8.RuneError {
		return false
	}
	if r <= 0x7f {
		return isASCIIAlphaNum(byte(r))
	}
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func nextKeptIsWord(b []byte) bool {
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		if r <= 0x7f && isASCIIAlphaNum(byte(r)) {
			return true
		}
		if r <= 0x7f && (r == ' ' || r == '\t' || r == '-' || r == '_') {
			return false
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
		if unicode.IsSpace(r) {
			return false
		}
		// more punctuation — keep scanning
		i += size
	}
	return false
}

func trimSpaceBytes(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n' || b[i] == '\r') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\n' || b[j-1] == '\r') {
		j--
	}
	return b[i:j]
}
