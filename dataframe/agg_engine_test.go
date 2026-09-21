package dataframe

import (
	"math"
	"runtime"
	"testing"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/series"
)

func f64(t *testing.T, df DataFrame, col string) []float64 {
	s, ok := df.Col(col).(series.Float64s)
	if !ok {
		t.Fatalf("col %q not Float64s", col)
	}
	return s.Float64s()
}

func TestAggregateSerialSingleKey(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"b", "a", "b", "a", "b"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 10, 2, 20, 3}, nil, false, ctx))

	out := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Sum("v"), Count()}, true)
	// sorted by key: a then b
	if got := out.Col("g").(series.Strings).GetAsString(0); got != "a" {
		t.Fatalf("row0 key = %q, want a (sorted)", got)
	}
	sum := f64(t, out, "sum(v)")
	if sum[0] != 30 || sum[1] != 6 { // a: 10+20, b: 1+2+3
		t.Fatalf("sum = %v, want [30 6]", sum)
	}
	n := out.Col("n").(series.Int64s).Int64s()
	if n[0] != 2 || n[1] != 3 {
		t.Fatalf("count = %v, want [2 3]", n)
	}
}

func TestAggregateSerialSkipNullsDefault(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a", "a", "a"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 0, 3}, []bool{false, true, false}, false, ctx))

	// removeNAs=true (new default): mean of {1,3} = 2
	out := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Mean("v")}, true)
	if got := f64(t, out, "mean(v)")[0]; got != 2 {
		t.Fatalf("skip-null mean = %v, want 2", got)
	}
	// removeNAs=false: NA-propagated -> NaN
	out2 := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Mean("v")}, false)
	if got := f64(t, out2, "mean(v)")[0]; !math.IsNaN(got) {
		t.Fatalf("NA-propagated mean = %v, want NaN", got)
	}
}

// TestAggregateSerialSkipNullsDefaultInt64 is the Int64 analogue of
// TestAggregateSerialSkipNullsDefault: it exercises the copy-free
// aggValueView I64 read path (accumulateChunk reading an Int64 value column
// inline instead of a []float64 copy) under both the skip-null default and
// RemoveNAs(false) NA propagation.
func TestAggregateSerialSkipNullsDefaultInt64(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a", "a", "a"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesInt64([]int64{1, 0, 3}, []bool{false, true, false}, false, ctx))

	// removeNAs=true (new default): mean of {1,3} = 2
	out := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Mean("v")}, true)
	if got := f64(t, out, "mean(v)")[0]; got != 2 {
		t.Fatalf("skip-null mean = %v, want 2", got)
	}
	// removeNAs=false: NA-propagated -> NaN
	out2 := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Mean("v")}, false)
	if got := f64(t, out2, "mean(v)")[0]; !math.IsNaN(got) {
		t.Fatalf("NA-propagated mean = %v, want NaN", got)
	}
}

func TestAggregateSerialMedianMultiKey(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a", "a", "a", "a"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 2, 3, 4}, nil, false, ctx))
	out := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Median("v"), Quantile("v", 0.25, WithInterpolation(Higher))}, true)
	if got := f64(t, out, "median(v)")[0]; got != 2.5 {
		t.Fatalf("median = %v, want 2.5", got)
	}
	if got := f64(t, out, "quantile_0.25(v)")[0]; got != 2 {
		t.Fatalf("q0.25 higher = %v, want 2", got)
	}
}

// Under RemoveNAs(false) a null in a group must propagate an NA to EVERY aggregate,
// holistic (Median) as well as reducible (Mean).
func TestAggregateSerialNAPropagationAllAggregates(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a", "a", "a"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 0, 3}, []bool{false, true, false}, false, ctx))

	out := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Mean("v"), Median("v")}, false)
	if got := f64(t, out, "mean(v)")[0]; !math.IsNaN(got) {
		t.Fatalf("NA-propagated mean = %v, want NaN", got)
	}
	if got := f64(t, out, "median(v)")[0]; !math.IsNaN(got) {
		t.Fatalf("NA-propagated median = %v, want NaN", got)
	}
}

