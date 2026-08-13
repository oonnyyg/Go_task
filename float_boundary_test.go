package decimal

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// TestNewFromFloatShortestBoundary guards against a regression in roundShortest
// where some large-magnitude float64 boundary values were rounded one digit too
// far down: the last significant digit ended up one too small (and the symmetric
// error occurred for negatives). Each value below previously produced a wrong
// Decimal; the expected string is the shortest decimal that round-trips, i.e.
// strconv.FormatFloat(f, 'f', -1, 64).
func TestNewFromFloatShortestBoundary(t *testing.T) {
	cases := []struct {
		bits uint64
		want string
	}{
		{0x439836f18e020b8f, "436211874468651970"},
		{0x439842ca588a210b, "437045521749131970"},
		{0x43901d1a5538a677, "290278308064107970"},
		{0x439d825bc741da41, "531590721358434370"},
		{0x43e03d1c2212bbcb, "9360979286839679000"},
		{0xc3bdf108b246e5bb, "-2157515258273839900"},
		{0xc3b185132807e1d3, "-1262436333199151900"},
		{0xc45f0c4f65094371, "-2290944420108287900000"},
	}
	for _, c := range cases {
		f := math.Float64frombits(c.bits)
		got := NewFromFloat(f)
		if got.String() != c.want {
			t.Errorf("bits=%#x f=%v: expected %s, got %s (value=%s exp=%d)",
				c.bits, f, c.want, got.String(), got.value.String(), got.exp)
		}
		shortest, err := NewFromString(strconv.FormatFloat(f, 'f', -1, 64))
		if err != nil {
			t.Fatalf("bits=%#x: strconv parse error: %v", c.bits, err)
		}
		if !got.Equal(shortest) {
			t.Errorf("bits=%#x f=%v: not equal to strconv shortest %s", c.bits, f, shortest.String())
		}
	}
}

// TestNewFromFloatShortestRandom compares NewFromFloat against the shortest
// round-trip decimal from strconv across the full float64 range, including
// subnormals and large magnitudes. The default quick.Check run is too small to
// reliably catch the boundary regression above, so this broadens coverage.
func TestNewFromFloatShortestRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5eed))
	check := func(f float64) {
		if math.IsNaN(f) || math.IsInf(f, 0) || f == 0 {
			return
		}
		want, err := NewFromString(strconv.FormatFloat(f, 'f', -1, 64))
		if err != nil {
			return
		}
		got := NewFromFloat(f)
		if !got.Equal(want) {
			t.Fatalf("f=%v bits=%#x want=%s got=%s", f, math.Float64bits(f), want.String(), got.String())
		}
	}
	for i := 0; i < 100000; i++ {
		check(math.Float64frombits(rng.Uint64()))
	}
	for i := 0; i < 30000; i++ {
		m := rng.Uint64() & ((1 << 52) - 1)
		if m == 0 {
			m = 1
		}
		bits := m
		if rng.Uint64()&1 == 1 {
			bits |= 1 << 63
		}
		check(math.Float64frombits(bits))
	}
}
