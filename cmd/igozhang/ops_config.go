package main

import (
	_ "embed"
	"crypto/sha256"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const opsConfigName = "igo.conf"

//go:embed igo.conf
var opsConfigExample []byte

type opsSnapshot struct {
	maintenanceInterval   time.Duration
	maxStashBytes         int
	stashTTL              time.Duration
	stashPollInterval     time.Duration
	stashPollIntervalMax  time.Duration
	fsFileMaxAge          time.Duration
	transferDirMaxBytes   int64
	usageBaselineBytes    int64
	transferMaxConcurrent int
	uploadMaxBytes        int64
	fetchUpstreamTimeout  time.Duration
	browseRLPerMinute     int
	stashRLPerMinute      int
	logFileMaxAge         time.Duration
	logRotateSizeBytes    int64
	logTruncateSizeBytes  int64
	previewMaxBytes       int64
	mp3MinBufferBytes     int
	mp3PrefetchAheadSec   int
	mp3StallMs            int
	mp3MaxStallRetries    int
	mp3CoverMaxBytes      int
	mp3GateEnabled        bool
	mp3GateTTL            time.Duration
	mp3RefererCheck       bool
	mp3MaxConcurrentPerIP int
	mp3MaxTracksPerMinute int
	siteTitle             string
	siteTagline           string
	siteBannerMaskPercent int
	beijingOffsetHours    int
	igokeyEnabled         bool
	authEnabled           bool
	opsEnabled            bool
	boardPollInterval        time.Duration
	visitorSessionIdleTimeout time.Duration
	footerGomoku          string
	footerXiangqi         string
	footerTexas           string
	footerMaxLinks        int
	xiangqiEasyDepth      int
	xiangqiEasyTimeMs     int
	xiangqiHardDepth      int
	xiangqiHardTimeMs     int
	texasStartingChips    int
	texasSmallBlind       int
	texasBigBlind         int
	texasMaxRaises        int
	texasEasyThinkMs      int
	texasHardThinkMs      int
	footerLinks           []footerLink
	diceRollMs            int
	diceResultShowMs      int
	zhijianFuseMs         int
	zhijianFuseComboAfter int
	zhijianFuseComboBonus int
	zhijianFuseUrgentMs   int
	zhijianSessionSec     int
	zhijianMinRecordMs    int
	zhijianComboAccum     int
	zhijianComboDanmaku   int
	zhijianComboDanmakuEa int
	zhijianComboDecay     float64
	zhijianTitleCPM       []int
	zhijianWord3kPoolSize int
	docHomeLatest              int
	docNavBallEnabled          bool
	pathProbeEnabled           bool
	pathProbeWindow            time.Duration
	pathProbeMaxPaths          int
	pathProbeBanDuration       time.Duration
	pathProbeExcludePrefixes   []string
}

var (
	opsMu        sync.RWMutex
	opsCurrent   opsSnapshot
	opsConfigRev uint64
	opsDataRoot  string
)

func defaultOpsSnapshot() opsSnapshot {
	return opsSnapshot{
		maintenanceInterval:   time.Hour,
		maxStashBytes:         200 * 1024,
		stashTTL:              12 * time.Hour,
		stashPollInterval:     3 * time.Second,
		stashPollIntervalMax:  30 * time.Second,
		fsFileMaxAge:          7 * 24 * time.Hour,
		transferDirMaxBytes:   9 * 1024 * 1024 * 1024,
		usageBaselineBytes:    9 * 1024 * 1024 * 1024,
		transferMaxConcurrent: 2,
		uploadMaxBytes:        2 << 30,
		fetchUpstreamTimeout:  30 * time.Minute,
		browseRLPerMinute:     150,
		stashRLPerMinute:      80,
		logFileMaxAge:         15 * 24 * time.Hour,
		logRotateSizeBytes:    500 * 1024,
		logTruncateSizeBytes:  1024 * 1024,
		previewMaxBytes:       4 << 20,
		mp3MinBufferBytes:     512 * 1024,
		mp3PrefetchAheadSec:   10,
		mp3StallMs:            4500,
		mp3MaxStallRetries:    4,
		mp3CoverMaxBytes:      8 << 20,
		mp3GateEnabled:        true,
		mp3GateTTL:            12 * time.Hour,
		mp3RefererCheck:       true,
		mp3MaxConcurrentPerIP: 4,
		mp3MaxTracksPerMinute: 40,
		siteTitle:             defaultSiteDisplayName,
		siteTagline:           defaultSiteTagline,
		siteBannerMaskPercent: siteBannerMaskDefault,
		beijingOffsetHours:    0,
		igokeyEnabled:         true,
		authEnabled:           true,
		opsEnabled:            true,
		boardPollInterval:         5 * time.Second,
		visitorSessionIdleTimeout: 30 * time.Minute,
		footerGomoku:          pathGomoku,
		footerXiangqi:         pathXiangqi,
		footerTexas:           pathTexas,
		footerMaxLinks:        6,
		xiangqiEasyDepth:      2,
		xiangqiEasyTimeMs:     700,
		xiangqiHardDepth:      3,
		xiangqiHardTimeMs:     1600,
		texasStartingChips:    2000,
		texasSmallBlind:       10,
		texasBigBlind:         20,
		texasMaxRaises:        2,
		texasEasyThinkMs:      500,
		texasHardThinkMs:      2000,
		footerLinks: []footerLink{
			{Text: "igozhang.cn", Href: "https://igozhang.cn"},
		},
		diceRollMs:            1150,
		diceResultShowMs:      2200,
		zhijianFuseMs:         4000,
		zhijianFuseComboAfter: 5,
		zhijianFuseComboBonus: 500,
		zhijianFuseUrgentMs:   800,
		zhijianSessionSec:     150,
		zhijianMinRecordMs:    30000,
		zhijianComboAccum:     3,
		zhijianComboDanmaku:   5,
		zhijianComboDanmakuEa: 8,
		zhijianComboDecay:     0.038,
		zhijianTitleCPM:       []int{30, 50, 65, 75, 95, 150},
		zhijianWord3kPoolSize: 80,
		docHomeLatest:            5,
		docNavBallEnabled:        true,
		pathProbeEnabled:         true,
		pathProbeWindow:          time.Minute,
		pathProbeMaxPaths:        20,
		pathProbeBanDuration:     time.Hour,
		pathProbeExcludePrefixes: []string{"/stream/mp3/", "/assets/doc-pic/", "/assets/doc/"},
	}
}

func opsConfigPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), opsConfigName)
}

