package series

import (
	"fmt"
	"testing"
	"time"

	"github.com/caerbannogwhite/enchanter"
)

// Round-trip tests for the Series -> Arrow -> Series conversion across all
// series types, byte-boundary lengths and null patterns. The null-mask
// inversion (enchanter bit-set=null vs Arrow bit-set=valid) is the invariant the
// whole Arrow integration rests on.

var rtLengths = []int{1, 7, 8, 9, 64, 65}

var rtPatterns = []struct {
	name string
	fn   func(i, n int) bool // true = null at index i; nil fn = not nullable
}{
	{"nonullable", nil},
	{"allnull", func(i, n int) bool { return true }},
	{"alternating", func(i, n int) bool { return i%2 == 0 }},
	{"lastonly", func(i, n int) bool { return i == n-1 }},
}

func rtMask(fn func(i, n int) bool, n int) []bool {
	if fn == nil {
		return nil
	}
	mask := make([]bool, n)
	for i := range mask {
		mask[i] = fn(i, n)
	}
	return mask
}

func rtIsNull(fn func(i, n int) bool, i, n int) bool {
	return fn != nil && fn(i, n)
}

func forEachLenPattern(t *testing.T, fn func(t *testing.T, n int, pat func(i, n int) bool)) {
	t.Helper()
	for _, n := range rtLengths {
		for _, p := range rtPatterns {
			t.Run(fmt.Sprintf("n=%d_%s", n, p.name), func(t *testing.T) {
				fn(t, n, p.fn)
			})
		}
	}
}

func checkRoundTripNulls(t *testing.T, got Series, n int, pat func(i, n int) bool) {
	t.Helper()
	if got.Err() != nil {
		t.Fatalf("round trip returned error series: %s", got.Err())
	}
	if got.Len() != n {
		t.Fatalf("length: got %d, want %d", got.Len(), n)
	}
	if got.IsNullable() != (pat != nil) {
		t.Fatalf("IsNullable: got %v, want %v", got.IsNullable(), pat != nil)
	}
	for i := 0; i < n; i++ {
		if got.IsNull(i) != rtIsNull(pat, i, n) {
			t.Fatalf("IsNull(%d): got %v, want %v", i, got.IsNull(i), rtIsNull(pat, i, n))
		}
	}
}

func TestArrowRoundTripFloat64(t *testing.T) {
	ctx := enchanter.NewContext()
	forEachLenPattern(t, func(t *testing.T, n int, pat func(i, n int) bool) {
		data := make([]float64, n)
		for i := range data {
			data[i] = float64(i)*1.5 - 3
		}
		s := NewSeriesFloat64(data, rtMask(pat, n), false, ctx)
		arr := s.ArrowArray()
		defer arr.Release()
		got := ArrowArrayToSeries(arr, ctx)
		checkRoundTripNulls(t, got, n, pat)
		g := got.(Float64s)
		for i := 0; i < n; i++ {
			if !rtIsNull(pat, i, n) && g.Data_[i] != data[i] {
				t.Fatalf("value[%d]: got %v, want %v", i, g.Data_[i], data[i])
			}
		}
	})
}

func TestArrowRoundTripInt64(t *testing.T) {
	ctx := enchanter.NewContext()
	forEachLenPattern(t, func(t *testing.T, n int, pat func(i, n int) bool) {
		data := make([]int64, n)
		for i := range data {
			data[i] = int64(i)*7 - 11
		}
		s := NewSeriesInt64(data, rtMask(pat, n), false, ctx)
		arr := s.ArrowArray()
		defer arr.Release()
		got := ArrowArrayToSeries(arr, ctx)
		checkRoundTripNulls(t, got, n, pat)
		g := got.(Int64s)
		for i := 0; i < n; i++ {
			if !rtIsNull(pat, i, n) && g.Data_[i] != data[i] {
				t.Fatalf("value[%d]: got %v, want %v", i, g.Data_[i], data[i])
			}
		}
	})
}

// Ints are stored as Arrow int64 (Arrow has no platform-dependent int), so the
// round trip intentionally comes back as Int64s.
func TestArrowRoundTripInt(t *testing.T) {
	ctx := enchanter.NewContext()
	forEachLenPattern(t, func(t *testing.T, n int, pat func(i, n int) bool) {
		data := make([]int, n)
		for i := range data {
			data[i] = i*2 - 5
		}
		s := NewSeriesInt(data, rtMask(pat, n), false, ctx)
		arr := s.ArrowArray()
		defer arr.Release()
		got := ArrowArrayToSeries(arr, ctx)
		checkRoundTripNulls(t, got, n, pat)
		g, ok := got.(Int64s)
		if !ok {
			t.Fatalf("expected Int64s after round trip, got %T", got)
		}
		for i := 0; i < n; i++ {
			if !rtIsNull(pat, i, n) && g.Data_[i] != int64(data[i]) {
				t.Fatalf("value[%d]: got %v, want %v", i, g.Data_[i], data[i])
			}
		}
	})
}

