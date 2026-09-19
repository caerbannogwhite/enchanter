package dataframe

import (
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

func TestArrowSchema(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeriesFromFloat64s("price", []float64{1.1, 2.2, 3.3}, nil, false).
		AddSeriesFromInts("qty", []int{10, 20, 30}, nil, false).
		AddSeriesFromStrings("name", []string{"a", "b", "c"}, nil, false)

	schema := df.ArrowSchema()
	if schema.NumFields() != 3 {
		t.Fatalf("expected 3 fields, got %d", schema.NumFields())
	}

	f0 := schema.Field(0)
	if f0.Name != "price" {
		t.Errorf("field 0 name: expected 'price', got %q", f0.Name)
	}
	if f0.Type.ID() != arrow.FLOAT64 {
		t.Errorf("field 0 type: expected FLOAT64, got %v", f0.Type)
	}

	f1 := schema.Field(1)
	if f1.Name != "qty" {
		t.Errorf("field 1 name: expected 'qty', got %q", f1.Name)
	}
	// Ints map to int64 in Arrow
	if f1.Type.ID() != arrow.INT64 {
		t.Errorf("field 1 type: expected INT64, got %v", f1.Type)
	}

	f2 := schema.Field(2)
	if f2.Name != "name" {
		t.Errorf("field 2 name: expected 'name', got %q", f2.Name)
	}
	if f2.Type.ID() != arrow.STRING {
		t.Errorf("field 2 type: expected STRING, got %v", f2.Type)
	}
}

func TestToArrowRecord(t *testing.T) {
	ctx := enchanter.NewContext()
	df := NewDataFrame(ctx).
		AddSeriesFromFloat64s("x", []float64{1.0, 2.0, 3.0}, nil, false).
		AddSeriesFromInt64s("y", []int64{10, 20, 30}, nil, false)

	rec := df.ToArrowRecord()
	defer rec.Release()

	if rec.NumCols() != 2 {
		t.Fatalf("expected 2 cols, got %d", rec.NumCols())
	}
	if rec.NumRows() != 3 {
		t.Fatalf("expected 3 rows, got %d", rec.NumRows())
	}

	col0 := rec.Column(0).(*array.Float64)
	if col0.Value(0) != 1.0 || col0.Value(1) != 2.0 || col0.Value(2) != 3.0 {
		t.Errorf("col 0 values mismatch")
	}

	col1 := rec.Column(1).(*array.Int64)
	if col1.Value(0) != 10 || col1.Value(1) != 20 || col1.Value(2) != 30 {
		t.Errorf("col 1 values mismatch")
	}
}

func TestNewDataFrameFromArrowRecord(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	// Build an Arrow record
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "a", Type: arrow.PrimitiveTypes.Float64, Nullable: false},
		{Name: "b", Type: arrow.BinaryTypes.String, Nullable: true},
	}, nil)

	fb := array.NewFloat64Builder(alloc)
	defer fb.Release()
	fb.AppendValues([]float64{1.1, 2.2}, nil)
	fArr := fb.NewFloat64Array()
	defer fArr.Release()

	sb := array.NewStringBuilder(alloc)
	defer sb.Release()
	sb.Append("hello")
	sb.AppendNull()
	sArr := sb.NewStringArray()
	defer sArr.Release()

	rec := array.NewRecordBatch(schema, []arrow.Array{fArr, sArr}, 2)
	defer rec.Release()

	df := NewDataFrameFromArrowRecord(rec, ctx)
	if df.Err() != nil {
		t.Fatal(df.Err())
	}
	if df.NCols() != 2 {
		t.Fatalf("expected 2 cols, got %d", df.NCols())
	}
	if df.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", df.NRows())
	}

	names := df.Names()
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("names mismatch: %v", names)
	}

	types := df.Types()
	if types[0] != meta.Float64Type {
		t.Errorf("expected Float64Type for col 0, got %v", types[0])
	}
	if types[1] != meta.StringType {
		t.Errorf("expected StringType for col 1, got %v", types[1])
	}

	// Check nullable string column
	s := df.C("b")
	if !s.IsNullable() {
		t.Error("col 'b' should be nullable")
	}
	if !s.IsNull(1) {
		t.Error("col 'b' index 1 should be null")
	}
}

func TestArrowRecordRoundTrip(t *testing.T) {
	ctx := enchanter.NewContext()

	// Build a DataFrame
	df := NewDataFrame(ctx).
		AddSeriesFromFloat64s("x", []float64{1.5, 2.5}, nil, false).
		AddSeriesFromInts("y", []int{10, 20}, nil, false).
		AddSeriesFromBools("z", []bool{true, false}, nil, false)

	// Convert to Arrow Record
	rec := df.ToArrowRecord()
	defer rec.Release()

	// Convert back to DataFrame
	df2 := NewDataFrameFromArrowRecord(rec, ctx)
	if df2.Err() != nil {
		t.Fatal(df2.Err())
	}

	if df2.NCols() != 3 {
		t.Fatalf("expected 3 cols, got %d", df2.NCols())
	}
	if df2.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", df2.NRows())
	}

	// Verify column names preserved
	names := df2.Names()
	if names[0] != "x" || names[1] != "y" || names[2] != "z" {
		t.Errorf("names mismatch: %v", names)
	}
}
