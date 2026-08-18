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

// initSecret07Files 在 DATA_ROOT/secret07 下补齐有效配置与说明文件（已存在则跳过）。
// igo.conf：写入内嵌完整配置；其余文件：格式说明 + 以 # 开头的不生效示例（common-commands 无生效行）。
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
	writeSecret07IfMissing(filepath.Join(dir, linkConfFileName), []byte(secret07LinkConfTemplate), 0o644, linkConfFileName)
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

const secret07SSHPassTemplate = `# 运维 SSH 目标：每行「主机 端口 用户名 密码」
# IPv4 末段可用 a-b 展开（如 10.1.1.2-5），展开后合计最多 10 台
# 以 # 开头的行不生效
# 示例：192.168.1.100 22 root your-password
# 示例：10.1.1.2-5 22 root your-password
`

const secret07CommonCommandsTemplate = `# 运维页「常用命令」快捷按钮：每行一条 shell 命令（最多 20 条）
# 以 # 开头的行不生效；仅注释时页面不显示快捷按钮
# 示例：uname -a
`

