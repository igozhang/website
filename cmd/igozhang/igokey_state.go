package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const igoKeyActiveFile = "igokey-active.json"

type igoKeyDiskState struct {
	Key       string `json:"key"`
	ExpiresAt int64  `json:"expiresAt"`
}

func igoKeyActivePath(dataRoot string) string {
	return filepath.Join(secret07Dir(dataRoot), igoKeyActiveFile)
}

// syncIgoKeyActiveFromDiskLocked 从 secret07 恢复活跃密钥（调用方已持 igoKeyMu）。
func syncIgoKeyActiveFromDiskLocked(dataRoot string, now time.Time) {
	if dataRoot == "" {
		return
	}
	raw, err := os.ReadFile(igoKeyActivePath(dataRoot))
	if err != nil {
		return
	}
	var st igoKeyDiskState
	if err := json.Unmarshal(raw, &st); err != nil || st.Key == "" || st.ExpiresAt <= 0 {
		return
	}
	exp := time.Unix(st.ExpiresAt, 0)
	if !exp.After(now) {
		_ = os.Remove(igoKeyActivePath(dataRoot))
		if igoKeyActive.key == st.Key {
			igoKeyActive.key = ""
			igoKeyActive.expiresAt = time.Time{}
		}
		delete(igoKeyStore, st.Key)
		return
	}
	igoKeyStore[st.Key] = igoKeyEntry{expiresAt: exp}
	igoKeyActive.key = st.Key
	igoKeyActive.expiresAt = exp
}

func persistIgoKeyActiveLocked(dataRoot, key string, expiresAt time.Time) {
	if dataRoot == "" {
		return
	}
	path := igoKeyActivePath(dataRoot)
	if key == "" {
		_ = os.Remove(path)
		return
	}
	_ = os.MkdirAll(secret07Dir(dataRoot), 0o755)
	st := igoKeyDiskState{Key: key, ExpiresAt: expiresAt.Unix()}
	b, err := json.Marshal(st)
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return
	}
	// Windows 上 Rename 不能覆盖已存在目标文件；先删再替换。
	_ = os.Remove(path)
	if err := os.Rename(tmp, path); err != nil {
		_ = os.WriteFile(path, b, 0o600)
		_ = os.Remove(tmp)
	}
}
