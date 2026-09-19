package dataframe

import (
	"strings"
	"testing"
	"time"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/utils"
)

var testCtx = enchanter.NewContext()

// Test data with null values for comprehensive testing
const testDataWithNulls = `
id,name,department,salary,active
1,Alice,HR,50000,true
2,Bob,IT,60000,true
3,Charlie,,45000,false
4,David,IT,,true
5,Eve,HR,55000,
6,,Finance,70000,false
`

func TestJoin_EmptyDataFrames(t *testing.T) {
	// Test joining with empty dataframes
	dfEmpty := NewDataFrame(testCtx)
	dfNormal := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2}, nil, false).
		AddSeriesFromStrings("name", []string{"Alice", "Bob"}, nil, false)

	// Empty left dataframe
	result := dfEmpty.Join(INNER_JOIN, dfNormal, "id")
	if result.GetError() == nil {
		t.Error("Expected error when joining empty dataframe without common columns")
	}

	// Test with properly structured empty dataframe
	dfEmptyStructured := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{}, nil, false).
		AddSeriesFromStrings("name", []string{}, nil, false)

	result = dfEmptyStructured.Join(INNER_JOIN, dfNormal, "id")
	if result.GetError() != nil {
		t.Error("Should handle empty structured dataframe:", result.GetError())
	}
	if result.NRows() != 0 {
		t.Errorf("Expected 0 rows from empty join, got %d", result.NRows())
	}
}

func TestJoin_MultipleColumns(t *testing.T) {
	df1 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("dept_id", []int64{1, 1, 2, 2, 3}, nil, false).
		AddSeriesFromStrings("category", []string{"A", "B", "A", "B", "A"}, nil, false).
		AddSeriesFromStrings("name", []string{"Alice", "Bob", "Charlie", "David", "Eve"}, nil, false)

	df2 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("dept_id", []int64{1, 1, 2, 4}, nil, false).
		AddSeriesFromStrings("category", []string{"A", "B", "A", "A"}, nil, false).
		AddSeriesFromFloat64s("budget", []float64{10000, 15000, 12000, 8000}, nil, false)

	// Multi-column join
	result := df1.Join(INNER_JOIN, df2, "dept_id", "category")
	if result.GetError() != nil {
		t.Error("Multi-column join failed:", result.GetError())
	}

	expectedRows := 3 // (1,A), (1,B), (2,A)
	if result.NRows() != expectedRows {
		t.Errorf("Expected %d rows, got %d", expectedRows, result.NRows())
	}

	// Verify the results
	names := result.C("name").Data().([]string)
	budgets := result.C("budget").Data().([]float64)
	expectedNames := []string{"Alice", "Bob", "Charlie"}
	expectedBudgets := []float64{10000, 15000, 12000}

	if !utils.CheckEqSliceString(names, expectedNames, nil, "Multi-column join names") {
		t.Errorf("Expected names %v, got %v", expectedNames, names)
	}
	if !utils.CheckEqSliceFloat64(budgets, expectedBudgets, nil, "Multi-column join budgets") {
		t.Errorf("Expected budgets %v, got %v", expectedBudgets, budgets)
	}
}

func TestJoin_WithNullValues(t *testing.T) {
	// Create dataframes with null values
	nullMask1 := []bool{false, false, true, false} // Charlie has null department
	df1 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3, 4}, nil, false).
		AddSeriesFromStrings("department", []string{"HR", "IT", "", "Finance"}, nullMask1, false)

	nullMask2 := []bool{false, false, true, false} // id=3 has null salary
	df2 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3, 5}, nil, false).
		AddSeriesFromFloat64s("salary", []float64{50000, 60000, 0, 45000}, nullMask2, false)

	// Inner join should handle nulls properly
	result := df1.Join(INNER_JOIN, df2, "id")
	if result.GetError() != nil {
		t.Error("Join with nulls failed:", result.GetError())
	}

	if result.NRows() != 3 { // ids 1, 2, 3 should match
		t.Errorf("Expected 3 rows, got %d", result.NRows())
	}
}

