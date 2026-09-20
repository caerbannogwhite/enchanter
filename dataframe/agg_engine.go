package dataframe

import (
	"math"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/caerbannogwhite/enchanter/series"
)

// aggMinParallel is the row-count threshold at or above which aggregate
// spawns worker goroutines instead of running the single-chunk core inline.
// Below it, the fixed cost of partitioning rows and merging worker state
// outweighs the benefit of parallelism.
const aggMinParallel = 1 << 16

// aggregateSerial runs the fused single-pass aggregation over df in a single
// chunk covering every row. It shares its accumulation core (accumulateChunk)
// and result-building core (finalizeAggregate) with aggregate's parallel
// workers; see aggregate for the parallel entry point and merge semantics.
//
// It builds one group per distinct composite key of keyCols (via groupTable),
// holding one reducibleState slot (reducible) or collector (holistic) per
// aggregator per group. A single pass over the rows updates the per-group state; each
// value column is read inline through an aggValueView (see
// prepAggValueColumns) rather than pre-copied to []float64, and a value is
// missing iff the column's null mask says so, or (for a Float64 column) the
// raw value is NaN — see accumulateChunk.
//
//   - removeNAs == true  (the new default): null values are skipped for every
//     aggregate; no NA is propagated.
//   - removeNAs == false: a null value propagates through that (group, aggregator) so its
//     result becomes NA — for every aggregate alike, reducible (Mean, ...) and
//     holistic (Median, Quantile). Count is the sole exception: it counts rows
//     and is independent of value nulls. Collectors still never receive NaN; the
//     NA-propagation flag overrides the collector's result at finalize.
//
// The result is the key columns (one value per group, taken from each group's
// representative row) followed by one aggregate column per aggregator, with the
// rows sorted by key ascending, nulls last, lexicographically across keyCols.
// Count is emitted as Int64s; every other aggregate is emitted as Float64s with
// a null mask from the finalize isNull flags.
//
// len(keyCols) == 0 is handled as a single group with no key columns.
func aggregateSerial(df DataFrame, keyCols []series.Series, aggs []aggregator, removeNAs bool) DataFrame {
	nRows := df.NRows()
	views, isHol := prepAggValueColumns(df, aggs)

	gt, states, cols, propagated := accumulateChunk(keyCols, aggs, views, isHol, removeNAs, 0, nRows)

	return finalizeAggregate(df, keyCols, aggs, isHol, removeNAs, gt, states, cols, propagated)
}

