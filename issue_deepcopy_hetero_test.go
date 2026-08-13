package mergo_test

import (
	"reflect"
	"testing"

	"github.com/imdario/mergo"
)

// When a []interface{} slice already holds a map value and the source slice
// holds a non-map value at the same index, WithSliceDeepCopy must not panic.
// The destination map content is preserved because the heterogeneous source
// value cannot be assigned to the map element.
func TestIssueSliceDeepCopyHeterogeneousMap(t *testing.T) {
	dst := []interface{}{map[string]int{"a": 1}}
	src := []interface{}{"hello"}

	if err := mergo.Merge(&dst, src, mergo.WithSliceDeepCopy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []interface{}{map[string]int{"a": 1}}
	if !reflect.DeepEqual(dst, expected) {
		t.Errorf("dst = %v, want %v", dst, expected)
	}
}
