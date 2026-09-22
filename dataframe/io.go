package dataframe

import (
	"encoding/binary"
	"io"

	"github.com/caerbannogwhite/enchanter"
	encio "github.com/caerbannogwhite/enchanter/io"
	"github.com/caerbannogwhite/enchanter/meta"
)

func FromIoData(iod *encio.IoData) DataFrame {
	df := NewDataFrame(iod.Context())

	if iod.Error != nil {
		df.err = iod.Error
	}

	for i, s := range iod.Series {
		df = df.AddSeries(iod.SeriesMeta[i].Name, s)
	}

	return df
}

func (df DataFrame) ToIoData() *encio.IoData {
	iod := encio.NewIoData(df.ctx)

	iod.Error = df.Err()

	for i, s := range df.series {
		iod.AddSeries(s, encio.SeriesMeta{
			Name: df.names[i],
		})
	}

	return iod
}

////////////////////////			CSV READER

// CsvReader reads CSV data into a DataFrame. Create one with ReadCsv,
// configure it with the Set methods, then call Read.
type CsvReader struct {
	reader *encio.CsvReader
}

// ReadCsv creates a CSV reader with the given context.
func ReadCsv(ctx *enchanter.Context) *CsvReader {
	return &CsvReader{
		reader: encio.NewCsvReader(ctx),
	}
}

func (r *CsvReader) SetHeader(header bool) *CsvReader {
	r.reader = r.reader.SetHeader(header)
	return r
}

func (r *CsvReader) SetDelimiter(delimiter rune) *CsvReader {
	r.reader = r.reader.SetDelimiter(delimiter)
	return r
}

func (r *CsvReader) SetGuessDataTypeLen(guessDataTypeLen int) *CsvReader {
	r.reader = r.reader.SetGuessDataTypeLen(guessDataTypeLen)
	return r
}

func (r *CsvReader) SetRows(rows int) *CsvReader {
	r.reader = r.reader.SetRows(rows)
	return r
}

func (r *CsvReader) SetPath(path string) *CsvReader {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *CsvReader) SetNullValues(nullValues bool) *CsvReader {
	r.reader = r.reader.SetNullValues(nullValues)
	return r
}

func (r *CsvReader) SetReader(reader io.Reader) *CsvReader {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *CsvReader) SetSchema(schema *meta.Schema) *CsvReader {
	r.reader = r.reader.SetSchema(schema)
	return r
}

func (r *CsvReader) SetContext(ctx *enchanter.Context) *CsvReader {
	r.reader = r.reader.SetContext(ctx)
	return r
}

func (r *CsvReader) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			CSV WRITER

// CsvWriter writes CSV data from a DataFrame. Create one with WriteCsv,
// configure it with the Set methods, then call Write.
type CsvWriter struct {
	writer *encio.CsvWriter
}

// WriteCsv starts a CSV write of the frame.
func (df DataFrame) WriteCsv() *CsvWriter {
	return &CsvWriter{
		writer: encio.NewCsvWriter().SetIoData(df.ToIoData()),
	}
}

func (w *CsvWriter) SetDelimiter(delimiter rune) *CsvWriter {
	w.writer = w.writer.SetDelimiter(delimiter)
	return w
}

func (w *CsvWriter) SetHeader(header bool) *CsvWriter {
	w.writer = w.writer.SetHeader(header)
	return w
}

func (w *CsvWriter) SetFormat(format bool) *CsvWriter {
	w.writer = w.writer.SetFormat(format)
	return w
}

func (w *CsvWriter) SetPath(path string) *CsvWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *CsvWriter) SetNaText(naText string) *CsvWriter {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *CsvWriter) SetEol(eol string) *CsvWriter {
	w.writer = w.writer.SetEol(eol)
	return w
}

func (w *CsvWriter) SetQuote(quote string) *CsvWriter {
	w.writer = w.writer.SetQuote(quote)
	return w
}

func (w *CsvWriter) SetQuoting(quoting encio.CsvQuotingType) *CsvWriter {
	w.writer = w.writer.SetQuoting(quoting)
	return w
}

