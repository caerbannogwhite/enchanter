package series

import "github.com/caerbannogwhite/enchanter/utils"

// binaryNullMask builds the null mask of a binary element-wise operation and
// reports whether the result is nullable.
//
// aScalar and bScalar are compile-time facts of the generated call site — the
// code generator emits one call per length case with literal booleans — so the
// dispatch below runs once per operation, not per element. The cases reproduce
// exactly what the generator used to emit inline for each of the four
// nullability combinations:
//
//   - both operands nullable: the masks are OR-combined, broadcasting a scalar
//     operand's single bit across the result;
//   - one operand nullable: its mask is broadcast (scalar) or copied (vector)
//     into a fresh mask of the result's size;
//   - neither nullable: the mask is empty and the result is not nullable.
func binaryNullMask(aNullable bool, aMask []uint8, aScalar bool, bNullable bool, bMask []uint8, bScalar bool, size int) ([]uint8, bool) {
	switch {
	case aNullable && bNullable:
		mask := utils.BinVecInit(size, false)
		switch {
		case aScalar && bScalar:
			utils.BinVecOrSS(aMask, bMask, mask)
		case aScalar:
			utils.BinVecOrSV(aMask, bMask, mask)
		case bScalar:
			utils.BinVecOrVS(aMask, bMask, mask)
		default:
			utils.BinVecOrVV(aMask, bMask, mask)
		}
		return mask, true

	case aNullable:
		if aScalar {
			return utils.BinVecInit(size, aMask[0] == 1), true
		}
		mask := utils.BinVecInit(size, false)
		copy(mask, aMask)
		return mask, true

	case bNullable:
		if bScalar {
			return utils.BinVecInit(size, bMask[0] == 1), true
		}
		mask := utils.BinVecInit(size, false)
		copy(mask, bMask)
		return mask, true

	default:
		return utils.BinVecInit(0, false), false
	}
}

// naOperandNullMask builds the null mask of an operation whose other operand
// is NAs but whose result is a typed series (for example Bools.Or(NAs)).
//
// It preserves the long-standing behavior of this case verbatim: the result
// carries a copy of the typed operand's null mask — broadcast when that
// operand is a scalar — while the result series itself is marked not nullable.
// That combination is questionable, but changing it is a semantic decision,
// not a code-generation one.
func naOperandNullMask(nullable bool, mask []uint8, scalar bool, size int) []uint8 {
	if !nullable {
		return make([]uint8, 0)
	}
	if scalar {
		return utils.BinVecInit(size, mask[0] == 1)
	}
	m := utils.BinVecInit(size, false)
	copy(m, mask)
	return m
}

// coalesceNullMask builds the null mask of Coalesce and reports whether the
// result is nullable. A result position is null only when both operands are
// null there, so if either operand has no nulls the result has none.
//
// Like binaryNullMask, the scalar flags are compile-time facts of the
// generated call site, and a scalar operand's single mask bit is broadcast
// across the result.
func coalesceNullMask(aNullable bool, aMask []uint8, aScalar bool, bNullable bool, bMask []uint8, bScalar bool, size int) ([]uint8, bool) {
	if !aNullable || !bNullable {
		return utils.BinVecInit(0, false), false
	}

	mask := utils.BinVecInit(size, false)
	switch {
	case aScalar && bScalar:
		utils.BinVecAndSS(aMask, bMask, mask)
	case aScalar:
		utils.BinVecAndSV(aMask, bMask, mask)
	case bScalar:
		utils.BinVecAndVS(aMask, bMask, mask)
	default:
		utils.BinVecAndVV(aMask, bMask, mask)
	}
	return mask, true
}
