package series

import (
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

func TestArrowArrayToSeriesFloat64(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	builder := array.NewFloat64Builder(alloc)
	defer builder.Release()
	builder.AppendValues([]float64{1.5, 2.5, 3.5}, nil)
	arr := builder.NewFloat64Array()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Err() != nil {
		t.Fatal(s.Err())
	}
	if s.Len() != 3 {
		t.Fatalf("expected len 3, got %d", s.Len())
	}
	if s.Type() != meta.Float64Type {
		t.Fatalf("expected Float64Type, got %v", s.Type())
	}
	f := s.(Float64s)
	if f.Data_[0] != 1.5 || f.Data_[1] != 2.5 || f.Data_[2] != 3.5 {
		t.Errorf("values mismatch: %v", f.Data_)
	}
	if s.IsNullable() {
		t.Error("should not be nullable")
	}
}

func TestArrowArrayToSeriesFloat64Nullable(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	builder := array.NewFloat64Builder(alloc)
	defer builder.Release()
	builder.Append(1.0)
	builder.AppendNull()
	builder.Append(3.0)
	arr := builder.NewFloat64Array()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Len() != 3 {
		t.Fatalf("expected len 3, got %d", s.Len())
	}
	if !s.IsNullable() {
		t.Fatal("should be nullable")
	}
	if s.NullCount() != 1 {
		t.Fatalf("expected 1 null, got %d", s.NullCount())
	}
	if !s.IsNull(1) {
		t.Error("index 1 should be null")
	}
}

func TestArrowArrayToSeriesInt64(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	builder := array.NewInt64Builder(alloc)
	defer builder.Release()
	builder.AppendValues([]int64{10, 20, 30}, nil)
	arr := builder.NewInt64Array()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Type() != meta.Int64Type {
		t.Fatalf("expected Int64Type, got %v", s.Type())
	}
	i64 := s.(Int64s)
	if i64.Data_[0] != 10 || i64.Data_[1] != 20 || i64.Data_[2] != 30 {
		t.Errorf("values mismatch: %v", i64.Data_)
	}
}

func TestArrowArrayToSeriesBool(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	builder := array.NewBooleanBuilder(alloc)
	defer builder.Release()
	builder.AppendValues([]bool{true, false, true}, nil)
	arr := builder.NewBooleanArray()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Type() != meta.BoolType {
		t.Fatalf("expected BoolType, got %v", s.Type())
	}
	b := s.(Bools)
	if b.Data_[0] != true || b.Data_[1] != false || b.Data_[2] != true {
		t.Errorf("values mismatch: %v", b.Data_)
	}
}

func TestArrowArrayToSeriesString(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	builder := array.NewStringBuilder(alloc)
	defer builder.Release()
	builder.AppendValues([]string{"hello", "world"}, nil)
	arr := builder.NewStringArray()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Type() != meta.StringType {
		t.Fatalf("expected StringType, got %v", s.Type())
	}
	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
	str := s.(Strings)
	if *str.Data_[0] != "hello" || *str.Data_[1] != "world" {
		t.Errorf("values mismatch: %v %v", *str.Data_[0], *str.Data_[1])
	}
}

func TestArrowArrayToSeriesTimestamp(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	t1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)

	builder := array.NewTimestampBuilder(alloc, &arrow.TimestampType{Unit: arrow.Nanosecond})
	defer builder.Release()
	builder.Append(arrow.Timestamp(t1.UnixNano()))
	builder.Append(arrow.Timestamp(t2.UnixNano()))
	arr := builder.NewTimestampArray()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Type() != meta.TimeType {
		t.Fatalf("expected TimeType, got %v", s.Type())
	}
	ts := s.(Times)
	if !ts.Data_[0].Equal(t1) {
		t.Errorf("time 0: expected %v, got %v", t1, ts.Data_[0])
	}
	if !ts.Data_[1].Equal(t2) {
		t.Errorf("time 1: expected %v, got %v", t2, ts.Data_[1])
	}
}

func TestArrowArrayToSeriesDuration(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	d1 := 5 * time.Second
	d2 := 100 * time.Millisecond

	builder := array.NewDurationBuilder(alloc, &arrow.DurationType{Unit: arrow.Nanosecond})
	defer builder.Release()
	builder.Append(arrow.Duration(d1.Nanoseconds()))
	builder.Append(arrow.Duration(d2.Nanoseconds()))
	arr := builder.NewDurationArray()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Type() != meta.DurationType {
		t.Fatalf("expected DurationType, got %v", s.Type())
	}
	ds := s.(Durations)
	if ds.Data_[0] != d1 {
		t.Errorf("duration 0: expected %v, got %v", d1, ds.Data_[0])
	}
	if ds.Data_[1] != d2 {
		t.Errorf("duration 1: expected %v, got %v", d2, ds.Data_[1])
	}
}

func TestArrowArrayToSeriesNull(t *testing.T) {
	ctx := enchanter.NewContext()
	alloc := memory.DefaultAllocator

	builder := array.NewNullBuilder(alloc)
	defer builder.Release()
	builder.AppendNull()
	builder.AppendNull()
	builder.AppendNull()
	arr := builder.NewNullArray()
	defer arr.Release()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Type() != meta.NullType {
		t.Fatalf("expected NullType, got %v", s.Type())
	}
	if s.Len() != 3 {
		t.Fatalf("expected len 3, got %d", s.Len())
	}
}

func TestArrowArrayToSeriesRoundTrip(t *testing.T) {
	ctx := enchanter.NewContext()

	// Create a series, convert to Arrow, convert back
	orig := NewSeriesFloat64([]float64{1.1, 2.2, 3.3}, nil, false, ctx)
	arr := orig.ArrowArray()

	s := ArrowArrayToSeries(arr, ctx)
	if s.Len() != 3 {
		t.Fatalf("expected len 3, got %d", s.Len())
	}
	f := s.(Float64s)
	if f.Data_[0] != 1.1 || f.Data_[1] != 2.2 || f.Data_[2] != 3.3 {
		t.Errorf("round-trip values mismatch: %v", f.Data_)
	}
}
