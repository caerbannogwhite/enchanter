package series

import (
	"math/rand"
	"testing"
	"time"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/meta"
	"github.com/caerbannogwhite/enchanter/utils"
)

func Test_SeriesString_Base(t *testing.T) {
	var data, expected []string

	data = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	mask := []bool{false, false, true, false, false, true, false, false, true, false}

	// Create a new Strings.
	s := NewSeriesString(data, mask, true, ctx)

	// Check the length.
	if s.Len() != 10 {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	// Check the type.
	if s.Type() != meta.StringType {
		t.Errorf("Expected type of StringType, got %s", s.Type().String())
	}

	// Check the data.
	expected = []string{"a", "b", NA_TEXT, "d", "e", NA_TEXT, "g", "h", NA_TEXT, "j"}
	if !utils.CheckEqSliceString(s.Data().([]string), expected, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", expected, s.Data())
	}

	// Check the nullability.
	if !s.IsNullable() {
		t.Errorf("Expected IsNullable() to be true, got false")
	}

	// Check the null mask.
	if !utils.CheckEqSliceBool(s.NullMask(), mask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", mask, s.NullMask())
	}

	// Check the null values.
	for i := range s.Data().([]string) {
		if s.IsNull(i) != mask[i] {
			t.Errorf("Expected IsNull(%d) to be %t, got %t", i, mask[i], s.IsNull(i))
		}
	}

	// Check the null count.
	if s.NullCount() != 3 {
		t.Errorf("Expected NullCount() to be 3, got %d", s.NullCount())
	}

	// Check the HasNull method.
	if !s.HasNull() {
		t.Errorf("Expected HasNull() to be true, got false")
	}

	// Check the Set() with null values.
	for i := range s.Data().([]string) {
		s.Set(i, nil)
	}

	// Check the null values.
	for i := range s.Data().([]string) {
		if !s.IsNull(i) {
			t.Errorf("Expected IsNull(%d) to be true, got false", i)
		}
	}

	// Check the null count.
	if s.NullCount() != 10 {
		t.Errorf("Expected NullCount() to be 10, got %d", s.NullCount())
	}

	// Check the Get method.
	for i, v := range s.Data().([]string) {
		if s.Get(i) != v {
			t.Errorf("Expected Get(%d) to be %s, got %s", i, v, s.Get(i))
		}
	}

	s = NewSeriesString(data, mask, true, ctx)

	// Check the Set method.
	for i, v := range s.Data().([]string) {
		s.Set(i, v+"x")
	}

	// Check the data.
	expected = []string{"ax", "bx", NA_TEXT + "x", "dx", "ex", NA_TEXT + "x", "gx", "hx", NA_TEXT + "x", "jx"}
	if !utils.CheckEqSliceString(s.Data().([]string), expected, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", expected, s.Data())
	}

	// Check the MakeNullable() method.
	p := NewSeriesString(data, nil, true, ctx)

	// Check the nullability.
	if p.IsNullable() {
		t.Errorf("Expected IsNullable() to be false, got true")
	}

	// Set values to null.
	p.Set(1, nil)
	p.Set(3, nil)
	p.Set(5, nil)

	// Check the null count.
	if p.NullCount() != 0 {
		t.Errorf("Expected NullCount() to be 0, got %d", p.NullCount())
	}

	// Make the series nullable.
	p = p.MakeNullable().(Strings)

	// Check the nullability.
	if !p.IsNullable() {
		t.Errorf("Expected IsNullable() to be true, got false")
	}

	// Check the null count.
	if p.NullCount() != 0 {
		t.Errorf("Expected NullCount() to be 0, got %d", p.NullCount())
	}

	p.Set(1, nil)
	p.Set(3, nil)
	p.Set(5, nil)

	// Check the null count.
	if p.NullCount() != 3 {
		t.Errorf("Expected NullCount() to be 3, got %d", p.NullCount())
	}
}

func Test_SeriesString_Append(t *testing.T) {
	var dataA, dataB, dataC, expected []string

	dataA = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	dataB = []string{"k", "l", "m", "n", "o", "p", "q", "r", "s", "t"}
	dataC = []string{"u", "v", "w", "x", "y", "z", "1", "2", "3", "4"}

	maskA := []bool{false, false, true, false, false, true, false, false, true, false}
	maskB := []bool{false, false, false, false, true, false, true, false, false, true}
	maskC := []bool{true, true, true, true, true, true, true, true, true, true}

	// Create two new series.
	sA := NewSeriesString(dataA, maskA, true, ctx)
	sB := NewSeriesString(dataB, maskB, true, ctx)
	sC := NewSeriesString(dataC, maskC, true, ctx)

	// Append the series.
	result := sA.Append(sB).Append(sC)

	// Check the length.
	if result.Len() != 30 {
		t.Errorf("Expected length of 30, got %d", result.Len())
	}

	// Check the data.
	expected = []string{
		"a", "b", NA_TEXT, "d", "e", NA_TEXT, "g", "h", NA_TEXT, "j",
		"k", "l", "m", "n", NA_TEXT, "p", NA_TEXT, "r", "s", NA_TEXT,
		NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT,
	}
	if !utils.CheckEqSliceString(result.Data().([]string), expected, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", expected, result.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSliceBool(result.NullMask(), append(maskA, append(maskB, maskC...)...), nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", append(maskA, append(maskB, maskC...)...), result.NullMask())
	}

	// Append random values.
	s := NewSeriesString([]string{}, nil, true, ctx)

	alpha := "abcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < 100; i++ {
		r := string(alpha[rand.Intn(len(alpha))])
		switch i % 4 {
		case 0:
			s = s.Append(r).(Strings)
		case 1:
			s = s.Append([]string{r}).(Strings)
		case 2:
			s = s.Append(enchanter.NullableString{Valid: true, Value: r}).(Strings)
		case 3:
			s = s.Append([]enchanter.NullableString{{Valid: false, Value: r}}).(Strings)
		}

		if s.Get(i) != r {
			t.Errorf("Expected %t, got %t at index %d (case %d)", true, s.Get(i), i, i%4)
		}
	}

	// Append nil
	s = NewSeriesString([]string{}, nil, true, ctx)

	for i := 0; i < 100; i++ {
		s = s.Append(nil).(Strings)
		if !s.IsNull(i) {
			t.Errorf("Expected %t, got %t at index %d", true, s.IsNull(i), i)
		}
	}

	// Append NAs
	s = NewSeriesString([]string{}, nil, true, ctx)
	na := NewSeriesNA(10, ctx)

	for i := 0; i < 100; i++ {
		s = s.Append(na).(Strings)
		if !utils.CheckEqSlice(s.NullMask()[s.Len()-10:], na.NullMask(), nil, "Strings.Append") {
			t.Errorf("Expected %v, got %v at index %d", na.NullMask(), s.NullMask()[s.Len()-10:], i)
		}
	}

	// Append NullableString
	s = NewSeriesString([]string{}, nil, true, ctx)

	for i := 0; i < 100; i++ {
		s = s.Append(enchanter.NullableString{Valid: false, Value: "a"}).(Strings)
		if !s.IsNull(i) {
			t.Errorf("Expected %t, got %t at index %d", true, s.IsNull(i), i)
		}
	}

	// Append []NullableString
	s = NewSeriesString([]string{}, nil, true, ctx)

	for i := 0; i < 100; i++ {
		s = s.Append([]enchanter.NullableString{{Valid: false, Value: "a"}}).(Strings)
		if !s.IsNull(i) {
			t.Errorf("Expected %t, got %t at index %d", true, s.IsNull(i), i)
		}
	}

	// Append Strings
	s = NewSeriesString([]string{}, nil, true, ctx)
	b := NewSeriesString(dataA, []bool{true, true, true, true, true, true, true, true, true, true}, true, ctx)

	for i := 0; i < 100; i++ {
		s = s.Append(b).(Strings)
		if !utils.CheckEqSlice(s.NullMask()[s.Len()-10:], b.NullMask(), nil, "Strings.Append") {
			t.Errorf("Expected %v, got %v at index %d", b.NullMask(), s.NullMask()[s.Len()-10:], i)
		}
	}
}

