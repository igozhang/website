package main

import (
	"os"
	"testing"
	"time"
)

func readPlainIgoKeyFile(t *testing.T, dataRoot string) string {
	t.Helper()
	raw, err := os.ReadFile(igoKeyPlainPath(dataRoot))
	if err != nil {
		t.Fatalf("read secret07/igokey: %v", err)
	}
	return string(raw)
}

func TestObtainIgoKey_PersistReuseAcrossReload(t *testing.T) {
	dir := t.TempDir()
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	igoKeyMu.Unlock()

	k1, _ := obtainIgoKey(dir, false)
	if got := readPlainIgoKeyFile(t, dir); got != k1 {
		t.Fatalf("secret07/igokey after issue=%q want %q", got, k1)
	}
	if err := os.Remove(igoKeyPlainPath(dir)); err != nil {
		t.Fatal(err)
	}
	igoKeyMu.Lock()
	igoKeyActive.key = ""
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyMu.Unlock()

	k2, _ := obtainIgoKey(dir, false)
	if k1 != k2 {
		t.Fatalf("want reuse from disk: %q vs %q", k1, k2)
	}
	if got := readPlainIgoKeyFile(t, dir); got != k2 {
		t.Fatalf("secret07/igokey=%q want %q", got, k2)
	}
}

func TestObtainIgoKey_NoReuseWhenRemainingLE15m(t *testing.T) {
	dir := t.TempDir()
	igoKeyMu.Lock()
	igoKeyStore = map[string]igoKeyEntry{}
	igoKeyActive.key = ""
	igoKeyMu.Unlock()

	k1, _ := obtainIgoKey(dir, false)
	igoKeyMu.Lock()
	igoKeyStore[k1] = igoKeyEntry{expiresAt: time.Now().Add(1 * time.Minute)}
	igoKeyActive.key = k1
	igoKeyActive.expiresAt = igoKeyStore[k1].expiresAt
	persistIgoKeyActiveLocked(dir, k1, igoKeyStore[k1].expiresAt)
	igoKeyMu.Unlock()

	k2, _ := obtainIgoKey(dir, false)
	if k2 == k1 {
		t.Fatal("expected new key when remaining <= 15m")
	}
	if got := readPlainIgoKeyFile(t, dir); got != k2 {
		t.Fatalf("secret07/igokey=%q want %q", got, k2)
	}
}