func TestJoin_TypeMismatch(t *testing.T) {
	df1 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3}, nil, false)

	df2 := NewDataFrame(testCtx).
		AddSeriesFromStrings("id", []string{"1", "2", "3"}, nil, false) // Different type

	result := df1.Join(INNER_JOIN, df2, "id")
	if result.GetError() == nil {
		t.Error("Expected error when joining columns with different types")
	}
}

func TestJoin_AllJoinTypes(t *testing.T) {
	df1 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3, 4}, nil, false).
		AddSeriesFromStrings("name", []string{"Alice", "Bob", "Charlie", "David"}, nil, false)

	df2 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{2, 3, 4, 5}, nil, false).
		AddSeriesFromFloat64s("salary", []float64{60000, 50000, 55000, 65000}, nil, false)

	// Test all join types
	tests := []struct {
		joinType     DataFrameJoinType
		expectedRows int
		name         string
	}{
		{INNER_JOIN, 3, "Inner Join"}, // ids 2, 3, 4
		{LEFT_JOIN, 4, "Left Join"},   // all from df1
		{RIGHT_JOIN, 4, "Right Join"}, // all from df2
		{OUTER_JOIN, 5, "Outer Join"}, // ids 1, 2, 3, 4, 5
	}

	for _, test := range tests {
		result := df1.Join(test.joinType, df2, "id")
		if result.GetError() != nil {
			t.Errorf("%s failed: %v", test.name, result.GetError())
			continue
		}
		if result.NRows() != test.expectedRows {
			t.Errorf("%s: expected %d rows, got %d", test.name, test.expectedRows, result.NRows())
		}
	}
}

func TestJoin_ColumnNameCollisions(t *testing.T) {
	df1 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3}, nil, false).
		AddSeriesFromStrings("value", []string{"A", "B", "C"}, nil, false)

	df2 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3}, nil, false).
		AddSeriesFromStrings("value", []string{"X", "Y", "Z"}, nil, false) // Same column name

	result := df1.Join(INNER_JOIN, df2, "id")
	if result.GetError() != nil {
		t.Error("Join with column name collision failed:", result.GetError())
	}

	// Should have value_x and value_y columns
	names := result.Names()
	expectedNames := []string{"id", "value_x", "value_y"}
	if !utils.CheckEqSliceString(names, expectedNames, nil, "Column collision names") {
		t.Errorf("Expected column names %v, got %v", expectedNames, names)
	}
}

func TestGroupBy_WithAggregations(t *testing.T) {
	// Use the working test data from the existing tests
	testData := `
name,age,weight,junior,department,salary band
Alice C,29,75.0,F,HR,4
John Doe,30,80.5,true,IT,2
Bob,31,85.0,T,IT,4
Jane H,25,60.0,false,IT,4
Mary,28,70.0,false,IT,3
Oliver,32,90.0,true,HR,1
Ursula,27,65.0,f,Business,4
Charlie,33,60.0,t,Business,2
`

	df := NewDataFrame(testCtx).FromCsv().
		SetReader(strings.NewReader(testData)).
		SetDelimiter(',').
		SetHeader(true).
		SetGuessDataTypeLen(3).
		Read()

	if df.GetError() != nil {
		t.Error("Failed to create test dataframe:", df.GetError())
	}

	// Test all aggregation functions including Min and Max
	result := df.GroupBy("department").
		Agg(Count(), Sum("age"), Mean("age"), Min("age"), Max("age"), Std("age")).
		Run()

	if result.GetError() != nil {
		t.Error("GroupBy with all aggregations failed:", result.GetError())
	}

	// Verify we have the expected columns
	expectedCols := []string{"department", "n", "sum(age)", "mean(age)", "min(age)", "max(age)", "std(age)"}
	actualCols := result.Names()
	if !utils.CheckEqSliceString(actualCols, expectedCols, nil, "Aggregation columns") {
		t.Errorf("Expected columns %v, got %v", expectedCols, actualCols)
	}

	// Test specific values for IT department
	for i := 0; i < result.NRows(); i++ {
		dept := result.C("department").Get(i).(string)
		if dept == "IT" {
			count := result.C("n").Get(i).(int64)
			minAge := result.C("min(age)").Get(i).(float64)
			maxAge := result.C("max(age)").Get(i).(float64)

			if count != 4 {
				t.Errorf("IT department: expected count 4, got %d", count)
			}
			if minAge != 25.0 {
				t.Errorf("IT department: expected min age 25.0, got %f", minAge)
			}
			if maxAge != 31.0 {
				t.Errorf("IT department: expected max age 31.0, got %f", maxAge)
			}
		}
	}
}

