package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	userAccessBundleZipEntry            = "secret07/user-access-bundle.json"
	userAccessBundleVersion             = 3
	userAccessBundleVersionLegacy       = 2
	zipImportAccessBundleMax            = 8 << 20
	userAccessBundleDetailPathsPerSess  = visitorTrailMax
)

type userAccessBundleStatsManifest struct {
	Day            int      `json:"day"`
	Week           int      `json:"week"`
	Month          int      `json:"month"`
	MonthsIncluded []string `json:"monthsIncluded,omitempty"`
}

type userAccessBundleMonthlyStat struct {
	Month          string           `json:"month"`
	UniqueVisitors int              `json:"uniqueVisitors"`
	EventCount     int              `json:"eventCount"`
	Samples        []accessLogEvent `json:"samples,omitempty"`
}

type userAccessBundle struct {
	Version            int                              `json:"version"`
	ExportedAt         string                           `json:"exportedAt"`
	DetailWindowDays   int                              `json:"detailWindowDays"`
	StatsRetentionDays int                              `json:"statsRetentionDays"`
	DetailSlimmed      bool                             `json:"detailSlimmed,omitempty"`
	ActiveLogDay       string                           `json:"activeLogDay,omitempty"`
	ActiveLog          []accessLogEvent                 `json:"activeLog,omitempty"`
	DailyArchives      map[string][]accessLogEvent      `json:"dailyArchives,omitempty"`
	MonthlyStats       []userAccessBundleMonthlyStat    `json:"monthlyStats,omitempty"`
	StatsManifest      *userAccessBundleStatsManifest   `json:"statsManifest,omitempty"`
	InsightsCache      *accessInsightsCacheFile         `json:"insightsCache,omitempty"`
}

// prepareUserAccessLogsForBundleExport 导出前压实月归档，避免日归档超窗后未进 summary 而丢失。
func prepareUserAccessLogsForBundleExport(dataRoot string) {
	dir := secret07Dir(dataRoot)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return
	}
	now := time.Now()
	maintainUserAccessActiveLog(dir, now)
	compactUserAccessMonthlyArchives(dir, now)
}

func userAccessMonthsWithSummary(dir string) map[string]struct{} {
	out := make(map[string]struct{})
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if monthKey, ok := parseUserAccessMonthlySummaryName(e.Name()); ok {
			out[monthKey] = struct{}{}
		}
	}
	return out
}

func shouldExportDailyArchiveForBundle(day time.Time, summaries map[string]struct{}) bool {
	if summaries == nil {
		return true
	}
	_, hasSummary := summaries[userAccessMonthKey(day)]
	return !hasSummary
}

func slimAccessLogEventsForBundle(events []accessLogEvent) []accessLogEvent {
	if len(events) == 0 {
		return nil
	}
	sorted := append([]accessLogEvent(nil), events...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].TSUnix != sorted[j].TSUnix {
			return sorted[i].TSUnix < sorted[j].TSUnix
		}
		return sorted[i].SessionID < sorted[j].SessionID
	})

	var sessions []accessLogEvent
	pathBySession := make(map[string][]accessLogEvent)
	for _, ev := range sorted {
		if accessLogEventCountsForSessionStats(ev) {
			sessions = append(sessions, ev)
			continue
		}
		if !accessLogEventCountsForStats(ev) {
			continue
		}
		sid := strings.TrimSpace(ev.SessionID)
		if sid == "" {
			continue
		}
		pathBySession[sid] = append(pathBySession[sid], ev)
	}

	capPaths := func(paths []accessLogEvent) []accessLogEvent {
		if len(paths) <= userAccessBundleDetailPathsPerSess {
			return paths
		}
		return paths[len(paths)-userAccessBundleDetailPathsPerSess:]
	}

	out := make([]accessLogEvent, 0, len(events))
	sessionSeen := make(map[string]bool)
	for _, ev := range sessions {
		out = append(out, ev)
		sid := strings.TrimSpace(ev.SessionID)
		sessionSeen[sid] = true
		out = append(out, capPaths(pathBySession[sid])...)
	}
	for sid, paths := range pathBySession {
		if sessionSeen[sid] {
			continue
		}
		out = append(out, capPaths(paths)...)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TSUnix != out[j].TSUnix {
			return out[i].TSUnix < out[j].TSUnix
		}
		return out[i].SessionID < out[j].SessionID
	})
	return out
}

