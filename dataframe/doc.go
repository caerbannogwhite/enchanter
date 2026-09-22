// Package dataframe implements the DataFrame: an ordered collection of named,
// equally sized series with relational-style operations.
//
// Read one from a file with the package-level reader constructors
// ([ReadCsv], [ReadParquet], ...), or construct one column by column with
// [NewDataFrame]. Build pipelines by chaining:
// Select, Filter, GroupBy + Agg, Join (inner, left, right, outer), OrderBy,
// Slice and TakeIndices. Operations never mutate their receiver; each call
// returns a new DataFrame value, and errors travel with the frame (check
// Err at the end of a chain).
//
//	df := dataframe.ReadCsv(ctx).
//		SetPath("people.csv").Read().
//		Filter(df.Col("age").Gt(int64(30))).
//		GroupBy("city").
//		Agg(dataframe.Count()).
//		Run()
//
// Reading and writing files (CSV, XLSX, XPT, JSON, HTML, Markdown, Parquet,
// Arrow IPC, and read-only SAS7BDAT) is exposed through builder chains such
// as ReadCsv/WriteCsv and ReadParquet/WriteParquet. A DataFrame also converts
// to and from an Apache Arrow record batch via ToArrowRecord and
// [NewDataFrameFromArrowRecord].
package dataframe
