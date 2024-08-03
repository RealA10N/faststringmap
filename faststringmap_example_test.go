package faststringmap_test

import (
	"fmt"
	"sort"

	"github.com/RealA10N/faststringmap"
)

func Example() {
	m := exampleSource{
		"key1": 42,
		"key2": 27644437,
		"l":    2,
	}

	fm := faststringmap.NewMap[uint32](m)

	// add an entry that is not in the fast map
	m["m"] = 4

	// sort the keys so output is the same for each test run
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// lookup every key in the fast map and print the corresponding value
	for _, k := range keys {
		v, ok := fm.LookupString(k)
		fmt.Printf("%q: %d, %v\n", k, v, ok)
	}

	// Output:
	//
	// "key1": 42, true
	// "key2": 27644437, true
	// "l": 2, true
	// "m": 0, false
}

type exampleSource map[string]uint32

func (s exampleSource) AppendKeys(a []string) []string {
	for k := range s {
		a = append(a, k)
	}
	return a
}

func (s exampleSource) Get(k string) uint32 {
	return s[k]
}
