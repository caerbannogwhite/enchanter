// Package gospike is a THROWAWAY 0.4.0 spike (not part of the library).
//
// Question: how much of the groupby performance gap does a pragmatic pure-Go,
// single-pass, PARALLEL hash-aggregation close vs enchanter's current eager
// groupby? And (cameo) how much does operator fusion save on an element-wise
// op chain by eliminating the intermediate arrays?
//
// Method: measure three regimes for `sum(v1) group by id1` on the h2oai G1
// datasets (1e6, 1e7 rows; id1 has 100 distinct groups):
//
//	baseline  = df.GroupBy("id1").Agg(Sum("v1")).Run()   (current library path)
//	hashSingle= single-pass hash-agg, one goroutine
//	hashPar   = single-pass hash-agg, GOMAXPROCS goroutines + merge
//
// The key trick: enchanter interns strings (StringPool), so equal id1 values
// share one *string. We can hash by the POINTER (cheap) instead of the string
// bytes — exactly what the current partition path does not exploit.
//
// Run:  go test ./benchmarking/gospike/ -run TestSpikeCorrect -v
//
//	go test ./benchmarking/gospike/ -run ^$ -bench 'GroupBy.*1e6' -benchmem -benchtime=3x
//	go test ./benchmarking/gospike/ -run ^$ -bench 'GroupBy.*1e7' -benchmem -benchtime=3x
//	go test ./benchmarking/gospike/ -run ^$ -bench 'Chain'        -benchmem
package gospike

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/dataframe"
	"github.com/caerbannogwhite/enchanter/series"
)

// ---- lazy per-size G1 loader (avoids the dataframe pkg's full 8-frame TestMain) ----

var (
	loadMu sync.Mutex
	loaded = map[string]*dataframe.DataFrame{}
	pkgCtx *enchanter.Context
)

func readG1(ctx *enchanter.Context, name string) *dataframe.DataFrame {
	f, err := os.OpenFile(filepath.Join("..", "..", "testdata", name), os.O_RDONLY, 0666)
	if err != nil {
		return nil
	}
	defer f.Close()
	df := dataframe.ReadCsv(ctx).SetDelimiter(',').SetNullValues(false).SetReader(f).Read()
	return &df
}

func getG1(tb testing.TB, name string) dataframe.DataFrame {
	loadMu.Lock()
	defer loadMu.Unlock()
	if pkgCtx == nil {
		pkgCtx = enchanter.NewContext()
	}
	if d, ok := loaded[name]; ok {
		if d == nil {
			tb.Skipf("G1 data not available: %s", name)
		}
		return *d
	}
	d := readG1(pkgCtx, name)
	loaded[name] = d
	if d == nil || d.Err() != nil {
		tb.Skipf("G1 data not available: %s", name)
	}
	return *d
}

// cols extracts the raw interned key slice (id1) and value slice (v1).
func cols(df dataframe.DataFrame) (keys []*string, vals []int64) {
	keys = df.Col("id1").(series.Strings).Data_
	switch v := df.Col("v1").(type) {
	case series.Int64s:
		vals = v.Data_
	case series.Ints:
		vals = make([]int64, len(v.Data_))
		for i, x := range v.Data_ {
			vals[i] = int64(x)
		}
	}
	return keys, vals
}

// ---- the spike: single-pass hash aggregation ----

func hashSumSingle(keys []*string, vals []int64) map[*string]int64 {
	m := make(map[*string]int64, 128)
	for i, k := range keys {
		m[k] += vals[i]
	}
	return m
}

func hashSumParallel(keys []*string, vals []int64, workers int) map[*string]int64 {
	n := len(keys)
	if workers < 1 {
		workers = 1
	}
	chunk := (n + workers - 1) / workers
	partials := make([]map[*string]int64, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		lo := w * chunk
		hi := lo + chunk
		if hi > n {
			hi = n
		}
		wg.Add(1)
		go func(w, lo, hi int) {
			defer wg.Done()
			m := make(map[*string]int64, 128)
			for i := lo; i < hi; i++ {
				m[keys[i]] += vals[i]
			}
			partials[w] = m
		}(w, lo, hi)
	}
	wg.Wait()
	out := make(map[*string]int64, 128)
	for _, m := range partials {
		for k, v := range m {
			out[k] += v
		}
	}
	return out
}

