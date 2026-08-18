package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func visitEvent(ts time.Time, ip, target string) accessLogEvent {
	return accessLogEvent{
		TSUnix: ts.Unix(),
		TS:     formatBeijingUnix(ts.Unix()),
		IP:     ip,
		Action: accessLogActionVisit,
		Target: target,
	}
}

func sessionEvent(ts time.Time, ip, target, sessionID string) accessLogEvent {
	return accessLogEvent{
		TSUnix:    ts.Unix(),
		TS:        formatBeijingUnix(ts.Unix()),
		IP:        ip,
		Action:    accessLogActionSession,
		Target:    target,
		SessionID: sessionID,
	}
}

func TestAppendAccessLogAndStats(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)

	now := time.Now()
	appendAccessLog(dir, sessionEvent(now, "203.0.113.20", "首页", "203.0.113.20:1"))
	appendAccessLog(dir, sessionEvent(now.Add(-10*24*time.Hour), "203.0.113.21", "首页", "203.0.113.21:1"))

	st := countAccessLogStats(dir, now)
	if st.Day != 1 {
		t.Fatalf("day=%d want 1", st.Day)
	}
	if st.Week != 1 {
		t.Fatalf("week=%d want 1", st.Week)
	}
	if st.Month != 2 {
		t.Fatalf("month=%d want 2", st.Month)
	}

	inv := siteLogInventory(dir)
	found := false
	for _, f := range inv {
		if f.Path == secret07RelDir+"/"+accessLogFileName {
			found = true
			if !f.Active {
				t.Fatal("user-access.log should be active")
			}
		}
	}
	if !found {
		t.Fatal("missing user-access.log in inventory")
	}
}

func TestSiteLogInventory_ignoresJunkInSecretLog(t *testing.T) {
	dir := t.TempDir()
	secretLog := filepath.Join(dir, logRelDir)
	if err := os.MkdirAll(secretLog, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secretLog, "readme.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secretLog, appLogFileName), []byte("log"), 0o644); err != nil {
		t.Fatal(err)
	}
	inv := siteLogInventory(dir)
	if len(inv) != 1 || inv[0].Path != logRelDir+"/"+appLogFileName {
		t.Fatalf("inv=%+v", inv)
	}
}

