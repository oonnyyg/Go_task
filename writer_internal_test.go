package logrus

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"time"
)

// runWriterScanner drives entry.writerScanner against a fresh pipe, writes the
// given fragments in order, closes the writer, and waits for the scanner
// goroutine to return. It fails the test if the scanner does not return within
// a short timeout, which happens when it panics or deadlocks.
func runWriterScanner(t *testing.T, writes ...[]byte) []string {
	t.Helper()

	entry := NewEntry(New())
	reader, writer := io.Pipe()

	done := make(chan struct{})
	var tokens []string
	printFunc := func(args ...interface{}) {
		tokens = append(tokens, args[0].(string))
	}

	go func() {
		entry.writerScanner(reader, printFunc)
		close(done)
	}()

	for _, w := range writes {
		if _, err := writer.Write(w); err != nil {
			reader.Close()
			t.Fatalf("pipe write: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		reader.Close()
		t.Fatalf("pipe close: %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("writerScanner did not return")
	}
	return tokens
}

// TestWriterScannerEOFNoPanic reproduces the panic
// "bufio.Scan: too many empty tokens without progressing" that was triggered
// when the writer reached EOF after writing newline-bearing fragments.
func TestWriterScannerEOFNoPanic(t *testing.T) {
	tokens := runWriterScanner(t,
		[]byte("line one\n"),
		[]byte("line two\nline three\n"),
		[]byte("partial without newline"),
	)

	want := []string{"line one", "line two", "line three", "partial without newline"}
	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens %q, want %d", len(tokens), tokens, len(want))
	}
	for i := range want {
		if tokens[i] != want[i] {
			t.Fatalf("token %d: got %q, want %q", i, tokens[i], want[i])
		}
	}
}

// TestWriterScannerLargeWriteNoPanic ensures that writing a single chunk
// larger than bufio.MaxScanTokenSize does not crash the writer.
func TestWriterScannerLargeWriteNoPanic(t *testing.T) {
	big := strings.Repeat("a", bufio.MaxScanTokenSize+4096)

	tokens := runWriterScanner(t, []byte(big))

	if got := strings.Join(tokens, ""); got != big {
		t.Fatalf("reconstructed %d bytes, want %d", len(got), len(big))
	}
}
