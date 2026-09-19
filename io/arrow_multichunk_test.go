package io

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/series"
)

// Multi-chunk reads: parquet row groups / IPC record batches after the first
// must keep their null masks when concatenated into a single series.

func TestParquetMultiRowGroupWithNulls(t *testing.T) {
	ctx := enchanter.NewContext()
	path := filepath.Join(t.TempDir(), "multi.parquet")

	const n = 10
	fdata := make([]float64, n)
	fmask := make([]bool, n)
	sdata := make([]string, n)
	smask := make([]bool, n)
	for i := 0; i < n; i++ {
		fdata[i] = float64(i) * 1.5
		fmask[i] = i%3 == 1 // nulls at 1, 4, 7 — row groups after the first contain nulls
		sdata[i] = fmt.Sprintf("row%02d", i)
		smask[i] = i >= 5 // nulls only in later row groups
	}
	fs := series.NewSeriesFloat64(fdata, fmask, false, ctx)
	ss := series.NewSeriesString(sdata, smask, false, ctx)

	farr := fs.ArrowArray()
	sarr := ss.ArrowArray()
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "f", Type: farr.DataType(), Nullable: true},
		{Name: "s", Type: sarr.DataType(), Nullable: true},
	}, nil)
	rec := array.NewRecordBatch(schema, []arrow.Array{farr, sarr}, int64(n))
	defer rec.Release()
	farr.Release()
	sarr.Release()
	tbl := array.NewTableFromRecords(schema, []arrow.RecordBatch{rec})
	defer tbl.Release()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// Chunk size 3 over 10 rows -> row groups of 3, 3, 3, 1.
	if err := pqarrow.WriteTable(tbl, f, 3, parquet.NewWriterProperties(), pqarrow.DefaultWriterProps()); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// Note: pqarrow.ReadTable may merge row groups into a single chunk, so
	// this is an end-to-end row-group check; the multi-chunk concatenation
	// path is covered directly by TestMultiChunkToSeriesPreservesNulls.
	iod := FromParquet(ctx).SetPath(path).Read()
	if iod.Error != nil {
		t.Fatal(iod.Error)
	}
	if len(iod.Series) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(iod.Series))
	}

	got := iod.Series[0]
	if got.Len() != n {
		t.Fatalf("f: length %d, want %d", got.Len(), n)
	}
	gf, ok := got.(series.Float64s)
	if !ok {
		t.Fatalf("f: expected Float64s, got %T", got)
	}
	for i := 0; i < n; i++ {
		if got.IsNull(i) != fmask[i] {
			t.Fatalf("f: IsNull(%d) = %v, want %v (nulls in row groups after the first must survive)", i, got.IsNull(i), fmask[i])
		}
		if !fmask[i] && gf.Data_[i] != fdata[i] {
			t.Fatalf("f: value[%d] = %v, want %v", i, gf.Data_[i], fdata[i])
		}
	}

	gots := iod.Series[1]
	gs, ok := gots.(series.Strings)
	if !ok {
		t.Fatalf("s: expected Strings, got %T", gots)
	}
	for i := 0; i < n; i++ {
		if gots.IsNull(i) != smask[i] {
			t.Fatalf("s: IsNull(%d) = %v, want %v (nulls in row groups after the first must survive)", i, gots.IsNull(i), smask[i])
		}
		if !smask[i] && *gs.Data_[i] != sdata[i] {
			t.Fatalf("s: value[%d] = %q, want %q", i, *gs.Data_[i], sdata[i])
		}
	}
}

