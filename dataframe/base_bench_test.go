package dataframe

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/caerbannogwhite/enchanter/series"
)

var G1_1e4_1e2_0_0_df *DataFrame
var G1_1e5_1e2_0_0_df *DataFrame
var G1_1e6_1e2_0_0_df *DataFrame
var G1_1e7_1e2_0_0_df *DataFrame
var G1_1e4_1e2_10_0_df *DataFrame
var G1_1e5_1e2_10_0_df *DataFrame
var G1_1e6_1e2_10_0_df *DataFrame
var G1_1e7_1e2_10_0_df *DataFrame

func read_G1_1e4_1e2_0_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e4_1e2_0_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(false).
			SetReader(f).
			Read()

		f.Close()

		G1_1e4_1e2_0_0_df = &df
	} else {
		G1_1e4_1e2_0_0_df = nil
	}
}

func read_G1_1e5_1e2_0_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e5_1e2_0_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(false).
			SetReader(f).
			Read()

		f.Close()

		G1_1e5_1e2_0_0_df = &df
	} else {
		G1_1e5_1e2_0_0_df = nil
	}
}

func read_G1_1e6_1e2_0_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e6_1e2_0_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(false).
			SetReader(f).
			Read()

		f.Close()

		G1_1e6_1e2_0_0_df = &df
	} else {
		G1_1e6_1e2_0_0_df = nil
	}
}

func read_G1_1e7_1e2_0_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e7_1e2_0_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(false).
			SetReader(f).
			Read()

		f.Close()

		G1_1e7_1e2_0_0_df = &df
	} else {
		G1_1e7_1e2_0_0_df = nil
	}
}

func read_G1_1e4_1e2_10_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e4_1e2_10_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(true).
			SetReader(f).
			Read()

		f.Close()

		G1_1e4_1e2_10_0_df = &df
	} else {
		G1_1e4_1e2_10_0_df = nil
	}
}

func read_G1_1e5_1e2_10_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e5_1e2_10_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(true).
			SetReader(f).
			Read()

		f.Close()

		G1_1e5_1e2_10_0_df = &df
	} else {
		G1_1e5_1e2_10_0_df = nil
	}
}

func read_G1_1e6_1e2_10_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e6_1e2_10_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(true).
			SetReader(f).
			Read()

		f.Close()

		G1_1e6_1e2_10_0_df = &df
	} else {
		G1_1e6_1e2_10_0_df = nil
	}
}

func read_G1_1e7_1e2_10_0() {
	f, err := os.OpenFile(filepath.Join(testDataDir, "G1_1e7_1e2_10_0.csv"), os.O_RDONLY, 0666)
	if err == nil {
		df := NewDataFrame(ctx).
			FromCsv().
			SetDelimiter(',').
			SetNullValues(true).
			SetReader(f).
			Read()

		f.Close()

		G1_1e7_1e2_10_0_df = &df
	} else {
		G1_1e7_1e2_10_0_df = nil
	}
}

func Benchmark_Filter_Q1_1e5(b *testing.B) {
	if G1_1e5_1e2_0_0_df == nil {
		b.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_0_0_df)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		df.Filter(
			df.Col("id6").Map(func(v any) any {
				return v.(int64) > 500
			}).(series.Bools).Or(
				df.Col("id1").Map(func(v any) any {
					return v.(string) == "id024"
				}).(series.Bools)).(series.Bools))
	}
	b.StopTimer()
}

func Benchmark_Filter_Q1_1e6(b *testing.B) {
	if G1_1e6_1e2_0_0_df == nil {
		b.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_0_0_df)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		df.Filter(
			df.Col("id6").Map(func(v any) any {
				return v.(int64) > 500
			}).(series.Bools).Or(
				df.Col("id1").Map(func(v any) any {
					return v.(string) == "id024"
				}).(series.Bools)).(series.Bools))
	}
	b.StopTimer()
}

func Benchmark_Filter_Q1_1e7(b *testing.B) {
	if G1_1e7_1e2_0_0_df == nil {
		b.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_0_0_df)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		df.Filter(
			df.Col("id6").Map(func(v any) any {
				return v.(int64) > 500
			}).(series.Bools).Or(
				df.Col("id1").Map(func(v any) any {
					return v.(string) == "id024"
				}).(series.Bools)).(series.Bools))
	}
	b.StopTimer()
}

