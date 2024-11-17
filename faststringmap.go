// Copyright 2021 The Sensible Code Company Ltd
// Author: Duncan Harris & Alon Krymgand

package faststringmap

import (
	"sort"
)

type Uint = uint32

type (
	// Map[T] is a fast read only map from string to generic type T
	// Lookups are about 5x faster than the built-in Go map type
	Map[T any] struct {
		store  []mapInternalNode[T]
		values []T
	}

	// MapEntry[T] is for supplying data to initialize a new map
	MapEntry[T any] struct {
		Key   string
		Value T
	}

	mapInternalNode[T any] struct {
		nextLo      Uint // index in store of next mapEntry
		nextLen     byte // number of mapEntries in store used for next possible bytes
		nextOffset  byte // offset from zero byte value of first element of range of mapEntries
		valueOffset Uint // index+1 in values for byte sequence with no more bytes. 0 if not valid
	}

	mapBuilder[T any] struct {
		stores [][]mapInternalNode[T]
		values []T
		len    Uint
	}
)

// NewMap[T] creates from the provided map entries
func NewMap[T any](entries []MapEntry[T]) Map[T] {
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })

	b := mapBuilder[T]{}
	root := b.allocateNodes(1)
	if len(entries) > 0 {
		b.makeEntry(&root[0], entries, 0)
	}

	return b.toMap()
}

// FromMap[T] constructs a new faststringmap from a builtin Go map
func FromMap[T any](m map[string]T) Map[T] {
	entries := make([]MapEntry[T], 0, len(m))
	for k, v := range m {
		entries = append(entries, MapEntry[T]{k, v})
	}

	return NewMap[T](entries)
}

// makeEntry will initialize the supplied mapInternalNode for
// the sorted strings in slice a considering bytes at entryIndex in the strings
func (b *mapBuilder[T]) makeEntry(node *mapInternalNode[T], entries []MapEntry[T], entryIndex int) {
	// if there is a string with no more bytes then it is always first because they are sorted
	if len(entries[0].Key) == entryIndex {
		b.values = append(b.values, entries[0].Value)
		node.valueOffset = uint32(len(b.values))
		entries = entries[1:]
	}

	if len(entries) == 0 {
		return
	}

	node.nextOffset = entries[0].Key[entryIndex]             // lowest value for next byte
	node.nextLen = entries[len(entries)-1].Key[entryIndex] - // highest value for next byte
		node.nextOffset + 1 // minus lowest value +1 = number of possible next bytes
	node.nextLo = uint32(b.len)           // first mapEntry struct in eventual built slice
	next := b.allocateNodes(node.nextLen) // new mapInternalNodes default to "not valid"

	for i, n := 0, len(entries); i < n; {
		// find range of strings starting with the same byte
		iSameByteHi := i + 1
		for iSameByteHi < n && entries[iSameByteHi].Key[entryIndex] == entries[i].Key[entryIndex] {
			iSameByteHi++
		}
		b.makeEntry(
			&next[(entries[i].Key[entryIndex]-node.nextOffset)],
			entries[i:iSameByteHi],
			entryIndex+1,
		)
		i = iSameByteHi
	}
}

func (b *mapBuilder[T]) allocateNodes(n byte) []mapInternalNode[T] {
	store := make([]mapInternalNode[T], n)
	b.stores = append(b.stores, store)
	b.len += uint32(n)
	return store
}

func (b *mapBuilder[T]) toMap() Map[T] {
	m := Map[T]{
		store:  make([]mapInternalNode[T], 0, b.len),
		values: b.values,
	}

	for _, store := range b.stores {
		m.store = append(m.store, store...)
	}

	return m
}

// MARK: Index

// IndexString returns the index of the value in the map for the supplied
// string, or 0 if the value is not present in the map. Use AtIndex() to get
// the value using the resulting index.
func (m *Map[T]) IndexString(s string) Uint {
	bv := &m.store[0]
	for i, n := 0, len(s); i < n; i++ {
		b := s[i]
		if b < bv.nextOffset {
			return 0
		}
		ni := b - bv.nextOffset
		if ni >= bv.nextLen {
			return 0
		}
		bv = &m.store[bv.nextLo+uint32(ni)]
	}

	if bv.valueOffset == 0 {
		return 0
	}

	return bv.valueOffset
}

// IndexBytes returns the index of the value in the map for the supplied
// byte slice, or 0 if the value is not present in the map. Use AtIndex() to get
// the value using the resulting index.
func (m *Map[T]) IndexBytes(s []byte) Uint {
	bv := &m.store[0]
	for _, b := range s {
		if b < bv.nextOffset {
			return 0
		}
		ni := b - bv.nextOffset
		if ni >= bv.nextLen {
			return 0
		}
		bv = &m.store[bv.nextLo+uint32(ni)]
	}

	if bv.valueOffset == 0 {
		return 0
	}

	return bv.valueOffset
}

// MARK: Lookup

// LookupString looks up the supplied string in the map
func (m *Map[T]) LookupString(s string) (t T, ok bool) {
	return m.AtIndex(m.IndexString(s))
}

// LookupBytes looks up the supplied byte slice in the map
func (m *Map[T]) LookupBytes(s []byte) (t T, ok bool) {
	return m.AtIndex(m.IndexBytes(s))
}

// MARK: At

// AtIndex returns the value in the map at the supplied internal index
func (m *Map[T]) AtIndex(index Uint) (t T, ok bool) {
	if index != 0 && index-1 < Uint(len(m.values)) {
		return m.values[index-1], true
	} else {
		return t, false
	}
}
