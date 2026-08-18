package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
)

func TestBuildUserAccessBundle_includesActiveAndDaily(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	secret := secret07Dir(dir)
	oldDay := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	oldPath := filepath.Join(secret, accessLogArchiveStem+"."+oldDay+".log")
	oldEv := sessionEvent(time.Now().AddDate(0, 0, -3), "1.2.3.4", "首页", "1.2.3.4:1")
	oldLine, err := json.Marshal(oldEv)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPath, append(oldLine, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	appendAccessLog(dir, sessionEvent(time.Now(), "5.6.7.8", "igokey", "5.6.7.8:1"))

	raw, err := buildUserAccessBundleBytes(dir)
	if err != nil {
		t.Fatal(err)
	}
	var bundle userAccessBundle
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.Version != userAccessBundleVersion {
		t.Fatalf("version=%d", bundle.Version)
	}
	if bundle.StatsManifest == nil || bundle.StatsManifest.Day < 1 {
		t.Fatalf("statsManifest=%+v", bundle.StatsManifest)
	}
	if !bundle.DetailSlimmed {
		t.Fatal("expected detailSlimmed")
	}
	if len(bundle.ActiveLog) != 1 || bundle.ActiveLog[0].IP != "5.6.7.8" {
		t.Fatalf("active=%+v", bundle.ActiveLog)
	}
	if len(bundle.DailyArchives[oldDay]) != 1 || bundle.DailyArchives[oldDay][0].IP != "1.2.3.4" {
		t.Fatalf("daily=%+v", bundle.DailyArchives)
	}
}

func TestBuildUserAccessBundle_rejectsCacheOnly(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	cf := accessInsightsCacheFile{
		Version:     accessInsightsCacheVersion,
		BuiltAtUnix: time.Now().Unix(),
		Day:         newAccessInsightsBucket().snapshot(),
	}
	b, err := json.Marshal(cf)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(accessInsightsCachePath(dir), b, 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := buildUserAccessBundleBytes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) > 0 {
		t.Fatalf("cache-only should not export bundle: %q", raw)
	}
}

func TestImportUserAccessBundle_rejectsCacheOnly(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	bundle := userAccessBundle{
		Version: userAccessBundleVersion,
		InsightsCache: &accessInsightsCacheFile{
			Version:     accessInsightsCacheVersion,
			BuiltAtUnix: time.Now().Unix(),
		},
	}
	raw, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := importUserAccessBundle(dir, raw); err == nil {
		t.Fatal("expected error for cache-only bundle")
	}
}

func TestImportUserAccessBundle_clearsStaleLocalLogs(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	appendAccessLog(dir, sessionEvent(time.Now(), "203.0.113.77", "首页", "203.0.113.77:1"))
	raw, err := buildUserAccessBundleBytes(dir)
	if err != nil {
		t.Fatal(err)
	}

	dir2 := t.TempDir()
	initSecret07Files(dir2)
	initAccessLog(dir2)
	appendAccessLog(dir2, sessionEvent(time.Now(), "198.51.100.1", "文件中转", "198.51.100.1:1"))

	if _, err := importUserAccessBundle(dir2, raw); err != nil {
		t.Fatal(err)
	}
	logRaw, err := os.ReadFile(accessLogPath(dir2))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(logRaw, []byte("198.51.100.1")) {
		t.Fatal("stale local IP should be cleared")
	}
	if !bytes.Contains(logRaw, []byte("203.0.113.77")) {
		t.Fatalf("imported IP missing: %q", logRaw)
	}
}

func TestImportUserAccessBundle_setsLastDayMarker(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	ev := sessionEvent(time.Now().AddDate(0, 0, -1), "203.0.113.88", "首页", "203.0.113.88:1")
	rawBundle, err := buildUserAccessBundleBytes(dir)
	if err != nil || len(rawBundle) == 0 {
		appendAccessLog(dir, ev)
		rawBundle, err = buildUserAccessBundleBytes(dir)
	}
	if err != nil || len(rawBundle) == 0 {
		t.Fatalf("bundle: err=%v len=%d", err, len(rawBundle))
	}
	var bundle userAccessBundle
	if err := json.Unmarshal(rawBundle, &bundle); err != nil {
		t.Fatal(err)
	}
	bundle.ActiveLogDay = day
	patched, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}

	dir2 := t.TempDir()
	initSecret07Files(dir2)
	if _, err := importUserAccessBundle(dir2, patched); err != nil {
		t.Fatal(err)
	}
	got := readRotatingLogLastDay(secret07Dir(dir2), accessLogLastDayMarker)
	if got != day {
		t.Fatalf("last-day marker=%q want %q", got, day)
	}
}

