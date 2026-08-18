package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestShouldSkipHTTPRequestLog(t *testing.T) {
	if !shouldSkipHTTPRequestLog("/health") {
		t.Fatal("/health should skip")
	}
	if !shouldSkipHTTPRequestLog("/favicon.ico") {
		t.Fatal("favicon should skip")
	}
	if shouldSkipHTTPRequestLog("/igokey") {
		t.Fatal("page should not skip")
	}
}

func TestFormatRequestLogDuration_ascii(t *testing.T) {
	if s := formatRequestLogDuration(13*time.Microsecond + 632*time.Nanosecond); s != "13.632us" {
		t.Fatalf("us=%q", s)
	}
	if s := formatRequestLogDuration(566*time.Microsecond + 276*time.Nanosecond); s != "566.276us" {
		t.Fatalf("us=%q", s)
	}
	if s := formatRequestLogDuration(28*time.Millisecond + 77*time.Microsecond); s != "28.077ms" {
		t.Fatalf("ms=%q", s)
	}
	if strings.Contains(formatRequestLogDuration(time.Microsecond), "µ") {
		t.Fatal("must not contain unicode micro sign")
	}
}

func TestLogRequests_skipsHealth(t *testing.T) {
	var lines []string
	orig := log.Writer()
	log.SetOutput(&captureLogWriter{append: func(b []byte) { lines = append(lines, string(b)) }})
	defer log.SetOutput(orig)

	h := logRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
	}
	req := httptest.NewRequest(http.MethodGet, "/igokey", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if len(lines) != 1 {
		t.Fatalf("lines=%d want 1 (health skipped): %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "GET /igokey") || !strings.Contains(lines[0], "us") {
		t.Fatalf("line=%q", lines[0])
	}
}

type captureLogWriter struct {
	append func([]byte)
}

func (w *captureLogWriter) Write(p []byte) (int, error) {
	if w.append != nil {
		w.append(p)
	}
	return len(p), nil
}