func (w *CsvWriter) SetWriter(writer io.Writer) *CsvWriter {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *CsvWriter) Write() error {
	return w.writer.Write()
}

////////////////////////			JSON READER

// JsonReader reads JSON data into a DataFrame. Create one with ReadJson,
// configure it with the Set methods, then call Read.
type JsonReader struct {
	reader *encio.JsonReader
}

// ReadJson creates a JSON reader with the given context.
func ReadJson(ctx *enchanter.Context) *JsonReader {
	return &JsonReader{
		reader: encio.NewJsonReader(ctx),
	}
}

func (r *JsonReader) SetPath(path string) *JsonReader {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *JsonReader) SetReader(reader io.Reader) *JsonReader {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *JsonReader) SetSchema(schema *meta.Schema) *JsonReader {
	r.reader = r.reader.SetSchema(schema)
	return r
}

func (r *JsonReader) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			JSON WRITER

// JsonWriter writes JSON data from a DataFrame. Create one with WriteJson,
// configure it with the Set methods, then call Write.
type JsonWriter struct {
	writer *encio.JsonWriter
}

// WriteJson starts a JSON write of the frame.
func (df DataFrame) WriteJson() *JsonWriter {
	return &JsonWriter{
		writer: encio.NewJsonWriter().SetIoData(df.ToIoData()),
	}
}

func (w *JsonWriter) SetPath(path string) *JsonWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *JsonWriter) SetNewLine(newLine string) *JsonWriter {
	w.writer = w.writer.SetNewLine(newLine)
	return w
}

func (w *JsonWriter) SetIndent(indent string) *JsonWriter {
	w.writer = w.writer.SetIndent(indent)
	return w
}

func (w *JsonWriter) SetWriter(writer io.Writer) *JsonWriter {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *JsonWriter) Write() error {
	return w.writer.Write()
}

////////////////////////			XPT READER

// XptReader reads XPT data into a DataFrame. Create one with ReadXpt,
// configure it with the Set methods, then call Read.
type XptReader struct {
	reader *encio.XptReader
}

// ReadXpt creates a XPT reader with the given context.
func ReadXpt(ctx *enchanter.Context) *XptReader {
	return &XptReader{
		reader: encio.NewXptReader(ctx),
	}
}

func (r *XptReader) SetMaxObservations(maxObservations int) *XptReader {
	r.reader = r.reader.SetMaxObservations(maxObservations)
	return r
}

func (r *XptReader) SetVersion(version encio.XptVersionType) *XptReader {
	r.reader = r.reader.SetVersion(version)
	return r
}

func (r *XptReader) SetByteOrder(byteOrder binary.ByteOrder) *XptReader {
	r.reader = r.reader.SetByteOrder(byteOrder)
	return r
}

func (r *XptReader) SetPath(path string) *XptReader {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *XptReader) SetReader(reader io.Reader) *XptReader {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *XptReader) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			SAS7BDAT READER

// Sas7bdatReader reads SAS7BDAT data into a DataFrame. Create one with ReadSas7bdat,
// configure it with the Set methods, then call Read.
type Sas7bdatReader struct {
	reader *encio.Sas7bdatReader
}

// ReadSas7bdat creates a SAS7BDAT reader with the given context.
func ReadSas7bdat(ctx *enchanter.Context) *Sas7bdatReader {
	return &Sas7bdatReader{
		reader: encio.NewSas7bdatReader(ctx),
	}
}

func (r *Sas7bdatReader) SetPath(path string) *Sas7bdatReader {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *Sas7bdatReader) SetReader(reader io.ReadSeeker) *Sas7bdatReader {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *Sas7bdatReader) SetTrimStrings(trim bool) *Sas7bdatReader {
	r.reader = r.reader.SetTrimStrings(trim)
	return r
}

func (r *Sas7bdatReader) SetConvertDates(convert bool) *Sas7bdatReader {
	r.reader = r.reader.SetConvertDates(convert)
	return r
}

func (r *Sas7bdatReader) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			XPT WRITER

// XptWriter writes XPT data from a DataFrame. Create one with WriteXpt,
// configure it with the Set methods, then call Write.
type XptWriter struct {
	writer *encio.XptWriter
}