func Test_SeriesString_Cast(t *testing.T) {
	data := []string{"true", "false", "0", "3", "4", "5", "hello", "7", "8", "world"}
	mask := []bool{false, false, true, false, false, true, false, false, true, false}

	// Create a new series.
	s := NewSeriesString(data, mask, true, ctx)

	// Cast to bool.
	resBool := s.Cast(meta.BoolType)
	expectedMask := []bool{false, false, true, true, true, true, true, true, true, true}

	// Check the data.
	if !utils.CheckEqSlice(resBool.Data().([]bool), []bool{true, false, false, false, false, false, false, false, false, false}, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", []bool{true, false, false, false, false, false, false, false, false, false}, resBool.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSlice(resBool.NullMask(), expectedMask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", mask, resBool.NullMask())
	}

	// Cast to int.
	resInt := s.Cast(meta.IntType)
	expectedInt := []int{0, 0, 0, 3, 4, 0, 0, 7, 0, 0}
	expectedMask = []bool{true, true, true, false, false, true, true, false, true, true}

	// Check the data.
	if !utils.CheckEqSlice(resInt.Data().([]int), expectedInt, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", expectedInt, resInt.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSlice(resInt.NullMask(), expectedMask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", expectedMask, resInt.NullMask())
	}

	// Cast to int64.
	resInt64 := s.Cast(meta.Int64Type)
	expectedInt64 := []int64{0, 0, 0, 3, 4, 0, 0, 7, 0, 0}

	// Check the data.
	if !utils.CheckEqSlice(resInt64.Data().([]int64), expectedInt64, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", expectedInt64, resInt64.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSlice(resInt64.NullMask(), expectedMask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", expectedMask, resInt64.NullMask())
	}

	// Cast to float64.
	resFloat64 := s.Cast(meta.Float64Type)
	expectedFloat64 := []float64{0, 0, 0, 3, 4, 0, 0, 7, 0, 0}

	// Check the data.
	if !utils.CheckEqSlice(resFloat64.Data().([]float64), expectedFloat64, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", expectedFloat64, resFloat64.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSlice(resFloat64.NullMask(), expectedMask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", expectedMask, resFloat64.NullMask())
	}

	// Cast to string.
	resString := s.Cast(meta.StringType)
	expectedString := []string{"true", "false", NA_TEXT, "3", "4", NA_TEXT, "hello", "7", NA_TEXT, "world"}

	// Check the data.
	if !utils.CheckEqSlice(resString.Data().([]string), expectedString, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", expectedString, resString.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSlice(resString.NullMask(), mask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", mask, resString.NullMask())
	}

	// Cast to time.
	if s.Cast(meta.TimeType).(Errors).Err().Error() != "Strings.Cast: cannot cast to Time, use Strings.ParseTime(layout) instead" {
		t.Errorf("Expected error, got %v", s.Cast(meta.TimeType))
	}

	// Cast to error.
	castError := s.Cast(meta.ErrorType)

	// Check the message.
	if castError.(Errors).msg != "Strings.Cast: invalid type Error" {
		t.Errorf("Expected error, got %v", castError)
	}

	// Parse time.
	s = NewSeriesString([]string{"2019-01-01", "2019-01-02", "2019-01-03", "2019-01-04"}, nil, true, ctx)
	expectedTime := []time.Time{
		time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 1, 4, 0, 0, 0, 0, time.UTC),
	}

	if !utils.CheckEqSlice(s.ParseTime("2006-01-02").Data().([]time.Time), expectedTime, nil, "Strings.ParseTime") {
		t.Errorf("Expected data of %v, got %v", expectedTime, s.ParseTime("2006-01-02").Data())
	}
}

