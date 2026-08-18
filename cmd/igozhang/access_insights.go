package main

import (
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
	accessInsightsTopN              = 8
	accessInsightsCacheName         = "user-access-board.cache.json"
	accessInsightsCacheVersion      = 5
	accessLogCountryNameMaxRunes    = 18
)

type accessInsightEntry struct {
	Label string `json:"label"`
	Code  string `json:"code,omitempty"`
	Count int    `json:"count"`
}

type accessIPInsightEntry struct {
	IP    string `json:"ip"`
	Geo   string `json:"geo,omitempty"`
	Count int    `json:"count"`
}

type accessPeriodInsights struct {
	GeoTop []accessInsightEntry   `json:"geoTop"`
	IPTop  []accessIPInsightEntry `json:"ipTop"`
}

type accessBoardInsights struct {
	Ready       bool                 `json:"ready"`
	Building    bool                 `json:"building,omitempty"`
	BuiltAtUnix int64                `json:"builtAtUnix,omitempty"`
	Day         accessPeriodInsights `json:"day"`
	Week        accessPeriodInsights `json:"week"`
	Month       accessPeriodInsights `json:"month"`
}

type accessInsightsCounterSnapshot struct {
	Geo     map[string]int            `json:"geo"`
	GeoMeta map[string]accessGeoMeta  `json:"geoMeta,omitempty"`
	IP      map[string]int            `json:"ip"`
	IPGeo   map[string]string         `json:"ipGeo,omitempty"`
}

