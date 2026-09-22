package io

import (
	"fmt"
	"os"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/series"
)

////////////////////////			ARROW IPC READER

type ArrowIPCReader struct {
	ctx   *enchanter.Context
	path  string
	alloc memory.Allocator
}

func NewArrowIPCReader(ctx *enchanter.Context) *ArrowIPCReader {
	alloc := memory.DefaultAllocator
	if ctx != nil {
		alloc = ctx.Allocator
	}
	return &ArrowIPCReader{ctx: ctx, alloc: alloc}
}

func FromArrowIPC(ctx *enchanter.Context) *ArrowIPCReader {
	return NewArrowIPCReader(ctx)
}

func (r *ArrowIPCReader) SetPath(path string) *ArrowIPCReader {
	r.path = path
	return r
}

func (r *ArrowIPCReader) Read() *IoData {
	iod := NewIoData(r.ctx)

	f, err := os.Open(r.path)
	if err != nil {
		iod.Error = fmt.Errorf("ArrowIPCReader.Read: %w", err)
		return iod
	}
	defer f.Close()

	reader, err := ipc.NewFileReader(f, ipc.WithAllocator(r.alloc))
	if err != nil {
		iod.Error = fmt.Errorf("ArrowIPCReader.Read: %w", err)
		return iod
	}
	defer reader.Close()

	schema := reader.Schema()

	// Read all record batches
	var allSeries []series.Series
	for i := 0; i < reader.NumRecords(); i++ {
		rec, err := reader.RecordBatch(i)
		if err != nil {
			iod.Error = fmt.Errorf("ArrowIPCReader.Read: record %d: %w", i, err)
			return iod
		}

		if i == 0 {
			// Initialize series from first record
			allSeries = make([]series.Series, rec.NumCols())
			for j := 0; j < int(rec.NumCols()); j++ {
				allSeries[j] = series.ArrowArrayToSeries(rec.Column(j), r.ctx)
				if allSeries[j].Err() != nil {
					iod.Error = fmt.Errorf("ArrowIPCReader.Read: column %q: %w", schema.Field(j).Name, allSeries[j].Err())
					return iod
				}
			}
		} else {
			// Append subsequent records
			for j := 0; j < int(rec.NumCols()); j++ {
				chunk := series.ArrowArrayToSeries(rec.Column(j), r.ctx)
				if chunk.Err() != nil {
					iod.Error = fmt.Errorf("ArrowIPCReader.Read: column %q record %d: %w", schema.Field(j).Name, i, chunk.Err())
					return iod
				}
				// Append the series (not its raw data) so the chunk's null
				// mask is merged instead of dropped.
				allSeries[j] = allSeries[j].Append(chunk)
			}
		}
	}

	// Build IoData
	for j := 0; j < schema.NumFields(); j++ {
		name := schema.Field(j).Name
		if j < len(allSeries) {
			iod.AddSeries(allSeries[j], SeriesMeta{Name: name, Type: allSeries[j].Type()})
		}
	}

	iod.FileMeta.FilePath = r.path
	iod.FileMeta.FileFormat = FILE_FORMAT_ARROW_IPC

	return iod
}

////////////////////////			ARROW IPC WRITER

type ArrowIPCWriter struct {
	iod   *IoData
	path  string
	alloc memory.Allocator
}

func NewArrowIPCWriter() *ArrowIPCWriter {
	return &ArrowIPCWriter{alloc: memory.DefaultAllocator}
}

func (w *ArrowIPCWriter) SetIoData(iod *IoData) *ArrowIPCWriter {
	w.iod = iod
	if iod != nil && iod.ctx != nil {
		w.alloc = iod.ctx.Allocator
	}
	return w
}

func (w *ArrowIPCWriter) SetPath(path string) *ArrowIPCWriter {
	w.path = path
	return w
}

func (w *ArrowIPCWriter) Write() error {
	if w.iod == nil {
		return fmt.Errorf("ArrowIPCWriter.Write: no data")
	}

	// Build Arrow schema and record
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
			return fmt.Errorf("ArrowIPCWriter.Write: column %d (%q) returned nil Arrow array", i, name)
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
		return fmt.Errorf("ArrowIPCWriter.Write: %w", err)
	}
	defer f.Close()

	writer, err := ipc.NewFileWriter(f, ipc.WithSchema(schema), ipc.WithAllocator(w.alloc))
	if err != nil {
		return fmt.Errorf("ArrowIPCWriter.Write: %w", err)
	}

	if err := writer.Write(rec); err != nil {
		writer.Close()
		return fmt.Errorf("ArrowIPCWriter.Write: %w", err)
	}

	return writer.Close()
}

func (iod *IoData) ToArrowIPC() *ArrowIPCWriter {
	return NewArrowIPCWriter().SetIoData(iod)
}
