package cron

import (
	"testing"
	"time"
)

// TestChainSkipIfStillRunningIsolatesJobs reproduces the isolation bug where
// two different jobs wrapped by the same SkipIfStillRunning chain share a
// single token. When one job is still running, an independent job may be
// skipped even though only re-entrancy of the *same* job should be skipped.
func TestChainSkipIfStillRunningIsolatesJobs(t *testing.T) {
	var jobA countJob
	jobA.delay = 20 * time.Millisecond
	var jobB countJob

	chain := NewChain(SkipIfStillRunning(DiscardLogger))
	wrappedA := chain.Then(&jobA)
	wrappedB := chain.Then(&jobB)

	// Start jobA; it will be running for ~20ms and hold its skip-token.
	go wrappedA.Run()
	time.Sleep(2 * time.Millisecond)

	// jobB is an independent job. It must be allowed to run even while jobA
	// is still running.
	go wrappedB.Run()
	time.Sleep(5 * time.Millisecond)

	if b := jobB.Done(); b != 1 {
		t.Errorf("expected independent job B to run while job A is running, got %d", b)
	}

	// Let jobA finish.
	time.Sleep(30 * time.Millisecond)
	if a := jobA.Done(); a != 1 {
		t.Errorf("expected job A to complete, got %d", a)
	}
}
