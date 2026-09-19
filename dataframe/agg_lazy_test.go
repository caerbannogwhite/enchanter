package dataframe

import (
	"testing"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/series"
)

func TestGroupByIsLazyButJoinStillWorks(t *testing.T) {
	ctx := enchanter.NewContext()
	a := NewDataFrame(ctx).
		AddSeries("k", series.NewSeriesString([]string{"x", "y"}, nil, false, ctx)).
		AddSeries("va", series.NewSeriesInt64([]int64{1, 2}, nil, false, ctx))
	b := NewDataFrame(ctx).
		AddSeries("k", series.NewSeriesString([]string{"y", "z"}, nil, false, ctx)).
		AddSeries("vb", series.NewSeriesInt64([]int64{3, 4}, nil, false, ctx))

	// GroupBy no longer eagerly builds partitions
	g := a.GroupBy("k")
	if !g.isGrouped {
		t.Fatalf("expected isGrouped")
	}
	if g.partitions != nil {
		t.Fatalf("GroupBy should be lazy: partitions must be nil until materialized")
	}

	// Join still produces the inner match on k=y
	j := a.Join(INNER_JOIN, b, "k")
	if j.IsErrored() {
		t.Fatalf("join errored: %v", j.GetError())
	}
	if j.NRows() != 1 {
		t.Fatalf("inner join rows = %d, want 1", j.NRows())
	}
}
