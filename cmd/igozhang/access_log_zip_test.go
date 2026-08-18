package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccessLogFileRelevantForMonth(t *testing.T) {
	cutMonth := time.Now().Add(-30 * 24 * time.Hour).Unix()
	oldDay := time.Now().Add(-40 * 24 * time.Hour).Format("2006-01-02")
	newDay := time.Now().Add(-2 * 24 * time.Hour).Format("2006-01-02")

	if accessLogFileRelevantForMonth(accessLogArchiveStem+"."+oldDay+".log", cutMonth) {
		t.Fatal("old daily archive should be skipped")
	}
	if !accessLogFileRelevantForMonth(accessLogArchiveStem+"."+newDay+".log", cutMonth) {
		t.Fatal("recent daily archive should be scanned")
	}
	if !accessLogFileRelevantForMonth(accessLogFileName, cutMonth) {
		t.Fatal("active log should always be scanned")
	}
	oldMonth := time.Now().AddDate(0, -2, 0).Format("2006-01")
	if accessLogFileRelevantForMonth(accessLogArchiveStem+"."+oldMonth+".summary.log", cutMonth) {
		t.Fatal("old monthly summary should be skipped for 30d stats")
	}
}

func TestCountAccessLogStats_skipsOldDailyArchive(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	secret := secret07Dir(dir)
	now := time.Now()
	oldDay := now.Add(-40 * 24 * time.Hour).Format("2006-01-02")
	oldPath := filepath.Join(secret, accessLogArchiveStem+"."+oldDay+".log")
	oldLine, err := json.Marshal(sessionEvent(now.Add(-40*24*time.Hour), "1.2.3.4", "首页", "1.2.3.4:1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPath, append(oldLine, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	appendAccessLog(dir, sessionEvent(now, "5.6.7.8", "igokey", "5.6.7.8:1"))
	st := countAccessLogStats(dir, now)
	if st.Month != 1 {
		t.Fatalf("month=%d want 1 (old archive skipped)", st.Month)
	}
}
