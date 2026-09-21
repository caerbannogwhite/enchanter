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
	isNullable bool
	sorted     enchanter.SeriesSortOrder
	data       []float64
	nullMask   []uint8
	partition  *SeriesFloat64Partition
	ctx        *enchanter.Context
}

// ArrowArray builds and returns a fresh Arrow array from the series data.
// The caller owns the returned array; releasing it is optional under
// GC-backed allocators (see enchanter.Context.Allocator).
func (s Float64s) ArrowArray() arrow.Array {
	return buildArrowFloat64(s.ctx.Allocator, s.data, s.isNullable, s.nullMask)
}

// Get the element at index i as a string.
func (s Float64s) GetAsString(i int) string {
	if s.isNullable && s.IsNull(i) {
		return enchanter.NA_TEXT
	}
	return floatToString(s.data[i])
}

// Set the element at index i. The value v can be any belonging to types:
// int8, int16, int, int, int64, float32, float64 and their nullable versions.
func (s Float64s) Set(i int, v any) Series {
	if s.partition != nil {
		return Errors{"Float64s.Set: cannot set values in a grouped series"}
	}

	switch val := v.(type) {
	case nil:
		s = s.MakeNullable().(Float64s)
		s.nullMask[i>>3] |= 1 << uint(i%8)

	case int8:
		s.data[i] = float64(val)

	case int16:
		s.data[i] = float64(val)

	case int:
		s.data[i] = float64(val)

	case int32:
		s.data[i] = float64(val)

	case int64:
		s.data[i] = float64(val)

	case float32:
		s.data[i] = float64(val)

	case float64:
		s.data[i] = val

	case enchanter.NullableInt8:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt8).Valid {
			s.data[i] = float64(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt16:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt16).Valid {
			s.data[i] = float64(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt).Valid {
			s.data[i] = float64(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt64:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableInt64).Valid {
			s.data[i] = float64(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableFloat32:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableFloat32).Valid {
			s.data[i] = float64(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableFloat64:
		s = s.MakeNullable().(Float64s)
		if v.(enchanter.NullableFloat64).Valid {
			s.data[i] = val.Value
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	default:
		return Errors{fmt.Sprintf("Float64s.Set: invalid type %T", v)}
	}

	s.sorted = enchanter.SORTED_NONE
	return s
}

////////////////////////			ALL DATA ACCESSORS

// Return the underlying data as a slice of float64.
func (s Float64s) Float64s() []float64 {
	return s.data
}

// Return the underlying data as a slice of NullableFloat64.
func (s Float64s) DataAsNullable() any {
	data := make([]enchanter.NullableFloat64, len(s.data))
	for i, v := range s.data {
		data[i] = enchanter.NullableFloat64{Valid: !s.IsNull(i), Value: v}
	}
	return data
}

// Return the underlying data as a slice of strings.
func (s Float64s) DataAsString() []string {
	data := make([]string, len(s.data))
	if s.isNullable {
		for i, v := range s.data {
			if s.IsNull(i) {
				data[i] = enchanter.NA_TEXT
			} else {
				data[i] = floatToString(v)
			}
		}
	} else {
		for i, v := range s.data {
			data[i] = floatToString(v)
		}
	}
	return data
}

// Casts the series to a given type.
func (s Float64s) Cast(t meta.BaseType) Series {
	switch t {
	case meta.BoolType:
		data := make([]bool, len(s.data))
		for i, v := range s.data {
			data[i] = v != 0
		}

		return Bools{
			isNullable: s.isNullable,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.IntType:
		data := make([]int, len(s.data))
		for i, v := range s.data {
			data[i] = int(v)
		}

		return Ints{
			isNullable: s.isNullable,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Int64Type:
		data := make([]int64, len(s.data))
		for i, v := range s.data {
			data[i] = int64(v)
		}

		return Int64s{
			isNullable: s.isNullable,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Float64Type:
		return s

	case meta.StringType:
		data := make([]*string, len(s.data))
		if s.isNullable {
			for i, v := range s.data {
				if s.IsNull(i) {
					data[i] = s.ctx.StringPool.Put(enchanter.NA_TEXT)
				} else {
					data[i] = s.ctx.StringPool.Put(floatToString(v))
				}
			}
		} else {
			for i, v := range s.data {
				data[i] = s.ctx.StringPool.Put(floatToString(v))
			}
		}

		return Strings{
			isNullable: s.isNullable,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.TimeType:
		data := make([]time.Time, len(s.data))
		for i, v := range s.data {
			data[i] = time.Unix(0, int64(v))
		}

		return Times{
			isNullable: s.isNullable,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.DurationType:
		data := make([]time.Duration, len(s.data))
		for i, v := range s.data {
			data[i] = time.Duration(v)
		}

		return Durations{
			isNullable: s.isNullable,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	default:
		return Errors{fmt.Sprintf("Float64s.Cast: invalid type %s", t.String())}
	}
}

////////////////////////			GROUPING OPERATIONS

// A SeriesFloat64Partition is a partition of a Float64s.
// Each key is a hash of a bool value, and each value is a slice of indices
// of the original series that are set to that value.
type SeriesFloat64Partition struct {
	partition    map[int64][]int
	indexToGroup []int
}

func (gp *SeriesFloat64Partition) GetSize() int {
	return len(gp.partition)
}

func (gp *SeriesFloat64Partition) GetMap() map[int64][]int {
	return gp.partition
}

func (s Float64s) Group() Series {

	// Define the worker callback
	worker := func(threadNum, start, end int, map_ map[int64][]int) {
		for i := start; i < end; i++ {
			map_[*(*int64)(unsafe.Pointer((&s.data[i])))] = append(map_[*(*int64)(unsafe.Pointer((&s.data[i])))], i)
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		for i := start; i < end; i++ {
			if s.IsNull(i) {
				(*nulls) = append((*nulls), i)
			} else {
				map_[*(*int64)(unsafe.Pointer((&s.data[i])))] = append(map_[*(*int64)(unsafe.Pointer((&s.data[i])))], i)
			}
		}
	}

	partition := SeriesFloat64Partition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_2, len(s.data), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &partition

	return s
}

func (s Float64s) GroupBy(partition SeriesPartition) Series {
	// collect all keys
	otherIndeces := partition.GetMap()
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
				newHash = *(*int64)(unsafe.Pointer((&(s.data)[index]))) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
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
					newHash = *(*int64)(unsafe.Pointer((&(s.data)[index]))) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	newPartition := SeriesFloat64Partition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, len(keys), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &newPartition

	return s
}

////////////////////////			SORTING OPERATIONS

func (s Float64s) Less(i, j int) bool {
	if s.isNullable {
		if s.nullMask[i>>3]&(1<<uint(i%8)) > 0 {
			return false
		}
		if s.nullMask[j>>3]&(1<<uint(j%8)) > 0 {
			return true
		}
	}

	return s.data[i] < s.data[j]
}

func (s Float64s) Equal(i, j int) bool {
	if s.isNullable {
		if (s.nullMask[i>>3] & (1 << uint(i%8))) > 0 {
			return (s.nullMask[j>>3] & (1 << uint(j%8))) > 0
		}
		if (s.nullMask[j>>3] & (1 << uint(j%8))) > 0 {
			return false
		}
	}

	return s.data[i] == s.data[j]
}

func (s Float64s) Swap(i, j int) {
	if s.isNullable {
		// i is null, j is not null
		if s.nullMask[i>>3]&(1<<uint(i%8)) > 0 && s.nullMask[j>>3]&(1<<uint(j%8)) == 0 {
			s.nullMask[i>>3] &= ^(1 << uint(i%8))
			s.nullMask[j>>3] |= 1 << uint(j%8)
		} else

		// i is not null, j is null
		if s.nullMask[i>>3]&(1<<uint(i%8)) == 0 && s.nullMask[j>>3]&(1<<uint(j%8)) > 0 {
			s.nullMask[i>>3] |= 1 << uint(i%8)
			s.nullMask[j>>3] &= ^(1 << uint(j%8))
		}
	}

	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s Float64s) Sort() Series {
	if s.sorted != enchanter.SORTED_ASC {
		sort.Sort(s)
		s.sorted = enchanter.SORTED_ASC
	}
	return s
}

func (s Float64s) SortRev() Series {
	if s.sorted != enchanter.SORTED_DESC {
		sort.Sort(sort.Reverse(s))
		s.sorted = enchanter.SORTED_DESC
	}
	return s
}
