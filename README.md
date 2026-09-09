## Enchanter 🧙‍♂️

[![Go Reference](https://pkg.go.dev/badge/github.com/caerbannogwhite/enchanter.svg)](https://pkg.go.dev/github.com/caerbannogwhite/enchanter)
[![CI](https://github.com/caerbannogwhite/enchanter/actions/workflows/ci.yml/badge.svg)](https://github.com/caerbannogwhite/enchanter/actions/workflows/ci.yml)

> Formerly known as **aargh** — renamed in v0.2.0. The v0.1.x releases remain
> available under the old module path.

Enchanter is a data wrangling library in pure Go, in the spirit of Pandas and
Polars in Python and dplyr in R: typed, nullable columns (series) composed
into DataFrames with select / filter / group by / join / sort / aggregate
pipelines. Series and DataFrames interoperate with
[Apache Arrow](https://arrow.apache.org/), which also powers the Parquet and
Arrow IPC readers and writers.

Enchanter is a work in progress and the API is not stable yet.

### Install

```sh
go get github.com/caerbannogwhite/enchanter
```

Requires Go 1.26+. Pure Go, no cgo.

### Quick start

```go
package main

import (
	"strings"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/dataframe"
)

func main() {
	data1 := `
name,age,weight,junior,department,salary band
Alice C,29,75.0,F,HR,4
John Doe,30,80.5,true,IT,2
Bob,31,85.0,F,IT,4
Jane H,25,60.0,false,IT,4
Mary,28,70.0,false,IT,3
Oliver,32,90.0,true,HR,1
Ursula,27,65.0,f,Business,4
Charlie,33,60.0,t,Business,2
Megan,26,55.0,F,IT,3`

	dataframe.NewBaseDataFrame(enchanter.NewContext()).
		FromCsv().
		SetReader(strings.NewReader(data1)).
		Read().
		Select("department", "age", "weight", "junior").
		GroupBy("department").
		Agg(dataframe.Min("age"), dataframe.Max("weight"), dataframe.Mean("junior"), dataframe.Count()).
		Run().
		PPrint(dataframe.NewPPrintParams().SetUseLipGloss(true))
}

//   BaseDataFrame: 3 rows, 5 columns
// ╭────────────┬──────────┬─────────────┬──────────────┬───────╮
// │ department │ min(age) │ max(weight) │ mean(junior) │ n     │
// ├────────────┼──────────┼─────────────┼──────────────┼───────┤
// │ String     │ Float64  │ Float64     │ Float64      │ Int64 │
// ├────────────┼──────────┼─────────────┼──────────────┼───────┤
// │ HR         │    29.00 │       90.00 │       0.5000 │ 2.000 │
// │ IT         │    25.00 │       85.00 │       0.2000 │ 5.000 │
// │ Business   │    27.00 │       65.00 │       0.5000 │ 2.000 │
// ╰────────────┴──────────┴─────────────┴──────────────┴───────╯
```

More runnable examples live in [examples](examples) and in the package
documentation on [pkg.go.dev](https://pkg.go.dev/github.com/caerbannogwhite/enchanter).

### Reading and writing files

| Format              | Read | Write | Notes                                    |
| ------------------- | :--: | :---: | ---------------------------------------- |
| CSV                 |  ✅  |  ✅   | delimiter, header, type guessing         |
| Parquet             |  ✅  |  ✅   | via Apache Arrow; types + nulls preserved |
| Arrow IPC (Feather) |  ✅  |  ✅   |                                          |
| XPT (SAS transport) |  ✅  |  ✅   | versions 5–9                             |
| XLSX                |  ✅  |  ✅   |                                          |
| JSON                |  ✅  |  ✅   | record-oriented                          |
| HTML                |  ✅  |  ✅   | tables                                   |
| Markdown            |  ✅  |  ✅   | tables                                   |
| SAS7BDAT            |  ✅  |   —   | read-only, via kshedden/datareader        |

All readers and writers share the same builder style:

```go
// Parquet round trip: types and nulls survive, unlike CSV.
err := df.ToParquet().SetPath("people.parquet").Write()

df2 := dataframe.NewBaseDataFrame(ctx).
	FromParquet().
	SetPath("people.parquet").
	Read()
```

See [examples/parquet](examples/parquet/main.go) for a runnable version.

### Apache Arrow interop

Every series can produce an Arrow array and every DataFrame an Arrow record
batch, which makes Enchanter data exchangeable with anything that speaks
Arrow (DuckDB, Polars, pandas, DataFusion, Spark, ...):

```go
rec := df.ToArrowRecord() // freshly built, owned by the record
defer rec.Release()       // optional under the default GC-backed allocator

df2 := dataframe.NewBaseDataFrameFromArrowRecord(rec, ctx)
```

Conversion notes:

- Enchanter null masks map onto Arrow validity bitmaps in both directions.
- `series.ArrowArrayToSeries` accepts more Arrow types than Enchanter has
  series types (all integer widths, Float32, Date32/64, every timestamp
  unit, LargeString, ...) and funnels them onto the native types: integers
  normalize to int64, timestamps to nanoseconds.
- Conversions are materialized copies — no reference to the source Arrow
  memory is retained, so you may release inputs immediately.
- Releasing values returned to you is optional under GC-backed allocators
  (the default); see `Context.Allocator` for the exact ownership contract.

### Data types

Seven nullable, columnar series types. Nulls live in a separate bit-packed
mask, so a column with no nulls carries no null overhead.

| Series      | Go element type                        |
| ----------- | -------------------------------------- |
| `Bools`     | `bool`                                 |
| `Ints`      | `int`                                  |
| `Int64s`    | `int64`                                |
| `Float64s`  | `float64`                              |
| `Strings`   | `*string` (interned via a shared pool) |
| `Times`     | `time.Time`                            |
| `Durations` | `time.Duration`                        |

Not implemented (would be added on demand): narrower integers (`Int8/16/32`),
`Float32`, complex numbers, and a bit-packed memory-optimized `Bool`.

### Operations

**Series** carry element-wise arithmetic (`Add`, `Sub`, `Mul`, `Div`, `Mod`,
`Exp`, `Neg`), comparison (`Eq`, `Ne`, `Lt`, `Le`, `Gt`, `Ge`), and boolean
(`And`, `Or`, `Not`) operators, plus `Coalesce` (fill nulls from another
series or a scalar), `Filter` (by a `[]bool` / `[]int` or a `Bools` / `Ints`
series), null-aware `Group` / `SubGroup` and `Sort` / `SortRev`, and `Map`,
`Take`, `Cast`, `Append`.

**DataFrame**

| Operation            | Status | Notes                                    |
| -------------------- | :----: | ---------------------------------------- |
| Select               |   ✅   | exact names; `SelectMatching` for regex  |
| Filter               |   ✅   | by a `Bools` series                      |
| GroupBy + Agg        |   ✅   | null-aware group keys                    |
| Join                 |   ✅   | inner / left / right / outer, null-aware |
| OrderBy              |   ✅   | multi-key, ascending / descending        |
| Take                 |   ✅   |                                          |
| Pivot (longer/wider) |   🚧   | in progress on `dev-pivot`               |
| Map                  |   ❌   | planned                                  |
| Stack / Append       |   ❌   | planned                                  |

`Select` takes **exact column names**, in the order given; naming a column that
does not exist is an error rather than a silently missing column. Pattern
selection lives in `SelectMatching`, which takes regular expressions matched
unanchored — so `SelectMatching("Car")` also selects `CarOrigin`.

**Aggregations** (via `Agg`): `Count`, `Sum`, `Mean`, `Min`, `Max`, `Std`,
`Variance`, `Median`, `Quantile`, `Any`, and `All` are all supported, with
three deliberate behaviors:

- Results are **sorted by group key** (ascending, nulls last).
- Null values (NAs) are **skipped by default**; `RemoveNAs(false)` makes
  NAs **propagate** — a single NA in a group's value column makes that
  group's result NA for every aggregate (holistic and reducible alike),
  surfaced as `NaN` for the numeric aggregates and as null for `Any`/`All`.
- `Any` / `All` return a `Bools` column, not `Float64s`.

Std/Variance take a `WithDDoF` option (default population, `ddof=0`);
Median/Quantile take a `WithInterpolation` option.

### Development

```sh
go test -short ./...     # fast unit tests (skips large local benchmark fixtures)
go test ./...            # full suite
go test -bench . ./...   # benchmarks (need local fixtures in testdata/)
go generate ./series/    # regenerate *_base.go / *_ops.go via the generators module
```

CI runs build, vet, gofmt, a generated-code freshness check, the test suite
and the race detector on Linux and Windows.

### Roadmap

Enchanter is pre-1.0. Releases are themed so each has a single focus, working
toward a stable 1.0.

**0.3.0 — measure & clean up**. Deprecation cleanup (`arrow.Record` →
`RecordBatch`), aggregation correctness fixes, CI + lint, and a *measured*
decision on Arrow-native storage: **not worth it** — in pure Go, Arrow-backed
columns only match plain Go slices, and arrow-go has no grouped-aggregation
kernels, so Arrow stays an interop layer (Parquet, IPC, ecosystem handoff). See
the [storage measurement](docs/superpowers/specs/2026-08-08-arrow-native-storage-migration.md).

**0.4.0 — make it fast** (current). The performance headline, validated by a
[spike](benchmarking/README.md):

- [x] Single-pass, parallel hash aggregation for `GroupBy().Agg()` — shipped;
      on the h2oai-style Q1 sum-by-id benchmark (1e7 rows) it went from ~142ms
      to ~31ms, beating Polars' ~43ms.
- [ ] Fused combinators for element-wise op chains (no intermediate arrays).
- [ ] Benchmark-regression tracking so the gains don't rot.

**0.5.0 — stabilize for 1.0** (breaking changes, batched together):

- [ ] Public API: hide internal fields (`Data_`, `NullMask_`), standardize the
      reader constructors.
- [x] `Select` split into an exact-name `Select` and a pattern-matching
      `SelectMatching`, so a plain column name can no longer silently
      over-select (`"Car"` used to also match `"CarOrigin"`).
- [ ] Generics to collapse the generated per-type code — *spike-gated* (only if
      it measurably shrinks the code without regressing speed).
- [ ] Broaden test coverage.
- [x] SAS7BDAT: decided — reading is delegated to
      [kshedden/datareader](https://github.com/kshedden/datareader)
      (BSD-3-Clause) rather than finishing the bespoke parser, verified
      against that project's reference data; there is no write path (use XPT
      to hand data to SAS).
- [x] Internal naming: the legacy `__gdl_` / leading-double-underscore helper
      names (Gandalff-era, pre-rename) are retired for idiomatic Go unexported
      names.
- [x] Generated `*_ops.go` readability: the length×nullability `if` nesting is
      flattened — nullability is resolved at run time by a shared null-mask
      helper and the length cases are a flat `switch` — shrinking the emitted
      operator code from ~44k to ~13k lines with byte-identical behavior.
- [x] `Coalesce` on every series type: fill null elements from another series
      or a scalar; a result element is null only when both operands are null
      there.

**1.0 — commit** to the stable API.

Parking lot (unversioned, picked up as they fit): dictionary-encoded (factor)
strings; pivot longer/wider (in progress on `dev-pivot`); custom aggregators
(archived on `archive/dev-fix-agg`); stricter CSV type guessing (archived on
`archive/dev-0.1.3`); chunked series; SPSS reader/writer; JSON-by-records;
configurable time format; `Set(i []int, …)` and `Slice(i []int)`; url
resolvers and format options on the I/O builders; PrettyPrint and
filtering-interface polish; memory-optimized `Bool` / `uint64` null mask.

### Dependencies

Built with:

- [arrow-go](https://github.com/apache/arrow-go)
- [xslx](https://github.com/tealeg/xlsx/tree/master)
- [lipgloss](https://github.com/charmbracelet/lipgloss)
- [datareader](https://github.com/kshedden/datareader) (SAS7BDAT reading)

### License

See [LICENSE](LICENSE).