func TestCountAccessLogStats_summaryHeaderBoost(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	secret := secret07Dir(dir)
	// 固定「月中」时刻，保证上月与 30 天窗口有足够重叠（月末跑会只剩 1 天重叠导致 boost≈0）。
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.Local)
	prevMonth := time.Date(2026, 6, 1, 12, 0, 0, 0, time.Local)
	monthKey := prevMonth.Format("2006-01")
	sum := userAccessLogSummary{
		Kind:           userAccessSummaryKind,
		Month:          monthKey,
		UniqueVisitors: 80,
		EventCount:     80,
	}
	head, err := json.Marshal(sum)
	if err != nil {
		t.Fatal(err)
	}
	ev := sessionEvent(prevMonth.AddDate(0, 0, 5), "203.0.113.60", "首页", "203.0.113.60:1")
	line, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	body := string(head) + "\n" + string(line) + "\n"
	path := userAccessSummaryPath(secret, monthKey)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	st := countAccessLogStats(dir, now)
	if st.Month < 10 {
		t.Fatalf("month UV=%d want boost from summary header", st.Month)
	}
}

func TestImportUserAccessBundle_restoresBoardInsights(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	appendAccessLog(dir, accessLogEvent{
		TSUnix:    time.Now().Unix(),
		IP:        "203.0.113.99",
		Geo:       "中国 CN",
		Action:    accessLogActionSession,
		Target:    "首页",
		SessionID: "203.0.113.99:1",
	})
	raw, err := buildUserAccessBundleBytes(dir)
	if err != nil {
		t.Fatal(err)
	}

	dir2 := t.TempDir()
	initSecret07Files(dir2)
	accessInsightsMu.Lock()
	accessInsightsDay = newAccessInsightsBucket()
	accessInsightsReady = false
	accessInsightsBuilding = false
	accessInsightsMu.Unlock()

	if _, err := importUserAccessBundle(dir2, raw); err != nil {
		t.Fatal(err)
	}
	active, err := os.ReadFile(accessLogPath(dir2))
	if err != nil || !bytes.Contains(active, []byte("203.0.113.99")) {
		t.Fatalf("active log not restored: %v %q", err, active)
	}
	snap := accessBoardInsightsSnapshot()
	if !snap.Ready || len(snap.Day.GeoTop) != 1 || snap.Day.GeoTop[0].Code != "CN" {
		t.Fatalf("insights not ready from cache: %+v", snap)
	}
}