// WriteXpt starts a XPT write of the frame.
func (df DataFrame) WriteXpt() *XptWriter {
	return &XptWriter{
		writer: encio.NewXptWriter().SetIoData(df.ToIoData()),
	}
}

func (w *XptWriter) SetVersion(version encio.XptVersionType) *XptWriter {
	w.writer = w.writer.SetVersion(version)
	return w
}

func (w *XptWriter) SetByteOrder(byteOrder binary.ByteOrder) *XptWriter {
	w.writer = w.writer.SetByteOrder(byteOrder)
	return w
}

func (w *XptWriter) SetPath(path string) *XptWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *XptWriter) SetWriter(writer io.Writer) *XptWriter {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *XptWriter) Write() error {
	return w.writer.Write()
}

////////////////////////			XLSX READER

// XlsxReader reads XLSX data into a DataFrame. Create one with ReadXlsx,
// configure it with the Set methods, then call Read.
type XlsxReader struct {
	reader *encio.XlsxReader
}

// ReadXlsx creates a XLSX reader with the given context.
func ReadXlsx(ctx *enchanter.Context) *XlsxReader {
	return &XlsxReader{
		reader: encio.NewXlsxReader(ctx),
	}
}

func (r *XlsxReader) SetPath(path string) *XlsxReader {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *XlsxReader) SetSheet(sheet string) *XlsxReader {
	r.reader = r.reader.SetSheet(sheet)
	return r
}

func (r *XlsxReader) SetHeader(header int) *XlsxReader {
	r.reader = r.reader.SetHeader(header)
	return r
}

func (r *XlsxReader) SetRows(rows int) *XlsxReader {
	r.reader = r.reader.SetRows(rows)
	return r
}

func (r *XlsxReader) SetGuessDataTypeLen(guessDataTypeLen int) *XlsxReader {
	r.reader = r.reader.SetGuessDataTypeLen(guessDataTypeLen)
	return r
}

func (r *XlsxReader) SetNullValues(nullValues bool) *XlsxReader {
	r.reader = r.reader.SetNullValues(nullValues)
	return r
}

func (r *XlsxReader) SetSchema(schema *meta.Schema) *XlsxReader {
	r.reader = r.reader.SetSchema(schema)
	return r
}

func (r *XlsxReader) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			XLSX WRITER

// XlsxWriter writes XLSX data from a DataFrame. Create one with WriteXlsx,
// configure it with the Set methods, then call Write.
type XlsxWriter struct {
	writer *encio.XlsxWriter
}

// WriteXlsx starts a XLSX write of the frame.
func (df DataFrame) WriteXlsx() *XlsxWriter {
	return &XlsxWriter{
		writer: encio.NewXlsxWriter().SetIoData(df.ToIoData()),
	}
}

func (w *XlsxWriter) SetPath(path string) *XlsxWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *XlsxWriter) SetSheet(sheet string) *XlsxWriter {
	w.writer = w.writer.SetSheet(sheet)
	return w
}

func (w *XlsxWriter) SetNaText(naText string) *XlsxWriter {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *XlsxWriter) SetWriter(writer io.Writer) *XlsxWriter {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *XlsxWriter) Write() error {
	return w.writer.Write()
}

////////////////////////			HTML WRITER

// HtmlWriter writes HTML data from a DataFrame. Create one with WriteHtml,
// configure it with the Set methods, then call Write.
type HtmlWriter struct {
	writer *encio.HtmlWriter
}

// WriteHtml starts a HTML write of the frame.
func (df DataFrame) WriteHtml() *HtmlWriter {
	return &HtmlWriter{
		writer: encio.NewHtmlWriter().SetIoData(df.ToIoData()),
	}
}

func (w *HtmlWriter) SetPath(path string) *HtmlWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *HtmlWriter) SetNaText(naText string) *HtmlWriter {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *HtmlWriter) SetNewLine(newLine string) *HtmlWriter {
	w.writer = w.writer.SetNewLine(newLine)
	return w
}

