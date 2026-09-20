package meta

import (
	"testing"
)

func TestTypeSystem(t *testing.T) {

	var res Primitive

	res = OP_BINARY_MUL.GetBinaryOpResultType(
		Primitive{Base: IntType},
		Primitive{Base: IntType},
	)

	if res.Base != IntType {
		t.Errorf("Expected IntType, got %v", res.Base)
	}

	res = OP_BINARY_MUL.GetBinaryOpResultType(
		Primitive{Base: IntType},
		Primitive{Base: Float64Type},
	)

	if res.Base != Float64Type {
		t.Errorf("Expected Float64Type, got %v", res.Base)
	}

	res = OP_BINARY_MUL.GetBinaryOpResultType(
		Primitive{Base: Float64Type},
		Primitive{Base: IntType},
	)

	if res.Base != Float64Type {
		t.Errorf("Expected Float64Type, got %v", res.Base)
	}
}

func Test_Coalesce_ResultTypes(t *testing.T) {
	cases := []struct {
		l, r, want BaseType
	}{
		{Float64Type, Float64Type, Float64Type},
		{IntType, Int64Type, Int64Type},
		{Int64Type, IntType, Int64Type},
		{IntType, Float64Type, Float64Type},
		{Float64Type, Int64Type, Float64Type},
		{BoolType, BoolType, BoolType},
		{StringType, StringType, StringType},
		{TimeType, TimeType, TimeType},
		{DurationType, DurationType, DurationType},
		{Float64Type, NullType, Float64Type},
		{NullType, StringType, StringType},
		{NullType, NullType, NullType},
		{BoolType, Int64Type, ErrorType},
		{StringType, Float64Type, ErrorType},
		{TimeType, DurationType, ErrorType},
	}
	for _, c := range cases {
		res := OP_BINARY_COALESCE.GetBinaryOpResultType(
			Primitive{Base: c.l, Size: 5}, Primitive{Base: c.r, Size: 5})
		if res.Base != c.want {
			t.Errorf("coalesce(%v, %v): expected %v, got %v", c.l, c.r, c.want, res.Base)
		}
	}
	if !OP_BINARY_COALESCE.IsBinaryOp() {
		t.Error("OP_BINARY_COALESCE must report itself as a binary operator")
	}
	if OP_BINARY_COALESCE.IsCommutative() {
		t.Error("coalesce is not commutative")
	}
}

// Every binary operation propagates NA: a valid combination with an NA
// operand yields NA. Coalesce is the exception, tested above.
func Test_NA_Propagation(t *testing.T) {
	ops := []OPCODE{
		OP_BINARY_MUL, OP_BINARY_DIV, OP_BINARY_MOD, OP_BINARY_EXP,
		OP_BINARY_ADD, OP_BINARY_SUB,
		OP_BINARY_EQ, OP_BINARY_NE, OP_BINARY_LT, OP_BINARY_LE, OP_BINARY_GT, OP_BINARY_GE,
		OP_BINARY_AND, OP_BINARY_OR,
	}
	types := []BaseType{NullType, BoolType, IntType, Int64Type, Float64Type, StringType, TimeType, DurationType}
	for _, op := range ops {
		for _, l := range types {
			for _, r := range types {
				if l != NullType && r != NullType {
					continue
				}
				res := op.GetBinaryOpResultType(Primitive{Base: l, Size: 3}, Primitive{Base: r, Size: 3})
				if res.Base != NullType && res.Base != ErrorType {
					t.Errorf("%v (%v, %v): expected NA or error, got %v", op, l, r, res.Base)
				}
			}
		}
	}
}
