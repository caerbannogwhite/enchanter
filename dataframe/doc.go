// Package dataframe implements the DataFrame: an ordered collection of named,
// equally sized series with relational-style operations.
//
// Construct one with [NewDataFrame] and build pipelines by chaining:
// Select, Filter, GroupBy + Agg, Join (inner, left, right, outer), OrderBy
// and Take. Operations never mutate their receiver; each call returns a new
// DataFrame value, and errors travel with the frame (check IsErrored /
// GetError at the end of a chain).
//
//	df := dataframe.NewDataFrame(ctx).
//		FromCsv().SetPath("people.csv").Read().
//		Filter(df.C("age").Gt(int64(30))).
//		GroupBy("city").
//		Agg(dataframe.Count()).
//		Run()
//
// Reading and writing files (CSV, XLSX, XPT, JSON, HTML, Markdown, Parquet,
// Arrow IPC) is exposed through builder chains such as FromCsv/ToCsv and
// FromParquet/ToParquet. A DataFrame also converts to and from an Apache
// Arrow record batch via ToArrowRecord and [NewDataFrameFromArrowRecord].
package dataframe
