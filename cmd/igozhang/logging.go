package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// logRelDir 位于 DATA_ROOT 下（不以 f 开头，不出现在目录浏览白名单）；服务端写入运行日志。
const logRelDir = "secret-log"

const appLogFileName = "igozhang.log"

const logLastDayMarker = ".last-day"

const appLogArchiveStem = "igozhang"

// parseAppLogDailyArchiveDay 解析 igozhang.YYYY-MM-DD.log。
func parseAppLogDailyArchiveDay(name string) (time.Time, bool) {
	if !strings.HasPrefix(name, appLogArchiveStem+".") || !strings.HasSuffix(name, ".log") {
		return time.Time{}, false
	}
	mid := strings.TrimSuffix(strings.TrimPrefix(name, appLogArchiveStem+"."), ".log")
	if strings.Contains(mid, ".") {
		return time.Time{}, false
	}
	if len(mid) != 10 || mid[4] != '-' || mid[7] != '-' {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", mid, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// parseAppLogPartArchiveUnix 解析 igozhang.part.{unix}.log。
func parseAppLogPartArchiveUnix(name string) (int64, bool) {
	prefix := appLogArchiveStem + ".part."
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".log") {
		return 0, false
	}
	mid := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".log")
	if mid == "" {
		return 0, false
	}
	for _, r := range mid {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	u, err := strconv.ParseInt(mid, 10, 64)
	if err != nil || u <= 0 {
		return 0, false
	}
	return u, true
}

// appLogArchiveRole 运行日志文件名 → 看板角色。
func appLogArchiveRole(name string) string {
	if name == appLogFileName {
		return "运行日志（当前）"
	}
	if name == logLastDayMarker {
		return ""
	}
	if _, ok := parseAppLogDailyArchiveDay(name); ok {
		return "运行日志日归档"
	}
	if _, ok := parseAppLogPartArchiveUnix(name); ok {
		return "运行日志过大切割"
	}
	return "运行日志归档"
}

// isSiteAppLogEntryName 判断 secret-log 目录下是否属于运行日志清单（忽略杂项文件）。
func isSiteAppLogEntryName(name string) bool {
	if name == appLogFileName || name == logLastDayMarker {
		return true
	}
	if strings.HasPrefix(name, "igozhang.") && strings.HasSuffix(name, ".log") {
		return true
	}
	return name == "igo-doc.log"
}

// logFileMaxAge：归档日志保留时长。
var logFileMaxAge = 15 * 24 * time.Hour

// logRotateSizeBytes：当前 igozhang.log 超过此大小则轮转为新文件。
var logRotateSizeBytes int64 = 500 * 1024 // 500 KiB

// logTruncateSizeBytes：当前 igozhang.log 超过此大小则截断为最近若干行并清空全部归档。
var logTruncateSizeBytes int64 = 1024 * 1024 // 1 MiB

const logTruncateKeepLines = 100

var (
	logRotateMu   sync.Mutex
	activeLogFile *os.File
)

// initAppLogging 将标准日志同时写入 stderr 与 DATA_ROOT/secret-log/igozhang.log。目录创建失败时仅保留 stderr。
func initAppLogging(dataRoot string) (close func() error) {
	close = func() error { return nil }
	dir := filepath.Join(dataRoot, logRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("logging: mkdir %s: %v (stderr only)", dir, err)
		return close
	}
	p := filepath.Join(dir, appLogFileName)
	legacy := filepath.Join(dir, "igo-doc.log")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if _, err2 := os.Stat(legacy); err2 == nil {
			if err3 := os.Rename(legacy, p); err3 == nil {
				log.Printf("logging: migrated %s -> %s", legacy, p)
			}
		}
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("logging: open %s: %v (stderr only)", p, err)
		return close
	}
	activeLogFile = f
	log.SetOutput(io.MultiWriter(os.Stderr, f))
	return func() error {
		logRotateMu.Lock()
		defer logRotateMu.Unlock()
		log.SetOutput(os.Stderr)
		var err error
		if activeLogFile != nil {
			err = activeLogFile.Close()
			activeLogFile = nil
		}
		return err
	}
}

func reopenLogAfterRotate(dataRoot string) error {
	dir := filepath.Join(dataRoot, logRelDir)
	p := filepath.Join(dir, appLogFileName)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.SetOutput(os.Stderr)
		return err
	}
	activeLogFile = f
	log.SetOutput(io.MultiWriter(os.Stderr, f))
	return nil
}

