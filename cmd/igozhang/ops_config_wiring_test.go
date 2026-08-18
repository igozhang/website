package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 自定义配置：用于验证解析 → applyOpsSnapshot → 注入脚本 / 全局变量 全链路。
const opsWiringSample = `
[global]
maintenance_interval = 2h

[footer]
gomoku = /gomoku-custom
max_links = 2
link = A | https://a.example
link = B | https://b.example

[stash]
max_size = 64K
ttl = 1h
poll_interval = 5s
poll_interval_max = 45s

[transfer]
file_max_age = 1d
dir_max_size = 1G
usage_baseline = 2G
max_concurrent_per_ip = 4
upload_max_size = 100M
fetch_timeout = 5m

[ratelimit]
browse_per_minute = 10
stash_per_minute = 11

[log]
file_max_age = 2d
rotate_size = 400K
truncate_size = 2M

[preview]
max_size = 1M

[mp3]
min_buffer = 1M
prefetch_ahead = 9
stall_timeout = 2s
max_stall_retries = 2
cover_max_size = 3M

[dice]
roll_duration = 800ms
result_show = 1s

[zhijian]
fuse_duration = 3s
fuse_combo_after = 7
fuse_combo_bonus = 600ms
fuse_urgent = 900ms
session_duration = 120s
min_record = 45s
combo_accum_after = 4
combo_danmaku_after = 6
combo_danmaku_each_after = 9
combo_decay_per_sec = 0.05
title_cpm = 10,20,30
`

func TestParseOpsConfig_AllSections(t *testing.T) {
	s, err := parseOpsConfig(opsWiringSample)
	if err != nil {
		t.Fatal(err)
	}
	if s.maintenanceInterval != 2*time.Hour {
		t.Fatalf("maintenance_interval: %v", s.maintenanceInterval)
	}
	if s.footerGomoku != "/gomoku-custom" || len(s.footerLinks) != 2 {
		t.Fatalf("footer: gomoku=%q links=%d", s.footerGomoku, len(s.footerLinks))
	}
	if s.maxStashBytes != 64*1024 || s.stashTTL != time.Hour {
		t.Fatalf("stash: size=%d ttl=%v", s.maxStashBytes, s.stashTTL)
	}
	if s.stashPollInterval != 5*time.Second || s.stashPollIntervalMax != 45*time.Second {
		t.Fatalf("stash poll: %v %v", s.stashPollInterval, s.stashPollIntervalMax)
	}
	if s.fsFileMaxAge != 24*time.Hour || s.transferDirMaxBytes != 1<<30 {
		t.Fatalf("transfer: age=%v dir=%d", s.fsFileMaxAge, s.transferDirMaxBytes)
	}
	if s.usageBaselineBytes != 2<<30 || s.transferMaxConcurrent != 4 {
		t.Fatalf("transfer: baseline=%d conc=%d", s.usageBaselineBytes, s.transferMaxConcurrent)
	}
	if s.uploadMaxBytes != 100<<20 || s.fetchUpstreamTimeout != 5*time.Minute {
		t.Fatalf("transfer: upload=%d fetch=%v", s.uploadMaxBytes, s.fetchUpstreamTimeout)
	}
	if s.browseRLPerMinute != 10 || s.stashRLPerMinute != 11 {
		t.Fatalf("ratelimit: browse=%d stash=%d", s.browseRLPerMinute, s.stashRLPerMinute)
	}
	if s.logFileMaxAge != 48*time.Hour || s.logRotateSizeBytes != 400*1024 || s.logTruncateSizeBytes != 2<<20 {
		t.Fatalf("log: age=%v rotate=%d truncate=%d", s.logFileMaxAge, s.logRotateSizeBytes, s.logTruncateSizeBytes)
	}
	if s.previewMaxBytes != 1<<20 {
		t.Fatalf("preview: %d", s.previewMaxBytes)
	}
	if s.mp3MinBufferBytes != 1<<20 || s.mp3PrefetchAheadSec != 9 || s.mp3StallMs != 2000 || s.mp3MaxStallRetries != 2 {
		t.Fatalf("mp3: buf=%d ahead=%d stall=%d retries=%d", s.mp3MinBufferBytes, s.mp3PrefetchAheadSec, s.mp3StallMs, s.mp3MaxStallRetries)
	}
	if s.mp3CoverMaxBytes != 3<<20 {
		t.Fatalf("mp3 cover: %d", s.mp3CoverMaxBytes)
	}
	if s.diceRollMs != 800 || s.diceResultShowMs != 1000 {
		t.Fatalf("dice: roll=%d show=%d", s.diceRollMs, s.diceResultShowMs)
	}
	if s.zhijianFuseMs != 3000 || s.zhijianSessionSec != 120 || len(s.zhijianTitleCPM) != 3 {
		t.Fatalf("zhijian: fuse=%d session=%d tiers=%v", s.zhijianFuseMs, s.zhijianSessionSec, s.zhijianTitleCPM)
	}
}

