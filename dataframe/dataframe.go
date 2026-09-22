package dataframe

// JoinType selects the flavor of DataFrame.Join.
type JoinType int8

const (
	JoinInner JoinType = iota
	JoinLeft
	JoinRight
	JoinOuter
)
