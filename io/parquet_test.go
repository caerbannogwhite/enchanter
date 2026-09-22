package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/series"
)

func TestParquetRoundTrip(t *testing.T) {
	ctx := enchanter.NewContext()

	// Build IoData
	iod := NewIoData(ctx)
	iod.AddSeries(
		series.NewSeriesFloat64([]float64{1.1, 2.2, 3.3}, nil, false, ctx),
		SeriesMeta{Name: "price", Type: meta.Float64Type},
	)
	iod.AddSeries(
		series.NewSeriesInt64([]int64{10, 20, 30}, nil, false, ctx),
		SeriesMeta{Name: "qty", Type: meta.Int64Type},
	)
	iod.AddSeries(
		series.NewSeriesString([]string{"a", "b", "c"}, nil, false, ctx),
		SeriesMeta{Name: "name", Type: meta.StringType},
	)

	// Write to temp file
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.parquet")

	err := iod.ToParquet().SetPath(path).Write()
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("parquet file not created")
	}

	// Read back
	iod2 := FromParquet(ctx).SetPath(path).Read()
	if iod2.Error != nil {
		t.Fatalf("read: %v", iod2.Error)
	}

	if iod2.NCols() != 3 {
		t.Fatalf("expected 3 cols, got %d", iod2.NCols())
	}
	if iod2.NRows() != 3 {
		t.Fatalf("expected 3 rows, got %d", iod2.NRows())
	}

	// Check column names
	if iod2.SeriesMeta[0].Name != "price" {
		t.Errorf("col 0 name: expected 'price', got %q", iod2.SeriesMeta[0].Name)
	}
	if iod2.SeriesMeta[1].Name != "qty" {
		t.Errorf("col 1 name: expected 'qty', got %q", iod2.SeriesMeta[1].Name)
	}
	if iod2.SeriesMeta[2].Name != "name" {
		t.Errorf("col 2 name: expected 'name', got %q", iod2.SeriesMeta[2].Name)
	}

	// Check values
	s0 := iod2.Series[0]
	if s0.Len() != 3 {
		t.Fatalf("col 0 len: expected 3, got %d", s0.Len())
	}
	f64 := s0.(series.Float64s)
	if f64.Float64s()[0] != 1.1 || f64.Float64s()[1] != 2.2 || f64.Float64s()[2] != 3.3 {
		t.Errorf("col 0 values mismatch: %v", f64.Float64s())
	}
}

func TestParquetRoundTripNullable(t *testing.T) {
	ctx := enchanter.NewContext()

	iod := NewIoData(ctx)
	iod.AddSeries(
		series.NewSeriesFloat64([]float64{1.0, 0, 3.0}, []bool{false, true, false}, false, ctx),
		SeriesMeta{Name: "x", Type: meta.Float64Type},
	)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nullable.parquet")

	err := iod.ToParquet().SetPath(path).Write()
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	iod2 := FromParquet(ctx).SetPath(path).Read()
	if iod2.Error != nil {
		t.Fatalf("read: %v", iod2.Error)
	}

	s := iod2.Series[0]
	if !s.IsNullable() {
		t.Error("should be nullable")
	}
	if s.NullCount() != 1 {
		t.Errorf("expected 1 null, got %d", s.NullCount())
	}
	if !s.IsNull(1) {
		t.Error("index 1 should be null")
	}
}
