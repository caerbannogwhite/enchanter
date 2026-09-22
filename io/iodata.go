package io

import (
	"fmt"
	"time"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/series"
)

type FileFormat string

const (
	FILE_FORMAT_XPT      FileFormat = "XPT"
	FILE_FORMAT_CSV      FileFormat = "CSV"
	FILE_FORMAT_XLSX     FileFormat = "XLSX"
	FILE_FORMAT_JSON     FileFormat = "JSON"
	FILE_FORMAT_HTML     FileFormat = "HTML"
	FILE_FORMAT_SAS7BDAT FileFormat = "SAS7BDAT"
)

type IoData struct {
	ctx *enchanter.Context

	FileMeta   FileMeta
	SeriesMeta []SeriesMeta
	Series     []series.Series
	Error      error
}

type FileMeta struct {
	FileSize   int64
	FileName   string
	FilePath   string
	FileExt    string
	FileFormat FileFormat

	Label          string
	Created        time.Time
	LastModified   time.Time
	SasLibVersion  string
	SasDataVersion string
	SasOs          string
	SasDsName      string
}

func (fileMeta *FileMeta) String() string {
	return fmt.Sprintf("FileMeta{FileName: %s, FilePath: %s, FileSize: %d, FileExt: %s, FileFormat: %s, Label: %s, Created: %s, LastModified: %s, SasLibVersion: %s, SasDataVersion: %s, SasOs: %s, SasDsName: %s}",
		fileMeta.FileName,
		fileMeta.FilePath,
		fileMeta.FileSize,
		fileMeta.FileExt,
		fileMeta.FileFormat,
		fileMeta.Label,
		fileMeta.Created,
		fileMeta.LastModified,
		fileMeta.SasLibVersion,
		fileMeta.SasDataVersion,
		fileMeta.SasOs,
		fileMeta.SasDsName,
	)
}

func (fileMeta *FileMeta) PrettyPrint() {
	fmt.Println("File Meta")
	fmt.Println("--------------------------------")
	fmt.Println("File Name: ", fileMeta.FileName)
	fmt.Println("File Path: ", fileMeta.FilePath)
	fmt.Println("File Size: ", fileMeta.FileSize)
	fmt.Println("File Ext: ", fileMeta.FileExt)
	fmt.Println("File Format: ", fileMeta.FileFormat)
	fmt.Println("Label: ", fileMeta.Label)
	fmt.Println("Created: ", fileMeta.Created)
	fmt.Println("Last Modified: ", fileMeta.LastModified)
	fmt.Println("Sas Lib Version: ", fileMeta.SasLibVersion)
	fmt.Println("Sas Data Version: ", fileMeta.SasDataVersion)
	fmt.Println("Sas Os: ", fileMeta.SasOs)
	fmt.Println("Sas Ds Name: ", fileMeta.SasDsName)
}

type SeriesMeta struct {
	Format      string
	Label       string
	Length      int
	KeySequence int
	Name        string
	Type        meta.BaseType
}

func (seriesMeta *SeriesMeta) String() string {
	return fmt.Sprintf("SeriesMeta{Format: %s, Label: %s, Length: %d, KeySequence: %d, Name: %s, Type: %s}",
		seriesMeta.Format,
		seriesMeta.Label,
		seriesMeta.Length,
		seriesMeta.KeySequence,
		seriesMeta.Name,
		seriesMeta.Type.String(),
	)
}

func (seriesMeta *SeriesMeta) PrettyPrint() {
	fmt.Println("Series Meta")
	fmt.Println("--------------------------------")
	fmt.Println("Format: ", seriesMeta.Format)
	fmt.Println("Label: ", seriesMeta.Label)
	fmt.Println("Length: ", seriesMeta.Length)
	fmt.Println("Key Sequence: ", seriesMeta.KeySequence)
	fmt.Println("Name: ", seriesMeta.Name)
	fmt.Println("Type: ", seriesMeta.Type)
}

func (iod *IoData) AddSeries(series series.Series, meta SeriesMeta) {
	iod.Series = append(iod.Series, series)
	iod.SeriesMeta = append(iod.SeriesMeta, meta)
}

func (iod *IoData) At(i int) series.Series {
	return iod.Series[i]
}

func (iod *IoData) ByName(name string) series.Series {
	for i, meta := range iod.SeriesMeta {
		if meta.Name == name {
			return iod.Series[i]
		}
	}
	return nil
}

func (iod *IoData) SeriesMetaAt(i int) SeriesMeta {
	return iod.SeriesMeta[i]
}

func (iod *IoData) Types() []meta.BaseType {
	types := make([]meta.BaseType, len(iod.Series))
	for i, series := range iod.Series {
		types[i] = series.Type()
	}
	return types
}

func (iod *IoData) Context() *enchanter.Context {
	return iod.ctx
}

func (iod *IoData) NRows() int {
	if len(iod.Series) >= 1 {
		return iod.Series[0].Len()
	}
	return 0
}

func (iod *IoData) NCols() int {
	return len(iod.Series)
}

func (iod *IoData) ToCsv() *CsvWriter {
	return NewCsvWriter().SetIoData(iod)
}

func (iod *IoData) ToHtml() *HtmlWriter {
	return NewHtmlWriter().SetIoData(iod)
}

func (iod *IoData) ToJson() *JsonWriter {
	return NewJsonWriter().SetIoData(iod)
}

func (iod *IoData) ToMarkdown() *MarkDownWriter {
	return NewMarkDownWriter().SetIoData(iod)
}

func (iod *IoData) ToXlsx() *XlsxWriter {
	return NewXlsxWriter().SetIoData(iod)
}

func (iod *IoData) ToXpt() *XptWriter {
	return NewXptWriter().SetIoData(iod)
}

func NewIoData(ctx *enchanter.Context) *IoData {
	return &IoData{
		ctx:        ctx,
		Series:     make([]series.Series, 0),
		SeriesMeta: make([]SeriesMeta, 0),
		FileMeta:   FileMeta{},
	}
}

func FromCsv(ctx *enchanter.Context) *CsvReader {
	return NewCsvReader(ctx)
}

func FromJson(ctx *enchanter.Context) *JsonReader {
	return NewJsonReader(ctx)
}

func FromXlsx(ctx *enchanter.Context) *XlsxReader {
	return NewXlsxReader(ctx)
}

func FromXpt(ctx *enchanter.Context) *XptReader {
	return NewXptReader(ctx)
}

func FromSas7bdat(ctx *enchanter.Context) *Sas7bdatReader {
	return NewSas7bdatReader(ctx)
}
