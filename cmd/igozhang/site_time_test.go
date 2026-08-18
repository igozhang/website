package main

import (
	"testing"
	"time"
)

func TestFormatBeijingUnix(t *testing.T) {
	opsMu.Lock()
	opsCurrent.beijingOffsetHours = 0
	opsMu.Unlock()

	utc := time.Date(2026, 5, 25, 10, 30, 45, 0, time.UTC)
	got := formatBeijingUnix(utc.Unix())
	want := "2026-05-25 18:30:45"
	if got != want {
		t.Fatalf("formatBeijingUnix() = %q, want %q", got, want)
	}

	opsMu.Lock()
	opsCurrent.beijingOffsetHours = -1
	opsMu.Unlock()
	got = formatBeijingUnix(utc.Unix())
	want = "2026-05-25 17:30:45"
	if got != want {
		t.Fatalf("offset -1: got %q want %q", got, want)
	}
}

func TestFormatBeijingCompact(t *testing.T) {
	opsMu.Lock()
	opsCurrent.beijingOffsetHours = 0
	opsMu.Unlock()
	utc := time.Date(2026, 5, 25, 10, 30, 45, 0, time.UTC)
	got := formatBeijingCompact(utc.Unix())
	if got != "20260525-1830" {
		t.Fatalf("formatBeijingCompact() = %q, want 20260525-1830", got)
	}
}
