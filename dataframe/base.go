package dataframe

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/formatter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/series"
	"github.com/caerbannogwhite/enchanter/utils"
)

type DataFramePartitionEntry struct {
	index     int
	name      string
	partition series.SeriesPartition
}

type DataFrame struct {
	isGrouped    bool
	err          error
	names        []string
	series       []series.Series
	partitions   []DataFramePartitionEntry
	groupByNames []string
	sortParams   []SortParam
	ctx          *enchanter.Context
}

func NewDataFrame(ctx *enchanter.Context) DataFrame {
	if ctx == nil {
		return DataFrame{err: fmt.Errorf("NewDataFrame: context is nil")}
	}

	return DataFrame{
		series: make([]series.Series, 0),
		ctx:    ctx,
	}
}

////////////////////////			BASIC ACCESSORS

// GetContext returns the context of the dataframe.
func (df DataFrame) Context() *enchanter.Context {
	return df.ctx
}

// Names returns the names of the series in the dataframe.
func (df DataFrame) Names() []string {
	return df.names
}

// Types returns the types of the series in the dataframe.
func (df DataFrame) Types() []meta.BaseType {
	types := make([]meta.BaseType, len(df.series))
	for i, series := range df.series {
		types[i] = series.Type()
	}
	return types
}

// NCols returns the number of columns in the dataframe.
func (df DataFrame) NCols() int {
	return len(df.series)
}

// NRows returns the number of rows in the dataframe.
func (df DataFrame) NRows() int {
	if len(df.series) == 0 {
		return 0
	}
	return df.series[0].Len()
}

// Err returns the error carried by the frame; nil when the frame is
// healthy. Operations never panic: a failure travels with the returned
// frame and is checked once at the end of a chain.
func (df DataFrame) Err() error {
	return df.err
}

func (df DataFrame) IsGrouped() bool {
	return df.isGrouped
}

// ColIndex returns the position of the named column, -1 when absent.
func (df DataFrame) ColIndex(name string) int {
	for i, name_ := range df.names {
		if name_ == name {
			return i
		}
	}
	return -1
}

func (df DataFrame) AddSeries(name string, series series.Series) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeries: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && series.Len() != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeries: series length (%d) does not match dataframe length (%d)", series.Len(), df.NRows())
		return df
	}

	df.names = append(df.names, name)
	df.series = append(df.series, series)

	return df
}

func (df DataFrame) AddSeriesFromBools(name string, data []bool, nullMask []bool, makeCopy bool) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromBools: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && len(data) != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromBools: series length (%d) does not match dataframe length (%d)", len(data), df.NRows())
		return df
	}

	return df.AddSeries(name, series.NewSeriesBool(data, nullMask, makeCopy, df.ctx))
}

func (df DataFrame) AddSeriesFromInts(name string, data []int, nullMask []bool, makeCopy bool) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromInts: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && len(data) != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromInts: series length (%d) does not match dataframe length (%d)", len(data), df.NRows())
		return df
	}

	return df.AddSeries(name, series.NewSeriesInt(data, nullMask, makeCopy, df.ctx))
}

func (df DataFrame) AddSeriesFromInt64s(name string, data []int64, nullMask []bool, makeCopy bool) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromInt64s: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && len(data) != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromInt64s: series length (%d) does not match dataframe length (%d)", len(data), df.NRows())
		return df
	}

	return df.AddSeries(name, series.NewSeriesInt64(data, nullMask, makeCopy, df.ctx))
}

func (df DataFrame) AddSeriesFromFloat64s(name string, data []float64, nullMask []bool, makeCopy bool) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromFloat64s: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && len(data) != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromFloat64s: series length (%d) does not match dataframe length (%d)", len(data), df.NRows())
		return df
	}

	return df.AddSeries(name, series.NewSeriesFloat64(data, nullMask, makeCopy, df.ctx))
}

func (df DataFrame) AddSeriesFromStrings(name string, data []string, nullMask []bool, makeCopy bool) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromStrings: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && len(data) != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromStrings: series length (%d) does not match dataframe length (%d)", len(data), df.NRows())
		return df
	}

	return df.AddSeries(name, series.NewSeriesString(data, nullMask, makeCopy, df.ctx))
}

func (df DataFrame) AddSeriesFromTimes(name string, data []time.Time, nullMask []bool, makeCopy bool) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromTimes: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && len(data) != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromTimes: series length (%d) does not match dataframe length (%d)", len(data), df.NRows())
		return df
	}

	return df.AddSeries(name, series.NewSeriesTime(data, nullMask, makeCopy, df.ctx))
}

