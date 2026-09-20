package series

import (
	"fmt"
	"sort"
	"time"
	"unsafe"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

// Float64s represents a series of floats.
type Float64s struct {
	IsNullable_ bool
	Sorted_     enchanter.SeriesSortOrder
	Data_       []float64
	NullMask_   []uint8
	Partition_  *SeriesFloat64Partition
	Ctx_        *enchanter.Context
}

// ArrowArray builds and returns a fresh Arrow array from the series data.
// The caller owns the returned array; releasing it is optional under
// GC-backed allocators (see enchanter.Context.Allocator).
func (s Float64s) ArrowArray() arrow.Array {
	return buildArrowFloat64(s.Ctx_.Allocator, s.Data_, s.IsNullable_, s.NullMask_)
}

// Get the element at index i as a string.
func (s Float64s) GetAsString(i int) string {
	if s.IsNullable_ && s.IsNull(i) {
		return enchanter.NA_TEXT
	}
	return floatToString(s.Data_[i])
}

// Set the element at index i. The value v can be any belonging to types:
// int8, int16, int, int, int64, float32, float64 and their nullable versions.
func (s Float64s) Set(i int, v any) Series {
	if s.Partition_ != nil {
		return Errors{"Float64s.Set: cannot set values in a grouped series"}
	}

	switch val := v.(type) {
	case nil:
		s = s.MakeNullable().(Float64s)
		s.NullMask_[i>>3] |= 1 << uint(i%8)

	case int8:
		s.Data_[i] = float64(val)

	case int16:
		s.Data_[i] = float64(val)

	case int:
		s.Data_[i] = float64(val)

	case int32:
		s.Data_[i] = float64(val)

	case int64:
		s.Data_[i] = float64(val)

	case float32:
		s.Data_[i] = float64(val)

	case float64:
		s.Data_[i] = val

	case enchanter.NullableInt8:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt8).Valid {
			s.Data_[i] = float64(val.Value)
		} else {
			s.Data_[i] = 0
			s.NullMask_[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt16:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt16).Valid {
			s.Data_[i] = float64(val.Value)
		} else {
			s.Data_[i] = 0
			s.NullMask_[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt).Valid {
			s.Data_[i] = float64(val.Value)
		} else {
			s.Data_[i] = 0
			s.NullMask_[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt64:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt64).Valid {
			s.Data_[i] = float64(val.Value)
		} else {
			s.Data_[i] = 0
			s.NullMask_[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableFloat32:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableFloat32).Valid {
			s.Data_[i] = float64(val.Value)
		} else {
			s.Data_[i] = 0
			s.NullMask_[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableFloat64:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableFloat64).Valid {
			s.Data_[i] = val.Value
		} else {
			s.Data_[i] = 0
			s.NullMask_[i>>3] |= 1 << uint(i%8)
		}

	default:
		return Errors{fmt.Sprintf("Float64s.Set: invalid type %T", v)}
	}

	s.Sorted_ = enchanter.SORTED_NONE
	return s
}

////////////////////////			ALL DATA ACCESSORS

// Return the underlying Data_ as a slice of float64.
func (s Float64s) Float64s() []float64 {
	return s.Data_
}

// Return the underlying Data_ as a slice of NullableFloat64.
func (s Float64s) DataAsNullable() any {
	Data_ := make([]enchanter.NullableFloat64, len(s.Data_))
	for i, v := range s.Data_ {
		Data_[i] = enchanter.NullableFloat64{Valid: !s.IsNull(i), Value: v}
	}
	return Data_
}

// Return the underlying Data_ as a slice of strings.
func (s Float64s) DataAsString() []string {
	Data_ := make([]string, len(s.Data_))
	if s.IsNullable_ {
		for i, v := range s.Data_ {
			if s.IsNull(i) {
				Data_[i] = enchanter.NA_TEXT
			} else {
				Data_[i] = floatToString(v)
			}
		}
	} else {
		for i, v := range s.Data_ {
			Data_[i] = floatToString(v)
		}
	}
	return Data_
}

// Casts the series to a given type.
func (s Float64s) Cast(t meta.BaseType) Series {
	switch t {
	case meta.BoolType:
		Data_ := make([]bool, len(s.Data_))
		for i, v := range s.Data_ {
			Data_[i] = v != 0
		}

		return Bools{
			IsNullable_: s.IsNullable_,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   s.NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.IntType:
		Data_ := make([]int, len(s.Data_))
		for i, v := range s.Data_ {
			Data_[i] = int(v)
		}

		return Ints{
			IsNullable_: s.IsNullable_,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   s.NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.Int64Type:
		Data_ := make([]int64, len(s.Data_))
		for i, v := range s.Data_ {
			Data_[i] = int64(v)
		}

		return Int64s{
			IsNullable_: s.IsNullable_,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   s.NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.Float64Type:
		return s

	case meta.StringType:
		Data_ := make([]*string, len(s.Data_))
		if s.IsNullable_ {
			for i, v := range s.Data_ {
				if s.IsNull(i) {
					Data_[i] = s.Ctx_.StringPool.Put(enchanter.NA_TEXT)
				} else {
					Data_[i] = s.Ctx_.StringPool.Put(floatToString(v))
				}
			}
		} else {
			for i, v := range s.Data_ {
				Data_[i] = s.Ctx_.StringPool.Put(floatToString(v))
			}
		}

		return Strings{
			IsNullable_: s.IsNullable_,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   s.NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.TimeType:
		Data_ := make([]time.Time, len(s.Data_))
		for i, v := range s.Data_ {
			Data_[i] = time.Unix(0, int64(v))
		}

		return Times{
			IsNullable_: s.IsNullable_,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   s.NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	case meta.DurationType:
		Data_ := make([]time.Duration, len(s.Data_))
		for i, v := range s.Data_ {
			Data_[i] = time.Duration(v)
		}

		return Durations{
			IsNullable_: s.IsNullable_,
			Sorted_:     enchanter.SORTED_NONE,
			Data_:       Data_,
			NullMask_:   s.NullMask_,
			Partition_:  nil,
			Ctx_:        s.Ctx_,
		}

	default:
		return Errors{fmt.Sprintf("Float64s.Cast: invalid type %s", t.String())}
	}
}

////////////////////////			GROUPING OPERATIONS

// A SeriesFloat64Partition is a Partition_ of a Float64s.
// Each key is a hash of a bool value, and each value is a slice of indices
// of the original series that are set to that value.
type SeriesFloat64Partition struct {
	Partition_   map[int64][]int
	indexToGroup []int
}

func (gp *SeriesFloat64Partition) GetSize() int {
	return len(gp.Partition_)
}

func (gp *SeriesFloat64Partition) GetMap() map[int64][]int {
	return gp.Partition_
}

func (s Float64s) Group() Series {

	// Define the worker callback
	worker := func(threadNum, start, end int, map_ map[int64][]int) {
		for i := start; i < end; i++ {
			map_[*(*int64)(unsafe.Pointer((&s.Data_[i])))] = append(map_[*(*int64)(unsafe.Pointer((&s.Data_[i])))], i)
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		for i := start; i < end; i++ {
			if s.IsNull(i) {
				(*nulls) = append((*nulls), i)
			} else {
				map_[*(*int64)(unsafe.Pointer((&s.Data_[i])))] = append(map_[*(*int64)(unsafe.Pointer((&s.Data_[i])))], i)
			}
		}
	}

	Partition_ := SeriesFloat64Partition{
		Partition_: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_2, len(s.Data_), s.HasNull(),
			worker, workerNulls),
	}

	s.Partition_ = &Partition_

	return s
}

func (s Float64s) GroupBy(Partition_ SeriesPartition) Series {
	// collect all keys
	otherIndeces := Partition_.GetMap()
	keys := make([]int64, len(otherIndeces))
	i := 0
	for k := range otherIndeces {
		keys[i] = k
		i++
	}

	// Define the worker callback
	worker := func(threadNum, start, end int, map_ map[int64][]int) {
		var newHash int64
		for _, h := range keys[start:end] { // keys is defined outside the function
			for _, index := range otherIndeces[h] { // otherIndeces is defined outside the function
				newHash = *(*int64)(unsafe.Pointer((&(s.Data_)[index]))) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		var newHash int64
		for _, h := range keys[start:end] { // keys is defined outside the function
			for _, index := range otherIndeces[h] { // otherIndeces is defined outside the function
				if s.IsNull(index) {
					newHash = enchanter.HASH_MAGIC_NUMBER_NULL + (h << 13) + (h >> 4)
				} else {
					newHash = *(*int64)(unsafe.Pointer((&(s.Data_)[index]))) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	newPartition := SeriesFloat64Partition{
		Partition_: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, len(keys), s.HasNull(),
			worker, workerNulls),
	}

	s.Partition_ = &newPartition

	return s
}

////////////////////////			SORTING OPERATIONS

func (s Float64s) Less(i, j int) bool {
	if s.IsNullable_ {
		if s.NullMask_[i>>3]&(1<<uint(i%8)) > 0 {
			return false
		}
		if s.NullMask_[j>>3]&(1<<uint(j%8)) > 0 {
			return true
		}
	}

	return s.Data_[i] < s.Data_[j]
}

func (s Float64s) Equal(i, j int) bool {
	if s.IsNullable_ {
		if (s.NullMask_[i>>3] & (1 << uint(i%8))) > 0 {
			return (s.NullMask_[j>>3] & (1 << uint(j%8))) > 0
		}
		if (s.NullMask_[j>>3] & (1 << uint(j%8))) > 0 {
			return false
		}
	}

	return s.Data_[i] == s.Data_[j]
}

func (s Float64s) Swap(i, j int) {
	if s.IsNullable_ {
		// i is null, j is not null
		if s.NullMask_[i>>3]&(1<<uint(i%8)) > 0 && s.NullMask_[j>>3]&(1<<uint(j%8)) == 0 {
			s.NullMask_[i>>3] &= ^(1 << uint(i%8))
			s.NullMask_[j>>3] |= 1 << uint(j%8)
		} else

		// i is not null, j is null
		if s.NullMask_[i>>3]&(1<<uint(i%8)) == 0 && s.NullMask_[j>>3]&(1<<uint(j%8)) > 0 {
			s.NullMask_[i>>3] |= 1 << uint(i%8)
			s.NullMask_[j>>3] &= ^(1 << uint(j%8))
		}
	}

	s.Data_[i], s.Data_[j] = s.Data_[j], s.Data_[i]
}

func (s Float64s) Sort() Series {
	if s.Sorted_ != enchanter.SORTED_ASC {
		sort.Sort(s)
		s.Sorted_ = enchanter.SORTED_ASC
	}
	return s
}

func (s Float64s) SortRev() Series {
	if s.Sorted_ != enchanter.SORTED_DESC {
		sort.Sort(sort.Reverse(s))
		s.Sorted_ = enchanter.SORTED_DESC
	}
	return s
}