// A nullable key column: the null-key group sorts LAST (after every non-null
// group), which come out ascending.
func TestAggregateSerialNullKeysSortLast(t *testing.T) {
	ctx := enchanter.NewContext()
	// keys: b, <null>, a, <null>, b  ->  groups b, <null>, a
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"b", "", "a", "", "b"}, []bool{false, true, false, true, false}, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 2, 3, 4, 5}, nil, false, ctx))

	out := aggregateSerial(df, []series.Series{df.Col("g")}, []aggregator{Sum("v")}, true)
	g := out.Col("g").(series.Strings)
	if g.IsNull(0) || g.GetAsString(0) != "a" {
		t.Fatalf("row0 = %q null=%v, want a", g.GetAsString(0), g.IsNull(0))
	}
	if g.IsNull(1) || g.GetAsString(1) != "b" {
		t.Fatalf("row1 = %q null=%v, want b", g.GetAsString(1), g.IsNull(1))
	}
	if !g.IsNull(2) {
		t.Fatalf("row2 null=%v, want null last", g.IsNull(2))
	}
	// sums follow the sorted rows: a=3, b=1+5=6, null=2+4=6
	sum := f64(t, out, "sum(v)")
	if sum[0] != 3 || sum[1] != 6 || sum[2] != 6 {
		t.Fatalf("sum = %v, want [3 6 6]", sum)
	}
}