func Benchmark_Filter_Q2_1e5(b *testing.B) {
	if G1_1e5_1e2_0_0_df == nil {
		b.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_0_0_df)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		df.Filter(
			df.Col("id6").Map(func(v any) any {
				return v.(int64) > 500
			}).(series.Bools).And(
				df.Col("v3").Map(func(v any) any {
					return v.(float64) < 50
				}).(series.Bools),
			).(series.Bools).And(
				df.Col("id1").Map(func(v any) any {
					return v.(string) == "id024"
				}).(series.Bools).Or(
					df.Col("id2").Map(func(v any) any {
						return v.(string) == "id024"
					}).(series.Bools),
				),
			).(series.Bools).And(
				df.Col("v1").Map(func(v any) any {
					return v.(int64) == 5
				}).(series.Bools),
			).(series.Bools).And(
				df.Col("v2").Map(func(v any) any {
					return v.(int64) == 1
				}).(series.Bools),
			).(series.Bools),
		)
	}
	b.StopTimer()
}

func Benchmark_Filter_Q2_1e6(b *testing.B) {
	if G1_1e6_1e2_0_0_df == nil {
		b.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_0_0_df)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		df.Filter(
			df.Col("id6").Map(func(v any) any {
				return v.(int64) > 500
			}).(series.Bools).And(
				df.Col("v3").Map(func(v any) any {
					return v.(float64) < 50
				}).(series.Bools),
			).(series.Bools).And(
				df.Col("id1").Map(func(v any) any {
					return v.(string) == "id024"
				}).(series.Bools).Or(
					df.Col("id2").Map(func(v any) any {
						return v.(string) == "id024"
					}).(series.Bools),
				),
			).(series.Bools).And(
				df.Col("v1").Map(func(v any) any {
					return v.(int64) == 5
				}).(series.Bools),
			).(series.Bools).And(
				df.Col("v2").Map(func(v any) any {
					return v.(int64) == 1
				}).(series.Bools),
			).(series.Bools),
		)
	}
	b.StopTimer()
}

func Benchmark_Filter_Q2_1e7(b *testing.B) {
	if G1_1e7_1e2_0_0_df == nil {
		b.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_0_0_df)
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		df.Filter(
			df.Col("id6").Map(func(v any) any {
				return v.(int64) > 500
			}).(series.Bools).And(
				df.Col("v3").Map(func(v any) any {
					return v.(float64) < 50
				}).(series.Bools),
			).(series.Bools).And(
				df.Col("id1").Map(func(v any) any {
					return v.(string) == "id024"
				}).(series.Bools).Or(
					df.Col("id2").Map(func(v any) any {
						return v.(string) == "id024"
					}).(series.Bools),
				),
			).(series.Bools).And(
				df.Col("v1").Map(func(v any) any {
					return v.(int64) == 5
				}).(series.Bools),
			).(series.Bools).And(
				df.Col("v2").Map(func(v any) any {
					return v.(int64) == 1
				}).(series.Bools),
			).(series.Bools),
		)
	}
	b.StopTimer()
}

////////////////////////			GROUP BY TESTS
//
// GroupBy challenge: more info here https://github.com/h2oai/db-benchmark/tree/master