func Test_SeriesString_Filter(t *testing.T) {
	data := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t"}
	mask := []bool{false, true, false, false, true, false, false, true, false, false, true, false, false, true, false, false, true, false, false, true}

	// Create a new series.
	s := NewSeriesString(data, mask, true, ctx)

	// Filter mask.
	filterMask := []bool{true, false, true, true, false, true, true, false, true, true, true, false, true, true, false, true, true, false, true, true}
	filterIndeces := []int{0, 2, 3, 5, 6, 8, 9, 10, 12, 13, 15, 16, 18, 19}

	result := []string{"a", "c", "d", "f", "g", "i", "j", NA_TEXT, "m", NA_TEXT, "p", NA_TEXT, "s", NA_TEXT}
	resultMask := []bool{false, false, false, false, false, false, false, true, false, true, false, true, false, true}

	/////////////////////////////////////////////////////////////////////////////////////
	// 							Check the Filter() method.
	filtered := s.Filter(NewSeriesBool(filterMask, nil, true, ctx))

	// Check the length.
	if filtered.Len() != len(result) {
		t.Errorf("Expected length of %d, got %d", len(result), filtered.Len())
	}

	// Check the data.
	if !utils.CheckEqSliceString(filtered.Data().([]string), result, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", result, filtered.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSliceBool(filtered.NullMask(), resultMask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", resultMask, filtered.NullMask())
	}

	/////////////////////////////////////////////////////////////////////////////////////
	// 							Check the Filter() method.
	filtered = s.Filter(filterMask)

	// Check the length.
	if filtered.Len() != len(result) {
		t.Errorf("Expected length of %d, got %d", len(result), filtered.Len())
	}

	// Check the data.
	if !utils.CheckEqSliceString(filtered.Data().([]string), result, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", result, filtered.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSliceBool(filtered.NullMask(), resultMask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", resultMask, filtered.NullMask())
	}

	/////////////////////////////////////////////////////////////////////////////////////
	// 							Check the Filter() method.
	filtered = s.Filter(filterIndeces)

	// Check the length.
	if filtered.Len() != len(result) {
		t.Errorf("Expected length of %d, got %d", len(result), filtered.Len())
	}

	// Check the data.
	if !utils.CheckEqSliceString(filtered.Data().([]string), result, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", result, filtered.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSliceBool(filtered.NullMask(), resultMask, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", resultMask, filtered.NullMask())
	}

	/////////////////////////////////////////////////////////////////////////////////////

	// try to filter by a series with a different length.
	filtered = filtered.Filter(filterMask)

	if e, ok := filtered.(Errors); !ok || e.Err().Error() != "Strings.Filter: mask length (20) does not match series length (14)" {
		t.Errorf("Expected Errors, got %v", filtered)
	}

	// Another test.
	data = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w"}
	mask = []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}

	// Create a new series.
	s = NewSeriesString(data, mask, true, ctx)

	// Filter mask.
	filterMask = []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, true}
	filterIndeces = []int{0, 15, 22}

	result = []string{NA_TEXT, NA_TEXT, NA_TEXT}

	/////////////////////////////////////////////////////////////////////////////////////
	// 							Check the Filter() method.
	filtered = s.Filter(filterMask)

	// Check the length.
	if filtered.Len() != 3 {
		t.Errorf("Expected length of 3, got %d", filtered.Len())
	}

	// Check the data.
	if !utils.CheckEqSliceString(filtered.Data().([]string), result, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", result, filtered.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSliceBool(filtered.NullMask(), []bool{true, true, true}, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", []bool{true, true, true}, filtered.NullMask())
	}

	/////////////////////////////////////////////////////////////////////////////////////
	// 							Check the Filter() method.
	filtered = s.Filter(filterIndeces)

	// Check the length.
	if filtered.Len() != 3 {
		t.Errorf("Expected length of 3, got %d", filtered.Len())
	}

	// Check the data.
	if !utils.CheckEqSliceString(filtered.Data().([]string), result, nil, "Strings.Data") {
		t.Errorf("Expected data of %v, got %v", result, filtered.Data())
	}

	// Check the null mask.
	if !utils.CheckEqSliceBool(filtered.NullMask(), []bool{true, true, true}, nil, "Strings.GetNullMask") {
		t.Errorf("Expected null mask of %v, got %v", []bool{true, true, true}, filtered.NullMask())
	}
}