func uniqueArchivePath(dir, base string) string {
	p := filepath.Join(dir, base)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return p
	}
	if len(base) > 4 && base[len(base)-4:] == ".log" {
		return filepath.Join(dir, fmt.Sprintf("%s.%d.log", base[:len(base)-4], time.Now().Unix()))
	}
	return filepath.Join(dir, fmt.Sprintf("%s.%d", base, time.Now().Unix()))
}

// rotateActiveLogFileTo 关闭当前 igozhang.log，重命名为 arcBase 并新建空 igozhang.log。
func rotateActiveLogFileTo(dataRoot, arcBase string) error {
	logRotateMu.Lock()
	defer logRotateMu.Unlock()
	dir := filepath.Join(dataRoot, logRelDir)
	p := filepath.Join(dir, appLogFileName)
	log.SetOutput(os.Stderr)
	if activeLogFile != nil {
		_ = activeLogFile.Close()
		activeLogFile = nil
	}
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return reopenLogAfterRotate(dataRoot)
		}
		return err
	}
	arc := uniqueArchivePath(dir, arcBase)
	if err := os.Rename(p, arc); err != nil {
		if err := os.Truncate(p, 0); err != nil {
			return err
		}
	} else {
		log.Printf("maintenance: log archived path=%s", arc)
	}
	return reopenLogAfterRotate(dataRoot)
}

func tailLines(data []byte, n int) []byte {
	if n <= 0 || len(data) == 0 {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) <= n {
		return data
	}
	out := strings.Join(lines[len(lines)-n:], "\n")
	if len(data) > 0 && data[len(data)-1] == '\n' {
		out += "\n"
	}
	return []byte(out)
}

func resetActiveLogForTest() {
	logRotateMu.Lock()
	defer logRotateMu.Unlock()
	log.SetOutput(os.Stderr)
	if activeLogFile != nil {
		_ = activeLogFile.Close()
		activeLogFile = nil
	}
}

func truncateActiveLogToLastLines(dataRoot string, keep int) error {
	logRotateMu.Lock()
	defer logRotateMu.Unlock()
	dir := filepath.Join(dataRoot, logRelDir)
	p := filepath.Join(dir, appLogFileName)
	log.SetOutput(os.Stderr)
	if activeLogFile != nil {
		_ = activeLogFile.Close()
		activeLogFile = nil
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return reopenLogAfterRotate(dataRoot)
		}
		return err
	}
	kept := tailLines(raw, keep)
	if err := os.WriteFile(p, kept, 0o644); err != nil {
		return err
	}
	log.Printf("maintenance: log truncated keep_lines=%d path=%s", keep, p)
	return nil
}

func removeLogArchives(dataRoot string) {
	dir := filepath.Join(dataRoot, logRelDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == appLogFileName || name == logLastDayMarker {
			continue
		}
		path := filepath.Join(dir, name)
		if err := os.Remove(path); err != nil {
			log.Printf("maintenance: secret-log remove %s: %v", path, err)
		} else {
			log.Printf("maintenance: delete reason=log_truncate_clear path=%s", path)
		}
	}
}

func readLogLastDay(dataRoot string) string {
	raw, err := os.ReadFile(filepath.Join(dataRoot, logRelDir, logLastDayMarker))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func writeLogLastDay(dataRoot string, day string) {
	_ = os.WriteFile(filepath.Join(dataRoot, logRelDir, logLastDayMarker), []byte(day), 0o644)
}

func logTodayLocal() string {
	return time.Now().Format("2006-01-02")
}

func purgeExpiredLogArchives(dir string, cutoff time.Time, skip map[string]bool) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if skip[base] {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				log.Printf("maintenance: %s remove %s: %v", filepath.Base(dir), path, err)
			} else {
				log.Printf("maintenance: delete reason=log_ttl_age path=%s", path)
			}
		}
		return nil
	})
}

func readRotatingLogLastDay(dir, markerName string) string {
	raw, err := os.ReadFile(filepath.Join(dir, markerName))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func writeRotatingLogLastDay(dir, markerName, day string) {
	_ = os.WriteFile(filepath.Join(dir, markerName), []byte(day), 0o644)
}

func removeRotatingLogArchives(dir string, skip map[string]bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if skip[name] {
			continue
		}
		path := filepath.Join(dir, name)
		if err := os.Remove(path); err != nil {
			log.Printf("maintenance: %s remove %s: %v", filepath.Base(dir), path, err)
		} else {
			log.Printf("maintenance: delete reason=log_truncate_clear path=%s", path)
		}
	}
}

