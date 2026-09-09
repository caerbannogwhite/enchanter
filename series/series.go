package series

// The *_base.go and *_ops.go files in this package are produced by the code
// generator in ../generators (its own Go module). Regenerate them with:
//go:generate go -C ../generators run .

import (
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

type Series interface {
	// Basic accessors.

	// Return the context of the series.
	GetContext() *enchanter.Context
	// Return the number of elements in the series.
	Len() int
	// Return the type of the series.
	Type() meta.BaseType
	// Return the type and cardinality of the series.
	TypeCard() meta.BaseTypeCard
	// Return if the series is grouped.
	IsGrouped() bool
	// Return if the series admits null values.
	IsNullable() bool
	// Return if the series is sorted.
	IsSorted() enchanter.SeriesSortOrder
	// Return if the series is error.
	IsError() bool
	// Return the error message of the series.
	GetError() string

	// Nullability operations.

	// Return if the series has null values.
	HasNull() bool
	// Return the number of null values in the series.
	NullCount() int
	// Return if the element at index i is null.
	IsNull(i int) bool
	// Return the null mask of the series.
	GetNullMask() []bool
	// Set the null mask of the series.
	SetNullMask(mask []bool) Series
	// Make the series nullable.
	MakeNullable() Series
	// Make the series non-nullable.
	MakeNonNullable() Series

	// Get the element at index i.
	Get(i int) any
	// Get the element at index i as a string.
	GetAsString(i int) string
	// Set the element at index i.
	Set(i int, v any) Series
	// Take the elements according to the given interval.
	Take(params ...int) Series

	// Append elements to the series.
	// Value can be a single value, slice of values,
	// a nullable value, a slice of nullable values or a series.
	Append(v any) Series

	// All-data accessors.

	// Return the actual data of the series.
	Data() any
	// Return the nullable data of the series.
	DataAsNullable() any
	// Return the data of the series as a slice of strings.
	DataAsString() []string

	// Cast the series to a given type.
	Cast(t meta.BaseType) Series
	// Copy the series.
	Copy() Series

	// Series operations.

	// Filter out the elements by the given mask.
	// Mask can be a bool series, a slice of bools or a slice of ints.
	Filter(mask any) Series
	FilterIntSlice(mask []int, check bool) Series

	// Apply the given function to each element of the series.
	Map(f enchanter.MapFunc) Series
	MapNull(f enchanter.MapFuncNull) Series

	// Group the elements in the series.
	Group() Series
	GroupBy(gp SeriesPartition) Series
	UnGroup() Series

	// Get the partition of the series.
	GetPartition() SeriesPartition

	// Sort Interface.
	Less(i, j int) bool
	Equal(i, j int) bool
	Swap(i, j int)

	// Sort the elements of the series.
	Sort() Series
	SortRev() Series

	// Boolean operations.
	And(other any) Series
	Or(other any) Series

	// Coalesce fills the null elements with the corresponding elements of
	// other. A result element is null only when both operands are null there.
	Coalesce(other any) Series

	// Arithmetic operations.
	Mul(other any) Series
	Div(other any) Series
	Mod(other any) Series
	Exp(other any) Series
	Add(other any) Series
	Sub(other any) Series

	// Logical operations.
	Eq(other any) Series
	Ne(other any) Series
	Gt(other any) Series
	Ge(other any) Series
	Lt(other any) Series
	Le(other any) Series

	// Arrow interop.
	// Return the underlying Arrow array. May build it lazily from Go slices.
	ArrowArray() arrow.Array
}

type SeriesNumeric interface {
	Series

	// Return the minimum value of the series.
	Min() any
	// Return the maximum value of the series.
	Max() any
	// Return the sum of the values of the series.
	Sum() any
	// Return the mean of the values of the series.
	Mean() any
	// Return the median of the values of the series.
	Median() any
	// Return the variance of the values of the series.
	Variance() any
	// Return the standard deviation of the values of the series.
	StdDev() any
	// Return the quantile of the values of the series.
	Quantile(q any) any
}

type SeriesPartition interface {
	// Return the number partitions.
	GetSize() int

	// Return the indices of the groups.
	GetMap() map[int64][]int
}
