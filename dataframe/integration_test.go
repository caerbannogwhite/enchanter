package dataframe

import (
	"testing"

	"github.com/caerbannogwhite/enchanter"
)

var integrationCtx = enchanter.NewContext()

func TestIntegration_ComplexJoinAndGroupBy(t *testing.T) {
	// Create employee dataframe
	employees := NewDataFrame(integrationCtx).
		AddSeriesFromInt64s("emp_id", []int64{1, 2, 3, 4, 5, 6}, nil, false).
		AddSeriesFromStrings("name", []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank"}, nil, false).
		AddSeriesFromStrings("department", []string{"HR", "IT", "IT", "HR", "Finance", "IT"}, nil, false).
		AddSeriesFromInt64s("manager_id", []int64{0, 1, 2, 1, 0, 2}, nil, false) // 0 means no manager

	// Create salary dataframe
	salaries := NewDataFrame(integrationCtx).
		AddSeriesFromInt64s("emp_id", []int64{1, 2, 3, 4, 5, 6, 7}, nil, false). // emp_id 7 doesn't exist in employees
		AddSeriesFromFloat64s("salary", []float64{75000, 65000, 60000, 70000, 80000, 55000, 50000}, nil, false).
		AddSeriesFromInt64s("year", []int64{2023, 2023, 2023, 2023, 2023, 2023, 2023}, nil, false)

	// Create bonus dataframe
	bonuses := NewDataFrame(integrationCtx).
		AddSeriesFromInt64s("emp_id", []int64{1, 2, 4, 5}, nil, false). // Not all employees have bonuses
		AddSeriesFromFloat64s("bonus", []float64{5000, 3000, 4000, 6000}, nil, false)

	// Complex scenario: Join employees with salaries (inner join)
	empSalaries := employees.Join(INNER_JOIN, salaries, "emp_id")
	if empSalaries.GetError() != nil {
		t.Fatal("Employee-Salary join failed:", empSalaries.GetError())
	}

	// Then join with bonuses (left join to include employees without bonuses)
	fullData := empSalaries.Join(LEFT_JOIN, bonuses, "emp_id")
	if fullData.GetError() != nil {
		t.Fatal("Full data join failed:", fullData.GetError())
	}

	// Now perform complex grouping: Group by department and calculate various metrics
	departmentStats := fullData.GroupBy("department").
		Agg(
			Count(),        // Number of employees
			Mean("salary"), // Average salary
			Sum("salary"),  // Total salary cost
			Min("salary"),  // Minimum salary
			Max("salary"),  // Maximum salary
			Std("salary"),  // Salary standard deviation
		).Run()

	if departmentStats.GetError() != nil {
		t.Fatal("Department statistics failed:", departmentStats.GetError())
	}

	// Verify results
	expectedDepartments := 3 // HR, IT, Finance
	if departmentStats.NRows() != expectedDepartments {
		t.Errorf("Expected %d departments, got %d", expectedDepartments, departmentStats.NRows())
	}

	// Verify column structure
	expectedCols := []string{"department", "n", "mean(salary)", "sum(salary)", "min(salary)", "max(salary)", "std(salary)"}
	actualCols := departmentStats.Names()
	for i, expected := range expectedCols {
		if i >= len(actualCols) || actualCols[i] != expected {
			t.Errorf("Expected column %s at position %d, got %v", expected, i, actualCols)
		}
	}

	// Verify specific department statistics
	for i := 0; i < departmentStats.NRows(); i++ {
		dept := departmentStats.C("department").Get(i).(string)
		count := departmentStats.C("n").Get(i).(int64)

		switch dept {
		case "IT":
			if count != 3 { // Bob, Charlie, Frank
				t.Errorf("IT department: expected 3 employees, got %d", count)
			}
		case "HR":
			if count != 2 { // Alice, David
				t.Errorf("HR department: expected 2 employees, got %d", count)
			}
		case "Finance":
			if count != 1 { // Eve
				t.Errorf("Finance department: expected 1 employee, got %d", count)
			}
		}
	}

	// Test multiple join types in sequence
	// Right join to see what salary records don't have employees
	rightJoinResult := employees.Join(RIGHT_JOIN, salaries, "emp_id")
	if rightJoinResult.GetError() != nil {
		t.Fatal("Right join failed:", rightJoinResult.GetError())
	}

	// Should have 7 rows (all salary records)
	if rightJoinResult.NRows() != 7 {
		t.Errorf("Right join: expected 7 rows, got %d", rightJoinResult.NRows())
	}

	// Outer join to see complete picture
	outerJoinResult := employees.Join(OUTER_JOIN, salaries, "emp_id")
	if outerJoinResult.GetError() != nil {
		t.Fatal("Outer join failed:", outerJoinResult.GetError())
	}

	// Should have 7 rows (employees 1-6 match, salary record 7 is unmatched)
	if outerJoinResult.NRows() != 7 {
		t.Errorf("Outer join: expected 7 rows, got %d", outerJoinResult.NRows())
	}
}

func TestIntegration_ChainedOperations(t *testing.T) {
	// Test a complex chain of operations
	sales := NewDataFrame(integrationCtx).
		AddSeriesFromStrings("region", []string{"North", "South", "East", "West", "North", "South"}, nil, false).
		AddSeriesFromStrings("product", []string{"A", "B", "A", "C", "B", "A"}, nil, false).
		AddSeriesFromFloat64s("revenue", []float64{1000, 1500, 800, 1200, 900, 1100}, nil, false).
		AddSeriesFromInt64s("units", []int64{10, 15, 8, 12, 9, 11}, nil, false)

	targets := NewDataFrame(integrationCtx).
		AddSeriesFromStrings("region", []string{"North", "South", "East", "West"}, nil, false).
		AddSeriesFromFloat64s("target", []float64{2000, 2500, 1000, 1500}, nil, false)

	// Chain operations: Join -> Group -> Calculate metrics
	result := sales.
		Join(INNER_JOIN, targets, "region").
		GroupBy("region").
		Agg(
			Sum("revenue"),
			Sum("units"),
			Mean("target"), // Target should be same for all products in region
		).Run()

	if result.GetError() != nil {
		t.Fatal("Chained operations failed:", result.GetError())
	}

	if result.NRows() != 4 { // 4 regions
		t.Errorf("Expected 4 regions, got %d", result.NRows())
	}

	// Verify that each region's data is correctly aggregated
	for i := 0; i < result.NRows(); i++ {
		region := result.C("region").Get(i).(string)
		totalRevenue := result.C("sum(revenue)").Get(i).(float64)
		target := result.C("mean(target)").Get(i).(float64)

		switch region {
		case "North":
			expectedRevenue := 1000.0 + 900.0 // A + B products
			if totalRevenue != expectedRevenue {
				t.Errorf("North region: expected revenue %f, got %f", expectedRevenue, totalRevenue)
			}
			if target != 2000.0 {
				t.Errorf("North region: expected target 2000, got %f", target)
			}
		case "South":
			expectedRevenue := 1500.0 + 1100.0 // B + A products
			if totalRevenue != expectedRevenue {
				t.Errorf("South region: expected revenue %f, got %f", expectedRevenue, totalRevenue)
			}
		}
	}
}

func TestIntegration_AllAggregationsWithComplexData(t *testing.T) {
	// Test all aggregation functions work correctly together
	data := NewDataFrame(integrationCtx).
		AddSeriesFromStrings("category", []string{"A", "A", "B", "B", "A", "C"}, nil, false).
		AddSeriesFromFloat64s("value1", []float64{10.5, 20.3, 15.7, 25.1, 12.9, 30.0}, nil, false).
		AddSeriesFromFloat64s("value2", []float64{100, 200, 150, 250, 120, 300}, nil, false).
		AddSeriesFromInt64s("count_field", []int64{1, 2, 1, 3, 1, 2}, nil, false)

	// Use all aggregation functions
	result := data.GroupBy("category").
		Agg(
			Count(),
			Sum("value1"),
			Mean("value1"),
			Min("value1"),
			Max("value1"),
			Std("value1"),
			Sum("value2"),
			Mean("count_field"),
		).Run()

	if result.GetError() != nil {
		t.Fatal("All aggregations failed:", result.GetError())
	}

	expectedCols := []string{
		"category", "n", "sum(value1)", "mean(value1)", "min(value1)",
		"max(value1)", "std(value1)", "sum(value2)", "mean(count_field)",
	}

	actualCols := result.Names()
	for i, expected := range expectedCols {
		if i >= len(actualCols) || actualCols[i] != expected {
			t.Errorf("Expected column %s at position %d, got %v", expected, i, actualCols)
		}
	}

	// Verify specific calculations for category A
	for i := 0; i < result.NRows(); i++ {
		category := result.C("category").Get(i).(string)
		if category == "A" {
			count := result.C("n").Get(i).(int64)
			minVal := result.C("min(value1)").Get(i).(float64)
			maxVal := result.C("max(value1)").Get(i).(float64)

			if count != 3 { // 3 records for category A
				t.Errorf("Category A: expected count 3, got %d", count)
			}
			if minVal != 10.5 {
				t.Errorf("Category A: expected min 10.5, got %f", minVal)
			}
			if maxVal != 20.3 {
				t.Errorf("Category A: expected max 20.3, got %f", maxVal)
			}
		}
	}
}
