package dataframe

import (
	"testing"

	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/series"
)

// ToArrowRecord's contract is that the caller Releases the record; under a
// CheckedAllocator that must balance to zero outstanding bytes.
func TestArrowRecordRoundTripNoLeaks(t *testing.T) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	ctx := enchanter.NewContext()
	ctx.Allocator = mem
	defer mem.AssertSize(t, 0)

	df := NewDataFrame(ctx).
		AddSeries("a", series.NewSeriesFloat64([]float64{1, 2, 3}, []bool{false, true, false}, false, ctx)).
		AddSeries("b", series.NewSeriesString([]string{"x", "y", "z"}, nil, false, ctx))
	if df.IsErrored() {
		t.Fatal(df.GetError())
	}

	rec := df.ToArrowRecord()
	df2 := NewDataFrameFromArrowRecord(rec, ctx)
	rec.Release()
	if df2.IsErrored() {
		t.Fatal(df2.GetError())
	}
	if df2.NRows() != 3 || df2.NCols() != 2 {
		t.Fatalf("round trip shape: got %dx%d, want 3x2", df2.NRows(), df2.NCols())
	}
}
