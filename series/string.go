package series

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/utils"
)

// Strings represents a series of strings.
type Strings struct {
	isNullable bool
	sorted     enchanter.SeriesSortOrder
	data       []*string
	nullMask   []uint8
	partition  *SeriesStringPartition
	ctx        *enchanter.Context
}

// ArrowArray builds and returns a fresh Arrow array from the series data.
// The caller owns the returned array; releasing it is optional under
// GC-backed allocators (see enchanter.Context.Allocator).
func (s Strings) ArrowArray() arrow.Array {
	return buildArrowString(s.ctx.Allocator, s.data, s.isNullable, s.nullMask)
}

// Get the element at index i as a string.
func (s Strings) GetAsString(i int) string {
	return *s.data[i]
}

// Set the element at index i. The value v must be of type string or NullableString.
func (s Strings) Set(i int, v any) Series {
	if s.partition != nil {
		return Errors{"Strings.Set: cannot set values on a grouped Series"}
	}

	switch v := v.(type) {
	case nil:
		s = s.MakeNullable().(Strings)
		s.data[i] = s.ctx.StringPool.Put(enchanter.NA_TEXT)
		s.nullMask[i>>3] |= 1 << uint(i%8)

	case string:
		s.data[i] = s.ctx.StringPool.Put(v)

	case enchanter.NullableString:
		s = s.MakeNullable().(Strings)
		if v.Valid {
			s.data[i] = s.ctx.StringPool.Put(v.Value)
		} else {
			s.data[i] = s.ctx.StringPool.Put(enchanter.NA_TEXT)
			s.nullMask[i>>3] |= 1 << uint(i%8)
		}

	default:
		return Errors{fmt.Sprintf("Strings.Set: invalid type %T", v)}
	}

	s.sorted = enchanter.SORTED_NONE
	return s
}

////////////////////////			ALL DATA ACCESSORS

// Return the underlying data as a slice of string.
// Interned returns the backing slice of interned string pointers as a
// view. Null elements hold the interned NA text. Callers must not
// modify it.
func (s Strings) Interned() []*string {
	return s.data
}

func (s Strings) Strings() []string {
	data := make([]string, len(s.data))
	if s.isNullable {
		for i, v := range s.data {
			if s.IsNull(i) {
				data[i] = enchanter.NA_TEXT
			} else {
				data[i] = *v
			}
		}
	} else {
		for i, v := range s.data {
			data[i] = *v
		}
	}
	return data
}

// Return the underlying data as a slice of NullableString.
func (s Strings) DataAsNullable() any {
	data := make([]enchanter.NullableString, len(s.data))
	for i, v := range s.data {
		data[i] = enchanter.NullableString{Valid: !s.IsNull(i), Value: *v}
	}
	return data
}

// Return the underlying data as a slice of string.
func (s Strings) DataAsString() []string {
	data := make([]string, len(s.data))
	if s.isNullable {
		for i, v := range s.data {
			if s.IsNull(i) {
				data[i] = enchanter.NA_TEXT
			} else {
				data[i] = *v
			}
		}
	} else {
		for i, v := range s.data {
			data[i] = *v
		}
	}
	return data
}

func atoBool(s string) (bool, error) {
	trueRegex := regexp.MustCompile(`^[Tt]([Rr][Uu][Ee])?$`)
	falseRegex := regexp.MustCompile(`^[Ff]([Aa][Ll][Ss][Ee])?$`)

	if trueRegex.MatchString(s) {
		return true, nil
	} else if falseRegex.MatchString(s) {
		return false, nil
	}
	return false, fmt.Errorf("cannot convert \"%s\" to bool", s)
}

