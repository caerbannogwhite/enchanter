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

type csvWriterWrapper struct {
	writer *encio.CsvWriter
}

func (df DataFrame) ToCsv() *csvWriterWrapper {
	return &csvWriterWrapper{
		writer: encio.NewCsvWriter().SetIoData(df.ToIoData()),
	}
}

func (w *csvWriterWrapper) SetDelimiter(delimiter rune) *csvWriterWrapper {
	w.writer = w.writer.SetDelimiter(delimiter)
	return w
}

func (w *csvWriterWrapper) SetHeader(header bool) *csvWriterWrapper {
	w.writer = w.writer.SetHeader(header)
	return w
}

func (w *csvWriterWrapper) SetFormat(format bool) *csvWriterWrapper {
	w.writer = w.writer.SetFormat(format)
	return w
}

func (w *csvWriterWrapper) SetPath(path string) *csvWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *csvWriterWrapper) SetNaText(naText string) *csvWriterWrapper {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *csvWriterWrapper) SetEol(eol string) *csvWriterWrapper {
	w.writer = w.writer.SetEol(eol)
	return w
}

func (w *csvWriterWrapper) SetQuote(quote string) *csvWriterWrapper {
	w.writer = w.writer.SetQuote(quote)
	return w
}

func (w *csvWriterWrapper) SetQuoting(quoting encio.CsvQuotingType) *csvWriterWrapper {
	w.writer = w.writer.SetQuoting(quoting)
	return w
}

func (w *csvWriterWrapper) SetWriter(writer io.Writer) *csvWriterWrapper {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *csvWriterWrapper) Write() error {
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

type jsonWriterWrapper struct {
	writer *encio.JsonWriter
}

func (df DataFrame) ToJson() *jsonWriterWrapper {
	return &jsonWriterWrapper{
		writer: encio.NewJsonWriter().SetIoData(df.ToIoData()),
	}
}

func (w *jsonWriterWrapper) SetPath(path string) *jsonWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *jsonWriterWrapper) SetNewLine(newLine string) *jsonWriterWrapper {
	w.writer = w.writer.SetNewLine(newLine)
	return w
}

func (w *jsonWriterWrapper) SetIndent(indent string) *jsonWriterWrapper {
	w.writer = w.writer.SetIndent(indent)
	return w
}

func (w *jsonWriterWrapper) SetWriter(writer io.Writer) *jsonWriterWrapper {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *jsonWriterWrapper) Write() error {
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

type xptWriterWrapper struct {
	writer *encio.XptWriter
}

func (df DataFrame) ToXpt() *xptWriterWrapper {
	return &xptWriterWrapper{
		writer: encio.NewXptWriter().SetIoData(df.ToIoData()),
	}
}

func (w *xptWriterWrapper) SetVersion(version encio.XptVersionType) *xptWriterWrapper {
	w.writer = w.writer.SetVersion(version)
	return w
}

func (w *xptWriterWrapper) SetByteOrder(byteOrder binary.ByteOrder) *xptWriterWrapper {
	w.writer = w.writer.SetByteOrder(byteOrder)
	return w
}

func (w *xptWriterWrapper) SetPath(path string) *xptWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *xptWriterWrapper) SetWriter(writer io.Writer) *xptWriterWrapper {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *xptWriterWrapper) Write() error {
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

type xlsxWriterWrapper struct {
	writer *encio.XlsxWriter
}

func (df DataFrame) ToXlsx() *xlsxWriterWrapper {
	return &xlsxWriterWrapper{
		writer: encio.NewXlsxWriter().SetIoData(df.ToIoData()),
	}
}

func (w *xlsxWriterWrapper) SetPath(path string) *xlsxWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *xlsxWriterWrapper) SetSheet(sheet string) *xlsxWriterWrapper {
	w.writer = w.writer.SetSheet(sheet)
	return w
}

func (w *xlsxWriterWrapper) SetNaText(naText string) *xlsxWriterWrapper {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *xlsxWriterWrapper) SetWriter(writer io.Writer) *xlsxWriterWrapper {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *xlsxWriterWrapper) Write() error {
	return w.writer.Write()
}

////////////////////////			HTML WRITER

type htmlWriterWrapper struct {
	writer *encio.HtmlWriter
}

func (df DataFrame) ToHtml() *htmlWriterWrapper {
	return &htmlWriterWrapper{
		writer: encio.NewHtmlWriter().SetIoData(df.ToIoData()),
	}
}

func (w *htmlWriterWrapper) SetPath(path string) *htmlWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *htmlWriterWrapper) SetNaText(naText string) *htmlWriterWrapper {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *htmlWriterWrapper) SetNewLine(newLine string) *htmlWriterWrapper {
	w.writer = w.writer.SetNewLine(newLine)
	return w
}

func (w *htmlWriterWrapper) SetIndent(indent string) *htmlWriterWrapper {
	w.writer = w.writer.SetIndent(indent)
	return w
}

func (w *htmlWriterWrapper) SetWriter(writer io.Writer) *htmlWriterWrapper {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *htmlWriterWrapper) SetDatatables(datatables bool) *htmlWriterWrapper {
	w.writer = w.writer.SetDatatables(datatables)
	return w
}

func (w *htmlWriterWrapper) Write() error {
	return w.writer.Write()
}

////////////////////////			MARKDOWN WRITER

type markDownWriterWrapper struct {
	writer *encio.MarkDownWriter
}

func (df DataFrame) ToMarkDown() *markDownWriterWrapper {
	return &markDownWriterWrapper{
		writer: encio.NewMarkDownWriter().SetIoData(df.ToIoData()),
	}
}

func (w *markDownWriterWrapper) SetHeader(header bool) *markDownWriterWrapper {
	w.writer = w.writer.SetHeader(header)
	return w
}

func (w *markDownWriterWrapper) SetIndex(index bool) *markDownWriterWrapper {
	w.writer = w.writer.SetIndex(index)
	return w
}

func (w *markDownWriterWrapper) SetPath(path string) *markDownWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *markDownWriterWrapper) SetNaText(naText string) *markDownWriterWrapper {
	w.writer = w.writer.SetNaText(naText)
	return w
}

func (w *markDownWriterWrapper) SetWriter(writer io.Writer) *markDownWriterWrapper {
	w.writer = w.writer.SetWriter(writer)
	return w
}

func (w *markDownWriterWrapper) Write() error {
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

type parquetWriterWrapper struct {
	writer *encio.ParquetWriter
}

func (df DataFrame) ToParquet() *parquetWriterWrapper {
	return &parquetWriterWrapper{
		writer: encio.NewParquetWriter().SetIoData(df.ToIoData()),
	}
}

func (w *parquetWriterWrapper) SetPath(path string) *parquetWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *parquetWriterWrapper) Write() error {
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

type arrowIPCWriterWrapper struct {
	writer *encio.ArrowIPCWriter
}

func (df DataFrame) ToArrowIPC() *arrowIPCWriterWrapper {
	return &arrowIPCWriterWrapper{
		writer: encio.NewArrowIPCWriter().SetIoData(df.ToIoData()),
	}
}

func (w *arrowIPCWriterWrapper) SetPath(path string) *arrowIPCWriterWrapper {
	w.writer = w.writer.SetPath(path)
	return w
}

func (w *arrowIPCWriterWrapper) Write() error {
	return w.writer.Write()
}
