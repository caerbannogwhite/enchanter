package formatter

import (
	"strings"
	"testing"

	"github.com/caerbannogwhite/enchanter"
)

// A single pushed string used to panic in Compute: the 80th-percentile
// index came out at -1.
func Test_StringFormatter_SingleValue(t *testing.T) {
	f := NewStringFormatter()
	f.Push("only")
	f.Compute()
	if f.GetMaxWidth() != 4 {
		t.Errorf("width: expected 4, got %d", f.GetMaxWidth())
	}
}

// No pushed values must not panic either.
func Test_StringFormatter_NoValues(t *testing.T) {
	f := NewStringFormatter()
	f.Compute()
	if f.GetMaxWidth() != 0 {
		t.Errorf("width: expected 0, got %d", f.GetMaxWidth())
	}
}

// A null element must be padded to the requested width like any value,
// otherwise the table row comes out shorter than its column.
func Test_StringFormatter_NaPadded(t *testing.T) {
	f := NewStringFormatter()
	want := enchanter.NA_TEXT + strings.Repeat(" ", 8-len(enchanter.NA_TEXT))
	if out := f.Format(8, "x", true); out != want {
		t.Errorf("NA cell: expected %q, got %q", want, out)
	}
	// A non-string value renders as a null too.
	if out := f.Format(8, nil, false); out != want {
		t.Errorf("non-string cell: expected %q, got %q", want, out)
	}
}