func applyOpsSnapshot(s opsSnapshot) {
	opsMu.Lock()
	opsCurrent = s
	maxStashBytes = s.maxStashBytes
	stashTTL = s.stashTTL
	fsFileMaxAge = s.fsFileMaxAge
	transferDirMaxBytes = s.transferDirMaxBytes
	logFileMaxAge = s.logFileMaxAge
	logRotateSizeBytes = s.logRotateSizeBytes
	logTruncateSizeBytes = s.logTruncateSizeBytes
	maxPreviewBytes = s.previewMaxBytes
	maxCoverBytes = s.mp3CoverMaxBytes
	browseRLPerMinute = s.browseRLPerMinute
	stashRLPerMinute = s.stashRLPerMinute
	maxTransferConc = s.transferMaxConcurrent
	mp3GateEnabled = s.mp3GateEnabled
	mp3GateTTL = s.mp3GateTTL
	mp3RefererCheck = s.mp3RefererCheck
	mp3MaxConcurrentPerIP = s.mp3MaxConcurrentPerIP
	mp3MaxTracksPerMinute = s.mp3MaxTracksPerMinute
	opsMu.Unlock()
}

func getOpsSnapshot() opsSnapshot {
	opsMu.RLock()
	defer opsMu.RUnlock()
	return opsCurrent
}

func usageBaselineBytes() int64 {
	s := getOpsSnapshot()
	if v := getenv("FS_USAGE_BASELINE_BYTES", ""); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return s.usageBaselineBytes
}

const (
	defaultSiteDisplayName = "igozhang"
	defaultSiteTagline     = "Practice makes perfect"
	siteTaglineMaxRunes    = 48
)