func TestDocExportZip_accessBundle(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(filepath.Join(assets, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets, "doc-pic"), 0o755)
	st, err := docindex.Open(docDBPath(dir), assets)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.WriteArticle(docindex.Meta{
		Slug: "demo", Title: "Demo", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	initSecret07Files(dir)
	initIPBlacklist(dir)
	if _, err := addIPBlacklist(dir, "203.0.113.50"); err != nil {
		t.Fatal(err)
	}
	initIPWhitelist(dir)
	if err := os.WriteFile(ipWhitelistPath(dir), []byte("# test\n203.0.113.60\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	reloadIPWhitelist(dir)
	initAccessLog(dir)
	appendAccessLog(dir, sessionEvent(time.Now(), "203.0.113.51", "首页", "203.0.113.51:1"))

	zipBytes, err := buildDocExportZip(dir, assets, st)
	if err != nil {
		t.Fatal(err)
	}
	if !zipHasEntry(zipBytes, userAccessBundleZipEntry) {
		t.Fatal("export missing access bundle")
	}
	if zipHasEntry(zipBytes, "secret07/user-access.log") {
		t.Fatal("should not export legacy user-access.log")
	}

	dir2 := t.TempDir()
	assets2 := filepath.Join(dir2, "assets")
	_ = os.MkdirAll(filepath.Join(assets2, "md"), 0o755)
	_ = os.MkdirAll(filepath.Join(assets2, "doc-pic"), 0o755)
	st2, err := docindex.Open(docDBPath(dir2), assets2)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	if err := st2.WriteArticle(docindex.Meta{
		Slug: "demo", Title: "Demo", Author: "a", PublishedAt: time.Now(),
	}, []byte("# Demo\n")); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := importDocZip(dir2, assets2, st2, zr); err != nil {
		t.Fatal(err)
	}
	if !isIPBlacklisted("203.0.113.50") {
		t.Fatal("blacklist not restored")
	}
	if !isIPWhitelisted("203.0.113.60") {
		t.Fatal("whitelist not restored")
	}
	logRaw, err := os.ReadFile(accessLogPath(dir2))
	if err != nil || !bytes.Contains(logRaw, []byte("203.0.113.51")) {
		t.Fatalf("access log not restored via bundle: %v %q", err, logRaw)
	}
}

func TestSlimAccessLogEventsForBundle_keepsSessionsAndCapsPaths(t *testing.T) {
	now := time.Now()
	sid := "203.0.113.1:100"
	var events []accessLogEvent
	events = append(events, sessionEvent(now, "203.0.113.1", "首页", sid))
	for i := 0; i < 20; i++ {
		events = append(events, accessLogEvent{
			TSUnix: now.Add(time.Duration(i) * time.Second).Unix(),
			IP:     "203.0.113.1",
			Action: accessLogActionVisit,
			Target: fmt.Sprintf("/p%d", i),
			SessionID: sid,
		})
	}
	out := slimAccessLogEventsForBundle(events)
	sessions := 0
	paths := 0
	for _, ev := range out {
		if accessLogEventCountsForSessionStats(ev) {
			sessions++
		} else {
			paths++
		}
	}
	if sessions != 1 {
		t.Fatalf("sessions=%d want 1", sessions)
	}
	if paths != userAccessBundleDetailPathsPerSess {
		t.Fatalf("paths=%d want %d", paths, userAccessBundleDetailPathsPerSess)
	}
}

func TestBuildUserAccessBundle_skipsDailyWhenSummaryExists(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	secret := secret07Dir(dir)
	now := time.Now()
	// 用「两月前 1 号」起算，保证 day 与 monthKey 同月（避免 AddDate 月末溢出跨月）。
	base := time.Date(now.Year(), now.Month()-2, 1, 12, 0, 0, 0, now.Location())
	monthKey := base.Format("2006-01")
	day := base.AddDate(0, 0, 4).Format("2006-01-02")
	dailyPath := filepath.Join(secret, accessLogArchiveStem+"."+day+".log")
	ev := sessionEvent(base.AddDate(0, 0, 4), "203.0.113.40", "首页", "203.0.113.40:1")
	line, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dailyPath, append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := userAccessLogSummary{
		Kind: userAccessSummaryKind, Month: monthKey,
		UniqueVisitors: 5, EventCount: 5,
	}
	head, err := json.Marshal(sum)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userAccessSummaryPath(secret, monthKey), append(append(head, '\n'), line...), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := buildUserAccessBundleBytes(dir)
	if err != nil {
		t.Fatal(err)
	}
	var bundle userAccessBundle
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatal(err)
	}
	if len(bundle.DailyArchives) > 0 {
		t.Fatalf("daily should be omitted when summary exists: %+v", bundle.DailyArchives)
	}
	if len(bundle.MonthlyStats) != 1 || bundle.MonthlyStats[0].UniqueVisitors != 5 {
		t.Fatalf("monthly=%+v", bundle.MonthlyStats)
	}
}

func TestImportUserAccessBundle_clearsVisitorBoard(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	appendAccessLog(dir, sessionEvent(time.Now(), "203.0.113.55", "首页", "203.0.113.55:1"))
	raw, err := buildUserAccessBundleBytes(dir)
	if err != nil {
		t.Fatal(err)
	}

	dir2 := t.TempDir()
	initSecret07Files(dir2)
	visitorBoardMu.Lock()
	visitorBoard["203.0.113.99"] = &visitorEntry{ip: "203.0.113.99", lastSeen: time.Now()}
	visitorBoardMu.Unlock()

	if _, err := importUserAccessBundle(dir2, raw); err != nil {
		t.Fatal(err)
	}
	visitorBoardMu.Lock()
	n := len(visitorBoard)
	visitorBoardMu.Unlock()
	if n != 0 {
		t.Fatalf("visitor board should be empty after import, got %d", n)
	}
}

func TestValidateUserAccessBundleStatsManifest_usesExportedAt(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	exportedAt := time.Now().Add(-25 * time.Hour)
	ev := sessionEvent(exportedAt, "203.0.113.70", "首页", "203.0.113.70:1")
	if err := writeAccessLogNDJSON(accessLogPath(dir), []accessLogEvent{ev}); err != nil {
		t.Fatal(err)
	}
	manifest := buildUserAccessBundleStatsManifest(dir, exportedAt, nil)
	if manifest == nil || manifest.Day != 1 || manifest.Week != 1 {
		t.Fatalf("manifest=%+v", manifest)
	}
	if msg := validateUserAccessBundleStatsManifest(dir, manifest, exportedAt); msg != "" {
		t.Fatalf("unexpected mismatch at exportedAt: %s", msg)
	}
	if msg := validateUserAccessBundleStatsManifest(dir, manifest, time.Now()); msg == "" {
		t.Fatal("expected mismatch when validated at later now (day window elapsed)")
	}
}

func TestImportUserAccessBundle_supportsLegacyV2(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	bundle := userAccessBundle{
		Version: userAccessBundleVersionLegacy,
		ActiveLog: []accessLogEvent{
			sessionEvent(time.Now(), "203.0.113.66", "首页", "203.0.113.66:1"),
		},
		InsightsCache: &accessInsightsCacheFile{
			Version:     accessInsightsCacheVersion,
			BuiltAtUnix: time.Now().Unix(),
			Day:         newAccessInsightsBucket().snapshot(),
		},
	}
	raw, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := importUserAccessBundle(dir, raw); err != nil {
		t.Fatal(err)
	}
	logRaw, err := os.ReadFile(accessLogPath(dir))
	if err != nil || !bytes.Contains(logRaw, []byte("203.0.113.66")) {
		t.Fatalf("legacy import failed: %v %q", err, logRaw)
	}
}
