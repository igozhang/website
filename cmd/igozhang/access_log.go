package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	accessLogFileName              = "user-access.log"
	accessLogLastDayMarker         = "user-access.log.last-day"
	accessLogArchiveStem           = "user-access"
	userAccessRecentFullDays       = 30
	userAccessStatsRetention       = 365 * 24 * time.Hour
	userAccessMonthlySampleLines   = 200
	userAccessMaxFileBytes         = 5 * 1024 * 1024
	userAccessTruncateKeepLines    = 300
	accessLogRefreshDedupe         = 5 * time.Second
	accessLogDetailLimit           = 30
)

const userAccessSummaryKind = "user-access-summary"

const (
	accessLogActionSession = "session"
	accessLogActionVisit   = "visit"
	accessLogActionFeature = "feature"
)

type accessLogEvent struct {
	TSUnix    int64  `json:"tsUnix"`
	TS        string `json:"ts"`
	IP        string `json:"ip"`
	Geo       string `json:"geo,omitempty"`
	Client    string `json:"client,omitempty"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	SessionID string `json:"sessionId,omitempty"`
}

type userAccessLogSummary struct {
	Kind           string `json:"kind"`
	Month          string `json:"month"`
	UniqueVisitors int    `json:"uniqueVisitors"`
	EventCount     int    `json:"eventCount"`
}

type accessLogStats struct {
	Day   int `json:"day"`
	Week  int `json:"week"`
	Month int `json:"month"`
}

type siteLogFileInfo struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	Active bool   `json:"active,omitempty"`
	More   int    `json:"more,omitempty"` // 看板折叠：该 role 另有若干归档未列出
}

const siteLogArchivePreviewPerRole = 2

var siteLogArchiveRoleOrder = []string{
	"运行日志日归档",
	"运行日志过大切割",
	"访客访问日归档",
	"访客访问月归档",
}

var (
	accessLogMu      sync.Mutex
	accessLogRoot    string
	accessLogDedupMu sync.Mutex
	accessLogDedup   = make(map[string]time.Time)
)

func initAccessLog(dataRoot string) {
	accessLogMu.Lock()
	defer accessLogMu.Unlock()
	accessLogRoot = dataRoot
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	initAccessInsights(dataRoot)
}

func accessLogPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), accessLogFileName)
}

func accessLogPathDedupeKey(sessionID, action, target string) string {
	return sessionID + "\x00" + action + "\x00" + target
}

func resetAccessLogDedupForImport() {
	accessLogDedupMu.Lock()
	defer accessLogDedupMu.Unlock()
	accessLogDedup = make(map[string]time.Time)
}

func shouldAppendPathAccessLog(sessionID, action, target string, now time.Time) bool {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false
	}
	key := accessLogPathDedupeKey(sessionID, action, target)
	accessLogDedupMu.Lock()
	defer accessLogDedupMu.Unlock()
	pruneAccessLogDedupLocked(now)
	if t, ok := accessLogDedup[key]; ok && now.Sub(t) < accessLogRefreshDedupe {
		return false
	}
	accessLogDedup[key] = now
	return true
}

func pruneAccessLogDedupLocked(now time.Time) {
	cutoff := now.Add(-time.Hour)
	for key, t := range accessLogDedup {
		if t.Before(cutoff) {
			delete(accessLogDedup, key)
		}
	}
}

func accessLogEventCountsForStats(ev accessLogEvent) bool {
	return strings.TrimSpace(ev.Action) != "" && strings.TrimSpace(ev.Target) != ""
}

func accessLogEventCountsForSessionStats(ev accessLogEvent) bool {
	return strings.TrimSpace(ev.Action) == accessLogActionSession && strings.TrimSpace(ev.Target) != ""
}

func formatSessionID(ip string, sessionStart time.Time) string {
	ip = strings.TrimSpace(ip)
	if ip == "" || sessionStart.IsZero() {
		return ""
	}
	return ip + ":" + fmt.Sprintf("%d", sessionStart.Unix())
}

// formatAccessActionLabel 将 action+target 格式化为看板展示用短标签。
func formatAccessActionLabel(action, target string) string {
	action = strings.TrimSpace(action)
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}
	switch action {
	case accessLogActionSession:
		return "访问 " + target
	case accessLogActionFeature, accessLogActionVisit:
		return target
	default:
		return target
	}
}

func appendAccessLog(dataRoot string, ev accessLogEvent) {
	if dataRoot == "" || strings.TrimSpace(ev.IP) == "" {
		return
	}
	accessLogMu.Lock()
	defer accessLogMu.Unlock()
	if accessLogRoot == "" {
		accessLogRoot = dataRoot
	}
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	b, err := json.Marshal(ev)
	if err != nil {
		return
	}
	p := accessLogPath(dataRoot)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
	_ = f.Close()
	enforceUserAccessFileSizeCap(p)
	if accessLogEventCountsForSessionStats(ev) {
		recordAccessInsightsEvent(ev)
	}
}

func enforceUserAccessFileSizeCap(path string) {
	fi, err := os.Stat(path)
	if err != nil || fi.Size() <= userAccessMaxFileBytes {
		return
	}
	if err := truncateRotatingLogFile(path, userAccessTruncateKeepLines); err != nil {
		log.Printf("user-access: truncate %s: %v", path, err)
		return
	}
	if fi2, err := os.Stat(path); err == nil && fi2.Size() > userAccessMaxFileBytes {
		raw, err := os.ReadFile(path)
		if err != nil {
			return
		}
		keep := int(userAccessMaxFileBytes)
		if len(raw) > keep {
			_ = os.WriteFile(path, raw[len(raw)-keep:], 0o644)
			log.Printf("user-access: hard-truncated oversize path=%s max=%d", path, userAccessMaxFileBytes)
		}
		return
	}
	log.Printf("user-access: truncated oversize path=%s keep_lines=%d", path, userAccessTruncateKeepLines)
}

func maintainAccessLogDir(dataRoot string) {
	dir := secret07Dir(dataRoot)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return
	}
	now := time.Now()
	enforceUserAccessArchivesSizeCap(dir)
	maintainUserAccessActiveLog(dir, now)
	compactUserAccessMonthlyArchives(dir, now)
	purgeExpiredUserAccessMonthlySummaries(dir, now.Add(-userAccessStatsRetention))
	reconcileAccessInsights(dataRoot)
}

func enforceUserAccessArchivesSizeCap(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == accessLogFileName {
			enforceUserAccessFileSizeCap(filepath.Join(dir, name))
			continue
		}
		if _, ok := parseUserAccessDailyArchiveDay(name); ok {
			enforceUserAccessFileSizeCap(filepath.Join(dir, name))
		}
	}
}

func maintainUserAccessActiveLog(dir string, now time.Time) {
	fp := filepath.Join(dir, accessLogFileName)
	enforceUserAccessFileSizeCap(fp)
	fi, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			writeRotatingLogLastDay(dir, accessLogLastDayMarker, logTodayLocal())
		}
		return
	}

	today := logTodayLocal()
	lastDay := readRotatingLogLastDay(dir, accessLogLastDayMarker)
	if lastDay == "" {
		writeRotatingLogLastDay(dir, accessLogLastDayMarker, today)
		lastDay = today
	}
	if lastDay != today {
		if fi.Size() > 0 {
			arc := fmt.Sprintf("%s.%s.log", accessLogArchiveStem, lastDay)
			if err := rotateRotatingLogFile(dir, accessLogFileName, arc); err != nil {
				log.Printf("user-access: daily rotate: %v", err)
			} else {
				enforceUserAccessFileSizeCap(filepath.Join(dir, arc))
			}
		}
		writeRotatingLogLastDay(dir, accessLogLastDayMarker, today)
	}
}

func userAccessMonthKey(t time.Time) string {
	return t.Format("2006-01")
}

func userAccessMonthStart(monthKey string) (time.Time, bool) {
	t, err := time.ParseInLocation("2006-01", monthKey, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func userAccessMonthEnd(monthKey string) (time.Time, bool) {
	start, ok := userAccessMonthStart(monthKey)
	if !ok {
		return time.Time{}, false
	}
	return start.AddDate(0, 1, 0).Add(-time.Second), true
}

func userAccessMonthCompactable(monthKey string, now time.Time) bool {
	end, ok := userAccessMonthEnd(monthKey)
	if !ok {
		return false
	}
	return now.After(end.Add(userAccessRecentFullDays * 24 * time.Hour))
}

func userAccessSummaryPath(dir, monthKey string) string {
	return filepath.Join(dir, fmt.Sprintf("%s.%s.summary.log", accessLogArchiveStem, monthKey))
}

func compactUserAccessMonthlyArchives(dir string, now time.Time) {
	months := collectUserAccessDailyArchiveMonths(dir)
	for monthKey := range months {
		if !userAccessMonthCompactable(monthKey, now) {
			continue
		}
		summaryPath := userAccessSummaryPath(dir, monthKey)
		if _, err := os.Stat(summaryPath); err == nil {
			removeUserAccessDailyArchivesForMonth(dir, monthKey)
			continue
		}
		if err := writeUserAccessMonthlySummary(dir, monthKey); err != nil {
			log.Printf("user-access: compact %s: %v", monthKey, err)
			continue
		}
		removeUserAccessDailyArchivesForMonth(dir, monthKey)
	}
}

func collectUserAccessDailyArchiveMonths(dir string) map[string]struct{} {
	out := make(map[string]struct{})
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if day, ok := parseUserAccessDailyArchiveDay(e.Name()); ok {
			out[userAccessMonthKey(day)] = struct{}{}
		}
	}
	return out
}

func parseUserAccessDailyArchiveDay(name string) (time.Time, bool) {
	if !strings.HasPrefix(name, accessLogArchiveStem+".") || !strings.HasSuffix(name, ".log") {
		return time.Time{}, false
	}
	mid := strings.TrimSuffix(strings.TrimPrefix(name, accessLogArchiveStem+"."), ".log")
	if strings.Contains(mid, ".") {
		return time.Time{}, false
	}
	if len(mid) != 10 || mid[4] != '-' || mid[7] != '-' {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", mid, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func userAccessDailyArchivePathsForMonth(dir, monthKey string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		day, ok := parseUserAccessDailyArchiveDay(e.Name())
		if !ok || userAccessMonthKey(day) != monthKey {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)
	return paths
}

func removeUserAccessDailyArchivesForMonth(dir, monthKey string) {
	for _, p := range userAccessDailyArchivePathsForMonth(dir, monthKey) {
		if err := os.Remove(p); err != nil {
			log.Printf("user-access: remove daily %s: %v", p, err)
		}
	}
}

func writeUserAccessMonthlySummary(dir, monthKey string) error {
	var events []accessLogEvent
	for _, p := range userAccessDailyArchivePathsForMonth(dir, monthKey) {
		collectAccessLogEventsFromFile(p, 0, &events)
	}
	if len(events) == 0 {
		return nil
	}
	sessionCount := 0
	for _, ev := range events {
		if accessLogEventCountsForSessionStats(ev) {
			sessionCount++
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].TSUnix > events[j].TSUnix
	})
	sample := events
	if len(sample) > userAccessMonthlySampleLines {
		sample = sample[:userAccessMonthlySampleLines]
	}
	summary := userAccessLogSummary{
		Kind:           userAccessSummaryKind,
		Month:          monthKey,
		UniqueVisitors: sessionCount,
		EventCount:     len(events),
	}
	head, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	var body strings.Builder
	body.Write(head)
	body.WriteByte('\n')
	for _, ev := range sample {
		b, err := json.Marshal(ev)
		if err != nil {
			continue
		}
		body.Write(b)
		body.WriteByte('\n')
	}
	path := userAccessSummaryPath(dir, monthKey)
	if err := os.WriteFile(path, []byte(body.String()), 0o644); err != nil {
		return err
	}
	enforceUserAccessFileSizeCap(path)
	return nil
}

func purgeExpiredUserAccessMonthlySummaries(dir string, cutoff time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		monthKey, ok := parseUserAccessMonthlySummaryName(e.Name())
		if !ok {
			continue
		}
		end, ok := userAccessMonthEnd(monthKey)
		if !ok || !end.Before(cutoff) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := os.Remove(path); err != nil {
			log.Printf("user-access: remove summary %s: %v", path, err)
		} else {
			log.Printf("user-access: delete reason=stats_ttl path=%s", path)
		}
	}
}

func parseUserAccessMonthlySummaryName(name string) (string, bool) {
	if !strings.HasPrefix(name, accessLogArchiveStem+".") || !strings.HasSuffix(name, ".summary.log") {
		return "", false
	}
	mid := strings.TrimSuffix(strings.TrimPrefix(name, accessLogArchiveStem+"."), ".summary.log")
	if len(mid) != 7 || mid[4] != '-' {
		return "", false
	}
	if _, err := time.ParseInLocation("2006-01", mid, time.Local); err != nil {
		return "", false
	}
	return mid, true
}

func accessLogStatsCutoffs(now time.Time) (cutDay, cutWeek, cutMonth int64) {
	return now.Add(-24 * time.Hour).Unix(),
		now.Add(-7 * 24 * time.Hour).Unix(),
		now.Add(-30 * 24 * time.Hour).Unix()
}

func accessLogFileRelevantSince(path string, sinceUnix int64) bool {
	name := filepath.Base(path)
	if name == accessLogFileName {
		return true
	}
	if day, ok := parseUserAccessDailyArchiveDay(name); ok {
		return day.Unix()+86400 >= sinceUnix
	}
	if monthKey, ok := parseUserAccessMonthlySummaryName(name); ok {
		end, ok := userAccessMonthEnd(monthKey)
		if !ok {
			return false
		}
		return end.Unix() >= sinceUnix
	}
	return false
}

func accessLogFileRelevantForMonth(path string, cutMonth int64) bool {
	name := filepath.Base(path)
	if name == accessLogFileName {
		return true
	}
	if day, ok := parseUserAccessDailyArchiveDay(name); ok {
		return day.Unix()+86400 >= cutMonth
	}
	if monthKey, ok := parseUserAccessMonthlySummaryName(name); ok {
		end, ok := userAccessMonthEnd(monthKey)
		if !ok {
			return false
		}
		return end.Unix() >= cutMonth
	}
	return false
}

func isAccessLogFileName(name string) bool {
	if name == accessLogFileName {
		return true
	}
	if name == accessLogLastDayMarker {
		return false
	}
	if _, ok := parseUserAccessDailyArchiveDay(name); ok {
		return true
	}
	_, ok := parseUserAccessMonthlySummaryName(name)
	return ok
}

func accessLogFiles(dataRoot string) []string {
	dir := secret07Dir(dataRoot)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if isAccessLogFileName(e.Name()) {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	return files
}

// clearAccessLogsInImportWindow 导入 bundle 前清理保留窗口内的本地访客 log，避免与 ZIP 数据合并污染看板。
func clearAccessLogsInImportWindow(dataRoot string) error {
	now := time.Now()
	_, _, cutMonth := accessLogStatsCutoffs(now)
	statsCutoff := now.Add(-userAccessStatsRetention)
	dir := secret07Dir(dataRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	_ = os.Remove(accessLogPath(dataRoot))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if day, ok := parseUserAccessDailyArchiveDay(name); ok {
			if day.Unix()+86400 >= cutMonth {
				if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
			continue
		}
		if monthKey, ok := parseUserAccessMonthlySummaryName(name); ok {
			end, ok := userAccessMonthEnd(monthKey)
			if ok && !end.Before(statsCutoff) {
				if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
		}
	}
	_ = os.Remove(accessInsightsCachePath(dataRoot))
	return nil
}

func readUserAccessSummaryHeader(path string) (userAccessLogSummary, bool) {
	f, err := os.Open(path)
	if err != nil {
		return userAccessLogSummary{}, false
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return userAccessLogSummary{}, false
	}
	line := strings.TrimSpace(sc.Text())
	if line == "" || !isUserAccessSummaryLine(line) {
		return userAccessLogSummary{}, false
	}
	var sum userAccessLogSummary
	if err := json.Unmarshal([]byte(line), &sum); err != nil || sum.Kind != userAccessSummaryKind {
		return userAccessLogSummary{}, false
	}
	return sum, true
}

// summaryVisitorBoostInWindow 月 summary 仅保留样本行时，用 header.uniqueVisitors 按窗口重叠比例补足 UV。
func summaryVisitorBoostInWindow(header userAccessLogSummary, monthKey string, sinceUnix int64, now time.Time, sampleUnique int) int {
	if header.UniqueVisitors <= sampleUnique {
		return 0
	}
	start, ok := userAccessMonthStart(monthKey)
	if !ok {
		return 0
	}
	end, ok := userAccessMonthEnd(monthKey)
	if !ok {
		return 0
	}
	windowStart := time.Unix(sinceUnix, 0).In(time.Local)
	overlapStart := start
	if windowStart.After(overlapStart) {
		overlapStart = windowStart
	}
	overlapEnd := now.In(time.Local)
	if end.Before(overlapEnd) {
		overlapEnd = end
	}
	if !overlapEnd.After(overlapStart) {
		return 0
	}
	monthDays := int(end.Sub(start).Hours()/24) + 1
	overlapDays := int(overlapEnd.Sub(overlapStart).Hours()/24) + 1
	if monthDays <= 0 || overlapDays <= 0 {
		return 0
	}
	estimated := header.UniqueVisitors * overlapDays / monthDays
	if estimated <= sampleUnique {
		return 0
	}
	return estimated - sampleUnique
}

func countAccessLogFileSessionMaps(path string, cutDay, cutWeek, cutMonth int64) (dayCount, weekCount, monthCount int) {
	return collectAccessLogFileSessionCount(path, cutDay),
		collectAccessLogFileSessionCount(path, cutWeek),
		collectAccessLogFileSessionCount(path, cutMonth)
}

func countAccessLogStats(dataRoot string, now time.Time) accessLogStats {
	cutDay, cutWeek, cutMonth := accessLogStatsCutoffs(now)
	dayCount, weekCount, monthCount := 0, 0, 0
	var dayBoost, weekBoost, monthBoost int
	for _, path := range accessLogFiles(dataRoot) {
		if !accessLogFileRelevantForMonth(path, cutMonth) {
			continue
		}
		fileDay, fileWeek, fileMonth := countAccessLogFileSessionMaps(path, cutDay, cutWeek, cutMonth)
		dayCount += fileDay
		weekCount += fileWeek
		monthCount += fileMonth
		if monthKey, ok := parseUserAccessMonthlySummaryName(filepath.Base(path)); ok {
			header, ok := readUserAccessSummaryHeader(path)
			if !ok {
				continue
			}
			dayBoost += summaryVisitorBoostInWindow(header, monthKey, cutDay, now, fileDay)
			weekBoost += summaryVisitorBoostInWindow(header, monthKey, cutWeek, now, fileWeek)
			monthBoost += summaryVisitorBoostInWindow(header, monthKey, cutMonth, now, fileMonth)
		}
	}
	return accessLogStats{
		Day:   dayCount + dayBoost,
		Week:  weekCount + weekBoost,
		Month: monthCount + monthBoost,
	}
}

func collectAccessLogFileSessionCount(path string, sinceUnix int64) int {
	n := 0
	collectAccessLogFileSessions(path, sinceUnix, func() { n++ })
	return n
}

func collectAccessLogFileSessions(path string, sinceUnix int64, add func()) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 4096), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if isUserAccessSummaryLine(line) {
			continue
		}
		var ev accessLogEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		if !accessLogEventCountsForSessionStats(ev) {
			continue
		}
		ts := ev.TSUnix
		if ts <= 0 || ts < sinceUnix {
			continue
		}
		add()
	}
}

func isUserAccessSummaryLine(line string) bool {
	return strings.Contains(line, `"kind":`) && strings.Contains(line, userAccessSummaryKind)
}

func accessLogPeriodCutoff(period string, now time.Time) (int64, bool) {
	cutDay, cutWeek, cutMonth := accessLogStatsCutoffs(now)
	switch period {
	case "day":
		return cutDay, true
	case "week":
		return cutWeek, true
	case "month":
		return cutMonth, true
	default:
		return 0, false
	}
}

func listRecentAccessLogEvents(dataRoot string, sinceUnix int64, limit int) []accessLogEvent {
	if limit <= 0 {
		limit = accessLogDetailLimit
	}
	if limit > 100 {
		limit = 100
	}
	var all []accessLogEvent
	for _, path := range accessLogFiles(dataRoot) {
		if !accessLogFileRelevantSince(path, sinceUnix) {
			continue
		}
		collectAccessLogEventsFromFile(path, sinceUnix, &all)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].TSUnix > all[j].TSUnix
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

func listRecentAccessLogSessionDetails(dataRoot string, sinceUnix int64, limit int) []map[string]any {
	if limit <= 0 {
		limit = accessLogDetailLimit
	}
	if limit > 100 {
		limit = 100
	}
	var all []accessLogEvent
	for _, path := range accessLogFiles(dataRoot) {
		if !accessLogFileRelevantSince(path, sinceUnix) {
			continue
		}
		collectAccessLogEventsFromFile(path, sinceUnix, &all)
	}
	if len(all) == 0 {
		return nil
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].TSUnix < all[j].TSUnix
	})

	type detail struct {
		session accessLogEvent
		paths   []string
	}
	bySession := make(map[string]*detail)
	order := make([]string, 0)

	appendPath := func(d *detail, label string) {
		if label == "" {
			return
		}
		n := len(d.paths)
		if n > 0 && d.paths[n-1] == label {
			return
		}
		d.paths = append(d.paths, label)
	}

	for _, ev := range all {
		sid := strings.TrimSpace(ev.SessionID)
		if accessLogEventCountsForSessionStats(ev) {
			if sid == "" {
				sid = formatSessionID(ev.IP, time.Unix(ev.TSUnix, 0))
			}
			if _, ok := bySession[sid]; !ok {
				order = append(order, sid)
				bySession[sid] = &detail{session: ev}
			} else {
				bySession[sid].session = ev
			}
			continue
		}
		if sid == "" || !accessLogEventCountsForStats(ev) {
			continue
		}
		if ev.Action != accessLogActionVisit && ev.Action != accessLogActionFeature {
			continue
		}
		d, ok := bySession[sid]
		if !ok {
			order = append(order, sid)
			d = &detail{}
			bySession[sid] = d
		}
		appendPath(d, formatAccessActionLabel(ev.Action, ev.Target))
	}

	out := make([]map[string]any, 0, limit)
	for i := len(order) - 1; i >= 0 && len(out) < limit; i-- {
		sid := order[i]
		d := bySession[sid]
		if d == nil || strings.TrimSpace(d.session.Action) == "" {
			continue
		}
		ev := d.session
		row := map[string]any{
			"ip":        ev.IP,
			"ts":        ev.TS,
			"tsUnix":    ev.TSUnix,
			"action":    ev.Action,
			"target":    ev.Target,
			"label":     formatAccessActionLabel(ev.Action, ev.Target),
			"sessionId": sid,
		}
		if g := strings.TrimSpace(ev.Geo); g != "" {
			row["geo"] = g
		}
		if c := strings.TrimSpace(ev.Client); c != "" {
			row["client"] = c
		}
		if len(d.paths) > 0 {
			row["paths"] = append([]string(nil), d.paths...)
		}
		out = append(out, row)
	}
	return out
}

func collectAccessLogEventsFromFile(path string, sinceUnix int64, out *[]accessLogEvent) {
	*out = append(*out, readAccessLogEventsFromFile(path, sinceUnix)...)
}

func readAccessLogEventsFromFile(path string, sinceUnix int64) []accessLogEvent {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	var out []accessLogEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 4096), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if isUserAccessSummaryLine(line) {
			continue
		}
		var ev accessLogEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		if !accessLogEventCountsForStats(ev) {
			continue
		}
		ts := ev.TSUnix
		if ts <= 0 || (sinceUnix > 0 && ts < sinceUnix) {
			continue
		}
		if strings.TrimSpace(ev.IP) == "" {
			continue
		}
		out = append(out, ev)
	}
	return out
}

const accessLogSummaryInsightWeightMax = 50

// expandSummaryEventsForInsights 月 summary 样本不足时按 header.uniqueVisitors（会话数）加权复现 session 事件。
func expandSummaryEventsForInsights(header userAccessLogSummary, events []accessLogEvent) []accessLogEvent {
	var sessions []accessLogEvent
	for _, ev := range events {
		if accessLogEventCountsForSessionStats(ev) {
			sessions = append(sessions, ev)
		}
	}
	if len(sessions) == 0 {
		return sessions
	}
	if header.UniqueVisitors <= len(sessions) {
		return sessions
	}
	weight := (header.UniqueVisitors + len(sessions) - 1) / len(sessions)
	if weight > accessLogSummaryInsightWeightMax {
		weight = accessLogSummaryInsightWeightMax
	}
	out := make([]accessLogEvent, 0, len(sessions)*weight)
	for _, ev := range sessions {
		for i := 0; i < weight; i++ {
			out = append(out, ev)
		}
	}
	return out
}

func accessLogEventsToJSON(events []accessLogEvent) []map[string]any {
	out := make([]map[string]any, 0, len(events))
	for _, ev := range events {
		row := map[string]any{
			"ip":     ev.IP,
			"ts":     ev.TS,
			"tsUnix": ev.TSUnix,
			"action": ev.Action,
			"target": ev.Target,
			"label":  formatAccessActionLabel(ev.Action, ev.Target),
		}
		if g := strings.TrimSpace(ev.Geo); g != "" {
			row["geo"] = g
		}
		if c := strings.TrimSpace(ev.Client); c != "" {
			row["client"] = c
		}
		if sid := strings.TrimSpace(ev.SessionID); sid != "" {
			row["sessionId"] = sid
		}
		out = append(out, row)
	}
	return out
}

func userAccessLogRole(name string) string {
	if name == accessLogFileName {
		return "访客访问（当前）"
	}
	if _, ok := parseUserAccessMonthlySummaryName(name); ok {
		return "访客访问月归档"
	}
	return "访客访问日归档"
}

func siteLogInventory(dataRoot string) []siteLogFileInfo {
	var out []siteLogFileInfo
	secretLog := filepath.Join(dataRoot, logRelDir)
	if entries, err := os.ReadDir(secretLog); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !isSiteAppLogEntryName(name) {
				continue
			}
			role := appLogArchiveRole(name)
			if role == "" {
				continue
			}
			out = append(out, siteLogFileInfo{
				Path:   logRelDir + "/" + name,
				Role:   role,
				Active: role == "运行日志（当前）",
			})
		}
	}
	for _, path := range accessLogFiles(dataRoot) {
		name := filepath.Base(path)
		out = append(out, siteLogFileInfo{
			Path:   secret07RelDir + "/" + name,
			Role:   userAccessLogRole(name),
			Active: name == accessLogFileName,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Active != out[j].Active {
			return out[i].Active
		}
		return out[i].Path < out[j].Path
	})
	return out
}

// siteLogBoardFiles 供 OPS 看板：当前文件全列，各归档类型仅预览最近若干条。
func siteLogBoardFiles(dataRoot string) []siteLogFileInfo {
	all := siteLogInventory(dataRoot)
	if len(all) == 0 {
		return all
	}
	var actives, archives []siteLogFileInfo
	for _, f := range all {
		if f.Active {
			actives = append(actives, f)
		} else {
			archives = append(archives, f)
		}
	}
	sort.Slice(actives, func(i, j int) bool {
		return actives[i].Path < actives[j].Path
	})
	byRole := make(map[string][]siteLogFileInfo)
	for _, f := range archives {
		byRole[f.Role] = append(byRole[f.Role], f)
	}
	out := append([]siteLogFileInfo{}, actives...)
	appendSiteLogArchivePreview := func(role string, items []siteLogFileInfo) {
		if len(items) == 0 {
			return
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].Path > items[j].Path
		})
		n := siteLogArchivePreviewPerRole
		if n > len(items) {
			n = len(items)
		}
		out = append(out, items[:n]...)
		if extra := len(items) - n; extra > 0 {
			out = append(out, siteLogFileInfo{Role: role, More: extra})
		}
	}
	for _, role := range siteLogArchiveRoleOrder {
		appendSiteLogArchivePreview(role, byRole[role])
		delete(byRole, role)
	}
	roles := make([]string, 0, len(byRole))
	for role := range byRole {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	for _, role := range roles {
		appendSiteLogArchivePreview(role, byRole[role])
	}
	return out
}

func buildAccessLogEvent(e *visitorEntry, now time.Time, action, target, sessionID string) accessLogEvent {
	ev := accessLogEvent{
		TSUnix: now.Unix(),
		TS:     formatBeijingUnix(now.Unix()),
		IP:     e.ip,
		Action: action,
		Target: target,
	}
	if sid := strings.TrimSpace(sessionID); sid != "" {
		ev.SessionID = sid
	}
	if g := strings.TrimSpace(e.geo); g != "" {
		ev.Geo = g
	}
	if c := strings.TrimSpace(e.client); c != "" {
		ev.Client = c
	}
	return ev
}

