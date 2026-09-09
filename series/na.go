package series

import (
	"fmt"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/utils"
)

// NAs represents a series with no Data_.
type NAs struct {
	size       int
	Partition_ *SeriesNAPartition
	Ctx_       *enchanter.Context
}

// Return the context of the series.
func (s NAs) GetContext() *enchanter.Context {
	return s.Ctx_
}

// Returns the length of the series.
func (s NAs) Len() int {
	return s.size
}

// Returns if the series is grouped.
func (s NAs) IsGrouped() bool {
	return s.Partition_ != nil
}

// Returns if the series admits null values.
func (s NAs) IsNullable() bool {
	return true
}

func (s NAs) IsSorted() enchanter.SeriesSortOrder {
	return enchanter.SORTED_ASC
}

// Returns if the series is error.
func (s NAs) IsError() bool {
	return false
}

// Returns the error message of the series.
func (s NAs) GetError() string {
	return ""
}

// Makes the series nullable.
func (s NAs) MakeNullable() Series {
	return s
}

// Make the series non-nullable.
func (s NAs) MakeNonNullable() Series {
	return s
}

// Returns the type of the series.
func (s NAs) Type() meta.BaseType {
	return meta.NullType
}

// Returns the type and cardinality of the series.
func (s NAs) TypeCard() meta.BaseTypeCard {
	return meta.BaseTypeCard{Base: meta.NullType, Card: s.Len()}
}

// Returns if the series has null values.
func (s NAs) HasNull() bool {
	return true
}

// Returns the number of null values in the series.
func (s NAs) NullCount() int {
	return s.size
}

// Returns if the element at index i is null.
func (s NAs) IsNull(i int) bool {
	return true
}

// Returns the null mask of the series.
func (s NAs) GetNullMask() []bool {
	NullMask_ := make([]bool, s.size)
	for i := 0; i < s.size; i++ {
		NullMask_[i] = true
	}
	return NullMask_
}

// Sets the null mask of the series.
func (s NAs) SetNullMask(mask []bool) Series {
	return s
}

// Get the element at index i.
func (s NAs) Get(i int) any {
	return nil
}

func (s NAs) GetAsString(i int) string {
	return enchanter.NA_TEXT
}

// Set the element at index i.
func (s NAs) Set(i int, v any) Series {
	return s
}

// Take the elements according to the given interval.
func (s NAs) Take(params ...int) Series {
	return s
}

