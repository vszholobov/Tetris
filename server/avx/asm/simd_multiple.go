package asm

// IntersectsAVXMany принимает слайсы указателей на [16]uint16.
// Возвращает количество пар, у которых AND ≠ 0.
//
//go:noescape
func IntersectsAVXMultiple(a, b **[16]uint16, n int) int