func TestGroupBy_WithNullValues(t *testing.T) {
	// Create dataframe with null values
	nullMaskSalary := []bool{false, false, true, false, false} // Charlie has null salary
	nullMaskDept := []bool{false, false, false, true, false}   // David has null department

	df := NewDataFrame(testCtx).
		AddSeriesFromStrings("name", []string{"Alice", "Bob", "Charlie", "David", "Eve"}, nil, false).
		AddSeriesFromStrings("department", []string{"HR", "IT", "IT", "", "HR"}, nullMaskDept, false).
		AddSeriesFromFloat64s("salary", []float64{50000, 60000, 0, 55000, 52000}, nullMaskSalary, false)

	// Group by department with null handling
	result := df.GroupBy("department").Agg(Count(), Sum("salary"), Mean("salary")).Run()
	if result.GetError() != nil {
		t.Error("GroupBy with nulls failed:", result.GetError())
	}

	// Test with RemoveNAs
	resultNoNAs := df.GroupBy("department").Agg(Count(), Sum("salary"), Mean("salary")).RemoveNAs(true).Run()
	if resultNoNAs.GetError() != nil {
		t.Error("GroupBy with RemoveNAs failed:", resultNoNAs.GetError())
	}

	// Verify the results handle nulls appropriately
	if result.NRows() <= 0 {
		t.Error("Expected non-empty result from groupby with nulls")
	}
}

func TestGroupBy_MultipleColumns(t *testing.T) {
	df := NewDataFrame(testCtx).
		AddSeriesFromStrings("department", []string{"IT", "IT", "HR", "HR", "IT", "HR"}, nil, false).
		AddSeriesFromBools("senior", []bool{true, false, true, false, true, true}, nil, false).
		AddSeriesFromFloat64s("salary", []float64{70000, 50000, 65000, 45000, 75000, 68000}, nil, false)

	// Multi-column grouping
	result := df.GroupBy("department", "senior").Agg(Count(), Mean("salary")).Run()
	if result.GetError() != nil {
		t.Error("Multi-column GroupBy failed:", result.GetError())
	}

	expectedRows := 4 // (IT,true), (IT,false), (HR,true), (HR,false)
	if result.NRows() != expectedRows {
		t.Errorf("Expected %d groups, got %d", expectedRows, result.NRows())
	}

	// Verify column structure
	expectedCols := []string{"department", "senior", "n", "mean(salary)"}
	actualCols := result.Names()
	if !utils.CheckEqSliceString(actualCols, expectedCols, nil, "Multi-column groupby columns") {
		t.Errorf("Expected columns %v, got %v", expectedCols, actualCols)
	}
}

func TestGroupBy_EmptyGroups(t *testing.T) {
	// Test with a dataframe that could result in empty groups
	df := NewDataFrame(testCtx).
		AddSeriesFromStrings("category", []string{"A", "A", "B", "B"}, nil, false).
		AddSeriesFromInt64s("value", []int64{1, 2, 3, 4}, nil, false)

	result := df.GroupBy("category").Agg(Count(), Sum("value")).Run()
	if result.GetError() != nil {
		t.Error("GroupBy with potential empty groups failed:", result.GetError())
	}

	if result.NRows() != 2 { // Should have A and B groups
		t.Errorf("Expected 2 groups, got %d", result.NRows())
	}
}

