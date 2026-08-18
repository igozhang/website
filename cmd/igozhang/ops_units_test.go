package main

import (
	"testing"
	"time"
)

func TestParseByteSize(t *testing.T) {
	cases := []struct {
		in  string
		out int64
	}{
		{"200K", 200 * 1024},
		{"9G", 9 * 1024 * 1024 * 1024},
		{"100M", 100 * 1024 * 1024},
		{"512K", 512 * 1024},
		{"4M", 4 * 1024 * 1024},
	}
	for _, c := range cases {
		got, err := parseByteSize(c.in)
		if err != nil || got != c.out {
			t.Fatalf("parseByteSize(%q) = %d, %v; want %d", c.in, got, err, c.out)
		}
	}
}

func TestParseBeijingOffsetHours(t *testing.T) {
	cases := []struct {
		in  string
		out int
		err bool
	}{
		{"", 0, false},
		{"0", 0, false},
		{"-1", -1, false},
		{"+2", 2, false},
		{"12", 12, false},
		{"-12", -12, false},
		{"13", 0, true},
		{"-13", 0, true},
		{"x", 0, true},
	}
	for _, c := range cases {
		got, err := parseBeijingOffsetHours(c.in)
		if c.err {
			if err == nil {
				t.Fatalf("parseBeijingOffsetHours(%q) want error", c.in)
			}
			continue
		}
		if err != nil || got != c.out {
			t.Fatalf("parseBeijingOffsetHours(%q) = %d, %v; want %d", c.in, got, err, c.out)
		}
	}
}

func TestParseOnOff(t *testing.T) {
	for _, c := range []struct{ in string; on bool }{
		{"on", true}, {"OFF", false}, {"yes", true}, {"0", false},
	} {
		got, err := parseOnOff(c.in)
		if err != nil || got != c.on {
			t.Fatalf("parseOnOff(%q)=%v %v want %v", c.in, got, err, c.on)
		}
	}
}

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in  string
		out time.Duration
	}{
		{"10s", 10 * time.Second},
		{"100ms", 100 * time.Millisecond},
		{"1m", time.Minute},
		{"1h", time.Hour},
		{"12h", 12 * time.Hour},
		{"30", 30 * time.Second},
	}
	for _, c := range cases {
		got, err := parseDuration(c.in)
		if err != nil || got != c.out {
			t.Fatalf("parseDuration(%q) = %v, %v; want %v", c.in, got, err, c.out)
		}
	}
}
