package main

import (
	"log"
	"path/filepath"
	"strings"
)

const (
	commonCommandsFileName = "common-commands.md"
	commonCommandsMaxLines = 20
)

func commonCommandsPath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), commonCommandsFileName)
}

// readCommonCommandsExportBytes 供文章 ZIP 导出：与 listCommonCommands 同源判定，
// 仅当文件存在且含至少一条生效命令（非空、非 # 注释），且与 secret07CommonCommandsTemplate 不等时返回全文。
func readCommonCommandsExportBytes(dataRoot string) []byte {
	if dataRoot == "" {
		return nil
	}
	raw, ok := secret07ReadFileOrMissing(commonCommandsPath(dataRoot), []byte(secret07CommonCommandsTemplate))
	if !ok {
		return nil
	}
	return raw
}

// listCommonCommands 读取 secret07/common-commands.md，最多 commonCommandsMaxLines 条。
// 仅说明模板时等同未配置（无快捷按钮，与内嵌无关）。
func listCommonCommands(dataRoot string) []string {
	raw, ok := secret07ReadFileOrMissing(commonCommandsPath(dataRoot), []byte(secret07CommonCommandsTemplate))
	if !ok {
		return []string{}
	}
	return parseCommonCommandsLines(raw)
}

func parseCommonCommandsLines(raw []byte) []string {
	out := make([]string, 0)
	truncated := false
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(out) >= commonCommandsMaxLines {
			truncated = true
			break
		}
		out = append(out, line)
	}
	if truncated {
		log.Printf("ops-common-cmds: truncated to %d lines", commonCommandsMaxLines)
	}
	return out
}
