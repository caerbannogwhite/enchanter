package meta

import "testing"

// Temporary scan: list every binary-op combination with an NA operand
// whose result is typed (neither NA nor error).
func Test_TempNAScan(t *testing.T) {
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
					t.Logf("TYPED: %v (%v, %v) -> %v", op, l, r, res.Base)
				}
			}
		}
	}
}