// Casts the series to a given type.
func (s Strings) Cast(t meta.BaseType) Series {
	switch t {
	case meta.BoolType:
		data := make([]bool, len(s.data))
		nullMask := utils.BinVecInit(len(s.data), false)
		if s.isNullable {
			copy(nullMask, s.nullMask)
		}

		if s.isNullable {
			for i, v := range s.data {
				if !s.IsNull(i) {
					b, err := atoBool(*v)
					if err != nil {
						nullMask[i>>3] |= (1 << uint(i%8))
					}
					data[i] = b
				}
			}
		} else {
			for i, v := range s.data {
				b, err := atoBool(*v)
				if err != nil {
					nullMask[i>>3] |= (1 << uint(i%8))
				}
				data[i] = b
			}
		}

		return Bools{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.IntType:
		data := make([]int, len(s.data))
		nullMask := utils.BinVecInit(len(s.data), false)
		if s.isNullable {
			copy(nullMask, s.nullMask)
		}

		if s.isNullable {
			for i, v := range s.data {
				if !s.IsNull(i) {
					d, err := strconv.Atoi(*v)
					if err != nil {
						nullMask[i>>3] |= (1 << uint(i%8))
					} else {
						data[i] = int(d)
					}
				}
			}
		} else {
			for i, v := range s.data {
				d, err := strconv.Atoi(*v)
				if err != nil {
					nullMask[i>>3] |= (1 << uint(i%8))
				} else {
					data[i] = int(d)
				}
			}
		}

		return Ints{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Int64Type:
		data := make([]int64, len(s.data))
		nullMask := utils.BinVecInit(len(s.data), false)
		if s.isNullable {
			copy(nullMask, s.nullMask)
		}

		if s.isNullable {
			for i, v := range s.data {
				if !s.IsNull(i) {
					d, err := strconv.ParseInt(*v, 10, 64)
					if err != nil {
						nullMask[i>>3] |= (1 << uint(i%8))
					} else {
						data[i] = d
					}
				}
			}
		} else {
			for i, v := range s.data {
				d, err := strconv.ParseInt(*v, 10, 64)
				if err != nil {
					nullMask[i>>3] |= (1 << uint(i%8))
				} else {
					data[i] = d
				}
			}
		}

		return Int64s{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.Float64Type:
		data := make([]float64, len(s.data))
		nullMask := utils.BinVecInit(len(s.data), false)
		if s.isNullable {
			copy(nullMask, s.nullMask)
		}

		if s.isNullable {
			for i, v := range s.data {
				if !s.IsNull(i) {
					f, err := strconv.ParseFloat(*v, 64)
					if err != nil {
						nullMask[i>>3] |= (1 << uint(i%8))
					} else {
						data[i] = f
					}
				}
			}
		} else {
			for i, v := range s.data {
				f, err := strconv.ParseFloat(*v, 64)
				if err != nil {
					nullMask[i>>3] |= (1 << uint(i%8))
				} else {
					data[i] = f
				}
			}
		}

		return Float64s{
			isNullable: true,
			sorted:     enchanter.SORTED_NONE,
			data:       data,
			nullMask:   nullMask,
			partition:  nil,
			ctx:        s.ctx,
		}

	case meta.StringType:
		return s

	case meta.TimeType:
		return Errors{"Strings.Cast: cannot cast to Time, use Strings.ParseTime(layout) instead"}

	default:
		return Errors{fmt.Sprintf("Strings.Cast: invalid type %s", t.String())}
	}
}

// Parse the series as a time series.
func (s Strings) ParseTime(layout string) Series {
	data := make([]time.Time, len(s.data))
	nullMask := utils.BinVecInit(len(s.data), false)
	if s.isNullable {
		copy(nullMask, s.nullMask)
	}

	for i, v := range s.data {
		if s.isNullable && s.IsNull(i) {
			continue
		}
		t, err := time.Parse(layout, *v)
		if err != nil {
			nullMask[i>>3] |= (1 << uint(i%8))
		} else {
			data[i] = t
		}
	}

	return Times{
		isNullable: true,
		sorted:     enchanter.SORTED_NONE,
		data:       data,
		nullMask:   nullMask,
		partition:  nil,
		ctx:        s.ctx,
	}
}

////////////////////////			GROUPING OPERATIONS

// A SeriesStringPartition is a partition of a Strings.
// Each key is a hash of a bool value, and each value is a slice of indices
// of the original series that are set to that value.
type SeriesStringPartition struct {
	partition map[int64][]int
	ctx       *enchanter.Context
}

func (gp *SeriesStringPartition) GetSize() int {
	return len(gp.partition)
}

func (gp *SeriesStringPartition) GetMap() map[int64][]int {
	return gp.partition
}

func (s Strings) Group() Series {

	// Define the worker callback
	worker := func(threadNum, start, end int, map_ map[int64][]int) {
		var ptr unsafe.Pointer
		for i := start; i < end; i++ {
			ptr = unsafe.Pointer(s.data[i])
			map_[(*(*int64)(unsafe.Pointer(&ptr)))] = append(map_[(*(*int64)(unsafe.Pointer(&ptr)))], i)
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		var ptr unsafe.Pointer
		for i := start; i < end; i++ {
			if s.IsNull(i) {
				(*nulls) = append((*nulls), i)
			} else {
				ptr = unsafe.Pointer(s.data[i])
				map_[(*(*int64)(unsafe.Pointer(&ptr)))] = append(map_[(*(*int64)(unsafe.Pointer(&ptr)))], i)
			}
		}
	}

	partition := SeriesStringPartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, len(s.data), s.HasNull(),
			worker, workerNulls),
		ctx: s.ctx,
	}

	s.partition = &partition

	return s
}

func (s Strings) GroupBy(partition SeriesPartition) Series {
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
		var ptr unsafe.Pointer
		for _, h := range keys[start:end] { // keys is defined outside the function
			for _, index := range otherIndeces[h] { // otherIndeces is defined outside the function
				ptr = unsafe.Pointer(s.data[index])
				newHash = *(*int64)(unsafe.Pointer(&ptr)) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	// Define the worker callback for nulls
	workerNulls := func(threadNum, start, end int, map_ map[int64][]int, nulls *[]int) {
		var newHash int64
		var ptr unsafe.Pointer
		for _, h := range keys[start:end] { // keys is defined outside the function
			for _, index := range otherIndeces[h] { // otherIndeces is defined outside the function
				if s.IsNull(index) {
					newHash = enchanter.HASH_MAGIC_NUMBER_NULL + (h << 13) + (h >> 4)
				} else {
					ptr = unsafe.Pointer(s.data[index])
					newHash = *(*int64)(unsafe.Pointer(&ptr)) + enchanter.HASH_MAGIC_NUMBER + (h << 13) + (h >> 4)
				}
				map_[newHash] = append(map_[newHash], index)
			}
		}
	}

	newPartition := SeriesStringPartition{
		partition: seriesGroupBy(
			enchanter.THREADS_NUMBER, enchanter.MINIMUM_PARALLEL_SIZE_1, len(keys), s.HasNull(),
			worker, workerNulls),
		ctx: s.ctx,
	}

	s.partition = &newPartition

	return s
}

////////////////////////			SORTING OPERATIONS

func (s Strings) Less(i, j int) bool {
	if s.isNullable {
		if s.nullMask[i>>3]&(1<<uint(i%8)) > 0 {
			return false
		}
		if s.nullMask[j>>3]&(1<<uint(j%8)) > 0 {
			return true
		}
	}

	return (*s.data[i]) < (*s.data[j])
}

func (s Strings) Equal(i, j int) bool {
	if s.isNullable {
		if (s.nullMask[i>>3] & (1 << uint(i%8))) > 0 {
			return (s.nullMask[j>>3] & (1 << uint(j%8))) > 0
		}
		if (s.nullMask[j>>3] & (1 << uint(j%8))) > 0 {
			return false
		}
	}

	return (*s.data[i]) == (*s.data[j])
}

func (s Strings) Swap(i, j int) {
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

func (s Strings) Sort() Series {
	if s.sorted != enchanter.SORTED_ASC {
		sort.Sort(s)
		s.sorted = enchanter.SORTED_ASC
	}
	return s
}

func (s Strings) SortRev() Series {
	if s.sorted != enchanter.SORTED_DESC {
		sort.Sort(sort.Reverse(s))
		s.sorted = enchanter.SORTED_DESC
	}
	return s
}

////////////////////////			STRING OPERATIONS

func (s Strings) ToUpper() Series {
	if s.partition != nil {
		return Errors{"Strings.ToUpper() not supported on grouped Series"}
	}

	for i := 0; i < len(s.data); i++ {
		s.data[i] = s.ctx.StringPool.Put(strings.ToUpper(*s.data[i]))
	}

	return s
}

func (s Strings) ToLower() Series {
	if s.partition != nil {
		return Errors{"Strings.ToLower() not supported on grouped Series"}
	}

	for i := 0; i < len(s.data); i++ {
		s.data[i] = s.ctx.StringPool.Put(strings.ToLower(*s.data[i]))
	}

	return s
}

func (s Strings) TrimSpace() Series {
	if s.partition != nil {
		return Errors{"Strings.TrimSpace() not supported on grouped Series"}
	}

	for i := 0; i < len(s.data); i++ {
		s.data[i] = s.ctx.StringPool.Put(strings.TrimSpace(*s.data[i]))
	}

	return s
}

func (s Strings) Trim(cutset string) Series {
	if s.partition != nil {
		return Errors{"Strings.Trim() not supported on grouped Series"}
	}

	for i := 0; i < len(s.data); i++ {
		s.data[i] = s.ctx.StringPool.Put(strings.Trim(*s.data[i], cutset))
	}

	return s
}

func (s Strings) Replace(old, new string, n int) Series {
	if s.partition != nil {
		return Errors{"Strings.Replace() not supported on grouped Series"}
	}

	for i := 0; i < len(s.data); i++ {
		s.data[i] = s.ctx.StringPool.Put(strings.Replace(*s.data[i], old, new, n))
	}

	return s
}
