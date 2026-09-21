package dataframe

import (
	"path/filepath"
	"testing"

	"github.com/caerbannogwhite/enchanter"
)

// A CSV written through the frame's WriteCsv builder reads back through
// ReadCsv with values and nulls intact.
func TestCsvRoundTrip(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{1, 0, 3}, []bool{false, true, false}, false).
		AddSeriesFromFloat64s("x", []float64{1.5, 2.5, 3.5}, nil, false).
		AddSeriesFromStrings("name", []string{"ada", "bob", "cyn"}, nil, false)

	path := filepath.Join(t.TempDir(), "roundtrip.csv")
	if err := df.WriteCsv().SetPath(path).SetEol("\n").Write(); err != nil {
		t.Fatal(err)
	}

	back := ReadCsv(ctx).SetPath(path).SetNullValues(true).Read()
	if back.Err() != nil {
		t.Fatal(back.Err())
	}
	if back.NRows() != 3 || back.NCols() != 3 {
		t.Fatalf("shape: expected 3x3, got %dx%d", back.NRows(), back.NCols())
	}

	if !back.Col("id").IsNull(1) {
		t.Error("the null id must survive the round trip")
	}
	if back.Col("id").IsNull(0) || back.Col("id").IsNull(2) {
		t.Error("non-null ids must stay non-null")
	}
	if got := back.Col("id").Get(0).(int64); got != 1 {
		t.Errorf("id[0]: expected 1, got %v", got)
	}
	if got := back.Col("x").Get(2).(float64); got != 3.5 {
		t.Errorf("x[2]: expected 3.5, got %v", got)
	}
	if got := back.Col("name").Get(1).(string); got != "bob" {
		t.Errorf("name[1]: expected bob, got %q", got)
	}
}
