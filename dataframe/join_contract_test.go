package dataframe

import (
	"testing"

	"github.com/caerbannogwhite/enchanter"
)

// A row with several matches produces consecutive output rows: left
// rows in left order, and each one's matches in right order.
func TestJoin_ManyToMany(t *testing.T) {
	ctx := enchanter.NewContext()
	left := NewDataFrame(ctx).
		AddSeriesFromInt64s("k", []int64{1, 1, 2}, nil, false).
		AddSeriesFromStrings("l", []string{"a", "b", "c"}, nil, false)
	right := NewDataFrame(ctx).
		AddSeriesFromInt64s("k", []int64{1, 1}, nil, false).
		AddSeriesFromStrings("r", []string{"x", "y"}, nil, false)

	res := left.Join(JoinInner, right, "k")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 4 {
		t.Fatalf("rows: expected 4, got %d", res.NRows())
	}
	wantL := []string{"a", "a", "b", "b"}
	wantR := []string{"x", "y", "x", "y"}
	for i := range wantL {
		if got := res.Col("l").Get(i).(string); got != wantL[i] {
			t.Errorf("l[%d]: expected %q, got %q", i, wantL[i], got)
		}
		if got := res.Col("r").Get(i).(string); got != wantR[i] {
			t.Errorf("r[%d]: expected %q, got %q", i, wantR[i], got)
		}
	}
}

// Null keys match null keys, on both sides.
func TestJoin_NullKeysMatch(t *testing.T) {
	ctx := enchanter.NewContext()
	left := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{1, 0, 2}, []bool{false, true, false}, false).
		AddSeriesFromStrings("l", []string{"l0", "l1", "l2"}, nil, false)
	right := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{0, 3}, []bool{true, false}, false).
		AddSeriesFromStrings("r", []string{"r0", "r1"}, nil, false)

	res := left.Join(JoinInner, right, "id")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 1 {
		t.Fatalf("rows: expected 1 (the null keys), got %d", res.NRows())
	}
	if !res.Col("id").IsNull(0) {
		t.Error("the joined key must be null")
	}
	if res.Col("l").Get(0).(string) != "l1" || res.Col("r").Get(0).(string) != "r0" {
		t.Errorf("values: expected l1 and r0, got %v and %v",
			res.Col("l").Get(0), res.Col("r").Get(0))
	}

	// Left join: unmatched left rows keep their place, right side null.
	res = left.Join(JoinLeft, right, "id")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 3 {
		t.Fatalf("left join rows: expected 3, got %d", res.NRows())
	}
	if res.Col("r").IsNull(0) != true || res.Col("r").IsNull(1) != false || res.Col("r").IsNull(2) != true {
		t.Errorf("r nulls: expected [true false true], got [%v %v %v]",
			res.Col("r").IsNull(0), res.Col("r").IsNull(1), res.Col("r").IsNull(2))
	}
}

// A right join follows the right frame's row order, and its unmatched
// rows leave the left columns null.
func TestJoin_RightOrder(t *testing.T) {
	ctx := enchanter.NewContext()
	left := NewDataFrame(ctx).
		AddSeriesFromInt64s("k", []int64{5, 4}, nil, false).
		AddSeriesFromStrings("l", []string{"p", "q"}, nil, false)
	right := NewDataFrame(ctx).
		AddSeriesFromInt64s("k", []int64{4, 4, 6}, nil, false).
		AddSeriesFromStrings("r", []string{"r0", "r1", "r2"}, nil, false)

	res := left.Join(JoinRight, right, "k")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 3 {
		t.Fatalf("rows: expected 3, got %d", res.NRows())
	}
	wantK := []int64{4, 4, 6}
	wantR := []string{"r0", "r1", "r2"}
	for i := range wantK {
		if got := res.Col("k").Get(i).(int64); got != wantK[i] {
			t.Errorf("k[%d]: expected %d, got %d", i, wantK[i], got)
		}
		if got := res.Col("r").Get(i).(string); got != wantR[i] {
			t.Errorf("r[%d]: expected %q, got %q", i, wantR[i], got)
		}
	}
	if res.Col("l").Get(0).(string) != "q" || res.Col("l").Get(1).(string) != "q" {
		t.Error("matched right rows must carry the left value q")
	}
	if !res.Col("l").IsNull(2) {
		t.Error("the unmatched right row must have a null left column")
	}
}

// An outer join with a colliding column name suffixes both sides and
// nulls each side's half where the row has no counterpart.
func TestJoin_OuterCollision(t *testing.T) {
	ctx := enchanter.NewContext()
	left := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{1, 2}, nil, false).
		AddSeriesFromStrings("v", []string{"a", "b"}, nil, false)
	right := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{2, 3}, nil, false).
		AddSeriesFromStrings("v", []string{"x", "y"}, nil, false)

	res := left.Join(JoinOuter, right, "id")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	names := res.Names()
	want := []string{"id", "v_x", "v_y"}
	for i, w := range want {
		if names[i] != w {
			t.Fatalf("names: expected %v, got %v", want, names)
		}
	}
	if res.NRows() != 3 {
		t.Fatalf("rows: expected 3, got %d", res.NRows())
	}
	// Row order: left rows first (1 unmatched, 2 matched), then the
	// unmatched right row 3.
	wantID := []int64{1, 2, 3}
	for i, w := range wantID {
		if got := res.Col("id").Get(i).(int64); got != w {
			t.Errorf("id[%d]: expected %d, got %d", i, w, got)
		}
	}
	if !res.Col("v_y").IsNull(0) || res.Col("v_y").IsNull(1) || !res.Col("v_x").IsNull(2) {
		t.Error("null pattern wrong: v_y null only at row 0, v_x null only at row 2")
	}
	if res.Col("v_x").Get(0).(string) != "a" || res.Col("v_y").Get(1).(string) != "x" {
		t.Error("values wrong in the joined halves")
	}
}

// Joining against an empty frame: inner drops everything, left keeps
// every left row with a null right side.
func TestJoin_EmptyRight(t *testing.T) {
	ctx := enchanter.NewContext()
	left := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{1, 2}, nil, false).
		AddSeriesFromStrings("l", []string{"a", "b"}, nil, false)
	right := NewDataFrame(ctx).
		AddSeriesFromInt64s("id", []int64{}, nil, false).
		AddSeriesFromStrings("r", []string{}, nil, false)

	res := left.Join(JoinInner, right, "id")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 0 {
		t.Fatalf("inner: expected 0 rows, got %d", res.NRows())
	}

	res = left.Join(JoinLeft, right, "id")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.NRows() != 2 {
		t.Fatalf("left: expected 2 rows, got %d", res.NRows())
	}
	if !res.Col("r").IsNull(0) || !res.Col("r").IsNull(1) {
		t.Error("left join against an empty frame must null the right columns")
	}
}