// aggregate is the public single-pass aggregation entry point. Below
// aggMinParallel rows it runs the single-chunk core inline (identical output
// to aggregateSerial). At or above the threshold, it partitions the rows into
// runtime.GOMAXPROCS(0) contiguous chunks, accumulates each chunk in its own
// goroutine into a chunk-local groupTable and accumulator/collector state
// (accumulateChunk over disjoint row ranges — no shared mutable state between
// workers), then merges every chunk's state into one global groupTable and
// finalizes exactly as the serial path.
//
// Merge correctness relies on two points:
//
//   - Chunk-local group ids are NOT comparable across chunks: groupTable
//     assigns dense ids in first-appearance order, which differs per chunk, so
//     the same logical group can get different ids in different chunks. Each
//     chunk-local group is therefore re-keyed into the global groupTable by
//     replaying its representative row through global.idOf, which re-reads
//     that row's actual key-column values, so equal groups collapse to the
//     same global id regardless of which chunk (or which local id) they came
//     from.
//
//   - Under removeNAs == false, NA-propagation flags are OR-merged per (aggregator,
//     group): a group NA-propagated in any one chunk is NA-propagated in the merged
//     result. This matches the serial engine's per-row NA propagation exactly,
//     since an NA-propagated group there stays NA-propagated once any row in it is null.
func aggregate(df DataFrame, keyCols []series.Series, aggs []aggregator, removeNAs bool) DataFrame {
	nRows := df.NRows()
	views, isHol := prepAggValueColumns(df, aggs)

	if nRows < aggMinParallel {
		gt, states, cols, propagated := accumulateChunk(keyCols, aggs, views, isHol, removeNAs, 0, nRows)
		return finalizeAggregate(df, keyCols, aggs, isHol, removeNAs, gt, states, cols, propagated)
	}

	nWorkers := runtime.GOMAXPROCS(0)
	if nWorkers < 1 {
		nWorkers = 1
	}
	if nWorkers > nRows {
		nWorkers = nRows
	}
	chunkSize := (nRows + nWorkers - 1) / nWorkers

	type chunkState struct {
		gt         *groupTable
		states     []*reducibleState
		cols       [][]collector
		propagated [][]bool
	}

	// Each goroutine writes only its own index w; disjoint slice elements are
	// safe to write concurrently. wg.Wait() below establishes the
	// happens-before edge before the main goroutine reads any of them.
	chunks := make([]chunkState, nWorkers)
	var wg sync.WaitGroup
	for w := 0; w < nWorkers; w++ {
		lo := w * chunkSize
		hi := lo + chunkSize
		if hi > nRows {
			hi = nRows
		}
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(w, lo, hi int) {
			defer wg.Done()
			gt, states, cols, propagated := accumulateChunk(keyCols, aggs, views, isHol, removeNAs, lo, hi)
			chunks[w] = chunkState{gt, states, cols, propagated}
		}(w, lo, hi)
	}
	wg.Wait()

	// Merge every chunk's local state into one global groupTable, re-keying
	// each chunk-local group by its representative row (see doc comment
	// above) and OR-merging NA-propagation flags. Chunks are merged in worker order
	// for determinism; the merge itself is order-independent — reducible
	// (mergeReducible) and collector merge are commutative, and a group's
	// representative row always carries the same key values no matter which
	// chunk supplied it.
	global := newGroupTable(keyCols)
	gStates := make([]*reducibleState, len(aggs))
	gCols := make([][]collector, len(aggs))
	gPropagated := make([][]bool, len(aggs))
	for j := range aggs {
		if !isHol[j] {
			gStates[j] = &reducibleState{}
		}
	}

	nGroups := 0
	for _, chunk := range chunks {
		if chunk.gt == nil { // empty chunk (lo >= hi above)
			continue
		}
		reps := chunk.gt.representativeRows()
		for localGid, row := range reps {
			gid := global.idOf(row)

			if gid == nGroups {
				for j, agg := range aggs {
					if isHol[j] {
						gCols[j] = append(gCols[j], newQuantileCollector(agg.p, agg.interp))
					} else {
						growReducible(gStates[j], agg.type_)
					}
					gPropagated[j] = append(gPropagated[j], false)
				}
				nGroups++
			}

			for j, agg := range aggs {
				if isHol[j] {
					gCols[j][gid].merge(chunk.cols[j][localGid])
				} else {
					mergeReducible(gStates[j], gid, chunk.states[j], localGid, agg.type_)
				}
				if chunk.propagated[j][localGid] {
					gPropagated[j][gid] = true
				}
			}
		}
	}

	return finalizeAggregate(df, keyCols, aggs, isHol, removeNAs, global, gStates, gCols, gPropagated)
}

// aggValueKind identifies the element type backing an aggValueView, so
// accumulateChunk's row loop can read+convert a value with a plain switch
// instead of a per-row interface call.
type aggValueKind uint8

const (
	// aggValUnsupported marks a value column whose type none of the other
	// kinds cover (e.g. Strings). Reading such a view mirrors the previous
	// behavior of the former stats-preprocessing helper, which returned nil for unsupported
	// types, which made accumulateChunk panic with an out-of-range index on
	// the first row read; see accumulateChunk's default case.
	aggValUnsupported aggValueKind = iota
	aggValF64
	aggValI64
	aggValInt
	aggValBool
	aggValDur
)

