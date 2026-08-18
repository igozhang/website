package docindex

import "strings"

// NormalizeSlug accepts letters, digits, spaces, '-' and '_'. Input may start with
// an uppercase letter; result is lowercased with separators normalized to '_'.
func NormalizeSlug(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	var b strings.Builder
	b.Grow(len(raw))
	prevSep := false
	for _, c := range strings.ToLower(raw) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b.WriteRune(c)
			prevSep = false
		case c == ' ', c == '-', c == '_':
			if b.Len() == 0 || prevSep {
				continue
			}
			b.WriteByte('_')
			prevSep = true
		default:
			return "", false
		}
	}
	s := strings.TrimRight(b.String(), "_")
	if !ValidSlug(s) {
		return "", false
	}
	return s, true
}