func (df DataFrame) AddSeriesFromDurations(name string, data []time.Duration, nullMask []bool, makeCopy bool) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromDurations: cannot add series to a grouped dataframe")
		return df
	}

	if df.NCols() > 0 && len(data) != df.NRows() {
		df.err = fmt.Errorf("DataFrame.AddSeriesFromDurations: series length (%d) does not match dataframe length (%d)", len(data), df.NRows())
		return df
	}

	return df.AddSeries(name, series.NewSeriesDuration(data, nullMask, makeCopy, df.ctx))
}

func (df DataFrame) Replace(name string, s series.Series) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.Replace: cannot replace series in a grouped dataframe")
		return df
	}

	index := df.ColIndex(name)
	if index == -1 {
		df.err = fmt.Errorf("DataFrame.Replace: series \"%s\" not found", name)
		return df
	}

	if s.Len() != df.NRows() {
		df.err = fmt.Errorf("DataFrame.Replace: series length (%d) does not match dataframe length (%d)", s.Len(), df.NRows())
		return df
	}

	df.series[index] = s
	return df
}

// Col returns the column with the given name.
func (df DataFrame) Col(name string) series.Series {
	for i, name_ := range df.names {
		if name_ == name {
			return df.series[i]
		}
	}

	return series.Errors{Msg_: fmt.Sprintf("DataFrame.C: series \"%s\" not found", name)}
}

// Returns the series with the given name.
// For internal use only: returns nil if the series is not found.
func (df DataFrame) seriesByName(name string) series.Series {
	for i, name_ := range df.names {
		if name_ == name {
			return df.series[i]
		}
	}

	return nil
}

// ColAt returns the column at the given index.
func (df DataFrame) ColAt(index int) series.Series {
	if index < 0 || index >= len(df.series) {
		return series.Errors{Msg_: fmt.Sprintf("DataFrame.ColAt: index %d out of bounds", index)}
	}
	return df.series[index]
}

// Returns the name of the series at the given index, or the empty string
// when the index is out of bounds.
func (df DataFrame) NameAt(index int) string {
	if index < 0 || index >= len(df.names) {
		return ""
	}
	return df.names[index]
}

// Select returns a DataFrame holding the named columns, in the order given.
// Names are matched exactly; a name repeated in the argument list is kept once,
// at its first position.
//
//	df.Select("Car", "Origin") // exactly those two columns
//
// It is an error to name a column the DataFrame does not have — a typo yields
// an errored DataFrame rather than a silently missing column. Use
// SelectMatching for pattern-based selection.
func (df DataFrame) Select(names ...string) DataFrame {
	if df.err != nil {
		return df
	}

	have := make(map[string]bool, len(df.names))
	for _, name := range df.names {
		have[name] = true
	}

	taken := make(map[string]bool, len(names))
	outNames := make([]string, 0, len(names))
	seriesList := make([]series.Series, 0, len(names))

	for _, name := range names {
		if !have[name] {
			df.err = fmt.Errorf("DataFrame.Select: column \"%s\" not found", name)
			return df
		}
		if taken[name] {
			continue
		}
		taken[name] = true
		outNames = append(outNames, name)
		seriesList = append(seriesList, df.Col(name))
	}

	return DataFrame{
		names:  outNames,
		series: seriesList,
		ctx:    df.ctx,
	}
}

// SelectMatching returns a DataFrame holding the columns matched by patterns,
// in pattern order; a column matched by several patterns is kept once, at its
// first match.
//
// Each pattern is a regular expression, matched UNANCHORED against the column
// name, so "Car" also selects a column named "CarOrigin". Anchor a pattern to
// pin it to a whole name:
//
//	df.SelectMatching("^EX.*1$", "_RAW")
//
// An invalid regular expression leaves the returned DataFrame in an error
// state. A pattern that matches nothing contributes no columns and is not an
// error, since a pattern is not a claim that a particular column exists.
func (df DataFrame) SelectMatching(patterns ...string) DataFrame {
	if df.err != nil {
		return df
	}

	regexes := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			df.err = fmt.Errorf("DataFrame.SelectMatching: invalid pattern \"%s\"", pattern)
			return df
		}
		regexes[i] = regex
	}

	selected := make(map[string]bool)
	for _, name := range df.names {
		selected[name] = false
	}

	names := make([]string, 0)
	seriesList := make([]series.Series, 0)

	for _, regex := range regexes {
		for _, name := range df.names {
			if !selected[name] && regex.MatchString(name) {
				selected[name] = true
				names = append(names, name)
				seriesList = append(seriesList, df.Col(name))
			}
		}
	}

	return DataFrame{
		names:  names,
		series: seriesList,
		ctx:    df.ctx,
	}
}

