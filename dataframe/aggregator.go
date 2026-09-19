package dataframe

import (
	"fmt"

	"github.com/caerbannogwhite/enchanter/series"
)

type aggregatorBuilder struct {
	df          DataFrame
	removeNAs   bool
	aggregators []aggregator
}

func (ab aggregatorBuilder) RemoveNAs(b bool) aggregatorBuilder {
	ab.removeNAs = b
	return ab
}

func (ab aggregatorBuilder) Run() DataFrame {
	df := ab.df
	if df.err != nil {
		return ab.df
	}

	if len(ab.aggregators) == 0 {
		return df
	}

	// CHECK: aggregators must have unique output names and input series must exist
	aggNewNames := make(map[string]bool)
	for _, agg := range ab.aggregators {

		// Check if output names are unique
		if aggNewNames[agg.newName] {
			df.err = fmt.Errorf("DataFrame.Agg: aggregator output names must be unique")
			return df
		}
		aggNewNames[agg.newName] = true

		// CASE: aggregator count doesn't need an input series
		if agg.type_ != AGGREGATE_COUNT {
			if df.seriesByName(agg.name) == nil {
				df.err = fmt.Errorf("DataFrame.Agg: series \"%s\" not found", agg.name)
				return df
			}

			// The value column must be a type the engine can accumulate
			// (Float64s/Int64s/Ints/Bools/Durations — see newAggValueView).
			// Anything else (e.g. Strings) would otherwise reach
			// accumulateChunk's aggValUnsupported default case and panic with
			// an out-of-range slice index.
			if newAggValueView(df.C(agg.name)).kind == aggValUnsupported {
				df.err = fmt.Errorf("DataFrame.Agg: series \"%s\" has unsupported type %s for aggregator \"%s\"", agg.name, df.C(agg.name).Type(), agg.newName)
				return df
			}
		}
	}

	// CHECK: option applicability — ddof only applies to Std/Variance,
	// interpolation only applies to Median/Quantile.
	for _, agg := range ab.aggregators {
		if agg.ddofSet && agg.type_ != AGGREGATE_STD && agg.type_ != AGGREGATE_VARIANCE {
			df.err = fmt.Errorf("DataFrame.Agg: WithDDoF is only applicable to Std/Variance, not to \"%s\"", agg.newName)
			return df
		}
		if agg.interpSet && agg.type_ != AGGREGATE_MEDIAN && agg.type_ != AGGREGATE_QUANTILE {
			df.err = fmt.Errorf("DataFrame.Agg: WithInterpolation is only applicable to Median/Quantile, not to \"%s\"", agg.newName)
			return df
		}
	}

	if df.isGrouped {
		return aggregate(df, df.buildGroupKeyCols(), ab.aggregators, ab.removeNAs)
	}

	return aggregate(df, nil, ab.aggregators, ab.removeNAs)
}

type AggregateType int8

const (
	AGGREGATE_COUNT AggregateType = iota
	AGGREGATE_SUM
	AGGREGATE_MEAN
	AGGREGATE_MEDIAN
	AGGREGATE_MIN
	AGGREGATE_MAX
	AGGREGATE_STD
	AGGREGATE_VARIANCE
	AGGREGATE_QUANTILE
	AGGREGATE_ANY
	AGGREGATE_ALL
)

func (t AggregateType) isHolistic() bool {
	return t == AGGREGATE_MEDIAN || t == AGGREGATE_QUANTILE
}

const DEFAULT_COUNT_NAME = "n"

type aggregator struct {
	name      string
	newName   string
	type_     AggregateType
	p         float64       // quantile probability (AGGREGATE_QUANTILE)
	ddof      int           // AGGREGATE_STD / AGGREGATE_VARIANCE
	interp    Interpolation // AGGREGATE_MEDIAN / AGGREGATE_QUANTILE
	ddofSet   bool
	interpSet bool
}

func mkAgg(name, newName string, t AggregateType, p float64, opts []AggOption) aggregator {
	c := newAggConfig(opts)
	return aggregator{name, newName, t, p, c.ddof, c.interp, c.ddofSet, c.interpSet}
}

func Count() aggregator {
	return aggregator{DEFAULT_COUNT_NAME, DEFAULT_COUNT_NAME, AGGREGATE_COUNT, 0, 0, Linear, false, false}
}

func Sum(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("sum(%s)", name), AGGREGATE_SUM, 0, opts)
}
func Mean(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("mean(%s)", name), AGGREGATE_MEAN, 0, opts)
}
func Min(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("min(%s)", name), AGGREGATE_MIN, 0, opts)
}
func Max(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("max(%s)", name), AGGREGATE_MAX, 0, opts)
}
func Std(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("std(%s)", name), AGGREGATE_STD, 0, opts)
}
func Variance(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("var(%s)", name), AGGREGATE_VARIANCE, 0, opts)
}
func Median(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("median(%s)", name), AGGREGATE_MEDIAN, 0.5, opts)
}
func Quantile(name string, p float64, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("quantile_%g(%s)", p, name), AGGREGATE_QUANTILE, p, opts)
}
func Any(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("any(%s)", name), AGGREGATE_ANY, 0, opts)
}
func All(name string, opts ...AggOption) aggregator {
	return mkAgg(name, fmt.Sprintf("all(%s)", name), AGGREGATE_ALL, 0, opts)
}

////////////////////////			SORT

type SortParam struct {
	asc     bool
	name    string
	_series series.Series
}

func Asc(name string) SortParam {
	return SortParam{asc: true, name: name}
}

func Desc(name string) SortParam {
	return SortParam{asc: false, name: name}
}