// siteDisplayName 返回全站页眉与首页 document.title 展示名（来自 igo.conf [site] title 的首词）。
func siteDisplayName() string {
	t := strings.TrimSpace(getOpsSnapshot().siteTitle)
	if t == "" {
		return defaultSiteDisplayName
	}
	return truncateRunes(t, 80)
}

// siteTagline 返回页眉站点名后的短句；未单独配置时使用 defaultSiteTagline。
func siteTagline() string {
	t := strings.TrimSpace(getOpsSnapshot().siteTagline)
	if t == "" {
		t = defaultSiteTagline
	}
	return truncateRunes(t, siteTaglineMaxRunes)
}

// parseSiteTitleConfig 解析 title =「网站名 短句」；仅一个词时短句由调用方填默认。
func parseSiteTitleConfig(val string) (name, tagline string, hasTagline bool) {
	val = strings.TrimSpace(val)
	if val == "" {
		return "", "", false
	}
	i := strings.IndexByte(val, ' ')
	if i < 0 {
		return val, "", false
	}
	return strings.TrimSpace(val[:i]), strings.TrimSpace(val[i+1:]), true
}

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func loadFooterConfigFromOps() footerConfig {
	s := getOpsSnapshot()
	return footerConfig{
		Gomoku:  s.footerGomoku,
		Xiangqi: s.footerXiangqi,
		Texas:   s.footerTexas,
		Links:   append([]footerLink(nil), s.footerLinks...),
	}
}

// opsFooterGomokuHref 将 [footer] gomoku 解析为可放入 href 的地址（含 BASE_PATH 前缀）。
func opsFooterGomokuHref(pub string) string {
	g := strings.TrimSpace(loadFooterConfigFromOps().Gomoku)
	if g == "" {
		g = pathGomoku
	}
	low := strings.ToLower(g)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		return g
	}
	return pubPath(pub, strings.TrimPrefix(g, "/"))
}

// opsFooterXiangqiHref 将 [footer] xiangqi 解析为可放入 href 的地址（含 BASE_PATH 前缀）。
func opsFooterXiangqiHref(pub string) string {
	g := strings.TrimSpace(loadFooterConfigFromOps().Xiangqi)
	if g == "" {
		g = pathXiangqi
	}
	low := strings.ToLower(g)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		return g
	}
	return pubPath(pub, strings.TrimPrefix(g, "/"))
}

// opsFooterTexasHref 将 [footer] texas 解析为可放入 href 的地址（含 BASE_PATH 前缀）。
func opsFooterTexasHref(pub string) string {
	g := strings.TrimSpace(loadFooterConfigFromOps().Texas)
	if g == "" {
		g = pathTexas
	}
	low := strings.ToLower(g)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		return g
	}
	return pubPath(pub, strings.TrimPrefix(g, "/"))
}

func initOpsConfig(dataRoot string) {
	opsDataRoot = dataRoot
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	if err := loadOpsConfig(dataRoot); err != nil {
		log.Printf("ops-config: load %v (using defaults)", err)
		applyOpsSnapshot(defaultOpsSnapshot())
	}
}

func ensureOpsConfigFile(dataRoot string) {
	p := opsConfigPath(dataRoot)
	if _, err := os.Stat(p); err == nil {
		return
	}
	if err := os.WriteFile(p, opsConfigExample, 0o644); err != nil {
		log.Printf("ops-config: write example %s: %v", p, err)
	}
}

func loadOpsConfig(dataRoot string) error {
	p := opsConfigPath(dataRoot)
	if _, err := os.Stat(p); err != nil {
		return err
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	if !secret07HasEffectiveLines(raw) {
		return os.ErrNotExist
	}
	s, err := parseOpsConfig(string(raw))
	if err != nil {
		return err
	}
	applyOpsSnapshot(s)
	if s.transferMaxConcurrent > 0 && s.transferMaxConcurrent <= 32 {
		maxTransferConc = s.transferMaxConcurrent
		if v := getenv("TRANSFER_MAX_CONCURRENT_PER_IP", ""); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 32 {
				maxTransferConc = n
			}
		}
	}
	rev := opsConfigRevFromRaw(raw)
	atomic.StoreUint64(&opsConfigRev, rev)
	log.Printf("ops-config: loaded rev=%d path=%s (restart pod to apply file changes)", rev, p)
	return nil
}

