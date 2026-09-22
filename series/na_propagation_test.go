package series

import "testing"

// Every binary operator propagates NA from either side: a valid pairing
// with an NA operand yields an NAs result of the right length. Coalesce
// is the deliberate exception.
func Test_NAOperand_YieldsNA(t *testing.T) {
	na := NewSeriesNA(2, ctx)
	f := NewSeriesFloat64([]float64{1, 2}, nil, true, ctx)
	i := NewSeriesInt([]int{1, 2}, nil, true, ctx)
	i64 := NewSeriesInt64([]int64{1, 2}, nil, true, ctx)
	b := NewSeriesBool([]bool{true, false}, nil, true, ctx)
	st := NewSeriesString([]string{"a", "b"}, nil, true, ctx)

	cases := []struct {
		name string
		res  Series
	}{
		{"f + na", f.Add(na)},
		{"na + f", na.Add(f)},
		{"f - na", f.Sub(na)},
		{"na - f", na.Sub(f)},
		{"f * na", f.Mul(na)},
		{"f / na", f.Div(na)},
		{"i %% na", i.Mod(na)},
		{"i64 ^ na", i64.Exp(na)},
		{"i == na", i.Eq(na)},
		{"na != i", na.Ne(i)},
		{"f < na", f.Lt(na)},
		{"f <= na", f.Le(na)},
		{"f > na", f.Gt(na)},
		{"na >= f", na.Ge(f)},
		{"b | na", b.Or(na)},
		{"na | b", na.Or(b)},
		{"b & na", b.And(na)},
		{"s + na", st.Add(na)},
		{"na + s", na.Add(st)},
		{"s == na", st.Eq(na)},
	}
	for _, c := range cases {
		res, ok := c.res.(NAs)
		if !ok {
			t.Errorf("%s: expected NAs, got %T (err %v)", c.name, c.res, c.res.Err())
			continue
		}
		if res.Len() != 2 {
			t.Errorf("%s: expected length 2, got %d", c.name, res.Len())
		}
	}

	// Coalesce replaces nulls, so it keeps the typed result.
	if _, ok := f.Coalesce(na).(Float64s); !ok {
		t.Errorf("f ?? na: expected Float64s, got %T", f.Coalesce(na))
	}

	// An invalid pairing stays an error from either side.
	if i.And(na).Err() == nil {
		t.Error("i & na: expected an error, AND is not defined for Ints")
	}
	if na.And(i).Err() == nil {
		t.Error("na & i: expected an error, AND is not defined for Ints")
	}
}