type accessGeoMeta struct {
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

type accessInsightsCacheFile struct {
	Version     int                             `json:"version"`
	BuiltAtUnix int64                           `json:"builtAtUnix"`
	Day         accessInsightsCounterSnapshot   `json:"day"`
	Week        accessInsightsCounterSnapshot   `json:"week"`
	Month       accessInsightsCounterSnapshot   `json:"month"`
}

type accessInsightsBucket struct {
	geo     map[string]int
	geoMeta map[string]accessGeoMeta
	ip      map[string]int
	ipGeo   map[string]string
}

func newAccessInsightsBucket() accessInsightsBucket {
	return accessInsightsBucket{
		geo:     make(map[string]int),
		geoMeta: make(map[string]accessGeoMeta),
		ip:      make(map[string]int),
		ipGeo:   make(map[string]string),
	}
}

func (b *accessInsightsBucket) add(ev accessLogEvent) {
	if !accessLogEventCountsForSessionStats(ev) {
		return
	}
	ip := strings.TrimSpace(ev.IP)
	if ip == "" {
		return
	}
	name, code := parseAccessLogGeoFields(ev.Geo)
	key := code
	if key == "" {
		key = name
	}
	b.geo[key]++
	if cur, ok := b.geoMeta[key]; !ok || len([]rune(name)) >= len([]rune(cur.Name)) {
		b.geoMeta[key] = accessGeoMeta{Name: countryDisplayNameZH(code, name), Code: code}
	}
	b.ip[ip]++
	if g := strings.TrimSpace(ev.Geo); g != "" {
		b.ipGeo[ip] = g
	}
}

func (b accessInsightsBucket) snapshot() accessInsightsCounterSnapshot {
	return accessInsightsCounterSnapshot{
		Geo:     copyStringIntMap(b.geo),
		GeoMeta: copyGeoMetaMap(b.geoMeta),
		IP:      copyStringIntMap(b.ip),
		IPGeo:   copyStringStringMap(b.ipGeo),
	}
}

func (b *accessInsightsBucket) load(s accessInsightsCounterSnapshot) {
	b.geo = copyStringIntMap(s.Geo)
	b.geoMeta = copyGeoMetaMap(s.GeoMeta)
	b.ip = copyStringIntMap(s.IP)
	b.ipGeo = copyStringStringMap(s.IPGeo)
	if b.geo == nil {
		b.geo = make(map[string]int)
	}
	if b.geoMeta == nil {
		b.geoMeta = make(map[string]accessGeoMeta)
	}
	if b.ip == nil {
		b.ip = make(map[string]int)
	}
	if b.ipGeo == nil {
		b.ipGeo = make(map[string]string)
	}
	for key := range b.geo {
		if _, ok := b.geoMeta[key]; ok {
			continue
		}
		if isISO3166Alpha2(key) {
			b.geoMeta[key] = accessGeoMeta{Name: key, Code: key}
		} else {
			b.geoMeta[key] = accessGeoMeta{Name: key}
		}
	}
}

var (
	accessInsightsMu        sync.RWMutex
	accessInsightsRebuildMu sync.Mutex
	accessInsightsReady     bool
	accessInsightsBuilding  bool
	accessInsightsBuiltAt   time.Time
	accessInsightsDay       accessInsightsBucket
	accessInsightsWeek      accessInsightsBucket
	accessInsightsMonth     accessInsightsBucket
)

func applyImportAccessInsightsCache(dataRoot string, cf *accessInsightsCacheFile) error {
	if cf == nil || cf.Version != accessInsightsCacheVersion {
		return fmt.Errorf("invalid insights cache version")
	}
	accessInsightsMu.Lock()
	accessInsightsDay.load(cf.Day)
	accessInsightsWeek.load(cf.Week)
	accessInsightsMonth.load(cf.Month)
	accessInsightsReady = true
	accessInsightsBuilding = false
	if cf.BuiltAtUnix > 0 {
		accessInsightsBuiltAt = time.Unix(cf.BuiltAtUnix, 0)
	} else {
		accessInsightsBuiltAt = time.Now()
	}
	accessInsightsMu.Unlock()
	return saveAccessInsightsCache(dataRoot)
}

func initAccessInsights(dataRoot string) {
	accessInsightsDay = newAccessInsightsBucket()
	accessInsightsWeek = newAccessInsightsBucket()
	accessInsightsMonth = newAccessInsightsBucket()
	if loadAccessInsightsCache(dataRoot) {
		return
	}
	// 同步重建：避免与测试 TempDir 清理、滚动退出时的异步写盘竞态。
	rebuildAccessInsightsFromLogs(dataRoot)
}

func isISO3166Alpha2(s string) bool {
	if len(s) != 2 {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func truncateCountryNameRunes(name string, max int) string {
	if max <= 0 {
		return name
	}
	r := []rune(strings.TrimSpace(name))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max]) + "…"
}

// parseAccessLogGeoFields 从访客 geo 串解析国名与二字代码（格式：国名 代码 大洲 …）。
func parseAccessLogGeoFields(geo string) (name, code string) {
	geo = strings.TrimSpace(geo)
	if geo == "" {
		return "未知", ""
	}
	parts := strings.Fields(geo)
	if len(parts) == 1 {
		return truncateCountryNameRunes(parts[0], accessLogCountryNameMaxRunes), ""
	}
	codeIdx := -1
	for i := 1; i < len(parts); i++ {
		if isISO3166Alpha2(parts[i]) {
			codeIdx = i
			break
		}
	}
	if codeIdx > 0 {
		name = strings.Join(parts[:codeIdx], " ")
		code = parts[codeIdx]
		return truncateCountryNameRunes(name, accessLogCountryNameMaxRunes), code
	}
	return truncateCountryNameRunes(parts[0], accessLogCountryNameMaxRunes), ""
}

// accessLogCountryLabel 保留供简单场景；统计请用 parseAccessLogGeoFields。
func accessLogCountryLabel(geo string) string {
	name, _ := parseAccessLogGeoFields(geo)
	return name
}

func recordAccessInsightsEvent(ev accessLogEvent) {
	if !accessLogEventCountsForSessionStats(ev) {
		return
	}
	accessInsightsMu.Lock()
	defer accessInsightsMu.Unlock()
	accessInsightsDay.add(ev)
	accessInsightsWeek.add(ev)
	accessInsightsMonth.add(ev)
	if accessInsightsReady {
		accessInsightsBuiltAt = time.Now()
	}
}

func rebuildAccessInsightsFromLogs(dataRoot string) {
	accessInsightsRebuildMu.Lock()
	defer accessInsightsRebuildMu.Unlock()

	accessInsightsMu.Lock()
	accessInsightsBuilding = true
	accessInsightsMu.Unlock()

	now := time.Now()
	cutDay, cutWeek, cutMonth := accessLogStatsCutoffs(now)
	day := newAccessInsightsBucket()
	week := newAccessInsightsBucket()
	month := newAccessInsightsBucket()

	var all []accessLogEvent
	for _, path := range accessLogFiles(dataRoot) {
		if !accessLogFileRelevantForMonth(path, cutMonth) {
			continue
		}
		events := readAccessLogEventsFromFile(path, cutMonth)
		if header, ok := readUserAccessSummaryHeader(path); ok {
			events = expandSummaryEventsForInsights(header, events)
		}
		all = append(all, events...)
	}
	for _, ev := range all {
		ts := ev.TSUnix
		if ts >= cutDay {
			day.add(ev)
		}
		if ts >= cutWeek {
			week.add(ev)
		}
		if ts >= cutMonth {
			month.add(ev)
		}
	}

	accessInsightsMu.Lock()
	accessInsightsDay = day
	accessInsightsWeek = week
	accessInsightsMonth = month
	accessInsightsReady = true
	accessInsightsBuilding = false
	accessInsightsBuiltAt = now
	accessInsightsMu.Unlock()

	if err := saveAccessInsightsCache(dataRoot); err != nil {
		log.Printf("user-access insights: save cache: %v", err)
	} else {
		log.Printf("user-access insights: rebuilt from logs")
	}
}

func reconcileAccessInsights(dataRoot string) {
	rebuildAccessInsightsFromLogs(dataRoot)
}

func accessInsightsCachePath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), accessInsightsCacheName)
}