// Two key columns: result rows are ordered lexicographically by (k1, k2) with
// k1 primary; rows sharing k1 but differing on k2 exercise the tie-break.
func TestAggregateSerialTwoKeyLexicographic(t *testing.T) {
	ctx := enchanter.NewContext()
	// (k1,k2): (b,y),(a,y),(b,x),(a,x) -> 4 distinct groups
	df := NewDataFrame(ctx).
		AddSeries("k1", series.NewSeriesString([]string{"b", "a", "b", "a"}, nil, false, ctx)).
		AddSeries("k2", series.NewSeriesString([]string{"y", "y", "x", "x"}, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64([]float64{1, 2, 3, 4}, nil, false, ctx))

	out := aggregateSerial(df, []series.Series{df.Col("k1"), df.Col("k2")}, []aggregator{Sum("v")}, true)
	k1 := out.Col("k1").(series.Strings)
	k2 := out.Col("k2").(series.Strings)
	// lexicographic: (a,x),(a,y),(b,x),(b,y)
	wantK1 := []string{"a", "a", "b", "b"}
	wantK2 := []string{"x", "y", "x", "y"}
	for i := range wantK1 {
		if k1.GetAsString(i) != wantK1[i] || k2.GetAsString(i) != wantK2[i] {
			t.Fatalf("row%d = (%q,%q), want (%q,%q)", i, k1.GetAsString(i), k2.GetAsString(i), wantK1[i], wantK2[i])
		}
	}
}

// TestAggregateParallelParity asserts that aggregate's parallel path (row
// count above aggMinParallel) produces results identical to aggregateSerial
// across reducible (Sum, Mean, Min, Max, Std) and non-value (Count)
// aggregators. Run with -race to confirm the worker chunks + merge carry no
// data races.
func TestAggregateParallelParity(t *testing.T) {
	ctx := enchanter.NewContext()
	n := 200_000
	keys := make([]string, n)
	vals := make([]float64, n)
	for i := range keys {
		keys[i] = []string{"a", "b", "c", "d"}[i%4]
		vals[i] = float64(i % 97)
	}
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString(keys, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64(vals, nil, false, ctx))

	aggs := []aggregator{Sum("v"), Mean("v"), Min("v"), Max("v"), Std("v"), Count()}
	ser := aggregateSerial(df, []series.Series{df.Col("g")}, aggs, true)
	par := aggregate(df, []series.Series{df.Col("g")}, aggs, true)

	for _, col := range []string{"sum(v)", "mean(v)", "min(v)", "max(v)", "std(v)"} {
		s, p := f64(t, ser, col), f64(t, par, col)
		if len(s) != len(p) {
			t.Fatalf("%s len mismatch", col)
		}
		for i := range s {
			if math.Abs(s[i]-p[i]) > 1e-9 {
				t.Fatalf("%s[%d] serial=%v parallel=%v", col, i, s[i], p[i])
			}
		}
	}
}

// TestAggregateParallelNAPropagationCrossChunk exercises the cross-chunk NA-propagation
// OR-merge (agg_engine.go's merge loop, which must OR a group's NA-propagation flag
// across every worker chunk rather than reset it) and the holistic
// (Median)/Count columns under the parallel path, neither of which
// TestAggregateParallelParity exercises (it only runs removeNAs=true, which
// never propagates anything).
//
// GOMAXPROCS is forced to 4 for the duration of the test so the null row and
// this group's many non-null rows are guaranteed to land in different worker
// chunks regardless of the host's actual CPU count: group "a" appears at
// every 4th row across the full 200,000-row range (50,000 rows total, spread
// evenly across all 4 forced chunks of 50,000 rows each), with a single null
// at row 0 (chunk 0) and ~49,999 non-null rows for "a" spread across every
// chunk, including chunks 1-3 which see no NA propagation locally at all.
func TestAggregateParallelNAPropagationCrossChunk(t *testing.T) {
	ctx := enchanter.NewContext()
	n := 200_000
	keys := make([]string, n)
	vals := make([]float64, n)
	nulls := make([]bool, n)
	for i := range keys {
		keys[i] = []string{"a", "b", "c", "d"}[i%4]
		vals[i] = float64(i % 97)
	}
	nulls[0] = true // row 0 is key "a"; this is the only null in the frame

	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString(keys, nil, false, ctx)).
		AddSeries("v", series.NewSeriesFloat64(vals, nulls, false, ctx))

	oldProcs := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(oldProcs)

	aggs := []aggregator{Sum("v"), Mean("v"), Median("v"), Count()}
	ser := aggregateSerial(df, []series.Series{df.Col("g")}, aggs, false)
	par := aggregate(df, []series.Series{df.Col("g")}, aggs, false)

	if ser.NRows() != par.NRows() {
		t.Fatalf("row count mismatch: serial=%d parallel=%d", ser.NRows(), par.NRows())
	}

	gcolPar := par.Col("g").(series.Strings)
	aIdx := -1
	for i := 0; i < par.NRows(); i++ {
		if gcolPar.GetAsString(i) == "a" {
			aIdx = i
			break
		}
	}
	if aIdx == -1 {
		t.Fatalf("group 'a' not found in parallel result")
	}

	// The NA-propagated group's reducible (Sum, Mean) and holistic (Median)
	// aggregates must all be non-null NaN, matching the serial engine's
	// NA-propagation-overrides-everything semantics.
	for _, col := range []string{"sum(v)", "mean(v)", "median(v)"} {
		s := par.Col(col).(series.Float64s)
		if s.IsNull(aIdx) {
			t.Fatalf("parallel %s[a] isNull = true, want non-null NaN (NA propagation)", col)
		}
		if !math.IsNaN(s.Float64s()[aIdx]) {
			t.Fatalf("parallel %s[a] = %v, want NaN (NA-propagated)", col, s.Float64s()[aIdx])
		}
	}

	// Count is unaffected by value nulls: it counts rows, not values.
	wantCount := int64(n / 4)
	cnt := par.Col("n").(series.Int64s)
	if cnt.Int64s()[aIdx] != wantCount {
		t.Fatalf("parallel count[a] = %d, want %d", cnt.Int64s()[aIdx], wantCount)
	}

	// Parallel must match serial exactly for the SAME inputs: same keys in
	// the same order, and every aggregate column equal (NaN-aware for the
	// float columns; exact for Count).
	gcolSer := ser.Col("g").(series.Strings)
	for i := 0; i < ser.NRows(); i++ {
		if gcolSer.GetAsString(i) != gcolPar.GetAsString(i) {
			t.Fatalf("key[%d] serial=%q parallel=%q", i, gcolSer.GetAsString(i), gcolPar.GetAsString(i))
		}
	}
	for _, col := range []string{"sum(v)", "mean(v)", "median(v)"} {
		sSer := ser.Col(col).(series.Float64s)
		sPar := par.Col(col).(series.Float64s)
		for i := 0; i < ser.NRows(); i++ {
			if sSer.IsNull(i) != sPar.IsNull(i) {
				t.Fatalf("%s[%d] null mask mismatch: serial=%v parallel=%v", col, i, sSer.IsNull(i), sPar.IsNull(i))
			}
			sv, pv := sSer.Float64s()[i], sPar.Float64s()[i]
			// NaN-aware equality: NaN != NaN in Go, but NA-propagated cells on
			// both sides must both be NaN for the rows to agree.
			if math.IsNaN(sv) || math.IsNaN(pv) {
				if !math.IsNaN(sv) || !math.IsNaN(pv) {
					t.Fatalf("%s[%d] NaN mismatch: serial=%v parallel=%v", col, i, sv, pv)
				}
				continue
			}
			if math.Abs(sv-pv) > 1e-9 {
				t.Fatalf("%s[%d] serial=%v parallel=%v", col, i, sv, pv)
			}
		}
	}
	nSer := ser.Col("n").(series.Int64s)
	nPar := par.Col("n").(series.Int64s)
	for i := 0; i < ser.NRows(); i++ {
		if nSer.Int64s()[i] != nPar.Int64s()[i] {
			t.Fatalf("n[%d] serial=%d parallel=%d", i, nSer.Int64s()[i], nPar.Int64s()[i])
		}
	}
}