func (df DataFrame) SelectAt(indices ...int) DataFrame {
	if df.err != nil {
		return df
	}

	selected := NewDataFrame(df.ctx)
	for _, index := range indices {
		if index < 0 || index >= len(df.series) {
			selected.AddSeries(df.names[index], df.series[index])
		} else {
			return DataFrame{err: fmt.Errorf("DataFrame.SelectAt: index %d out of bounds", index)}
		}
	}

	return selected
}

func (df DataFrame) Filter(mask any) DataFrame {
	if df.err != nil {
		return df
	}

	var maskSeries series.Bools
	if _, ok := mask.(series.Bools); ok {
		maskSeries = mask.(series.Bools)

	} else {
		if mask, ok := mask.([]bool); ok {
			maskSeries = series.NewSeriesBool(mask, nil, false, df.ctx)
		} else {
			df.err = fmt.Errorf("DataFrame.Filter: mask is not a bool series")
			return df
		}
	}

	if maskSeries.Len() != df.NRows() {
		df.err = fmt.Errorf("DataFrame.Filter: mask length (%d) does not match dataframe length (%d)", maskSeries.Len(), df.NRows())
		return df
	}

	seriesList := make([]series.Series, 0)
	for _, series := range df.series {
		seriesList = append(seriesList, series.Filter(maskSeries))
	}

	return DataFrame{
		names:  df.names,
		series: seriesList,
		ctx:    df.ctx,
	}
}

func (df DataFrame) GroupBy(by ...string) DataFrame {
	if df.err != nil {
		return df
	}

	// Grouping an already grouped dataframe replaces the existing grouping.
	if df.isGrouped {
		df = df.Ungroup()
	}

	// Check that all the group by columns exist
	for _, name := range by {
		found := false
		for _, name_ := range df.names {
			if name_ == name {
				found = true
				break
			}
		}

		if !found {
			df.err = fmt.Errorf("DataFrame.GroupBy: column \"%s\" not found", name)
			return df
		}
	}

	// Lazy: just record the grouping. No partitions are built here; they are
	// materialized on demand (see buildPartitions) by the few callers that
	// still need them (Join).
	df.isGrouped = true
	df.groupByNames = by
	df.partitions = nil

	return df
}

// buildPartitions materializes the partition chain for the recorded
// group-by columns (df.groupByNames). DataFrame is passed by value, so
// there is nowhere durable to cache the result across calls; it is rebuilt
// fresh every time. Only the legacy Join path calls this — the new Agg
// engine (Task 8) reads the raw series directly and never needs it.
func (df DataFrame) buildPartitions() []DataFramePartitionEntry {
	partitions := make([]DataFramePartitionEntry, len(df.groupByNames))

	for partitionsIndex, name := range df.groupByNames {
		i := df.ColIndex(name)
		s := df.series[i]

		// First partition: group the series
		if partitionsIndex == 0 {
			partitions[partitionsIndex] = DataFramePartitionEntry{
				index:     i,
				name:      name,
				partition: s.Group().Partition(),
			}
		} else

		// Subsequent partitions: sub-group the series
		{
			partitions[partitionsIndex] = DataFramePartitionEntry{
				index:     i,
				name:      name,
				partition: s.GroupBy(partitions[partitionsIndex-1].partition).Partition(),
			}
		}
	}

	return partitions
}

func (df DataFrame) Ungroup() DataFrame {
	if df.err != nil {
		return df
	}

	df.isGrouped = false
	df.groupByNames = nil
	df.partitions = nil
	return df
}

func (df DataFrame) getPartitions() []series.SeriesPartition {
	if df.err != nil {
		return nil
	}

	if df.isGrouped {
		built := df.buildPartitions()
		partitions := make([]series.SeriesPartition, len(built))
		for i, partition := range built {
			partitions[i] = partition.partition
		}
		return partitions
	} else {
		return nil
	}
}

