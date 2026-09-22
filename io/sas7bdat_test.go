package io

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/caerbannogwhite/enchanter/series"
)

// The fixture pair comes from github.com/kshedden/datareader
// (test_files/data/test1.sas7bdat and test1.csv, BSD-3-Clause): the CSV is an
// independently produced reference for the SAS file's contents, so the test
// checks the whole read path against ground truth this library did not
// generate. 100 columns in a repeating float/string/float/float pattern, with
// missing values in both numeric and character columns; two of the numeric
// columns carry the SAS MMDDYY date format.
func sas7bdatReference(t *testing.T) ([]string, [][]string) {
	t.Helper()
	f, err := os.Open(filepath.Join(testDataFolder, "sas7bdat_test1_ref.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return records[0], records[1:]
}

// With date conversion off, every column is numeric or character and must
// match the reference cell for cell.
func Test_Sas7bdat_ReadAgainstReference(t *testing.T) {
	header, rows := sas7bdatReference(t)

	iod := FromSas7bdat(ctx).
		SetPath(filepath.Join(testDataFolder, "sas7bdat_test1.sas7bdat")).
		SetConvertDates(false).
		Read()
	if iod.Error != nil {
		t.Fatal(iod.Error)
	}

	if got, want := len(iod.Series), len(header); got != want {
		t.Fatalf("column count: got %d, want %d", got, want)
	}
	for j, meta := range iod.SeriesMeta {
		if meta.Name != header[j] {
			t.Fatalf("column %d name: got %q, want %q", j, meta.Name, header[j])
		}
	}
	if got, want := iod.NRows(), len(rows); got != want {
		t.Fatalf("row count: got %d, want %d", got, want)
	}

	for j, s := range iod.Series {
		switch s := s.(type) {
		case series.Float64s:
			for i := range rows {
				cell := rows[i][j]
				if cell == "" {
					if !s.IsNull(i) {
						t.Fatalf("col %d row %d: reference is missing, got %v", j, i, s.Get(i))
					}
					continue
				}
				want, err := strconv.ParseFloat(cell, 64)
				if err != nil {
					t.Fatalf("col %d row %d: reference %q is not numeric but the column read as Float64s", j, i, cell)
				}
				if s.IsNull(i) {
					t.Fatalf("col %d row %d: got null, reference has %v", j, i, want)
				}
				// The reference stores rounded decimals; integers are exact.
				if got := s.Float64s()[i]; math.Abs(got-want) > 5e-4 {
					t.Fatalf("col %d row %d: got %v, want %v", j, i, got, want)
				}
			}

		case series.Strings:
			for i := range rows {
				cell := rows[i][j]
				got := ""
				if !s.IsNull(i) {
					got = s.Get(i).(string)
				}
				if got != cell {
					t.Fatalf("col %d row %d: got %q, want %q", j, i, got, cell)
				}
			}

		default:
			t.Fatalf("col %d (%s): unexpected series type %T", j, header[j], s)
		}
	}

	// Pin a few concrete cells so a systematic offset cannot slip through the
	// tolerance: the first reference row starts 0.636, pear, 84, 2170.
	if got := iod.Series[0].(series.Float64s).Float64s()[0]; math.Abs(got-0.636) > 5e-4 {
		t.Fatalf("Column1[0]: got %v, want 0.636", got)
	}
	if got := iod.Series[1].Get(0).(string); got != "pear" {
		t.Fatalf("Column2[0]: got %q, want \"pear\"", got)
	}
	if s := iod.Series[7].(series.Float64s); !s.IsNull(0) {
		t.Fatalf("Column8[0]: want null (missing in the reference), got %v", s.Float64s()[0])
	}
}

// With the default date conversion on, the two MMDDYY columns become Times
// holding the SAS epoch (1960-01-01) plus the reference's day count; every
// other column is unaffected.
func Test_Sas7bdat_ConvertDates(t *testing.T) {
	header, rows := sas7bdatReference(t)

	iod := FromSas7bdat(ctx).
		SetPath(filepath.Join(testDataFolder, "sas7bdat_test1.sas7bdat")).
		Read()
	if iod.Error != nil {
		t.Fatal(iod.Error)
	}

	epoch := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)
	var timeCols []int
	for j, s := range iod.Series {
		if _, ok := s.(series.Times); ok {
			timeCols = append(timeCols, j)
		}
	}
	if len(timeCols) != 2 {
		t.Fatalf("expected exactly 2 date columns, got %d (%v)", len(timeCols), timeCols)
	}

	for _, j := range timeCols {
		s := iod.Series[j].(series.Times)
		for i := range rows {
			cell := rows[i][j]
			if cell == "" {
				if !s.IsNull(i) {
					t.Fatalf("col %d (%s) row %d: reference is missing, got %v", j, header[j], i, s.Get(i))
				}
				continue
			}
			days, err := strconv.Atoi(cell)
			if err != nil {
				t.Fatalf("col %d row %d: reference %q is not a day count", j, i, cell)
			}
			want := epoch.AddDate(0, 0, days)
			if got := s.Times()[i]; !got.Equal(want) {
				t.Fatalf("col %d (%s) row %d: got %v, want %v (%d days after the SAS epoch)",
					j, header[j], i, got, want, days)
			}
		}
	}
}
