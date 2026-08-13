package cron

import (
	"testing"
	"time"
)

// TestReproSteppedStarOr reproduces the bug where a stepped star (e.g. */10)
// in day-of-month combined with a restricted day-of-week (e.g. Sun) is treated
// as an AND instead of an OR. Both fields are restricted, so standard cron
// semantics require firing when EITHER matches.
func TestReproSteppedStarOr(t *testing.T) {
	sched, err := ParseStandard("* * */10 * Sun")
	if err != nil {
		t.Fatal(err)
	}
	// Sun Jul 15 00:00 2012 is a Sunday; day 15 is NOT in */10 (1,11,21,31).
	// Because day-of-week matches, the schedule should still fire here.
	after := getTime("Sun Jul 15 00:00 2012").Add(-1 * time.Second)
	got := sched.Next(after)
	want := getTime("Sun Jul 15 00:00 2012")
	if !got.Equal(want) {
		t.Errorf("stepped-star OR: expected %v, got %v", want, got)
	}
}
