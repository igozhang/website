package main

import (
	"strings"
	"testing"
)

func TestSiteBannerMaskGradient(t *testing.T) {
	opsMu.Lock()
	opsCurrent.siteBannerMaskPercent = siteBannerMaskDefault
	opsMu.Unlock()
	gDef := siteBannerMaskGradient()
	if !strings.Contains(gDef, "0.3000") || !strings.Contains(gDef, "0.4800") || !strings.Contains(gDef, "0.6600") {
		t.Fatalf("default %d: %q", siteBannerMaskDefault, gDef)
	}

	opsMu.Lock()
	opsCurrent.siteBannerMaskPercent = 100
	opsMu.Unlock()
	g100 := siteBannerMaskGradient()
	if !strings.Contains(g100, "1.0000") {
		t.Fatalf("100%% block: %q", g100)
	}

	opsMu.Lock()
	opsCurrent.siteBannerMaskPercent = 0
	opsMu.Unlock()
	g0 := siteBannerMaskGradient()
	if !strings.Contains(g0, "0.0000") {
		t.Fatalf("0%% transparent: %q", g0)
	}
}

func TestParseOpsConfig_BannerMask(t *testing.T) {
	s, err := parseOpsConfig("[site]\nbanner_mask = 40\n")
	if err != nil {
		t.Fatal(err)
	}
	if s.siteBannerMaskPercent != 40 {
		t.Fatalf("got %d", s.siteBannerMaskPercent)
	}
	s2, err := parseOpsConfig(string(opsConfigExample))
	if err != nil {
		t.Fatal(err)
	}
	if s2.siteBannerMaskPercent != siteBannerMaskDefault {
		t.Fatalf("example default=%d want %d", s2.siteBannerMaskPercent, siteBannerMaskDefault)
	}
	if _, err := parseOpsConfig("[site]\nbanner_mask = 101\n"); err == nil {
		t.Fatal("expected range error")
	}
}
