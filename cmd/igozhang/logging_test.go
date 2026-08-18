package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func secretLogDir(t *testing.T, dataRoot string) string {
	t.Helper()
	dir := filepath.Join(dataRoot, logRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeActiveLog(t *testing.T, dataRoot string, content []byte) {
	t.Helper()
	dir := secretLogDir(t, dataRoot)
	if err := os.WriteFile(filepath.Join(dir, appLogFileName), content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTailLines(t *testing.T) {
	raw := []byte("line1\nline2\nline3\n")
	got := string(tailLines(raw, 2))
	if got != "line2\nline3\n" {
		t.Fatalf("tailLines(2) = %q, want line2+line3", got)
	}
	if string(tailLines(raw, 10)) != string(raw) {
		t.Fatal("tailLines beyond count should return all")
	}
}

func TestMaintainSecretLogDir_purgeExpiredArchives(t *testing.T) {
	dataRoot := t.TempDir()
	t.Cleanup(resetActiveLogForTest)
	dir := secretLogDir(t, dataRoot)
	writeActiveLog(t, dataRoot, []byte("active\n"))

	oldArc := filepath.Join(dir, "igozhang.2020-01-01.log")
	if err := os.WriteFile(oldArc, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-20 * 24 * time.Hour)
	if err := os.Chtimes(oldArc, old, old); err != nil {
		t.Fatal(err)
	}

	maintainSecretLogDir(dataRoot)
	if _, err := os.Stat(oldArc); !os.IsNotExist(err) {
		t.Fatal("expired archive should be removed")
	}
}

func TestMaintainSecretLogDir_sizeRotate500K(t *testing.T) {
	dataRoot := t.TempDir()
	t.Cleanup(resetActiveLogForTest)
	dir := secretLogDir(t, dataRoot)
	today := logTodayLocal()
	if err := os.WriteFile(filepath.Join(dir, logLastDayMarker), []byte(today), 0o644); err != nil {
		t.Fatal(err)
	}

	padding := bytes.Repeat([]byte("x"), int(logRotateSizeBytes)+1)
	writeActiveLog(t, dataRoot, padding)

	maintainSecretLogDir(dataRoot)

	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var parts []string
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), "igozhang.part.") {
			parts = append(parts, e.Name())
		}
	}
	if len(parts) != 1 {
		t.Fatalf("expected one part archive, got %v", parts)
	}
	fi, err := os.Stat(filepath.Join(dir, appLogFileName))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Size() != 0 {
		t.Fatalf("active log should be empty after rotate, size=%d", fi.Size())
	}
}

func TestMaintainSecretLogDir_truncate1M_clearsArchives(t *testing.T) {
	dataRoot := t.TempDir()
	t.Cleanup(resetActiveLogForTest)
	dir := secretLogDir(t, dataRoot)
	today := logTodayLocal()
	if err := os.WriteFile(filepath.Join(dir, logLastDayMarker), []byte(today), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "igozhang.2026-01-01.log"), []byte("arc\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var b strings.Builder
	for i := 0; i < 200; i++ {
		b.WriteString(strings.Repeat("a", 6000))
		b.WriteByte('\n')
	}
	writeActiveLog(t, dataRoot, []byte(b.String()))

	maintainSecretLogDir(dataRoot)

	if _, err := os.Stat(filepath.Join(dir, "igozhang.2026-01-01.log")); !os.IsNotExist(err) {
		t.Fatal("archives should be cleared after truncate")
	}
	raw, err := os.ReadFile(filepath.Join(dir, appLogFileName))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) > logTruncateKeepLines {
		t.Fatalf("truncated log has %d lines, want <= %d", len(lines), logTruncateKeepLines)
	}
	if fi, _ := os.Stat(filepath.Join(dir, appLogFileName)); fi.Size() >= logTruncateSizeBytes {
		t.Fatalf("active log still too large: %d", fi.Size())
	}
}

func TestMaintainSecretLogDir_dailyRotate(t *testing.T) {
	dataRoot := t.TempDir()
	t.Cleanup(resetActiveLogForTest)
	dir := secretLogDir(t, dataRoot)
	yesterday := time.Now().Add(-25 * time.Hour).Format("2006-01-02")
	if err := os.WriteFile(filepath.Join(dir, logLastDayMarker), []byte(yesterday), 0o644); err != nil {
		t.Fatal(err)
	}
	writeActiveLog(t, dataRoot, []byte("daily\n"))

	maintainSecretLogDir(dataRoot)

	arc := filepath.Join(dir, "igozhang."+yesterday+".log")
	if _, err := os.Stat(arc); err != nil {
		t.Fatalf("missing daily archive %s: %v", arc, err)
	}
	raw, _ := os.ReadFile(arc)
	if string(raw) != "daily\n" {
		t.Fatalf("archive content = %q", raw)
	}
	got := readLogLastDay(dataRoot)
	if got != logTodayLocal() {
		t.Fatalf(".last-day = %q, want %q", got, logTodayLocal())
	}
}
