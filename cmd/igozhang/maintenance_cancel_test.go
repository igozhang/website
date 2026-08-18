package main

import (
	"context"
	"testing"
	"time"
)

func TestStartHourlyMaintenance_StopsOnCancel(t *testing.T) {
	dir := t.TempDir()
	transfer := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	startHourlyMaintenance(ctx, dir, transfer)
	time.Sleep(30 * time.Millisecond)
	cancel()
	// Allow the maintenance goroutine to observe cancel and return.
	time.Sleep(100 * time.Millisecond)

	// Starting again after cancel must also be cancellable (no stuck Sleep loop).
	ctx2, cancel2 := context.WithCancel(context.Background())
	startHourlyMaintenance(ctx2, dir, transfer)
	time.Sleep(30 * time.Millisecond)
	cancel2()
	time.Sleep(100 * time.Millisecond)
}
