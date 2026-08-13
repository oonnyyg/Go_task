// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package expirable

import (
	"testing"
	"time"
)

// TestExpiredEntryReadReturnsMiss reproduces the inconsistency where an entry
// that has expired but has not yet been swept by the background cleanup is
// returned by Get/Peek with ok=true and the zero value instead of a clean miss.
func TestExpiredEntryReadReturnsMiss(t *testing.T) {
	lc := NewLRU[string, string](0, nil, time.Second)

	lc.Add("key1", "val1")

	// Force the stored entry into the "expired but not yet cleaned up" window:
	// its ExpiresAt is in the past while the entry is still physically present.
	lc.mu.Lock()
	ent, ok := lc.items["key1"]
	if !ok {
		lc.mu.Unlock()
		t.Fatalf("entry missing before expiration")
	}
	ent.ExpiresAt = time.Now().Add(-time.Second)
	if lc.evictList.Length() != 1 {
		lc.mu.Unlock()
		t.Fatalf("expected entry to still be present, got len %d", lc.evictList.Length())
	}
	lc.mu.Unlock()

	if v, ok := lc.Get("key1"); ok {
		t.Fatalf("Get on expired entry returned ok=true, value=%q; want ok=false", v)
	}
	if v, ok := lc.Peek("key1"); ok {
		t.Fatalf("Peek on expired entry returned ok=true, value=%q; want ok=false", v)
	}
}
