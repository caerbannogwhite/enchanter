package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/series"
)

func TestArrowIPCRoundTrip(t *testing.T) {
	ctx := enchanter.NewContext()

	iod := NewIoData(ctx)
	iod.AddSeries(
		series.NewSeriesFloat64([]float64{1.5, 2.5, 3.5}, nil, false, ctx),
		SeriesMeta{Name: "x", Type: meta.Float64Type},
	)
	iod.AddSeries(
		series.NewSeriesInt64([]int64{100, 200, 300}, nil, false, ctx),
		SeriesMeta{Name: "y", Type: meta.Int64Type},
	)
	iod.AddSeries(
		series.NewSeriesBool([]bool{true, false, true}, nil, false, ctx),
		SeriesMeta{Name: "flag", Type: meta.BoolType},
	)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.arrow")

	err := iod.ToArrowIPC().SetPath(path).Write()
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("arrow ipc file not created")
	}

	iod2 := FromArrowIPC(ctx).SetPath(path).Read()
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
	if iod2.SeriesMeta[0].Name != "x" {
		t.Errorf("col 0 name: expected 'x', got %q", iod2.SeriesMeta[0].Name)
	}
	if iod2.SeriesMeta[1].Name != "y" {
		t.Errorf("col 1 name: expected 'y', got %q", iod2.SeriesMeta[1].Name)
	}
	if iod2.SeriesMeta[2].Name != "flag" {
		t.Errorf("col 2 name: expected 'flag', got %q", iod2.SeriesMeta[2].Name)
	}

	// Check float values
	f64 := iod2.Series[0].(series.Float64s)
	if f64.Float64s()[0] != 1.5 || f64.Float64s()[1] != 2.5 || f64.Float64s()[2] != 3.5 {
		t.Errorf("col 0 values mismatch: %v", f64.Float64s())
	}

	// Check int values
	i64 := iod2.Series[1].(series.Int64s)
	if i64.Int64s()[0] != 100 || i64.Int64s()[1] != 200 || i64.Int64s()[2] != 300 {
		t.Errorf("col 1 values mismatch: %v", i64.Int64s())
	}

	// Check bool values
	bools := iod2.Series[2].(series.Bools)
	if bools.Bools()[0] != true || bools.Bools()[1] != false || bools.Bools()[2] != true {
		t.Errorf("col 2 values mismatch: %v", bools.Bools())
	}
}

func TestArrowIPCRoundTripNullable(t *testing.T) {
	ctx := enchanter.NewContext()

	iod := NewIoData(ctx)
	iod.AddSeries(
		series.NewSeriesFloat64([]float64{1.0, 0, 3.0}, []bool{false, true, false}, false, ctx),
		SeriesMeta{Name: "val", Type: meta.Float64Type},
	)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nullable.arrow")

	err := iod.ToArrowIPC().SetPath(path).Write()
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	iod2 := FromArrowIPC(ctx).SetPath(path).Read()
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

func TestArrowIPCRoundTripStrings(t *testing.T) {
	ctx := enchanter.NewContext()

	iod := NewIoData(ctx)
	iod.AddSeries(
		series.NewSeriesString([]string{"hello", "world", "test"}, nil, false, ctx),
		SeriesMeta{Name: "text", Type: meta.StringType},
	)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "strings.arrow")

	err := iod.ToArrowIPC().SetPath(path).Write()
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	iod2 := FromArrowIPC(ctx).SetPath(path).Read()
	if iod2.Error != nil {
		t.Fatalf("read: %v", iod2.Error)
	}

	s := iod2.Series[0].(series.Strings)
	if *s.Interned()[0] != "hello" || *s.Interned()[1] != "world" || *s.Interned()[2] != "test" {
		t.Errorf("string values mismatch")
	}
}