// aggValueView is a read-only, typed view over one aggregator's value
// column, built once outside accumulateChunk's row loop (see
// prepAggValueColumns) so the loop can read each row's value without any
// per-row allocation or interface dispatch. nullable/nullMask mirror the
// column's own bit-packed null representation (the same convention as, e.g.,
// series.Int64s.IsNull: bit-packed, bit-set == null, 8 rows per byte), read
// directly here instead of through the Series interface.
type aggValueView struct {
	kind     aggValueKind
	nullable bool
	nullMask []uint8
	f64      []float64
	i64      []int64
	ints     []int
	bools    []bool
	dur      []time.Duration
}

// newAggValueView builds the typed view for col. Unsupported column types
// (anything but Float64s/Int64s/Ints/Bools/Durations — the same set
// the former stats-preprocessing helper supported) yield aggValUnsupported.
func newAggValueView(col series.Series) aggValueView {
	switch c := col.(type) {
	case series.Float64s:
		return aggValueView{kind: aggValF64, nullable: c.IsNullable_, nullMask: c.NullMask_, f64: c.Data_}
	case series.Int64s:
		return aggValueView{kind: aggValI64, nullable: c.IsNullable_, nullMask: c.NullMask_, i64: c.Data_}
	case series.Ints:
		return aggValueView{kind: aggValInt, nullable: c.IsNullable_, nullMask: c.NullMask_, ints: c.Data_}
	case series.Bools:
		return aggValueView{kind: aggValBool, nullable: c.IsNullable_, nullMask: c.NullMask_, bools: c.Data_}
	case series.Durations:
		return aggValueView{kind: aggValDur, nullable: c.IsNullable_, nullMask: c.NullMask_, dur: c.Data_}
	default:
		return aggValueView{kind: aggValUnsupported}
	}
}

// prepAggValueColumns builds a typed aggValueView per aggregator once, up
// front (Count needs no value column). The returned slice is read-only from
// this point on and safe to share across accumulateChunk calls running
// concurrently over disjoint row ranges — no per-op []float64 copy is made.
func prepAggValueColumns(df DataFrame, aggs []aggregator) (views []aggValueView, isHol []bool) {
	views = make([]aggValueView, len(aggs))
	isHol = make([]bool, len(aggs))
	for j, agg := range aggs {
		isHol[j] = agg.type_.isHolistic()
		if agg.type_ == AGGREGATE_COUNT {
			continue
		}
		views[j] = newAggValueView(df.Col(agg.name))
	}
	return views, isHol
}

// reducibleState holds per-group struct-of-arrays state for one reducible
// aggregator (Count/Sum/Mean/Min/Max/Std/Variance/Any/All), indexed by group
// id. It replaces the per-group accumulator OBJECTS behind the accumulator
// interface (accumulator.go) with typed arrays, so accumulateChunk's row
// loop can update them with a `switch kind` instead of a per-row virtual
// call. Only the fields relevant to a given aggregator's AggregateType are
// grown/read (see growReducible); the others stay nil for that aggregator.
//
// growReducible/updateReducible/mergeReducible/finalizeReducible below
// reproduce, verbatim in behavior, countAcc/sumAcc/meanAcc/minAcc/maxAcc/
// varAcc's update/merge/result from accumulator.go — that file remains the
// documented reference and the 0.5.0 custom-aggregator extension point; this
// is the fast lane the built-in engine now uses instead of calling it. Any/All
// have no accumulator.go equivalent — the engine's SoA path is their only
// implementation (see the plan's Task 11 note on the obsolete accumulator
// path).
type reducibleState struct {
	n     []int64   // COUNT, SUM, MEAN, ANY, ALL: row/value count per group
	sum   []float64 // SUM, MEAN: running sum per group
	val   []float64 // MIN, MAX: running extreme per group
	set   []bool    // MIN, MAX: whether val has been set for that group yet
	wn    []float64 // STD, VARIANCE: Welford sample count per group
	mean  []float64 // STD, VARIANCE: Welford running mean per group
	m2    []float64 // STD, VARIANCE: Welford running sum of squared deviations
	bflag []bool    // ANY: any truthy value seen; ALL: any falsy value seen (per group)
}

