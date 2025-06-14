package main

import (
	"testing"
	"tetrisServer/asm"
)

func TestIntersectsAVXFunctions(t *testing.T) {
	testDataA := [][16]uint16{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{65535, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30},
	}
	testDataB := [][16]uint16{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31},
	}

	n := len(testDataA)
	if len(testDataB) != n {
		t.Fatal("Test data length mismatch")
	}

	// Проверяем IntersectsAVXSingle и intersectsSimdeSingle поэлементно
	for i := 0; i < n; i++ {
		a := &testDataA[i]
		b := &testDataB[i]

		gotAsm := asm.IntersectsAVXSingle(a, b)
		gotSimde := intersectsSimdeSingle(a[:], b[:]) // <-- здесь

		if gotAsm != gotSimde {
			t.Errorf("IntersectsAVXSingle mismatch at index %d: asm=%v simde=%v", i, gotAsm, gotSimde)
		}
	}

	// Подготавливаем массивы указателей для IntersectsAVXMultiple
	ptrA := make([]*[16]uint16, n)
	ptrB := make([]*[16]uint16, n)
	for i := 0; i < n; i++ {
		ptrA[i] = &testDataA[i]
		ptrB[i] = &testDataB[i]
	}

	gotMultipleAsm := asm.IntersectsAVXMultiple(&ptrA[0], &ptrB[0], n)
	gotMultipleSimde := intersectsSimdeMany(flatten(testDataA), flatten(testDataB))

	if gotMultipleAsm != gotMultipleSimde {
		t.Errorf("IntersectsAVXMultiple mismatch: asm=%d simde=%d", gotMultipleAsm, gotMultipleSimde)
	}

	var sumSingles int
	for i := 0; i < n; i++ {
		a := &testDataA[i]
		b := &testDataB[i]

		gotAsm := asm.IntersectsAVXSingle(a, b)
		if gotAsm {
			sumSingles++
		}
	}
	if sumSingles != gotMultipleAsm {
		t.Errorf("Sum of single intersections (%d) != multiple intersections (%d)", sumSingles, gotMultipleAsm)
	}
}

// flatten превращает слайс [][16]uint16 в []uint16 (конкатенация)
func flatten(slices [][16]uint16) []uint16 {
	out := make([]uint16, 0, len(slices)*16)
	for _, arr := range slices {
		out = append(out, arr[:]...)
	}
	return out
}