func (w *HtmlWriter) SetIndent(indent string) *HtmlWriter {
	w.writer = w.writer.SetIndent(indent)
	return w
}

func (w *HtmlWriter) SetWriter(writer io.Writer) *HtmlWriter {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *HtmlWriter) SetDatatables(datatables bool) *HtmlWriter {
	w.writer = w.writer.SetDatatables(datatables)
	return w
}

func (w *HtmlWriter) Write() error {
	return w.writer.Write()
}

////////////////////////			MARKDOWN WRITER

// MarkdownWriter writes Markdown data from a DataFrame. Create one with WriteMarkdown,
// configure it with the Set methods, then call Write.
type MarkdownWriter struct {
	writer *encio.MarkDownWriter
}

// WriteMarkdown starts a Markdown write of the frame.
func (df DataFrame) WriteMarkdown() *MarkdownWriter {
	return &MarkdownWriter{
		writer: encio.NewMarkDownWriter().SetIoData(df.ToIoData()),
	}
}

func (w *MarkdownWriter) SetHeader(header bool) *MarkdownWriter {
	w.writer = w.writer.SetHeader(header)
	return w
}

func (w *MarkdownWriter) SetIndex(index bool) *MarkdownWriter {
	w.writer = w.writer.SetIndex(index)
	return w
}

func (w *MarkdownWriter) SetPath(path string) *MarkdownWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *MarkdownWriter) SetNaText(naText string) *MarkdownWriter {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *MarkdownWriter) SetWriter(writer io.Writer) *MarkdownWriter {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *MarkdownWriter) Write() error {
	return w.writer.Write()
}

////////////////////////			PARQUET READER

// ParquetReader reads Parquet data into a DataFrame. Create one with ReadParquet,
// configure it with the Set methods, then call Read.
type ParquetReader struct {
	reader *encio.ParquetReader
}

// ReadParquet creates a Parquet reader with the given context.
func ReadParquet(ctx *enchanter.Context) *ParquetReader {
	return &ParquetReader{
		reader: encio.NewParquetReader(ctx),
	}
}

func (r *ParquetReader) SetPath(path string) *ParquetReader {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *ParquetReader) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			PARQUET WRITER

// ParquetWriter writes Parquet data from a DataFrame. Create one with WriteParquet,
// configure it with the Set methods, then call Write.
type ParquetWriter struct {
	writer *encio.ParquetWriter
}

// WriteParquet starts a Parquet write of the frame.
func (df DataFrame) WriteParquet() *ParquetWriter {
	return &ParquetWriter{
		writer: encio.NewParquetWriter().SetIoData(df.ToIoData()),
	}
}

func (w *ParquetWriter) SetPath(path string) *ParquetWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *ParquetWriter) Write() error {
	return w.writer.Write()
}

////////////////////////			ARROW IPC READER

// ArrowIPCReader reads Arrow IPC data into a DataFrame. Create one with ReadArrowIPC,
// configure it with the Set methods, then call Read.
type ArrowIPCReader struct {
	reader *encio.ArrowIPCReader
}

// ReadArrowIPC creates a Arrow IPC reader with the given context.
func ReadArrowIPC(ctx *enchanter.Context) *ArrowIPCReader {
	return &ArrowIPCReader{
		reader: encio.NewArrowIPCReader(ctx),
	}
}

func (r *ArrowIPCReader) SetPath(path string) *ArrowIPCReader {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *ArrowIPCReader) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			ARROW IPC WRITER

// ArrowIPCWriter writes Arrow IPC data from a DataFrame. Create one with WriteArrowIPC,
// configure it with the Set methods, then call Write.
type ArrowIPCWriter struct {
	writer *encio.ArrowIPCWriter
}

// WriteArrowIPC starts a Arrow IPC write of the frame.
func (df DataFrame) WriteArrowIPC() *ArrowIPCWriter {
	return &ArrowIPCWriter{
		writer: encio.NewArrowIPCWriter().SetIoData(df.ToIoData()),
	}
}

func (w *ArrowIPCWriter) SetPath(path string) *ArrowIPCWriter {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *ArrowIPCWriter) Write() error {
	return w.writer.Write()
}
