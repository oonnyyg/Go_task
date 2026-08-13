// Copyright 2023 Google Inc.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uuid

import (
	"io"
	"sync"
)

// UUID version 7 features a time-ordered value field derived from the widely
// implemented and well known Unix Epoch timestamp source,
// the number of milliseconds seconds since midnight 1 Jan 1970 UTC, leap seconds excluded.
// As well as improved entropy characteristics over versions 1 or 6.
//
// see https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format-03#name-uuid-version-7
//
// Implementations SHOULD utilize UUID version 7 over UUID version 1 and 6 if possible.
//
// NewV7 returns a Version 7 UUID based on the current time(Unix Epoch).
// Uses the randomness pool if it was enabled with EnableRandPool.
//
// Successive calls to NewV7 are strictly increasing in byte order: within the
// same millisecond the random bits (rand_a and rand_b) are treated as a
// monotonic counter instead of being re-randomized, so the returned UUIDs can
// be sorted as strings or byte slices and stay in generation order.
//
// On error, NewV7 returns Nil and an error
func NewV7() (UUID, error) {
	uuid, err := NewRandom()
	if err != nil {
		return uuid, err
	}
	makeV7(uuid[:])
	return uuid, nil
}

// NewV7FromReader returns a Version 7 UUID based on the current time(Unix Epoch).
// it use NewRandomFromReader fill random bits.
// On error, NewV7FromReader returns Nil and an error.
//
// Unlike NewV7, NewV7FromReader does not enforce monotonic ordering: the random
// bits are taken from r on every call, so two UUIDs produced in the same
// millisecond are not ordered by their creation order.
func NewV7FromReader(r io.Reader) (UUID, error) {
	uuid, err := NewRandomFromReader(r)
	if err != nil {
		return uuid, err
	}

	makeV7Random(uuid[:])
	return uuid, nil
}

// makeV7 fill 48 bits time (uuid[0] - uuid[5]), set version b0111 (uuid[6])
// uuid[8] already has the right version number (Variant is 10)
// see function  NewV7 and NewV7FromReader
//
// makeV7 is monotonic: when the millisecond timestamp has not advanced since
// the previous call it reuses that timestamp and increments the 74-bit random
// counter (rand_a then rand_b), carrying into the timestamp on overflow, so
// that successive UUIDs are strictly increasing in byte order and can be
// sorted directly as strings or byte slices.
func makeV7(uuid []byte) {
	/*
		 0                   1                   2                   3
		 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		|                           unix_ts_ms                          |
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		|          unix_ts_ms           |  ver  |       rand_a          |
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		|var|                        rand_b                             |
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		|                            rand_b                             |
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	*/
	_ = uuid[15] // bounds check

	t := uint64(timeNow().UnixMilli())

	v7Mu.Lock()
	defer v7Mu.Unlock()

	if t > v7Last {
		// New timestamp: adopt the fresh random bits that NewRandom
		// already placed in uuid.
		v7RandA = uint16(uuid[6]&0x0f)<<8 | uint16(uuid[7])
		v7RandB = uint64(uuid[8]&0x3f)<<56 |
			uint64(uuid[9])<<48 |
			uint64(uuid[10])<<40 |
			uint64(uuid[11])<<32 |
			uint64(uuid[12])<<24 |
			uint64(uuid[13])<<16 |
			uint64(uuid[14])<<8 |
			uint64(uuid[15])
	} else {
		// Same or earlier millisecond: reuse the last timestamp and
		// increment the random counter so the result is strictly greater
		// than the previous UUID, carrying on overflow.
		t = v7Last
		v7RandB++
		if v7RandB >= 1<<62 {
			v7RandB = 0
			v7RandA++
			if v7RandA >= 1<<12 {
				v7RandA = 0
				t++
			}
		}
	}
	v7Last = t

	uuid[0] = byte(t >> 40)
	uuid[1] = byte(t >> 32)
	uuid[2] = byte(t >> 24)
	uuid[3] = byte(t >> 16)
	uuid[4] = byte(t >> 8)
	uuid[5] = byte(t)

	uuid[6] = 0x70 | byte(v7RandA>>8)
	uuid[7] = byte(v7RandA)

	uuid[8] = 0x80 | byte(v7RandB>>56)
	uuid[9] = byte(v7RandB >> 48)
	uuid[10] = byte(v7RandB >> 40)
	uuid[11] = byte(v7RandB >> 32)
	uuid[12] = byte(v7RandB >> 24)
	uuid[13] = byte(v7RandB >> 16)
	uuid[14] = byte(v7RandB >> 8)
	uuid[15] = byte(v7RandB)
}

// v7 monotonic state shared across NewV7 calls.
var (
	v7Mu    sync.Mutex
	v7Last  uint64 // last unix_ts_ms written into a UUIDv7
	v7RandA uint16 // last 12-bit rand_a (after the version nibble)
	v7RandB uint64 // last 62-bit rand_b (after the variant bits)
)

// makeV7Random fills uuid as a Version 7 UUID using the random bits already
// present in uuid (placed by NewRandomFromReader). It sets only the timestamp
// and version bits and, unlike makeV7, does not enforce monotonic ordering.
func makeV7Random(uuid []byte) {
	_ = uuid[15] // bounds check

	t := timeNow().UnixMilli()

	uuid[0] = byte(t >> 40)
	uuid[1] = byte(t >> 32)
	uuid[2] = byte(t >> 24)
	uuid[3] = byte(t >> 16)
	uuid[4] = byte(t >> 8)
	uuid[5] = byte(t)

	uuid[6] = 0x70 | (uuid[6] & 0x0F)
	// uuid[8] has already has right version
}