func loadAccessInsightsCache(dataRoot string) bool {
	raw, err := os.ReadFile(accessInsightsCachePath(dataRoot))
	if err != nil || len(raw) == 0 {
		return false
	}
	var cf accessInsightsCacheFile
	if err := json.Unmarshal(raw, &cf); err != nil || cf.Version != accessInsightsCacheVersion {
		return false
	}
	accessInsightsMu.Lock()
	defer accessInsightsMu.Unlock()
	accessInsightsDay.load(cf.Day)
	accessInsightsWeek.load(cf.Week)
	accessInsightsMonth.load(cf.Month)
	accessInsightsReady = true
	accessInsightsBuilding = false
	if cf.BuiltAtUnix > 0 {
		accessInsightsBuiltAt = time.Unix(cf.BuiltAtUnix, 0)
	} else {
		accessInsightsBuiltAt = time.Now()
	}
	log.Printf("user-access insights: loaded cache")
	return true
}

func saveAccessInsightsCache(dataRoot string) error {
	accessInsightsMu.RLock()
	cf := accessInsightsCacheFile{
		Version:     accessInsightsCacheVersion,
		BuiltAtUnix: accessInsightsBuiltAt.Unix(),
		Day:         accessInsightsDay.snapshot(),
		Week:        accessInsightsWeek.snapshot(),
		Month:       accessInsightsMonth.snapshot(),
	}
	accessInsightsMu.RUnlock()

	b, err := json.Marshal(cf)
	if err != nil {
		return err
	}
	dir := secret07Dir(dataRoot)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return fmt.Errorf("insights cache dir: %w", err)
	}
	p := accessInsightsCachePath(dataRoot)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	// Windows 上 Rename 不能覆盖已存在文件；失败时回退直写，避免 Remove 后留下空洞。
	_ = os.Remove(p)
	if err := os.Rename(tmp, p); err != nil {
		if werr := os.WriteFile(p, b, 0o644); werr != nil {
			_ = os.Remove(tmp)
			return werr
		}
		_ = os.Remove(tmp)
	}
	return nil
}

func accessBoardInsightsSnapshot() accessBoardInsights {
	accessInsightsMu.RLock()
	defer accessInsightsMu.RUnlock()
	out := accessBoardInsights{
		Ready:       accessInsightsReady,
		Building:    accessInsightsBuilding,
		BuiltAtUnix: accessInsightsBuiltAt.Unix(),
		Day:         accessPeriodInsightsFromBucket(accessInsightsDay),
		Week:        accessPeriodInsightsFromBucket(accessInsightsWeek),
		Month:       accessPeriodInsightsFromBucket(accessInsightsMonth),
	}
	if !accessInsightsReady && accessInsightsBuilding {
		out.Building = true
	}
	return out
}

func accessPeriodInsightsFromBucket(b accessInsightsBucket) accessPeriodInsights {
	return accessPeriodInsights{
		GeoTop: topAccessGeoTop(b.geo, b.geoMeta, accessInsightsTopN),
		IPTop:  topAccessIPTop(b.ip, b.ipGeo, accessInsightsTopN),
	}
}

func topAccessGeoTop(counts map[string]int, meta map[string]accessGeoMeta, n int) []accessInsightEntry {
	type pair struct {
		label string
		code  string
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for key, count := range counts {
		if count <= 0 {
			continue
		}
		m := meta[key]
		label := countryDisplayNameZH(m.Code, m.Name)
		if label == "" {
			label = key
		}
		pairs = append(pairs, pair{label: label, code: m.Code, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		if pairs[i].code != pairs[j].code {
			return pairs[i].code < pairs[j].code
		}
		return pairs[i].label < pairs[j].label
	})
	if n <= 0 || len(pairs) < n {
		n = len(pairs)
	}
	out := make([]accessInsightEntry, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, accessInsightEntry{
			Label: pairs[i].label,
			Code:  pairs[i].code,
			Count: pairs[i].count,
		})
	}
	return out
}

func topAccessIPTop(counts map[string]int, ipGeo map[string]string, n int) []accessIPInsightEntry {
	type pair struct {
		ip    string
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for ip, count := range counts {
		if count > 0 {
			pairs = append(pairs, pair{ip: ip, count: count})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].ip < pairs[j].ip
	})
	if n <= 0 || len(pairs) < n {
		n = len(pairs)
	}
	out := make([]accessIPInsightEntry, 0, n)
	for i := 0; i < n; i++ {
		entry := accessIPInsightEntry{IP: pairs[i].ip, Count: pairs[i].count}
		if g, ok := ipGeo[pairs[i].ip]; ok {
			entry.Geo = formatAccessGeoBoardZH(g)
		}
		out = append(out, entry)
	}
	return out
}

func copyGeoMetaMap(m map[string]accessGeoMeta) map[string]accessGeoMeta {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]accessGeoMeta, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyStringIntMap(m map[string]int) map[string]int {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyStringStringMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
