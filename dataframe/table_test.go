package dataframe

import (
	"strings"
	"testing"

	"github.com/caerbannogwhite/enchanter"
)

// Table returns the rendered text table, so a consumer can place it
// itself instead of having it printed.
func TestDataFrame_Table(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3}, nil, false).
		AddSeriesFromStrings("name", []string{"ada", "bob", "cyn"}, nil, false)

	out := df.Table(NewPPrintParams())
	for _, want := range []string{"3 rows, 2 columns", "id", "name", "ada", "Int64", "String"} {
		if !strings.Contains(out, want) {
			t.Errorf("table must contain %q, got:\n%s", want, out)
		}
	}

	// An errored frame renders as its error.
	broken := df.Select("missing")
	if out := broken.Table(NewPPrintParams()); !strings.Contains(out, "missing") {
		t.Errorf("an errored frame must render its error, got %q", out)
	}

	// An empty frame says so.
	empty := NewDataFrame(ctx).AddSeriesFromInt64s("id", []int64{}, nil, false)
	if out := empty.Table(NewPPrintParams()); !strings.Contains(out, "Empty DataFrame") {
		t.Errorf("an empty frame must say so, got %q", out)
	}
}
