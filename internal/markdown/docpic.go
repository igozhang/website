package markdown

// ExtractDocPicRefs returns unique doc-pic filenames referenced in markdown (not including "doc-pic/" prefix).
func ExtractDocPicRefs(src string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, m := range docPicRE.FindAllStringSubmatch(src, -1) {
		if len(m) < 3 {
			continue
		}
		name := m[2]
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
