package series

import "testing"

// Data returns the backing slice: writes through the returned slice are
// visible in the series.
func Test_Data_IsAView(t *testing.T) {
	s := NewSeriesFloat64([]float64{1, 2, 3}, nil, true, ctx)

	s.Float64s()[1] = 99
	if s.Get(1).(float64) != 99 {
		t.Fatal("Float64s() must be a view of the series storage")
	}

	s.Data().([]float64)[2] = 42
	if s.Get(2).(float64) != 42 {
		t.Fatal("Data() must be a view of the series storage")
	}
}

// NullMask materializes a fresh []bool: writes to it do not touch the
// series.
func Test_NullMask_IsACopy(t *testing.T) {
	s := NewSeriesInt([]int{1, 2, 3}, []bool{false, true, false}, true, ctx)

	m := s.NullMask()
	if !m[1] || m[0] || m[2] {
		t.Fatalf("mask: expected [false true false], got %v", m)
	}
	m[0] = true
	if s.IsNull(0) {
		t.Fatal("writing to the NullMask() result must not change the series")
	}
}

// PackedNullMask exposes the stored bit-packed mask: empty without
// nulls, and in agreement with IsNull when nullable.
func Test_PackedNullMask(t *testing.T) {
	plain := NewSeriesInt64([]int64{1, 2, 3}, nil, true, ctx)
	if len(plain.PackedNullMask()) != 0 {
		t.Fatal("a series without nulls has an empty packed mask")
	}

	nulls := []bool{false, true, false, true, false, false, false, false, true, false}
	s := NewSeriesFloat64(make([]float64, 10), nulls, true, ctx)
	packed := s.PackedNullMask()
	for i := 0; i < s.Len(); i++ {
		bit := packed[i>>3]&(1<<uint(i%8)) != 0
		if bit != s.IsNull(i) {
			t.Fatalf("bit %d: packed says %v, IsNull says %v", i, bit, s.IsNull(i))
		}
	}
}

// Interned returns the raw pool-interned pointers: equal values share a
// pointer, and null elements hold the interned NA text.
func Test_Interned(t *testing.T) {
	s := NewSeriesString([]string{"a", "b", "a"}, []bool{false, true, false}, true, ctx)

	p := s.Interned()
	if len(p) != 3 {
		t.Fatalf("length: expected 3, got %d", len(p))
	}
	if *p[0] != "a" {
		t.Fatalf("value: expected a, got %q", *p[0])
	}
	if p[0] != p[2] {
		t.Fatal("equal values must share one interned pointer")
	}
}
