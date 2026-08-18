package main

import "time"

const beijingBaseOffsetHours = 8

// beijingDisplayOffsetSec 返回用于显示的 UTC→本地偏移秒数（8 小时 + igo.conf 修正）。
func beijingDisplayOffsetSec() int64 {
	h := getOpsSnapshot().beijingOffsetHours
	if h > 12 {
		h = 12
	} else if h < -12 {
		h = -12
	}
	return int64(beijingBaseOffsetHours+h) * 3600
}

// formatBeijingUnix 将 Unix 秒（UTC 瞬间）格式化为北京时间字符串。
func formatBeijingUnix(unixSec int64) string {
	if unixSec <= 0 {
		return ""
	}
	return time.Unix(unixSec+beijingDisplayOffsetSec(), 0).UTC().Format("2006-01-02 15:04:05")
}

// formatBeijingCompact 用于文件名等紧凑场景（UTC 瞬间 + igo.conf 偏移，如 20060102-1504）。
func formatBeijingCompact(unixSec int64) string {
	if unixSec <= 0 {
		unixSec = time.Now().Unix()
	}
	return time.Unix(unixSec+beijingDisplayOffsetSec(), 0).UTC().Format("20060102-1504")
}
