package dataframe_test

import (
	"fmt"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/dataframe"
	"github.com/caerbannogwhite/enchanter/series"
)

func ExampleNewDataFrame() {
	ctx := enchanter.NewContext()

	df := dataframe.NewDataFrame(ctx).
		AddSeries("name", series.NewSeriesString([]string{"Alice", "Bob", "Carol"}, nil, false, ctx)).
		AddSeries("age", series.NewSeriesInt64([]int64{29, 35, 31}, nil, false, ctx))

	adults := df.Filter(df.C("age").Gt(int64(30)))
	if adults.Err() != nil {
		fmt.Println(adults.Err())
		return
	}

	fmt.Println(adults.NRows())
	fmt.Println(adults.C("name").DataAsString())
	// Output:
	// 2
	// [Bob Carol]
}
