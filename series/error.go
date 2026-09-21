package series

import (
	"errors"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

// Dummy series for error handling.
type Errors struct {
	msg string
}

// Return the context of the series.
func (s Errors) Context() *enchanter.Context {
	return nil
}

// Returns the length of the series.
func (s Errors) Len() int {
	return 0
}

// Returns if the series is grouped.
func (s Errors) IsGrouped() bool {
	return false
}

// Returns if the series admits null values.
func (s Errors) IsNullable() bool {
	return false
}

func (s Errors) SortOrder() enchanter.SeriesSortOrder {
	return enchanter.SORTED_NONE
}

// Err returns the error carried by the series, built from the stored
// message on each call.
func (s Errors) Err() error {
	return errors.New(s.msg)
}

// Makes the series nullable.
func (s Errors) MakeNullable() Series {
	return s
}

// Make the series non-nullable.
func (s Errors) MakeNonNullable() Series {
	return s
}

// Returns the type of the series.
func (s Errors) Type() meta.BaseType {
	return meta.ErrorType
}

// Returns the type and cardinality of the series.
func (s Errors) TypeCard() meta.BaseTypeCard {
	return meta.BaseTypeCard{Base: meta.ErrorType, Card: s.Len()}
}

// Returns if the series has null values.
func (s Errors) HasNull() bool {
	return false
}

// Returns the number of null values in the series.
func (s Errors) NullCount() int {
	return 0
}

// Returns if the element at index i is null.
func (s Errors) IsNull(i int) bool {
	return false
}

// Returns the null mask of the series.
func (s Errors) NullMask() []bool {
	return []bool{}
}

// Sets the null mask of the series.
func (s Errors) SetNullMask(mask []bool) Series {
	return s
}

// Get the element at index i.
func (s Errors) Get(i int) any {
	return nil
}

func (s Errors) GetAsString(i int) string {
	return ""
}

// Set the element at index i.
func (s Errors) Set(i int, v any) Series {
	return s
}

// Slice on an errored series keeps the error.
func (s Errors) Slice(start, end int) Series {
	return s
}

// TakeIndices on an errored series keeps the error.
func (s Errors) TakeIndices(indices []int) Series {
	return s
}

// Append elements to the series.
func (s Errors) Append(v any) Series {
	return s
}

// All-data accessors.

// Returns the actual data of the series.
func (s Errors) Data() any {
	return s
}

// Returns the nullable data of the series.
func (s Errors) DataAsNullable() any {
	return s
}

// Returns the data of the series as a slice of strings.
func (s Errors) DataAsString() []string {
	return []string{s.msg}
}

// Casts the series to a given type.
func (s Errors) Cast(t meta.BaseType) Series {
	return s
}

// Copies the series.
func (s Errors) Copy() Series {
	return s
}

// Series operations.

// Filters out the elements by the given mask.
// Mask can be a bool series, a slice of bools or a slice of ints.
func (s Errors) Filter(mask any) Series {
	return s
}

func (s Errors) filterIntSlice(mask []int, check bool) Series {
	return s
}

func (s Errors) Map(f enchanter.MapFunc) Series {
	return s
}

func (s Errors) MapNull(f enchanter.MapFuncNull) Series {
	return s
}

// Group the elements in the series.
func (s Errors) Group() Series {
	return s
}

func (s Errors) GroupBy(gp SeriesPartition) Series {
	return s
}

func (s Errors) UnGroup() Series {
	return s
}

func (s Errors) Partition() SeriesPartition {
	return nil
}

// Sort interface.
func (s Errors) Less(i, j int) bool {
	return false
}

func (s Errors) Equal(i, j int) bool {
	return false
}

func (s Errors) Swap(i, j int) {}

func (s Errors) Sort() Series {
	return s
}

func (s Errors) SortRev() Series {
	return s
}

////////////////////////			ARITHMETIC OPERATIONS

func (s Errors) And(other any) Series {
	return s
}

func (s Errors) Or(other any) Series {
	return s
}

// Coalesce on an errored series keeps the error.
func (s Errors) Coalesce(other any) Series {
	return s
}

// Not on an errored series keeps the error.
func (s Errors) Not() Series {
	return s
}

func (s Errors) Mul(other any) Series {
	return s
}

func (s Errors) Div(other any) Series {
	return s
}

func (s Errors) Mod(other any) Series {
	return s
}

func (s Errors) Exp(other any) Series {
	return s
}

func (s Errors) Add(other any) Series {
	return s
}

func (s Errors) Sub(other any) Series {
	return s
}

////////////////////////			LOGICAL OPERATIONS

func (s Errors) Eq(other any) Series {
	return s
}

func (s Errors) Ne(other any) Series {
	return s
}

func (s Errors) Gt(other any) Series {
	return s
}

func (s Errors) Ge(other any) Series {
	return s
}

func (s Errors) Lt(other any) Series {
	return s
}

func (s Errors) Le(other any) Series {
	return s
}

func (s Errors) ArrowArray() arrow.Array {
	return nil
}
