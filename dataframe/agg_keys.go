package dataframe

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/caerbannogwhite/enchanter/series"
)

// groupTableKind selects the keying strategy a groupTable uses, decided once
// in newGroupTable from the number and types of key columns. idOf dispatches
// on kind with a plain switch (a predictable branch), not a per-row
// closure/interface call, so the hot aggregation loop stays allocation-free
// for the common 1-key and 2..8-key cases.
type groupTableKind int

const (
	// gtSingleString: exactly one Strings key column, keyed directly by its
	// interned *string pointer (Data_[row]).
	gtSingleString groupTableKind = iota
	// gtSingleInt64/Ints/Bool/Time/Duration: exactly one key column of an
	// integer-like type, each normalized to a raw int64 and keyed in a
	// shared map[int64]int (Bool -> 0/1, Time -> UnixNano(), Duration ->
	// int64(d)).
	gtSingleInt64
	gtSingleInts
	gtSingleBool
	gtSingleTime
	gtSingleDuration
	// gtSingleFloat64: exactly one Float64s key column, keyed by
	// math.Float64bits(Data_[row]) in a map[uint64]int.
	gtSingleFloat64
	// gtMultiArray: 2..8 key columns, dense-coded per column (makeCellCoder)
	// and packed into a comparable, allocation-free [8]uint64 array key.
	gtMultiArray
	// gtFallback: 0 key columns, more than 8 key columns, or a single key
	// column of a type with no dedicated raw-keying case above. Dense-codes
	// each column and packs the codes into a byte buffer converted to a
	// string map key, exactly as the pre-optimization implementation did.
	gtFallback
)

// groupTable assigns dense, stable group ids to rows based on the exact
// composite value of one or more key columns. Two rows map to the same id
// iff every key column has an equal value for those rows, with null==null
// per column (nulls in a given column group together, distinct from every
// non-null value in that column).
//
// newGroupTable picks a groupTableKind once based on the key columns; idOf
// dispatches on that kind. For 1 key column of a supported type, the row is
// keyed directly off its raw cell value (no dense-coding, no allocation per
// row: an interned *string pointer, a raw int64, or a Float64bits uint64).
// For 2..8 key columns, each column is still dense-coded (makeCellCoder) but
// the per-column codes are packed into a fixed-size comparable [8]uint64
// array instead of a byte-buffer string, avoiding the per-row allocation
// that string(buf) required. 0 keys, >8 keys, and single-key types with no
// dedicated case keep the original dense-code + string(buf) path.
type groupTable struct {
	kind groupTableKind
	reps []int // representative row per group id (index = gid)

	// nullGid is the group id reserved for the null-key group in the
	// single-key raw-keying paths (gtSingle*): -1 until the first null row
	// is seen, after which it holds that group's id. Assigned lazily, like
	// every other group, by order of first appearance — a null row is not
	// special-cased to group 0; whichever distinct key (null or not) is
	// encountered first in row order gets id 0.
	nullGid int

	// Single-key raw column + map: exactly one (col, map) pair below is
	// populated, selected by kind. i64Map is shared by
	// gtSingleInt64/Ints/Bool/Time/Duration since they all key off a raw
	// int64.
	strCol series.Strings
	strMap map[*string]int

	i64Col      series.Int64s
	intsCol     series.Ints
	boolCol     series.Bools
	timeCol     series.Times
	durationCol series.Durations
	i64Map      map[int64]int

	f64Col series.Float64s
	f64Map map[uint64]int

	// 2..8 keys: one dense-code function per column, packed into arrMap.
	coders []func(int) uint64
	arrMap map[[8]uint64]int

	// Fallback (0 keys, >8 keys, unsupported single type): dense-coded
	// per-column, packed into buf and converted to a string map key.
	ids map[string]int
	buf []byte // reused scratch buffer for packing per-row codes
}