func TestApplyOpsSnapshot_WiresServerGlobals(t *testing.T) {
	s, err := parseOpsConfig(opsWiringSample)
	if err != nil {
		t.Fatal(err)
	}
	applyOpsSnapshot(s)
	if maxStashBytes != 64*1024 {
		t.Fatalf("maxStashBytes=%d", maxStashBytes)
	}
	if stashTTL != time.Hour || fsFileMaxAge != 24*time.Hour {
		t.Fatalf("ttl=%v fileAge=%v", stashTTL, fsFileMaxAge)
	}
	if transferDirMaxBytes != 1<<30 || maxPreviewBytes != 1<<20 {
		t.Fatalf("transfer/preview: %d %d", transferDirMaxBytes, maxPreviewBytes)
	}
	if maxCoverBytes != 3<<20 || browseRLPerMinute != 10 || stashRLPerMinute != 11 {
		t.Fatalf("cover/rl: %d %d %d", maxCoverBytes, browseRLPerMinute, stashRLPerMinute)
	}
	if logRotateSizeBytes != 400*1024 || logTruncateSizeBytes != 2<<20 || logFileMaxAge != 48*time.Hour {
		t.Fatalf("log: rotate=%d truncate=%d age=%v", logRotateSizeBytes, logTruncateSizeBytes, logFileMaxAge)
	}
	if maxTransferConc != 4 {
		t.Fatalf("maxTransferConc=%d", maxTransferConc)
	}
	applyOpsSnapshot(defaultOpsSnapshot())
	initTransferLimits()
}

func TestOpsInjectScripts_UseSnapshot(t *testing.T) {
	s, err := parseOpsConfig(opsWiringSample)
	if err != nil {
		t.Fatal(err)
	}
	applyOpsSnapshot(s)
	stash := opsStashInjectScript()
	if !strings.Contains(stash, `"maxBytes":65536`) || !strings.Contains(stash, `"pollIntervalMs":5000`) {
		t.Fatalf("stash inject: %s", stash)
	}
	mp3 := opsMp3OnlyInjectScript()
	if !strings.Contains(mp3, `"prefetchAheadSec":9`) || !strings.Contains(mp3, `"stallMs":2000`) {
		t.Fatalf("mp3 inject: %s", mp3)
	}
	dice := opsDiceInjectScript()
	if !strings.Contains(dice, `"rollMs":800`) || !strings.Contains(dice, `"resultShowMs":1000`) {
		t.Fatalf("dice inject: %s", dice)
	}
	zj := opsZjInjectScript()
	if !strings.Contains(zj, `"fuseMs":3000`) || !strings.Contains(zj, `"sessionSec":120`) {
		t.Fatalf("zj inject: %s", zj)
	}
	applyOpsSnapshot(defaultOpsSnapshot())
}

func TestOpsClientJSON_MatchesSnapshot(t *testing.T) {
	s, err := parseOpsConfig(opsWiringSample)
	if err != nil {
		t.Fatal(err)
	}
	applyOpsSnapshot(s)
	raw := string(opsClientJSON())
	for _, want := range []string{
		`"stashMaxBytes":65536`,
		`"stashPollIntervalMs":5000`,
		`"rollMs":800`,
		`"fuseMs":3000`,
		`"sessionSec":120`,
	} {
		if !strings.Contains(raw, want) {
			t.Fatalf("opsClientJSON missing %q in %s", want, raw)
		}
	}
	applyOpsSnapshot(defaultOpsSnapshot())
}

func TestOpsFooterGomokuHref(t *testing.T) {
	s, err := parseOpsConfig("[footer]\ngomoku = /gomoku-custom\n")
	if err != nil {
		t.Fatal(err)
	}
	applyOpsSnapshot(s)
	if got := opsFooterGomokuHref("/fs"); got != "/fs/gomoku-custom" {
		t.Fatalf("got %q", got)
	}
	applyOpsSnapshot(defaultOpsSnapshot())
}

func TestLoadOpsConfig_FromDataRoot(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, secret07RelDir)
	if err := os.MkdirAll(secret, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, opsConfigName), []byte(opsWiringSample), 0o644); err != nil {
		t.Fatal(err)
	}
	applyOpsSnapshot(defaultOpsSnapshot())
	if err := loadOpsConfig(dir); err != nil {
		t.Fatal(err)
	}
	if maxStashBytes != 64*1024 {
		t.Fatalf("after load: maxStashBytes=%d", maxStashBytes)
	}
	applyOpsSnapshot(defaultOpsSnapshot())
	initTransferLimits()
}

func TestFsPage_StashOpsBeforeStashScript(t *testing.T) {
	out := string(renderFsHTML("", nil))
	if !strings.Contains(out, "window.__fsStashOps") && !strings.Contains(out, "__fsStashOps") {
		t.Fatal("fs page must inject stash ops config")
	}
	if !strings.Contains(out, "window.__fsStashInit") {
		t.Fatal("fs page must load stash module")
	}
}

func TestLoadFooterConfigFromOps(t *testing.T) {
	s, err := parseOpsConfig(opsWiringSample)
	if err != nil {
		t.Fatal(err)
	}
	applyOpsSnapshot(s)
	cfg := loadFooterConfigFromOps()
	if cfg.Gomoku != "/gomoku-custom" || len(cfg.Links) != 2 {
		t.Fatalf("footer api: %+v", cfg)
	}
	applyOpsSnapshot(defaultOpsSnapshot())
}