// Join joins the two frames on the given columns, or on every same-named
// column when none are given. Null keys match null keys.
//
// The output row order is part of the contract. Rows follow the left
// frame's row order, and a row with several matches produces consecutive
// output rows, ordered by the right frame's rows. A right join follows
// the right frame's row order instead. In an outer join the unmatched
// right rows come last, in the right frame's row order.
func (df DataFrame) Join(how JoinType, other DataFrame, on ...string) DataFrame {
	if df.err != nil {
		return df
	}

	// CASE: the dataframes have different contexts
	if df.ctx != other.Context() {
		df.err = fmt.Errorf("DataFrame.Join: dataframes have different contexts")
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.Join: cannot join a grouped dataframe")
		return df
	}

	if other.IsGrouped() {
		df.err = fmt.Errorf("DataFrame.Join: cannot join with a grouped dataframe")
		return df
	}

	// CHECK: all the join columns must exist
	// CHECK: all the join columns must have the same type
	types := make([]meta.BaseType, len(on))
	for _, name := range on {

		// Series A
		found := false
		for idx, series := range df.series {
			if df.names[idx] == name {
				found = true

				// keep track of the types
				types = append(types, series.Type())
				break
			}
		}
		if !found {
			df.err = fmt.Errorf("DataFrame.Join: column \"%s\" not found in left dataframe", name)
			return df
		}

		// Series B
		found = false
		otherNames := other.Names()
		for idx, series := range other.series {
			if idx < len(otherNames) && otherNames[idx] == name {
				found = true

				// CHECK: the types must match
				if types[len(types)-1] != series.Type() {
					df.err = fmt.Errorf("DataFrame.Join: columns \"%s\" have different types", name)
					return df
				}
				break
			}
		}
		if !found {
			df.err = fmt.Errorf("DataFrame.Join: column \"%s\" not found in right dataframe", name)
			return df
		}
	}

	// CASE: on is empty -> use all columns with the same name
	if len(on) == 0 {
		for _, name := range df.Names() {
			if other.ColIndex(name) != -1 {
				on = append(on, name)
			}
		}
	}

	// CASE: on is still empty -> error
	if len(on) == 0 {
		df.err = fmt.Errorf("DataFrame.Join: no columns to join on")
		return df
	}

	// CHECK: all columns in on must have the same type
	for _, name := range on {
		if df.Col(name).Type() != other.Col(name).Type() {
			df.err = fmt.Errorf("DataFrame.Join: columns \"%s\" have different types", name)
			return df
		}
	}

	// Group the dataframes by the join columns
	dfGrouped := df.GroupBy(on...)
	otherGrouped := other.GroupBy(on...)

	colsDiffA := make([]string, 0)
	colsDiffB := make([]string, 0)

	// Get the columns that are not in the join columns
	for _, name := range df.Names() {
		found := false
		for _, joinName := range on {
			if name == joinName {
				found = true
				break
			}
		}
		if !found {
			colsDiffA = append(colsDiffA, name)
		}
	}

	for _, name := range other.Names() {
		found := false
		for _, joinName := range on {
			if name == joinName {
				found = true
				break
			}
		}
		if !found {
			colsDiffB = append(colsDiffB, name)
		}
	}

	// Get the columns that are in both dataframes
	commonCols := make(map[string]bool)
	for _, name := range df.Names() {
		for _, otherName := range other.Names() {
			if name == otherName {
				commonCols[name] = true
				break
			}
		}
	}

	joined := NewDataFrame(df.ctx)

	pA := dfGrouped.getPartitions()
	pB := otherGrouped.getPartitions()

	// Get the maps, keys and sort them
	mapA := pA[len(pA)-1].GetMap()
	mapB := pB[len(pB)-1].GetMap()

	keysA := make([]int64, 0, len(mapA))
	keysB := make([]int64, 0, len(mapB))

	for key := range mapA {
		keysA = append(keysA, key)
	}

	for key := range mapB {
		keysB = append(keysB, key)
	}

	sort.Slice(keysA, func(i, j int) bool { return keysA[i] < keysA[j] })
	sort.Slice(keysB, func(i, j int) bool { return keysB[i] < keysB[j] })

	// Find the intersection
	keysAOnly := make([]int64, 0, len(keysA))
	keysBOnly := make([]int64, 0, len(keysB))
	keysIntersection := make([]int64, 0, len(keysA))

	var i, j int = 0, 0
	for i < len(keysA) && j < len(keysB) {
		if keysA[i] < keysB[j] {
			keysAOnly = append(keysAOnly, keysA[i])
			i++
		} else if keysA[i] > keysB[j] {
			keysBOnly = append(keysBOnly, keysB[j])
			j++
		} else {
			keysIntersection = append(keysIntersection, keysA[i])
			i++
			j++
		}
	}

	for i < len(keysA) {
		keysAOnly = append(keysAOnly, keysA[i])
		i++
	}

	for j < len(keysB) {
		keysBOnly = append(keysBOnly, keysB[j])
		j++
	}

	// Materialize the matching row pairs. The grouping code assigns group
	// ids in no useful order, so every output segment is sorted by row
	// index: the promised row order does not depend on grouping internals.
	type pair struct{ a, b int }

	matched := make([]pair, 0)
	for _, key := range keysIntersection {
		for _, indexA := range mapA[key] {
			for _, indexB := range mapB[key] {
				matched = append(matched, pair{indexA, indexB})
			}
		}
	}

	aOnly := make([]int, 0)
	for _, key := range keysAOnly {
		aOnly = append(aOnly, mapA[key]...)
	}
	sort.Ints(aOnly)

	bOnly := make([]int, 0)
	for _, key := range keysBOnly {
		bOnly = append(bOnly, mapB[key]...)
	}
	sort.Ints(bOnly)

	byLeftRow := func(p []pair) {
		sort.Slice(p, func(i, j int) bool {
			if p[i].a != p[j].a {
				return p[i].a < p[j].a
			}
			return p[i].b < p[j].b
		})
	}

	// indexB is -1 where a left row has no match; indexA is -1 where a
	// right row has none.
	var pairs []pair
	switch how {
	case JoinInner:
		pairs = matched
		byLeftRow(pairs)

	case JoinLeft:
		pairs = matched
		for _, a := range aOnly {
			pairs = append(pairs, pair{a, -1})
		}
		byLeftRow(pairs)

	case JoinRight:
		pairs = matched
		for _, b := range bOnly {
			pairs = append(pairs, pair{-1, b})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].b != pairs[j].b {
				return pairs[i].b < pairs[j].b
			}
			return pairs[i].a < pairs[j].a
		})

	case JoinOuter:
		pairs = matched
		for _, a := range aOnly {
			pairs = append(pairs, pair{a, -1})
		}
		byLeftRow(pairs)
		for _, b := range bOnly {
			pairs = append(pairs, pair{-1, b})
		}
	}

	indicesA := make([]int, len(pairs))
	indicesB := make([]int, len(pairs))
	for i, p := range pairs {
		indicesA[i] = p.a
		indicesB[i] = p.b
	}

	// Join columns: the key values come from the side that has every row.
	// An outer join has no such side, so the left-ordered part comes from
	// the left frame and the unmatched right rows at the tail from the
	// right frame.
	for _, name := range on {
		var keyCol series.Series
		switch how {
		case JoinRight:
			keyCol = other.Col(name).FilterIntSlice(indicesB, false)
		case JoinOuter:
			cut := len(pairs) - len(bOnly)
			keyCol = df.Col(name).FilterIntSlice(indicesA[:cut], false).
				Append(other.Col(name).FilterIntSlice(indicesB[cut:], false))
		default:
			keyCol = df.Col(name).FilterIntSlice(indicesA, false)
		}
		joined = joined.AddSeries(name, keyCol)
	}

	// The remaining left columns, null where the row exists only in the
	// right frame.
	for _, name := range colsDiffA {
		ser_ := joinGather(df.Col(name), indicesA, df.ctx)
		if commonCols[name] {
			name += "_x"
		}
		joined = joined.AddSeries(name, ser_)
	}

	// The remaining right columns, null where the row exists only in the
	// left frame.
	for _, name := range colsDiffB {
		ser_ := joinGather(other.Col(name), indicesB, df.ctx)
		if commonCols[name] {
			name += "_y"
		}
		joined = joined.AddSeries(name, ser_)
	}

	return joined
}

