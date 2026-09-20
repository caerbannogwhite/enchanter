package series

import (
	"fmt"
	"strconv"

	"github.com/caerbannogwhite/enchanter"
)

func boolToString(b bool) string {
	if b {
		return enchanter.BOOL_TRUE_TEXT
	} else {
		return enchanter.BOOL_FALSE_TEXT
	}
}

func intToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func debugPrintPartition(p SeriesPartition, series ...Series) {
	map_ := p.GetMap()

	header := ""
	separators := ""
	for i := range series {
		header += fmt.Sprintf("| %-10d ", i)
		separators += "|------------"
	}

	fmt.Println()
	fmt.Printf("    | %-20s %s | %-20s |\n", "Key", header, "Indices")
	fmt.Printf("    |%s%s-|%s|\n", "----------------------", separators, "----------------------")
	for k, v := range map_ {
		vals := ""
		for _, s := range series {
			vals += fmt.Sprintf("| %-10s ", s.GetAsString(v[0]))
		}

		indices := ""
		for _, i := range v {
			indices += fmt.Sprintf("%d ", i)
		}
		fmt.Printf("    | %-20d %s | %-20s |\n", k, vals, indices)
	}
	fmt.Println()
}
