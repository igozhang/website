package main

import (
	"html"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/igokorea/igozhang/internal/ipgeo"
)

var (
	ipgeoDB   *ipgeo.DB
	ipgeoOnce sync.Once
)

var continentNameZH = map[string]string{
	"AS": "亚洲",
	"EU": "欧洲",
	"AF": "非洲",
	"NA": "北美洲",
	"SA": "南美洲",
	"OC": "大洋洲",
	"AN": "南极洲",
}

var continentNameEN = map[string]string{
	"AS": "Asia",
	"EU": "Europe",
	"AF": "Africa",
	"NA": "North America",
	"SA": "South America",
	"OC": "Oceania",
	"AN": "Antarctica",
}

func continentDisplayName(c ipgeo.Country, code string) string {
	if code == "" || code == "XX" {
		return ""
	}
	if c.ChinesePrimary {
		return continentNameZH[code]
	}
	return continentNameEN[code]
}

func initIPGeo() {
	ipgeoOnce.Do(func() {
		db, err := ipgeo.OpenEmbedded()
		if err != nil {
			log.Printf("ipgeo: embedded load failed: %v", err)
			return
		}
		ipgeoDB = db
		c, l1, l2, w, cpt := ipgeoDB.Stats()
		log.Printf("ipgeo: embedded countries=%d v4L1=%d v4L2=%d v6wide=%d v6compact=%d", c, l1, l2, w, cpt)
	})
}

func visitorClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	ipStr := strings.TrimSpace(clientIP(r))
	if ipStr == "" {
		return ""
	}
	if ip := net.ParseIP(ipStr); ip != nil {
		return ip.String()
	}
	return ipStr
}

func visitorGeoParts(res *ipgeo.Result, countryLang string, continentEN bool) []string {
	if res == nil {
		return nil
	}
	var parts []string
	if name := strings.TrimSpace(res.Country.Name(countryLang)); name != "" {
		parts = append(parts, name)
	}
	if cc := strings.TrimSpace(res.Country.Code); cc != "" {
		parts = append(parts, cc)
	}
	cont := strings.TrimSpace(res.Country.ContinentCode)
	if cont != "" && cont != "XX" {
		var contName string
		if continentEN {
			contName = continentNameEN[cont]
		} else {
			contName = continentDisplayName(res.Country, cont)
		}
		if contName != "" {
			parts = append(parts, contName)
		}
		parts = append(parts, cont)
	}
	return parts
}

// visitorGeoLogEN 返回访问日志用英文 geo 摘要；未命中返回 "-"。
func visitorGeoLogEN(r *http.Request) (ip, geo string) {
	ip = visitorClientIP(r)
	if ip == "" {
		return "-", "-"
	}
	if ipgeoDB == nil {
		return ip, "-"
	}
	res, ok := ipgeoDB.Lookup(strings.TrimSpace(clientIP(r)))
	if !ok {
		return ip, "-"
	}
	parts := visitorGeoParts(res, "en", true)
	if len(parts) == 0 {
		return ip, "-"
	}
	return ip, strings.Join(parts, " ")
}

// countryDisplayNameZH 看板用中文国名；优先按二字代码查库，否则保留已有中文或回退原文。
func countryDisplayNameZH(code, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback == "" || fallback == "未知" {
		return "未知"
	}
	if ipgeoDB != nil {
		code = strings.TrimSpace(strings.ToUpper(code))
		if code != "" {
			if c, ok := ipgeoDB.CountryByCode(code); ok {
				if name := strings.TrimSpace(c.Name("zh")); name != "" {
					return truncateCountryNameRunes(name, accessLogCountryNameMaxRunes)
				}
			}
		}
	}
	if containsCJKRunes(fallback) {
		return truncateCountryNameRunes(fallback, accessLogCountryNameMaxRunes)
	}
	return truncateCountryNameRunes(fallback, accessLogCountryNameMaxRunes)
}

// formatAccessGeoBoardZH 将日志 geo 串格式化为看板用「中文国名 代码」。
func formatAccessGeoBoardZH(geo string) string {
	geo = strings.TrimSpace(geo)
	if geo == "" || geo == "-" {
		return ""
	}
	name, code := parseAccessLogGeoFields(geo)
	label := countryDisplayNameZH(code, name)
	if code != "" {
		return label + " " + code
	}
	return label
}

func containsCJKRunes(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// visitorGeoLabel 格式化页脚来访 IP；空串表示不渲染。
// 命中时按「国家 国家代码 大洲 大洲代码 IP」拼接有值字段；未命中仅 IP。
func visitorGeoLabel(r *http.Request) string {
	ip := visitorClientIP(r)
	if ip == "" {
		return ""
	}
	if ipgeoDB == nil {
		return html.EscapeString(ip)
	}
	res, ok := ipgeoDB.Lookup(strings.TrimSpace(clientIP(r)))
	if !ok {
		return html.EscapeString(ip)
	}
	parts := visitorGeoParts(res, "auto", false)
	parts = append(parts, ip)
	return html.EscapeString(strings.Join(parts, " "))
}
