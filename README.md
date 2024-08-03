# faststringmap

`faststringmap` is a fast read-only string keyed map for Go (golang).
For our use case it is approximately 5 times faster than using Go's
built-in map type with a string key. It also has the following advantages:

* look up strings and byte slices without use of the `unsafe` package
* minimal impact on GC due to lack of pointers in the data structure
* data structure can be trivially serialized to disk or network
* supports any type as the value type (via generics)

The module provided a generic `Map[T]` type that implements a map
string from `string` (or `[]byte`) to a generic type `T`.

`faststringmap` is a variant of a data structure called a [Trie](https://en.wikipedia.org/wiki/Trie).
At each level we use a slice to hold the next possible byte values.
This slice is of length one plus the difference between the lowest and highest
possible next bytes of strings in the map. Not all the entries in the slice are
valid next bytes. `faststringmap` is thus more space efficient for keys using a
small set of nearby runes, for example those using a lot of digits.

## Example

Example usage can be found in [``faststringmap_example_test.go``](faststringmap_example_test.go).

## Motivation

[Duncan Harris](https://github.com/duncanharris) first created
[faststringmap](https://github.com/sensiblecodeio/faststringmap) in order to
improve the speed of parsing CSV where the fields were category codes from
survey data.
The majority of these were numeric (`"1"`, `"2"`, `"3"`...) plus a distinct
code for "not applicable".
I was struck that in the simplest possible cases (e.g. `"1"` ... `"5"`) the map
should be a single slice lookup.

I then forked the origin repo, and improved on Duncan's implementation:

* Added generic value type support
* Simplified overall implementation

## Benchmarks

Below are example benchmarks from my laptop which are for looking up every element
in a map of size 1000. So approximate times are 25ns per lookup for the Go native map
and 5ns per lookup for the ``faststringmap``.

```
cpu: Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz
BenchmarkUint32Store
BenchmarkUint32Store-8        	  218463	      4959 ns/op
BenchmarkGoStringToUint32
BenchmarkGoStringToUint32-8   	   49279	     24483 ns/op
```
