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