func TestArrowRoundTripBool(t *testing.T) {
	ctx := enchanter.NewContext()
	forEachLenPattern(t, func(t *testing.T, n int, pat func(i, n int) bool) {
		data := make([]bool, n)
		for i := range data {
			data[i] = i%3 == 0
		}
		s := NewSeriesBool(data, rtMask(pat, n), false, ctx)
		arr := s.ArrowArray()
		defer arr.Release()
		got := ArrowArrayToSeries(arr, ctx)
		checkRoundTripNulls(t, got, n, pat)
		g := got.(Bools)
		for i := 0; i < n; i++ {
			if !rtIsNull(pat, i, n) && g.Data_[i] != data[i] {
				t.Fatalf("value[%d]: got %v, want %v", i, g.Data_[i], data[i])
			}
		}
	})
}

func TestArrowRoundTripString(t *testing.T) {
	ctx := enchanter.NewContext()
	forEachLenPattern(t, func(t *testing.T, n int, pat func(i, n int) bool) {
		data := make([]string, n)
		for i := range data {
			data[i] = fmt.Sprintf("val_%03d", i)
		}
		s := NewSeriesString(data, rtMask(pat, n), false, ctx)
		arr := s.ArrowArray()
		defer arr.Release()
		got := ArrowArrayToSeries(arr, ctx)
		checkRoundTripNulls(t, got, n, pat)
		g := got.(Strings)
		for i := 0; i < n; i++ {
			if !rtIsNull(pat, i, n) && *g.Data_[i] != data[i] {
				t.Fatalf("value[%d]: got %q, want %q", i, *g.Data_[i], data[i])
			}
		}
	})
}

func TestArrowRoundTripTime(t *testing.T) {
	ctx := enchanter.NewContext()
	base := time.Date(2020, 1, 2, 3, 4, 5, 123456789, time.UTC)
	forEachLenPattern(t, func(t *testing.T, n int, pat func(i, n int) bool) {
		data := make([]time.Time, n)
		for i := range data {
			data[i] = base.Add(time.Duration(i) * time.Minute)
		}
		s := NewSeriesTime(data, rtMask(pat, n), false, ctx)
		arr := s.ArrowArray()
		defer arr.Release()
		got := ArrowArrayToSeries(arr, ctx)
		checkRoundTripNulls(t, got, n, pat)
		g := got.(Times)
		for i := 0; i < n; i++ {
			if !rtIsNull(pat, i, n) && !g.Data_[i].Equal(data[i]) {
				t.Fatalf("value[%d]: got %v, want %v", i, g.Data_[i], data[i])
			}
		}
	})
}

func TestArrowRoundTripDuration(t *testing.T) {
	ctx := enchanter.NewContext()
	forEachLenPattern(t, func(t *testing.T, n int, pat func(i, n int) bool) {
		data := make([]time.Duration, n)
		for i := range data {
			data[i] = time.Duration(i+1) * time.Millisecond
		}
		s := NewSeriesDuration(data, rtMask(pat, n), false, ctx)
		arr := s.ArrowArray()
		defer arr.Release()
		got := ArrowArrayToSeries(arr, ctx)
		checkRoundTripNulls(t, got, n, pat)
		g := got.(Durations)
		for i := 0; i < n; i++ {
			if !rtIsNull(pat, i, n) && g.Data_[i] != data[i] {
				t.Fatalf("value[%d]: got %v, want %v", i, g.Data_[i], data[i])
			}
		}
	})
}

// Regression tests: a series created FROM an Arrow array must not serve a
// stale Arrow representation after its data is mutated (Sort/Set mutate
// Data_ in place).

func TestArrowBornSeriesSortNotStale(t *testing.T) {
	ctx := enchanter.NewContext()
	orig := NewSeriesFloat64([]float64{3, 1, 2}, nil, false, ctx)
	arrowBorn := ArrowArrayToSeries(orig.ArrowArray(), ctx)

	sorted := arrowBorn.Sort()

	arr := sorted.ArrowArray()
	defer arr.Release()
	got := ArrowArrayToSeries(arr, ctx).(Float64s)
	want := []float64{1, 2, 3}
	for i, w := range want {
		if got.Data_[i] != w {
			t.Fatalf("stale Arrow array after Sort(): got %v, want %v", got.Data_, want)
		}
	}
}

func TestArrowBornSeriesSetNotStale(t *testing.T) {
	ctx := enchanter.NewContext()
	orig := NewSeriesInt64([]int64{10, 20, 30}, nil, false, ctx)
	arrowBorn := ArrowArrayToSeries(orig.ArrowArray(), ctx)

	updated := arrowBorn.Set(1, int64(99))

	arr := updated.ArrowArray()
	defer arr.Release()
	got := ArrowArrayToSeries(arr, ctx).(Int64s)
	if got.Data_[1] != 99 {
		t.Fatalf("stale Arrow array after Set(): got %v, want [10 99 30]", got.Data_)
	}
}
