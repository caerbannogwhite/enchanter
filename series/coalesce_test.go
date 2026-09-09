package series

import (
	"strings"
	"testing"
	"time"

	"github.com/caerbannogwhite/enchanter/meta"
)

// Coalesce takes the left element where it is not null and the right element
// where it is. A result element is null only when both operands are null at
// that position.
func Test_Coalesce_VectorVector(t *testing.T) {
	s := NewSeriesFloat64([]float64{1, 0, 3, 0, 5}, []bool{false, true, false, true, false}, true, ctx)
	o := NewSeriesFloat64([]float64{10, 20, 30, 0, 50}, []bool{false, false, true, true, false}, true, ctx)

	res := s.Coalesce(o)
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Type() != meta.Float64Type {
		t.Fatalf("type: expected Float64, got %v", res.Type())
	}
	if !res.IsNullable() {
		t.Fatal("both operands have nulls at position 3, the result must be nullable")
	}

	wantVals := []float64{1, 20, 3, 0, 5}
	wantNull := []bool{false, false, false, true, false}
	for i := range wantVals {
		if res.IsNull(i) != wantNull[i] {
			t.Fatalf("null[%d]: expected %v, got %v", i, wantNull[i], res.IsNull(i))
		}
		if !wantNull[i] && res.Get(i).(float64) != wantVals[i] {
			t.Fatalf("value[%d]: expected %v, got %v", i, wantVals[i], res.Get(i))
		}
	}
}

// Filling with a raw Go scalar is the most common use. The scalar has no
// nulls, so the result has none either.
func Test_Coalesce_ScalarFill(t *testing.T) {
	s := NewSeriesFloat64([]float64{1, 0, 3}, []bool{false, true, false}, true, ctx)

	res := s.Coalesce(9.5)
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.IsNullable() {
		t.Fatal("the fill value has no nulls, the result must not be nullable")
	}
	want := []float64{1, 9.5, 3}
	for i, w := range want {
		if res.Get(i).(float64) != w {
			t.Fatalf("value[%d]: expected %v, got %v", i, w, res.Get(i))
		}
	}
}

// A null scalar on the left broadcasts: every position takes the right side.
func Test_Coalesce_NullScalarLeft(t *testing.T) {
	s := NewSeriesFloat64([]float64{0}, []bool{true}, true, ctx)
	o := NewSeriesFloat64([]float64{10, 20, 30}, []bool{false, true, false}, true, ctx)

	res := s.Coalesce(o)
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Len() != 3 {
		t.Fatalf("length: expected 3, got %d", res.Len())
	}
	if !res.IsNull(1) || res.IsNull(0) || res.IsNull(2) {
		t.Fatalf("nulls: expected only position 1, got %v %v %v", res.IsNull(0), res.IsNull(1), res.IsNull(2))
	}
	if res.Get(0).(float64) != 10 || res.Get(2).(float64) != 30 {
		t.Fatalf("values: expected 10 and 30, got %v and %v", res.Get(0), res.Get(2))
	}
}

// Mixed numeric operands widen: Ints with a float fill become Float64s, and
// Int64s with an Ints operand stay Int64s.
func Test_Coalesce_NumericWidening(t *testing.T) {
	i := NewSeriesInt([]int{1, 0, 3}, []bool{false, true, false}, true, ctx)
	res := i.Coalesce(2.5)
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Type() != meta.Float64Type {
		t.Fatalf("Ints coalesce float: expected Float64, got %v", res.Type())
	}
	want := []float64{1, 2.5, 3}
	for j, w := range want {
		if res.Get(j).(float64) != w {
			t.Fatalf("value[%d]: expected %v, got %v", j, w, res.Get(j))
		}
	}

	i64 := NewSeriesInt64([]int64{0, 2}, []bool{true, false}, true, ctx)
	ints := NewSeriesInt([]int{7, 8}, nil, true, ctx)
	res = i64.Coalesce(ints)
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Type() != meta.Int64Type {
		t.Fatalf("Int64s coalesce Ints: expected Int64, got %v", res.Type())
	}
	if res.Get(0).(int64) != 7 || res.Get(1).(int64) != 2 {
		t.Fatalf("values: expected 7 and 2, got %v and %v", res.Get(0), res.Get(1))
	}
}

