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

// NAs represents a series with no data.
type NAs struct {
	size      int
	partition *SeriesNAPartition
	ctx       *enchanter.Context
}

// Return the context of the series.
func (s NAs) Context() *enchanter.Context {
	return s.ctx
}

// Returns the length of the series.
func (s NAs) Len() int {
	return s.size
}

// Returns if the series is grouped.
func (s NAs) IsGrouped() bool {
	return s.partition != nil
}

// Returns if the series admits null values.
func (s NAs) IsNullable() bool {
	return true
}

func (s NAs) SortOrder() enchanter.SeriesSortOrder {
	return enchanter.SORTED_ASC
}

// Err returns the error carried by the series: always nil, an NAs
// series is never in an error state.
func (s NAs) Err() error {
	return nil
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
func (s NAs) NullMask() []bool {
	nullMask := make([]bool, s.size)
	for i := 0; i < s.size; i++ {
		nullMask[i] = true
	}
	return nullMask
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

// Slice returns the elements in the half-open interval [start, end).
func (s NAs) Slice(start, end int) Series {
	if start < 0 || end < start || end > s.size {
		return Errors{fmt.Sprintf("NAs.Slice: invalid interval [%d, %d) for a series of length %d", start, end, s.size)}
	}
	s.size = end - start
	return s
}

// TakeIndices returns the elements at the given indices: every element
// is null, so only the count matters.
func (s NAs) TakeIndices(indices []int) Series {
	for _, v := range indices {
		if v < 0 || v >= s.size {
			return Errors{fmt.Sprintf("NAs.TakeIndices: index %d is out of range", v)}
		}
	}
	s.size = len(indices)
	return s
}

// Append elements to the series.
func (s NAs) Append(v any) Series {
	var nullMask []byte
	switch v := v.(type) {
	case nil:
		s.size++
		return s

	case NAs:
		s.size += v.size
		return s

	case bool, enchanter.NullableBool, []bool, []enchanter.NullableBool, Bools:
		var data []bool
		switch v := v.(type) {
		case bool:
			data = make([]bool, s.size+1)
			data[s.size] = v
			nullMask = utils.BinVecInit(s.size+1, true)
			nullMask[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableBool:
			data = make([]bool, s.size+1)
			nullMask = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				data[s.size] = v.Value
				nullMask[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []bool:
			data = append(make([]bool, s.size), v...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableBool:
			data = make([]bool, s.size+len(v))
			nullMask = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					data[s.size+i] = v.Value
				} else {
					nullMask[i>>3] |= 1 << uint(i%8)
				}
			}

			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, nullMask)

		case Bools:
			data = append(make([]bool, s.size), v.data...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.nullMask)
		}

		return Bools{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case int, enchanter.NullableInt, []int, []enchanter.NullableInt, Ints:
		var data []int
		switch v := v.(type) {
		case int:
			data = make([]int, s.size+1)
			data[s.size] = v
			nullMask = utils.BinVecInit(s.size+1, true)
			nullMask[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableInt:
			data = make([]int, s.size+1)
			nullMask = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				data[s.size] = v.Value
				nullMask[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []int:
			data = append(make([]int, s.size), v...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableInt:
			data = make([]int, s.size+len(v))
			nullMask = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					data[s.size+i] = v.Value
				} else {
					nullMask[i>>3] |= 1 << uint(i%8)
				}
			}

			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, nullMask)

		case Ints:
			data = append(make([]int, s.size), v.data...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.nullMask)
		}

		return Ints{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case int64, enchanter.NullableInt64, []int64, []enchanter.NullableInt64, Int64s:
		var data []int64
		switch v := v.(type) {
		case int64:
			data = make([]int64, s.size+1)
			data[s.size] = v
			nullMask = utils.BinVecInit(s.size+1, true)
			nullMask[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableInt64:
			data = make([]int64, s.size+1)
			nullMask = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				data[s.size] = v.Value
				nullMask[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []int64:
			data = append(make([]int64, s.size), v...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableInt64:
			data = make([]int64, s.size+len(v))
			nullMask = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					data[s.size+i] = v.Value
				} else {
					nullMask[i>>3] |= 1 << uint(i%8)
				}
			}

			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, nullMask)

		case Int64s:
			data = append(make([]int64, s.size), v.data...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.nullMask)
		}

		return Int64s{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case float64, enchanter.NullableFloat64, []float64, []enchanter.NullableFloat64, Float64s:
		var data []float64
		switch v := v.(type) {
		case float64:
			data = make([]float64, s.size+1)
			data[s.size] = v
			nullMask = utils.BinVecInit(s.size+1, true)
			nullMask[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableFloat64:
			data = make([]float64, s.size+1)
			nullMask = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				data[s.size] = v.Value
				nullMask[s.size>>3] &= ^(1 << uint(s.size%8))
			}

		case []float64:
			data = append(make([]float64, s.size), v...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableFloat64:
			data = make([]float64, s.size+len(v))
			nullMask = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					data[s.size+i] = v.Value
				} else {
					nullMask[i>>3] |= 1 << uint(i%8)
				}
			}

			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, nullMask)

		case Float64s:
			data = append(make([]float64, s.size), v.data...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.nullMask)
		}

		return Float64s{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case string, enchanter.NullableString, []string, []enchanter.NullableString, Strings:
		data := make([]*string, s.size)
		for i := 0; i < s.size; i++ {
			data[i] = s.ctx.StringPool.Put(enchanter.NA_TEXT)
		}

		switch v := v.(type) {
		case string:
			data = append(data, s.ctx.StringPool.Put(v))
			nullMask = utils.BinVecInit(s.size+1, true)
			nullMask[s.size>>3] &= ^(1 << uint(s.size%8))

		case enchanter.NullableString:
			nullMask = utils.BinVecInit(s.size+1, true)
			if v.Valid {
				data = append(data, s.ctx.StringPool.Put(v.Value))
				nullMask[s.size>>3] &= ^(1 << uint(s.size%8))
			} else {
				data = append(data, s.ctx.StringPool.Put(enchanter.NA_TEXT))
			}

		case []string:
			data = append(data, make([]*string, len(v))...)
			for i, v := range v {
				data[s.size+i] = s.ctx.StringPool.Put(v)
			}
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), false, make([]uint8, 0))

		case []enchanter.NullableString:
			data = append(data, make([]*string, len(v))...)
			nullMask = utils.BinVecInit(len(v), false)
			for i, v := range v {
				if v.Valid {
					data[s.size+i] = s.ctx.StringPool.Put(v.Value)
				} else {
					nullMask[i>>3] |= 1 << uint(i%8)
					data[s.size+i] = s.ctx.StringPool.Put(enchanter.NA_TEXT)
				}
			}

			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), len(v), true, nullMask)

		case Strings:
			data = append(data, v.data...)
			_, nullMask = utils.MergeNullMasks(s.size, true, utils.BinVecInit(s.size, true), v.Len(), v.IsNullable(), v.nullMask)
		}

		return Strings{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	default:
		return Errors{fmt.Sprintf("NAs.Append: invalid type %T", v)}
	}
}

// All-data accessors.

// Returns the actual data of the series.
func (s NAs) Data() any {
	return make([]bool, s.size)
}

// Returns the nullable data of the series.
func (s NAs) DataAsNullable() any {
	return make([]enchanter.NullableBool, s.size)
}

// Returns the data of the series as a slice of strings.
func (s NAs) DataAsString() []string {
	data := make([]string, s.size)
	for i := 0; i < s.size; i++ {
		data[i] = enchanter.NA_TEXT
	}
	return data
}

// Casts the series to a given type.
func (s NAs) Cast(t meta.BaseType) Series {
	switch t {
	case meta.NullType:
		return s

	case meta.BoolType:
		return Bools{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       make([]bool, s.size),
			nullMask:   utils.BinVecInit(s.size, true),
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.IntType:
		return Ints{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       make([]int, s.size),
			nullMask:   utils.BinVecInit(s.size, true),
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Int64Type:
		return Int64s{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       make([]int64, s.size),
			nullMask:   utils.BinVecInit(s.size, true),
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Float64Type:
		return Float64s{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       make([]float64, s.size),
			nullMask:   utils.BinVecInit(s.size, true),
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.StringType:
		// Null string elements hold the interned NA text by convention;
		// nil pointers crash the generated operators.
		data := make([]*string, s.size)
		na := s.ctx.StringPool.Put(enchanter.NA_TEXT)
		for i := range data {
			data[i] = na
		}
		return Strings{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   utils.BinVecInit(s.size, true),
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.TimeType:
		return Times{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       make([]time.Time, s.size),
			nullMask:   utils.BinVecInit(s.size, true),
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.DurationType:
		return Durations{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       make([]time.Duration, s.size),
			nullMask:   utils.BinVecInit(s.size, true),
			partition:  nil,
			ctx:        s.ctx,
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
		return s.filterIntSlice(mask, true)
	default:
		return Errors{fmt.Sprintf("NAs.Filter: invalid type %T", mask)}
	}
}

func (s NAs) filterBool(mask Bools) Series {
	elementCount := 0
	for _, v := range mask.data {
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

func (s NAs) filterIntSlice(indexes []int, check bool) Series {
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
	partition map[int64][]int
}

func (gp *SeriesNAPartition) GetSize() int {
	return len(gp.partition)
}

func (gp *SeriesNAPartition) GetMap() map[int64][]int {
	return gp.partition
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

func (s NAs) Partition() SeriesPartition {
	return s.partition
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
		otherSeries = NewSeries(other, nil, false, false, s.ctx)
	}

	if e, ok := otherSeries.(Errors); ok {
		return e
	}

	if s.ctx != otherSeries.Context() {
		return Errors{fmt.Sprintf("Cannot operate on series with different contexts: %v and %v", s.ctx, otherSeries.Context())}
	}

	switch {
	case otherSeries.Len() == s.Len() || s.Len() == 1:
		return otherSeries
	case otherSeries.Len() == 1:
		// Broadcast the scalar to the receiver's length.
		indices := make([]int, s.Len())
		return otherSeries.TakeIndices(indices)
	}
	return Errors{fmt.Sprintf("Cannot coalesce %s and %s", s.Type().String(), otherSeries.Type().String())}
}