// Append elements to the series.
func (s NAs) Append(v any) Series {
	var NullMask_ []byte
	switch v := v.(type) {
	case nil:
		s.size++
		return s

	case NAs:
		s.size += v.size
		return s

	case bool, enchanter.NullableBool, []bool, []enchanter.NullableBool, Bools:
		var Data_ []bool
		switch v := v.(type) {
		case bool:
			Data_ = make([]bool, s.size+1)
			Data_[s.size] = v
			NullMask_ = utils.BinVecInit(s.size+1, true)
			NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableBool:
			Data_ = make([]bool, s.size+1)
			NullMask_ = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				Data_[s.size] = v.Value
				NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []bool:
			Data_ = append(make([]bool, s.size), v...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableBool:
			Data_ = make([]bool, s.size+len(v))
			NullMask_ = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					Data_[s.size+i] = v.Value
				} else {
					NullMask_[i>>3] |= 1 << uint(i%8)
				}
			}

			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, NullMask_)

		case Bools:
			Data_ = append(make([]bool, s.size), v.Data_...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.NullMask_)
		}

		return Bools{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case int, enchanter.NullableInt, []int, []enchanter.NullableInt, Ints:
		var Data_ []int
		switch v := v.(type) {
		case int:
			Data_ = make([]int, s.size+1)
			Data_[s.size] = v
			NullMask_ = utils.BinVecInit(s.size+1, true)
			NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableInt:
			Data_ = make([]int, s.size+1)
			NullMask_ = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				Data_[s.size] = v.Value
				NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []int:
			Data_ = append(make([]int, s.size), v...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableInt:
			Data_ = make([]int, s.size+len(v))
			NullMask_ = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					Data_[s.size+i] = v.Value
				} else {
					NullMask_[i>>3] |= 1 << uint(i%8)
				}
			}

			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, NullMask_)

		case Ints:
			Data_ = append(make([]int, s.size), v.Data_...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.NullMask_)
		}

		return Ints{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case int64, enchanter.NullableInt64, []int64, []enchanter.NullableInt64, Int64s:
		var Data_ []int64
		switch v := v.(type) {
		case int64:
			Data_ = make([]int64, s.size+1)
			Data_[s.size] = v
			NullMask_ = utils.BinVecInit(s.size+1, true)
			NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableInt64:
			Data_ = make([]int64, s.size+1)
			NullMask_ = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				Data_[s.size] = v.Value
				NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []int64:
			Data_ = append(make([]int64, s.size), v...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableInt64:
			Data_ = make([]int64, s.size+len(v))
			NullMask_ = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					Data_[s.size+i] = v.Value
				} else {
					NullMask_[i>>3] |= 1 << uint(i%8)
				}
			}

			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, NullMask_)

		case Int64s:
			Data_ = append(make([]int64, s.size), v.Data_...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.NullMask_)
		}

		return Int64s{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case float64, enchanter.NullableFloat64, []float64, []enchanter.NullableFloat64, Float64s:
		var Data_ []float64
		switch v := v.(type) {
		case float64:
			Data_ = make([]float64, s.size+1)
			Data_[s.size] = v
			NullMask_ = utils.BinVecInit(s.size+1, true)
			NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableFloat64:
			Data_ = make([]float64, s.size+1)
			NullMask_ = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				Data_[s.size] = v.Value
				NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []float64:
			Data_ = append(make([]float64, s.size), v...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableFloat64:
			Data_ = make([]float64, s.size+len(v))
			NullMask_ = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					Data_[s.size+i] = v.Value
				} else {
					NullMask_[i>>3] |= 1 << uint(i%8)
				}
			}

			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, NullMask_)

		case Float64s:
			Data_ = append(make([]float64, s.size), v.Data_...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.NullMask_)
		}

		return Float64s{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case string, enchanter.NullableString, []string, []enchanter.NullableString, Strings:
		Data_ := make([]*string, s.size)
		for i := 0; i < s.size; i++ {
			Data_[i] = s.Ctx_.StringPool.Put(enchanter.NA_TEXT)
		}

		switch v := v.(type) {
		case string:
			Data_ = append(Data_, s.Ctx_.StringPool.Put(v))
			NullMask_ = utils.BinVecInit(s.size+1, true)
			NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableString:
			NullMask_ = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				Data_ = append(Data_, s.Ctx_.StringPool.Put(v.Value))
				NullMask_[s.size>>3] &= ^(1 << uint(s.size%8))
			} else {
				Data_ = append(Data_, s.Ctx_.StringPool.Put(enchanter.NA_TEXT))
			}

		case []string:
			Data_ = append(Data_, make([]*string, len(v))...)
			for i, v := range v {
				Data_[s.size+i] = s.Ctx_.StringPool.Put(v)
			}
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableString:
			Data_ = append(Data_, make([]*string, len(v))...)
			NullMask_ = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					Data_[s.size+i] = s.Ctx_.StringPool.Put(v.Value)
				} else {
					NullMask_[i>>3] |= 1 << uint(i%8)
					Data_[s.size+i] = s.Ctx_.StringPool.Put(enchanter.NA_TEXT)
				}
			}

			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, NullMask_)

		case Strings:
			Data_ = append(Data_, v.Data_...)
			_, NullMask_ = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.NullMask_)
		}

		return Strings{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	default:
		return Errors{fmt.Sprintf("NAs.Append: invalid type %T", v)}
	}
}

// All-Data_ accessors.

// Returns the actual Data_ of the series.
func (s NAs) Data() any {
	return make([]bool, s.size)
}

// Returns the nullable Data_ of the series.
func (s NAs) DataAsNullable() any {
	return make([]enchanter.NullableBool, s.size)
}

// Returns the Data_ of the series as a slice of strings.
func (s NAs) DataAsString() []string {
	Data_ := make([]string, s.size)
	for i := 0; i < s.size; i++ {
		Data_[i] = enchanter.NA_TEXT
	}
	return Data_
}