func buildUserAccessBundleStatsManifest(dataRoot string, now time.Time, monthly []userAccessBundleMonthlyStat) *userAccessBundleStatsManifest {
	st := countAccessLogStats(dataRoot, now)
	months := make([]string, 0, len(monthly))
	for _, m := range monthly {
		if m.Month != "" {
			months = append(months, m.Month)
		}
	}
	sort.Strings(months)
	return &userAccessBundleStatsManifest{
		Day:            st.Day,
		Week:           st.Week,
		Month:          st.Month,
		MonthsIncluded: months,
	}
}

func buildUserAccessBundleBytes(dataRoot string) ([]byte, error) {
	return buildUserAccessBundleBytesOpts(dataRoot, true)
}

// buildUserAccessBundleBytesOpts 当 prepare=false 时跳过导出前日志压实（供 HTTP 导出热路径使用）。
func buildUserAccessBundleBytesOpts(dataRoot string, prepare bool) ([]byte, error) {
	if dataRoot == "" {
		return nil, nil
	}
	if prepare {
		prepareUserAccessLogsForBundleExport(dataRoot)
	}

	now := time.Now()
	statsCutoff := now.Add(-userAccessStatsRetention)
	dir := secret07Dir(dataRoot)
	summaries := userAccessMonthsWithSummary(dir)

	bundle := userAccessBundle{
		Version:            userAccessBundleVersion,
		ExportedAt:         now.Format(time.RFC3339),
		DetailWindowDays:   userAccessRecentFullDays,
		StatsRetentionDays: int(userAccessStatsRetention / (24 * time.Hour)),
		DetailSlimmed:      true,
		DailyArchives:      make(map[string][]accessLogEvent),
	}

	activePath := accessLogPath(dataRoot)
	if events := collectAccessLogEventsFromPath(activePath, 0); len(events) > 0 {
		bundle.ActiveLog = slimAccessLogEventsForBundle(events)
	}
	bundle.ActiveLogDay = computeBundleActiveLogDay(bundle.ActiveLog, now)

	for _, path := range accessLogFiles(dataRoot) {
		name := filepath.Base(path)
		if name == accessLogFileName {
			continue
		}
		if day, ok := parseUserAccessDailyArchiveDay(name); ok {
			if !shouldExportDailyArchiveForBundle(day, summaries) {
				continue
			}
			events := collectAccessLogEventsFromPath(path, 0)
			if len(events) == 0 {
				continue
			}
			bundle.DailyArchives[day.Format("2006-01-02")] = slimAccessLogEventsForBundle(events)
			continue
		}
		if monthKey, ok := parseUserAccessMonthlySummaryName(name); ok {
			end, ok := userAccessMonthEnd(monthKey)
			if !ok || end.Before(statsCutoff) {
				continue
			}
			stat, err := parseUserAccessMonthlySummaryFile(path)
			if err != nil {
				continue
			}
			stat.Samples = slimAccessLogEventsForBundle(stat.Samples)
			bundle.MonthlyStats = append(bundle.MonthlyStats, stat)
		}
	}

	if cache := exportAccessInsightsCacheSnapshot(dataRoot); cache != nil {
		bundle.InsightsCache = cache
	}
	bundle.StatsManifest = buildUserAccessBundleStatsManifest(dataRoot, now, bundle.MonthlyStats)

	if !userAccessBundleHasData(bundle) {
		return nil, nil
	}
	return json.Marshal(bundle)
}

func peekBundleStatsManifest(raw []byte) *userAccessBundleStatsManifest {
	if len(raw) == 0 {
		return nil
	}
	var peek struct {
		StatsManifest *userAccessBundleStatsManifest `json:"statsManifest"`
	}
	if err := json.Unmarshal(raw, &peek); err != nil {
		return nil
	}
	return peek.StatsManifest
}

func userAccessBundleHasData(b userAccessBundle) bool {
	return len(b.ActiveLog) > 0 || len(b.DailyArchives) > 0 || len(b.MonthlyStats) > 0
}

func computeBundleActiveLogDay(activeLog []accessLogEvent, exportedAt time.Time) string {
	var maxTS int64
	for _, ev := range activeLog {
		if ev.TSUnix > maxTS {
			maxTS = ev.TSUnix
		}
	}
	if maxTS > 0 {
		return time.Unix(maxTS, 0).In(time.Local).Format("2006-01-02")
	}
	return exportedAt.In(time.Local).Format("2006-01-02")
}

