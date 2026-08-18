package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	sshPassMaxLines    = 10
	sshPassMaxTargets  = 10
	sshPassMaxIPRange  = 32
)

// sshTarget 一条可连接的 SSH 目标（展开 IP 段后计为一条）。
type sshTarget struct {
	ID       int
	Host     string
	Port     int
	User     string
	Password string
}

func (t sshTarget) poolKey() string {
	return fmt.Sprintf("%s:%d@%s", t.Host, t.Port, t.User)
}

func (t sshTarget) label() string {
	return fmt.Sprintf("%s:%d (%s)", t.Host, t.Port, t.User)
}

// listSSHTargets 解析 secret07/ssh.pass，最多 10 行、展开后最多 10 个目标。
// 仅说明模板或无有效条目时等同未配置（返回 ErrNotExist）。
func listSSHTargets(dataRoot string) ([]sshTarget, error) {
	raw, ok := secret07ReadFileOrMissing(sshPassPath(dataRoot), []byte(secret07SSHPassTemplate))
	if !ok {
		return nil, os.ErrNotExist
	}
	var out []sshTarget
	lines := 0
	for _, line := range strings.Split(string(raw), "\n") {
		if len(out) >= sshPassMaxTargets {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines++
		if lines > sshPassMaxLines {
			break
		}
		parts := strings.SplitN(line, " ", 4)
		if len(parts) < 4 {
			continue
		}
		port, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || port < 1 || port > 65535 {
			continue
		}
		hostPat := strings.TrimSpace(parts[0])
		user := strings.TrimSpace(parts[2])
		pass := parts[3]
		if hostPat == "" || user == "" || pass == "" {
			continue
		}
		hosts, err := expandHostPattern(hostPat)
		if err != nil {
			continue
		}
		for _, host := range hosts {
			if len(out) >= sshPassMaxTargets {
				break
			}
			out = append(out, sshTarget{
				ID:       len(out),
				Host:     host,
				Port:     port,
				User:     user,
				Password: pass,
			})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid entry in %s", sshPassFileName)
	}
	return out, nil
}

// expandHostPattern 仅当主机为 IPv4 且末段为 a-b 时展开（如 10.1.1.2-5）；否则原样返回。
func expandHostPattern(host string) ([]string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	lastDash := strings.LastIndex(host, "-")
	if lastDash < 0 {
		return []string{host}, nil
	}
	left := host[:lastDash]
	endPart := host[lastDash+1:]
	dot := strings.LastIndex(left, ".")
	if dot < 0 {
		return []string{host}, nil
	}
	prefix := left[:dot+1]
	startStr := left[dot+1:]
	if !ipv4OctetOK(startStr) || !ipv4OctetOK(endPart) {
		return []string{host}, nil
	}
	start, err1 := strconv.Atoi(startStr)
	end, err2 := strconv.Atoi(endPart)
	if err1 != nil || err2 != nil || start > end || start < 0 || end > 255 {
		return []string{host}, nil
	}
	n := end - start + 1
	if n > sshPassMaxIPRange {
		return nil, fmt.Errorf("host range %s too large (max %d hosts)", host, sshPassMaxIPRange)
	}
	out := make([]string, 0, n)
	for i := start; i <= end; i++ {
		out = append(out, fmt.Sprintf("%s%d", prefix, i))
	}
	return out, nil
}

func ipv4OctetOK(s string) bool {
	if s == "" || len(s) > 3 {
		return false
	}
	n, err := strconv.Atoi(s)
	return err == nil && n >= 0 && n <= 255
}

func sshTargetByID(dataRoot string, id int) (*sshTarget, error) {
	all, err := listSSHTargets(dataRoot)
	if err != nil {
		return nil, err
	}
	if id < 0 || id >= len(all) {
		return nil, fmt.Errorf("invalid target id %d", id)
	}
	t := all[id]
	return &t, nil
}
