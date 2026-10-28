package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
)

//go:embed pic.readme.md
var secret07PicReadmeTemplate []byte

const secret07PicReadmeName = "pic.readme.md"

// initSecret07Files 在 DATA_ROOT/secret07 下补齐配置文件（缺失或不含生效行时写入默认）。
// 约定：生效配置在文件头部，必要说明与案例在尾部（均为 # 注释）。
// igo.conf / link.conf：默认含生效行；ssh.pass、common-commands、黑白名单等默认头部留空。
func initSecret07Files(dataRoot string) {
	if dataRoot == "" {
		return
	}
	dir := secret07Dir(dataRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("secret07: mkdir %s: %v", dir, err)
		return
	}
	ensureOpsConfigFile(dataRoot)
	writeSecret07IfMissing(filepath.Join(dir, sshPassFileName), []byte(secret07SSHPassTemplate), 0o600, "ssh.pass")
	writeSecret07IfMissing(filepath.Join(dir, commonCommandsFileName), []byte(secret07CommonCommandsTemplate), 0o644, "common-commands.md")
	writeSecret07IfMissing(filepath.Join(dir, secret07PicReadmeName), secret07PicReadmeTemplate, 0o644, secret07PicReadmeName)
	ensureSecret07Config(filepath.Join(dir, linkConfFileName), []byte(secret07LinkConfTemplate), 0o644, linkConfFileName)
	ensureIPBlacklistFile(dataRoot)
	ensureIPWhitelistFile(dataRoot)
}

func writeSecret07IfMissing(path string, content []byte, perm os.FileMode, tag string) bool {
	if _, err := os.Stat(path); err == nil {
		return false
	}
	if err := os.WriteFile(path, content, perm); err != nil {
		log.Printf("secret07: write %s: %v", tag, err)
		return false
	}
	log.Printf("secret07: created %s (edit on server as needed)", tag)
	return true
}

// ensureSecret07Config 在文件缺失或无任何生效行时写入默认配置（头部生效、尾部说明）。
func ensureSecret07Config(path string, content []byte, perm os.FileMode, tag string) bool {
	if raw, err := os.ReadFile(path); err == nil && secret07HasEffectiveLines(raw) {
		return false
	}
	if err := os.WriteFile(path, content, perm); err != nil {
		log.Printf("secret07: write %s: %v", tag, err)
		return false
	}
	log.Printf("secret07: created %s (edit on server as needed)", tag)
	return true
}

const secret07SSHPassTemplate = `# === 生效配置（非 # 行生效；默认留空）===

# === 说明与案例（均为注释，不生效；生效配置写在上方）===
# 运维 SSH 目标：每行「主机 端口 用户名 密码」
# IPv4 末段可用 a-b 展开（如 10.1.1.2-5），展开后合计最多 10 台
# 以 # 开头的行不生效
# 示例：192.168.1.100 22 root your-password
# 示例：10.1.1.2-5 22 root your-password
`

const secret07CommonCommandsTemplate = `# === 生效配置（非 # 行生效；默认留空）===

# === 说明与案例（均为注释，不生效；生效配置写在上方）===
# 运维页「常用命令」快捷按钮：每行一条 shell 命令（最多 20 条）
# 以 # 开头的行不生效；仅注释时页面不显示快捷按钮
# 示例：uname -a
`

