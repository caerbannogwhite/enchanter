package series

import (
	"testing"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/utils"
)

func Test_SeriesNA_Append(t *testing.T) {
	var res Series
	var baseMask, expectedMask []bool

	nas := NewSeriesNA(10, ctx)

	baseMask = []bool{true, true, true, true, true, true, true, true, true, true}
	int64s := NewSeriesInt64([]int64{1, 2, 3, 4, 5}, []bool{false, true, false, true, false}, false, ctx)
	strings := NewSeriesString([]string{"a", "b", "c", "d", "e"}, []bool{false, true, false, true, false}, false, ctx)

	// Append nil
	res = nas.Append(nil)
	expectedMask = append(baseMask, true)
	if res.Len() != 11 {
		t.Errorf("Expected length 11, got %d", res.Len())
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append NAs
	res = nas.Append(NewSeriesNA(5, ctx))
	expectedMask = append(baseMask, true, true, true, true, true)
	if res.Len() != 15 {
		t.Errorf("Expected length 15, got %d", res.Len())
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append int64
	res = nas.Append(int64(1))
	expectedMask = append(baseMask, false)
	if res.Len() != 11 {
		t.Errorf("Expected length 11, got %d", res.Len())
	}
	if res.Get(10).(int64) != 1 {
		t.Errorf("Expected last element to be 1, got %v", res.Get(10))
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append NullableInt64
	res = nas.Append(enchanter.NullableInt64{Value: 1, Valid: true})
	expectedMask = append(baseMask, false)
	if res.Len() != 11 {
		t.Errorf("Expected length 11, got %d", res.Len())
	}
	if res.Get(10).(int64) != 1 {
		t.Errorf("Expected last element to be 1, got %v", res.Get(10))
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append NullableInt64
	res = nas.Append(enchanter.NullableInt64{Value: 1, Valid: false})
	expectedMask = append(baseMask, true)
	if res.Len() != 11 {
		t.Errorf("Expected length 11, got %d", res.Len())
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append []int64
	res = nas.Append([]int64{1, 2, 3, 4, 5})
	expectedMask = append(baseMask, false, false, false, false, false)
	if res.Len() != 15 {
		t.Errorf("Expected length 15, got %d", res.Len())
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append []NullableInt64
	res = nas.Append([]enchanter.NullableInt64{
		{Value: 1, Valid: true},
		{Value: 2, Valid: false},
		{Value: 3, Valid: true},
		{Value: 4, Valid: false},
		{Value: 5, Valid: true}})
	expectedMask = append(baseMask, false, true, false, true, false)
	if res.Len() != 15 {
		t.Errorf("Expected length 15, got %d", res.Len())
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append Int64s
	res = nas.Append(int64s)
	expectedMask = append(baseMask, false, true, false, true, false)
	if res.Len() != 15 {
		t.Errorf("Expected length 15, got %d", res.Len())
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append string
	res = nas.Append("a")
	expectedMask = append(baseMask, false)
	if res.Len() != 11 {
		t.Errorf("Expected length 11, got %d", res.Len())
	}
	if res.Get(10).(string) != "a" {
		t.Errorf("Expected last element to be a, got %v", res.Get(10))
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append NullableString
	res = nas.Append(enchanter.NullableString{Value: "a", Valid: true})
	expectedMask = append(baseMask, false)
	if res.Len() != 11 {
		t.Errorf("Expected length 11, got %d", res.Len())
	}
	if res.Get(10).(string) != "a" {
		t.Errorf("Expected last element to be a, got %v", res.Get(10))
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append NullableString
	res = nas.Append(enchanter.NullableString{Value: "a", Valid: false})
	expectedMask = append(baseMask, true)
	if res.Len() != 11 {
		t.Errorf("Expected length 11, got %d", res.Len())
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append []string
	res = nas.Append([]string{"a", "b", "c", "d", "e"})
	expectedMask = append(baseMask, false, false, false, false, false)
	if res.Len() != 15 {
		t.Errorf("Expected length 15, got %d", res.Len())
	}
	if !utils.CheckEqSliceString(res.Data().([]string), []string{NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, "a", "b", "c", "d", "e"}, nil, "Append") {
		t.Errorf("Expecting %v, got %v", []string{NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, "a", "b", "c", "d", "e"}, res.Data().([]string))
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append []NullableString
	res = nas.Append([]enchanter.NullableString{
		{Value: "a", Valid: true},
		{Value: "b", Valid: false},
		{Value: "c", Valid: true},
		{Value: "d", Valid: false},
		{Value: "e", Valid: true}})
	expectedMask = append(baseMask, false, true, false, true, false)
	if res.Len() != 15 {
		t.Errorf("Expected length 15, got %d", res.Len())
	}
	if !utils.CheckEqSliceString(res.Data().([]string), []string{NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, "a", NA_TEXT, "c", NA_TEXT, "e"}, nil, "Append") {
		t.Errorf("Expecting %v, got %v", []string{NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, "a", NA_TEXT, "c", NA_TEXT, "e"}, res.Data().([]string))
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}

	// Append Strings
	res = nas.Append(strings)
	expectedMask = append(baseMask, false, true, false, true, false)
	if res.Len() != 15 {
		t.Errorf("Expected length 15, got %d", res.Len())
	}
	if !utils.CheckEqSliceString(res.Data().([]string), []string{NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, "a", NA_TEXT, "c", NA_TEXT, "e"}, nil, "Append") {
		t.Errorf("Expecting %v, got %v", []string{NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, "a", NA_TEXT, "c", NA_TEXT, "e"}, res.Data().([]string))
	}
	if !utils.CheckEqSlice(res.NullMask(), expectedMask, nil, "Append") {
		t.Errorf("Expected %v, got %v", expectedMask, res.NullMask())
	}
}

func Test_SeriesNA_Arithmetic_Mul(t *testing.T) {
	nas := NewSeriesNA(1, ctx)
	nav := NewSeriesNA(10, ctx)

	int64s := NewSeriesInt64([]int64{1}, nil, false, ctx)
	int64v := NewSeriesInt64([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, false, ctx)
	int64s_ := NewSeriesInt64([]int64{1}, nil, false, ctx).SetNullMask([]bool{true})
	int64v_ := NewSeriesInt64([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, false, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	float64s := NewSeriesFloat64([]float64{1}, nil, false, ctx)
	float64v := NewSeriesFloat64([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, false, ctx)
	float64s_ := NewSeriesFloat64([]float64{1}, nil, false, ctx).SetNullMask([]bool{true})
	float64v_ := NewSeriesFloat64([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, false, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	// scalar | na
	if !utils.CheckEqSlice(nas.Mul(nas).NullMask(), []bool{true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Mul(nas).NullMask())
	}
	if !utils.CheckEqSlice(nas.Mul(nav).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Mul(nav).NullMask())
	}

	// scalar | int64
	if res, ok := nas.Mul(int64s).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := nas.Mul(int64v).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if !utils.CheckEqSlice(nas.Mul(int64s).NullMask(), []bool{true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Mul(int64s).NullMask())
	}
	if !utils.CheckEqSlice(nas.Mul(int64v).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Mul(int64v).NullMask())
	}
	if !utils.CheckEqSlice(nas.Mul(int64s_).NullMask(), []bool{true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Mul(int64s_).NullMask())
	}
	if !utils.CheckEqSlice(nas.Mul(int64v_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Mul(int64v_).NullMask())
	}

	// scalar | float64
	if res, ok := nas.Mul(float64s).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := nas.Mul(float64v).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if !utils.CheckEqSlice(nas.Mul(float64s).NullMask(), []bool{true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Mul(float64s).NullMask())
	}
	if !utils.CheckEqSlice(nas.Mul(float64v).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Mul(float64v).NullMask())
	}
	if !utils.CheckEqSlice(nas.Mul(float64s_).NullMask(), []bool{true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Mul(float64s_).NullMask())
	}
	if !utils.CheckEqSlice(nas.Mul(float64v_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Mul(float64v_).NullMask())
	}

	// vector | na
	if !utils.CheckEqSlice(nav.Mul(nas).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Mul(nas).NullMask())
	}
	if !utils.CheckEqSlice(nav.Mul(nav).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Mul(nav).NullMask())
	}

	// vector | int64
	if res, ok := nav.Mul(int64s).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Mul(int64v).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if !utils.CheckEqSlice(nav.Mul(int64s).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Mul(int64s).NullMask())
	}
	if !utils.CheckEqSlice(nav.Mul(int64v).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Mul(int64v).NullMask())
	}
	if !utils.CheckEqSlice(nav.Mul(int64s_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Mul(int64s_).NullMask())
	}
	if !utils.CheckEqSlice(nav.Mul(int64v_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Mul(int64v_).NullMask())
	}

	// vector | float64
	if res, ok := nav.Mul(float64s).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Mul(float64v).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if !utils.CheckEqSlice(nav.Mul(float64s).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nav.Mul(float64s).NullMask())
	}
	if !utils.CheckEqSlice(nav.Mul(float64v).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nav.Mul(float64v).NullMask())
	}
	if !utils.CheckEqSlice(nav.Mul(float64s_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nav.Mul(float64s_).NullMask())
	}
	if !utils.CheckEqSlice(nav.Mul(float64v_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Mul") {
		t.Errorf("Expected %v, got %v", []bool{true}, nav.Mul(float64v_).NullMask())
	}
}

func Test_SeriesNA_Arithmetic_Add(t *testing.T) {
	nas := NewSeriesNA(1, ctx)
	nav := NewSeriesNA(10, ctx)

	ints := NewSeriesInt64([]int64{1}, nil, false, ctx)
	intv := NewSeriesInt64([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, false, ctx)
	ints_ := NewSeriesInt64([]int64{1}, nil, false, ctx).SetNullMask([]bool{true})
	intv_ := NewSeriesInt64([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, false, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	strings := NewSeriesString([]string{"a"}, nil, false, ctx)
	stringv := NewSeriesString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, nil, false, ctx)
	strings_ := NewSeriesString([]string{"a"}, nil, false, ctx).SetNullMask([]bool{true})
	stringv_ := NewSeriesString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, nil, false, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	// scalar | na
	if !utils.CheckEqSlice(nas.Add(nas).NullMask(), []bool{true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Add(nas).NullMask())
	}
	if !utils.CheckEqSlice(nas.Add(nav).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Add(nav).NullMask())
	}

	// scalar | int64
	if res, ok := nas.Add(ints).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := nas.Add(intv).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if !utils.CheckEqSlice(nas.Add(ints).NullMask(), []bool{true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Add(ints).NullMask())
	}
	if !utils.CheckEqSlice(nas.Add(intv).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Add(intv).NullMask())
	}
	if !utils.CheckEqSlice(nas.Add(ints_).NullMask(), []bool{true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Add(ints_).NullMask())
	}
	if !utils.CheckEqSlice(nas.Add(intv_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Add(intv_).NullMask())
	}

	// scalar | string: NA propagation, the result is all null
	if res, ok := nas.Add(strings).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := nas.Add(stringv).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nas.Add(strings_).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := nas.Add(stringv_).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}

	// vector | na
	if !utils.CheckEqSlice(nav.Add(nas).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Add(nas).NullMask())
	}
	if !utils.CheckEqSlice(nav.Add(nav).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Add(nav).NullMask())
	}

	// vector | int64
	if res, ok := nav.Add(ints).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Add(intv).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if !utils.CheckEqSlice(nav.Add(ints).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Add(ints).NullMask())
	}
	if !utils.CheckEqSlice(nav.Add(intv).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Add(intv).NullMask())
	}
	if !utils.CheckEqSlice(nav.Add(ints_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Add(ints_).NullMask())
	}
	if !utils.CheckEqSlice(nav.Add(intv_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Add(intv_).NullMask())
	}

	// vector | string: NA propagation, the result is all null
	if res, ok := nav.Add(strings).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Add(stringv).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Add(strings_).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Add(stringv_).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
}

func Test_SeriesNA_Boolean_Or(t *testing.T) {
	nas := NewSeriesNA(1, ctx)
	nav := NewSeriesNA(10, ctx)

	bools := NewSeriesBool([]bool{true}, nil, true, ctx)
	boolv := NewSeriesBool([]bool{true, false, true, false, true, false, true, true, false, false}, nil, true, ctx)
	bools_ := NewSeriesBool([]bool{true}, nil, true, ctx).SetNullMask([]bool{true})
	boolv_ := NewSeriesBool([]bool{true, false, true, false, true, false, true, true, false, false}, nil, true, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	// scalar | na
	if !utils.CheckEqSlice(nas.Or(nas).NullMask(), []bool{true}, nil, "NA Or") {
		t.Errorf("Expected %v, got %v", []bool{true}, nas.Or(nas).NullMask())
	}
	if !utils.CheckEqSlice(nas.Or(nav).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Or") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nas.Or(nav).NullMask())
	}

	// scalar | bool: NA propagation, the result is all null
	if res, ok := nas.Or(bools).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := nas.Or(boolv).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nas.Or(bools_).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := nas.Or(boolv_).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}

	// vector | na
	if !utils.CheckEqSlice(nav.Or(nas).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Or") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Or(nas).NullMask())
	}
	if !utils.CheckEqSlice(nav.Or(nav).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "NA Or") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, nav.Or(nav).NullMask())
	}

	// vector | bool: NA propagation, the result is all null
	if res, ok := nav.Or(bools).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Or(boolv).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Or(bools_).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := nav.Or(boolv_).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
}

// Casting NAs to Strings must fill the interned NA text, not nil
// pointers: the generated operators dereference the data uncheckedly by
// the NA-text convention. This is the path a mixed list like
// ['a', na] takes downstream, where the na element is cast to the list
// type before it is appended.
func Test_SeriesNA_CastToStrings(t *testing.T) {
	c := NewSeriesNA(2, ctx).Cast(meta.StringType)
	for _, p := range c.(Strings).Interned() {
		if p == nil {
			t.Fatal("cast null slots must hold the interned NA text, not nil")
		}
	}

	s := NewSeriesString([]string{"a"}, nil, true, ctx).Append(c)
	res := s.Add("f")
	if res.Err() != nil {
		t.Fatal(res.Err())
	}
	if res.Get(0).(string) != "af" {
		t.Errorf("value: expected af, got %v", res.Get(0))
	}
	if !res.IsNull(1) || !res.IsNull(2) {
		t.Error("nulls must stay null through the concatenation")
	}
}