// growReducible appends one zero-value slot (a new group) to exactly the
// fields t's kind uses, initializing that group's state to its zero value
// (0 count/sum, unset min/max, zeroed Welford accumulators) exactly as a
// fresh built-in accumulator would start out.
func growReducible(st *reducibleState, t AggregateType) {
	switch t {
	case AGGREGATE_COUNT:
		st.n = append(st.n, 0)
	case AGGREGATE_SUM, AGGREGATE_MEAN:
		st.sum = append(st.sum, 0)
		st.n = append(st.n, 0)
	case AGGREGATE_MIN, AGGREGATE_MAX:
		st.val = append(st.val, 0)
		st.set = append(st.set, false)
	case AGGREGATE_STD, AGGREGATE_VARIANCE:
		st.wn = append(st.wn, 0)
		st.mean = append(st.mean, 0)
		st.m2 = append(st.m2, 0)
	case AGGREGATE_ANY, AGGREGATE_ALL:
		st.n = append(st.n, 0)
		st.bflag = append(st.bflag, false)
	}
}

// updateReducible folds value v into group gid's state for aggregator kind
// t. Copied verbatim (in behavior) from accumulator.go: countAcc.update,
// sumAcc.update, meanAcc.update, minAcc.update, maxAcc.update, and
// varAcc.update's Welford recurrence.
func updateReducible(st *reducibleState, t AggregateType, gid int, v float64) {
	switch t {
	case AGGREGATE_COUNT:
		st.n[gid]++
	case AGGREGATE_SUM, AGGREGATE_MEAN:
		st.sum[gid] += v
		st.n[gid]++
	case AGGREGATE_MIN:
		if !st.set[gid] || v < st.val[gid] {
			st.val[gid] = v
			st.set[gid] = true
		}
	case AGGREGATE_MAX:
		if !st.set[gid] || v > st.val[gid] {
			st.val[gid] = v
			st.set[gid] = true
		}
	case AGGREGATE_STD, AGGREGATE_VARIANCE:
		st.wn[gid]++
		d := v - st.mean[gid]
		st.mean[gid] += d / st.wn[gid]
		st.m2[gid] += d * (v - st.mean[gid])
	case AGGREGATE_ANY:
		st.n[gid]++
		if v != 0 {
			st.bflag[gid] = true
		}
	case AGGREGATE_ALL:
		st.n[gid]++
		if v == 0 {
			st.bflag[gid] = true
		}
	}
}

// mergeReducible folds src's group sgid into dst's group dgid for aggregator
// kind t (the parallel worker merge). Copied verbatim (in behavior) from
// accumulator.go: countAcc.merge, sumAcc.merge, meanAcc.merge (min/max reuse
// updateReducible exactly as minAcc.merge/maxAcc.merge call a.update(b.v)),
// and varAcc.merge's pairwise Welford combination (Chan et al.).
func mergeReducible(dst *reducibleState, dgid int, src *reducibleState, sgid int, t AggregateType) {
	switch t {
	case AGGREGATE_COUNT:
		dst.n[dgid] += src.n[sgid]
	case AGGREGATE_SUM, AGGREGATE_MEAN:
		dst.sum[dgid] += src.sum[sgid]
		dst.n[dgid] += src.n[sgid]
	case AGGREGATE_MIN:
		if src.set[sgid] {
			updateReducible(dst, AGGREGATE_MIN, dgid, src.val[sgid])
		}
	case AGGREGATE_MAX:
		if src.set[sgid] {
			updateReducible(dst, AGGREGATE_MAX, dgid, src.val[sgid])
		}
	case AGGREGATE_STD, AGGREGATE_VARIANCE:
		bn := src.wn[sgid]
		if bn == 0 {
			return
		}
		an := dst.wn[dgid]
		if an == 0 {
			dst.wn[dgid], dst.mean[dgid], dst.m2[dgid] = src.wn[sgid], src.mean[sgid], src.m2[sgid]
			return
		}
		delta := src.mean[sgid] - dst.mean[dgid]
		tot := an + bn
		dst.m2[dgid] += src.m2[sgid] + delta*delta*an*bn/tot
		dst.mean[dgid] += delta * bn / tot
		dst.wn[dgid] = tot
	case AGGREGATE_ANY, AGGREGATE_ALL:
		dst.n[dgid] += src.n[sgid]
		if src.bflag[sgid] {
			dst.bflag[dgid] = true
		}
	}
}

