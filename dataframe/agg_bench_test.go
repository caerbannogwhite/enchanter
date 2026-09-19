package dataframe

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/caerbannogwhite/enchanter"
)

func loadG1(b *testing.B, name string) DataFrame {
	p := filepath.Join("..", "testdata", name)
	if _, err := os.Stat(p); err != nil {
		b.Skipf("G1 fixture absent: %s", name)
	}
	f, _ := os.Open(p)
	defer f.Close()
	df := NewDataFrame(enchanter.NewContext()).
		FromCsv().SetDelimiter(',').SetNullValues(false).SetReader(f).Read()
	if df.IsErrored() {
		b.Skipf("load failed: %v", df.GetError())
	}
	return df
}

func BenchmarkAgg_Q1_sum_v1_by_id1_1e7(b *testing.B) {
	df := loadG1(b, "G1_1e7_1e2_0_0.csv")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		df.GroupBy("id1").Agg(Sum("v1")).Run()
	}
}

func BenchmarkAgg_multi_key_multi_agg_1e7(b *testing.B) {
	df := loadG1(b, "G1_1e7_1e2_0_0.csv")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		df.GroupBy("id1", "id2").Agg(Sum("v1"), Mean("v3"), Std("v3")).Run()
	}
}
