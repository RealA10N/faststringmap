// Copyright 2021 The Sensible Code Company Ltd
// Author: Duncan Harris & Alon Krymgand

package faststringmap

import (
	"sort"
)

type (
	// Map[T] is a fast read only map from string to generic type T
	// Lookups are about 5x faster than the built-in Go map type
	Map[T any] struct {
		store  []mapEntry[T]
		values []T
	}

	mapEntry[T any] struct {
		nextLo      uint32 // index in store of next mapEntry
		nextLen     byte   // number of mapEntries in store used for next possible bytes
		nextOffset  byte   // offset from zero byte value of first element of range of mapEntries
		valueOffset uint32 // index+1 in values for byte sequence with no more bytes. 0 if not valid
	}

	// Uint32Source is for supplying data to initialize Uint32Store
	MapSource[T any] interface {
		// AppendKeys should append the keys of the maps to the supplied slice and return the resulting slice
		AppendKeys([]string) []string
		// Get should return the value for the supplied key
		Get(string) T
	}

	mapBuilder[T any] struct {
		stores [][]mapEntry[T]
		values []T
		src    MapSource[T]
		len    uint32
	}
)

// NewUint32Store creates from the data supplied in src
func NewMap[T any](src MapSource[T]) Map[T] {
	keys := src.AppendKeys([]string(nil))
	sort.Strings(keys)

	b := mapBuilder[T]{src: src}
	root := b.allocateEntries(1)
	if len(keys) > 0 {
		b.makeEntry(&root[0], keys, 0)
	}

	return b.toMap()
}

// makeEntry will initialize the supplied mapEntry for
// the sorted strings in slice a considering bytes at entryIndex in the strings
func (b *mapBuilder[T]) makeEntry(bv *mapEntry[T], a []string, entryIndex int) {
	// if there is a string with no more bytes then it is always first because they are sorted
	if len(a[0]) == entryIndex {
		b.values = append(b.values, b.src.Get(a[0]))
		bv.valueOffset = uint32(len(b.values))
		a = a[1:]
	}

	if len(a) == 0 {
		return
	}

	bv.nextOffset = a[0][entryIndex]       // lowest value for next byte
	bv.nextLen = a[len(a)-1][entryIndex] - // highest value for next byte
		bv.nextOffset + 1 // minus lowest value +1 = number of possible next bytes
	bv.nextLo = uint32(b.len)             // first mapEntry struct in eventual built slice
	next := b.allocateEntries(bv.nextLen) // new mapEntries default to "not valid"

	for i, n := 0, len(a); i < n; {
		// find range of strings starting with the same byte
		iSameByteHi := i + 1
		for iSameByteHi < n && a[iSameByteHi][entryIndex] == a[i][entryIndex] {
			iSameByteHi++
		}
		b.makeEntry(&next[(a[i][entryIndex]-bv.nextOffset)], a[i:iSameByteHi], entryIndex+1)
		i = iSameByteHi
	}
}

func (b *mapBuilder[T]) allocateEntries(n byte) []mapEntry[T] {
	store := make([]mapEntry[T], n)
	b.stores = append(b.stores, store)
	b.len += uint32(n)
	return store
}

func (b *mapBuilder[T]) toMap() Map[T] {
	m := Map[T]{
		store:  make([]mapEntry[T], 0, b.len),
		values: b.values,
	}

	for _, store := range b.stores {
		m.store = append(m.store, store...)
	}

	return m
}

// LookupString looks up the supplied string in the map
func (m *Map[T]) LookupString(s string) (t T, ok bool) {
	bv := &m.store[0]
	for i, n := 0, len(s); i < n; i++ {
		b := s[i]
		if b < bv.nextOffset {
			return t, false
		}
		ni := b - bv.nextOffset
		if ni >= bv.nextLen {
			return t, false
		}
		bv = &m.store[bv.nextLo+uint32(ni)]
	}

	if bv.valueOffset == 0 {
		return t, false
	}

	return m.values[bv.valueOffset-1], true
}

// LookupBytes looks up the supplied byte slice in the map
func (m *Map[T]) LookupBytes(s []byte) (t T, ok bool) {
	bv := &m.store[0]
	for _, b := range s {
		if b < bv.nextOffset {
			return t, false
		}
		ni := b - bv.nextOffset
		if ni >= bv.nextLen {
			return t, false
		}
		bv = &m.store[bv.nextLo+uint32(ni)]
	}

	if bv.valueOffset == 0 {
		return t, false
	}

	return m.values[bv.valueOffset-1], true
}