func TestSiteLogBoardFiles_perRolePreview(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	secretLog := filepath.Join(dir, logRelDir)
	secret := secret07Dir(dir)
	if err := os.MkdirAll(secretLog, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"igozhang.2026-06-30.log",
		"igozhang.2026-07-01.log",
		"igozhang.part.100.log",
		"igozhang.part.200.log",
		"igozhang.part.300.log",
	} {
		if err := os.WriteFile(filepath.Join(secretLog, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(secretLog, appLogFileName), []byte("active"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"user-access.2026-06-01.log",
		"user-access.2026-06-02.log",
		"user-access.2026-06-03.log",
		"user-access.2026-01.summary.log",
		"user-access.2026-02.summary.log",
	} {
		if err := os.WriteFile(filepath.Join(secret, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(accessLogPath(dir), []byte("ua"), 0o644); err != nil {
		t.Fatal(err)
	}

	board := siteLogBoardFiles(dir)
	var (
		activeN         int
		appDayArch      int
		appPartArch     int
		uaDayArch       int
		uaMonthArch     int
		appDayMore      int
		appPartMore     int
		uaDayMore       int
		uaMonthMore     int
	)
	for _, f := range board {
		switch {
		case f.Active:
			activeN++
		case f.More > 0:
			switch f.Role {
			case "运行日志日归档":
				appDayMore = f.More
			case "运行日志过大切割":
				appPartMore = f.More
			case "访客访问日归档":
				uaDayMore = f.More
			case "访客访问月归档":
				uaMonthMore = f.More
			}
		case f.Role == "运行日志日归档":
			appDayArch++
		case f.Role == "运行日志过大切割":
			appPartArch++
		case f.Role == "访客访问日归档":
			uaDayArch++
		case f.Role == "访客访问月归档":
			uaMonthArch++
		}
	}
	if activeN != 2 {
		t.Fatalf("active=%d want 2", activeN)
	}
	if appDayArch != 2 || appDayMore != 0 {
		t.Fatalf("app day archive=%d more=%d want 2 preview + 0 more", appDayArch, appDayMore)
	}
	if appPartArch != 2 || appPartMore != 1 {
		t.Fatalf("app part archive=%d more=%d want 2 preview + 1 more", appPartArch, appPartMore)
	}
	if uaDayArch != 2 || uaDayMore != 1 {
		t.Fatalf("ua day archive=%d more=%d want 2 preview + 1 more", uaDayArch, uaDayMore)
	}
	if uaMonthArch != 2 || uaMonthMore != 0 {
		t.Fatalf("ua month archive=%d more=%d want 2 preview + 0 more", uaMonthArch, uaMonthMore)
	}
}

func TestEnforceUserAccessFileSizeCap(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	p := accessLogPath(dir)
	padding := make([]byte, int(userAccessMaxFileBytes)+1)
	for i := range padding {
		padding[i] = 'a'
	}
	if err := os.WriteFile(p, padding, 0o644); err != nil {
		t.Fatal(err)
	}
	enforceUserAccessFileSizeCap(p)
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Size() > userAccessMaxFileBytes {
		t.Fatalf("size=%d want <= %d", fi.Size(), userAccessMaxFileBytes)
	}
}

func TestUserAccessMonthlyCompact(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	secret := secret07Dir(dir)
	month := time.Now().AddDate(0, -2, 0)
	monthKey := userAccessMonthKey(month)
	dayPath := filepath.Join(secret, accessLogArchiveStem+"."+month.Format("2006-01-02")+".log")
	line, _ := json.Marshal(sessionEvent(month, "203.0.113.60", "首页", "203.0.113.60:1"))
	if err := os.WriteFile(dayPath, append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	maintainAccessLogDir(dir)

	summaryPath := userAccessSummaryPath(secret, monthKey)
	if _, err := os.Stat(summaryPath); err != nil {
		t.Fatalf("summary missing: %v", err)
	}
	if _, err := os.Stat(dayPath); !os.IsNotExist(err) {
		t.Fatal("daily archive should be removed after compact")
	}
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) < 2 {
		t.Fatalf("summary lines=%d", len(lines))
	}
	var sum userAccessLogSummary
	if err := json.Unmarshal([]byte(lines[0]), &sum); err != nil {
		t.Fatal(err)
	}
	if sum.Kind != userAccessSummaryKind || sum.UniqueVisitors != 1 || sum.EventCount != 1 {
		t.Fatalf("summary=%+v", sum)
	}
}

func TestBuildUserAccessBundle_prefersActiveOverDaily(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	secret := secret07Dir(dir)
	oldDay := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	oldPath := filepath.Join(secret, accessLogArchiveStem+"."+oldDay+".log")
	if err := os.WriteFile(oldPath, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	appendAccessLog(dir, sessionEvent(time.Now(), "1.2.3.4", "首页", "1.2.3.4:1"))
	raw, err := buildUserAccessBundleBytes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "1.2.3.4") {
		t.Fatalf("bundle should include active log: %q", raw)
	}
}

func TestListRecentAccessLogEvents(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	now := time.Now()
	appendAccessLog(dir, accessLogEvent{
		TSUnix:    now.Unix(),
		TS:        formatBeijingUnix(now.Unix()),
		IP:        "203.0.113.40",
		Client:    "Chrome",
		Action:    accessLogActionSession,
		Target:    "首页",
		SessionID: "203.0.113.40:" + fmt.Sprintf("%d", now.Unix()),
	})
	appendAccessLog(dir, visitEvent(now.Add(-2*24*time.Hour), "203.0.113.41", "igokey"))
	appendAccessLog(dir, visitEvent(now.Add(-10*24*time.Hour), "203.0.113.42", "文件中转"))

	cutDay, _ := accessLogPeriodCutoff("day", now)
	dayEvents := listRecentAccessLogEvents(dir, cutDay, 30)
	if len(dayEvents) != 1 || dayEvents[0].IP != "203.0.113.40" {
		t.Fatalf("day events=%+v", dayEvents)
	}

	cutWeek, _ := accessLogPeriodCutoff("week", now)
	weekEvents := listRecentAccessLogEvents(dir, cutWeek, 30)
	if len(weekEvents) != 2 {
		t.Fatalf("week=%d want 2", len(weekEvents))
	}
	if weekEvents[0].IP != "203.0.113.40" || weekEvents[1].IP != "203.0.113.41" {
		t.Fatalf("week order=%v,%v", weekEvents[0].IP, weekEvents[1].IP)
	}

	cutMonth, _ := accessLogPeriodCutoff("month", now)
	monthEvents := listRecentAccessLogEvents(dir, cutMonth, 2)
	if len(monthEvents) != 2 {
		t.Fatalf("limit=%d want 2", len(monthEvents))
	}

	rows := accessLogEventsToJSON(dayEvents)
	if len(rows) != 1 || rows[0]["ip"] != "203.0.113.40" || rows[0]["target"] != "首页" {
		t.Fatalf("json=%+v", rows)
	}
}

func TestBuildAccessLogEvent_atomic(t *testing.T) {
	now := time.Unix(1719900000, 0)
	e := &visitorEntry{
		ip:           "1.2.3.4",
		geo:          "中国",
		client:       "Chrome",
		sessionStart: now.Add(-10 * time.Minute),
		lastSeen:     now,
	}
	ev := buildAccessLogEvent(e, now, accessLogActionFeature, "mp3", "1.2.3.4:1719900000")
	if ev.Action != accessLogActionFeature || ev.Target != "mp3" || ev.IP != "1.2.3.4" {
		t.Fatalf("ev=%+v", ev)
	}
}

func TestCountAccessLogStats_sessions(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	now := time.Now()
	sid := "203.0.113.30:" + fmt.Sprintf("%d", now.Unix())
	for i := 0; i < 3; i++ {
		appendAccessLog(dir, sessionEvent(now, "203.0.113.30", "首页", sid))
	}
	appendAccessLog(dir, visitEvent(now, "203.0.113.30", "文件中转"))
	st := countAccessLogStats(dir, now)
	if st.Day != 3 || st.Week != 3 || st.Month != 3 {
		t.Fatalf("session stats=%+v want 3/3/3", st)
	}
}

func TestShouldAppendPathAccessLog_dedupe(t *testing.T) {
	accessLogDedupMu.Lock()
	accessLogDedup = make(map[string]time.Time)
	accessLogDedupMu.Unlock()
	now := time.Now()
	sid := "1.2.3.4:100"
	if !shouldAppendPathAccessLog(sid, accessLogActionVisit, "首页", now) {
		t.Fatal("first append")
	}
	if shouldAppendPathAccessLog(sid, accessLogActionVisit, "首页", now.Add(2*time.Second)) {
		t.Fatal("deduped within refresh window")
	}
	if !shouldAppendPathAccessLog(sid, accessLogActionVisit, "首页", now.Add(6*time.Second)) {
		t.Fatal("should append after refresh window")
	}
	if !shouldAppendPathAccessLog(sid, accessLogActionVisit, "文件中转", now.Add(1*time.Second)) {
		t.Fatal("different target should append")
	}
}

func TestFormatAccessActionLabel_session(t *testing.T) {
	if got := formatAccessActionLabel(accessLogActionSession, "首页"); got != "访问 首页" {
		t.Fatalf("got %q", got)
	}
}

func TestListRecentAccessLogSessionDetails(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	now := time.Now()
	sid := "203.0.113.40:" + fmt.Sprintf("%d", now.Unix())
	appendAccessLog(dir, accessLogEvent{
		TSUnix:    now.Unix(),
		TS:        formatBeijingUnix(now.Unix()),
		IP:        "203.0.113.40",
		Action:    accessLogActionSession,
		Target:    "首页",
		SessionID: sid,
	})
	appendAccessLog(dir, accessLogEvent{
		TSUnix:    now.Unix() + 10,
		TS:        formatBeijingUnix(now.Unix() + 10),
		IP:        "203.0.113.40",
		Action:    accessLogActionVisit,
		Target:    "五子棋",
		SessionID: sid,
	})
	cutDay, _ := accessLogPeriodCutoff("day", now)
	rows := listRecentAccessLogSessionDetails(dir, cutDay, 10)
	if len(rows) != 1 {
		t.Fatalf("rows=%+v", rows)
	}
	if rows[0]["label"] != "访问 首页" {
		t.Fatalf("label=%v", rows[0]["label"])
	}
	paths, _ := rows[0]["paths"].([]string)
	if len(paths) != 1 || paths[0] != "五子棋" {
		t.Fatalf("paths=%v", rows[0]["paths"])
	}
}

func TestFormatVisitorTrail(t *testing.T) {
	ts := time.Date(2026, 7, 3, 7, 30, 0, 0, time.UTC).Unix()
	trail := []visitorActionStep{
		{TSUnix: ts, Action: accessLogActionVisit, Target: "首页"},
		{TSUnix: ts + 15, Action: accessLogActionFeature, Target: "mp3"},
		{TSUnix: ts + 62, Action: accessLogActionVisit, Target: "/redis_cfg"},
		{TSUnix: ts + 120, Action: accessLogActionVisit, Target: "五子棋"},
	}
	got := formatVisitorTrail(trail)
	var prevDate string
	wantParts := make([]string, 0, len(trail))
	for _, s := range trail {
		wantParts = append(wantParts, formatVisitorTrailCompactStep(s, &prevDate))
	}
	want := strings.Join(wantParts, " -> ")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if strings.Count(got, "2026-07-03") != 1 {
		t.Fatalf("same-day trail should show date once: %q", got)
	}
	j := visitorTrailToJSON(trail)
	if len(j) != 4 || j[0]["ts"] == "" || j[0]["tsUnix"] == nil {
		t.Fatalf("trail json=%+v", j)
	}
}
