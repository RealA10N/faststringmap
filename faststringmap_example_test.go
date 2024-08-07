package faststringmap_test

import (
	"fmt"
	"sort"

	"alon.kr/x/faststringmap"
)

func Example() {
	m := []faststringmap.MapEntry[uint32]{
		{"key1", 42},
		{"key2", 27644437},
		{"l", 2},
	}

	fm := faststringmap.NewMap[uint32](m)

	// add an entry that is not in the fast map
	m = append(m, faststringmap.MapEntry[uint32]{"m", 4})

	// sort the keys so output is the same for each test run
	keys := make([]string, 0, len(m))
	for _, e := range m {
		keys = append(keys, e.Key)
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
