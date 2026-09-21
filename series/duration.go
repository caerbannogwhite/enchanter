package series

import (
	"fmt"
	"sort"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

// Durations represents a duration series.
type Durations struct {
	isNullable bool
	sorted     enchanter.SeriesSortOrder
	data       []time.Duration
	nullMask   []uint8
	partition  *SeriesDurationPartition
	ctx        *enchanter.Context
}

// ArrowArray builds and returns a fresh Arrow array from the series data.
// The caller owns the returned array; releasing it is optional under
// GC-backed allocators (see enchanter.Context.Allocator).
func (s Durations) ArrowArray() arrow.Array {
	return buildArrowDuration(s.ctx.Allocator, s.data, s.isNullable, s.nullMask)
}

// Get the element at index i as a string.
func (s Durations) GetAsString(i int) string {
	if s.isNullable && s.nullMask[i>>3]&(1<<uint(i%8)) != 0 {
		return enchanter.NA_TEXT
	}
	return s.data[i].String()
}

// Set the element at index i. The value v must be of type time.Duration or NullableDuration.
func (s Durations) Set(i int, v any) Series {
	if s.partition != nil {
		return Errors{"Durations.Set: cannot set values on a grouped Series"}
	}

	switch v := v.(type) {
	case nil:
		s = s.MakeNullable().(Durations)
		s.nullMask[i>>3] |= 1 << uint(i%8)

	case time.Duration:
		s.data[i] = v

	case enchanter.NullableDuration:
		s = s.MakeNullable().(Durations)
		if v.Valid {
			s.data[i] = v.Value
		} else {
			s.data[i] = time.Duration(0)
			s.nullMask[i/8] |= 1 << uint(i%8)
		}

	default:
		return Errors{fmt.Sprintf("Durations.Set: invalid type %T", v)}
	}

	s.sorted = enchanter.SORTED_NONE
	return s
}

////////////////////////			ALL DATA ACCESSORS

// Return the underlying data as a slice of time.Duration.
func (s Durations) Durations() []time.Duration {
	return s.data
}

// Return the underlying data as a slice of NullableDuration.
func (s Durations) DataAsNullable() any {
	data := make([]enchanter.NullableDuration, len(s.data))
	for i, v := range s.data {
		data[i] = enchanter.NullableDuration{Valid: !s.IsNull(i), Value: v}
	}
	return data
}

// Return the underlying data as a slice of strings.
func (s Durations) DataAsString() []string {
	data := make([]string, len(s.data))
	if s.isNullable {
		for i, v := range s.data {
			if s.IsNull(i) {
				data[i] = enchanter.NA_TEXT
			} else {
				data[i] = v.String()
			}
		}
	} else {
		for i, v := range s.data {
			data[i] = v.String()
		}
	}
	return data
}

// Casts the series to a given type.
func (s Durations) Cast(t meta.BaseType) Series {

	switch t {
	case meta.BoolType:
		return Errors{fmt.Sprintf("Durations.Cast: cannot cast to %s", t.String())}

	case meta.IntType:
		data := make([]int, len(s.data))
		for i, v := range s.data {
			data[i] = int(v.Nanoseconds())
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
			data[i] = v.Nanoseconds()
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
			data[i] = float64(v.Nanoseconds())
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
		if s.isNullable {
			for i, v := range s.data {
				if s.IsNull(i) {
					data[i] = s.ctx.StringPool.Put(enchanter.NA_TEXT)
				} else {
					data[i] = s.ctx.StringPool.Put(v.String())
				}
			}
		} else {
			for i, v := range s.data {
				data[i] = s.ctx.StringPool.Put(v.String())
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
		return Errors{fmt.Sprintf("Durations.Cast: invalid type %T", t)}
	}
}

////////////////////////			GROUPING OPERATIONS

// A SeriesDurationPartition is a partition of a Durations.
// Each key is a hash of a bool value, and each value is a slice of indices
// of the original series that are set to that value.
type SeriesDurationPartition struct {
	partition map[int64][]int
}

func (gp *SeriesDurationPartition) GetSize() int {
	return len(gp.partition)
}

func (gp *SeriesDurationPartition) GetMap() map[int64][]int {
	return gp.partition
}

func (s Durations) Group() Series {

	// Define the worker callback
	worker := func(threadNum, start, end int, map_ map[int64][]int) {
		for i := start; i < end; i++ {
			map_[s.data[i].Nanoseconds()] = append(map_[s.data[i].Nanoseconds()], i)
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		for i := start; i < end; i++ {
			if s.IsNull(i) {
				(*nulls) = append((*nulls), i)
			} else {
				map_[s.data[i].Nanoseconds()] = append(map_[s.data[i].Nanoseconds()], i)
			}
		}
	}

	partition := SeriesDurationPartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, s.Len(), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &partition

	return s
}

func (s Durations) GroupBy(partition SeriesPartition) Series {
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
				newHash = s.data[index].Nanoseconds() + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
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
					newHash = s.data[index].Nanoseconds() + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	newPartition := SeriesDurationPartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, len(keys), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &newPartition

	return s
}

////////////////////////			SORTING OPERATIONS

func (s Durations) Less(i, j int) bool {
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

func (s Durations) Equal(i, j int) bool {
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

func (s Durations) Swap(i, j int) {
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

func (s Durations) Sort() Series {
	if s.sorted != enchanter.SORTED_ASC {
		sort.Sort(s)
		s.sorted = enchanter.SORTED_ASC
	}
	return s
}

func (s Durations) SortRev() Series {
	if s.sorted != enchanter.SORTED_DESC {
		sort.Sort(sort.Reverse(s))
		s.sorted = enchanter.SORTED_DESC
	}
	return s
}