func collectAccessLogEventsFromPath(path string, sinceUnix int64) []accessLogEvent {
	var out []accessLogEvent
	collectAccessLogEventsFromFile(path, sinceUnix, &out)
	return out
}

func parseUserAccessMonthlySummaryFile(path string) (userAccessBundleMonthlyStat, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return userAccessBundleMonthlyStat{}, err
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) == 0 {
		return userAccessBundleMonthlyStat{}, fmt.Errorf("empty summary")
	}
	var sum userAccessLogSummary
	if err := json.Unmarshal([]byte(lines[0]), &sum); err != nil {
		return userAccessBundleMonthlyStat{}, err
	}
	if sum.Kind != userAccessSummaryKind || sum.Month == "" {
		return userAccessBundleMonthlyStat{}, fmt.Errorf("bad summary header")
	}
	stat := userAccessBundleMonthlyStat{
		Month:          sum.Month,
		UniqueVisitors: sum.UniqueVisitors,
		EventCount:     sum.EventCount,
	}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev accessLogEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		if !accessLogEventCountsForStats(ev) {
			continue
		}
		stat.Samples = append(stat.Samples, ev)
	}
	return stat, nil
}

func exportAccessInsightsCacheSnapshot(dataRoot string) *accessInsightsCacheFile {
	accessInsightsMu.RLock()
	hasData := len(accessInsightsDay.geo) > 0 || len(accessInsightsWeek.geo) > 0 || len(accessInsightsMonth.geo) > 0
	if accessInsightsReady || hasData {
		cf := &accessInsightsCacheFile{
			Version:     accessInsightsCacheVersion,
			BuiltAtUnix: accessInsightsBuiltAt.Unix(),
			Day:         accessInsightsDay.snapshot(),
			Week:        accessInsightsWeek.snapshot(),
			Month:       accessInsightsMonth.snapshot(),
		}
		if cf.BuiltAtUnix <= 0 {
			cf.BuiltAtUnix = time.Now().Unix()
		}
		accessInsightsMu.RUnlock()
		return cf
	}
	accessInsightsMu.RUnlock()

	raw, err := os.ReadFile(accessInsightsCachePath(dataRoot))
	if err != nil || len(raw) == 0 {
		return nil
	}
	var cf accessInsightsCacheFile
	if err := json.Unmarshal(raw, &cf); err != nil || cf.Version != accessInsightsCacheVersion {
		return nil
	}
	return &cf
}