// finalizeReducible computes group gid's (value, isNull) for aggregator kind
// t (with ddof, applicable to STD/VARIANCE only). Copied verbatim (in
// behavior) from accumulator.go: countAcc.result, sumAcc.result,
// meanAcc.result, minAcc.result/maxAcc.result, and varAcc.result (sample
// variance/std with n-ddof denominator, null when n==0 or n-ddof<=0).
func finalizeReducible(st *reducibleState, gid int, t AggregateType, ddof int) (float64, bool) {
	switch t {
	case AGGREGATE_COUNT:
		return float64(st.n[gid]), false
	case AGGREGATE_SUM:
		if st.n[gid] == 0 {
			return 0, true
		}
		return st.sum[gid], false
	case AGGREGATE_MEAN:
		if st.n[gid] == 0 {
			return 0, true
		}
		return st.sum[gid] / float64(st.n[gid]), false
	case AGGREGATE_MIN, AGGREGATE_MAX:
		return st.val[gid], !st.set[gid]
	case AGGREGATE_STD, AGGREGATE_VARIANCE:
		n := st.wn[gid]
		denom := n - float64(ddof)
		if n == 0 || denom <= 0 {
			return 0, true
		}
		v := st.m2[gid] / denom
		if t == AGGREGATE_STD {
			return math.Sqrt(v), false
		}
		return v, false
	case AGGREGATE_ANY:
		if st.n[gid] == 0 {
			return 0, true
		}
		if st.bflag[gid] {
			return 1, false
		}
		return 0, false
	case AGGREGATE_ALL:
		if st.n[gid] == 0 {
			return 0, true
		}
		if st.bflag[gid] {
			return 0, false
		}
		return 1, false
	}
	panic("finalizeReducible: not a reducible aggregate")
}