func Test_SeriesString_Map(t *testing.T) {
	data := []string{"", "hello", "world", "this", "is", "a", "test", "of", "the", "map", "function", "in", "the", "series", "", "this", "is", "a", "", "test"}
	nullMask := []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, true, true, false, false, true}

	// Create a new series.
	s := NewSeriesString(data, nullMask, true, ctx)

	// Map the series to bool.
	resBool := s.Map(func(v any) any {
		return v.(string) == ""
	})

	expectedBool := []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, true, false}
	if !utils.CheckEqSliceBool(resBool.Data().([]bool), expectedBool, nil, "Strings.Map") {
		t.Errorf("Expected data of %v, got %v", expectedBool, resBool.Data())
	}

	// Map the series to int.
	resInt := s.Map(func(v any) any {
		return int(len(v.(string)))
	})

	expectedInt := []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 6, 0, 2, 2, 1, 0, 2}
	if !utils.CheckEqSliceInt(resInt.Data().([]int), expectedInt, nil, "Strings.Map") {
		t.Errorf("Expected data of %v, got %v", expectedInt, resInt.Data())
	}

	// Map the series to int64.
	resInt64 := s.Map(func(v any) any {
		return int64(len(v.(string)))
	})

	expectedInt64 := []int64{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 6, 0, 2, 2, 1, 0, 2}
	if !utils.CheckEqSliceInt64(resInt64.Data().([]int64), expectedInt64, nil, "Strings.Map") {
		t.Errorf("Expected data of %v, got %v", expectedInt64, resInt64.Data())
	}

	// Map the series to float64.
	resFloat64 := s.Map(func(v any) any {
		return -float64(len(v.(string)))
	})

	expectedFloat64 := []float64{-2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -6, -0, -2, -2, -1, -0, -2}
	if !utils.CheckEqSliceFloat64(resFloat64.Data().([]float64), expectedFloat64, nil, "Strings.Map") {
		t.Errorf("Expected data of %v, got %v", expectedFloat64, resFloat64.Data())
	}

	// Map the series to string.
	resString := s.Map(func(v any) any {
		if v.(string) == "" {
			return "empty"
		}
		return ""
	})

	expectedString := []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "empty", "", "", "", "empty", ""}
	if !utils.CheckEqSliceString(resString.Data().([]string), expectedString, nil, "Strings.Map") {
		t.Errorf("Expected data of %v, got %v", expectedString, resString.Data())
	}
}

