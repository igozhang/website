package markdown

import (
	"strings"
	"testing"
)

func TestValidate_unclosedFence(t *testing.T) {
	v := Validate([]byte("```bash\necho hi\n"))
	if v.OK || len(v.Errors) == 0 {
		t.Fatalf("want error, got %+v", v)
	}
}

func TestValidate_image_noFalsePositiveOnParens(t *testing.T) {
	v := Validate([]byte("说明 ![a](doc-pic/x.png) 见文档 (备注)\n"))
	for _, e := range v.Errors {
		if strings.Contains(e.Message, "图片") {
			t.Fatalf("unexpected image error: %+v", e)
		}
	}
}

func TestValidate_okSample(t *testing.T) {
	v := Validate([]byte("# Title\n\n```go\nfmt.Println(1)\n```\n"))
	if !v.OK {
		t.Fatalf("want ok, got %+v", v)
	}
}

func TestFormatValidationSummary(t *testing.T) {
	s := FormatValidationSummary(ValidationResult{
		OK: false,
		Errors: []ValidationIssue{{Message: "代码块未闭合", Line: 3}},
	})
	if s == "" {
		t.Fatal("expected summary")
	}
}
