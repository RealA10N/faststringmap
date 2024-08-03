// Copyright 2021 The Sensible Code Company Ltd
// Author: Duncan Harris

package faststringmap

import (
	"sort"
)

type (
	// Uint32Store is a fast read only map from string to uint32
	// Lookups are about 5x faster than the built-in Go map type
	Map[T any] struct {
		store []mapEntry[T]
	}

	mapEntry[T any] struct {
		nextLo     uint32 // index in store of next byteValues
		nextLen    byte   // number of byteValues in store used for next possible bytes
		nextOffset byte   // offset from zero byte value of first element of range of byteValues
		valid      bool   // is the byte sequence with no more bytes in the map?
		value      T      // value for byte sequence with no more bytes
	}

	// Uint32Source is for supplying data to initialize Uint32Store
	MapSource[T any] interface {
		// AppendKeys should append the keys of the maps to the supplied slice and return the resulting slice
		AppendKeys([]string) []string
		// Get should return the value for the supplied key
		Get(string) T
	}

	// uint32Builder is used only during construction
	mapBuilder[T any] struct {
		all [][]mapEntry[T]
		src MapSource[T]
		len int
	}
)

// NewUint32Store creates from the data supplied in src
func NewMap[T any](src MapSource[T]) Map[T] {
	if keys := src.AppendKeys([]string(nil)); len(keys) > 0 {
		sort.Strings(keys)
		return Map[T]{store: mapBuild(keys, src)}
	}
	return Map[T]{store: []mapEntry[T]{{}}}
}

// uint32Build constructs the map by allocating memory in blocks
// and then copying into the eventual slice at the end. This is
// more efficient than continually using append.
func mapBuild[T any](keys []string, src MapSource[T]) []mapEntry[T] {
	b := mapBuilder[T]{
		all: [][]mapEntry[T]{make([]mapEntry[T], 1, firstBufSize(len(keys)))},
		src: src,
		len: 1,
	}
	b.makeEntry(&b.all[0][0], keys, 0)
	// copy all blocks to one slice
	s := make([]mapEntry[T], 0, b.len)
	for _, a := range b.all {
		s = append(s, a...)
	}
	return s
}

// makeByteValue will initialize the supplied byteValue for
// the sorted strings in slice a considering bytes at byteIndex in the strings
func (b *mapBuilder[T]) makeEntry(bv *mapEntry[T], a []string, byteIndex int) {
	// if there is a string with no more bytes then it is always first because they are sorted
	if len(a[0]) == byteIndex {
		bv.valid = true
		bv.value = b.src.Get(a[0])
		a = a[1:]
	}
	if len(a) == 0 {
		return
	}
	bv.nextOffset = a[0][byteIndex]       // lowest value for next byte
	bv.nextLen = a[len(a)-1][byteIndex] - // highest value for next byte
		bv.nextOffset + 1 // minus lowest value +1 = number of possible next bytes
	bv.nextLo = uint32(b.len)   // first byteValue struct in eventual built slice
	next := b.alloc(bv.nextLen) // new byteValues default to "not valid"

	for i, n := 0, len(a); i < n; {
		// find range of strings starting with the same byte
		iSameByteHi := i + 1
		for iSameByteHi < n && a[iSameByteHi][byteIndex] == a[i][byteIndex] {
			iSameByteHi++
		}
		b.makeEntry(&next[(a[i][byteIndex]-bv.nextOffset)], a[i:iSameByteHi], byteIndex+1)
		i = iSameByteHi
	}
}

const maxBuildBufSize = 1 << 20

func firstBufSize(mapSize int) int {
	size := 1 << 4
	for size < mapSize && size < maxBuildBufSize {
		size <<= 1
	}
	return size
}

// alloc will grab space in the current block if available or allocate a new one if not
func (b *mapBuilder[T]) alloc(nByteValues byte) []mapEntry[T] {
	n := int(nByteValues)
	b.len += n
	cur := &b.all[len(b.all)-1] // current
	curCap, curLen := cap(*cur), len(*cur)
	if curCap-curLen >= n { // enough space in current
		*cur = (*cur)[: curLen+n : curCap]
		return (*cur)[curLen:]
	}
	newCap := curCap * 2
	for newCap < n {
		newCap *= 2
	}
	if newCap > maxBuildBufSize {
		newCap = maxBuildBufSize
	}
	a := make([]mapEntry[T], n, newCap)
	b.all = append(b.all, a)
	return a
}

// LookupString looks up the supplied string in the map
func (m *Map[T]) LookupString(s string) (T, bool) {
	bv := &m.store[0]
	for i, n := 0, len(s); i < n; i++ {
		b := s[i]
		if b < bv.nextOffset {
			var t T
			return t, false
		}
		ni := b - bv.nextOffset
		if ni >= bv.nextLen {
			var t T
			return t, false
		}
		bv = &m.store[bv.nextLo+uint32(ni)]
	}
	return bv.value, bv.valid
}

// LookupBytes looks up the supplied byte slice in the map
func (m *Map[T]) LookupBytes(s []byte) (T, bool) {
	bv := &m.store[0]
	for _, b := range s {
		if b < bv.nextOffset {
			var t T
			return t, false
		}
		ni := b - bv.nextOffset
		if ni >= bv.nextLen {
			var t T
			return t, false
		}
		bv = &m.store[bv.nextLo+uint32(ni)]
	}
	return bv.value, bv.valid
}