// An NAs operand contributes only nulls, so the typed side comes through
// unchanged, nullability included.
func Test_Coalesce_WithNAs(t *testing.T) {
	s := NewSeriesFloat64([]float64{1, 0, 3}, []bool{false, true, false}, true, ctx)
	na := NewSeriesNA(3, ctx)

	res := s.Coalesce(na)
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Type() != meta.Float64Type {
		t.Fatalf("type: expected Float64, got %v", res.Type())
	}
	if !res.IsNullable() || !res.IsNull(1) || res.IsNull(0) {
		t.Fatal("coalescing with NAs must keep the receiver's nulls")
	}
	if res.Get(0).(float64) != 1 || res.Get(2).(float64) != 3 {
		t.Fatalf("values: expected 1 and 3, got %v and %v", res.Get(0), res.Get(2))
	}

	// The other direction: NAs filled from a typed series.
	res = na.Coalesce(s)
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Type() != meta.Float64Type || res.Len() != 3 {
		t.Fatalf("NAs coalesce Float64s: expected Float64 of length 3, got %v of %d", res.Type(), res.Len())
	}
	if !res.IsNull(1) || res.Get(0).(float64) != 1 {
		t.Fatal("NAs coalesce must equal the typed operand")
	}

	// NAs filled from a raw scalar broadcasts it.
	res = NewSeriesNA(4, ctx).Coalesce(int64(7))
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Len() != 4 || res.Get(3).(int64) != 7 {
		t.Fatalf("NAs scalar fill: expected four 7s, got len %d, last %v", res.Len(), res.Get(3))
	}
}

func Test_Coalesce_OtherTypes(t *testing.T) {
	// Bools.
	b := NewSeriesBool([]bool{true, false}, []bool{false, true}, true, ctx)
	res := b.Coalesce(NewSeriesBool([]bool{false, true}, nil, true, ctx))
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Get(0).(bool) != true || res.Get(1).(bool) != true {
		t.Fatalf("bools: expected true,true got %v,%v", res.Get(0), res.Get(1))
	}

	// Strings.
	st := NewSeriesString([]string{"a", ""}, []bool{false, true}, true, ctx)
	res = st.Coalesce(NewSeriesString([]string{"x", "y"}, nil, true, ctx))
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Get(0).(string) != "a" || res.Get(1).(string) != "y" {
		t.Fatalf("strings: expected a,y got %v,%v", res.Get(0), res.Get(1))
	}

	// Times.
	t0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2021, 6, 2, 0, 0, 0, 0, time.UTC)
	tm := NewSeriesTime([]time.Time{t0, {}}, []bool{false, true}, true, ctx)
	res = tm.Coalesce(NewSeriesTime([]time.Time{t1, t1}, nil, true, ctx))
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if !res.Get(0).(time.Time).Equal(t0) || !res.Get(1).(time.Time).Equal(t1) {
		t.Fatalf("times: expected %v,%v got %v,%v", t0, t1, res.Get(0), res.Get(1))
	}

	// Durations.
	d := NewSeriesDuration([]time.Duration{time.Second, 0}, []bool{false, true}, true, ctx)
	res = d.Coalesce(NewSeriesDuration([]time.Duration{time.Minute, time.Hour}, nil, true, ctx))
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.Get(0).(time.Duration) != time.Second || res.Get(1).(time.Duration) != time.Hour {
		t.Fatalf("durations: expected 1s,1h got %v,%v", res.Get(0), res.Get(1))
	}
}

// Two non-nullable operands: the left side always wins and no mask is built.
func Test_Coalesce_NoNulls(t *testing.T) {
	s := NewSeriesInt64([]int64{1, 2, 3}, nil, true, ctx)
	res := s.Coalesce(NewSeriesInt64([]int64{7, 8, 9}, nil, true, ctx))
	if res.IsError() {
		t.Fatal(res.GetError())
	}
	if res.IsNullable() {
		t.Fatal("no operand has nulls, the result must not be nullable")
	}
	for i, w := range []int64{1, 2, 3} {
		if res.Get(i).(int64) != w {
			t.Fatalf("value[%d]: expected %v, got %v", i, w, res.Get(i))
		}
	}
}

func Test_Coalesce_Errors(t *testing.T) {
	// Length mismatch between two vectors.
	s := NewSeriesFloat64([]float64{1, 2, 3, 4, 5}, nil, true, ctx)
	res := s.Coalesce(NewSeriesFloat64([]float64{1, 2, 3, 4}, nil, true, ctx))
	if !res.IsError() {
		t.Fatal("length mismatch must be an error")
	}
	if !strings.Contains(res.GetError(), "coalesce") {
		t.Fatalf("error should mention coalesce, got %q", res.GetError())
	}

	// A pair with no defined result type.
	res = NewSeriesBool([]bool{true}, nil, true, ctx).Coalesce(NewSeriesInt([]int{1}, nil, true, ctx))
	if !res.IsError() {
		t.Fatal("Bools coalesce Ints must be an error")
	}

	// An errored series keeps its error.
	res = NewSeriesError("boom").Coalesce(1.0)
	if !res.IsError() || res.GetError() != "boom" {
		t.Fatalf("expected the original error, got %v", res.GetError())
	}
}
