package dataframe

import (
	"fmt"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/arrowutil"
	"github.com/caerbannogwhite/enchanter/series"
)

// ArrowSchema returns the Arrow schema corresponding to this DataFrame's columns.
func (df DataFrame) ArrowSchema() *arrow.Schema {
	fields := make([]arrow.Field, len(df.series))
	for i, s := range df.series {
		dt := arrowutil.BaseTypeToArrowType(s.Type())
		if dt == nil {
			dt = arrow.Null
		}
		fields[i] = arrow.Field{
			Name:     df.names[i],
			Type:     dt,
			Nullable: s.IsNullable(),
		}
	}
	return arrow.NewSchema(fields, nil)
}

// ToArrowRecord converts this DataFrame to an Arrow RecordBatch (columnar batch).
// The column arrays are freshly built copies owned by the record: the caller
// should Release() the returned record when done (optional but recommended
// under GC-backed allocators, see enchanter.Context.Allocator).
func (df DataFrame) ToArrowRecord() arrow.RecordBatch {
	alloc := memory.DefaultAllocator
	if df.ctx != nil {
		alloc = df.ctx.Allocator
	}
	schema := df.ArrowSchema()
	cols := make([]arrow.Array, len(df.series))
	for i, s := range df.series {
		arr := s.ArrowArray()
		if arr == nil {
			// Build a null array for unsupported types
			builder := array.NewNullBuilder(alloc)
			for j := 0; j < s.Len(); j++ {
				builder.AppendNull()
			}
			cols[i] = builder.NewNullArray()
			builder.Release()
		} else {
			cols[i] = arr
		}
	}
	rec := array.NewRecordBatch(schema, cols, int64(df.NRows()))
	// NewRecordBatch retains the columns; drop our references so the record is
	// the sole owner and rec.Release() frees everything.
	for _, c := range cols {
		c.Release()
	}
	return rec
}

// NewDataFrameFromArrowRecord creates a DataFrame from an Arrow RecordBatch.
// Each column in the record becomes a Series in the DataFrame. Column data is
// materialized into Go slices, so this function does not take ownership of the
// record: the caller may Release it as soon as this function returns.
func NewDataFrameFromArrowRecord(record arrow.RecordBatch, ctx *enchanter.Context) DataFrame {
	if ctx == nil {
		return DataFrame{err: fmt.Errorf("NewDataFrameFromArrowRecord: context is nil")}
	}

	df := NewDataFrame(ctx)

	schema := record.Schema()
	for i := 0; i < int(record.NumCols()); i++ {
		col := record.Column(i)
		name := schema.Field(i).Name
		s := series.ArrowArrayToSeries(col, ctx)
		if s.Err() != nil {
			df.err = fmt.Errorf("NewDataFrameFromArrowRecord: column %q: %w", name, s.Err())
			return df
		}
		df = df.AddSeries(name, s)
	}

	return df
}
