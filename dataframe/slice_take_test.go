package dataframe

import (
	"testing"

	"github.com/caerbannogwhite/enchanter"
)

func sliceFixture(t *testing.T) DataFrame {
	t.Helper()
	ctx := enchanter.NewContext()
	return NewDataFrame(ctx).
		AddSeriesFromInt64s("n", []int64{10, 20, 30, 40, 50}, nil, false).
		AddSeriesFromStrings("s", []string{"a", "b", "c", "d", "e"}, nil, false)
}

func TestDataFrame_Slice(t *testing.T) {
	df := sliceFixture(t)

	res := df.Slice(1, 4)
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 3 {
		t.Fatalf("rows: expected 3, got %d", res.NRows())
	}
	for i, w := range []int64{20, 30, 40} {
		if got := res.Col("n").Get(i).(int64); got != w {
			t.Errorf("n[%d]: expected %d, got %d", i, w, got)
		}
	}

	if got := df.Slice(2, 2).NRows(); got != 0 {
		t.Errorf("empty interval: expected 0 rows, got %d", got)
	}

	for _, bad := range [][2]int{{-1, 2}, {3, 2}, {2, 6}} {
		if df.Slice(bad[0], bad[1]).Err() == nil {
			t.Errorf("Slice(%d, %d): expected an error", bad[0], bad[1])
		}
	}

	// An errored frame passes its error through.
	broken := df.Select("missing")
	if broken.Slice(0, 1).Err() == nil {
		t.Error("an errored frame must keep its error through Slice")
	}
}

func TestDataFrame_TakeIndices(t *testing.T) {
	df := sliceFixture(t)

	res := df.TakeIndices([]int{4, 0, 0})
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 3 {
		t.Fatalf("rows: expected 3, got %d", res.NRows())
	}
	for i, w := range []string{"e", "a", "a"} {
		if got := res.Col("s").Get(i).(string); got != w {
			t.Errorf("s[%d]: expected %q, got %q", i, w, got)
		}
	}

	if got := df.TakeIndices(nil).NRows(); got != 0 {
		t.Errorf("no indices: expected 0 rows, got %d", got)
	}

	if df.TakeIndices([]int{5}).Err() == nil {
		t.Error("an out-of-range index must be an error")
	}
	if df.TakeIndices([]int{-1}).Err() == nil {
		t.Error("a negative index must be an error")
	}
}

func TestDataFrame_SelectAt(t *testing.T) {
	df := sliceFixture(t)

	res := df.SelectAt(1, 0)
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	names := res.Names()
	if len(names) != 2 || names[0] != "s" || names[1] != "n" {
		t.Fatalf("names: expected [s n], got %v", names)
	}
	if got := res.Col("n").Get(2).(int64); got != 30 {
		t.Errorf("values must survive the reorder, got n[2]=%d", got)
	}

	if df.SelectAt(2).Err() == nil {
		t.Error("an out-of-range column index must be an error")
	}
	if df.SelectAt(-1).Err() == nil {
		t.Error("a negative column index must be an error")
	}
}
