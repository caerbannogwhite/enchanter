package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/dataframe"
	"github.com/caerbannogwhite/enchanter/io"
)

const (
	data1 = `
name,surname,age,junior,department,salary band
Alice,C,29,F,HR,4
John,Doe,30,true,IT,2
Bob,Smith,31,F,IT,4
Jane,H,25,false,IT,4
Mary,Jane,28,false,IT,3
Oliver,C,32,true,HR,1
Ursula,Alle,27,f,Business,4
Charlie,Brown,33,False,Business,2
Eve,Black,26,F,Business,3
Frank,White,34,T,Business,1
Anna,Green,39,f,Operations,4
`

	data2 = `
department,number of employees,budget
IT,4,100000
HR,2,50000
Business,4,100000
Operations,5,250000
`
)

var ctx = enchanter.NewContext()

func Example01() {
	dataframe.NewDataFrame(ctx).
		FromCsv().
		SetReader(strings.NewReader(data1)).
		SetDelimiter(',').
		SetHeader(true).
		Read().
		Select("department", "age", "junior", "salary band").
		GroupBy("department").
		Agg(dataframe.Max("age"), dataframe.Min("salary band"), dataframe.Mean("junior"), dataframe.Count()).
		Run().
		PPrint(
			dataframe.NewPPrintParams().
				SetUseLipGloss(true).
				SetWidth(130).
				SetNRows(50))

	// Output:
	// ╭────────────┬─────────┬─────────────┬─────────┬───────╮
	// │ department │ age     │ salary band │ junior  │ n     │
	// ├────────────┼─────────┼─────────────┼─────────┼───────┤
	// │ String     │ Float64 │ Float64     │ Float64 │ Int64 │
	// ├────────────┼─────────┼─────────────┼─────────┼───────┤
	// │ Business   │   34.00 │       1.000 │  0.2500 │ 4.000 │
	// │ Operations │   39.00 │       4.000 │       0 │ 1.000 │
	// │ HR         │   32.00 │       1.000 │  0.5000 │ 2.000 │
	// │ IT         │   31.00 │       2.000 │  0.2500 │ 4.000 │
	// ╰────────────┴─────────┴─────────────┴─────────┴───────╯
}

func Example02() {
	employees := dataframe.NewDataFrame(ctx).
		FromCsv().
		SetReader(strings.NewReader(data1)).
		SetDelimiter(',').
		SetHeader(true).
		Read()

	departments := dataframe.NewDataFrame(ctx).
		FromCsv().
		SetReader(strings.NewReader(data2)).
		SetDelimiter(',').
		SetHeader(true).
		Read()

	departments.PPrint(dataframe.NewPPrintParams())

	employees.Join(dataframe.JoinLeft, departments, "department").
		PPrint(dataframe.NewPPrintParams())
}

func Example03() {
	df := dataframe.NewDataFrame(ctx).
		FromCsv().
		SetReader(strings.NewReader(data1)).
		SetDelimiter(',').
		SetHeader(true).
		Read()

	df.Filter(
		df.Col("age").Ge(30).
			And(df.Col("junior").
				Or(df.Col("department").Eq("Business")))).
		PPrint(dataframe.NewPPrintParams())
}

func Example04() {
	x := `
a,b
1,2
2,2
3,3
3,4
4,4
`

	y := `
a,b
1,2
2,2
2,3
3,3
4,4
`

	ppp := dataframe.NewPPrintParams()

	dfX := dataframe.NewDataFrame(ctx).
		FromCsv().
		SetReader(strings.NewReader(x)).
		SetDelimiter(',').
		SetHeader(true).
		Read()

	dfY := dataframe.NewDataFrame(ctx).
		FromCsv().
		SetReader(strings.NewReader(y)).
		SetDelimiter(',').
		SetHeader(true).
		Read()

	dfX.Join(dataframe.JoinInner, dfY, "a", "b").
		PPrint(ppp)

	dfX.Join(dataframe.JoinLeft, dfY, "a", "b").
		PPrint(ppp)

	dfX.Join(dataframe.JoinRight, dfY, "a", "b").
		PPrint(ppp)

	dfX.Join(dataframe.JoinOuter, dfY, "a", "b").
		PPrint(ppp)
}

func Example05() {
	dataframe.NewDataFrame(ctx).
		FromXpt().
		SetPath("../testdata/CDBRFS90.XPT").
		// SetPath("../testdata/xpt_test_mixed.xpt").
		SetVersion(io.XPT_VERSION_9).
		// SetMaxObservations(10).
		Read().
		Take(100).

		// to SAS XPT
		// ToXpt().
		// SetPath("../testdata/mixed_out.XPT").
		// SetVersion(XPT_VERSION_9).
		// Write().

		// to Excel
		// ToXlsx().
		// SetPath("../testdata/test.xlsx").
		// SetSheet("test").
		// SetNaText("").
		// Write().

		// to HTML
		// ToHtml().
		// SetPath("../testdata/test.html").
		// SetDataTable(true).
		// SetNaText("-").
		// SetNewLine("\n").
		// SetIndent("  ").
		// Write().

		// to JSON
		// ToJson().
		// SetPath("../testdata/test.json").
		// Write().

		// Pretty print
		PPrint(
			dataframe.NewPPrintParams().
				SetUseLipGloss(true).
				SetWidth(200).
				SetNRows(10))
}

func Example06() {
	df := dataframe.NewDataFrame(ctx).
		FromCsv().
		SetNullValues(true).
		// SetRows(20).
		SetPath(filepath.Join("..", "testdata", "G1_1e4_1e2_10_0.csv")).
		Read()

	df.PPrint(dataframe.NewPPrintParams().SetNRows(10).SetUseLipGloss(true))

	df = df.GroupBy("id6").
		Agg(dataframe.Sum("v1"), dataframe.Sum("v2"), dataframe.Sum("v3")).
		RemoveNAs(true).
		Run().
		PPrint(dataframe.NewPPrintParams().SetNRows(10).SetUseLipGloss(true))

	fmt.Println(df.Agg(dataframe.Sum("sum(v1)")).Run().Col("sum(sum(v1))"))
}

func main() {
	fmt.Println("Example01:")
	Example01()

	fmt.Println("Example02:")
	Example02()

	fmt.Println("Example03:")
	Example03()

	fmt.Println("Example04:")
	Example04()

	fmt.Println("Example05:")
	Example05()

	fmt.Println("Example06:")
	Example06()
}