// allNullSeries builds a series of the given type whose elements are all
// null.
func allNullSeries(t meta.BaseType, size int, ctx *enchanter.Context) series.Series {
	mask := make([]bool, size)
	for i := range mask {
		mask[i] = true
	}
	switch t {
	case meta.BoolType:
		return series.NewSeriesBool(make([]bool, size), mask, false, ctx)
	case meta.IntType:
		return series.NewSeriesInt(make([]int, size), mask, false, ctx)
	case meta.Int64Type:
		return series.NewSeriesInt64(make([]int64, size), mask, false, ctx)
	case meta.Float64Type:
		return series.NewSeriesFloat64(make([]float64, size), mask, false, ctx)
	case meta.StringType:
		return series.NewSeriesString(make([]string, size), mask, false, ctx)
	case meta.TimeType:
		return series.NewSeriesTime(make([]time.Time, size), mask, false, ctx)
	case meta.DurationType:
		return series.NewSeriesDuration(make([]time.Duration, size), mask, false, ctx)
	}
	return series.Errors{Msg_: fmt.Sprintf("allNullSeries: unsupported type %v", t)}
}

// joinGather returns the elements of s at the given indices, in order,
// where index -1 produces a null element. One null element is put in
// front of a fresh copy of the series, so index v gathers as v+1 and -1
// lands on the null.
func joinGather(s series.Series, indices []int, ctx *enchanter.Context) series.Series {
	hasNull := false
	for _, v := range indices {
		if v == -1 {
			hasNull = true
			break
		}
	}
	if !hasNull {
		return s.FilterIntSlice(indices, false)
	}

	ext := allNullSeries(s.Type(), 1, ctx).Append(s)
	safe := make([]int, len(indices))
	for i, v := range indices {
		safe[i] = v + 1
	}
	return ext.FilterIntSlice(safe, false)
}