func importUserAccessBundle(dataRoot string, raw []byte) ([]string, error) {
	if dataRoot == "" {
		return nil, fmt.Errorf("empty data root")
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty bundle")
	}
	var bundle userAccessBundle
	if err := json.Unmarshal(raw, &bundle); err != nil {
		return nil, err
	}
	if bundle.Version != userAccessBundleVersion && bundle.Version != userAccessBundleVersionLegacy {
		return nil, fmt.Errorf("unsupported bundle version %d", bundle.Version)
	}
	if !userAccessBundleHasDetailData(bundle) {
		return nil, fmt.Errorf("bundle has no access log detail data")
	}
	for dayKey := range bundle.DailyArchives {
		if _, err := time.ParseInLocation("2006-01-02", dayKey, time.Local); err != nil {
			return nil, fmt.Errorf("invalid daily archive key %q", dayKey)
		}
	}
	for _, stat := range bundle.MonthlyStats {
		if stat.Month == "" {
			continue
		}
		if _, err := time.ParseInLocation("2006-01", stat.Month, time.Local); err != nil {
			return nil, fmt.Errorf("invalid monthly stat month %q", stat.Month)
		}
	}

	if err := clearAccessLogsInImportWindow(dataRoot); err != nil {
		return nil, fmt.Errorf("clear access logs: %w", err)
	}
	if err := os.MkdirAll(secret07Dir(dataRoot), 0o755); err != nil {
		return nil, err
	}

	if bundle.ActiveLog != nil {
		if err := writeAccessLogNDJSON(accessLogPath(dataRoot), bundle.ActiveLog); err != nil {
			return nil, fmt.Errorf("active log: %w", err)
		}
	}
	dir := secret07Dir(dataRoot)
	for dayKey, events := range bundle.DailyArchives {
		path := filepath.Join(dir, fmt.Sprintf("%s.%s.log", accessLogArchiveStem, dayKey))
		if err := writeAccessLogNDJSON(path, events); err != nil {
			return nil, fmt.Errorf("daily %s: %w", dayKey, err)
		}
	}
	for _, stat := range bundle.MonthlyStats {
		if stat.Month == "" {
			continue
		}
		if err := writeUserAccessMonthlySummaryFromStat(dir, stat); err != nil {
			return nil, fmt.Errorf("monthly %s: %w", stat.Month, err)
		}
	}

	activeLogDay := strings.TrimSpace(bundle.ActiveLogDay)
	if activeLogDay == "" {
		activeLogDay = computeBundleActiveLogDay(bundle.ActiveLog, time.Now())
	}
	if _, err := time.ParseInLocation("2006-01-02", activeLogDay, time.Local); err != nil {
		activeLogDay = logTodayLocal()
	}
	writeRotatingLogLastDay(dir, accessLogLastDayMarker, activeLogDay)

	resetVisitorBoardForImport()
	resetAccessLogDedupForImport()

	var warnings []string
	if bundle.InsightsCache != nil {
		if err := applyImportAccessInsightsCache(dataRoot, bundle.InsightsCache); err != nil {
			log.Printf("user-access bundle: restore insights cache: %v", err)
			resetAccessInsightsForImport()
			rebuildAccessInsightsFromLogs(dataRoot)
			warnings = append(warnings, "insights cache restore failed, rebuilt from logs")
		}
	} else {
		resetAccessInsightsForImport()
		rebuildAccessInsightsFromLogs(dataRoot)
	}

	exportedAt := time.Now()
	if ts, err := time.Parse(time.RFC3339, strings.TrimSpace(bundle.ExportedAt)); err == nil {
		exportedAt = ts
	}
	if msg := validateUserAccessBundleStatsManifest(dataRoot, bundle.StatsManifest, exportedAt); msg != "" {
		log.Printf("user-access bundle: %s", msg)
		warnings = append(warnings, msg)
	}
	return warnings, nil
}

func userAccessBundleHasDetailData(b userAccessBundle) bool {
	return userAccessBundleHasData(b)
}

func validateUserAccessBundleStatsManifest(dataRoot string, manifest *userAccessBundleStatsManifest, exportedAt time.Time) string {
	if manifest == nil {
		return ""
	}
	got := countAccessLogStats(dataRoot, exportedAt)
	if got.Day == manifest.Day && got.Week == manifest.Week && got.Month == manifest.Month {
		return ""
	}
	return fmt.Sprintf(
		"stats manifest mismatch: expected day=%d week=%d month=%d, got day=%d week=%d month=%d",
		manifest.Day, manifest.Week, manifest.Month, got.Day, got.Week, got.Month,
	)
}

func resetAccessInsightsForImport() {
	accessInsightsMu.Lock()
	defer accessInsightsMu.Unlock()
	accessInsightsReady = false
	accessInsightsBuilding = true
	accessInsightsDay = newAccessInsightsBucket()
	accessInsightsWeek = newAccessInsightsBucket()
	accessInsightsMonth = newAccessInsightsBucket()
}

func writeAccessLogNDJSON(path string, events []accessLogEvent) error {
	var body strings.Builder
	for _, ev := range events {
		b, err := json.Marshal(ev)
		if err != nil {
			continue
		}
		body.Write(b)
		body.WriteByte('\n')
	}
	return zipWriteFile(path, []byte(body.String()), 0o644)
}

func writeUserAccessMonthlySummaryFromStat(dir string, stat userAccessBundleMonthlyStat) error {
	summary := userAccessLogSummary{
		Kind:           userAccessSummaryKind,
		Month:          stat.Month,
		UniqueVisitors: stat.UniqueVisitors,
		EventCount:     stat.EventCount,
	}
	head, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	var body strings.Builder
	body.Write(head)
	body.WriteByte('\n')
	for _, ev := range stat.Samples {
		b, err := json.Marshal(ev)
		if err != nil {
			continue
		}
		body.Write(b)
		body.WriteByte('\n')
	}
	path := userAccessSummaryPath(dir, stat.Month)
	if err := os.WriteFile(path, []byte(body.String()), 0o644); err != nil {
		return err
	}
	enforceUserAccessFileSizeCap(path)
	return nil
}