func Test_SeriesString_Group(t *testing.T) {
	var partMap map[int64][]int

	data1 := []string{"foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo"}
	data1Mask := []bool{false, false, false, false, false, true, true, true, true, true}
	data2 := []string{"foo", "foo", "bar", "bar", "foo", "foo", "bar", "bar", "foo", "foo"}
	data2Mask := []bool{false, true, false, true, false, true, false, true, false, true}
	data3 := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	data3Mask := []bool{false, false, false, false, false, true, true, true, true, true}

	// Test 1
	s1 := NewSeriesString(data1, data1Mask, true, ctx).
		Group()

	p1 := s1.Partition().GetMap()
	if len(p1) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(p1))
	}

	partMap = map[int64][]int{
		0: {0, 1, 2, 3, 4},
		1: {5, 6, 7, 8, 9},
	}
	if !utils.CheckEqPartitionMap(p1, partMap, nil, "String Group") {
		t.Errorf("Expected partition map of %v, got %v", partMap, p1)
	}

	// Test 2
	s2 := NewSeriesString(data2, data2Mask, true, ctx).
		GroupBy(s1.Partition())

	p2 := s2.Partition().GetMap()
	if len(p2) != 6 {
		t.Errorf("Expected 6 groups, got %d", len(p2))
	}

	partMap = map[int64][]int{
		0: {0, 4},
		1: {1, 3},
		2: {2},
		3: {5, 7, 9},
		4: {6},
		5: {8},
	}
	if !utils.CheckEqPartitionMap(p2, partMap, nil, "String Group") {
		t.Errorf("Expected partition map of %v, got %v", partMap, p2)
	}

	// Test 3
	s3 := NewSeriesString(data3, data3Mask, true, ctx).
		GroupBy(s2.Partition())

	p3 := s3.Partition().GetMap()
	if len(p3) != 8 {
		t.Errorf("Expected 8 groups, got %d", len(p3))
	}

	partMap = map[int64][]int{
		0: {0},
		1: {1},
		2: {2},
		3: {3},
		4: {4},
		5: {5, 7, 9},
		6: {6},
		7: {8},
	}
	if !utils.CheckEqPartitionMap(p3, partMap, nil, "String Group") {
		t.Errorf("Expected partition map of %v, got %v", partMap, p3)
	}

	// debugPrintPartition(s1.Partition(), s1)
	// debugPrintPartition(s2.Partition(), s1, s2)
	// debugPrintPartition(s3.Partition(), s1, s2, s3)
}

func Test_SeriesString_Sort(t *testing.T) {
	data := []string{"w", "w", "d", "y", "b", "e", "a", "e", "e", "b", "l", "u", "a", "g", "w", "u", "{", "x", "t", "h"}
	mask := []bool{false, true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false, true}

	// Create a new series.
	s := NewSeriesString(data, nil, true, ctx)

	// Sort the series.
	sorted := s.Sort()

	// Check the data.
	expected := []string{"a", "a", "b", "b", "d", "e", "e", "e", "g", "h", "l", "t", "u", "u", "w", "w", "w", "x", "y", "{"}
	if !utils.CheckEqSliceString(sorted.Data().([]string), expected, nil, "") {
		t.Errorf("Strings.Sort() failed, expecting %v, got %v", expected, sorted.Data().([]string))
	}

	// Create a new series.
	s = NewSeriesString(data, mask, true, ctx)

	// Sort the series.
	sorted = s.Sort()

	// Check the data.
	expected = []string{"a", "a", "b", "d", "e", "l", "t", "w", "w", "{", NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT, NA_TEXT}
	if !utils.CheckEqSliceString(sorted.Data().([]string), expected, nil, "") {
		t.Errorf("Strings.Sort() failed, expecting %v, got %v", expected, sorted.Data().([]string))
	}

	// Check the null mask.
	expectedMask := []bool{false, false, false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true, true, true}
	if !utils.CheckEqSliceBool(sorted.NullMask(), expectedMask, nil, "") {
		t.Errorf("Strings.Sort() failed, expecting %v, got %v", expectedMask, sorted.NullMask())
	}
}

