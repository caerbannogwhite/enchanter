package series

import (
	"fmt"
	"sort"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

// Times represents a datetime series.
type Times struct {
	isNullable bool
	sorted     enchanter.SeriesSortOrder
	data       []time.Time
	nullMask   []uint8
	partition  *SeriesTimePartition
	ctx        *enchanter.Context
	timeFormat string
}

// ArrowArray builds and returns a fresh Arrow array from the series data.
// The caller owns the returned array; releasing it is optional under
// GC-backed allocators (see enchanter.Context.Allocator).
func (s Times) ArrowArray() arrow.Array {
	return buildArrowTimestamp(s.ctx.Allocator, s.data, s.isNullable, s.nullMask)
}

// Get the time format of the series.
func (s Times) GetTimeFormat() string {
	return s.timeFormat
}

// Set the time format of the series.
func (s Times) SetTimeFormat(format string) Series {
	s.timeFormat = format
	return s
}

// Get the element at index i as a string.
func (s Times) GetAsString(i int) string {
	if s.isNullable && s.nullMask[i>>3]&(1<<uint(i%8)) != 0 {
		return enchanter.NA_TEXT
	}
	return s.data[i].Format(s.timeFormat)
}

// Set the element at index i. The value v must be of type time.Time or NullableTime.
func (s Times) Set(i int, v any) Series {
	if s.partition != nil {
		return Errors{"Times.Set: cannot set values on a grouped Series"}
	}

	switch v := v.(type) {
	case nil:
		s = s.MakeNullable().(Times)
		s.nullMask[i>>3] |= 1 << uint(i%8)

	case time.Time:
		s.data[i] = v

	case enchanter.NullableTime:
		s = s.MakeNullable().(Times)
		if v.Valid {
			s.data[i] = v.Value
		} else {
			s.data[i] = time.Time{}
			s.nullMask[i/8] |= 1 << uint(i%8)
		}

	default:
		return Errors{fmt.Sprintf("Times.Set: invalid type %T", v)}
	}

	s.sorted = enchanter.SORTED_NONE
	return s
}

////////////////////////			ALL DATA ACCESSORS

// Return the underlying data as a slice of time.Time.
func (s Times) Times() []time.Time {
	return s.data
}

// Return the underlying data as a slice of NullableTime.
func (s Times) DataAsNullable() any {
	data := make([]enchanter.NullableTime, len(s.data))
	for i, v := range s.data {
		data[i] = enchanter.NullableTime{Valid: !s.IsNull(i), Value: v}
	}
	return data
}

// Return the underlying data as a slice of strings.
func (s Times) DataAsString() []string {
	data := make([]string, len(s.data))
	if s.isNullable {
		for i, v := range s.data {
			if s.IsNull(i) {
				data[i] = enchanter.NA_TEXT
			} else {
				data[i] = v.Format(s.timeFormat)
			}
		}
	} else {
		for i, v := range s.data {
			data[i] = v.Format(s.timeFormat)
		}
	}
	return data
}

// Casts the series to a given type.
func (s Times) Cast(t meta.BaseType) Series {
	switch t {
	case meta.BoolType:
		return Errors{fmt.Sprintf("Times.Cast: cannot cast to %s", t.String())}

	case meta.IntType:
		data := make([]int, len(s.data))
		for i, v := range s.data {
			data[i] = int(v.UnixNano())
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
			data[i] = v.UnixNano()
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
			data[i] = float64(v.UnixNano())
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
					data[i] = s.ctx.StringPool.Put(v.Format(s.timeFormat))
				}
			}
		} else {
			for i, v := range s.data {
				data[i] = s.ctx.StringPool.Put(v.Format(s.timeFormat))
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

	case meta.TimeType:
		return s

	case meta.DurationType:
		data := make([]time.Duration, len(s.data))
		for i, v := range s.data {
			data[i] = v.Sub(time.Time{})
		}

		return Durations{
			isNullable: s.isNullable,
			sorted:     s.sorted,
			data:       data,
			nullMask:   s.nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	default:
		return Errors{fmt.Sprintf("Times.Cast: invalid type %T", t)}
	}
}

////////////////////////			GROUPING OPERATIONS

// A SeriesTimePartition is a partition of a Times.
// Each key is a hash of a bool value, and each value is a slice of indices
// of the original series that are set to that value.
type SeriesTimePartition struct {
	partition map[int64][]int
}

func (gp *SeriesTimePartition) GetSize() int {
	return len(gp.partition)
}

func (gp *SeriesTimePartition) GetMap() map[int64][]int {
	return gp.partition
}

func (s Times) Group() Series {

	// Define the worker callback
	worker := func(threadNum, start, end int, map_ map[int64][]int) {
		for i := start; i < end; i++ {
			map_[s.data[i].UnixNano()] = append(map_[s.data[i].UnixNano()], i)
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		for i := start; i < end; i++ {
			if s.IsNull(i) {
				(*nulls) = append((*nulls), i)
			} else {
				map_[s.data[i].UnixNano()] = append(map_[s.data[i].UnixNano()], i)
			}
		}
	}

	partition := SeriesTimePartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, s.Len(), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &partition

	return s
}

func (s Times) GroupBy(partition SeriesPartition) Series {
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
				newHash = s.data[index].UnixNano() + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
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
					newHash = s.data[index].UnixNano() + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	newPartition := SeriesTimePartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, len(keys), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &newPartition

	return s
}

////////////////////////			SORTING OPERATIONS

func (s Times) Less(i, j int) bool {
	if s.isNullable {
		if s.nullMask[i>>3]&(1<<uint(i%8)) > 0 {
			return false
		}
		if s.nullMask[j>>3]&(1<<uint(j%8)) > 0 {
			return true
		}
	}
	return s.data[i].Compare(s.data[j]) < 0
}

func (s Times) Equal(i, j int) bool {
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

func (s Times) Swap(i, j int) {
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

func (s Times) Sort() Series {
	if s.sorted != enchanter.SORTED_ASC {
		sort.Sort(s)
		s.sorted = enchanter.SORTED_ASC
	}
	return s
}

func (s Times) SortRev() Series {
	if s.sorted != enchanter.SORTED_DESC {
		sort.Sort(sort.Reverse(s))
		s.sorted = enchanter.SORTED_DESC
	}
	return s
}