// Slice returns the rows in the half-open interval [start, end).
func (df DataFrame) Slice(start, end int) DataFrame {
	if df.err != nil {
		return df
	}

	if start < 0 || end < start || end > df.NRows() {
		df.err = fmt.Errorf("DataFrame.Slice: invalid interval [%d, %d) for a frame of %d rows", start, end, df.NRows())
		return df
	}

	indices := make([]int, end-start)
	for i := range indices {
		indices[i] = start + i
	}
	return df.takeIndices(indices)
}

// TakeIndices returns the rows at the given indices, in the given
// order. An index may repeat.
func (df DataFrame) TakeIndices(indices []int) DataFrame {
	if df.err != nil {
		return df
	}

	for _, v := range indices {
		if v < 0 || v >= df.NRows() {
			df.err = fmt.Errorf("DataFrame.TakeIndices: index %d is out of range", v)
			return df
		}
	}
	return df.takeIndices(indices)
}

func (df DataFrame) takeIndices(indices []int) DataFrame {
	taken := NewDataFrame(df.ctx)
	for idx, series := range df.series {
		taken = taken.AddSeries(df.names[idx], series.FilterIntSlice(indices, false))
	}
	return taken
}

func (df DataFrame) Len() int {
	if df.err != nil || len(df.series) < 1 {
		return 0
	}

	return df.series[0].Len()
}

// seriesSorter is the sorting surface the dataframe needs from a series.
// Every concrete series type implements these methods; they are not part
// of the public Series interface.
type seriesSorter interface {
	Less(i, j int) bool
	Equal(i, j int) bool
	Swap(i, j int)
}

func (df DataFrame) Less(i, j int) bool {
	for _, param := range df.sortParams {
		s := param._series.(seriesSorter)
		if !s.Equal(i, j) {
			return (param.asc && s.Less(i, j)) || (!param.asc && s.Less(j, i))
		}
	}

	return false
}

func (df DataFrame) Swap(i, j int) {
	for _, series := range df.series {
		series.(seriesSorter).Swap(i, j)
	}
}

func (df DataFrame) OrderBy(params ...SortParam) DataFrame {
	if df.err != nil {
		return df
	}

	if df.isGrouped {
		df.err = fmt.Errorf("DataFrame.OrderBy: cannot order grouped DataFrame")
		return df
	}

	// CHECK: params must have unique names and names must be valid
	paramNames := make(map[string]bool)
	for i, param := range params {
		if paramNames[param.name] {
			df.err = fmt.Errorf("DataFrame.OrderBy: series names must be unique")
			return df
		}
		paramNames[param.name] = true

		if series := df.seriesByName(param.name); series != nil {
			params[i]._series = series
		} else {
			df.err = fmt.Errorf("DataFrame.OrderBy: series \"%s\" not found", param.name)
			return df
		}
	}

	df.sortParams = params
	sort.Sort(df)
	df.sortParams = nil

	return df
}

////////////////////////			SUMMARY

func (df DataFrame) Agg(aggregators ...aggregator) aggregatorBuilder {
	return aggregatorBuilder{df, true, aggregators}
}

// buildGroupKeyCols returns the group-by columns, in the order recorded by
// GroupBy (df.groupByNames), for the aggregation engine (agg_engine.go) to
// key its groups on.
func (df DataFrame) buildGroupKeyCols() []series.Series {
	keyCols := make([]series.Series, len(df.groupByNames))
	for i, name := range df.groupByNames {
		keyCols[i] = df.series[df.ColIndex(name)]
	}
	return keyCols
}

////////////////////////			PRINTING

func (df DataFrame) Describe() string {
	return ""
}

func (df DataFrame) Records(header bool) [][]string {
	var out [][]string
	if header {
		out = make([][]string, df.NRows()+1)
	} else {
		out = make([][]string, df.NRows())
	}

	h := 0
	if header {
		out[0] = make([]string, df.NCols())
		for j := 0; j < df.NCols(); j++ {
			out[0][j] = df.names[j]
		}

		h = 1
	}

	for i := 0 + h; i < df.NRows()+h; i++ {
		out[i] = make([]string, df.NCols())
		for j := 0; j < df.NCols(); j++ {
			out[i][j] = df.series[j].GetAsString(i - h)
		}
	}

	return out
}