func TestGroupBy_AllDataTypes(t *testing.T) {
	// Test grouping with different data types
	timeData := []time.Time{
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	df := NewDataFrame(testCtx).
		AddSeriesFromBools("active", []bool{true, true, false}, nil, false).
		AddSeriesFromTimes("date", timeData, nil, false).
		AddSeriesFromInt64s("count", []int64{5, 3, 8}, nil, false)

	// Group by boolean
	result1 := df.GroupBy("active").Agg(Count(), Sum("count")).Run()
	if result1.GetError() != nil {
		t.Error("GroupBy boolean failed:", result1.GetError())
	}

	// Group by time
	result2 := df.GroupBy("date").Agg(Count(), Sum("count")).Run()
	if result2.GetError() != nil {
		t.Error("GroupBy time failed:", result2.GetError())
	}
}

func TestJoinThenGroupBy(t *testing.T) {
	// Test combination: join then group by
	df1 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("emp_id", []int64{1, 2, 3, 4}, nil, false).
		AddSeriesFromStrings("name", []string{"Alice", "Bob", "Charlie", "David"}, nil, false).
		AddSeriesFromStrings("department", []string{"HR", "IT", "IT", "HR"}, nil, false)

	df2 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("emp_id", []int64{1, 2, 3, 4, 5}, nil, false).
		AddSeriesFromFloat64s("salary", []float64{50000, 60000, 55000, 52000, 65000}, nil, false)

	// Join first
	joined := df1.Join(INNER_JOIN, df2, "emp_id")
	if joined.GetError() != nil {
		t.Error("Join failed:", joined.GetError())
	}

	// Then group by
	result := joined.GroupBy("department").Agg(Count(), Mean("salary")).Run()
	if result.GetError() != nil {
		t.Error("GroupBy after join failed:", result.GetError())
	}

	if result.NRows() != 2 { // HR and IT departments
		t.Errorf("Expected 2 departments, got %d", result.NRows())
	}
}

func TestGroupByThenJoin(t *testing.T) {
	// Test combination: group by then join (should fail as grouped dataframes can't be joined)
	df1 := NewDataFrame(testCtx).
		AddSeriesFromStrings("department", []string{"HR", "IT", "IT", "HR"}, nil, false).
		AddSeriesFromFloat64s("salary", []float64{50000, 60000, 55000, 52000}, nil, false)

	df2 := NewDataFrame(testCtx).
		AddSeriesFromStrings("department", []string{"HR", "IT", "Finance"}, nil, false).
		AddSeriesFromFloat64s("budget", []float64{100000, 150000, 80000}, nil, false)

	// Group first
	grouped := df1.GroupBy("department").Agg(Mean("salary")).Run()
	if grouped.GetError() != nil {
		t.Error("GroupBy failed:", grouped.GetError())
	}

	// Try to join grouped dataframe - this should work now since grouped result is ungrouped
	result := grouped.Join(INNER_JOIN, df2, "department")
	if result.GetError() != nil {
		t.Error("Join after GroupBy failed:", result.GetError())
	}
}

func TestErrorConditions(t *testing.T) {
	df1 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3}, nil, false)

	df2 := NewDataFrame(testCtx).
		AddSeriesFromInt64s("id", []int64{1, 2, 3}, nil, false)

	// Test joining non-existent columns
	result := df1.Join(INNER_JOIN, df2, "nonexistent")
	if result.GetError() == nil {
		t.Error("Expected error when joining on non-existent column")
	}

	// Test grouping by non-existent column
	result2 := df1.GroupBy("nonexistent").Agg(Count()).Run()
	if result2.GetError() == nil {
		t.Error("Expected error when grouping by non-existent column")
	}

	// Test aggregating non-existent column
	result3 := df1.GroupBy("id").Agg(Sum("nonexistent")).Run()
	if result3.GetError() == nil {
		t.Error("Expected error when aggregating non-existent column")
	}
}
