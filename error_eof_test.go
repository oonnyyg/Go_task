package toml_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// TestErrorEOFNoNewline verifies that a syntax error reaching EOF on a final
// line without a trailing newline is reported with a readable position.
// Previously this panicked with "index out of range" because the EOF error
// position was moved below line 1.
func TestErrorEOFNoNewline(t *testing.T) {
	var m map[string]any
	_, err := toml.Decode("x = \"unclosed", &m)
	if err == nil {
		t.Fatal("err is nil")
	}
	var pErr toml.ParseError
	if !errors.As(err, &pErr) {
		t.Fatalf("not a ParseError: %T", err)
	}

	// 10 spaces line prefix + (Col-1) spaces point at the last character.
	want := "toml: error: unexpected EOF; expected '\"'\n\n" +
		"At line 1, column 13:\n\n" +
		"      1 | x = \"unclosed\n" +
		strings.Repeat(" ", 22) + "^\n"
	if have := pErr.ErrorWithUsage(); have != want {
		t.Errorf("\nwant:\n%s\nhave:\n%s", want, have)
	}

	// Multi-line input without a trailing newline must point at the line with
	// the unclosed string, not be moved back to the previous line.
	_, err = toml.Decode("a = 1\nx = \"unclosed", &m)
	if err == nil {
		t.Fatal("err is nil")
	}
	if !errors.As(err, &pErr) {
		t.Fatalf("not a ParseError: %T", err)
	}
	if pErr.Position.Line != 2 {
		t.Errorf("line: want 2, have %d", pErr.Position.Line)
	}
}
