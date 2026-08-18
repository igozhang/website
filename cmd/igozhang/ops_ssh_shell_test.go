package main

import "testing"

func TestOpsSSHShellCommand(t *testing.T) {
	got := opsSSHShellCommand("uname -a")
	want := "bash -lc 'uname -a'"
	if got != want {
		t.Fatalf("opsSSHShellCommand simple = %q; want %q", got, want)
	}
	got = opsSSHShellCommand("echo it's fine")
	want = "bash -lc 'echo it'\\''s fine'"
	if got != want {
		t.Fatalf("opsSSHShellCommand quote = %q; want %q", got, want)
	}
}