// accumulateChunk is the single-chunk accumulation core shared by
// aggregateSerial and aggregate's parallel workers. It builds a fresh
// groupTable over rows [lo, hi) of keyCols, with one reducibleState slot
// (reducible) or collector (holistic) per aggregator per group encountered
// in that row range, and returns the per-(aggregator, group) NA-propagation flags
// for removeNAs == false (see aggregateSerial's doc comment for NA-propagation
// semantics). views/isHol come from prepAggValueColumns and are read-only.
//
// A value at (j, row) is missing iff views[j].nullable and the column's own
// null mask says row is null, OR — for a Float64 value column only —
// math.IsNaN(views[j].f64[row]). The Float64 case preserves the historical
// behavior of the former stats-preprocessing helper, which mapped nulls to NaN, with the engine
// detecting missing values with math.IsNaN: a genuine NaN stored in a
// Float64 column (as opposed to a null cell) is therefore still treated as
// missing, exactly as before. For every other kind, only the null mask
// applies — there is no NaN representation to alias against.
//
// The returned groupTable, reducible states, collectors and NA-propagation flags
// are local to [lo, hi): group ids are dense in first-appearance order
// within this chunk only and are not meaningful outside it (see aggregate's
// doc comment on merging).
func accumulateChunk(keyCols []series.Series, aggs []aggregator, views []aggValueView, isHol []bool, removeNAs bool, lo, hi int) (*groupTable, []*reducibleState, [][]collector, [][]bool) {
	gt := newGroupTable(keyCols)

	// Per-aggregator, per-gid state; grown lazily as new gids appear.
	states := make([]*reducibleState, len(aggs)) // reducible (incl. Count)
	cols := make([][]collector, len(aggs))       // holistic
	propagated := make([][]bool, len(aggs))      // reducible NA-propagation flags (removeNAs == false)
	for j := range aggs {
		if !isHol[j] {
			states[j] = &reducibleState{}
		}
	}

	nGroups := 0
	for row := lo; row < hi; row++ {
		gid := gt.idOf(row)

		// New group: append fresh state for every aggregator.
		if gid == nGroups {
			for j, agg := range aggs {
				if isHol[j] {
					cols[j] = append(cols[j], newQuantileCollector(agg.p, agg.interp))
				} else {
					growReducible(states[j], agg.type_)
				}
				propagated[j] = append(propagated[j], false)
			}
			nGroups++
		}

		for j, agg := range aggs {
			if agg.type_ == AGGREGATE_COUNT {
				updateReducible(states[j], AGGREGATE_COUNT, gid, 0) // value ignored; Count counts rows
				continue
			}

			view := &views[j]
			missing := view.nullable && view.nullMask[row>>3]&(1<<uint(row%8)) != 0
			var vj float64
			if !missing {
				switch view.kind {
				case aggValF64:
					vj = view.f64[row]
					missing = math.IsNaN(vj) // a genuine NaN value is treated as missing, same as today
				case aggValI64:
					vj = float64(view.i64[row])
				case aggValInt:
					vj = float64(view.ints[row])
				case aggValBool:
					if view.bools[row] {
						vj = 1
					}
				case aggValDur:
					vj = float64(view.dur[row])
				default:
					// Unsupported value column type: mirrors the previous
					// out-of-range panic the former stats-preprocessing helper produced.
					vj = view.f64[row]
				}
			}
			if missing {
				if !removeNAs {
					// Propagates the NA to every aggregate for this group under RemoveNAs(false),
					// reducible and holistic alike (Count never reaches here).
					propagated[j][gid] = true
				}
				continue
			}
			if isHol[j] {
				cols[j][gid].collect(vj)
			} else {
				updateReducible(states[j], agg.type_, gid, vj)
			}
		}
	}

	return gt, states, cols, propagated
}

// finalizeAggregate builds the sorted result DataFrame from a fully-populated
// (or fully-merged) group table and its per-aggregator per-group state: the
// key columns (one row per group, taken from each group's representative
// row) followed by one aggregate column per aggregator, sorted by key
// ascending, nulls last, lexicographic across keyCols. Count is emitted as
// Int64s; Any/All are emitted as Bools with a null mask (NA-propagated or
// empty/all-null groups are NULL — there is no bool NaN sentinel); every
// other aggregate as Float64s with a null mask from the per-row isNull flags
// (NA-propagated groups under removeNAs == false surface as non-null NaN,
// matching aggregateSerial).
func finalizeAggregate(df DataFrame, keyCols []series.Series, aggs []aggregator, isHol []bool, removeNAs bool, gt *groupTable, states []*reducibleState, cols [][]collector, propagated [][]bool) DataFrame {
	ctx := df.Context()
	nGroups := gt.numGroups()
	reps := gt.representativeRows()
	order := sortGroupOrder(keyCols, reps)

	// Build the result: key columns first, then one aggregate column per agg.
	result := NewDataFrame(ctx)
	result = appendKeyColumns(result, df, keyCols, reps, order)

	for j, agg := range aggs {
		if agg.type_ == AGGREGATE_COUNT {
			data := make([]int64, nGroups)
			for i, gid := range order {
				v, _ := finalizeReducible(states[j], gid, AGGREGATE_COUNT, agg.ddof)
				data[i] = int64(v)
			}
			result = result.AddSeries(agg.newName, series.NewSeriesInt64(data, nil, false, ctx))
			continue
		}

		if agg.type_ == AGGREGATE_ANY || agg.type_ == AGGREGATE_ALL {
			// Any/All are reducible and never NA-propagated into a non-null NaN
			// sentinel like the other aggregates below — bool has no NaN, so
			// an NA-propagated or empty/all-null group's result is NULL instead.
			data := make([]bool, nGroups)
			var nulls []bool
			for i, gid := range order {
				var v float64
				var isNull bool
				switch {
				case !removeNAs && propagated[j][gid]:
					v, isNull = 0, true
				default:
					v, isNull = finalizeReducible(states[j], gid, agg.type_, agg.ddof)
				}
				data[i] = v != 0
				if isNull {
					if nulls == nil {
						nulls = make([]bool, nGroups)
					}
					nulls[i] = true
				}
			}
			result = result.AddSeries(agg.newName, series.NewSeriesBool(data, nulls, false, ctx))
			continue
		}

		data := make([]float64, nGroups)
		var nulls []bool
		for i, gid := range order {
			var v float64
			var isNull bool
			switch {
			case !removeNAs && propagated[j][gid]:
				// NA propagation overrides both reducible and holistic results.
				v, isNull = math.NaN(), false
			case isHol[j]:
				v, isNull = cols[j][gid].result()
			default:
				v, isNull = finalizeReducible(states[j], gid, agg.type_, agg.ddof)
			}
			data[i] = v
			if isNull {
				if nulls == nil {
					nulls = make([]bool, nGroups)
				}
				nulls[i] = true
			}
		}
		result = result.AddSeries(agg.newName, series.NewSeriesFloat64(data, nulls, false, ctx))
	}

	return result
}

