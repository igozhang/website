package main

import (
	"testing"
	"time"
)

func TestParseOpsConfig_IgnoresDeprecatedReloadInterval(t *testing.T) {
	const legacy = `
[global]
reload_interval = 30s
maintenance_interval = 2h
`
	s, err := parseOpsConfig(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if s.maintenanceInterval != 2*time.Hour {
		t.Fatalf("maintenance=%v", s.maintenanceInterval)
	}
}

func TestParseOpsConfigExample(t *testing.T) {
	s, err := parseOpsConfig(string(opsConfigExample))
	if err != nil {
		t.Fatal(err)
	}
	if s.maxStashBytes != 200*1024 {
		t.Fatalf("maxStashBytes=%d", s.maxStashBytes)
	}
	if s.zhijianFuseMs != 4000 {
		t.Fatalf("fuseMs=%d", s.zhijianFuseMs)
	}
	if s.transferDirMaxBytes != 9*1024*1024*1024 {
		t.Fatalf("dir max=%d", s.transferDirMaxBytes)
	}
	if s.maintenanceInterval != time.Hour {
		t.Fatalf("maintenance=%v", s.maintenanceInterval)
	}
	if len(s.footerLinks) != 1 || s.footerLinks[0].Href != "https://igozhang.cn" {
		t.Fatalf("footer=%+v", s.footerLinks)
	}
	if s.siteTitle != defaultSiteDisplayName {
		t.Fatalf("site title=%q", s.siteTitle)
	}
}

func TestParseOpsConfig_BeijingOffsetHours(t *testing.T) {
	s, err := parseOpsConfig("[site]\nbeijing_offset_hours = -1\n")
	if err != nil {
		t.Fatal(err)
	}
	if s.beijingOffsetHours != -1 {
		t.Fatalf("offset=%d", s.beijingOffsetHours)
	}
	if _, err := parseOpsConfig("[site]\nbeijing_offset_hours = 13\n"); err == nil {
		t.Fatal("expected error for 13")
	}
}

func TestParseOpsConfig_SiteTitle(t *testing.T) {
	s, err := parseOpsConfig("[site]\ntitle = igozhang Custom line\n")
	if err != nil {
		t.Fatal(err)
	}
	if s.siteTitle != "igozhang" || s.siteTagline != "Custom line" {
		t.Fatalf("title=%q tagline=%q", s.siteTitle, s.siteTagline)
	}
	s2, err := parseOpsConfig("[site]\ntitle = igozhang\n")
	if err != nil {
		t.Fatal(err)
	}
	if s2.siteTitle != "igozhang" || s2.siteTagline != defaultSiteTagline {
		t.Fatalf("default tagline: title=%q tagline=%q", s2.siteTitle, s2.siteTagline)
	}
	if siteTagline() != defaultSiteTagline {
		t.Fatalf("siteTagline()=%q", siteTagline())
	}
}

func TestParseOpsConfig_IgokeyEnabled(t *testing.T) {
	s, err := parseOpsConfig("[igokey]\nenabled = off\n")
	if err != nil {
		t.Fatal(err)
	}
	if s.igokeyEnabled {
		t.Fatal("expected igokey off")
	}
	s2, err := parseOpsConfig(string(opsConfigExample))
	if err != nil || !s2.igokeyEnabled {
		t.Fatalf("example igokey=%v err=%v", s2.igokeyEnabled, err)
	}
}
