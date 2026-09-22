package io

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/series"
	"github.com/kshedden/datareader"
)

// Sas7bdatReader reads a SAS dataset in SAS7BDAT format into an IoData.
//
// Parsing is delegated to github.com/kshedden/datareader (BSD-3-Clause); this
// type adapts its output to enchanter series. SAS has two column types:
// numeric (always float64, missing values become nulls) and character
// (fixed-width, right-padded with spaces — trimmed by default, see
// SetTrimStrings). Columns carrying a date format that datareader recognizes
// are converted to Times when SetConvertDates is left on; other date formats
// stay numeric, holding days or seconds since the SAS epoch (1960-01-01).
//
// Writing SAS7BDAT is not supported. Use the XPT writer for a SAS-readable
// output format.
type Sas7bdatReader struct {
	path         string
	reader       io.ReadSeeker
	ctx          *enchanter.Context
	trimStrings  bool
	convertDates bool
}

func NewSas7bdatReader(ctx *enchanter.Context) *Sas7bdatReader {
	return &Sas7bdatReader{
		ctx:          ctx,
		trimStrings:  true,
		convertDates: true,
	}
}

func (r *Sas7bdatReader) SetPath(path string) *Sas7bdatReader {
	r.path = path
	return r
}

// SetReader sets the source. SAS7BDAT is a paged format, so the source must
// seek; a plain io.Reader is not enough.
func (r *Sas7bdatReader) SetReader(reader io.ReadSeeker) *Sas7bdatReader {
	r.reader = reader
	return r
}

// SetTrimStrings controls whether the fixed-width right-padding of SAS
// character values is trimmed. Default true.
func (r *Sas7bdatReader) SetTrimStrings(trim bool) *Sas7bdatReader {
	r.trimStrings = trim
	return r
}

// SetConvertDates controls whether columns with a recognized SAS date format
// are converted to Times. Default true; unrecognized date formats stay
// numeric either way.
func (r *Sas7bdatReader) SetConvertDates(convert bool) *Sas7bdatReader {
	r.convertDates = convert
	return r
}

func (r *Sas7bdatReader) Read() *IoData {
	if r.path != "" {
		file, err := os.OpenFile(r.path, os.O_RDONLY, 0666)
		if err != nil {
			return &IoData{Error: err}
		}
		defer file.Close()
		r.reader = file
	}

	if r.reader == nil {
		return &IoData{Error: fmt.Errorf("Sas7bdatReader: no reader specified")}
	}

	if r.ctx == nil {
		return &IoData{Error: fmt.Errorf("Sas7bdatReader: no context specified")}
	}

	sas, err := datareader.NewSAS7BDATReader(r.reader)
	if err != nil {
		return &IoData{Error: fmt.Errorf("Sas7bdatReader: %w", err)}
	}
	sas.TrimStrings = r.trimStrings
	sas.ConvertDates = r.convertDates

	iod := NewIoData(r.ctx)

	cols, err := sas.Read(-1)
	if err == io.EOF {
		// A dataset with zero observations: emit the columns empty, so the
		// schema still comes through.
		names := sas.ColumnNames()
		for i, colType := range sas.ColumnTypes() {
			switch colType {
			case datareader.SASStringType:
				iod.AddSeries(series.NewSeriesString(nil, nil, false, r.ctx),
					SeriesMeta{Name: names[i], Type: meta.StringType})
			default:
				iod.AddSeries(series.NewSeriesFloat64(nil, nil, false, r.ctx),
					SeriesMeta{Name: names[i], Type: meta.Float64Type})
			}
		}
		return iod
	}
	if err != nil {
		return &IoData{Error: fmt.Errorf("Sas7bdatReader: %w", err)}
	}

	for _, col := range cols {
		// datareader hands over freshly allocated slices, so the series can
		// take ownership without copying. A nil Missing() simply yields a
		// non-nullable series.
		missing := col.Missing()

		switch data := col.Data().(type) {
		case []float64:
			// A SAS numeric holds missing values as NaN - SAS has no
			// non-missing NaN - so NaNs are folded into the null mask
			// whether or not datareader flagged them.
			for i, v := range data {
				if math.IsNaN(v) {
					if missing == nil {
						missing = make([]bool, len(data))
					}
					missing[i] = true
				}
			}
			iod.AddSeries(series.NewSeriesFloat64(data, missing, false, r.ctx),
				SeriesMeta{Name: col.Name, Type: meta.Float64Type})
		case []string:
			iod.AddSeries(series.NewSeriesString(data, missing, false, r.ctx),
				SeriesMeta{Name: col.Name, Type: meta.StringType})
		case []time.Time:
			iod.AddSeries(series.NewSeriesTime(data, missing, false, r.ctx),
				SeriesMeta{Name: col.Name, Type: meta.TimeType})
		default:
			return &IoData{Error: fmt.Errorf(
				"Sas7bdatReader: unsupported column type %T for column \"%s\"", data, col.Name)}
		}
	}

	return iod
}
