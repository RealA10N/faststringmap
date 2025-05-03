// Copyright 2021 The Sensible Code Company Ltd
// Author: Duncan Harris & Alon Krymgand

package faststringmap_test

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"alon.kr/x/faststringmap"
)

// The expected behavior is for the lookup methods to work even when the map is
// a nil pointer.
func TestNilMapLookup(t *testing.T) {
	m := (*faststringmap.Map[uint32])(nil)

	v, found := m.LookupString("foo")
	if found {
		t.Errorf("LookupString(foo) = %v, expected not to be present", v)
	}

	v, found = m.LookupBytes([]byte{1, 2, 3})
	if found {
		t.Errorf("LookupBytes(1,2,3) = %v, expected not to be present", v)
	}
}

func TestZeroLengthMapLookup(t *testing.T) {
	// This creates a map with zero length on the stack, but not nil.
	m := faststringmap.NewMap[uint32](nil)

	v, found := m.LookupString("foo")
	if found {
		t.Errorf("LookupString(foo) = %v, expected not to be present", v)
	}

	v, found = m.LookupBytes([]byte{1, 2, 3})
	if found {
		t.Errorf("LookupBytes(1,2,3) = %v, expected not to be present", v)
	}
}

func TestUintMapSimpleCase(t *testing.T) {
	desc := mapTestDescription[uint32]{
		in: []faststringmap.MapEntry[uint32]{
			{"aaa", 1},
			{"aab", 1},
		},
	}
	testAgainstDescriptor(t, desc)
}

func TestFastStringToUint32Empty(t *testing.T) {
	desc := mapTestDescription[uint32]{
		in: []faststringmap.MapEntry[uint32]{
			{"", 1},
			{"a", 2},
			{"foo", 3},
			{"ß", 4},
		},
	}
	testAgainstDescriptor(t, desc)
}
func TestFastStringToUint32BigSpan(t *testing.T) {
	desc := mapTestDescription[uint32]{
		in: []faststringmap.MapEntry[uint32]{
			{"a!", 1},
			{"a~", 2},
		},
	}
	testAgainstDescriptor(t, desc)
}

func TestFastStringToUint32(t *testing.T) {
	const nStrs = 8192
	allEntries := randomSmallStrings(nStrs, 8)
	inEntries := allEntries[:nStrs/2]
	outKeys := make([]string, 0, nStrs/2)
	for _, e := range allEntries[nStrs/2:] {
		outKeys = append(outKeys, e.Key)
	}
	testAgainstDescriptor(t, mapTestDescription[uint32]{in: inEntries, out: outKeys})
}

type mapTestDescription[T any] struct {
	in  []faststringmap.MapEntry[T]
	out []string
}

func testAgainstDescriptor[T comparable](t *testing.T, desc mapTestDescription[T]) {
	m := faststringmap.NewMap(desc.in)

	for _, e := range desc.in {
		v, ok := m.LookupString(e.Key)
		if !ok || v != e.Value {
			t.Errorf("LookupString(%q) = %v, %v want %v, true", e.Key, v, ok, e.Value)
		}
		v, ok = m.LookupBytes([]byte(e.Key))
		if !ok || v != e.Value {
			t.Errorf("LookupBytes(%q) = %v, %v want %v, true", e.Key, v, ok, e.Value)
		}
	}

	for _, k := range desc.out {
		v, ok := m.LookupString(k)
		if ok {
			t.Errorf("LookupString(%q) = %v, expected not to be present", k, v)
		}
		v, ok = m.LookupBytes([]byte(k))
		if ok {
			t.Errorf("LookupBytes(%q) = %v, expected not to be present", k, v)
		}
	}
}

func randomSmallStrings(nStrs int, maxLen uint8) []faststringmap.MapEntry[uint32] {
	m := map[string]uint32{"": 0}
	for len(m) < nStrs {
		s := randomSmallString(maxLen)
		if _, ok := m[s]; !ok {
			m[s] = uint32(len(m))
		}
	}

	entries := make([]faststringmap.MapEntry[uint32], 0, len(m))
	for k, v := range m {
		entries = append(entries, faststringmap.MapEntry[uint32]{k, v})
	}

	return entries
}

func randomSmallString(maxLen uint8) string {
	var sb strings.Builder
	n := rand.Intn(int(maxLen) + 1)
	for i := 0; i <= n; i++ {
		sb.WriteRune(rand.Int31n(94) + 33)
	}
	return sb.String()
}

func typicalCodeStrings(n int) (m map[string]uint32, keys []string) {
	m = make(map[string]uint32, n)

	add := func(s string) {
		m[s] = uint32(len(m))
		keys = append(keys, s)
	}

	for i := 1; i < n; i++ {
		add(strconv.Itoa(i))
	}
	add("-9")

	return
}

const nStrsBench = 1000

func BenchmarkFastStringMap(b *testing.B) {
	m, keys := typicalCodeStrings(nStrsBench)
	fm := faststringmap.FromMap(m)

	b.ResetTimer()
	for bi := 0; bi < b.N; bi++ {
		for si, n := uint32(0), uint32(len(keys)); si < n; si++ {
			v, ok := fm.LookupString(keys[si])
			if !ok || v != si {
				b.Fatalf("ok=%v, value got %d want %d", ok, v, si)
			}
		}
	}
}

func BenchmarkGoStringMap(b *testing.B) {
	m, keys := typicalCodeStrings(nStrsBench)

	b.ResetTimer()
	for bi := 0; bi < b.N; bi++ {
		for si, n := uint32(0), uint32(len(keys)); si < n; si++ {
			v, ok := m[keys[si]]
			if !ok || v != si {
				b.Fatalf("ok=%v, value got %d want %d", ok, v, si)
			}
		}
	}
}
