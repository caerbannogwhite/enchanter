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

// A one-row string column used to panic inside the string formatter's
// width computation.
func TestDataFrame_Table_OneRow(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).AddSeriesFromStrings("s", []string{"only"}, nil, false)
	if out := df.Table(NewPPrintParams()); !strings.Contains(out, "only") {
		t.Errorf("table must contain the value, got:\n%s", out)
	}
}

// Every box line of the table has the same width, null string cells
// included: an unpadded NA cell used to shorten its row.
func TestDataFrame_Table_NullStringAlignment(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeriesFromStrings("s", []string{"abcdef", "x", "ghijkl"}, []bool{false, true, false}, true)
	out := df.Table(NewPPrintParams())
	want := -1
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(line, "╭") {
			want = len([]rune(line))
		}
		if want > 0 && strings.HasPrefix(line, "│") && len([]rune(line)) != want {
			t.Errorf("misaligned row (%d runes, want %d): %q", len([]rune(line)), want, line)
		}
	}
}

// The empty-frame text survives the lipgloss branch: it used to discard
// the rendered string.
func TestDataFrame_Table_EmptyLipGloss(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).AddSeriesFromStrings("s", []string{}, nil, false)
	out := df.Table(NewPPrintParams().SetUseLipGloss(true))
	if !strings.Contains(out, "Empty DataFrame") {
		t.Errorf("empty frame with lipgloss lost its text, got %q", out)
	}
}