// TestAggregateParallelInt64ValueNAPropagationCrossChunk is the Int64 analogue of
// TestAggregateParallelNAPropagationCrossChunk: it exercises the copy-free
// aggValueView I64 read path (accumulateChunk reading an Int64 value column
// inline, converting to float64 per row instead of pre-copying to
// []float64) across the parallel engine's cross-chunk NA-propagation OR-merge, and
// checks serial/parallel parity for the same inputs.
func TestAggregateParallelInt64ValueNAPropagationCrossChunk(t *testing.T) {
	ctx := enchanter.NewContext()
	n := 200_000
	keys := make([]string, n)
	vals := make([]int64, n)
	nulls := make([]bool, n)
	for i := range keys {
		keys[i] = []string{"a", "b", "c", "d"}[i%4]
		vals[i] = int64(i % 97)
	}
	nulls[0] = true // row 0 is key "a"; this is the only null in the frame

	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString(keys, nil, false, ctx)).
		AddSeries("v", series.NewSeriesInt64(vals, nulls, false, ctx))

	oldProcs := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(oldProcs)

	aggs := []aggregator{Sum("v"), Mean("v"), Median("v"), Count()}
	ser := aggregateSerial(df, []series.Series{df.Col("g")}, aggs, false)
	par := aggregate(df, []series.Series{df.Col("g")}, aggs, false)

	if ser.NRows() != par.NRows() {
		t.Fatalf("row count mismatch: serial=%d parallel=%d", ser.NRows(), par.NRows())
	}

	gcolPar := par.Col("g").(series.Strings)
	aIdx := -1
	for i := 0; i < par.NRows(); i++ {
		if gcolPar.GetAsString(i) == "a" {
			aIdx = i
			break
		}
	}
	if aIdx == -1 {
		t.Fatalf("group 'a' not found in parallel result")
	}

	// The NA-propagated group's reducible (Sum, Mean) and holistic (Median)
	// aggregates must all be non-null NaN, matching the serial engine's
	// NA-propagation-overrides-everything semantics.
	for _, col := range []string{"sum(v)", "mean(v)", "median(v)"} {
		s := par.Col(col).(series.Float64s)
		if s.IsNull(aIdx) {
			t.Fatalf("parallel %s[a] isNull = true, want non-null NaN (NA propagation)", col)
		}
		if !math.IsNaN(s.Float64s()[aIdx]) {
			t.Fatalf("parallel %s[a] = %v, want NaN (NA-propagated)", col, s.Float64s()[aIdx])
		}
	}

	// Count is unaffected by value nulls: it counts rows, not values.
	wantCount := int64(n / 4)
	cnt := par.Col("n").(series.Int64s)
	if cnt.Int64s()[aIdx] != wantCount {
		t.Fatalf("parallel count[a] = %d, want %d", cnt.Int64s()[aIdx], wantCount)
	}

	// Parallel must match serial exactly for the SAME inputs.
	gcolSer := ser.Col("g").(series.Strings)
	for i := 0; i < ser.NRows(); i++ {
		if gcolSer.GetAsString(i) != gcolPar.GetAsString(i) {
			t.Fatalf("key[%d] serial=%q parallel=%q", i, gcolSer.GetAsString(i), gcolPar.GetAsString(i))
		}
	}
	for _, col := range []string{"sum(v)", "mean(v)", "median(v)"} {
		sSer := ser.Col(col).(series.Float64s)
		sPar := par.Col(col).(series.Float64s)
		for i := 0; i < ser.NRows(); i++ {
			if sSer.IsNull(i) != sPar.IsNull(i) {
				t.Fatalf("%s[%d] null mask mismatch: serial=%v parallel=%v", col, i, sSer.IsNull(i), sPar.IsNull(i))
			}
			sv, pv := sSer.Float64s()[i], sPar.Float64s()[i]
			if math.IsNaN(sv) || math.IsNaN(pv) {
				if !math.IsNaN(sv) || !math.IsNaN(pv) {
					t.Fatalf("%s[%d] NaN mismatch: serial=%v parallel=%v", col, i, sv, pv)
				}
				continue
			}
			if math.Abs(sv-pv) > 1e-9 {
				t.Fatalf("%s[%d] serial=%v parallel=%v", col, i, sv, pv)
			}
		}
	}
	nSer := ser.Col("n").(series.Int64s)
	nPar := par.Col("n").(series.Int64s)
	for i := 0; i < ser.NRows(); i++ {
		if nSer.Int64s()[i] != nPar.Int64s()[i] {
			t.Fatalf("n[%d] serial=%d parallel=%d", i, nSer.Int64s()[i], nPar.Int64s()[i])
		}
	}
}

