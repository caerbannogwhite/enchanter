// Round-trip a DataFrame through Parquet: types and nulls survive, unlike CSV.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/dataframe"
	"github.com/caerbannogwhite/enchanter/series"
)

func main() {
	ctx := enchanter.NewContext()

	df := dataframe.NewDataFrame(ctx).
		AddSeries("name", series.NewSeriesString([]string{"Alice", "Bob", "Charlie", "Dana"}, nil, false, ctx)).
		AddSeries("age", series.NewSeriesInt64([]int64{29, 31, 0, 25}, []bool{false, false, true, false}, false, ctx)).
		AddSeries("score", series.NewSeriesFloat64([]float64{7.5, 8.25, 9.0, 6.75}, nil, false, ctx))
	if df.Err() != nil {
		fmt.Fprintln(os.Stderr, df.Err())
		os.Exit(1)
	}

	dir, err := os.MkdirTemp("", "enchanter-parquet-example")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "people.parquet")

	if err := df.ToParquet().SetPath(path).Write(); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}

	back := dataframe.NewDataFrame(ctx).FromParquet().SetPath(path).Read()
	if back.Err() != nil {
		fmt.Fprintln(os.Stderr, "read:", back.Err())
		os.Exit(1)
	}

	// The null in "age" and all column types come back intact.
	back.PPrint(dataframe.NewPPrintParams())
}
