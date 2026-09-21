package series

import (
	"fmt"
	"sort"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

// Bools represents a series of bools.
type Bools struct {
	isNullable bool
	sorted     enchanter.SeriesSortOrder
	data       []bool
	nullMask   []uint8
	partition  *SeriesBoolPartition
	ctx        *enchanter.Context
}

// ArrowArray builds and returns a fresh Arrow array from the series data.
// The caller owns the returned array; releasing it is optional under
// GC-backed allocators (see enchanter.Context.Allocator).
func (s Bools) ArrowArray() arrow.Array {
	return buildArrowBoolean(s.ctx.Allocator, s.data, s.isNullable, s.nullMask)
}

// Get the element at index i as a string.
func (s Bools) GetAsString(i int) string {
	if s.isNullable && s.nullMask[i>>3]&(1<<uint(i%8)) != 0 {
		return enchanter.NA_TEXT
	} else if s.data[i] {
		return enchanter.BOOL_TRUE_TEXT
	} else {
		return enchanter.BOOL_FALSE_TEXT
	}
}

// Set the element at index i. The value must be of type bool or NullableBool.
func (s Bools) Set(i int, v any) Series {
	if s.partition != nil {
		return Errors{"Bools.Set: cannot set values in a grouped series"}
	}

	switch v := v.(type) {
	case nil:
		s = s.MakeNullable().(Bools)
		s.nullMask[i>>3] |= 1 << uint(i%8)

	case bool:
		s.data[i] = v

	case enchanter.NullableBool:
		s = s.MakeNullable().(Bools)
		if v.Valid {
			s.data[i] = v.Value
		} else {
			s.nullMask[i>>3] |= 1 << uint(i%8)
			s.data[i] = false
		}

	default:
		return Errors{fmt.Sprintf("Bools.Set: invalid type %T", v)}
	}

	s.sorted = enchanter.SORTED_NONE
	return s
}

////////////////////////			ALL DATA ACCESSORS

// Return the underlying data as a slice of bools.
func (s Bools) Bools() []bool {
	return s.data
}

// Return the underlying data as a slice of NullableBool.
func (s Bools) DataAsNullable() any {
	data := make([]enchanter.NullableBool, len(s.data))
	for i, v := range s.data {
		data[i] = enchanter.NullableBool{Valid: !s.IsNull(i), Value: v}
	}
	return data
}

// Return the data as a slice of strings.
func (s Bools) DataAsString() []string {
	data := make([]string, len(s.data))
	if s.isNullable {
		for i, v := range s.data {
			if s.IsNull(i) {
				data[i] = enchanter.NA_TEXT
			} else if v {
				data[i] = enchanter.BOOL_TRUE_TEXT
			} else {
				data[i] = enchanter.BOOL_FALSE_TEXT
			}
		}
	} else {
		for i, v := range s.data {
			if v {
				data[i] = enchanter.BOOL_TRUE_TEXT
			} else {
				data[i] = enchanter.BOOL_FALSE_TEXT
			}
		}
	}
	return data
}

// Cast the series to a given type.
func (s Bools) Cast(t meta.BaseType) Series {
	switch t {
	case meta.BoolType:
		return s

	case meta.IntType:
		data := make([]int, len(s.data))
		for i, v := range s.data {
			if v {
				data[i] = 1
			}
		}

		return Ints{
			isNullable: s.isNullable,
			sorted:     s.sorted,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Int64Type:
		data := make([]int64, len(s.data))
		for i, v := range s.data {
			if v {
				data[i] = 1
			}
		}

		return Int64s{
			isNullable: s.isNullable,
			sorted:     s.sorted,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Float64Type:
		data := make([]float64, len(s.data))
		for i, v := range s.data {
			if v {
				data[i] = 1
			}
		}

		return Float64s{
			isNullable: s.isNullable,
			sorted:     s.sorted,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.StringType:
		data := make([]*string, len(s.data))

		naTextPtr := s.ctx.StringPool.Put(enchanter.NA_TEXT)
		trueTextPtr := s.ctx.StringPool.Put(enchanter.BOOL_TRUE_TEXT)
		falseTextPtr := s.ctx.StringPool.Put(enchanter.BOOL_FALSE_TEXT)

		if s.isNullable {
			for i, v := range s.data {
				if s.IsNull(i) {
					data[i] = naTextPtr
				} else if v {
					data[i] = trueTextPtr
				} else {
					data[i] = falseTextPtr
				}
			}
		} else {
			for i, v := range s.data {
				if v {
					data[i] = trueTextPtr
				} else {
					data[i] = falseTextPtr
				}
			}
		}

		return Strings{
			isNullable: s.isNullable,
			sorted:     s.sorted,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	default:
		return Errors{fmt.Sprintf("Bools.Cast: invalid type %s", t.String())}
	}
}

////////////////////////			GROUPING OPERATIONS

// A SeriesBoolPartition is a partition of a Bools.
// Each key is a hash of a bool value, and each value is a slice of indices
// of the original series that are set to that value.
type SeriesBoolPartition struct {
	partition map[int64][]int
}

func (gp *SeriesBoolPartition) GetSize() int {
	return len(gp.partition)
}

func (gp *SeriesBoolPartition) GetMap() map[int64][]int {
	return gp.partition
}

func (s Bools) Group() Series {

	// Define the worker callback
	worker := func(threadNum, start, end int, map_ map[int64][]int) {
		for i := start; i < end; i++ {
			if s.data[i] {
				map_[1] = append(map_[1], i)
			} else {
				map_[0] = append(map_[0], i)
			}
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		for i := start; i < end; i++ {
			if s.IsNull(i) {
				(*nulls) = append((*nulls), i)
			} else if s.data[i] {
				map_[1] = append(map_[1], i)
			} else {
				map_[0] = append(map_[0], i)
			}

		}
	}

	partition := SeriesBoolPartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, s.Len(), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &partition

	return s
}

func (s Bools) GroupBy(partition SeriesPartition) Series {
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
				if s.data[index] {
					newHash = (1 + enchanter.HASH_MAGIC_NUMBER) + (h << 13) + (h >> 4)
				} else {
					newHash = enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
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
				} else if s.data[index] {
					newHash = (1 + enchanter.HASH_MAGIC_NUMBER) + (h << 13) + (h >> 4)
				} else {
					newHash = enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	newPartition := SeriesBoolPartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, len(keys), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &newPartition

	return s
}

////////////////////////			SORTING OPERATIONS

func (s Bools) Less(i, j int) bool {
	if s.isNullable {
		if s.nullMask[i>>3]&(1<<uint(i%8)) > 0 {
			return false
		}
		if s.nullMask[j>>3]&(1<<uint(j%8)) > 0 {
			return true
		}
	}
	return !s.data[i] && s.data[j]
}

func (s Bools) Equal(i, j int) bool {
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

func (s Bools) Swap(i, j int) {
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

func (s Bools) Sort() Series {
	if s.sorted != enchanter.SORTED_ASC {
		sort.Sort(s)
		s.sorted = enchanter.SORTED_ASC
	}
	return s
}

func (s Bools) SortRev() Series {
	if s.sorted != enchanter.SORTED_DESC {
		sort.Sort(sort.Reverse(s))
		s.sorted = enchanter.SORTED_DESC
	}
	return s
}