func TestAggregateAnyAll(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a", "a", "b", "b"}, nil, false, ctx)).
		AddSeries("b", series.NewSeriesBool([]bool{true, false, true, true}, nil, false, ctx))
	out := aggregate(df, []series.Series{df.Col("g")}, []aggregator{Any("b"), All("b")}, true)
	anyC := out.Col("any(b)").(series.Bools) // sorted: a, b
	allC := out.Col("all(b)").(series.Bools)
	if anyC.Get(0) != true || allC.Get(0) != false || anyC.Get(1) != true || allC.Get(1) != true {
		t.Fatalf("any/all wrong: any=%v all=%v", anyC.Bools(), allC.Bools())
	}
}

func TestAggregateAnyAllNAPropagationNull(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString([]string{"a", "a", "b"}, nil, false, ctx)).
		AddSeries("b", series.NewSeriesBool([]bool{true, false, true}, []bool{false, true, false}, false, ctx))
	out := aggregate(df, []series.Series{df.Col("g")}, []aggregator{All("b")}, false)
	allC := out.Col("all(b)").(series.Bools)
	if !allC.IsNull(0) {
		t.Fatalf("NA-propagated group a: all(b) should be null")
	}
	if allC.IsNull(1) || allC.Get(1) != true {
		t.Fatalf("group b: all(b) should be true")
	}
}

