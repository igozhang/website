// 将 cmd/igozhang/igo.conf（go:embed 内置配置唯一源码）同步到 deploy/secret07/igo.conf。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "sync_igo_conf: runtime.Caller failed")
		os.Exit(1)
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	src := filepath.Join(root, "cmd", "igozhang", "igo.conf")
	dst := filepath.Join(root, "deploy", "secret07", "igo.conf")

	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sync_igo_conf: read %s: %v\n", src, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sync_igo_conf: mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sync_igo_conf: write %s: %v\n", dst, err)
		os.Exit(1)
	}
	fmt.Printf("sync_igo_conf: %s -> %s\n", src, dst)
}