// sortGroupOrder returns the group ids permuted into sorted key order:
// ascending, nulls last, lexicographic across the key columns (compared at each
// group's representative row). With no key columns the identity order (group id
// order, i.e. first-appearance) is returned.
func sortGroupOrder(keyCols []series.Series, reps []int) []int {
	order := make([]int, len(reps))
	for i := range order {
		order[i] = i
	}
	if len(keyCols) == 0 {
		return order
	}
	sort.SliceStable(order, func(i, j int) bool {
		ri, rj := reps[order[i]], reps[order[j]]
		for _, col := range keyCols {
			if c := compareKeyCells(col, ri, rj); c != 0 {
				return c < 0
			}
		}
		return false
	})
	return order
}

// compareKeyCells compares the cells of col at rows ra and rb, returning -1, 0
// or +1. A null cell sorts after any non-null cell (nulls last); two nulls are
// equal. The type switch covers the same key types the group-key coder supports.
func compareKeyCells(col series.Series, ra, rb int) int {
	na, nb := col.IsNull(ra), col.IsNull(rb)
	if na || nb {
		switch {
		case na && nb:
			return 0
		case na:
			return 1 // a is null -> sorts after b
		default:
			return -1 // b is null -> a sorts before
		}
	}

	switch c := col.(type) {
	case series.Bools:
		return compareBool(c.Data_[ra], c.Data_[rb])
	case series.Ints:
		return cmpOrdered(c.Data_[ra], c.Data_[rb])
	case series.Int64s:
		return cmpOrdered(c.Data_[ra], c.Data_[rb])
	case series.Float64s:
		return cmpOrdered(c.Data_[ra], c.Data_[rb])
	case series.Strings:
		return cmpOrdered(*c.Data_[ra], *c.Data_[rb])
	case series.Times:
		return cmpOrdered(c.Data_[ra].UnixNano(), c.Data_[rb].UnixNano())
	case series.Durations:
		return cmpOrdered(int64(c.Data_[ra]), int64(c.Data_[rb]))
	default:
		return cmpOrdered(col.GetAsString(ra), col.GetAsString(rb))
	}
}

// cmpOrdered is a three-way comparison for the ordered key cell types.
func cmpOrdered[T int | int64 | float64 | string](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// compareBool orders false before true.
func compareBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case !a:
		return -1
	default:
		return 1
	}
}

