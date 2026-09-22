package series

import (
	"testing"
)

// Not is callable through the Series interface on every type. Bools
// negates, NAs stays NAs, everything else returns an error series.
func Test_Not_EveryType(t *testing.T) {
	b := NewSeriesBool([]bool{true, false, true}, nil, true, ctx)
	res := b.Not()
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	want := []bool{false, true, false}
	for i, w := range want {
		if res.Get(i).(bool) != w {
			t.Fatalf("bools not [%d]: expected %v, got %v", i, w, res.Get(i))
		}
	}

	na := NewSeriesNA(3, ctx).Not()
	if na.Err() != nil || na.Len() != 3 {
		t.Fatalf("NAs not: expected NAs of length 3, got %v (err %v)", na, na.Err())
	}

	unsupported := []Series{
		NewSeriesInt([]int{1}, nil, true, ctx),
		NewSeriesInt64([]int64{1}, nil, true, ctx),
		NewSeriesFloat64([]float64{1}, nil, true, ctx),
		NewSeriesString([]string{"a"}, nil, true, ctx),
	}
	for _, s := range unsupported {
		if s.Not().Err() == nil {
			t.Fatalf("%s.Not(): expected an error series", s.Type().String())
		}
	}

	e := NewSeriesError("boom").Not()
	if e.Err() == nil || e.Err().Error() != "boom" {
		t.Fatalf("Errors.Not(): expected the original error, got %v", e.Err())
	}
}
