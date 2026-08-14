package humanize

import (
	"math/big"
	"testing"
)

func TestBigCommaNoMutation(t *testing.T) {
	tests := []struct {
		name string
		in   string
		exp  string
	}{
		{"trillion", "1000000000000", "1,000,000,000,000"},
		{"large negative", "-84889279597249724975972597249849757294578485", "-84,889,279,597,249,724,975,972,597,249,849,757,294,578,485"},
		{"small", "123", "123"},
		{"zero", "0", "0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			n, ok := (&big.Int{}).SetString(test.in, 10)
			if !ok {
				t.Fatalf("bad input %q", test.in)
			}
			orig := (&big.Int{}).Set(n)

			got1 := BigComma(n)
			if got1 != test.exp {
				t.Errorf("first call: expected %q, got %q", test.exp, got1)
			}
			if n.Cmp(orig) != 0 {
				t.Errorf("first call mutated input: have %s, want %s", n.String(), orig.String())
			}

			got2 := BigComma(n)
			if got2 != test.exp {
				t.Errorf("second call: expected %q, got %q", test.exp, got2)
			}
			if n.Cmp(orig) != 0 {
				t.Errorf("second call mutated input: have %s, want %s", n.String(), orig.String())
			}
		})
	}
}
