package dataframe

import (
	"math"
	"testing"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/series"
)

func TestAggRunSortedAndSkipNullByDefault(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"b", "a", "b"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 5, 0}, []bool{false, false, true}, false, ctx))

	out := df.GroupBy("g").Agg(Sum("v"), Std("v", WithDDoF(1))).Run()
	if out.Err() != nil {
		t.Fatalf("agg errored: %v", out.Err())
	}
	// sorted: a, b
	if out.Col("g").(series.Strings).GetAsString(0) != "a" {
		t.Fatalf("not sorted by key")
	}
	// b's null value skipped by default: sum(b) = 1, sample std of single value -> null
	sum := out.Col("sum(v)").(series.Float64s)
	if sum.Data_[1] != 1 {
		t.Fatalf("skip-null sum(b) = %v, want 1", sum.Data_[1])
	}
	std := out.Col("std(v)").(series.Float64s)
	if !std.IsNull(1) {
		t.Fatalf("sample std of single non-null should be null")
	}
}

// TestAggRunUnsupportedValueTypeErrors guards against a mistyped value
// column (e.g. Strings, which newAggValueView cannot accumulate) reaching
// accumulateChunk and panicking with an out-of-range slice index: Run must
// catch it up front and return an errored frame instead.
func TestAggRunUnsupportedValueTypeErrors(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a", "a", "b"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesString([]string{"x", "y", "z"}, nil, false, ctx))

	out := df.GroupBy("g").Agg(Sum("v")).Run()
	if out.Err() == nil {
		t.Fatalf("expected error for Sum over a Strings column, got none")
	}
	if out.Err() == nil {
		t.Fatalf("expected non-nil GetError() for Sum over a Strings column")
	}
}

// TestAggRunStdHandComputed builds a frame with three unequal-size groups
// (A: 2 rows, B: 3 rows, C: 1 row) and checks each group's Std() result
// against a value computed by hand (population std, ddof=0, the default) —
// an independent expectation, not re-derived from accumulator.go or
// agg_engine.go.
//
//	A: [2, 4]     mean=3   variance=((2-3)^2+(4-3)^2)/2 = 1        std=1
//	B: [1, 2, 3]  mean=2   variance=((1-2)^2+0+(3-2)^2)/3 = 2/3     std=sqrt(2/3)
//	C: [10]       mean=10  variance=0/1 = 0 (single point, ddof=0)  std=0
func TestAggRunStdHandComputed(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"A", "A", "B", "B", "B", "C"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{2, 4, 1, 2, 3, 10}, nil, false, ctx))

	out := df.GroupBy("g").Agg(Std("v")).Run()
	if out.Err() != nil {
		t.Fatalf("agg errored: %v", out.Err())
	}

	const eps = 1e-9
	want := map[string]float64{
		"A": 1.0,
		"B": math.Sqrt(2.0 / 3.0),
		"C": 0.0,
	}
	g := out.Col("g").(series.Strings)
	std := out.Col("std(v)").(series.Float64s)
	if out.NRows() != 3 {
		t.Fatalf("NRows = %d, want 3", out.NRows())
	}
	for i := 0; i < out.NRows(); i++ {
		key := g.GetAsString(i)
		w, ok := want[key]
		if !ok {
			t.Fatalf("unexpected group key %q", key)
		}
		if std.IsNull(i) {
			t.Fatalf("group %q: std(v) unexpectedly null", key)
		}
		if got := std.Data_[i]; math.Abs(got-w) > eps {
			t.Fatalf("group %q: std(v) = %v, want %v", key, got, w)
		}
	}
}

func TestAggRunOptionValidation(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1}, nil, false, ctx))
	out := df.GroupBy("g").Agg(Sum("v", WithDDoF(1))).Run() // ddof on Sum → error
	if out.Err() == nil {
		t.Fatalf("expected error for WithDDoF on Sum")
	}
	_ = math.Inf
}
