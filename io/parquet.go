package io

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/series"
)

////////////////////////			PARQUET READER

type ParquetReader struct {
	ctx   *enchanter.Context
	path  string
	alloc memory.Allocator
}

func NewParquetReader(ctx *enchanter.Context) *ParquetReader {
	alloc := memory.DefaultAllocator
	if ctx != nil {
		alloc = ctx.Allocator
	}
	return &ParquetReader{ctx: ctx, alloc: alloc}
}

func FromParquet(ctx *enchanter.Context) *ParquetReader {
	return NewParquetReader(ctx)
}

func (r *ParquetReader) SetPath(path string) *ParquetReader {
	r.path = path
	return r
}

func (r *ParquetReader) Read() *IoData {
	iod := NewIoData(r.ctx)

	f, err := os.Open(r.path)
	if err != nil {
		iod.Error = fmt.Errorf("ParquetReader.Read: %w", err)
		return iod
	}
	defer f.Close()

	tbl, err := pqarrow.ReadTable(context.Background(), f, parquet.NewReaderProperties(nil), pqarrow.ArrowReadProperties{}, r.alloc)
	if err != nil {
		iod.Error = fmt.Errorf("ParquetReader.Read: %w", err)
		return iod
	}
	defer tbl.Release()

	schema := tbl.Schema()
	for i := 0; i < int(tbl.NumCols()); i++ {
		col := tbl.Column(i)
		name := schema.Field(i).Name

		// Arrow tables store columns as chunked arrays; combine chunks.
		chunks := col.Data().Chunks()
		if len(chunks) == 0 {
			iod.AddSeries(series.NewSeriesNA(0, r.ctx), SeriesMeta{Name: name})
			continue
		}

		// For single-chunk columns, use directly.
		if len(chunks) == 1 {
			s := series.ArrowArrayToSeries(chunks[0], r.ctx)
			if s.Err() != nil {
				iod.Error = fmt.Errorf("ParquetReader.Read: column %q: %w", name, s.Err())
				return iod
			}
			iod.AddSeries(s, SeriesMeta{Name: name, Type: s.Type()})
			continue
		}

		// Multi-chunk: materialize into a single array via concatenation.
		s := multiChunkToSeries(chunks, r.ctx)
		if s.Err() != nil {
			iod.Error = fmt.Errorf("ParquetReader.Read: column %q: %w", name, s.Err())
			return iod
		}
		iod.AddSeries(s, SeriesMeta{Name: name, Type: s.Type()})
	}

	iod.FileMeta.FilePath = r.path
	iod.FileMeta.FileFormat = FILE_FORMAT_PARQUET

	return iod
}

// multiChunkToSeries handles multi-chunk Arrow columns by converting each
// chunk to a series and appending them together.
func multiChunkToSeries(chunks []arrow.Array, ctx *enchanter.Context) series.Series {
	if len(chunks) == 0 {
		return series.NewSeriesNA(0, ctx)
	}

	result := series.ArrowArrayToSeries(chunks[0], ctx)
	for i := 1; i < len(chunks); i++ {
		chunk := series.ArrowArrayToSeries(chunks[i], ctx)
		if chunk.Err() != nil {
			return chunk
		}
		// Append the series (not its raw data) so the chunk's null mask is
		// merged instead of dropped.
		result = result.Append(chunk)
	}
	return result
}

////////////////////////			PARQUET WRITER

type ParquetWriter struct {
	iod   *IoData
	path  string
	alloc memory.Allocator
}

func NewParquetWriter() *ParquetWriter {
	return &ParquetWriter{alloc: memory.DefaultAllocator}
}

func (w *ParquetWriter) SetIoData(iod *IoData) *ParquetWriter {
	w.iod = iod
	if iod != nil && iod.ctx != nil {
		w.alloc = iod.ctx.Allocator
	}
	return w
}

func (w *ParquetWriter) SetPath(path string) *ParquetWriter {
	w.path = path
	return w
}

func (w *ParquetWriter) Write() error {
	if w.iod == nil {
		return fmt.Errorf("ParquetWriter.Write: no data")
	}

	// Build Arrow schema and columns
	fields := make([]arrow.Field, len(w.iod.Series))
	cols := make([]arrow.Array, len(w.iod.Series))
	var nrows int64 = 0
	if len(w.iod.Series) > 0 {
		nrows = int64(w.iod.Series[0].Len())
	}

	for i, s := range w.iod.Series {
		name := ""
		if i < len(w.iod.SeriesMeta) {
			name = w.iod.SeriesMeta[i].Name
		}

		arr := s.ArrowArray()
		if arr == nil {
			return fmt.Errorf("ParquetWriter.Write: column %d (%q) returned nil Arrow array", i, name)
		}

		fields[i] = arrow.Field{
			Name:     name,
			Type:     arr.DataType(),
			Nullable: s.IsNullable(),
		}
		cols[i] = arr
	}

	schema := arrow.NewSchema(fields, nil)
	rec := makeRecord(schema, cols, nrows)
	defer rec.Release()
	// makeRecord retains the columns; drop our references so rec.Release()
	// frees everything.
	for _, c := range cols {
		c.Release()
	}

	// Write to file
	f, err := os.Create(w.path)
	if err != nil {
		return fmt.Errorf("ParquetWriter.Write: %w", err)
	}
	defer f.Close()

	return writeRecordAsParquet(rec, f)
}

func writeRecordAsParquet(rec arrow.RecordBatch, w io.Writer) error {
	tbl := recordToTable(rec)
	defer tbl.Release()

	writerProps := parquet.NewWriterProperties()
	arrowProps := pqarrow.DefaultWriterProps()

	return pqarrow.WriteTable(tbl, w, rec.NumRows(), writerProps, arrowProps)
}

func (iod *IoData) ToParquet() *ParquetWriter {
	return NewParquetWriter().SetIoData(iod)
}