// newGroupTable builds a groupTable over the given key columns, choosing a
// keying strategy once (see groupTableKind and the groupTable doc comment).
func newGroupTable(keyCols []series.Series) *groupTable {
	g := &groupTable{nullGid: -1}

	if len(keyCols) == 1 {
		switch c := keyCols[0].(type) {
		case series.Strings:
			g.kind = gtSingleString
			g.strCol = c
			g.strMap = make(map[*string]int, 128)
			return g
		case series.Int64s:
			g.kind = gtSingleInt64
			g.i64Col = c
			g.i64Map = make(map[int64]int, 128)
			return g
		case series.Ints:
			g.kind = gtSingleInts
			g.intsCol = c
			g.i64Map = make(map[int64]int, 128)
			return g
		case series.Bools:
			g.kind = gtSingleBool
			g.boolCol = c
			g.i64Map = make(map[int64]int, 128)
			return g
		case series.Times:
			g.kind = gtSingleTime
			g.timeCol = c
			g.i64Map = make(map[int64]int, 128)
			return g
		case series.Durations:
			g.kind = gtSingleDuration
			g.durationCol = c
			g.i64Map = make(map[int64]int, 128)
			return g
		case series.Float64s:
			g.kind = gtSingleFloat64
			g.f64Col = c
			g.f64Map = make(map[uint64]int, 128)
			return g
		}
	}

	if len(keyCols) >= 2 && len(keyCols) <= 8 {
		g.kind = gtMultiArray
		g.coders = make([]func(int) uint64, len(keyCols))
		for i, col := range keyCols {
			g.coders[i] = makeCellCoder(col)
		}
		g.arrMap = make(map[[8]uint64]int, 128)
		return g
	}

	// 0 keys, >8 keys, or a single key column of an unrecognized type.
	g.kind = gtFallback
	g.coders = make([]func(int) uint64, len(keyCols))
	for i, col := range keyCols {
		g.coders[i] = makeCellCoder(col)
	}
	g.buf = make([]byte, 8*len(keyCols))
	g.ids = make(map[string]int, 128)
	return g
}

// idOf returns the dense group id for row, assigning a new id (and
// capturing row as that group's representative) the first time this exact
// composite key is seen. Dispatches on g.kind with a plain switch — a
// predictable branch, not a per-row closure/interface call.
func (g *groupTable) idOf(row int) int {
	switch g.kind {
	case gtSingleString:
		if g.strCol.IsNull(row) {
			return g.nullGroup(row)
		}
		p := g.strCol.Data_[row]
		if id, ok := g.strMap[p]; ok {
			return id
		}
		id := g.newGroup(row)
		g.strMap[p] = id
		return id

	case gtSingleInt64:
		if g.i64Col.IsNull(row) {
			return g.nullGroup(row)
		}
		return g.internInt64(g.i64Col.Data_[row], row)

	case gtSingleInts:
		if g.intsCol.IsNull(row) {
			return g.nullGroup(row)
		}
		return g.internInt64(int64(g.intsCol.Data_[row]), row)

	case gtSingleBool:
		if g.boolCol.IsNull(row) {
			return g.nullGroup(row)
		}
		v := int64(0)
		if g.boolCol.Data_[row] {
			v = 1
		}
		return g.internInt64(v, row)

	case gtSingleTime:
		if g.timeCol.IsNull(row) {
			return g.nullGroup(row)
		}
		// UnixNano (exact instant), not raw time.Time equality: see
		// makeCellCoder's series.Times case for why.
		return g.internInt64(g.timeCol.Data_[row].UnixNano(), row)

	case gtSingleDuration:
		if g.durationCol.IsNull(row) {
			return g.nullGroup(row)
		}
		return g.internInt64(int64(g.durationCol.Data_[row]), row)

	case gtSingleFloat64:
		if g.f64Col.IsNull(row) {
			return g.nullGroup(row)
		}
		// Bit-identity keying, not value equality: two float bit patterns
		// group together iff Float64bits(a) == Float64bits(b). This means a
		// single Float64 key column groups by *bit pattern*, not by ==: every
		// non-null NaN bit pattern groups with identical NaN bit patterns
		// (NaN != NaN under ==, but here NaNs with the same bits DO group
		// together), and +0.0/-0.0 are DISTINCT groups (their bit patterns
		// differ, even though +0.0 == -0.0). This differs from the
		// multi-key/fallback path below (see makeCellCoder's series.Float64s
		// case), which codes Float64 via a map[float64] keyed by ==
		// value-equality: there, +0.0 and -0.0 collapse into one group, and
		// each NaN value compares unequal to every other value (including
		// itself), so every non-null NaN row becomes its own singleton group.
		// In short: float64 grouping-key semantics differ by key arity.
		v := math.Float64bits(g.f64Col.Data_[row])
		if id, ok := g.f64Map[v]; ok {
			return id
		}
		id := g.newGroup(row)
		g.f64Map[v] = id
		return id

	case gtMultiArray:
		var key [8]uint64
		for i, code := range g.coders {
			key[i] = code(row)
		}
		if id, ok := g.arrMap[key]; ok {
			return id
		}
		id := g.newGroup(row)
		g.arrMap[key] = id
		return id

	default: // gtFallback
		for i, code := range g.coders {
			binary.LittleEndian.PutUint64(g.buf[i*8:], code(row))
		}
		key := string(g.buf) // copies bytes; g.buf is safely reused after this
		if id, ok := g.ids[key]; ok {
			return id
		}
		id := g.newGroup(row)
		g.ids[key] = id
		return id
	}
}