// appendKeyColumns emits one key column per keyCol, one value per group taken
// from the group's representative row, with the groups laid out in the sorted
// order. The type switch covers the same key types as the former groupHelper;
// the representative row's null flag is carried into the emitted column.
func appendKeyColumns(result DataFrame, df DataFrame, keyCols []series.Series, reps, order []int) DataFrame {
	ctx := df.Context()
	n := len(order)

	for k, col := range keyCols {
		name := keyColumnName(df, col, k)

		var keyNulls []bool
		if col.IsNullable() {
			keyNulls = make([]bool, n)
			for i, gid := range order {
				keyNulls[i] = col.IsNull(reps[gid])
			}
		}

		switch c := col.(type) {
		case series.Bools:
			vals := make([]bool, n)
			for i, gid := range order {
				vals[i] = c.Data_[reps[gid]]
			}
			result = result.AddSeries(name, series.NewSeriesBool(vals, keyNulls, false, ctx))

		case series.Ints:
			vals := make([]int, n)
			for i, gid := range order {
				vals[i] = c.Data_[reps[gid]]
			}
			result = result.AddSeries(name, series.NewSeriesInt(vals, keyNulls, false, ctx))

		case series.Int64s:
			vals := make([]int64, n)
			for i, gid := range order {
				vals[i] = c.Data_[reps[gid]]
			}
			result = result.AddSeries(name, series.NewSeriesInt64(vals, keyNulls, false, ctx))

		case series.Float64s:
			vals := make([]float64, n)
			for i, gid := range order {
				vals[i] = c.Data_[reps[gid]]
			}
			result = result.AddSeries(name, series.NewSeriesFloat64(vals, keyNulls, false, ctx))

		case series.Strings:
			vals := make([]*string, n)
			for i, gid := range order {
				vals[i] = c.Data_[reps[gid]]
			}
			result = result.AddSeries(name, series.NewSeriesStringFromPtrs(vals, keyNulls, false, ctx))

		case series.Times:
			vals := make([]time.Time, n)
			for i, gid := range order {
				vals[i] = c.Data_[reps[gid]]
			}
			result = result.AddSeries(name, series.NewSeriesTime(vals, keyNulls, false, ctx))

		case series.Durations:
			vals := make([]time.Duration, n)
			for i, gid := range order {
				vals[i] = c.Data_[reps[gid]]
			}
			result = result.AddSeries(name, series.NewSeriesDuration(vals, keyNulls, false, ctx))

		default:
			// Unsupported key type: keep names and rows aligned.
			result = result.AddSeries(name, series.NewSeriesNA(n, ctx))
		}
	}

	return result
}

// keyColumnName recovers the dataframe name of a key column. The engine is
// called with keyCols drawn from df.series (see buildGroupKeyCols), so the
// column is matched back to its name by the identity of its backing data array.
// The k-based fallback is defensive: it is unreachable when keyCols come from df.
func keyColumnName(df DataFrame, col series.Series, k int) string {
	if p := seriesBackingPtr(col); p != nil {
		for i, s := range df.series {
			if seriesBackingPtr(s) == p {
				return df.names[i]
			}
		}
	}
	return "key_" + strconv.Itoa(k)
}

// seriesBackingPtr returns a comparable identity for a series' backing data
// array (the address of its first element), or nil when the series is empty or
// of an unrecognized type. Two series that share a backing array (e.g. the
// stored column and the value returned by df.C) yield equal identities.
func seriesBackingPtr(s series.Series) any {
	switch c := s.(type) {
	case series.Bools:
		if len(c.Data_) == 0 {
			return nil
		}
		return &c.Data_[0]
	case series.Ints:
		if len(c.Data_) == 0 {
			return nil
		}
		return &c.Data_[0]
	case series.Int64s:
		if len(c.Data_) == 0 {
			return nil
		}
		return &c.Data_[0]
	case series.Float64s:
		if len(c.Data_) == 0 {
			return nil
		}
		return &c.Data_[0]
	case series.Strings:
		if len(c.Data_) == 0 {
			return nil
		}
		return &c.Data_[0]
	case series.Times:
		if len(c.Data_) == 0 {
			return nil
		}
		return &c.Data_[0]
	case series.Durations:
		if len(c.Data_) == 0 {
			return nil
		}
		return &c.Data_[0]
	default:
		return nil
	}
}
