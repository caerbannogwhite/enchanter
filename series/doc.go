// Package series implements typed, nullable columns — the building block of
// dataframes.
//
// A [Series] holds a Go slice of one element type plus a bit-packed null
// mask. The concrete types are [Bools], [Ints], [Int64s], [Float64s],
// [Strings], [Times] and [Durations], built with the NewSeries* constructors:
//
//	ctx := enchanter.NewContext()
//	s := series.NewSeriesFloat64([]float64{1, 2, 3}, []bool{false, true, false}, false, ctx)
//
// Series support filtering, sorting, grouping, casting and element-wise
// arithmetic, comparison and boolean operations. Nulls propagate: an element
// is null when either operand is null there, and a whole NA operand yields
// an all-NA result. Coalesce is the deliberate exception, it exists to fill
// nulls. Failed operations return an [Errors] series that propagates through
// subsequent calls — check with Err.
//
// Every series converts to and from Apache Arrow: ArrowArray builds a fresh
// Arrow array from the data, and [ArrowArrayToSeries] materializes any
// supported Arrow array (including types enchanter does not have natively,
// such as narrower integers or non-nanosecond timestamps) into a series.
package series