func Test_SeriesString_Arithmetic_Add(t *testing.T) {
	nas := NewSeriesNA(1, ctx)
	nav := NewSeriesNA(10, ctx)

	bools := NewSeriesBool([]bool{true}, nil, true, ctx)
	boolv := NewSeriesBool([]bool{true, false, true, false, true, false, true, true, false, false}, nil, true, ctx)
	bools_ := NewSeriesBool([]bool{true}, nil, true, ctx).SetNullMask([]bool{true})
	boolv_ := NewSeriesBool([]bool{true, false, true, false, true, false, true, true, false, false}, nil, true, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	i32s := NewSeriesInt([]int{2}, nil, true, ctx)
	i32v := NewSeriesInt([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, true, ctx)
	i32s_ := NewSeriesInt([]int{2}, nil, true, ctx).SetNullMask([]bool{true})
	i32v_ := NewSeriesInt([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, true, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	i64s := NewSeriesInt64([]int64{2}, nil, true, ctx)
	i64v := NewSeriesInt64([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, true, ctx)
	i64s_ := NewSeriesInt64([]int64{2}, nil, true, ctx).SetNullMask([]bool{true})
	i64v_ := NewSeriesInt64([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, true, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	f64s := NewSeriesFloat64([]float64{2}, nil, true, ctx)
	f64v := NewSeriesFloat64([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, true, ctx)
	f64s_ := NewSeriesFloat64([]float64{2}, nil, true, ctx).SetNullMask([]bool{true})
	f64v_ := NewSeriesFloat64([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, true, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	ss := NewSeriesString([]string{"2"}, nil, true, ctx)
	sv := NewSeriesString([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, nil, true, ctx)
	ss_ := NewSeriesString([]string{"2"}, nil, true, ctx).SetNullMask([]bool{true})
	sv_ := NewSeriesString([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, nil, true, ctx).
		SetNullMask([]bool{false, true, false, true, false, true, false, true, false, true})

	// scalar | NA: NA propagation, the result is all null
	if res, ok := ss.Add(nas).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := ss.Add(nav).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := ss_.Add(nas).(NAs); !ok || res.Len() != 1 {
		t.Errorf("Expected NAs of length 1, got %v", res)
	}
	if res, ok := ss_.Add(nav).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}

	// scalar | bool
	if !utils.CheckEqSlice(ss.Add(bools).Data().([]string), []string{"2true"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"2true"}, ss.Add(bools).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(boolv).Data().([]string), []string{"2true", "2false", "2true", "2false", "2true", "2false", "2true", "2true", "2false", "2false"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"2true", "2false", "2true", "2false", "2true", "2false", "2true", "2true", "2false", "2false"}, ss.Add(boolv).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(bools_).NullMask(), []bool{true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, ss.Add(bools_).NullMask())
	}
	if !utils.CheckEqSlice(ss.Add(boolv_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, ss.Add(boolv_).NullMask())
	}

	// scalar | int
	if !utils.CheckEqSlice(ss.Add(i32s).Data().([]string), []string{"22"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"22"}, ss.Add(i32s).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(i32v).Data().([]string), []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, ss.Add(i32v).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(i32s_).NullMask(), []bool{true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, ss.Add(i32s_).NullMask())
	}
	if !utils.CheckEqSlice(ss.Add(i32v_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, ss.Add(i32v_).NullMask())
	}

	// scalar | int64
	if !utils.CheckEqSlice(ss.Add(i64s).Data().([]string), []string{"22"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"22"}, ss.Add(i64s).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(i64v).Data().([]string), []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, ss.Add(i64v).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(i64s_).NullMask(), []bool{true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, ss.Add(i64s_).NullMask())
	}
	if !utils.CheckEqSlice(ss.Add(i64v_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, ss.Add(i64v_).NullMask())
	}

	// scalar | float64
	if !utils.CheckEqSlice(ss.Add(f64s).Data().([]string), []string{"22"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"22"}, ss.Add(f64s).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(f64v).Data().([]string), []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, ss.Add(f64v).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(f64s_).NullMask(), []bool{true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, ss.Add(f64s_).NullMask())
	}
	if !utils.CheckEqSlice(ss.Add(f64v_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, ss.Add(f64v_).NullMask())
	}

	// scalar | string
	if !utils.CheckEqSlice(ss.Add(ss).Data().([]string), []string{"22"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"22"}, ss.Add(ss).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(sv).Data().([]string), []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"21", "22", "23", "24", "25", "26", "27", "28", "29", "210"}, ss.Add(sv).Data().([]string))
	}
	if !utils.CheckEqSlice(ss.Add(ss_).NullMask(), []bool{true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true}, ss.Add(ss_).NullMask())
	}
	if !utils.CheckEqSlice(ss.Add(sv_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, ss.Add(sv_).NullMask())
	}

	// vector | NA: NA propagation, the result is all null
	if res, ok := sv.Add(nas).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := sv.Add(nav).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := sv_.Add(nas).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}
	if res, ok := sv_.Add(nav).(NAs); !ok || res.Len() != 10 {
		t.Errorf("Expected NAs of length 10, got %v", res)
	}

	// vector | bool
	if !utils.CheckEqSlice(sv.Add(bools).Data().([]string), []string{"1true", "2true", "3true", "4true", "5true", "6true", "7true", "8true", "9true", "10true"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"1true", "2true", "3true", "4true", "5true", "6true", "7true", "8true", "9true", "10true"}, sv.Add(bools).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(boolv).Data().([]string), []string{"1true", "2false", "3true", "4false", "5true", "6false", "7true", "8true", "9false", "10false"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"1true", "2false", "3true", "4false", "5true", "6false", "7true", "8true", "9false", "10false"}, sv.Add(boolv).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(bools_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, sv.Add(bools_).NullMask())
	}
	if !utils.CheckEqSlice(sv.Add(boolv_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, sv.Add(boolv_).NullMask())
	}

	// vector | int
	if !utils.CheckEqSlice(sv.Add(i32s).Data().([]string), []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, sv.Add(i32s).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(i32v).Data().([]string), []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, sv.Add(i32v).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(i32s_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, sv.Add(i32s_).NullMask())
	}
	if !utils.CheckEqSlice(sv.Add(i32v_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, sv.Add(i32v_).NullMask())
	}

	// vector | int64
	if !utils.CheckEqSlice(sv.Add(i64s).Data().([]string), []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, sv.Add(i64s).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(i64v).Data().([]string), []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, sv.Add(i64v).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(i64s_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, sv.Add(i64s_).NullMask())
	}
	if !utils.CheckEqSlice(sv.Add(i64v_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, sv.Add(i64v_).NullMask())
	}

	// vector | float64
	if !utils.CheckEqSlice(sv.Add(f64s).Data().([]string), []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, sv.Add(f64s).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(f64v).Data().([]string), []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, sv.Add(f64v).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(f64s_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, sv.Add(f64s_).NullMask())
	}
	if !utils.CheckEqSlice(sv.Add(f64v_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, sv.Add(f64v_).NullMask())
	}

	// vector | string
	if !utils.CheckEqSlice(sv.Add(ss).Data().([]string), []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"12", "22", "32", "42", "52", "62", "72", "82", "92", "102"}, sv.Add(ss).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(sv).Data().([]string), []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []string{"11", "22", "33", "44", "55", "66", "77", "88", "99", "1010"}, sv.Add(sv).Data().([]string))
	}
	if !utils.CheckEqSlice(sv.Add(ss_).NullMask(), []bool{true, true, true, true, true, true, true, true, true, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{true, true, true, true, true, true, true, true, true, true}, sv.Add(ss_).NullMask())
	}
	if !utils.CheckEqSlice(sv.Add(sv_).NullMask(), []bool{false, true, false, true, false, true, false, true, false, true}, nil, "String Add") {
		t.Errorf("Expected %v, got %v", []bool{false, true, false, true, false, true, false, true, false, true}, sv.Add(sv_).NullMask())
	}
}

func Test_SeriesString_Logical_Eq(t *testing.T) {
	ss := NewSeriesString([]string{"1"}, nil, true, ctx)
	ss_ := NewSeriesString([]string{"1"}, nil, true, ctx).SetNullMask([]bool{true})
	sv := NewSeriesString([]string{"1", "2", "3"}, nil, true, ctx)
	sv_ := NewSeriesString([]string{"1", "2", "3"}, nil, true, ctx).SetNullMask([]bool{true, true, false})

	// scalar | scalar
	res := ss.Eq(ss)
	if res.Data().([]bool)[0] != true {
		t.Errorf("Expected %v, got %v", true, res.Data().([]bool)[0])
	}

	res = ss.Eq(ss_)
	if res.IsNull(0) == false {
		t.Errorf("Expected %v, got %v", true, res.IsNull(0))
	}

	// scalar | vector
	res = ss.Eq(sv)
	if res.Data().([]bool)[0] != true || res.Data().([]bool)[1] != false || res.Data().([]bool)[2] != false {
		t.Errorf("Expected %v, got %v", []bool{true, false, false}, res.Data().([]bool))
	}

	res = ss.Eq(sv_)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{true, true, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	// vector | scalar
	res = sv.Eq(ss)
	if res.Data().([]bool)[0] != true || res.Data().([]bool)[1] != false || res.Data().([]bool)[2] != false {
		t.Errorf("Expected %v, got %v", []bool{true, false, false}, res.Data().([]bool))
	}

	res = sv_.Eq(ss)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{true, true, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	// vector | vector
	res = sv.Eq(sv)
	if res.Data().([]bool)[0] != true || res.Data().([]bool)[1] != true || res.Data().([]bool)[2] != true {
		t.Errorf("Expected %v, got %v", []bool{true, true, true}, res.Data().([]bool))
	}

	res = sv.Eq(sv_)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{true, true, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	res = sv_.Eq(sv)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{true, true, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	res = sv_.Eq(sv_)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{true, true, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}
}

func Test_SeriesString_Logical_Ne(t *testing.T) {
	// TODO: add tests for all types
}

func Test_SeriesString_Logical_Lt(t *testing.T) {
	ss := NewSeriesString([]string{"1"}, nil, true, ctx)
	ss_ := NewSeriesString([]string{"1"}, nil, true, ctx).SetNullMask([]bool{true})
	sv := NewSeriesString([]string{"1", "2", "3"}, nil, true, ctx)
	sv_ := NewSeriesString([]string{"1", "2", "3"}, nil, true, ctx).SetNullMask([]bool{true, true, false})

	// scalar | scalar
	res := ss.Lt(ss)
	if res.Data().([]bool)[0] != false {
		t.Errorf("Expected %v, got %v", false, res.Data().([]bool)[0])
	}

	res = ss.Lt(ss_)
	if res.IsNull(0) == false {
		t.Errorf("Expected %v, got %v", true, res.IsNull(0))
	}

	// scalar | vector
	res = ss.Lt(sv)
	if res.Data().([]bool)[0] != false || res.Data().([]bool)[1] != true || res.Data().([]bool)[2] != true {
		t.Errorf("Expected %v, got %v", []bool{false, true, true}, res.Data().([]bool))
	}

	res = ss.Lt(sv_)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{false, false, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	// vector | scalar
	res = sv.Lt(ss)
	if res.Data().([]bool)[0] != false || res.Data().([]bool)[1] != false || res.Data().([]bool)[2] != false {
		t.Errorf("Expected %v, got %v", []bool{false, false, false}, res.Data().([]bool))
	}

	res = sv_.Lt(ss)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{false, false, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	// vector | vector
	res = sv.Lt(sv)
	if res.Data().([]bool)[0] != false || res.Data().([]bool)[1] != false || res.Data().([]bool)[2] != false {
		t.Errorf("Expected %v, got %v", []bool{false, false, false}, res.Data().([]bool))
	}

	res = sv.Lt(sv_)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{false, false, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	res = sv_.Lt(sv)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{false, false, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}

	res = sv_.Lt(sv_)
	if res.IsNull(0) == false || res.IsNull(1) == false || res.IsNull(2) == true {
		t.Errorf("Expected %v, got %v", []bool{false, false, false}, []bool{res.IsNull(0), res.IsNull(1), res.IsNull(2)})
	}
}

func Test_SeriesString_Logical_Le(t *testing.T) {
	// TODO: add tests for all types
}

func Test_SeriesString_Logical_Gt(t *testing.T) {
	// TODO: add tests for all types
}

func Test_SeriesString_Logical_Ge(t *testing.T) {
	// TODO: add tests for all types
}
