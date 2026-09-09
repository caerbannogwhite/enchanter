package utils

// This function merges two null masks into one.
// Any of the two masks can be empty.
func MergeNullMasks(s1Len int, s1Nullable bool, s1Mask []uint8, s2Len int, s2Nullable bool, s2Mask []uint8) (bool, []uint8) {
	dataLen := s1Len + s2Len

	if s1Nullable {
		if dataLen > len(s1Mask)<<3 {
			s1Mask = append(s1Mask, make([]uint8, (dataLen>>3)-len(s1Mask)+1)...)
		}

		if s2Nullable {

			sIdx := s1Len
			oIdx := 0
			for _, v := range s2Mask {
				for j := 0; j < 8; j++ {
					if v&(1<<uint(j)) != 0 {
						s1Mask[sIdx>>3] |= 1 << uint(sIdx%8)
					}
					sIdx++
					oIdx++
				}
			}

			return true, s1Mask
		} else {
			return true, s1Mask
		}
	} else {
		if s2Nullable {
			s1Mask = make([]uint8, (dataLen>>3)+1)

			sIdx := s1Len
			oIdx := 0
			for _, v := range s2Mask {
				for j := 0; j < 8; j++ {
					if v&(1<<uint(j)) != 0 {
						s1Mask[sIdx>>3] |= 1 << uint(sIdx%8)
					}
					sIdx++
					oIdx++
				}
			}

			return true, s1Mask
		} else {
			return false, make([]uint8, 0)
		}
	}
}

func BinVecInit(size int, flag bool) []uint8 {
	vec := make([]uint8, (size+7)>>3)
	if flag {
		for i := range vec {
			vec[i] = 0xFF
		}
		if size%8 != 0 {
			vec[len(vec)-1] >>= uint(8 - (size % 8))
		}
	}
	return vec
}

func BinVecFromBools(v []bool) []uint8 {
	binVec := make([]uint8, (len(v)+7)>>3)
	for i := 0; i < len(v); i++ {
		if v[i] {
			binVec[i>>3] |= 1 << uint(i%8)
		}
	}
	return binVec
}

func BinVecSet(v []uint8, i int, flag bool) {
	if flag {
		v[i>>3] |= 1 << uint(i%8)
	} else {
		v[i>>3] &= ^(1 << uint(i%8))
	}
}

func BinVecResize(v []uint8, size int) []uint8 {
	if size <= len(v)<<3 {
		return v
	}
	v2 := make([]uint8, (size+7)>>3)
	copy(v2, v)
	return v2
}

func BinVecFilterByIndices(srcVec []uint8, indices []int) []uint8 {
	dstVec := make([]uint8, (len(indices)+7)>>3)
	for dstIdx, srcIdx := range indices {
		if srcIdx%8 > dstIdx%8 {
			dstVec[dstIdx>>3] |= ((srcVec[srcIdx>>3] & (1 << uint(srcIdx%8))) >> uint(srcIdx%8-dstIdx%8))
		} else {
			dstVec[dstIdx>>3] |= ((srcVec[srcIdx>>3] & (1 << uint(srcIdx%8))) << uint(dstIdx%8-srcIdx%8))
		}
	}
	return dstVec
}

func BinVecCount(v []uint8) int {
	count := 0
	for _, x := range v {
		for x != 0 {
			count += int(x & 1)
			x >>= 1
		}
	}
	return count
}

// This function computes the bitwise OR of two binary vectors.
// The result is stored in the third argument.
func BinVecOrSS(a, b, res []uint8) {
	res[0] = a[0] | b[0]
}

// This function computes the bitwise OR of a binary vectors.
// The result is stored in the second argument.
func BinVecOrSV(a, b, res []uint8) {
	if a[0] == 0 {
		copy(res, b)
	} else {
		for i := range res {
			res[i] = 0xFF
		}
	}
}

// This function computes the bitwise OR of a binary vectors.
// The result is stored in the second argument.
func BinVecOrVS(a, b, res []uint8) {
	if b[0] == 0 {
		copy(res, a)
	} else {
		for i := range res {
			res[i] = 0xFF
		}
	}
}

// This function computes the bitwise OR of two binary vectors.
// The result is stored in the third argument.
func BinVecOrVV(a, b, res []uint8) {
	for i := range res {
		res[i] = a[i] | b[i]
	}
}

// This function computes the bitwise AND of two scalar binary vectors.
// The result is stored in the third argument.
func BinVecAndSS(a, b, res []uint8) {
	res[0] = a[0] & b[0]
}

// This function computes the bitwise AND of a scalar and a vector binary
// vector, broadcasting the scalar. The result is stored in the third
// argument.
func BinVecAndSV(a, b, res []uint8) {
	if a[0] == 0 {
		for i := range res {
			res[i] = 0
		}
	} else {
		copy(res, b)
	}
}

// This function computes the bitwise AND of a vector and a scalar binary
// vector, broadcasting the scalar. The result is stored in the third
// argument.
func BinVecAndVS(a, b, res []uint8) {
	if b[0] == 0 {
		for i := range res {
			res[i] = 0
		}
	} else {
		copy(res, a)
	}
}

// This function computes the bitwise AND of two binary vectors.
// The result is stored in the third argument.
func BinVecAndVV(a, b, res []uint8) {
	for i := range res {
		res[i] = a[i] & b[i]
	}
}
