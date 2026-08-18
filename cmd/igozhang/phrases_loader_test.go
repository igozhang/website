package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestParsePhrasesMarkdown_SkipsComments(t *testing.T) {
	md := `# title
# comment
Hello｜你好
# trailing

World｜世界
`
	got := parsePhrasesMarkdown([]byte(md), false)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if got[0].En != "Hello" || got[1].En != "World" {
		t.Fatalf("%+v", got)
	}
}

func TestLoadPhrasesSource_EmbeddedWhenExternalTooSmall(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(docAssetsEngDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; i < 10; i++ {
		b.WriteString("Line ")
		b.WriteString(strings.Repeat("x", i))
		b.WriteString("｜行")
		b.WriteString(strings.Repeat("x", i))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(docAssetsPhrasesPath(dir), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	pool, _ := buildPhraseServePool(dir)
	empty, _ := buildPhraseServePool(t.TempDir())
	if len(pool) != len(empty) {
		t.Fatalf("external too small: got %d want embedded sample %d", len(pool), len(empty))
	}
}

func TestLoadPhrasesSource_UsesExternalWhenEnough(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(docAssetsEngDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; i < phrasesMinExternal+80; i++ {
		b.WriteString("Ext ")
		b.WriteString(strings.Repeat("e", i%5))
		b.WriteString("｜外")
		b.WriteString(strings.Repeat("文", i%5))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(docAssetsPhrasesPath(dir), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	pool, _ := buildPhraseServePool(dir)
	if len(pool) != phrasesServePoolSize {
		t.Fatalf("len=%d want serve pool %d", len(pool), phrasesServePoolSize)
	}
	if !strings.HasPrefix(pool[0].En, "Ext") {
		t.Fatalf("expected external corpus, got %+v", pool[0])
	}
}

func TestBuildPhraseServePool_Samples200(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(docAssetsEngDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	const total = 300
	for i := 0; i < total; i++ {
		b.WriteString("P")
		b.WriteString(strings.Repeat("q", i%7))
		b.WriteString("｜句")
		b.WriteString(strings.Repeat("子", i%7))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(docAssetsPhrasesPath(dir), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	pool, _ := buildPhraseServePool(dir)
	if len(pool) != phrasesServePoolSize {
		t.Fatalf("pool len=%d want %d", len(pool), phrasesServePoolSize)
	}
}

func TestBuildPhraseServePool_ExternalFlag(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(docAssetsEngDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; i < phrasesMinExternal+10; i++ {
		b.WriteString("Flag ")
		b.WriteString(strings.Repeat("e", i%4))
		b.WriteString("｜标")
		b.WriteString(strings.Repeat("记", i%4))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(docAssetsPhrasesPath(dir), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	_, ext := buildPhraseServePool(dir)
	if !ext {
		t.Fatal("expected external=true")
	}
	initPhrasesCache(dir)
	if !phraseSourceIsExternal() {
		t.Fatal("cache external flag")
	}
	empty, ext2 := buildPhraseServePool(t.TempDir())
	if ext2 || len(empty) == 0 {
		t.Fatalf("empty root: external=%v len=%d", ext2, len(empty))
	}
	initPhrasesCache(t.TempDir())
	if phraseSourceIsExternal() {
		t.Fatal("expected embedded after init on empty dir")
	}
}

func TestInitPhrasesCache_ServesPoolOnly(t *testing.T) {
	initPhrasesCache(t.TempDir())
	bank := phraseBankForUI()
	if len(bank) == 0 {
		t.Fatal("expected phrase pool")
	}
	if len(bank) > phrasesServePoolSize {
		t.Fatalf("pool len=%d want <= %d", len(bank), phrasesServePoolSize)
	}
}

func TestInitPhrasesCache_SiteHeadJSON(t *testing.T) {
	initPhrasesCache(t.TempDir())
	raw, err := jsonMarshalPhraseBank()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 10 {
		t.Fatalf("empty phrase json: %s", raw)
	}
}

func TestParsePhrasesExternal_RejectsASCIIPipe(t *testing.T) {
	md := strings.Repeat("Hello|你好\n", phrasesMinExternal)
	_, ok := parsePhrasesExternalMarkdown([]byte(md))
	if ok {
		t.Fatal("expected reject for ASCII pipe delimiter")
	}
}

func TestParsePhrasesExternal_RejectsNoCJK(t *testing.T) {
	var b strings.Builder
	for i := 0; i < phrasesMinExternal; i++ {
		b.WriteString("Valid English line ")
		b.WriteString(strings.Repeat("x", i%5))
		b.WriteString("｜only latin\n")
	}
	_, ok := parsePhrasesExternalMarkdown([]byte(b.String()))
	if ok {
		t.Fatal("expected reject when zh has no Han characters")
	}
}

func TestParsePhrasesExternal_RejectsLowValidRatio(t *testing.T) {
	var b strings.Builder
	for i := 0; i < phrasesMinExternal; i++ {
		b.WriteString("Good line ")
		b.WriteString(strings.Repeat("e", i%4))
		b.WriteString("｜好句")
		b.WriteString(strings.Repeat("子", i%4))
		b.WriteByte('\n')
	}
	for i := 0; i < 80; i++ {
		b.WriteString("not a phrase line\n")
	}
	_, ok := parsePhrasesExternalMarkdown([]byte(b.String()))
	if ok {
		t.Fatal("expected reject when too many malformed lines")
	}
}

func TestParsePhrasesExternal_AcceptsWellFormedFile(t *testing.T) {
	var b strings.Builder
	for i := 0; i < phrasesMinExternal+20; i++ {
		b.WriteString("Good line ")
		b.WriteString(strings.Repeat("e", i%4))
		b.WriteString("｜好句")
		b.WriteString(strings.Repeat("子", i%4))
		b.WriteByte('\n')
	}
	pairs, ok := parsePhrasesExternalMarkdown([]byte(b.String()))
	if !ok {
		t.Fatal("expected accept for well-formed external file")
	}
	if len(pairs) < phrasesMinExternal {
		t.Fatalf("len=%d", len(pairs))
	}
}

func TestPhraseFromLine_Strict(t *testing.T) {
	_, ok := phraseFromLine("Hi｜你好", true)
	if !ok {
		t.Fatal("expected valid")
	}
	_, ok = phraseFromLine("Hi｜hello", true)
	if ok {
		t.Fatal("zh without Han should fail")
	}
	_, ok = phraseFromLine("你好｜句子", true)
	if ok {
		t.Fatal("en without Latin should fail")
	}
}

func jsonMarshalPhraseBank() (string, error) {
	b, err := json.Marshal(phraseBankForUI())
	if err != nil {
		return "", err
	}
	return string(b), nil
}
