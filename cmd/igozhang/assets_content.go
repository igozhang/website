package main

import (
	"log"
	"os"
	"path/filepath"
)

const assetsEngRelDir = "eng"

const assetsPhrasesTemplate = `# === 生效配置（非 # 行生效；默认留空；≥150 条有效行时启用此外部库）===

# === 说明与案例（均为注释，不生效；生效配置写在上方）===
# 全站页眉轮播引语；不足 150 条有效行时使用内嵌样本
# 格式：英文｜中文（全角竖线 U+FF5C 分隔，非半角 |）
# 以 # 开头的行不生效
# 示例：Hello world｜你好世界
`

const assetsWord3kTemplate = `# === 生效配置（非 # 行生效；默认留空；≥50 条有效行时启用此外部库）===

# === 说明与案例（均为注释，不生效；生效配置写在上方）===
# 知键词库；不足 50 条有效行时使用内嵌样本
# 格式：单词 /音标/ 释义（音标两侧为半角斜杠）
# 以 # 开头的行不生效
# 示例：hello /həˈloʊ/ 你好
`

func docAssetsEngDir(dataRoot string) string {
	return filepath.Join(docAssetsRoot(dataRoot), assetsEngRelDir)
}

func docAssetsPhrasesPath(dataRoot string) string {
	return filepath.Join(docAssetsEngDir(dataRoot), phrasesFileName)
}

func docAssetsWord3kPath(dataRoot string) string {
	return filepath.Join(docAssetsEngDir(dataRoot), word3kFileName)
}

// migrateAssetsEngFiles 将旧路径 assets/phrases.md、assets/word3k-full.md 迁入 assets/eng/。
func migrateAssetsEngFiles(dataRoot string) {
	if dataRoot == "" {
		return
	}
	assets := docAssetsRoot(dataRoot)
	eng := docAssetsEngDir(dataRoot)
	_ = os.MkdirAll(eng, 0o755)
	for _, name := range []string{phrasesFileName, word3kFileName} {
		old := filepath.Join(assets, name)
		new := filepath.Join(eng, name)
		if _, err := os.Stat(new); err == nil {
			continue
		}
		if _, err := os.Stat(old); err != nil {
			continue
		}
		if err := os.Rename(old, new); err != nil {
			log.Printf("assets: migrate %s -> eng/: %v", name, err)
		} else {
			log.Printf("assets: migrated %s to assets/eng/", name)
		}
	}
}

func docAssetsMp3Dir(dataRoot string) string {
	return filepath.Join(docAssetsRoot(dataRoot), "mp3")
}

// initAssetsContentFiles 在 DATA_ROOT/assets 下补齐 mp3 目录与引语/知键词库模板（已存在则跳过）。
func initAssetsContentFiles(dataRoot string) {
	if dataRoot == "" {
		return
	}
	assets := docAssetsRoot(dataRoot)
	if err := os.MkdirAll(filepath.Join(assets, "md"), 0o755); err != nil {
		log.Printf("assets: mkdir md: %v", err)
		return
	}
	if err := os.MkdirAll(docAssetsMp3Dir(dataRoot), 0o755); err != nil {
		log.Printf("assets: mkdir mp3: %v", err)
		return
	}
	migrateAssetsEngFiles(dataRoot)
	if err := os.MkdirAll(docAssetsEngDir(dataRoot), 0o755); err != nil {
		log.Printf("assets: mkdir eng: %v", err)
		return
	}
	writeAssetsIfMissing(docAssetsPhrasesPath(dataRoot), []byte(assetsPhrasesTemplate), 0o644, "assets/eng/phrases.md")
	writeAssetsIfMissing(docAssetsWord3kPath(dataRoot), []byte(assetsWord3kTemplate), 0o644, "assets/eng/word3k-full.md")
}

func writeAssetsIfMissing(path string, content []byte, perm os.FileMode, tag string) bool {
	if _, err := os.Stat(path); err == nil {
		return false
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("assets: mkdir %s: %v", tag, err)
		return false
	}
	if err := os.WriteFile(path, content, perm); err != nil {
		log.Printf("assets: write %s: %v", tag, err)
		return false
	}
	log.Printf("assets: created %s (edit on server as needed)", tag)
	return true
}

// postAssetsEngZipImport 导入 eng/ 后刷新页眉引语与知键词库内存缓存。
func postAssetsEngZipImport(dataRoot string) {
	if dataRoot == "" {
		return
	}
	reloadPhrasesCache()
	reloadWord3kCache()
}