func truncateRotatingLogFile(path string, keep int) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	kept := tailLines(raw, keep)
	if err := os.WriteFile(path, kept, 0o644); err != nil {
		return err
	}
	log.Printf("maintenance: log truncated keep_lines=%d path=%s", keep, path)
	return nil
}

func rotateRotatingLogFile(dir, activeName, arcBase string) error {
	p := filepath.Join(dir, activeName)
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return err
			}
			return f.Close()
		}
		return err
	}
	arc := uniqueArchivePath(dir, arcBase)
	if err := os.Rename(p, arc); err != nil {
		if err := os.Truncate(p, 0); err != nil {
			return err
		}
	} else {
		log.Printf("maintenance: log archived path=%s", arc)
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

// maintainRotatingLogDir 通用日志轮转：按 archiveMaxAge 清理归档；按日/体积轮转；超大截断并清空归档。
func maintainRotatingLogDir(dir, activeName, markerName, archiveStem string, archiveMaxAge time.Duration) {
	if archiveMaxAge <= 0 {
		archiveMaxAge = logFileMaxAge
	}
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return
	}
	skip := map[string]bool{activeName: true, markerName: true}
	purgeExpiredLogArchives(dir, time.Now().Add(-archiveMaxAge), skip)

	fp := filepath.Join(dir, activeName)
	fi, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			writeRotatingLogLastDay(dir, markerName, logTodayLocal())
		}
		return
	}

	if fi.Size() > logTruncateSizeBytes {
		if err := truncateRotatingLogFile(fp, logTruncateKeepLines); err != nil {
			log.Printf("maintenance: %s truncate: %v", filepath.Base(dir), err)
			return
		}
		removeRotatingLogArchives(dir, skip)
		return
	}

	today := logTodayLocal()
	lastDay := readRotatingLogLastDay(dir, markerName)
	if lastDay == "" {
		writeRotatingLogLastDay(dir, markerName, today)
		lastDay = today
	}
	if lastDay != today {
		if fi.Size() > 0 {
			arc := fmt.Sprintf("%s.%s.log", archiveStem, lastDay)
			if err := rotateRotatingLogFile(dir, activeName, arc); err != nil {
				log.Printf("maintenance: %s daily rotate: %v", filepath.Base(dir), err)
			}
		}
		writeRotatingLogLastDay(dir, markerName, today)
		return
	}

	if fi.Size() > logRotateSizeBytes {
		arc := fmt.Sprintf("%s.part.%d.log", archiveStem, time.Now().Unix())
		if err := rotateRotatingLogFile(dir, activeName, arc); err != nil {
			log.Printf("maintenance: %s size rotate: %v", filepath.Base(dir), err)
		}
	}
}

// maintainSecretLogDir 由 runMaintenance 约每小时调用：保留 15 天归档；按日/体积轮转；超大单文件截断并清空归档。
func maintainSecretLogDir(dataRoot string) {
	dir := filepath.Join(dataRoot, logRelDir)
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return
	}
	skip := map[string]bool{appLogFileName: true, logLastDayMarker: true}
	purgeExpiredLogArchives(dir, time.Now().Add(-logFileMaxAge), skip)

	fp := filepath.Join(dir, appLogFileName)
	fi, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			writeLogLastDay(dataRoot, logTodayLocal())
		}
		return
	}

	if fi.Size() > logTruncateSizeBytes {
		if err := truncateActiveLogToLastLines(dataRoot, logTruncateKeepLines); err != nil {
			log.Printf("maintenance: secret-log truncate: %v", err)
			return
		}
		removeLogArchives(dataRoot)
		if err := reopenLogAfterRotate(dataRoot); err != nil {
			log.Printf("maintenance: secret-log reopen: %v", err)
		}
		return
	}

	today := logTodayLocal()
	lastDay := readLogLastDay(dataRoot)
	if lastDay == "" {
		writeLogLastDay(dataRoot, today)
		lastDay = today
	}
	if lastDay != today {
		if fi.Size() > 0 {
			arc := fmt.Sprintf("igozhang.%s.log", lastDay)
			if err := rotateActiveLogFileTo(dataRoot, arc); err != nil {
				log.Printf("maintenance: secret-log daily rotate: %v", err)
			}
		}
		writeLogLastDay(dataRoot, today)
		return
	}

	if fi.Size() > logRotateSizeBytes {
		arc := fmt.Sprintf("igozhang.part.%d.log", time.Now().Unix())
		if err := rotateActiveLogFileTo(dataRoot, arc); err != nil {
			log.Printf("maintenance: secret-log size rotate: %v", err)
		}
	}
}
