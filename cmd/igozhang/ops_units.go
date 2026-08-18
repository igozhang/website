package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseByteSize 解析带单位的大小：B、K/KB/KiB、M、G（二进制 1024 阶）。
func parseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	low := strings.ToLower(s)
	mult := int64(1)
	switch {
	case strings.HasSuffix(low, "kib"):
		low, mult = strings.TrimSuffix(low, "kib"), 1024
	case strings.HasSuffix(low, "mib"):
		low, mult = strings.TrimSuffix(low, "mib"), 1024*1024
	case strings.HasSuffix(low, "gib"):
		low, mult = strings.TrimSuffix(low, "gib"), 1024*1024*1024
	case strings.HasSuffix(low, "kb"):
		low, mult = strings.TrimSuffix(low, "kb"), 1000
	case strings.HasSuffix(low, "mb"):
		low, mult = strings.TrimSuffix(low, "mb"), 1000*1000
	case strings.HasSuffix(low, "gb"):
		low, mult = strings.TrimSuffix(low, "gb"), 1000*1000*1000
	case strings.HasSuffix(low, "k"):
		low, mult = strings.TrimSuffix(low, "k"), 1024
	case strings.HasSuffix(low, "m"):
		low, mult = strings.TrimSuffix(low, "m"), 1024*1024
	case strings.HasSuffix(low, "g"):
		low, mult = strings.TrimSuffix(low, "g"), 1024*1024*1024
	case strings.HasSuffix(low, "b"):
		low = strings.TrimSuffix(low, "b")
		mult = 1
	}
	low = strings.TrimSpace(low)
	n, err := strconv.ParseFloat(low, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("negative size %q", s)
	}
	v := int64(n * float64(mult))
	if v < 0 {
		return 0, fmt.Errorf("size overflow %q", s)
	}
	return v, nil
}

// parseDuration 解析时长：ms、s、m、h、d（天=24h）。
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	low := strings.ToLower(s)
	switch {
	case strings.HasSuffix(low, "ms"):
		n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(low, "ms")), 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(n * float64(time.Millisecond)), nil
	case strings.HasSuffix(low, "min"):
		n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(low, "min")), 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(n * float64(time.Minute)), nil
	case strings.HasSuffix(low, "h"):
		n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(low, "h")), 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(n * float64(time.Hour)), nil
	case strings.HasSuffix(low, "d"):
		n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(low, "d")), 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(n * 24 * float64(time.Hour)), nil
	case strings.HasSuffix(low, "m"):
		n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(low, "m")), 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(n * float64(time.Minute)), nil
	case strings.HasSuffix(low, "s"):
		n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(low, "s")), 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(n * float64(time.Second)), nil
	}
	// 纯数字视为秒
	if n, err := strconv.ParseFloat(low, 64); err == nil {
		return time.Duration(n * float64(time.Second)), nil
	}
	return 0, fmt.Errorf("invalid duration %q", s)
}

func parseIntInRange(s string, min, max int) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty integer")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n < min || n > max {
		return 0, fmt.Errorf("value %d out of range [%d,%d]", n, min, max)
	}
	return n, nil
}

func parseFloatInRange(s string, min, max float64) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty float")
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if n < min || n > max {
		return 0, fmt.Errorf("value %g out of range [%g,%g]", n, min, max)
	}
	return n, nil
}

// parseBeijingOffsetHours 解析北京时间显示修正：± 与 0–12，如 -1、+2、0（默认 0 不偏移）。
func parseBeijingOffsetHours(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0, nil
	}
	sign := 1
	if strings.HasPrefix(s, "+") {
		s = strings.TrimSpace(s[1:])
	} else if strings.HasPrefix(s, "-") {
		sign = -1
		s = strings.TrimSpace(s[1:])
	}
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid beijing_offset_hours %q", s)
	}
	if n < 0 || n > 12 {
		return 0, fmt.Errorf("beijing_offset_hours must be between 0 and 12")
	}
	return sign * n, nil
}

// parseOnOff 解析 on/off（亦接受 yes/no、1/0、true/false，大小写不敏感）。
func parseOnOff(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "on", "yes", "1", "true":
		return true, nil
	case "off", "no", "0", "false":
		return false, nil
	default:
		return false, fmt.Errorf("want on or off")
	}
}
