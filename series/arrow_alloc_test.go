package series

import (
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/caerbannogwhite/enchanter"
)

// Leak test: run the Arrow interop round-trip under a CheckedAllocator and
// assert that every internal Arrow reference is released. Values returned to
// callers (Series.ArrowArray) are released explicitly here, per the documented
// ownership contract.

func newCheckedCtx() (*enchanter.Context, *memory.CheckedAllocator) {
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	ctx := enchanter.NewContext()
	ctx.Allocator = mem
	return ctx, mem
}

func TestArrowInteropNoLeaks(t *testing.T) {
	ctx, mem := newCheckedCtx()
	defer mem.AssertSize(t, 0)

	mask := []bool{false, true, false, false, true}

	roundTrip := func(s Series) {
		t.Helper()
		arr := s.ArrowArray()
		got := ArrowArrayToSeries(arr, ctx)
		arr.Release()
		if got.Err() != nil {
			t.Fatal(got.Err())
		}
	}

	roundTrip(NewSeriesFloat64([]float64{1, 2, 3, 4, 5}, mask, false, ctx))
	roundTrip(NewSeriesInt64([]int64{1, 2, 3, 4, 5}, mask, false, ctx))
	roundTrip(NewSeriesInt([]int{1, 2, 3, 4, 5}, mask, false, ctx))
	roundTrip(NewSeriesBool([]bool{true, false, true, false, true}, mask, false, ctx))
	roundTrip(NewSeriesString([]string{"a", "b", "c", "d", "e"}, mask, false, ctx))
	roundTrip(NewSeriesTime(make([]time.Time, 5), mask, false, ctx))
	roundTrip(NewSeriesDuration(make([]time.Duration, 5), mask, false, ctx))
}
