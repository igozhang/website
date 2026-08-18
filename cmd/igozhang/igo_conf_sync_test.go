package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// 内置 igo.conf（go:embed）须与 deploy/secret07/igo.conf 一致；改 cmd/igozhang/igo.conf 后执行 go generate ./cmd/igozhang/...
func TestIgoConfEmbeddedMatchesDeploy(t *testing.T) {
	deployPath := filepath.Join("..", "..", "deploy", "secret07", "igo.conf")
	deploy, err := os.ReadFile(deployPath)
	if err != nil {
		t.Fatalf("read deploy igo.conf: %v (run: go generate ./cmd/igozhang/...)", err)
	}
	if !bytes.Equal(opsConfigExample, deploy) {
		t.Fatalf("igo.conf out of sync: edit cmd/igozhang/igo.conf then run go generate ./cmd/igozhang/...")
	}
}
