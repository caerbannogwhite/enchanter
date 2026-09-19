package dataframe

import (
	"encoding/binary"
	"io"

	"github.com/caerbannogwhite/enchanter"
	encio "github.com/caerbannogwhite/enchanter/io"
	"github.com/caerbannogwhite/enchanter/meta"
)

func FromIoData(iod *encio.IoData) DataFrame {
	df := NewDataFrame(iod.GetContext())

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

type csvReaderWrapper struct {
	reader *encio.CsvReader
}

func (df DataFrame) FromCsv() *csvReaderWrapper {
	return &csvReaderWrapper{
		reader: encio.NewCsvReader(df.ctx),
	}
}

func (r *csvReaderWrapper) SetHeader(header bool) *csvReaderWrapper {
	r.reader = r.reader.SetHeader(header)
	return r
}

func (r *csvReaderWrapper) SetDelimiter(delimiter rune) *csvReaderWrapper {
	r.reader = r.reader.SetDelimiter(delimiter)
	return r
}

func (r *csvReaderWrapper) SetGuessDataTypeLen(guessDataTypeLen int) *csvReaderWrapper {
	r.reader = r.reader.SetGuessDataTypeLen(guessDataTypeLen)
	return r
}

func (r *csvReaderWrapper) SetRows(rows int) *csvReaderWrapper {
	r.reader = r.reader.SetRows(rows)
	return r
}

func (r *csvReaderWrapper) SetPath(path string) *csvReaderWrapper {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *csvReaderWrapper) SetNullValues(nullValues bool) *csvReaderWrapper {
	r.reader = r.reader.SetNullValues(nullValues)
	return r
}

func (r *csvReaderWrapper) SetReader(reader io.Reader) *csvReaderWrapper {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *csvReaderWrapper) SetSchema(schema *meta.Schema) *csvReaderWrapper {
	r.reader = r.reader.SetSchema(schema)
	return r
}

func (r *csvReaderWrapper) SetContext(ctx *enchanter.Context) *csvReaderWrapper {
	r.reader = r.reader.SetContext(ctx)
	return r
}

func (r *csvReaderWrapper) Read() DataFrame {
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

type jsonReaderWrapper struct {
	reader *encio.JsonReader
}

func (df DataFrame) FromJson() *jsonReaderWrapper {
	return &jsonReaderWrapper{
		reader: encio.NewJsonReader(df.ctx),
	}
}

func (r *jsonReaderWrapper) SetPath(path string) *jsonReaderWrapper {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *jsonReaderWrapper) SetReader(reader io.Reader) *jsonReaderWrapper {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *jsonReaderWrapper) SetSchema(schema *meta.Schema) *jsonReaderWrapper {
	r.reader = r.reader.SetSchema(schema)
	return r
}

func (r *jsonReaderWrapper) Read() DataFrame {
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

type xptReaderWrapper struct {
	reader *encio.XptReader
}

func (df DataFrame) FromXpt() *xptReaderWrapper {
	return &xptReaderWrapper{
		reader: encio.NewXptReader(df.ctx),
	}
}

func (r *xptReaderWrapper) SetMaxObservations(maxObservations int) *xptReaderWrapper {
	r.reader = r.reader.SetMaxObservations(maxObservations)
	return r
}

func (r *xptReaderWrapper) SetVersion(version encio.XptVersionType) *xptReaderWrapper {
	r.reader = r.reader.SetVersion(version)
	return r
}

func (r *xptReaderWrapper) SetByteOrder(byteOrder binary.ByteOrder) *xptReaderWrapper {
	r.reader = r.reader.SetByteOrder(byteOrder)
	return r
}

func (r *xptReaderWrapper) SetPath(path string) *xptReaderWrapper {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *xptReaderWrapper) SetReader(reader io.Reader) *xptReaderWrapper {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *xptReaderWrapper) Read() DataFrame {
	iod := r.reader.Read()
	return FromIoData(iod)
}

////////////////////////			SAS7BDAT READER

type sas7bdatReaderWrapper struct {
	reader *encio.Sas7bdatReader
}

func (df DataFrame) FromSas7bdat() *sas7bdatReaderWrapper {
	return &sas7bdatReaderWrapper{
		reader: encio.NewSas7bdatReader(df.ctx),
	}
}

func (r *sas7bdatReaderWrapper) SetPath(path string) *sas7bdatReaderWrapper {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *sas7bdatReaderWrapper) SetReader(reader io.ReadSeeker) *sas7bdatReaderWrapper {
	r.reader = r.reader.SetReader(reader)
	return r
}

func (r *sas7bdatReaderWrapper) SetTrimStrings(trim bool) *sas7bdatReaderWrapper {
	r.reader = r.reader.SetTrimStrings(trim)
	return r
}

func (r *sas7bdatReaderWrapper) SetConvertDates(convert bool) *sas7bdatReaderWrapper {
	r.reader = r.reader.SetConvertDates(convert)
	return r
}

func (r *sas7bdatReaderWrapper) Read() DataFrame {
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

type xlsxReaderWrapper struct {
	reader *encio.XlsxReader
}

func (df DataFrame) FromXlsx() *xlsxReaderWrapper {
	return &xlsxReaderWrapper{
		reader: encio.NewXlsxReader(df.ctx),
	}
}

func (r *xlsxReaderWrapper) SetPath(path string) *xlsxReaderWrapper {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *xlsxReaderWrapper) SetSheet(sheet string) *xlsxReaderWrapper {
	r.reader = r.reader.SetSheet(sheet)
	return r
}

func (r *xlsxReaderWrapper) SetHeader(header int) *xlsxReaderWrapper {
	r.reader = r.reader.SetHeader(header)
	return r
}

func (r *xlsxReaderWrapper) SetRows(rows int) *xlsxReaderWrapper {
	r.reader = r.reader.SetRows(rows)
	return r
}

func (r *xlsxReaderWrapper) SetGuessDataTypeLen(guessDataTypeLen int) *xlsxReaderWrapper {
	r.reader = r.reader.SetGuessDataTypeLen(guessDataTypeLen)
	return r
}

func (r *xlsxReaderWrapper) SetNullValues(nullValues bool) *xlsxReaderWrapper {
	r.reader = r.reader.SetNullValues(nullValues)
	return r
}

func (r *xlsxReaderWrapper) SetSchema(schema *meta.Schema) *xlsxReaderWrapper {
	r.reader = r.reader.SetSchema(schema)
	return r
}

func (r *xlsxReaderWrapper) Read() DataFrame {
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

type parquetReaderWrapper struct {
	reader *encio.ParquetReader
}

func (df DataFrame) FromParquet() *parquetReaderWrapper {
	return &parquetReaderWrapper{
		reader: encio.NewParquetReader(df.ctx),
	}
}

func (r *parquetReaderWrapper) SetPath(path string) *parquetReaderWrapper {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *parquetReaderWrapper) Read() DataFrame {
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

type arrowIPCReaderWrapper struct {
	reader *encio.ArrowIPCReader
}

func (df DataFrame) FromArrowIPC() *arrowIPCReaderWrapper {
	return &arrowIPCReaderWrapper{
		reader: encio.NewArrowIPCReader(df.ctx),
	}
}

func (r *arrowIPCReaderWrapper) SetPath(path string) *arrowIPCReaderWrapper {
	r.reader = r.reader.SetPath(path)
	return r
}

func (r *arrowIPCReaderWrapper) Read() DataFrame {
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