// ---- correctness: parallel == single == direct total ----

func TestSpikeCorrect(t *testing.T) {
	df := getG1(t, "G1_1e6_1e2_0_0.csv")
	keys, vals := cols(df)
	if len(keys) == 0 || len(vals) != len(keys) {
		t.Fatalf("bad columns: keys=%d vals=%d", len(keys), len(vals))
	}
	var total int64
	for _, v := range vals {
		total += v
	}
	single := hashSumSingle(keys, vals)
	par := hashSumParallel(keys, vals, runtime.GOMAXPROCS(0))

	if len(single) != len(par) {
		t.Fatalf("group count: single=%d parallel=%d", len(single), len(par))
	}
	var st int64
	for k, v := range single {
		st += v
		if par[k] != v {
			t.Fatalf("group %v: single=%d parallel=%d", *k, v, par[k])
		}
	}
	if st != total {
		t.Fatalf("sum of group sums=%d, want direct total=%d", st, total)
	}
	t.Logf("OK: rows=%d groups=%d total=%d", len(keys), len(single), total)
}

// ---- groupby benchmarks: baseline vs single-pass vs parallel ----

func benchBaseline(b *testing.B, name string) {
	df := getG1(b, name)
	runtime.GC()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		df.GroupBy("id1").Agg(dataframe.Sum("v1")).Run()
	}
}

func benchHashSingle(b *testing.B, name string) {
	keys, vals := cols(getG1(b, name))
	runtime.GC()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = hashSumSingle(keys, vals)
	}
}

func benchHashPar(b *testing.B, name string) {
	keys, vals := cols(getG1(b, name))
	nw := runtime.GOMAXPROCS(0)
	runtime.GC()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = hashSumParallel(keys, vals, nw)
	}
}

func BenchmarkGroupBy_baseline_1e6(b *testing.B)   { benchBaseline(b, "G1_1e6_1e2_0_0.csv") }
func BenchmarkGroupBy_hashSingle_1e6(b *testing.B) { benchHashSingle(b, "G1_1e6_1e2_0_0.csv") }
func BenchmarkGroupBy_hashPar_1e6(b *testing.B)    { benchHashPar(b, "G1_1e6_1e2_0_0.csv") }

func BenchmarkGroupBy_baseline_1e7(b *testing.B)   { benchBaseline(b, "G1_1e7_1e2_0_0.csv") }
func BenchmarkGroupBy_hashSingle_1e7(b *testing.B) { benchHashSingle(b, "G1_1e7_1e2_0_0.csv") }
func BenchmarkGroupBy_hashPar_1e7(b *testing.B)    { benchHashPar(b, "G1_1e7_1e2_0_0.csv") }

// ---- fusion cameo: (a+b)>0 then filter, 3-pass library chain vs 1-pass fused ----

func makeF(n int) (a, c []float64) {
	a = make([]float64, n)
	c = make([]float64, n)
	for i := range a {
		a[i] = float64(i%1000) - 500
		c[i] = float64((i*7)%1000) - 500
	}
	return a, c
}

func fusedChain(a, c []float64) []float64 {
	out := make([]float64, 0, len(a)/2)
	for i := range a {
		s := a[i] + c[i]
		if s > 0 {
			out = append(out, s)
		}
	}
	return out
}

func BenchmarkChain_eager_1e6(b *testing.B) {
	ctx := enchanter.NewContext()
	a, c := makeF(1_000_000)
	sa := series.NewSeriesFloat64(a, nil, false, ctx)
	sc := series.NewSeriesFloat64(c, nil, false, ctx)
	runtime.GC()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sum := sa.Add(sc)
		mask := sum.Gt(0.0)
		_ = sum.Filter(mask)
	}
}

func BenchmarkChain_fused_1e6(b *testing.B) {
	a, c := makeF(1_000_000)
	runtime.GC()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fusedChain(a, c)
	}
}
