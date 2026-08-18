package main

import (
	"strings"
	"testing"
)

func TestRenderOpsHTML_siteHeadCSSInsideStyle(t *testing.T) {
	dir := t.TempDir()
	initSecret07Files(dir)
	html := string(renderOpsHTML("", dir, nil))
	if strings.Contains(html, "</style>\n.site-head{") {
		t.Fatal("SITE_HEAD_CSS must be inside <style>, not raw text after </style>")
	}
	if !strings.Contains(html, ".ops-cmd") || !strings.Contains(html, "id=\"opsCmd\"") {
		t.Fatal("ops command field missing")
	}
	if !strings.Contains(html, "id=\"opsCommonCmds\"") || !strings.Contains(html, "var opsCommonCmds=") {
		t.Fatal("ops common command chips missing")
	}
	if !strings.Contains(html, "var opsCommonCmds=") {
		t.Fatal("ops common commands JSON missing")
	}
	if strings.Contains(html, "var opsCommonCmds=null") {
		t.Fatal("ops common commands should be empty array not null")
	}
	if authStart := strings.Index(html, `__igoAuthInit`); authStart < 0 {
		t.Fatal("ops must use shared __igoAuthInit")
	}
	if strings.Contains(html, "btnOpsAuth") || strings.Contains(html, "id=\"opsAuth\"") {
		t.Fatal("ops must not keep legacy opsAuth UI")
	}
	if !strings.Contains(html, "IGO密钥认证") || !strings.Contains(html, `id="igoMainSec"`) {
		t.Fatal("ops must use unified IGO auth panel")
	}
	if strings.Contains(html, "if(!sshReady)") && strings.Contains(html, "__igoAuthInit") {
		// auth itself must not gate on ssh.pass; remote exec may still check sshReady
		authIdx := strings.Index(html, `__igoAuthInit`)
		authBlock := html[authIdx : authIdx+400]
		if strings.Contains(authBlock, "if(!sshReady)") {
			t.Fatal("ops auth must not require ssh.pass")
		}
	}
	if strings.Contains(html, "<style><style>") {
		t.Fatal("duplicate nested style tags")
	}
}