func Test_GroupBy_Q1_1e4(t *testing.T) {
	if G1_1e4_1e2_0_0_df == nil {
		t.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_0_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 30027 {
		t.Errorf("Expected 30027, got %f", check)
	}
}

func Test_GroupBy_Q1_1e5(t *testing.T) {
	if G1_1e5_1e2_0_0_df == nil {
		t.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_0_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 300292 {
		t.Errorf("Expected 300292, got %f", check)
	}
}

func Test_GroupBy_Q1_1e6(t *testing.T) {
	if G1_1e6_1e2_0_0_df == nil {
		t.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_0_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 3000297 {
		t.Errorf("Expected 3000297, got %f", check)
	}
}

func Test_GroupBy_Q1_1e7(t *testing.T) {
	if G1_1e7_1e2_0_0_df == nil {
		t.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_0_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 29998789 {
		t.Errorf("Expected 29998789, got %f", check)
	}
}

func Test_GroupBy_Q1_1e4_10PercNAs(t *testing.T) {
	if G1_1e4_1e2_10_0_df == nil {
		t.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_10_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 27044 {
		t.Errorf("Expected 27044, got %f", check)
	}
}

func Test_GroupBy_Q1_1e5_10PercNAs(t *testing.T) {
	if G1_1e5_1e2_10_0_df == nil {
		t.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_10_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 270421 {
		t.Errorf("Expected 270421, got %f", check)
	}
}

func Test_GroupBy_Q1_1e6_10PercNAs(t *testing.T) {
	if G1_1e6_1e2_10_0_df == nil {
		t.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_10_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 2700684 {
		t.Errorf("Expected 2700684, got %f", check)
	}
}

func Test_GroupBy_Q1_1e7_10PercNAs(t *testing.T) {
	if G1_1e7_1e2_10_0_df == nil {
		t.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_10_0_df).
		GroupBy("id1").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 26998588 {
		t.Errorf("Expected 26998588, got %f", check)
	}
}

func Test_GroupBy_Q2_1e4(t *testing.T) {
	if G1_1e4_1e2_0_0_df == nil {
		t.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_0_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 6272 {
		t.Errorf("Expected 6272 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 30027 {
		t.Errorf("Expected 30027, got %f", check)
	}
}

func Test_GroupBy_Q2_1e5(t *testing.T) {
	if G1_1e5_1e2_0_0_df == nil {
		t.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_0_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 9999 {
		t.Errorf("Expected 9999 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 300292 {
		t.Errorf("Expected 300292, got %f", check)
	}
}

func Test_GroupBy_Q2_1e6(t *testing.T) {
	if G1_1e6_1e2_0_0_df == nil {
		t.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_0_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 10000 {
		t.Errorf("Expected 10000 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 3000297 {
		t.Errorf("Expected 3000297, got %f", check)
	}
}

func Test_GroupBy_Q2_1e7(t *testing.T) {
	if G1_1e7_1e2_0_0_df == nil {
		t.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_0_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		Run()

	if df.NRows() != 10000 {
		t.Errorf("Expected 10000 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 29998789 {
		t.Errorf("Expected 29998789, got %f", check)
	}
}

func Test_GroupBy_Q2_1e4_10PercNAs(t *testing.T) {
	if G1_1e4_1e2_10_0_df == nil {
		t.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_10_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 5276 {
		t.Errorf("Expected 5276 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 27044 {
		t.Errorf("Expected 27044, got %f", check)
	}
}

func Test_GroupBy_Q2_1e5_10PercNAs(t *testing.T) {
	if G1_1e5_1e2_10_0_df == nil {
		t.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_10_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 8280 {
		t.Errorf("Expected 8280 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 270421 {
		t.Errorf("Expected 270421, got %f", check)
	}
}

func Test_GroupBy_Q2_1e6_10PercNAs(t *testing.T) {
	if G1_1e6_1e2_10_0_df == nil {
		t.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_10_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 8281 {
		t.Errorf("Expected 8281 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 2700684 {
		t.Errorf("Expected 2700684, got %f", check)
	}
}

func Test_GroupBy_Q2_1e7_10PercNAs(t *testing.T) {
	if G1_1e7_1e2_10_0_df == nil {
		t.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_10_0_df).
		GroupBy("id1", "id2").
		Agg(Sum("v1")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 8281 {
		t.Errorf("Expected 8281 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check != 26998588 {
		t.Errorf("Expected 26998588, got %f", check)
	}
}

func Test_GroupBy_Q3_1e4(t *testing.T) {
	if G1_1e4_1e2_0_0_df == nil {
		t.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_0_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 30027 {
		t.Errorf("Expected 30027, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 4999.383247863238, 10e-6) {
		t.Errorf("Expected 4999.383247863238, got %f", check2)
	}
}

func Test_GroupBy_Q3_1e5(t *testing.T) {
	if G1_1e5_1e2_0_0_df == nil {
		t.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_0_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		Run()

	if df.NRows() != 1000 {
		t.Errorf("Expected 1000 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 300292 {
		t.Errorf("Expected 300292, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 50095.22836212861, 10e-6) {
		t.Errorf("Expected 50095.22836212861, got %f", check2)
	}
}

func Test_GroupBy_Q3_1e6(t *testing.T) {
	if G1_1e6_1e2_0_0_df == nil {
		t.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_0_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		Run()

	if df.NRows() != 10000 {
		t.Errorf("Expected 10000 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 3000297 {
		t.Errorf("Expected 3000297, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 500393.46150263766, 10e-6) {
		t.Errorf("Expected 500393.46150263766, got %f", check2)
	}
}

func Test_GroupBy_Q3_1e7(t *testing.T) {
	if G1_1e7_1e2_0_0_df == nil {
		t.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_0_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		Run()

	if df.NRows() != 100000 {
		t.Errorf("Expected 100000 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 29998789 {
		t.Errorf("Expected 29998789, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 4999719.62234443, 10e-6) {
		t.Errorf("Expected 4999719.62234443, got %f", check2)
	}
}

func Test_GroupBy_Q3_1e4_10PercNAs(t *testing.T) {
	if G1_1e4_1e2_10_0_df == nil {
		t.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_10_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 27044 {
		t.Errorf("Expected 27044, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 4546.896912, 10e-6) {
		t.Errorf("Expected 4546.896912, got %f", check2)
	}
}

func Test_GroupBy_Q3_1e5_10PercNAs(t *testing.T) {
	if G1_1e5_1e2_10_0_df == nil {
		t.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_10_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 901 {
		t.Errorf("Expected 901 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 270421 {
		t.Errorf("Expected 270421, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 45102.875061, 10e-6) {
		t.Errorf("Expected 45102.875061, got %f", check2)
	}
}

func Test_GroupBy_Q3_1e6_10PercNAs(t *testing.T) {
	if G1_1e6_1e2_10_0_df == nil {
		t.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_10_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 9001 {
		t.Errorf("Expected 9001 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 2700684 {
		t.Errorf("Expected 2700684, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 450405.970609, 10e-6) {
		t.Errorf("Expected 450405.970609, got %f", check2)
	}
}

func Test_GroupBy_Q3_1e7_10PercNAs(t *testing.T) {
	if G1_1e7_1e2_10_0_df == nil {
		t.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_10_0_df).
		GroupBy("id3").
		Agg(Sum("v1"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 90001 {
		t.Errorf("Expected 90001 rows, got %d", df.NRows())
	}

	if df.NCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if check1 != 26998588 {
		t.Errorf("Expected 26998588, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check2, 4498637.521119, 10e-6) {
		t.Errorf("Expected 4498637.521119, got %f", check2)
	}
}

func Test_GroupBy_Q4_1e4(t *testing.T) {
	if G1_1e4_1e2_0_0_df == nil {
		t.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_0_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 300.1460223942026, 10e-6) {
		t.Errorf("Expected 300.1460223942026, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 803.8206781360852, 10e-6) {
		t.Errorf("Expected 803.8206781360852, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 5008.9079567058325, 10e-6) {
		t.Errorf("Expected 5008.9079567058325, got %f", check3)
	}
}

func Test_GroupBy_Q4_1e5(t *testing.T) {
	if G1_1e5_1e2_0_0_df == nil {
		t.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_0_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 300.29996127903826, 10e-6) {
		t.Errorf("Expected 300.29996127903826, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 800.8632014058803, 10e-6) {
		t.Errorf("Expected 800.8632014058803, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 5009.03811345283, 10e-6) {
		t.Errorf("Expected 5009.03811345283, got %f", check3)
	}
}

func Test_GroupBy_Q4_1e6(t *testing.T) {
	if G1_1e6_1e2_0_0_df == nil {
		t.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_0_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 300.0300474405866, 10e-6) {
		t.Errorf("Expected 300.0300474405866, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 799.8113837581368, 10e-6) {
		t.Errorf("Expected 799.8113837581368, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 5003.666447664572, 10e-6) {
		t.Errorf("Expected 5003.666447664572, got %f", check3)
	}
}

func Test_GroupBy_Q4_1e7(t *testing.T) {
	if G1_1e7_1e2_0_0_df == nil {
		t.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_0_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 299.9879818750654, 10e-6) {
		t.Errorf("Expected 299.9879818750654, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 799.8941794099782, 10e-6) {
		t.Errorf("Expected 799.8941794099782, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 4999.766872833688, 10e-6) {
		t.Errorf("Expected 4999.766872833688, got %f", check3)
	}
}

func Test_GroupBy_Q4_1e4_10PercNAs(t *testing.T) {
	if G1_1e4_1e2_10_0_df == nil {
		t.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_10_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 272.526949, 10e-6) {
		t.Errorf("Expected 272.526949, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 730.148493, 10e-6) {
		t.Errorf("Expected 730.148493, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 4555.480100, 10e-6) {
		t.Errorf("Expected 4555.480100, got %f", check3)
	}
}

func Test_GroupBy_Q4_1e5_10PercNAs(t *testing.T) {
	if G1_1e5_1e2_10_0_df == nil {
		t.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_10_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 273.552775, 10e-6) {
		t.Errorf("Expected 273.552775, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 728.905170, 10e-6) {
		t.Errorf("Expected 728.905170, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 4557.974075, 10e-6) {
		t.Errorf("Expected 4557.974075, got %f", check3)
	}
}

func Test_GroupBy_Q4_1e6_10PercNAs(t *testing.T) {
	if G1_1e6_1e2_10_0_df == nil {
		t.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_10_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 273.093643, 10e-6) {
		t.Errorf("Expected 273.093643, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 727.793263, 10e-6) {
		t.Errorf("Expected 727.793263, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 4554.072670, 10e-6) {
		t.Errorf("Expected 4554.072670, got %f", check3)
	}
}

func Test_GroupBy_Q4_1e7_10PercNAs(t *testing.T) {
	if G1_1e7_1e2_10_0_df == nil {
		t.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_10_0_df).
		GroupBy("id4").
		Agg(Mean("v1"), Mean("v2"), Mean("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("mean(v1)")).Run().Col("sum(mean(v1))").Get(0).(float64)
	if !equalFloats(check1, 272.973598, 10e-6) {
		t.Errorf("Expected 272.973598, got %f", check1)
	}

	check2 := df.Agg(Sum("mean(v2)")).Run().Col("sum(mean(v2))").Get(0).(float64)
	if !equalFloats(check2, 727.937711, 10e-6) {
		t.Errorf("Expected 727.937711, got %f", check2)
	}

	check3 := df.Agg(Sum("mean(v3)")).Run().Col("sum(mean(v3))").Get(0).(float64)
	if !equalFloats(check3, 4549.183890, 10e-6) {
		t.Errorf("Expected 4549.183890, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e4(t *testing.T) {
	if G1_1e4_1e2_0_0_df == nil {
		t.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_0_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		Run()

	if df.NRows() != 100 {
		t.Errorf("Expected 100 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 30027, 10e-6) {
		t.Errorf("Expected 30027, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 80396, 10e-6) {
		t.Errorf("Expected 80396, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 500378.166716, 10e-6) {
		t.Errorf("Expected 500378.166716, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e5(t *testing.T) {
	if G1_1e5_1e2_0_0_df == nil {
		t.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_0_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		Run()

	if df.NRows() != 1000 {
		t.Errorf("Expected 1000 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 300292, 10e-6) {
		t.Errorf("Expected 300292, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 800809, 10e-6) {
		t.Errorf("Expected 800809, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 5009219.2870470015, 10e-6) {
		t.Errorf("Expected 5009219.2870470015, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e6(t *testing.T) {
	if G1_1e6_1e2_0_0_df == nil {
		t.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_0_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		Run()

	if df.NRows() != 10000 {
		t.Errorf("Expected 10000 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 3000297, 10e-6) {
		t.Errorf("Expected 3000297, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 7998131, 10e-6) {
		t.Errorf("Expected 7998131, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 50037098.685274005, 10e-6) {
		t.Errorf("Expected 50037098.685274005, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e7(t *testing.T) {
	if G1_1e7_1e2_0_0_df == nil {
		t.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_0_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		Run()

	if df.NRows() != 100000 {
		t.Errorf("Expected 100000 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 29998789, 10e-6) {
		t.Errorf("Expected 29998789, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 79989360, 10e-6) {
		t.Errorf("Expected 79989360, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 499976651.4080609, 10e-6) {
		t.Errorf("Expected 499976651.4080609, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e4_10PercNAs(t *testing.T) {
	if G1_1e4_1e2_10_0_df == nil {
		t.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e4_1e2_10_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 91 {
		t.Errorf("Expected 91 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 27044, 10e-6) {
		t.Errorf("Expected 27044, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 72373, 10e-6) {
		t.Errorf("Expected 72373, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 449477.651724, 10e-6) {
		t.Errorf("Expected 449477.651724, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e5_10PercNAs(t *testing.T) {
	if G1_1e5_1e2_10_0_df == nil {
		t.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e5_1e2_10_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 901 {
		t.Errorf("Expected 901 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 270421, 10e-6) {
		t.Errorf("Expected 270421, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 720829, 10e-6) {
		t.Errorf("Expected 720829, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 4508009.682434, 10e-6) {
		t.Errorf("Expected 4508009.682434, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e6_10PercNAs(t *testing.T) {
	if G1_1e6_1e2_10_0_df == nil {
		t.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e6_1e2_10_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 9001 {
		t.Errorf("Expected 9001 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 2700684, 10e-6) {
		t.Errorf("Expected 2700684, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 7198551, 10e-6) {
		t.Errorf("Expected 7198551, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 45036571.337614, 10e-6) {
		t.Errorf("Expected 45036571.337614, got %f", check3)
	}
}

func Test_GroupBy_Q5_1e7_10PercNAs(t *testing.T) {
	if G1_1e7_1e2_10_0_df == nil {
		t.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	df := (*G1_1e7_1e2_10_0_df).
		GroupBy("id6").
		Agg(Sum("v1"), Sum("v2"), Sum("v3")).
		RemoveNAs(true).
		Run()

	if df.NRows() != 90001 {
		t.Errorf("Expected 90001 rows, got %d", df.NRows())
	}

	if df.NCols() != 4 {
		t.Errorf("Expected 4 columns, got %d", df.NCols())
	}

	check1 := df.Agg(Sum("sum(v1)")).Run().Col("sum(sum(v1))").Get(0).(float64)
	if !equalFloats(check1, 26998588, 10e-6) {
		t.Errorf("Expected 26998588, got %f", check1)
	}

	check2 := df.Agg(Sum("sum(v2)")).Run().Col("sum(sum(v2))").Get(0).(float64)
	if !equalFloats(check2, 71993788, 10e-6) {
		t.Errorf("Expected 71993788, got %f", check2)
	}

	check3 := df.Agg(Sum("sum(v3)")).Run().Col("sum(sum(v3))").Get(0).(float64)
	if !equalFloats(check3, 449932870.447177, 10e-6) {
		t.Errorf("Expected 449932870.447177, got %f", check3)
	}
}

////////////////////////			GROUP BY BENCHMARKS

func Benchmark_GroupBy_Q1_1e4(b *testing.B) {
	if G1_1e4_1e2_0_0_df == nil {
		b.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_0_0_df).GroupBy("id1").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q1_1e5(b *testing.B) {
	if G1_1e5_1e2_0_0_df == nil {
		b.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_0_0_df).GroupBy("id1").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q1_1e6(b *testing.B) {
	if G1_1e6_1e2_0_0_df == nil {
		b.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_0_0_df).GroupBy("id1").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q1_1e7(b *testing.B) {
	if G1_1e7_1e2_0_0_df == nil {
		b.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_0_0_df).GroupBy("id1").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q1_1e4_10PercNAs(b *testing.B) {
	if G1_1e4_1e2_10_0_df == nil {
		b.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_10_0_df).GroupBy("id1").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q1_1e5_10PercNAs(b *testing.B) {
	if G1_1e5_1e2_10_0_df == nil {
		b.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_10_0_df).GroupBy("id1").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q1_1e6_10PercNAs(b *testing.B) {
	if G1_1e6_1e2_10_0_df == nil {
		b.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_10_0_df).GroupBy("id1").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q1_1e7_10PercNAs(b *testing.B) {
	if G1_1e7_1e2_10_0_df == nil {
		b.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_10_0_df).GroupBy("id1").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e4(b *testing.B) {
	if G1_1e4_1e2_0_0_df == nil {
		b.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_0_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e5(b *testing.B) {
	if G1_1e5_1e2_0_0_df == nil {
		b.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_0_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e6(b *testing.B) {
	if G1_1e6_1e2_0_0_df == nil {
		b.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_0_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e7(b *testing.B) {
	if G1_1e7_1e2_0_0_df == nil {
		b.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_0_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e4_10PercNAs(b *testing.B) {
	if G1_1e4_1e2_10_0_df == nil {
		b.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_10_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e5_10PercNAs(b *testing.B) {
	if G1_1e5_1e2_10_0_df == nil {
		b.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_10_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e6_10PercNAs(b *testing.B) {
	if G1_1e6_1e2_10_0_df == nil {
		b.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_10_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q2_1e7_10PercNAs(b *testing.B) {
	if G1_1e7_1e2_10_0_df == nil {
		b.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_10_0_df).GroupBy("id1", "id2").
			Agg(Sum("v1")).RemoveNAs(true).Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e4(b *testing.B) {
	if G1_1e4_1e2_0_0_df == nil {
		b.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_0_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e5(b *testing.B) {
	if G1_1e5_1e2_0_0_df == nil {
		b.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_0_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e6(b *testing.B) {
	if G1_1e6_1e2_0_0_df == nil {
		b.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_0_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e7(b *testing.B) {
	if G1_1e7_1e2_0_0_df == nil {
		b.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_0_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e4_10PercNAs(b *testing.B) {
	if G1_1e4_1e2_10_0_df == nil {
		b.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_10_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e5_10PercNAs(b *testing.B) {
	if G1_1e5_1e2_10_0_df == nil {
		b.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_10_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e6_10PercNAs(b *testing.B) {
	if G1_1e6_1e2_10_0_df == nil {
		b.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_10_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q3_1e7_10PercNAs(b *testing.B) {
	if G1_1e7_1e2_10_0_df == nil {
		b.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_10_0_df).GroupBy("id3").
			Agg(Sum("v1"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e4(b *testing.B) {
	if G1_1e4_1e2_0_0_df == nil {
		b.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_0_0_df).
			GroupBy("id4").
			Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e5(b *testing.B) {
	if G1_1e5_1e2_0_0_df == nil {
		b.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_0_0_df).
			GroupBy("id4").
			Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e6(b *testing.B) {
	if G1_1e6_1e2_0_0_df == nil {
		b.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_0_0_df).GroupBy("id4").Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e7(b *testing.B) {
	if G1_1e7_1e2_0_0_df == nil {
		b.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_0_0_df).GroupBy("id4").Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e4_10PercNAs(b *testing.B) {
	if G1_1e4_1e2_10_0_df == nil {
		b.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_10_0_df).
			GroupBy("id4").
			Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e5_10PercNAs(b *testing.B) {
	if G1_1e5_1e2_10_0_df == nil {
		b.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_10_0_df).
			GroupBy("id4").
			Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e6_10PercNAs(b *testing.B) {
	if G1_1e6_1e2_10_0_df == nil {
		b.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_10_0_df).
			GroupBy("id4").
			Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q4_1e7_10PercNAs(b *testing.B) {
	if G1_1e7_1e2_10_0_df == nil {
		b.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_10_0_df).
			GroupBy("id4").
			Agg(Mean("v1"), Mean("v2"), Mean("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e4(b *testing.B) {
	if G1_1e4_1e2_0_0_df == nil {
		b.Skip("G1_1e4_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_0_0_df).
			GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e5(b *testing.B) {
	if G1_1e5_1e2_0_0_df == nil {
		b.Skip("G1_1e5_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_0_0_df).
			GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e6(b *testing.B) {
	if G1_1e6_1e2_0_0_df == nil {
		b.Skip("G1_1e6_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_0_0_df).GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e7(b *testing.B) {
	if G1_1e7_1e2_0_0_df == nil {
		b.Skip("G1_1e7_1e2_0_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_0_0_df).GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e4_10PercNAs(b *testing.B) {
	if G1_1e4_1e2_10_0_df == nil {
		b.Skip("G1_1e4_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e4_1e2_10_0_df).
			GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e5_10PercNAs(b *testing.B) {
	if G1_1e5_1e2_10_0_df == nil {
		b.Skip("G1_1e5_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e5_1e2_10_0_df).
			GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e6_10PercNAs(b *testing.B) {
	if G1_1e6_1e2_10_0_df == nil {
		b.Skip("G1_1e6_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e6_1e2_10_0_df).
			GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}

func Benchmark_GroupBy_Q5_1e7_10PercNAs(b *testing.B) {
	if G1_1e7_1e2_10_0_df == nil {
		b.Skip("G1_1e7_1e2_10_0 dataframe not loaded")
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(*G1_1e7_1e2_10_0_df).
			GroupBy("id6").
			Agg(Sum("v1"), Sum("v2"), Sum("v3")).
			RemoveNAs(true).
			Run()
	}
	b.StopTimer()
}