// Casts the series to a given type.
func (s NAs) Cast(t meta.BaseType) Series {
	switch t {
	case meta.NullType:
		return s

	case meta.BoolType:
		return Bools{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       make([]bool, s.size),
			NullMask_:   utils.BinVecInit(s.size, true),
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.IntType:
		return Ints{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       make([]int, s.size),
			NullMask_:   utils.BinVecInit(s.size, true),
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.Int64Type:
		return Int64s{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       make([]int64, s.size),
			NullMask_:   utils.BinVecInit(s.size, true),
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.Float64Type:
		return Float64s{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       make([]float64, s.size),
			NullMask_:   utils.BinVecInit(s.size, true),
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.StringType:
		return Strings{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       make([]*string, s.size),
			NullMask_:   utils.BinVecInit(s.size, true),
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.TimeType:
		return Times{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       make([]time.Time, s.size),
			NullMask_:   utils.BinVecInit(s.size, true),
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.DurationType:
		return Durations{
			IsNullable_: true,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       make([]time.Duration, s.size),
			NullMask_:   utils.BinVecInit(s.size, true),
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	default:
		return Errors{fmt.Sprintf("NAs.Cast: invalid type %s", t.String())}
	}
}

// Copies the series.
func (s NAs) Copy() Series {
	return s
}

// Series operations.

// Filters out the elements by the given mask.
// Mask can be a bool series, a slice of bools or a slice of ints.
func (s NAs) Filter(mask any) Series {
	switch mask := mask.(type) {
	case Bools:
		return s.filterBool(mask)
	case []bool:
		return s.filterBoolSlice(mask)
	case []int:
		return s.FilterIntSlice(mask, true)
	default:
		return Errors{fmt.Sprintf("NAs.Filter: invalid type %T", mask)}
	}
}

func (s NAs) filterBool(mask Bools) Series {
	elementCount := 0
	for _, v := range mask.Data_ {
		if v {
			elementCount++
		}
	}

	s.size = elementCount
	return s
}

func (s NAs) filterBoolSlice(mask []bool) Series {
	elementCount := 0
	for _, v := range mask {
		if v {
			elementCount++
		}
	}

	s.size = elementCount
	return s
}

func (s NAs) FilterIntSlice(indexes []int, check bool) Series {
	// check if indexes are in range
	if check {
		for _, v := range indexes {
			if v < 0 || v >= s.size {
				return Errors{fmt.Sprintf("NAs.Filter: index %d is out of range", v)}
			}
		}
	}

	s.size = len(indexes)
	return s
}

func (s NAs) Map(f enchanter.MapFunc) Series {
	return s
}

func (s NAs) MapNull(f enchanter.MapFuncNull) Series {
	return s
}

type SeriesNAPartition struct {
	Partition_ map[int64][]int
}

func (gp *SeriesNAPartition) GetSize() int {
	return len(gp.Partition_)
}

func (gp *SeriesNAPartition) GetMap() map[int64][]int {
	return gp.Partition_
}

// Group the elements in the series.
func (s NAs) Group() Series {
	return s
}

func (s NAs) GroupBy(gp SeriesPartition) Series {
	return s
}

func (s NAs) UnGroup() Series {
	return s
}

func (s NAs) GetPartition() SeriesPartition {
	return s.Partition_
}

// Sort interface.
func (s NAs) Less(i, j int) bool {
	return false
}

func (s NAs) Equal(i, j int) bool {
	return false
}

func (s NAs) Swap(i, j int) {}

func (s NAs) Sort() Series {
	return s
}

func (s NAs) SortRev() Series {
	return s
}

// ArrowArray returns an all-null Arrow array of the NAs size.
func (s NAs) ArrowArray() arrow.Array {
	builder := array.NewNullBuilder(memory.DefaultAllocator)
	defer builder.Release()
	builder.AppendNulls(s.size)
	return builder.NewNullArray()
}

// Coalesce fills the null elements of the series with the corresponding
// elements of other. Every element of NAs is null, so the result is the
// other operand: broadcast when it is a scalar, unchanged when the lengths
// match. The returned series may share the operand's storage.
func (s NAs) Coalesce(other any) Series {
	var otherSeries Series
	if o, ok := other.(Series); ok {
		otherSeries = o
	} else {
		otherSeries = NewSeries(other, nil, false, false, s.Ctx_)
	}

	if e, ok := otherSeries.(Errors); ok {
		return e
	}

	if s.Ctx_ != otherSeries.GetContext() {
		return Errors{fmt.Sprintf("Cannot operate on series with different contexts: %v and %v", s.Ctx_, otherSeries.GetContext())}
	}

	switch {
	case otherSeries.Len() == s.Len() || s.Len() == 1:
		return otherSeries
	case otherSeries.Len() == 1:
		// Broadcast the scalar to the receiver's length.
		indices := make([]int, s.Len())
		return otherSeries.FilterIntSlice(indices, false)
	}
	return Errors{fmt.Sprintf("Cannot coalesce %s and %s", s.Type().String(), otherSeries.Type().String())}
}
