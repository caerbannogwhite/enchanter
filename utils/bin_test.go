package utils

import (
	"testing"
)

func Test_BinVec_Init(t *testing.T) {
	var vecA, vecB []uint8

	// 6
	vecA = BinVecFromBools([]bool{true, true, true, true, true, false})
	vecB = BinVecInit(6, true)
	BinVecSet(vecB, 5, false)
	if vecA[0] != 0x1F {
		t.Errorf("Expected 0x1F, got %x", vecA[0])
	}
	if vecB[0] != 0x1F {
		t.Errorf("Expected 0x1F, got %x", vecB[0])
	}

	vecA = BinVecFromBools([]bool{true, true, true, true, true, true})
	vecB = BinVecInit(6, true)
	if vecA[0] != 0x3F {
		t.Errorf("Expected 0x3F, got %x", vecA[0])
	}
	if vecB[0] != 0x3F {
		t.Errorf("Expected 0x3F, got %x", vecB[0])
	}

	// 7
	vecA = BinVecFromBools([]bool{true, true, true, true, true, true, false})
	vecB = BinVecInit(7, true)
	BinVecSet(vecB, 6, false)
	if vecA[0] != 0x3F {
		t.Errorf("Expected 0x6F, got %x", vecA[0])
	}
	if vecB[0] != 0x3F {
		t.Errorf("Expected 0x6F, got %x", vecB[0])
	}

	vecA = BinVecFromBools([]bool{true, true, true, true, true, true, true})
	vecB = BinVecInit(7, true)
	if vecA[0] != 0x7F {
		t.Errorf("Expected 0x7F, got %x", vecA[0])
	}
	if vecB[0] != 0x7F {
		t.Errorf("Expected 0x7F, got %x", vecB[0])
	}

	// 8
	vecA = BinVecFromBools([]bool{true, true, true, true, true, true, true, false})
	vecB = BinVecInit(8, true)
	BinVecSet(vecB, 7, false)
	if vecA[0] != 0x7F {
		t.Errorf("Expected 0x7F, got %x", vecA[0])
	}
	if vecB[0] != 0x7F {
		t.Errorf("Expected 0x7F, got %x", vecB[0])
	}

	vecA = BinVecFromBools([]bool{true, true, true, true, true, true, true, true})
	vecB = BinVecInit(8, true)
	if vecA[0] != 0xFF {
		t.Errorf("Expected 0xFF, got %x", vecA[0])
	}
	if vecB[0] != 0xFF {
		t.Errorf("Expected 0xFF, got %x", vecB[0])
	}

	// 9
	vecA = BinVecFromBools([]bool{true, true, true, true, true, true, true, true, false})
	vecB = BinVecInit(9, true)
	BinVecSet(vecB, 8, false)
	if vecA[0] != 0xFF || vecA[1] != 0x00 {
		t.Errorf("Expected 0xFF 0x00, got %x %x", vecA[0], vecA[1])
	}
	if vecB[0] != 0xFF || vecB[1] != 0x00 {
		t.Errorf("Expected 0xFF 0x00, got %x %x", vecB[0], vecB[1])
	}

	vecA = BinVecFromBools([]bool{true, true, true, true, true, true, true, true, true})
	vecB = BinVecInit(9, true)
	if vecA[0] != 0xFF || vecA[1] != 0x01 {
		t.Errorf("Expected 0xFF 0x01, got %x %x", vecA[0], vecA[1])
	}
	if vecB[0] != 0xFF || vecB[1] != 0x01 {
		t.Errorf("Expected 0xFF 0x01, got %x %x", vecB[0], vecB[1])
	}
}

func Test_BinVec_Count(t *testing.T) {

	var v []uint8

	v = []uint8{0x00}
	if c := BinVecCount(v); c != 0 {
		t.Errorf("Expected 0, got %d", c)
	}

	v = []uint8{0x01}
	if c := BinVecCount(v); c != 1 {
		t.Errorf("Expected 1, got %d", c)
	}

	v = []uint8{0x00, 0x00, 0x00, 0x00, 0x00}
	if c := BinVecCount(v); c != 0 {
		t.Errorf("Expected 0, got %d", c)
	}

	v = []uint8{0x00, 0x01, 0x00, 0x00, 0x00}
	if c := BinVecCount(v); c != 1 {
		t.Errorf("Expected 1, got %d", c)
	}

	v = []uint8{0x00, 0x01, 0x00, 0x00, 0x01}
	if c := BinVecCount(v); c != 2 {
		t.Errorf("Expected 2, got %d", c)
	}

	v = []uint8{0x01, 0x01, 0x01, 0x01, 0x01}
	if c := BinVecCount(v); c != 5 {
		t.Errorf("Expected 5, got %d", c)
	}

	v = []uint8{0x01, 0x02, 0x04, 0x08}
	if c := BinVecCount(v); c != 4 {
		t.Errorf("Expected 4, got %d", c)
	}

	v = []uint8{0x01, 0x03, 0x07, 0x0F}
	if c := BinVecCount(v); c != 10 {
		t.Errorf("Expected 10, got %d", c)
	}
}

func Benchmark_BinVec_Count(b *testing.B) {
	v := make([]uint8, 1_000_000)
	for i := 0; i < len(v); i++ {
		v[i] = uint8(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BinVecCount(v)
	}
}

func Test_BinVec_And(t *testing.T) {
	// Vector and vector: 10 elements, byte-wise AND.
	a := BinVecFromBools([]bool{true, true, false, false, true, false, true, true, true, false})
	b := BinVecFromBools([]bool{true, false, true, false, true, true, false, true, false, false})
	res := BinVecInit(10, false)
	BinVecAndVV(a, b, res)
	want := BinVecFromBools([]bool{true, false, false, false, true, false, false, true, false, false})
	if res[0] != want[0] || res[1] != want[1] {
		t.Errorf("AndVV: expected %x %x, got %x %x", want[0], want[1], res[0], res[1])
	}

	// Scalar and scalar.
	res = BinVecInit(1, false)
	BinVecAndSS(BinVecInit(1, true), BinVecInit(1, true), res)
	if res[0] != 1 {
		t.Errorf("AndSS true&true: expected 1, got %x", res[0])
	}
	res = BinVecInit(1, false)
	BinVecAndSS(BinVecInit(1, true), BinVecInit(1, false), res)
	if res[0] != 0 {
		t.Errorf("AndSS true&false: expected 0, got %x", res[0])
	}

	// Scalar broadcast: a false scalar clears everything.
	res = BinVecInit(10, true)
	BinVecAndSV(BinVecInit(1, false), b, res)
	if res[0] != 0 || res[1] != 0 {
		t.Errorf("AndSV false: expected zeros, got %x %x", res[0], res[1])
	}

	// Scalar broadcast: a true scalar keeps the vector.
	res = BinVecInit(10, false)
	BinVecAndSV(BinVecInit(1, true), b, res)
	if res[0] != b[0] || res[1] != b[1] {
		t.Errorf("AndSV true: expected %x %x, got %x %x", b[0], b[1], res[0], res[1])
	}

	// The mirrored forms.
	res = BinVecInit(10, false)
	BinVecAndVS(a, BinVecInit(1, true), res)
	if res[0] != a[0] || res[1] != a[1] {
		t.Errorf("AndVS true: expected %x %x, got %x %x", a[0], a[1], res[0], res[1])
	}
	res = BinVecInit(10, true)
	BinVecAndVS(a, BinVecInit(1, false), res)
	if res[0] != 0 || res[1] != 0 {
		t.Errorf("AndVS false: expected zeros, got %x %x", res[0], res[1])
	}
}
