package markdown

import "testing"

func TestExtractDocPicRefs_unique(t *testing.T) {
	md := "![](doc-pic/a.png)\n![x](doc-pic/b.jpg)\n![](doc-pic/a.png)"
	refs := ExtractDocPicRefs(md)
	if len(refs) != 2 || refs[0] != "a.png" || refs[1] != "b.jpg" {
		t.Fatalf("refs=%v", refs)
	}
}
