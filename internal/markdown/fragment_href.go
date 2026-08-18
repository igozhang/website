package markdown

import (
	"bytes"
	"regexp"
)

// fragmentHrefRE matches href="#..." (fragment-only links).
// Used to counteract <base href="/"> which would resolve #id to origin/#id (homepage).
var fragmentHrefRE = regexp.MustCompile(`href="#([^"]*)"`)

// RewriteFragmentHrefs rewrites href="#id" to href="{pagePath}#id" so in-page
// anchors work when the document has <base href="/"> or a BASE_PATH base tag.
// pagePath should be an absolute path without query/hash, e.g. "/k8s_crud".
func RewriteFragmentHrefs(html []byte, pagePath string) []byte {
	pagePath = normalizePagePath(pagePath)
	if pagePath == "" || !bytes.Contains(html, []byte(`href="#`)) {
		return html
	}
	return fragmentHrefRE.ReplaceAllFunc(html, func(m []byte) []byte {
		sub := fragmentHrefRE.FindSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		frag := sub[1]
		// Keep bare "#" (often unused) unchanged.
		if len(frag) == 0 {
			return m
		}
		out := make([]byte, 0, len(pagePath)+len(frag)+8)
		out = append(out, `href="`...)
		out = append(out, pagePath...)
		out = append(out, '#')
		out = append(out, frag...)
		out = append(out, '"')
		return out
	})
}

func normalizePagePath(p string) string {
	p = string(bytes.TrimSpace([]byte(p)))
	if p == "" {
		return ""
	}
	// strip query/hash if callers pass a full URL path by mistake
	if i := bytes.IndexByte([]byte(p), '?'); i >= 0 {
		p = p[:i]
	}
	if i := bytes.IndexByte([]byte(p), '#'); i >= 0 {
		p = p[:i]
	}
	if p == "" {
		return ""
	}
	if p[0] != '/' {
		p = "/" + p
	}
	// avoid trailing slash except root
	for len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}
