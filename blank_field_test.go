package mergo_test

import (
	"testing"

	"github.com/imdario/mergo"
)

type withBlankField struct {
	_    struct{}
	Name string
}

type withBlankNonEmptyField struct {
	_    [4]byte
	Name string
}

func TestMergeBlankFieldWithOverride(t *testing.T) {
	dst := withBlankField{Name: "dst"}
	src := withBlankField{Name: "src"}
	if err := mergo.Merge(&dst, src, mergo.WithOverride); err != nil {
		t.Fatalf("merge returned error: %v", err)
	}
	if dst.Name != "src" {
		t.Errorf("expected Name %q, got %q", "src", dst.Name)
	}
}

func TestMergeBlankFieldWithoutOverride(t *testing.T) {
	dst := withBlankField{Name: "dst"}
	src := withBlankField{Name: "src"}
	if err := mergo.Merge(&dst, src); err != nil {
		t.Fatalf("merge returned error: %v", err)
	}
	// Without override the non-empty dst value is preserved.
	if dst.Name != "dst" {
		t.Errorf("expected Name %q, got %q", "dst", dst.Name)
	}
}

func TestMergeBlankNonEmptyFieldWithOverride(t *testing.T) {
	dst := withBlankNonEmptyField{Name: "dst"}
	src := withBlankNonEmptyField{Name: "src"}
	if err := mergo.Merge(&dst, src, mergo.WithOverride); err != nil {
		t.Fatalf("merge returned error: %v", err)
	}
	if dst.Name != "src" {
		t.Errorf("expected Name %q, got %q", "src", dst.Name)
	}
}
