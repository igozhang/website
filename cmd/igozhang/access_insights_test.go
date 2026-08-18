package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAccessLogCountryLabel(t *testing.T) {
	if accessLogCountryLabel("中国 CN 亚洲 AS") != "中国" {
		t.Fatal("zh country")
	}
	if accessLogCountryLabel("") != "未知" {
		t.Fatal("empty")
	}
	if accessLogCountryLabel("Hungary") != "Hungary" {
		t.Fatal("single token")
	}
	if accessLogCountryLabel("United States US North America NA") != "United States" {
		t.Fatal("multi-word country")
	}
}

func TestParseAccessLogGeoFields(t *testing.T) {
	name, code := parseAccessLogGeoFields("United States US North America NA")
	if name != "United States" || code != "US" {
		t.Fatalf("us: name=%q code=%q", name, code)
	}
	name, code = parseAccessLogGeoFields("中国 CN 亚洲 AS")
	if name != "中国" || code != "CN" {
		t.Fatalf("cn: name=%q code=%q", name, code)
	}
	long := strings.Repeat("国", 20)
	name, code = parseAccessLogGeoFields(long + " XX")
	if code != "XX" || !strings.HasSuffix(name, "…") || len([]rune(name)) != accessLogCountryNameMaxRunes+1 {
		t.Fatalf("truncate: name=%q len=%d", name, len([]rune(name)))
	}
}

func TestTopAccessGeoAndIP(t *testing.T) {
	geo := map[string]int{"CN": 10, "US": 5, "JP": 5}
	meta := map[string]accessGeoMeta{
		"CN": {Name: "中国", Code: "CN"},
		"US": {Name: "United States", Code: "US"},
		"JP": {Name: "日本", Code: "JP"},
	}
	top := topAccessGeoTop(geo, meta, 2)
	if len(top) != 2 || top[0].Label != "中国" || top[0].Code != "CN" || top[0].Count != 10 {
		t.Fatalf("geo top=%+v", top)
	}
	ip := map[string]int{"1.2.3.4": 3, "5.6.7.8": 7}
	ipGeo := map[string]string{"5.6.7.8": "中国 CN"}
	topIP := topAccessIPTop(ip, ipGeo, 8)
	if len(topIP) != 2 || topIP[0].IP != "5.6.7.8" || topIP[0].Count != 7 {
		t.Fatalf("ip top=%+v", topIP)
	}
}

func TestRecordAccessInsightsEvent(t *testing.T) {
	accessInsightsMu.Lock()
	accessInsightsDay = newAccessInsightsBucket()
	accessInsightsWeek = newAccessInsightsBucket()
	accessInsightsMonth = newAccessInsightsBucket()
	accessInsightsReady = true
	accessInsightsMu.Unlock()

	recordAccessInsightsEvent(accessLogEvent{
		IP:     "203.0.113.1",
		Geo:    "中国 CN",
		Action: accessLogActionSession,
		Target: "首页",
	})
	snap := accessBoardInsightsSnapshot()
	if len(snap.Day.GeoTop) != 1 || snap.Day.GeoTop[0].Label != "中国" || snap.Day.GeoTop[0].Code != "CN" || snap.Day.GeoTop[0].Count != 1 {
		t.Fatalf("day geo=%+v", snap.Day.GeoTop)
	}
	if len(snap.Day.IPTop) != 1 || snap.Day.IPTop[0].IP != "203.0.113.1" || snap.Day.IPTop[0].Count != 1 {
		t.Fatalf("day ip=%+v", snap.Day.IPTop)
	}
}

func TestAccessInsightsCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	accessInsightsMu.Lock()
	accessInsightsDay = newAccessInsightsBucket()
	accessInsightsWeek = newAccessInsightsBucket()
	accessInsightsMonth = newAccessInsightsBucket()
	accessInsightsReady = true
	accessInsightsBuiltAt = time.Now()
	accessInsightsDay.add(accessLogEvent{IP: "1.1.1.1", Geo: "中国 CN", Action: accessLogActionSession, Target: "首页"})
	accessInsightsWeek = accessInsightsDay
	accessInsightsMonth = accessInsightsDay
	accessInsightsMu.Unlock()

	if err := saveAccessInsightsCache(dir); err != nil {
		t.Fatal(err)
	}
	accessInsightsMu.Lock()
	accessInsightsDay = newAccessInsightsBucket()
	accessInsightsReady = false
	accessInsightsMu.Unlock()

	if !loadAccessInsightsCache(dir) {
		t.Fatal("load cache")
	}
	snap := accessBoardInsightsSnapshot()
	if !snap.Ready || len(snap.Day.GeoTop) != 1 || snap.Day.GeoTop[0].Count != 1 || snap.Day.GeoTop[0].Code != "CN" {
		t.Fatalf("snap=%+v", snap)
	}
}

func TestRebuildAccessInsightsFromLogs(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	initAccessLog(dir)
	now := time.Now()
	sid := "203.0.113.77:" + fmt.Sprintf("%d", now.Unix())
	appendAccessLog(dir, accessLogEvent{
		TSUnix:    now.Unix(),
		IP:        "203.0.113.77",
		Geo:       "中国 CN",
		Action:    accessLogActionSession,
		Target:    "igokey",
		SessionID: sid,
	})
	appendAccessLog(dir, accessLogEvent{
		TSUnix:    now.Unix(),
		IP:        "203.0.113.77",
		Geo:       "中国 CN",
		Action:    accessLogActionVisit,
		Target:    "首页",
		SessionID: sid,
	})
	appendAccessLog(dir, accessLogEvent{
		TSUnix:    now.Add(-10 * 24 * time.Hour).Unix(),
		IP:        "203.0.113.88",
		Geo:       "美国 US",
		Action:    accessLogActionSession,
		Target:    "首页",
		SessionID: "203.0.113.88:1",
	})

	rebuildAccessInsightsFromLogs(dir)
	snap := accessBoardInsightsSnapshot()
	if !snap.Ready {
		t.Fatal("not ready")
	}
	if len(snap.Day.IPTop) != 1 || snap.Day.IPTop[0].Count != 1 {
		t.Fatalf("day ip=%+v", snap.Day.IPTop)
	}
	if len(snap.Month.IPTop) != 2 {
		t.Fatalf("month ip=%+v", snap.Month.IPTop)
	}
	if _, err := os.Stat(accessInsightsCachePath(dir)); err != nil {
		t.Fatalf("cache file: %v", err)
	}
}

func TestMaintainAccessLogDir_reconcileInsights(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	secret := secret07Dir(dir)
	month := time.Now().AddDate(0, -2, 0)
	dayPath := filepath.Join(secret, accessLogArchiveStem+"."+month.Format("2006-01-02")+".log")
	line := []byte(`{"tsUnix":1,"ip":"9.9.9.9","geo":"中国 CN","action":"session","target":"首页","sessionId":"9.9.9.9:1"}` + "\n")
	if err := os.WriteFile(dayPath, line, 0o644); err != nil {
		t.Fatal(err)
	}
	maintainAccessLogDir(dir)
	snap := accessBoardInsightsSnapshot()
	if !snap.Ready {
		t.Fatal("insights not ready after maintain")
	}
}
