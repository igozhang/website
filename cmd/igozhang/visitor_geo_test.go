package main

import (
	"net/http"
	"strings"
	"testing"
)

func ensureIPGeoLoaded(t *testing.T) {
	t.Helper()
	initIPGeo()
	if ipgeoDB == nil {
		t.Fatal("ipgeo DB not loaded")
	}
}

func TestCountryDisplayNameZH(t *testing.T) {
	ensureIPGeoLoaded(t)
	if got := countryDisplayNameZH("US", "United States"); got != "美国" {
		t.Fatalf("US=%q", got)
	}
	if got := countryDisplayNameZH("CN", "中国"); got != "中国" {
		t.Fatalf("CN=%q", got)
	}
	if got := formatAccessGeoBoardZH("United States US North America NA"); got != "美国 US" {
		t.Fatalf("board=%q", got)
	}
}

func TestVisitorGeoLabel_ipv4Hit(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "114.114.114.114")
	got := visitorGeoLabel(r)
	want := "中国 CN 亚洲 AS 114.114.114.114"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestVisitorGeoLabel_ipv4US(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "8.8.8.8")
	got := visitorGeoLabel(r)
	want := "United States US North America NA 8.8.8.8"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestVisitorGeoLabel_ipv6Hit(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "2400:3200::1")
	got := visitorGeoLabel(r)
	want := "中国 CN 亚洲 AS 2400:3200::1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestVisitorGeoLabel_hungaryEnglishContinent(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "188.44.128.106")
	got := visitorGeoLabel(r)
	want := "Hungary HU Europe EU 188.44.128.106"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestVisitorGeoLogEN_hungary(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "188.44.128.106")
	ip, geo := visitorGeoLogEN(r)
	if ip != "188.44.128.106" {
		t.Fatalf("ip %q", ip)
	}
	if geo != "Hungary HU Europe EU" {
		t.Fatalf("geo %q", geo)
	}
}

func TestVisitorGeoLogEN_chinaEnglish(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "114.114.114.114")
	_, geo := visitorGeoLogEN(r)
	if geo != "China CN Asia AS" {
		t.Fatalf("geo %q want English", geo)
	}
}

func TestVisitorGeoLogEN_miss(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "192.168.1.1")
	ip, geo := visitorGeoLogEN(r)
	if ip != "192.168.1.1" || geo != "-" {
		t.Fatalf("ip=%q geo=%q", ip, geo)
	}
}

func TestVisitorGeoLabel_ipv6UnknownOnlyIP(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "2001:db8::1")
	got := visitorGeoLabel(r)
	if got != "2001:db8::1" {
		t.Fatalf("got %q want bare IPv6", got)
	}
	if strings.Contains(got, "未知") {
		t.Fatal("must not show unknown label")
	}
}

func TestVisitorGeoLabel_privateIPv4OnlyIP(t *testing.T) {
	ensureIPGeoLoaded(t)
	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("X-Real-IP", "192.168.1.10")
	got := visitorGeoLabel(r)
	if got != "192.168.1.10" {
		t.Fatalf("got %q want bare private IP", got)
	}
}

func TestSiteFootBlock_visitorLabel(t *testing.T) {
	html := siteFootBlock("", "/", "/", "/", "/", "中国 CN 亚洲 AS 1.2.3.4")
	if !strings.Contains(html, `class="site-foot-visitor"`) {
		t.Fatal("missing visitor span")
	}
	if !strings.Contains(html, "中国 CN 亚洲 AS 1.2.3.4") {
		t.Fatal("missing visitor label text")
	}
	if !strings.Contains(html, `class="site-foot-links-nav"`) {
		t.Fatal("missing links nav wrapper")
	}
}

func TestSiteFootBlock_noVisitorLabel(t *testing.T) {
	html := siteFootBlock("", "/", "/", "/", "/", "")
	if strings.Contains(html, "site-foot-visitor") {
		t.Fatal("visitor span should be omitted when label empty")
	}
}