// pqarrow.ReadTable may hand back a single merged chunk, so exercise the
// multi-chunk concatenation helper directly: null masks in chunks after the
// first must be merged, including promotion of a non-nullable first chunk.
func TestMultiChunkToSeriesPreservesNulls(t *testing.T) {
	ctx := enchanter.NewContext()

	c1 := series.NewSeriesFloat64([]float64{0, 1, 2}, nil, false, ctx) // not nullable
	c2 := series.NewSeriesFloat64([]float64{3, 4, 5}, []bool{false, true, false}, false, ctx)
	c3 := series.NewSeriesFloat64([]float64{6, 7, 8}, []bool{true, false, false}, false, ctx)

	a1 := c1.ArrowArray()
	defer a1.Release()
	a2 := c2.ArrowArray()
	defer a2.Release()
	a3 := c3.ArrowArray()
	defer a3.Release()

	got := multiChunkToSeries([]arrow.Array{a1, a2, a3}, ctx)
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
	if got.Len() != 9 {
		t.Fatalf("length %d, want 9", got.Len())
	}
	wantNull := map[int]bool{4: true, 6: true}
	g := got.(series.Float64s)
	for i := 0; i < 9; i++ {
		if got.IsNull(i) != wantNull[i] {
			t.Fatalf("IsNull(%d) = %v, want %v (chunk null masks must be merged)", i, got.IsNull(i), wantNull[i])
		}
		if !wantNull[i] && g.Data_[i] != float64(i) {
			t.Fatalf("value[%d] = %v, want %d", i, g.Data_[i], i)
		}
	}

	// String chunks exercise the pointer-data + string-pool path.
	s1 := series.NewSeriesString([]string{"a", "b"}, nil, false, ctx)
	s2 := series.NewSeriesString([]string{"c", "d"}, []bool{true, false}, false, ctx)
	b1 := s1.ArrowArray()
	defer b1.Release()
	b2 := s2.ArrowArray()
	defer b2.Release()
	gs := multiChunkToSeries([]arrow.Array{b1, b2}, ctx)
	if gs.Err() != nil {
		t.Fatal(gs.Err())
	}
	if gs.Len() != 4 || !gs.IsNull(2) || gs.IsNull(3) {
		t.Fatalf("string chunks: len=%d IsNull(2)=%v IsNull(3)=%v, want 4/true/false", gs.Len(), gs.IsNull(2), gs.IsNull(3))
	}
	if v := *(gs.(series.Strings)).Data_[3]; v != "d" {
		t.Fatalf("string value[3] = %q, want %q", v, "d")
	}
}

func TestArrowIPCMultiRecordWithNulls(t *testing.T) {
	ctx := enchanter.NewContext()
	path := filepath.Join(t.TempDir(), "multi.arrow")

	s1 := series.NewSeriesFloat64([]float64{0, 1, 2, 3, 4}, []bool{false, true, false, false, false}, false, ctx)
	s2 := series.NewSeriesFloat64([]float64{5, 6, 7, 8, 9}, []bool{false, false, true, false, false}, false, ctx)

	a1 := s1.ArrowArray()
	a2 := s2.ArrowArray()
	schema := arrow.NewSchema([]arrow.Field{{Name: "f", Type: a1.DataType(), Nullable: true}}, nil)

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w, err := ipc.NewFileWriter(f, ipc.WithSchema(schema))
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	rec1 := array.NewRecordBatch(schema, []arrow.Array{a1}, 5)
	rec2 := array.NewRecordBatch(schema, []arrow.Array{a2}, 5)
	a1.Release()
	a2.Release()
	if err := w.Write(rec1); err != nil {
		t.Fatal(err)
	}
	if err := w.Write(rec2); err != nil {
		t.Fatal(err)
	}
	rec1.Release()
	rec2.Release()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	iod := FromArrowIPC(ctx).SetPath(path).Read()
	if iod.Error != nil {
		t.Fatal(iod.Error)
	}
	got := iod.Series[0]
	if got.Len() != 10 {
		t.Fatalf("length %d, want 10", got.Len())
	}
	g := got.(series.Float64s)
	wantNull := map[int]bool{1: true, 7: true}
	for i := 0; i < 10; i++ {
		if got.IsNull(i) != wantNull[i] {
			t.Fatalf("IsNull(%d) = %v, want %v (nulls in record batches after the first must survive)", i, got.IsNull(i), wantNull[i])
		}
		if !wantNull[i] && g.Data_[i] != float64(i) {
			t.Fatalf("value[%d] = %v, want %v", i, g.Data_[i], float64(i))
		}
	}
}
