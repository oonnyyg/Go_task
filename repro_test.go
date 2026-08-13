package semver

import (
	"math"
	"strconv"
	"testing"
)

// TestIncOverflowBoundary verifies that incrementing a version component
// which is already at the maximum uint64 value panics instead of silently
// overflowing back to 0, while the existing prerelease handling is preserved.
func TestIncOverflowBoundary(t *testing.T) {
	max := uint64(math.MaxUint64)

	// IncPatch on a maxed-out patch (no prerelease) must panic.
	v := New(1, 2, max, "", "")
	assertPanics(t, "IncPatch max patch", func() { v.IncPatch() })

	// The same boundary is reached when the max value comes from parsing.
	vs, err := NewVersion("1.2." + strconv.FormatUint(max, 10))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	assertPanics(t, "IncPatch parsed max patch", func() { vs.IncPatch() })

	// IncPatch with a prerelease does not increment patch, so it must NOT
	// panic and must preserve the (max) patch value while clearing pre/meta.
	vp := New(1, 2, max, "beta", "meta")
	got := vp.IncPatch()
	if got.Patch() != max {
		t.Fatalf("IncPatch with prerelease: expected patch %d, got %d", max, got.Patch())
	}
	if got.Prerelease() != "" || got.Metadata() != "" {
		t.Fatalf("IncPatch with prerelease: expected pre/meta cleared, got %q/%q", got.Prerelease(), got.Metadata())
	}

	// IncMinor on a maxed-out minor must panic.
	vm := New(1, max, 3, "", "")
	assertPanics(t, "IncMinor max minor", func() { vm.IncMinor() })

	// IncMajor on a maxed-out major must panic.
	vM := New(max, 2, 3, "", "")
	assertPanics(t, "IncMajor max major", func() { vM.IncMajor() })
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("%s: expected panic, got none", name)
		}
	}()
	fn()
}