// newGroup assigns the next dense group id and records row as that group's
// representative.
func (g *groupTable) newGroup(row int) int {
	id := len(g.reps)
	g.reps = append(g.reps, row)
	return id
}

// nullGroup returns the reserved null-key group id for the single-key raw
// paths, assigning it (via newGroup, exactly like any other first-seen key)
// the first time a null row is encountered.
func (g *groupTable) nullGroup(row int) int {
	if g.nullGid == -1 {
		g.nullGid = g.newGroup(row)
	}
	return g.nullGid
}

// internInt64 looks up v in the shared single-key int64 map, assigning a new
// group id on first sight. Shared by gtSingleInt64/Ints/Bool/Time/Duration.
func (g *groupTable) internInt64(v int64, row int) int {
	if id, ok := g.i64Map[v]; ok {
		return id
	}
	id := g.newGroup(row)
	g.i64Map[v] = id
	return id
}

// numGroups returns the number of distinct groups seen so far.
func (g *groupTable) numGroups() int { return len(g.reps) }

// representativeRows returns the representative row index for each group,
// indexed by group id.
func (g *groupTable) representativeRows() []int { return g.reps }

// makeCellCoder returns a per-row dense code for one key column: 0 means
// null, distinct non-null values get stable codes 1..k in order of first
// appearance. The type switch covers the same concrete series types as the
// former DataFrame.groupHelper.
//
// For interned Strings the map is keyed by the *string pointer (cheap,
// pointer-identity is stable within a StringPool); the other supported key
// types are keyed by their typed value in a dedicated map to avoid `any`
// boxing on the hot path. Unrecognized series types fall back to the
// generic Get(row) any accessor exposed by series.Series.
//
// makeCellCoder backs groupTable's gtMultiArray (2..8 keys) and gtFallback
// (0, >8, or an unsupported single-type key) paths; the gtSingle* paths key
// directly off the raw cell instead (see groupTable.idOf).
func makeCellCoder(col series.Series) func(int) uint64 {
	switch c := col.(type) {
	case series.Strings:
		codes := make(map[*string]uint64, 16)
		return func(row int) uint64 {
			if c.IsNull(row) {
				return 0
			}
			p := c.Data_[row]
			if v, ok := codes[p]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[p] = v
			return v
		}

	case series.Bools:
		codes := make(map[bool]uint64, 2)
		return func(row int) uint64 {
			if c.IsNull(row) {
				return 0
			}
			cell := c.Data_[row]
			if v, ok := codes[cell]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[cell] = v
			return v
		}

	case series.Ints:
		codes := make(map[int]uint64, 16)
		return func(row int) uint64 {
			if c.IsNull(row) {
				return 0
			}
			cell := c.Data_[row]
			if v, ok := codes[cell]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[cell] = v
			return v
		}

	case series.Int64s:
		codes := make(map[int64]uint64, 16)
		return func(row int) uint64 {
			if c.IsNull(row) {
				return 0
			}
			cell := c.Data_[row]
			if v, ok := codes[cell]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[cell] = v
			return v
		}

	case series.Float64s:
		codes := make(map[float64]uint64, 16)
		return func(row int) uint64 {
			if c.IsNull(row) {
				return 0
			}
			cell := c.Data_[row]
			if v, ok := codes[cell]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[cell] = v
			return v
		}

	case series.Times:
		// Key by UnixNano (exact instant), not the raw time.Time value:
		// time.Time equality via == also compares the monotonic reading and
		// *Location pointer, so two values representing the same instant
		// (t.Equal(t2) == true) can otherwise land in different groups.
		// Mirrors series.Times.Group() (series/time.go), which also keys by
		// UnixNano().
		codes := make(map[int64]uint64, 16)
		return func(row int) uint64 {
			if c.IsNull(row) {
				return 0
			}
			cell := c.Data_[row].UnixNano()
			if v, ok := codes[cell]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[cell] = v
			return v
		}

	case series.Durations:
		codes := make(map[time.Duration]uint64, 16)
		return func(row int) uint64 {
			if c.IsNull(row) {
				return 0
			}
			cell := c.Data_[row]
			if v, ok := codes[cell]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[cell] = v
			return v
		}

	default:
		// Unrecognized series type: fall back to the generic any accessor.
		codes := make(map[any]uint64, 16)
		return func(row int) uint64 {
			if col.IsNull(row) {
				return 0
			}
			cell := col.Get(row)
			if v, ok := codes[cell]; ok {
				return v
			}
			v := uint64(len(codes)) + 1
			codes[cell] = v
			return v
		}
	}
}
