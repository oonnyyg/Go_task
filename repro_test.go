package env

import "testing"

// Reproduces the edge case where a pointer-to-basic-type field is already
// non-nil when env.Parse runs and the corresponding environment variable is
// set. It should be parsed like a non-pointer field instead of triggering a
// "expected a pointer to a Struct" error.
func TestReproPreinitializedBasicPtr(t *testing.T) {
	type config struct {
		BoolPtr   *bool   `env:"BOOL"`
		StringPtr *string `env:"STRING"`
	}

	t.Setenv("BOOL", "true")
	t.Setenv("STRING", "hello")

	b := false
	s := "old"
	cfg := config{
		BoolPtr:   &b,
		StringPtr: &s,
	}

	if err := Parse(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BoolPtr == nil || *cfg.BoolPtr != true {
		t.Errorf("BoolPtr = %v, want true", cfg.BoolPtr)
	}
	if cfg.StringPtr == nil || *cfg.StringPtr != "hello" {
		t.Errorf("StringPtr = %v, want hello", cfg.StringPtr)
	}
}
