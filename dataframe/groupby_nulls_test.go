package dataframe

import (
	"testing"
	"time"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/series"
)

// Null keys group together, and the resulting key column must keep the null
// flag instead of surfacing the zero value as valid data.
func TestGroupByNullKeyKeepsNullInResult(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("k", series.NewSeriesInt64([]int64{1, 1, 2, 0, 2}, []bool{false, false, false, true, false}, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{10, 20, 30, 40, 50}, nil, false, ctx))
	if df.Err() != nil {
		t.Fatal(df.Err())
	}

	res := df.GroupBy("k").Agg(Count()).Run()
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 3 {
		t.Fatalf("expected 3 groups (1, 2, null), got %d", res.NRows())
	}

	k := res.C("k")
	nullCount := 0
	for i := 0; i < k.Len(); i++ {
		if k.IsNull(i) {
			nullCount++
		}
	}
	if nullCount != 1 {
		t.Fatalf("expected exactly one null group key, got %d (null keys must stay null in the result)", nullCount)
	}
}

// GroupBy on an already grouped dataframe must replace the grouping, not
// silently keep the old one.
func TestGroupByOnGroupedRegroups(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("a", series.NewSeriesInt64([]int64{1, 1, 2, 2}, nil, false, ctx)).
		AddSeries("b", series.NewSeriesString([]string{"x", "y", "x", "y"}, nil, false, ctx))
	if df.Err() != nil {
		t.Fatal(df.Err())
	}

	res := df.GroupBy("a").GroupBy("b").Agg(Count()).Run()
	if res.Err() != nil {
		t.Fatal(res.Err())
	}

	foundB := false
	for _, n := range res.Names() {
		if n == "b" {
			foundB = true
		}
	}
	if !foundB {
		t.Fatalf("expected the regrouped key column %q in result, got %v", "b", res.Names())
	}
}

// Grouping by a Times key must produce an aligned result frame (the key
// column used to be silently dropped, leaving names and series misaligned).
func TestGroupByTimeKeyProducesAlignedResult(t *testing.T) {
	ctx := enchanter.NewContext()
	t0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)
	df := NewDataFrame(ctx).
		AddSeries("k", series.NewSeriesTime([]time.Time{t0, t0, t1}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 2, 3}, nil, false, ctx))
	if df.Err() != nil {
		t.Fatal(df.Err())
	}

	res := df.GroupBy("k").Agg(Count()).Run()
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 2 {
		t.Fatalf("expected 2 groups, got %d", res.NRows())
	}
	k := res.C("k")
	if k.Err() != nil {
		t.Fatalf("key column missing from result: %s", k.Err())
	}
	if k.Len() != 2 {
		t.Fatalf("key column length %d, want 2 (result frame must stay aligned)", k.Len())
	}
	if k.Type() != meta.TimeType {
		t.Fatalf("key column type %v, want %v (a dropped key series shifts the columns)", k.Type(), meta.TimeType)
	}
}