// TestAggregateParallelAnyAllCrossChunk exercises the cross-chunk OR-merge of
// reducibleState.bflag for AGGREGATE_ANY/AGGREGATE_ALL (agg_engine.go's
// mergeReducible ANY/ALL case, and the merge-time growReducible call), which
// TestAggregateAnyAll/TestAggregateAnyAllNAPropagationNull never exercise: those use
// 3-4 rows, always below aggMinParallel, so aggregate() takes the
// single-chunk inline branch and never calls mergeReducible at all.
//
// GOMAXPROCS is forced to 4 so the 200_000 rows split into exactly 4 worker
// chunks of 50_000 rows each (regardless of host CPU count); every group
// ("a".."d", by i%4) appears at every 4th row, spread evenly across all 4
// chunks. Two "flip" rows are placed in specific chunks so the flipped value
// can only reach the merged result via the cross-chunk bflag OR-merge, not
// from any single chunk's local state alone:
//   - group "a" is true everywhere except one row in chunk 3 (the LAST
//     chunk) -> all(a) must be false, which only the OR-merge of chunk 3's
//     bflag into the other chunks' (locally all-true) state can produce.
//   - group "b" is false everywhere except one row in chunk 0 (the FIRST
//     chunk) -> any(b) must be true, symmetric to the above.
//
// A second phase re-runs the same shape under RemoveNAs(false) with a single
// null value for group "a" in chunk 2, verifying the NA-propagation flag OR-merges
// across chunks for Any/All exactly as it already does for Sum/Mean/Median
// (see TestAggregateParallelNAPropagationCrossChunk).
func TestAggregateParallelAnyAllCrossChunk(t *testing.T) {
	ctx := enchanter.NewContext()
	n := 200_000
	keys := make([]string, n)
	vals := make([]bool, n)
	for i := range keys {
		keys[i] = []string{"a", "b", "c", "d"}[i%4]
		switch i % 4 {
		case 0: // group "a": true everywhere except one late-chunk row below
			vals[i] = true
		case 1: // group "b": false everywhere except one early-chunk row below
			vals[i] = false
		case 2: // group "c": true everywhere (control: any=all=true)
			vals[i] = true
		case 3: // group "d": false everywhere (control: any=all=false)
			vals[i] = false
		}
	}
	vals[150000] = false // group "a", chunk 3 (150000 % 4 == 0): the lone false
	vals[1] = true       // group "b", chunk 0 (1 % 4 == 1): the lone true

	df := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString(keys, nil, false, ctx)).
		AddSeries("b", series.NewSeriesBool(vals, nil, false, ctx))

	oldProcs := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(oldProcs)

	aggs := []aggregator{Any("b"), All("b")}
	ser := aggregateSerial(df, []series.Series{df.Col("g")}, aggs, true)
	par := aggregate(df, []series.Series{df.Col("g")}, aggs, true)

	if ser.NRows() != par.NRows() || par.NRows() != 4 {
		t.Fatalf("row count mismatch: serial=%d parallel=%d, want 4", ser.NRows(), par.NRows())
	}

	findIdx := func(d DataFrame, key string) int {
		gcol := d.Col("g").(series.Strings)
		for i := 0; i < d.NRows(); i++ {
			if gcol.GetAsString(i) == key {
				return i
			}
		}
		return -1
	}

	cases := []struct {
		key     string
		wantAny bool
		wantAll bool
	}{
		{"a", true, false}, // lone chunk-3 false must survive the OR-merge into all(a)=false
		{"b", true, false}, // lone chunk-0 true must survive the OR-merge into any(b)=true
		{"c", true, true},
		{"d", false, false},
	}
	anyC := par.Col("any(b)").(series.Bools)
	allC := par.Col("all(b)").(series.Bools)
	for _, c := range cases {
		idx := findIdx(par, c.key)
		if idx == -1 {
			t.Fatalf("group %q not found in parallel result", c.key)
		}
		if anyC.IsNull(idx) || allC.IsNull(idx) {
			t.Fatalf("group %q: any/all unexpectedly null", c.key)
		}
		if got := anyC.Get(idx); got != c.wantAny {
			t.Fatalf("parallel any(b)[%s] = %v, want %v", c.key, got, c.wantAny)
		}
		if got := allC.Get(idx); got != c.wantAll {
			t.Fatalf("parallel all(b)[%s] = %v, want %v", c.key, got, c.wantAll)
		}
	}

	// Full parity: parallel must match serial exactly, value and null mask,
	// for every group (not just the four spot-checked above).
	gcolSer := ser.Col("g").(series.Strings)
	gcolPar := par.Col("g").(series.Strings)
	anySer := ser.Col("any(b)").(series.Bools)
	allSer := ser.Col("all(b)").(series.Bools)
	anyPar := par.Col("any(b)").(series.Bools)
	allPar := par.Col("all(b)").(series.Bools)
	for i := 0; i < ser.NRows(); i++ {
		if gcolSer.GetAsString(i) != gcolPar.GetAsString(i) {
			t.Fatalf("key[%d] serial=%q parallel=%q", i, gcolSer.GetAsString(i), gcolPar.GetAsString(i))
		}
		if anySer.IsNull(i) != anyPar.IsNull(i) || anySer.Get(i) != anyPar.Get(i) {
			t.Fatalf("any(b)[%d] serial=(%v,null=%v) parallel=(%v,null=%v)", i, anySer.Get(i), anySer.IsNull(i), anyPar.Get(i), anyPar.IsNull(i))
		}
		if allSer.IsNull(i) != allPar.IsNull(i) || allSer.Get(i) != allPar.Get(i) {
			t.Fatalf("all(b)[%d] serial=(%v,null=%v) parallel=(%v,null=%v)", i, allSer.Get(i), allSer.IsNull(i), allPar.Get(i), allPar.IsNull(i))
		}
	}

	// NA propagation across chunks (RemoveNAs(false)): a single null value for group
	// "a" placed in chunk 2 must propagate an NA to any(b)/all(b) as NULL for that
	// group, in both engines — the parallel merge OR's the NA-propagation flag
	// exactly as it OR's bflag above.
	nulls := make([]bool, n)
	nulls[100000] = true // group "a" (100000 % 4 == 0), chunk 2
	dfNAProp := NewDataFrame(ctx).
		AddSeries("g", series.NewSeriesString(keys, nil, false, ctx)).
		AddSeries("b", series.NewSeriesBool(vals, nulls, false, ctx))

	aggsNAProp := []aggregator{Any("b"), All("b")}
	serNA := aggregateSerial(dfNAProp, []series.Series{dfNAProp.Col("g")}, aggsNAProp, false)
	parNA := aggregate(dfNAProp, []series.Series{dfNAProp.Col("g")}, aggsNAProp, false)

	for _, d := range []DataFrame{serNA, parNA} {
		idx := findIdx(d, "a")
		if idx == -1 {
			t.Fatalf("group \"a\" not found in NA-propagated result")
		}
		anyD := d.Col("any(b)").(series.Bools)
		allD := d.Col("all(b)").(series.Bools)
		if !anyD.IsNull(idx) || !allD.IsNull(idx) {
			t.Fatalf("NA-propagated group \"a\": any(b)/all(b) should be null, got any.IsNull=%v all.IsNull=%v", anyD.IsNull(idx), allD.IsNull(idx))
		}
	}

	// Groups without an NA must be unaffected by the other group's NA propagation.
	for _, key := range []string{"b", "c", "d"} {
		for _, d := range []DataFrame{serNA, parNA} {
			idx := findIdx(d, key)
			if idx == -1 {
				t.Fatalf("group %q not found in NA-propagated result", key)
			}
			anyD := d.Col("any(b)").(series.Bools)
			allD := d.Col("all(b)").(series.Bools)
			if anyD.IsNull(idx) || allD.IsNull(idx) {
				t.Fatalf("group %q unexpectedly null after NA-propagating group \"a\"", key)
			}
		}
	}
}
