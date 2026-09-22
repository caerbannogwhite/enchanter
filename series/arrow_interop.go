package series

import (
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/caerbannogwhite/enchanter"
)

// ArrowArrayToSeries converts an Arrow array into the corresponding enchanter Series
// type. The array's data is materialized (copied) into Go slices; no reference to
// the input array is retained, so the caller keeps ownership and may Release it
// as soon as this function returns.
func ArrowArrayToSeries(arr arrow.Array, ctx *enchanter.Context) Series {
	if arr == nil {
		return Errors{"ArrowArrayToSeries: nil array"}
	}

	n := arr.Len()

	// Build enchanter null mask from Arrow validity bitmap.
	var isNullable bool
	var nullMask []uint8
	if arr.NullN() > 0 {
		isNullable = true
		nbytes := (n + 7) / 8
		nullMask = make([]uint8, nbytes)
		for i := 0; i < n; i++ {
			if arr.IsNull(i) {
				nullMask[i>>3] |= 1 << uint(i%8)
			}
		}
	} else {
		nullMask = make([]uint8, 0)
	}

	switch a := arr.(type) {
	case *array.Boolean:
		data := make([]bool, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = a.Value(i)
			}
		}
		return Bools{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Int64:
		data := make([]int64, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = a.Value(i)
			}
		}
		return Int64s{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Int32:
		data := make([]int, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = int(a.Value(i))
			}
		}
		return Ints{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Int16:
		data := make([]int, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = int(a.Value(i))
			}
		}
		return Ints{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Int8:
		data := make([]int, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = int(a.Value(i))
			}
		}
		return Ints{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Uint64:
		data := make([]int64, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = int64(a.Value(i))
			}
		}
		return Int64s{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Uint32:
		data := make([]int64, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = int64(a.Value(i))
			}
		}
		return Int64s{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Uint16:
		data := make([]int, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = int(a.Value(i))
			}
		}
		return Ints{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Uint8:
		data := make([]int, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = int(a.Value(i))
			}
		}
		return Ints{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Float32:
		data := make([]float64, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = float64(a.Value(i))
			}
		}
		return Float64s{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Float64:
		data := make([]float64, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = a.Value(i)
			}
		}
		return Float64s{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.String:
		data := make([]*string, n)
		for i := 0; i < n; i++ {
			if a.IsNull(i) {
				data[i] = ctx.StringPool.Put(enchanter.NA_TEXT)
			} else {
				data[i] = ctx.StringPool.Put(a.Value(i))
			}
		}
		return Strings{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.LargeString:
		data := make([]*string, n)
		for i := 0; i < n; i++ {
			if a.IsNull(i) {
				data[i] = ctx.StringPool.Put(enchanter.NA_TEXT)
			} else {
				data[i] = ctx.StringPool.Put(a.Value(i))
			}
		}
		return Strings{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Timestamp:
		data := make([]time.Time, n)
		unit := a.DataType().(*arrow.TimestampType).Unit
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = timestampToTime(int64(a.Value(i)), unit)
			}
		}
		return Times{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
			timeFormat: ctx.GetDateTimeFormat(),
		}

	case *array.Duration:
		data := make([]time.Duration, n)
		unit := a.DataType().(*arrow.DurationType).Unit
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = durationToGoDuration(int64(a.Value(i)), unit)
			}
		}
		return Durations{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
		}

	case *array.Date32:
		data := make([]time.Time, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = a.Value(i).ToTime()
			}
		}
		return Times{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
			timeFormat: ctx.GetDateTimeFormat(),
		}

	case *array.Date64:
		data := make([]time.Time, n)
		for i := 0; i < n; i++ {
			if !a.IsNull(i) {
				data[i] = a.Value(i).ToTime()
			}
		}
		return Times{
			isNullable: isNullable,
			data:       data,
			nullMask:   nullMask,
			ctx:        ctx,
			timeFormat: ctx.GetDateTimeFormat(),
		}

	case *array.Null:
		return NewSeriesNA(n, ctx)

	default:
		return Errors{"ArrowArrayToSeries: unsupported Arrow type " + arr.DataType().Name()}
	}
}

// timestampToTime converts an Arrow timestamp value to time.Time based on the unit.
func timestampToTime(v int64, unit arrow.TimeUnit) time.Time {
	switch unit {
	case arrow.Second:
		return time.Unix(v, 0)
	case arrow.Millisecond:
		return time.Unix(v/1000, (v%1000)*1e6)
	case arrow.Microsecond:
		return time.Unix(v/1e6, (v%1e6)*1000)
	case arrow.Nanosecond:
		return time.Unix(0, v)
	default:
		return time.Unix(0, v)
	}
}

// durationToGoDuration converts an Arrow duration value to time.Duration based on the unit.
func durationToGoDuration(v int64, unit arrow.TimeUnit) time.Duration {
	switch unit {
	case arrow.Second:
		return time.Duration(v) * time.Second
	case arrow.Millisecond:
		return time.Duration(v) * time.Millisecond
	case arrow.Microsecond:
		return time.Duration(v) * time.Microsecond
	case arrow.Nanosecond:
		return time.Duration(v)
	default:
		return time.Duration(v)
	}
}
