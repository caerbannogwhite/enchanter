package series

import (
	"fmt"
	"sort"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
)

// Ints represents a series of ints.
type Ints struct {
	isNullable bool
	sorted     enchanter.SeriesSortOrder
	data       []int
	nullMask   []uint8
	partition  *SeriesIntPartition
	ctx        *enchanter.Context
}

// ArrowArray builds and returns a fresh Arrow array from the series data.
// The caller owns the returned array; releasing it is optional under
// GC-backed allocators (see enchanter.Context.Allocator).
func (s Ints) ArrowArray() arrow.Array {
	return buildArrowInt64FromInts(s.ctx.Allocator, s.data, s.isNullable, s.nullMask)
}

// Get the element at index i as a string.
func (s Ints) GetAsString(i int) string {
	if s.isNullable && s.IsNull(i) {
		return enchanter.NA_TEXT
	}
	return intToString(int64(s.data[i]))
}

// Set the element at index i. The value v can be any belonging to types:
// int8, int16, int, int, int64 and their nullable versions.
func (s Ints) Set(i int, v any) Series {
	if s.partition != nil {
		return Errors{"Ints.Set: cannot set values on a grouped Series"}
	}

	switch val := v.(type) {
	case nil:
		s = s.MakeNullable().(Ints)
		s.nullMask[i>>3] |= 1 << uint(i%8)

	case int8:
		s.data[i] = int(val)

	case int16:
		s.data[i] = int(val)

	case int:
		s.data[i] = int(val)

	case enchanter.NullableInt8:
		s = s.MakeNullable().(Ints)
		if v.(enchanter.NullableInt8).Valid {
			s.data[i] = int(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt16:
		s = s.MakeNullable().(Ints)
		if v.(enchanter.NullableInt16).Valid {
			s.data[i] = int(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	case enchanter.NullableInt:
		s = s.MakeNullable().(Ints)
		if v.(enchanter.NullableInt).Valid {
			s.data[i] = int(val.Value)
		} else {
			s.data[i] = 0
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	default:
		return Errors{fmt.Sprintf("Ints.Set: invalid type %T", v)}
	}

	s.sorted = enchanter.SORTED_NONE
	return s
}

////////////////////////			ALL DATA ACCESSORS

// Return the underlying data as a slice of int.
func (s Ints) Ints() []int {
	return s.data
}

// Return the underlying data as a slice of NullableInt.
func (s Ints) DataAsNullable() any {
	data := make([]enchanter.NullableInt, len(s.data))
	for i, v := range s.data {
		data[i] = enchanter.NullableInt{Valid: !s.IsNull(i), Value: v}
	}
	return data
}

// Return the underlying data as a slice of strings.
func (s Ints) DataAsString() []string {
	data := make([]string, len(s.data))
	if s.isNullable {
		for i, v := range s.data {
			if s.IsNull(i) {
				data[i] = enchanter.NA_TEXT
			} else {
				data[i] = intToString(int64(v))
			}
		}
	} else {
		for i, v := range s.data {
			data[i] = intToString(int64(v))
		}
	}
	return data
}

// Casts the series to a given type.
func (s Ints) Cast(t meta.BaseType) Series {
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
		return s

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
		data := make([]float64, len(s.data))
		for i, v := range s.data {
			data[i] = float64(v)
		}

		return Float64s{
			isNullable: s.isNullable,
			sorted:     enchanter.SORTED_NONE,
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
					data[i] = s.ctx.StringPool.Put(intToString(int64(v)))
				}
			}
		} else {
			for i, v := range s.data {
				data[i] = s.ctx.StringPool.Put(intToString(int64(v)))
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
		return Errors{fmt.Sprintf("Ints.Cast: invalid type %s", t.String())}
	}
}

////////////////////////			GROUPING OPERATIONS

// A SeriesIntPartition is a partition of a Ints.
// Each key is a hash of a bool value, and each value is a slice of indices
// of the original series that are set to that value.
type SeriesIntPartition struct {
	partition           map[int64][]int
	isDense             bool
	partitionDenseMin   int
	partitionDense      [][]int
	partitionDenseNulls []int
}

func (gp *SeriesIntPartition) GetSize() int {
	if gp.isDense {
		if gp.partitionDenseNulls != nil && len(gp.partitionDenseNulls) > 0 {
			return len(gp.partitionDense) + 1
		}
		return len(gp.partitionDense)
	}
	return len(gp.partition)
}

func (gp *SeriesIntPartition) GetMap() map[int64][]int {
	if gp.isDense {
		map_ := make(map[int64][]int, len(gp.partitionDense))
		for i, part := range gp.partitionDense {
			map_[int64(i)+int64(gp.partitionDenseMin)] = part
		}

		// Merge the nulls to the map
		if gp.partitionDenseNulls != nil && len(gp.partitionDenseNulls) > 0 {
			nullKey := seriesNullKey(map_, enchanter.HASH_NULL_KEY)
			map_[nullKey] = gp.partitionDenseNulls
		}

		return map_
	}

	return gp.partition
}

func (s Ints) Group() Series {
	var useDenseMap bool
	var min, max int
	var partition SeriesIntPartition

	// If the number of elements is small,
	// look for the minimum and maximum values
	// if len(s.data) < MINIMUM_PARALLEL_SIZE_2 {
	// 	useDenseMap = true
	// 	max = s.data[0]
	// 	min = s.data[0]
	// 	for _, v := range s.data {
	// 		if v > max {
	// 			max = v
	// 		}
	// 		if v < min {
	// 			min = v
	// 		}
	// 	}
	// }

	// If the difference between the maximum and minimum values is acceptable,
	// then we can use a dense map, otherwise we use a sparse map
	if useDenseMap && (max-min >= enchanter.MINIMUM_PARALLEL_SIZE_1) {
		useDenseMap = false //nolint:ineffassign // unreachable while the dense-map fast path above stays disabled (see TODO: FIX DENSE MAP below); kept so this guard is correct once that path is restored
	}

	// TODO: FIX DENSE MAP
	// if useDenseMap {
	// 	var nulls []int
	// 	map_ := make([][]int, max-min+1)
	// 	for i := 0; i < len(map_); i++ {
	// 		map_[i] = make([]int, 0, DEFAULT_DENSE_MAP_ARRAY_INITIAL_CAPACITY)
	// 	}

	// 	if s.HasNull() {
	// 		nulls = make([]int, 0, DEFAULT_DENSE_MAP_ARRAY_INITIAL_CAPACITY)
	// 		for i, v := range s.data {
	// 			if s.IsNull(i) {
	// 				nulls = append(nulls, i)
	// 			} else {
	// 				map_[v-min] = append(map_[v-min], i)
	// 			}
	// 		}
	// 	} else {
	// 		for i, v := range s.data {
	// 			map_[v-min] = append(map_[v-min], i)
	// 		}
	// 	}

	// 	partition = SeriesIntPartition{
	// 		isDense:             true,
	// 		partitionDenseMin:   min,
	// 		partitionDense:      map_,
	// 		partitionDenseNulls: nulls,
	// 	}
	// } else

	// SPARSE MAP
	{
		// Define the worker callback
		worker := func(threadNum, start, end int, map_ map[int64][]int) {
			up := end - ((end - start) % 8)
			for i := start; i < up; {
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				i++
			}

			for i := up; i < end; i++ {
				map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
			}
		}

		// Define the worker callback for nulls
		workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
			for i := start; i < end; i++ {
				if s.IsNull(i) {
					(*nulls) = append((*nulls), i)
				} else {
					map_[int64(s.data[i])] = append(map_[int64(s.data[i])], i)
				}
			}
		}

		partition = SeriesIntPartition{
			isDense: false,
			partition: seriesGroupBy(
				enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_2, len(s.data), s.HasNull(),
				worker, workerNulls),
		}
	}

	s.partition = &partition

	return s
}

func (s Ints) GroupBy(partition SeriesPartition) Series {
	if partition == nil {
		return s
	}

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
				newHash = int64(s.data[index]) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
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
					newHash = int64(s.data[index]) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	newPartition := SeriesIntPartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_2, len(keys), s.HasNull(),
			worker, workerNulls),
	}

	s.partition = &newPartition

	return s
}

////////////////////////			SORTING OPERATIONS

func (s Ints) Less(i, j int) bool {
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

func (s Ints) Equal(i, j int) bool {
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

func (s Ints) Swap(i, j int) {
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

func (s Ints) Sort() Series {
	if s.sorted != enchanter.SORTED_ASC {
		sort.Sort(s)
		s.sorted = enchanter.SORTED_ASC
	}
	return s
}

func (s Ints) SortRev() Series {
	if s.sorted != enchanter.SORTED_DESC {
		sort.Sort(sort.Reverse(s))
		s.sorted = enchanter.SORTED_DESC
	}
	return s
}