// Pretty print the dataframe.
func (df DataFrame) PPrint(params PPrintParams) DataFrame {
	if df.err != nil {
		fmt.Println(df.err)
		return df
	}

	buffer := ""

	// check if the dataframe is empty
	if df.NRows() == 0 {
		buffer += params.indent
		if params.useLipGloss {
			params.styleNames.Render("  Empty DataFrame\n")
		} else {
			buffer += "  Empty DataFrame\n"
		}
		fmt.Println(buffer)
		return df
	}

	// print the shape
	buffer += params.indent
	if params.useLipGloss {
		buffer += params.styleTypes.Render(fmt.Sprintf("  DataFrame: %d rows, %d columns", df.NRows(), df.NCols()))
	} else {
		buffer += fmt.Sprintf("  DataFrame: %d rows, %d columns", df.NRows(), df.NCols())
	}
	buffer += "\n"

	// print the group by columns
	if df.isGrouped {
		buffer += params.indent
		if params.useLipGloss {
			buffer += params.styleTypes.Render("  Grouped by: ")
		} else {
			buffer += "  Grouped by: "
		}
		for i, name := range df.groupByNames {
			if params.useLipGloss {
				buffer += params.styleTypes.Render(fmt.Sprintf("%s", name))
			} else {
				buffer += fmt.Sprintf("%s", name)
			}
			if i < len(df.groupByNames)-1 {
				if params.useLipGloss {
					buffer += params.styleTypes.Render(",")
				} else {
					buffer += ","
				}
			}
		}
		buffer += "\n"
	}

	// check how many variables can fit in the screen
	nColsOut := 0
	actualWidthsSum := 0

	widths := make([]int, df.NCols())
	for i, name := range df.names {
		widths[i] = max(len(df.series[i].Type().String()), len(name))
		actualWidthsSum += widths[i] + 3
		if actualWidthsSum > params.width {
			break
		}
		nColsOut++
	}
	widths = widths[:nColsOut]

	nRowsOut := min(10, df.NRows())
	if params.nrows > 0 {
		nRowsOut = min(params.nrows, df.NRows())
	}

	addTail := false
	if df.NRows() > nRowsOut+params.tailLen {
		addTail = true
		nRowsOut -= params.tailLen
	}

	formatters := make([]formatter.Formatter, nColsOut)
	for i := 0; i < nColsOut; i++ {
		switch df.series[i].Type() {
		case meta.BoolType, meta.StringType, meta.TimeType:
			formatters[i] = formatter.NewStringFormatter().
				SetUseLipGloss(params.useLipGloss)
		case meta.IntType, meta.Int64Type, meta.Float64Type, meta.DurationType:
			formatters[i] = formatter.NewNumericFormatter().
				SetUseLipGloss(params.useLipGloss).
				SetNaText(df.ctx.GetNaText()).
				SetTruncateOutput(true)
		}

		switch s := df.series[i].(type) {
		case series.Bools:
			for _, v := range s.DataAsString()[:nRowsOut] {
				formatters[i].Push(v)
			}

			if addTail {
				for _, v := range s.DataAsString()[df.NRows()-params.tailLen:] {
					formatters[i].Push(v)
				}
			}

		case series.Ints:
			for _, v := range s.Ints()[:nRowsOut] {
				formatters[i].Push(v)
			}

			if addTail {
				for _, v := range s.Ints()[df.NRows()-params.tailLen:] {
					formatters[i].Push(v)
				}
			}

		case series.Int64s:
			for _, v := range s.Int64s()[:nRowsOut] {
				formatters[i].Push(v)
			}

			if addTail {
				for _, v := range s.Int64s()[df.NRows()-params.tailLen:] {
					formatters[i].Push(v)
				}
			}

		case series.Float64s:
			for _, v := range s.Float64s()[:nRowsOut] {
				formatters[i].Push(v)
			}

			if addTail {
				for _, v := range s.Float64s()[df.NRows()-params.tailLen:] {
					formatters[i].Push(v)
				}
			}

		case series.Strings:
			for _, v := range s.Strings()[:nRowsOut] {
				formatters[i].Push(v)
			}

			if addTail {
				for _, v := range s.Strings()[df.NRows()-params.tailLen:] {
					formatters[i].Push(v)
				}
			}

		case series.Times:
			for _, v := range s.DataAsString()[:nRowsOut] {
				formatters[i].Push(v)
			}

			if addTail {
				for _, v := range s.DataAsString()[df.NRows()-params.tailLen:] {
					formatters[i].Push(v)
				}
			}

		case series.Durations:
			for _, v := range s.Data().([]time.Duration)[:nRowsOut] {
				formatters[i].Push(v)
			}

			if addTail {
				for _, v := range s.Data().([]time.Duration)[df.NRows()-params.tailLen:] {
					formatters[i].Push(v)
				}
			}
		}
	}

	// compute the optimal width for each column with the given formatter
	// get the new number of columns that can fit in the screen
	actualWidthsSum = 0
	nColsOut = 0
	for i, f := range formatters {
		f.Compute()
		widths[i] = max(f.GetMaxWidth(), widths[i])
		actualWidthsSum += widths[i] + 3
		if actualWidthsSum > params.width {
			break
		}
		nColsOut++
	}
	widths = widths[:nColsOut]
	formatters = formatters[:nColsOut]

	// header
	buffer += params.indent + "╭"
	for i, w := range widths {
		for j := 0; j < w+2; j++ {
			buffer += "─"
		}
		if i < nColsOut-1 {
			buffer += "┬"
		}
	}
	buffer += "╮\n"

	// column names
	buffer += params.indent + "│"
	if params.useLipGloss {
		for i, name := range df.names[:nColsOut] {
			buffer += params.styleNames.Render(fmt.Sprintf(" %-*s ", widths[i], utils.Truncate(name, widths[i]))) + "│"
		}
	} else {
		for i, name := range df.names[:nColsOut] {
			buffer += fmt.Sprintf(" %-*s ", widths[i], utils.Truncate(name, widths[i])) + "│"
		}
	}
	buffer += "\n"

	// separator
	buffer += params.indent + "├"
	for i, w := range widths {
		for j := 0; j < w+2; j++ {
			buffer += "─"
		}
		if i < nColsOut-1 {
			buffer += "┼"
		}
	}
	buffer += "┤\n"

	// column types
	buffer += params.indent + "│"
	if params.useLipGloss {
		for i, c := range df.series[:nColsOut] {
			buffer += params.styleTypes.Render(fmt.Sprintf(" %-*s ", widths[i], c.Type().String())) + "│"
		}
	} else {
		for i, c := range df.series[:nColsOut] {
			buffer += fmt.Sprintf(" %-*s ", widths[i], c.Type().String()) + "│"
		}
	}
	buffer += "\n"

	// separator
	buffer += params.indent + "├"
	for i, w := range widths {
		for j := 0; j < w+2; j++ {
			buffer += "─"
		}
		if i < nColsOut-1 {
			buffer += "┼"
		}
	}
	buffer += "┤\n"

	// data
	for i := 0; i < nRowsOut; i++ {
		buffer += params.indent + "│"
		for j, c := range df.series[:nColsOut] {
			switch s := c.(type) {
			case series.Bools:
				buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.GetAsString(i), s.IsNull(i))) + "│"
			case series.Ints:
				buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
			case series.Int64s:
				buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
			case series.Float64s:
				buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
			case series.Strings:
				buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
			case series.Times:
				buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.GetAsString(i), s.IsNull(i))) + "│"
			case series.Durations:
				buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
			}
		}
		buffer += "\n"
	}

	if addTail {
		// separator (bottom)
		buffer += params.indent + "┊"
		for j := range df.series[:nColsOut] {
			buffer += utils.Center("⋮", widths[j]+2) + "┊"
		}
		buffer += "\n"

		// tail
		for i := df.NRows() - params.tailLen; i < df.NRows(); i++ {
			buffer += params.indent + "│"
			for j, c := range df.series[:nColsOut] {
				switch s := c.(type) {
				case series.Bools:
					buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.GetAsString(i), s.IsNull(i))) + "│"
				case series.Ints:
					buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
				case series.Int64s:
					buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
				case series.Float64s:
					buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
				case series.Strings:
					buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
				case series.Times:
					buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.GetAsString(i), s.IsNull(i))) + "│"
				case series.Durations:
					buffer += fmt.Sprintf(" %s ", formatters[j].Format(widths[j], s.Get(i), s.IsNull(i))) + "│"
				}
			}
			buffer += "\n"
		}
	}

	// end
	buffer += params.indent + "╰"
	for i, w := range widths {
		for j := 0; j < w+2; j++ {
			buffer += "─"
		}
		if i < nColsOut-1 {
			buffer += "┴"
		}
	}
	buffer += "╯\n"

	// Non-displayed column names
	if df.NCols() > nColsOut {
		for i := nColsOut; i < df.NCols(); i++ {
			buffer += df.names[i] + ", "
		}
	}

	fmt.Println(buffer)

	return df
}