func opsConfigRevFromRaw(raw []byte) uint64 {
	sum := sha256.Sum256(raw)
	return uint64(sum[0])<<56 | uint64(sum[1])<<48 | uint64(sum[2])<<40 | uint64(sum[3])<<32 |
		uint64(sum[4])<<24 | uint64(sum[5])<<16 | uint64(sum[6])<<8 | uint64(sum[7])
}

func parseOpsConfig(text string) (opsSnapshot, error) {
	s := defaultOpsSnapshot()
	section := ""
	lines := strings.Split(text, "\n")
	var footerLinks []footerLink

	setDur := func(dst *time.Duration, key, val string) error {
		d, err := parseDuration(val)
		if err != nil {
			return fmtErr(key, err)
		}
		*dst = d
		return nil
	}
	setBytes := func(dst *int64, key, val string) error {
		n, err := parseByteSize(val)
		if err != nil {
			return fmtErr(key, err)
		}
		*dst = n
		return nil
	}
	setInt := func(dst *int, key, val string, min, max int) error {
		n, err := parseIntInRange(val, min, max)
		if err != nil {
			return fmtErr(key, err)
		}
		*dst = n
		return nil
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)

		switch section {
		case "site":
			switch key {
			case "title":
				if val == "" {
					break
				}
				name, tag, hasTag := parseSiteTitleConfig(val)
				if name != "" {
					s.siteTitle = name
				}
				if hasTag && tag != "" {
					s.siteTagline = tag
				} else if name != "" {
					s.siteTagline = defaultSiteTagline
				}
			case "beijing_offset_hours":
				n, err := parseBeijingOffsetHours(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.beijingOffsetHours = n
			case "banner_mask":
				if err := setInt(&s.siteBannerMaskPercent, key, val, 0, 100); err != nil {
					return s, err
				}
			}
		case "igokey":
			switch key {
			case "enabled":
				on, err := parseOnOff(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.igokeyEnabled = on
			}
		case "auth":
			switch key {
			case "enabled":
				on, err := parseOnOff(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.authEnabled = on
			}
		case "ops":
			switch key {
			case "enabled":
				on, err := parseOnOff(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.opsEnabled = on
			case "board_poll_interval":
				if err := setDur(&s.boardPollInterval, key, val); err != nil {
					return s, err
				}
			case "session_idle_timeout":
				if err := setDur(&s.visitorSessionIdleTimeout, key, val); err != nil {
					return s, err
				}
			}
		case "global":
			switch key {
			case "maintenance_interval":
				if err := setDur(&s.maintenanceInterval, key, val); err != nil {
					return s, err
				}
			}
		case "footer":
			switch key {
			case "gomoku":
				if val != "" {
					s.footerGomoku = val
				}
			case "xiangqi":
				if val != "" {
					s.footerXiangqi = val
				}
			case "texas":
				if val != "" {
					s.footerTexas = val
				}
			case "max_links":
				if err := setInt(&s.footerMaxLinks, key, val, 1, 20); err != nil {
					return s, err
				}
			case "link":
				parts := strings.Split(val, "|")
				if len(parts) >= 2 {
					footerLinks = append(footerLinks, footerLink{
						Text: strings.TrimSpace(parts[0]),
						Href: strings.TrimSpace(parts[1]),
					})
				}
			}
		case "stash":
			switch key {
			case "max_size":
				n, err := parseByteSize(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.maxStashBytes = int(n)
			case "ttl":
				if err := setDur(&s.stashTTL, key, val); err != nil {
					return s, err
				}
			case "poll_interval":
				if err := setDur(&s.stashPollInterval, key, val); err != nil {
					return s, err
				}
			case "poll_interval_max":
				if err := setDur(&s.stashPollIntervalMax, key, val); err != nil {
					return s, err
				}
			}
		case "transfer":
			switch key {
			case "file_max_age":
				if err := setDur(&s.fsFileMaxAge, key, val); err != nil {
					return s, err
				}
			case "dir_max_size":
				if err := setBytes(&s.transferDirMaxBytes, key, val); err != nil {
					return s, err
				}
			case "usage_baseline":
				if err := setBytes(&s.usageBaselineBytes, key, val); err != nil {
					return s, err
				}
			case "max_concurrent_per_ip":
				if err := setInt(&s.transferMaxConcurrent, key, val, 1, 32); err != nil {
					return s, err
				}
			case "upload_max_size":
				if err := setBytes(&s.uploadMaxBytes, key, val); err != nil {
					return s, err
				}
			case "fetch_timeout":
				if err := setDur(&s.fetchUpstreamTimeout, key, val); err != nil {
					return s, err
				}
			}
		case "ratelimit":
			switch key {
			case "browse_per_minute":
				if err := setInt(&s.browseRLPerMinute, key, val, 1, 10000); err != nil {
					return s, err
				}
			case "stash_per_minute":
				if err := setInt(&s.stashRLPerMinute, key, val, 1, 10000); err != nil {
					return s, err
				}
			}
		case "security":
			switch key {
			case "path_probe_enabled":
				on, err := parseOnOff(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.pathProbeEnabled = on
			case "path_probe_window":
				if err := setDur(&s.pathProbeWindow, key, val); err != nil {
					return s, err
				}
			case "path_probe_max_paths":
				if err := setInt(&s.pathProbeMaxPaths, key, val, 5, 100); err != nil {
					return s, err
				}
			case "path_probe_ban_duration":
				if err := setDur(&s.pathProbeBanDuration, key, val); err != nil {
					return s, err
				}
			case "path_probe_exclude_prefix":
				s.pathProbeExcludePrefixes = parseCommaSeparatedPrefixes(val)
			}
		case "log":
			switch key {
			case "file_max_age":
				if err := setDur(&s.logFileMaxAge, key, val); err != nil {
					return s, err
				}
			case "rotate_size":
				if err := setBytes(&s.logRotateSizeBytes, key, val); err != nil {
					return s, err
				}
			case "truncate_size":
				if err := setBytes(&s.logTruncateSizeBytes, key, val); err != nil {
					return s, err
				}
			case "dir_max_size":
				// 已废弃：目录总占用上限策略已移除，保留键以兼容旧配置。
			}
		case "preview":
			switch key {
			case "max_size":
				if err := setBytes(&s.previewMaxBytes, key, val); err != nil {
					return s, err
				}
			}
		case "doc":
			switch key {
			case "home_latest":
				if err := setInt(&s.docHomeLatest, key, val, 1, 50); err != nil {
					return s, err
				}
			case "nav_ball":
				on, err := parseOnOff(val)
				if err != nil {
					return s, err
				}
				s.docNavBallEnabled = on
			}
		case "mp3":
			switch key {
			case "min_buffer":
				n, err := parseByteSize(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.mp3MinBufferBytes = int(n)
			case "prefetch_ahead":
				if err := setInt(&s.mp3PrefetchAheadSec, key, val, 1, 600); err != nil {
					return s, err
				}
			case "stall_timeout":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.mp3StallMs = int(d.Milliseconds())
			case "max_stall_retries":
				if err := setInt(&s.mp3MaxStallRetries, key, val, 0, 32); err != nil {
					return s, err
				}
			case "cover_max_size":
				n, err := parseByteSize(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.mp3CoverMaxBytes = int(n)
			case "gate_enabled":
				on, err := parseOnOff(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.mp3GateEnabled = on
			case "gate_ttl":
				if err := setDur(&s.mp3GateTTL, key, val); err != nil {
					return s, err
				}
			case "referer_check":
				on, err := parseOnOff(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.mp3RefererCheck = on
			case "max_concurrent_per_ip":
				if err := setInt(&s.mp3MaxConcurrentPerIP, key, val, 1, 32); err != nil {
					return s, err
				}
			case "max_tracks_per_minute":
				if err := setInt(&s.mp3MaxTracksPerMinute, key, val, 5, 500); err != nil {
					return s, err
				}
			}
		case "xiangqi":
			switch key {
			case "easy_depth":
				if err := setInt(&s.xiangqiEasyDepth, key, val, 1, 6); err != nil {
					return s, err
				}
			case "easy_time_ms":
				if err := setInt(&s.xiangqiEasyTimeMs, key, val, 200, 10000); err != nil {
					return s, err
				}
			case "hard_depth":
				if err := setInt(&s.xiangqiHardDepth, key, val, 1, 8); err != nil {
					return s, err
				}
			case "hard_time_ms":
				if err := setInt(&s.xiangqiHardTimeMs, key, val, 500, 15000); err != nil {
					return s, err
				}
			}
		case "texas":
			switch key {
			case "starting_chips":
				if err := setInt(&s.texasStartingChips, key, val, 100, 1000000); err != nil {
					return s, err
				}
			case "small_blind":
				if err := setInt(&s.texasSmallBlind, key, val, 1, 10000); err != nil {
					return s, err
				}
			case "big_blind":
				if err := setInt(&s.texasBigBlind, key, val, 1, 10000); err != nil {
					return s, err
				}
			case "max_raises_per_round":
				if err := setInt(&s.texasMaxRaises, key, val, 0, 10); err != nil {
					return s, err
				}
			case "easy_think_ms":
				if err := setInt(&s.texasEasyThinkMs, key, val, 100, 10000); err != nil {
					return s, err
				}
			case "hard_think_ms":
				if err := setInt(&s.texasHardThinkMs, key, val, 200, 15000); err != nil {
					return s, err
				}
			}
		case "dice":
			switch key {
			case "roll_duration":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.diceRollMs = int(d.Milliseconds())
			case "result_show":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.diceResultShowMs = int(d.Milliseconds())
			}
		case "zhijian":
			switch key {
			case "fuse_duration":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.zhijianFuseMs = int(d.Milliseconds())
			case "fuse_combo_after":
				if err := setInt(&s.zhijianFuseComboAfter, key, val, 1, 99); err != nil {
					return s, err
				}
			case "fuse_combo_bonus":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.zhijianFuseComboBonus = int(d.Milliseconds())
			case "fuse_urgent":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.zhijianFuseUrgentMs = int(d.Milliseconds())
			case "session_duration":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.zhijianSessionSec = int(d.Seconds())
			case "min_record":
				d, err := parseDuration(val)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.zhijianMinRecordMs = int(d.Milliseconds())
			case "combo_accum_after":
				if err := setInt(&s.zhijianComboAccum, key, val, 1, 99); err != nil {
					return s, err
				}
			case "combo_danmaku_after":
				if err := setInt(&s.zhijianComboDanmaku, key, val, 1, 99); err != nil {
					return s, err
				}
			case "combo_danmaku_each_after":
				if err := setInt(&s.zhijianComboDanmakuEa, key, val, 1, 99); err != nil {
					return s, err
				}
			case "combo_decay_per_sec":
				f, err := parseFloatInRange(val, 0, 10)
				if err != nil {
					return s, fmtErr(key, err)
				}
				s.zhijianComboDecay = f
			case "title_cpm":
				parts := strings.Split(val, ",")
				var tiers []int
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					n, err := strconv.Atoi(p)
					if err != nil || n < 1 {
						return s, fmtErr(key, err)
					}
					tiers = append(tiers, n)
				}
				if len(tiers) >= 2 {
					s.zhijianTitleCPM = tiers
				}
			case "word3k_pool_size":
				if err := setInt(&s.zhijianWord3kPoolSize, key, val, 10, 500); err != nil {
					return s, err
				}
			}
		}
	}
	if len(footerLinks) > 0 {
		s.footerLinks = sanitizeFooterLinks(footerLinks, s.footerMaxLinks)
	}
	if s.usageBaselineBytes <= 0 {
		s.usageBaselineBytes = s.transferDirMaxBytes
	}
	if s.maintenanceInterval < time.Hour {
		s.maintenanceInterval = time.Hour
	}
	if s.stashPollInterval < 3*time.Second {
		s.stashPollInterval = 3 * time.Second
	}
	if s.stashPollIntervalMax < s.stashPollInterval {
		s.stashPollIntervalMax = s.stashPollInterval
	}
	if s.boardPollInterval < 3*time.Second {
		s.boardPollInterval = 3 * time.Second
	}
	if s.boardPollInterval > 60*time.Second {
		s.boardPollInterval = 60 * time.Second
	}
	if s.visitorSessionIdleTimeout < 5*time.Minute {
		s.visitorSessionIdleTimeout = 5 * time.Minute
	}
	if s.visitorSessionIdleTimeout > 6*time.Hour {
		s.visitorSessionIdleTimeout = 6 * time.Hour
	}
	if s.pathProbeWindow < 10*time.Second {
		s.pathProbeWindow = 10 * time.Second
	}
	if s.pathProbeWindow > 10*time.Minute {
		s.pathProbeWindow = 10 * time.Minute
	}
	if s.pathProbeMaxPaths < 5 {
		s.pathProbeMaxPaths = 5
	}
	if s.pathProbeMaxPaths > 100 {
		s.pathProbeMaxPaths = 100
	}
	if s.pathProbeBanDuration < 5*time.Minute {
		s.pathProbeBanDuration = 5 * time.Minute
	}
	if s.pathProbeBanDuration > 24*time.Hour {
		s.pathProbeBanDuration = 24 * time.Hour
	}
	if len(s.pathProbeExcludePrefixes) == 0 {
		s.pathProbeExcludePrefixes = []string{"/stream/mp3/", "/assets/doc-pic/", "/assets/doc/"}
	}
	if s.mp3GateTTL < time.Minute {
		s.mp3GateTTL = time.Minute
	}
	if s.mp3GateTTL > 7*24*time.Hour {
		s.mp3GateTTL = 7 * 24 * time.Hour
	}
	if s.mp3MaxConcurrentPerIP < 1 {
		s.mp3MaxConcurrentPerIP = 1
	}
	if s.mp3MaxTracksPerMinute < 5 {
		s.mp3MaxTracksPerMinute = 5
	}
	return s, nil
}

func parseCommaSeparatedPrefixes(val string) []string {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(val, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		out = append(out, p)
	}
	return out
}

func fmtErr(key string, err error) error {
	return &opsConfigError{key: key, err: err}
}

type opsConfigError struct {
	key string
	err error
}

func (e *opsConfigError) Error() string {
	return opsConfigName + ": " + e.key + ": " + e.err.Error()
}

func sanitizeFooterLinks(in []footerLink, max int) []footerLink {
	out := make([]footerLink, 0, len(in))
	for _, l := range in {
		if len(out) >= max {
			break
		}
		text := strings.TrimSpace(l.Text)
		href := strings.TrimSpace(l.Href)
		if text == "" || href == "" {
			continue
		}
		if !isSafeFooterHref(href) {
			continue
		}
		out = append(out, footerLink{Text: text, Href: href})
	}
	return out
}

func opsClientJSON() []byte {
	s := getOpsSnapshot()
	payload := map[string]any{
		"rev":                  atomic.LoadUint64(&opsConfigRev),
		"stashMaxBytes":        s.maxStashBytes,
		"stashPollIntervalMs":  int(s.stashPollInterval.Milliseconds()),
		"stashPollIntervalMax": int(s.stashPollIntervalMax.Milliseconds()),
		"mp3": map[string]any{
			"minBufferBytes":   s.mp3MinBufferBytes,
			"prefetchAheadSec": s.mp3PrefetchAheadSec,
			"stallMs":          s.mp3StallMs,
			"maxStallRetries":  s.mp3MaxStallRetries,
		},
		"dice": map[string]any{
			"rollMs":       s.diceRollMs,
			"resultShowMs": s.diceResultShowMs,
		},
		"zhijian": map[string]any{
			"fuseMs":              s.zhijianFuseMs,
			"fuseComboAfter":      s.zhijianFuseComboAfter,
			"fuseComboBonusMs":    s.zhijianFuseComboBonus,
			"fuseUrgentMs":        s.zhijianFuseUrgentMs,
			"sessionSec":          s.zhijianSessionSec,
			"minRecordMs":         s.zhijianMinRecordMs,
			"comboAccumAfter":     s.zhijianComboAccum,
			"comboDanmakuAfter":   s.zhijianComboDanmaku,
			"comboDanmakuEach":  s.zhijianComboDanmakuEa,
			"comboDecayPerSec":    s.zhijianComboDecay,
			"titleCpm":            s.zhijianTitleCPM,
			"word3kPoolSize":      s.zhijianWord3kPoolSize,
		},
	}
	b, _ := json.Marshal(payload)
	return b
}

// 仅注入 mp3 子集
func opsMp3OnlyInjectScript() string {
	s := getOpsSnapshot()
	payload := map[string]any{
		"minBufferBytes":   s.mp3MinBufferBytes,
		"prefetchAheadSec": s.mp3PrefetchAheadSec,
		"stallMs":          s.mp3StallMs,
		"maxStallRetries":  s.mp3MaxStallRetries,
	}
	b, _ := json.Marshal(payload)
	return `<script>window.__fsMp3Ops=` + string(b) + `;</script>`
}

func opsZjInjectScript() string {
	s := getOpsSnapshot()
	payload := map[string]any{
		"fuseMs":             s.zhijianFuseMs,
		"fuseComboAfter":     s.zhijianFuseComboAfter,
		"fuseComboBonusMs":   s.zhijianFuseComboBonus,
		"fuseUrgentMs":       s.zhijianFuseUrgentMs,
		"sessionSec":         s.zhijianSessionSec,
		"minRecordMs":        s.zhijianMinRecordMs,
		"comboAccumAfter":    s.zhijianComboAccum,
		"comboDanmakuAfter":  s.zhijianComboDanmaku,
		"comboDanmakuEach":   s.zhijianComboDanmakuEa,
		"comboDecayPerSec":   s.zhijianComboDecay,
		"titleCpm":           s.zhijianTitleCPM,
		"word3kPoolSize":     s.zhijianWord3kPoolSize,
	}
	b, _ := json.Marshal(payload)
	return `<script>window.__fsZjOps=` + string(b) + `;</script>`
}

func opsDiceInjectScript() string {
	s := getOpsSnapshot()
	payload := map[string]any{
		"rollMs":       s.diceRollMs,
		"resultShowMs": s.diceResultShowMs,
	}
	b, _ := json.Marshal(payload)
	return `<script>window.__fsDiceOps=` + string(b) + `;</script>`
}

func opsXiangqiInjectScript() string {
	s := getOpsSnapshot()
	payload := map[string]any{
		"easyDepth":   s.xiangqiEasyDepth,
		"easyTimeMs":  s.xiangqiEasyTimeMs,
		"hardDepth":   s.xiangqiHardDepth,
		"hardTimeMs":  s.xiangqiHardTimeMs,
	}
	b, _ := json.Marshal(payload)
	return `<script>window.__fsXiangqiOps=` + string(b) + `;</script>`
}

func opsTexasInjectScript() string {
	s := getOpsSnapshot()
	payload := map[string]any{
		"startingChips":    s.texasStartingChips,
		"smallBlind":       s.texasSmallBlind,
		"bigBlind":         s.texasBigBlind,
		"maxRaisesPerRound": s.texasMaxRaises,
		"easyThinkMs":      s.texasEasyThinkMs,
		"hardThinkMs":      s.texasHardThinkMs,
	}
	b, _ := json.Marshal(payload)
	return `<script>window.__fsTexasOps=` + string(b) + `;</script>`
}

func opsStashInjectScript() string {
	s := getOpsSnapshot()
	payload := map[string]any{
		"maxBytes":          s.maxStashBytes,
		"pollIntervalMs":    int(s.stashPollInterval.Milliseconds()),
		"pollIntervalMaxMs": int(s.stashPollIntervalMax.Milliseconds()),
	}
	b, _ := json.Marshal(payload)
	return `<script>window.__fsStashOps=` + string(b) + `;</script>`
}
