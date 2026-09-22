// Package enchanter is a data-wrangling library for Go, in the spirit of
// pandas and Polars in Python and dplyr in R.
//
// The entry point is a [Context] (see [NewContext]), which carries the string
// interning pool and the Apache Arrow memory allocator shared by every series
// and dataframe created from it:
//
//	ctx := enchanter.NewContext()
//	df := dataframe.ReadCsv(ctx).
//		SetReader(r).Read().
//		GroupBy("department").
//		Agg(dataframe.Mean("salary")).
//		Run()
//
// The functionality lives in the subpackages:
//
//   - series: typed, nullable columns (Bool, Int, Int64, Float64, String,
//     Time, Duration) with filtering, grouping, sorting and arithmetic.
//   - dataframe: the DataFrame type — select, filter, group by, join,
//     sort, aggregate — plus Apache Arrow record conversion.
//   - io: readers and writers for CSV, XLSX, XPT (SAS), JSON, HTML,
//     Markdown, Parquet and Arrow IPC, plus a read-only SAS7BDAT reader.
//   - arrowutil: conversion helpers between enchanter null masks and Arrow
//     validity bitmaps.
//   - meta: the type system shared by the packages above.
package enchanter
