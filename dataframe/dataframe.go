package dataframe

type DataFrameJoinType int8

const (
	INNER_JOIN DataFrameJoinType = iota
	LEFT_JOIN
	RIGHT_JOIN
	OUTER_JOIN
)
